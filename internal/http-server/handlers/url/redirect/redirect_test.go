package redirect_test

import (
	"context"
	"net/http/httptest"
	"niko1ay-pr/url-shrtnr/internal/http-server/handlers/url/redirect"
	"niko1ay-pr/url-shrtnr/internal/http-server/handlers/url/redirect/mocks"
	"niko1ay-pr/url-shrtnr/internal/lib/api"
	"niko1ay-pr/url-shrtnr/internal/lib/logger/sldiscard"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
	"gopkg.in/go-playground/assert.v1"
)

func TestRedirectHandler(t *testing.T) {
	cases := []struct {
		name    string
		alias   string
		url     string
		respErr string
		mockErr error
	}{
		{
			name:  "Success",
			alias: "somealias",
			url:   "https://google.com",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			urlGetterMock := mocks.NewURLGetter(t)

			if tc.respErr == "" || tc.mockErr != nil {
				urlGetterMock.On("GetURL", context.Background(), tc.alias).
					Return(tc.url, tc.mockErr).Once()
			}

			handler := redirect.New(context.Background(), sldiscard.NewDescardLogger(), urlGetterMock)

			r := chi.NewRouter()
			r.Get("/url/{alias}", handler)

			ts := httptest.NewServer(r)
			defer ts.Close()

			redirectToURL, err := api.GetRedirect(ts.URL + "/url/" + tc.alias)

			require.NoError(t, err)

			assert.Equal(t, tc.url, redirectToURL)
		})
	}
}
