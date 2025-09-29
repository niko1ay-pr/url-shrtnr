package register_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"niko1ay-pr/url-shrtnr/internal/http-server/handlers/users/register"
	"niko1ay-pr/url-shrtnr/internal/http-server/handlers/users/register/mocks"
	"niko1ay-pr/url-shrtnr/internal/lib/logger/sldiscard"
	"niko1ay-pr/url-shrtnr/internal/repository"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var registerCases = []struct {
	name           string
	requireCall    bool
	email          string
	password       string
	expectedErr    string
	mockErr        error
	expectedStatus int
}{
	{
		name:           "Success registration",
		requireCall:    true,
		email:          "valid@example.com",
		password:       "validPassword123",
		expectedStatus: http.StatusCreated,
	},
	{
		name:           "Error user exists",
		requireCall:    true,
		email:          "valid@example.com",
		password:       "validPass123",
		expectedStatus: http.StatusInternalServerError,
		expectedErr:    "failed to register",
		mockErr:        repository.ErrUserExists,
	},
	{
		name:           "Empty password",
		email:          "valid@example.com",
		password:       "",
		expectedErr:    "field Password is required",
		expectedStatus: http.StatusBadRequest,
	},
	{
		name:           "Invalid email",
		email:          "invalid-email",
		password:       "someValidPassword123",
		expectedErr:    "field Email is not valid",
		expectedStatus: http.StatusBadRequest,
	},
}

func TestRegisterHandler(t *testing.T) {
	for _, tc := range registerCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			userSaverMock := mocks.NewUserSaver(t)
			if tc.requireCall {
				userSaverMock.On("UserSave", mock.Anything, mock.Anything).
					Return(tc.mockErr).Once()
			}

			handler := register.New(context.Background(), sldiscard.NewDescardLogger(), userSaverMock)
			requestBody := map[string]string{
				"email":    tc.email,
				"password": tc.password,
			}

			request, _ := json.Marshal(requestBody)
			req, err := http.NewRequest("POST", "/users/register", bytes.NewReader(request))
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			require.Equal(t, tc.expectedStatus, rr.Code)

			var resp register.Response
			err = json.Unmarshal(rr.Body.Bytes(), &resp)
			require.NoError(t, err)

			if tc.expectedErr != "" {
				require.Contains(t, resp.Error, tc.expectedErr)
			} else {
				require.Equal(t, resp.Email, tc.email)
				require.Empty(t, resp.Error)
			}

		})

	}

}
