package refresh_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"niko1ay-pr/url-shrtnr/internal/http-server/handlers/users/refresh"
	"niko1ay-pr/url-shrtnr/internal/http-server/handlers/users/refresh/mocks"
	"niko1ay-pr/url-shrtnr/internal/lib/logger/sldiscard"
	"niko1ay-pr/url-shrtnr/internal/models"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockSetup func(*mocks.RefreshTokenGetStorer)

var testCases = []struct {
	name             string
	request          refresh.Request
	mockSetup        mockSetup
	expectedStatus   int
	validateResponse func(*testing.T, *httptest.ResponseRecorder)
}{
	{name: "Success refresh",
		request: refresh.Request{
			Id:           1,
			Email:        "somevalid@mail.com",
			RefreshToken: "ValidRefreshToken",
		},
		mockSetup: func(m *mocks.RefreshTokenGetStorer) {
			m.On("GetRefreshToken", mock.Anything, mock.AnythingOfType("int64")).
				Return(&models.RefreshToken{Id: 1, Token: "ValidRefreshToken"}, nil).Once()
			m.On("StoreRefreshToken", mock.Anything, mock.AnythingOfType("*models.RefreshToken")).
				Return(nil).Once()
		},
		expectedStatus: http.StatusOK,
		validateResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
			var resp refresh.Response
			err := json.Unmarshal(rr.Body.Bytes(), &resp)
			require.NoError(t, err)
			require.NotEmpty(t, resp.Token)
			require.NotEmpty(t, resp.RefreshToken)
			require.Equal(t, int64(1), resp.Id)
		},
	},
	{name: "Refresh token not match",
		request: refresh.Request{
			Id:           1,
			Email:        "somevalid@mail.com",
			RefreshToken: "InvalidRefreshToken",
		},
		mockSetup: func(m *mocks.RefreshTokenGetStorer) {
			m.On("GetRefreshToken", mock.Anything, mock.AnythingOfType("int64")).
				Return(&models.RefreshToken{Id: 1, Token: "ValidRefreshToken"}, nil).Once()
		},
		expectedStatus: http.StatusBadRequest,
		validateResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
		},
	},
	{name: "Required fields missed",
		request: refresh.Request{
			Id: 1,
		},
		expectedStatus: http.StatusBadRequest,
	},
	{name: "Negative Id",
		request: refresh.Request{
			Id:           -1,
			Email:        "somevalid@mail.com",
			RefreshToken: "InvalidRefreshToken",
		},
		expectedStatus: http.StatusBadRequest,
	},
}

func TestRefreshHandler(t *testing.T) {

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			storageMock := mocks.NewRefreshTokenGetStorer(t)
			if tc.mockSetup != nil {
				tc.mockSetup(storageMock)
			}

			log := sldiscard.NewDescardLogger()
			secret := []byte("secret")
			handler := refresh.New(context.Background(), log, storageMock, secret)

			request, err := json.Marshal(tc.request)
			require.NoError(t, err)
			req, err := http.NewRequest("POST", "/users/refresh", bytes.NewReader(request))
			require.NoError(t, err)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			require.Equal(t, tc.expectedStatus, rr.Code)

			if tc.validateResponse != nil {
				tc.validateResponse(t, rr)
			}
			storageMock.AssertExpectations(t)
		})
	}

}
