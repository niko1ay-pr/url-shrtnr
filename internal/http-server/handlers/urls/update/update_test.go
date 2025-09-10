package update_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"niko1ay-pr/url-shrtnr/internal/http-server/handlers/urls/update"
	"niko1ay-pr/url-shrtnr/internal/http-server/handlers/urls/update/mocks"
	"niko1ay-pr/url-shrtnr/internal/http-server/middleware/jwtauth"
	"niko1ay-pr/url-shrtnr/internal/lib/logger/sldiscard"
	"niko1ay-pr/url-shrtnr/internal/repository"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var cases = []struct {
	name       string
	ownerID    int64
	alias      string
	newURL     string
	respErr    string
	mockErr    error
	statusCode int
}{
	{
		name:       "Success",
		ownerID:    1,
		alias:      "test_alias",
		newURL:     "https://google.com",
		statusCode: http.StatusOK,
	},
	{
		name:       "Empty ownerID",
		ownerID:    0,
		alias:      "valid_alias",
		newURL:     "https://google.com",
		respErr:    "failed to decode request body",
		statusCode: http.StatusBadRequest,
	},
	{
		name:       "Empty alias",
		ownerID:    1,
		alias:      "",
		respErr:    "url not found",
		statusCode: http.StatusBadRequest,
	},

	{
		name:       "Empty URL",
		ownerID:    1,
		alias:      "test_empty_url",
		newURL:     "",
		respErr:    "field URL is required",
		statusCode: http.StatusBadRequest,
	},
	{
		name:       "Invalid URL",
		ownerID:    1,
		alias:      "invalid_url",
		newURL:     "invalid.url",
		respErr:    "field URL is not a valid URL",
		statusCode: http.StatusBadRequest,
	},
	{
		name:       "UpdateURL error",
		ownerID:    1,
		alias:      "some_valid_alias",
		newURL:     "https://google.com",
		respErr:    "failed to update url",
		mockErr:    errors.New("unexpected error"),
		statusCode: http.StatusInternalServerError,
	},
	{
		name:       "Alias not found",
		ownerID:    1,
		alias:      "any_alias",
		newURL:     "https://google.com",
		respErr:    "alias not found",
		mockErr:    repository.ErrAliasNotFound,
		statusCode: http.StatusInternalServerError,
	},
	//{
	//TODO: trying to update foreign url
	//},
}

func TestUpdateHandler(t *testing.T) {
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			urlUpdaterMock := mocks.NewURLUpdater(t)

			if tc.respErr == "" || tc.mockErr != nil {
				urlUpdaterMock.On("UpdateURL", context.Background(),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("int64")).
					Return(tc.mockErr).Once()
			}

			handler := update.New(context.Background(), sldiscard.NewDescardLogger(), urlUpdaterMock)

			input := fmt.Sprintf(`{"url": "%s"}`, tc.newURL)

			req, err := http.NewRequest(http.MethodPut, "/url/{alias}", strings.NewReader(input))
			require.NoError(t, err)

			ctxWithID := jwtauth.ContextWithUserID(req.Context(), tc.ownerID)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("alias", tc.alias)

			req = req.WithContext(context.WithValue(ctxWithID, chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.statusCode, rr.Code)

			var resp update.Response

			require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))

			if tc.respErr != "" {
				require.Equal(t, tc.respErr, resp.Error)
			} else {
				require.Equal(t, tc.alias, resp.Alias)
			}
		})
	}

}
