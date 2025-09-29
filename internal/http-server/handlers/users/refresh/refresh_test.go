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
	expectedErr      string
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
		expectedErr:    "failed to refresh token",
	},
	{name: "Required fields missed",
		request: refresh.Request{
			Id: 1,
		},
		expectedStatus: http.StatusBadRequest,
		expectedErr:    "field Email is not valid, field RefreshToken is required",
	},
	{name: "Negative Id",
		request: refresh.Request{
			Id:           -1,
			Email:        "somevalid@mail.com",
			RefreshToken: "InvalidRefreshToken",
		},
		expectedStatus: http.StatusBadRequest,
		expectedErr:    "field Id is not valid",
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

			var resp refresh.Response
			err = json.Unmarshal(rr.Body.Bytes(), &resp)
			if tc.expectedErr != "" {
				require.Equal(t, tc.expectedErr, resp.Error)
			} else {
				require.NoError(t, err)
				require.Empty(t, resp.Error)
				require.NotEmpty(t, resp.Token)
				require.NotEmpty(t, resp.RefreshToken)
				require.Equal(t, int64(1), resp.Id)
				storageMock.AssertExpectations(t)
			}
		})
	}

}
