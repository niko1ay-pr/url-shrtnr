package delete_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	deleteurl "niko1ay-pr/url-shrtnr/internal/http-server/handlers/urls/delete"
	"niko1ay-pr/url-shrtnr/internal/http-server/handlers/urls/delete/mocks"
	"niko1ay-pr/url-shrtnr/internal/lib/logger/sldiscard"

	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

var cases = []struct {
	name    string
	alias   string
	respErr string
	mockErr error
}{
	{
		name:  "Success",
		alias: "validalias",
	},
	{
		name:    "Empty alias",
		alias:   "",
		respErr: "url not found",
	},
	{
		name:    "Internal error",
		alias:   "validalias",
		respErr: "internal error",
		mockErr: errors.New(""),
	},
}

func TestDeleteHandler(t *testing.T) {

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			urlDeleterMock := mocks.NewURLDeleter(t)

			if tc.respErr == "" || tc.mockErr != nil {
				urlDeleterMock.On("DeleteURL", context.Background(), tc.alias).Return(tc.mockErr).Once()
			}

			handler := deleteurl.New(context.Background(), sldiscard.NewDescardLogger(), urlDeleterMock)

			req, err := http.NewRequest(http.MethodDelete, "/url/{alias}", nil)
			require.NoError(t, err)

			r := chi.NewRouteContext()
			r.URLParams.Add("alias", tc.alias)

			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, r))

			rr := httptest.NewRecorder()
			handler(rr, req)

			require.Equal(t, http.StatusOK, rr.Code)

			var resp deleteurl.Response
			require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))

			require.Equal(t, tc.respErr, resp.Error)

		})
	}
}
