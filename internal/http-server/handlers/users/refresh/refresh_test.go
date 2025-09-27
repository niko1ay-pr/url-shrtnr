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

var refreshCases = []struct {
	name                    string
	id                      int64
	email                   string
	refreshToken            string
	rtMock                  *models.RefreshToken
	expectedStatus          int
	shouldStopBeforeStoreRT bool
	expectedErr             string
	// mockUser           *models.User
	// mockErr            error
	// expectHashStatus   int
	// expectedRTStoreErr string
}{
	{
		name:         "Success refresh",
		id:           1,
		email:        "valid@mail.com",
		refreshToken: "ValidRefreshToken",
		rtMock: &models.RefreshToken{
			Id:    1,
			Token: "ValidRefreshToken",
		},
		expectedStatus: http.StatusOK,
	},
	{
		name:         "Refresh token not match",
		id:           1,
		email:        "valid@mail.com",
		refreshToken: "wrongRequestToken",
		rtMock: &models.RefreshToken{
			Id:    1,
			Token: "ValidRefreshToken",
		},
		shouldStopBeforeStoreRT: true,
		expectedStatus:          http.StatusBadRequest,
		expectedErr:             "failed to refresh token",
	},
}

func TestRefreshHandler(t *testing.T) {

	for _, tc := range refreshCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			storageMock := mocks.NewRefreshTokenGetStorer(t)
			storageMock.On("GetRefreshToken", mock.Anything, mock.AnythingOfType("int64")).
				Return(tc.rtMock, nil).Once()
			if !tc.shouldStopBeforeStoreRT {
				storageMock.On("StoreRefreshToken", mock.Anything, mock.AnythingOfType("*models.RefreshToken")).
					Return(nil).Once()
			}
			log := sldiscard.NewDescardLogger()
			secret := []byte("secret")
			handler := refresh.New(context.Background(), log, storageMock, secret)

			jsonBody := map[string]any{
				"id":            tc.id,
				"email":         tc.email,
				"refresh_token": tc.refreshToken,
			}
			request, err := json.Marshal(jsonBody)
			require.NoError(t, err)
			req, err := http.NewRequest("POST", "/users/refresh", bytes.NewReader(request))
			require.NoError(t, err)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			require.Equal(t, tc.expectedStatus, rr.Code)

			var resp refresh.Response

			err = json.Unmarshal(rr.Body.Bytes(), &resp)
			require.NoError(t, err)

			if tc.expectedErr == "" {
				require.NotEmpty(t, resp.Id)
				require.NotEmpty(t, resp.Token)
				require.NotEmpty(t, resp.RefreshToken)
			}
		})
	}

}
