package save_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"niko1ay-pr/url-shrtnr/internal/http-server/handlers/urls/save"
	"niko1ay-pr/url-shrtnr/internal/http-server/handlers/urls/save/mocks"
	"niko1ay-pr/url-shrtnr/internal/lib/logger/sldiscard"
	"niko1ay-pr/url-shrtnr/internal/repository"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var cases = []struct {
	name    string
	alias   string
	url     string
	respErr string
	mockErr error
}{
	{
		name:  "Success",
		alias: "test_alias",
		url:   "https://google.com",
	},
	{
		name:  "Empty alias",
		alias: "",
		url:   "https://google.com",
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
		name:    "SaveURL error",
		alias:   "some_valid_alias",
		url:     "https://google.com",
		respErr: "failed to store url",
		mockErr: errors.New("unexpected error"),
	},
	{
		name:    "Url already exists",
		alias:   "any_alias",
		url:     "https://google.com",
		respErr: "url already exists",
		mockErr: repository.ErrUrlExists,
	},
}

func TestSaveHandler(t *testing.T) {
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			urlSaverMock := mocks.NewURLSaver(t)

			if tc.respErr == "" || tc.mockErr != nil {
				urlSaverMock.On("SaveURL", context.Background(), tc.url, mock.AnythingOfType("string")).
					Return(tc.mockErr).Once()
			}

			handler := save.New(context.Background(), sldiscard.NewDescardLogger(), urlSaverMock)

			input := fmt.Sprintf(`{"url": "%s", "alias": "%s"}`, tc.url, tc.alias)

			req, err := http.NewRequest(http.MethodPost, "/save", strings.NewReader(input))
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			require.Equal(t, http.StatusOK, rr.Code)

			var resp save.Response

			require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))

			require.Equal(t, tc.respErr, resp.Error)
		})
	}

}
