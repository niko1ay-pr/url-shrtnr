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
	"niko1ay-pr/url-shrtnr/internal/repository"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var loginCases = []struct {
	name         string
	email        string
	password     string
	expectedErr  string
	mockUser     *models.User
	mockErr      error
	expectStatus int
}{
	{
		name:     "Success login",
		email:    "example@gmail.com",
		password: "123passWord",
		mockUser: &models.User{
			Id:       1,
			Email:    "example@gmail.com",
			Password: "",
			Salt:     "123",
		},
		expectStatus: http.StatusOK,
	},
	{
		name:         "Invalid email",
		email:        "invalid-email",
		password:     "123passWord",
		expectedErr:  "field Email is not valid",
		expectStatus: http.StatusBadRequest,
	},
	{
		name:         "Empty password",
		email:        "example@gmail.com",
		expectedErr:  "field Password is required",
		expectStatus: http.StatusBadRequest,
	},
	{
		name:     "Wrong password",
		email:    "example@gmail.com",
		password: "wrongpass",
		mockUser: &models.User{
			Id:       1,
			Email:    "example@gmail.com",
			Password: "HashedPassword",
			Salt:     "anysalt",
		},
		expectStatus: http.StatusInternalServerError,
		expectedErr:  "failed to login user",
	},
	{
		name:         "User already exists",
		email:        "existinguser@example.com",
		password:     "123passWord",
		mockErr:      repository.ErrUserExists,
		expectedErr:  "failed to login user",
		expectStatus: http.StatusInternalServerError,
	},
}

func TestLoginHandler(t *testing.T) {

	for _, tc := range loginCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			userHashGetterMock := mocks.NewUserHashGetter(t)
			if tc.mockUser != nil && tc.mockErr == nil {
				if tc.mockUser.Password == "" {
					generatedHash, _ := hash.GenerateHash(tc.password, tc.mockUser.Salt)
					tc.mockUser.Password = string(generatedHash)
				}
				userHashGetterMock.On("GetUserHash", mock.Anything, tc.email).
					Return(tc.mockUser, nil).Once()
			} else if tc.mockErr != nil {
				userHashGetterMock.On("GetUserHash", mock.Anything, tc.email).
					Return(nil, tc.mockErr).Once()

			secret := []byte("test-secret")
			handler := login.New(context.Background(), sldiscard.NewDescardLogger(), userHashGetterMock, secret)

			requestBody := map[string]string{
				"email":    tc.email,
				"password": tc.password,
			}
			request, _ := json.Marshal(requestBody)
			req, err := http.NewRequest("POST", "/users/login", bytes.NewReader(request))
			require.NoError(t, err)

			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)
			require.Equal(t, tc.expectStatus, rr.Code)

			var resp login.Response
			err = json.Unmarshal(rr.Body.Bytes(), &resp)

			require.NoError(t, err)

			if tc.expectedErr != "" {

				require.Contains(t, resp.Error, tc.expectedErr)
			} else {
				require.NotEmpty(t, resp.Token)
				require.Empty(t, resp.Error)
				require.Equal(t, resp.Email, tc.email)
			}

			userHashGetterMock.AssertExpectations(t)
		})
	}
}
