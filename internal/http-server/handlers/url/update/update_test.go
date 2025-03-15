package update_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"niko1ay-pr/url-shrtnr/internal/http-server/handlers/url/update"
	"niko1ay-pr/url-shrtnr/internal/http-server/handlers/url/update/mocks"
	"niko1ay-pr/url-shrtnr/internal/lib/logger/sldiscard"
	"niko1ay-pr/url-shrtnr/internal/storage"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var cases = []struct {
	name     string
	alias    string
	newAlias string
	url      string
	respErr  string
	mockErr  error
}{
	{
		name:     "Success",
		alias:    "test_alias",
		newAlias: "newalias",
		url:      "https://google.com",
	},
	{
		name:    "Empty alias",
		alias:   "",
		respErr: "url not found",
	},
	{
		name:     "Empty new alias",
		alias:    "test_alias",
		newAlias: "",
		url:      "https://google.com",
	},
	{
		name:    "Empty URL",
		alias:   "test_empty_url",
		url:     "",
		respErr: "field URL is required",
	},
	{
		name:    "Invalid URL",
		alias:   "invalid_ulr",
		url:     "invalid.url",
		respErr: "field URL is not a valid URL",
	},
	{
		name:    "UpdateURL error",
		alias:   "some_valid_alias",
		url:     "https://google.com",
		respErr: "failed to update url",
		mockErr: errors.New("unexpected error"),
	},
	{
		name:     "Url not found",
		alias:    "any_alias",
		newAlias: "any_another_alias",
		url:      "https://google.com",
		respErr:  "url not found",
		mockErr:  storage.ErrUrlNotFound,
	},
}

func TestUpdateHandler(t *testing.T) {
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			urlUpdaterMock := mocks.NewURLUpdater(t)

			if tc.respErr == "" || tc.mockErr != nil {
				urlUpdaterMock.On("UpdateURL", context.Background(), tc.url, mock.AnythingOfType("string"), mock.AnythingOfType("string")).
					Return(tc.mockErr).Once()
			}

			handler := update.New(context.Background(), sldiscard.NewDescardLogger(), urlUpdaterMock)

			input := fmt.Sprintf(`{"url": "%s", "new_alias": "%s"}`, tc.url, tc.newAlias)

			req, err := http.NewRequest(http.MethodPut, "/url/{alias}", strings.NewReader(input))
			require.NoError(t, err)

			r := chi.NewRouteContext()
			r.URLParams.Add("alias", tc.alias)

			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, r))

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			require.Equal(t, http.StatusOK, rr.Code)

			var resp update.Response

			require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))

			require.Equal(t, tc.respErr, resp.Error)
		})
	}

}
