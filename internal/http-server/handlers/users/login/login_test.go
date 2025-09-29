package login_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"niko1ay-pr/url-shrtnr/internal/http-server/handlers/users/login"
	"niko1ay-pr/url-shrtnr/internal/http-server/handlers/users/login/mocks"
	"niko1ay-pr/url-shrtnr/internal/lib/hash"
	"niko1ay-pr/url-shrtnr/internal/lib/logger/sldiscard"
	"niko1ay-pr/url-shrtnr/internal/models"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockSetup func(*mocks.LoginStorage)

var testCases = []struct {
	name           string
	request        login.Request
	mockSetup      mockSetup
	expectedStatus int
	expectedErr    string
}{
	{
		name: "Success login",
		request: login.Request{
			Email:    "example@gmail.com",
			Password: "123passWord",
		},
		mockSetup: func(m *mocks.LoginStorage) {
			generatedHash, _ := hash.GenerateHash("123passWord", "123")
			m.On("GetUserHash", mock.Anything, mock.AnythingOfType("string")).
				Return(&models.User{
					Id:       1,
					Email:    "example@gmail.com",
					Password: string(generatedHash),
					Salt:     "123",
				}, nil).Once()
			m.On("StoreRefreshToken", mock.Anything, mock.AnythingOfType("*models.RefreshToken")).
				Return(nil).Once()
		},
		expectedStatus: http.StatusOK,
	},
	{
		name: "Wrong password",
		request: login.Request{
			Email:    "example@gmail.com",
			Password: "wrongpass",
		},
		mockSetup: func(m *mocks.LoginStorage) {
			m.On("GetUserHash", mock.Anything, mock.AnythingOfType("string")).
				Return(&models.User{
					Id:       1,
					Email:    "example@gmail.com",
					Password: "validPassword",
					Salt:     "123",
				}, nil).Once()
		},
		expectedStatus: http.StatusInternalServerError,
		expectedErr:    "failed to login user",
	},
	{
		name: "Empty password",
		request: login.Request{
			Email:    "valid@mail.com",
			Password: "",
		},
		expectedErr:    "field Password is required",
		expectedStatus: http.StatusBadRequest,
	},
	{
		name: "Invalid email",
		request: login.Request{
			Email:    "invalid-email",
			Password: "validpass",
		},
		expectedErr:    "field Email is not valid",
		expectedStatus: http.StatusBadRequest,
	},
}

func TestLoginHandler(t *testing.T) {

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			loginStorageMock := mocks.NewLoginStorage(t)
			if tc.mockSetup != nil {
				tc.mockSetup(loginStorageMock)
			}

			secret := []byte("test-secret")
			handler := login.New(context.Background(), sldiscard.NewDescardLogger(), loginStorageMock, secret)

			request, _ := json.Marshal(tc.request)
			req, err := http.NewRequest("POST", "/users/login", bytes.NewReader(request))
			require.NoError(t, err)

			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)
			require.Equal(t, tc.expectedStatus, rr.Code)

			var resp login.Response
			err = json.Unmarshal(rr.Body.Bytes(), &resp)
			require.NoError(t, err)

			if tc.expectedErr != "" {
				require.Equal(t, tc.expectedErr, resp.Error)
			} else {
				require.Empty(t, resp.Error)
				require.NotEmpty(t, resp.Token)
				require.NotEmpty(t, resp.RefreshToken)
			}

			loginStorageMock.AssertExpectations(t)
		})
	}
}
