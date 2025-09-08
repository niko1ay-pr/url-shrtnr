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
	"niko1ay-pr/url-shrtnr/internal/http-server/middleware/jwtauth"
	"niko1ay-pr/url-shrtnr/internal/lib/logger/sldiscard"
	"niko1ay-pr/url-shrtnr/internal/repository"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var cases = []struct {
	name       string
	entities   []save.Entity
	ownerID    int64
	respErr    string
	mockErr    error
	statusCode int
}{
	{
		name: "Success One Entity",
		entities: []save.Entity{
			{
				URL:   "https://google.com",
				Alias: "test_alias",
			},
		},
		ownerID:    1,
		statusCode: http.StatusOK,
	},
	{
		name: "Success Multiple Entities",
		entities: []save.Entity{
			{
				URL:   "https://google.com",
				Alias: "test_alias",
			},
			{
				URL:   "https://yandex.ru",
				Alias: "another_test",
			},
		},
		ownerID:    1,
		statusCode: http.StatusOK,
	},
	{
		name: "Success empty alias",
		entities: []save.Entity{
			{
				URL:   "https://google.com",
				Alias: "",
			},
		},
		ownerID:    1,
		statusCode: http.StatusOK,
	},
	{
		name: "Empty URL",
		entities: []save.Entity{
			{
				URL:   "",
				Alias: "test_empty_url",
			},
		},
		ownerID:    1,
		respErr:    "field URL is required",
		statusCode: http.StatusBadRequest,
	},
	{
		name: "Invalid URL",
		entities: []save.Entity{
			{
				URL:   "invalid.url",
				Alias: "invalid_url",
			},
		},
		ownerID:    1,
		respErr:    "field URL is not a valid URL",
		statusCode: http.StatusBadRequest,
	},
	{
		name: "SaveURL error",
		entities: []save.Entity{
			{
				URL:   "https://google.com",
				Alias: "some_valid_alias",
			},
		},
		ownerID:    1,
		respErr:    "failed to store url",
		mockErr:    errors.New("unexpected error"),
		statusCode: http.StatusInternalServerError,
	},
	{
		name: "Alias already exists",
		entities: []save.Entity{
			{
				URL:   "https://google.com",
				Alias: "any_alias",
			},
		},
		ownerID:    1,
		respErr:    "alias already exists",
		mockErr:    repository.ErrAliasExists,
		statusCode: http.StatusInternalServerError,
	},
}

func TestSaveHandler(t *testing.T) {
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			urlSaverMock := mocks.NewURLSaver(t)

			var entitiesInput string
			if tc.respErr == "" || tc.mockErr != nil {
				for _, ent := range tc.entities {
					urlSaverMock.On("SaveURL", context.Background(), ent.URL, mock.AnythingOfType("string"), tc.ownerID).
						Return(tc.mockErr).Once()
				}
			}

			handler := save.New(context.Background(), sldiscard.NewDescardLogger(), urlSaverMock)
			for n, ent := range tc.entities {
				entitiesInput = fmt.Sprintf(`%s{"url": "%s","alias":"%s"}`, entitiesInput, ent.URL, ent.Alias)
				if n < len(tc.entities)-1 {
					entitiesInput += ","
				}
			}
			input := fmt.Sprintf(`{"entities":[%s]}`, entitiesInput)
			req, err := http.NewRequest(http.MethodPost, "/url/save", strings.NewReader(input))
			ctx := jwtauth.ContextWithUserID(req.Context(), tc.ownerID)
			req = req.WithContext(ctx)
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.statusCode, rr.Code)

			var resp save.Response

			require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
			require.Equal(t, tc.respErr, resp.Error)
		})
	}

}
