package delete_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	deleteurl "niko1ay-pr/url-shrtnr/internal/http-server/handlers/urls/delete"
	"niko1ay-pr/url-shrtnr/internal/http-server/handlers/urls/delete/mocks"
	"niko1ay-pr/url-shrtnr/internal/http-server/middleware/jwtauth"
	"niko1ay-pr/url-shrtnr/internal/lib/logger/sldiscard"

	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var cases = []struct {
	name    string
	ownerID int64
	alias   string
	respErr string
	mockErr error
}{
	{
		name:    "Success",
		ownerID: 1,
		alias:   "validalias",
	},
	{
		name:    "Empty owner ID",
		ownerID: 0,
		alias:   "validalias",
		respErr: "failed to decode request body",
	},
	{
		name:    "Empty alias",
		ownerID: 1,
		alias:   "",
		respErr: "url not found",
	},
	{
		name:    "Internal error",
		ownerID: 1,
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
				urlDeleterMock.On("DeleteURL", mock.MatchedBy(func(ctx context.Context) bool {
					userId := jwtauth.GetUserID(ctx)
					return userId == tc.ownerID
				}), tc.alias, tc.ownerID).Return(tc.mockErr).Once()
			}

			handler := deleteurl.New(sldiscard.NewDescardLogger(), urlDeleterMock)

			req, err := http.NewRequest(http.MethodDelete, "/url/{alias}", nil)
			require.NoError(t, err)

			ctxWithID := jwtauth.ContextWithUserID(req.Context(), tc.ownerID)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("alias", tc.alias)

			req = req.WithContext(context.WithValue(ctxWithID, chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()
			handler(rr, req)

			require.Equal(t, http.StatusOK, rr.Code)

			var resp deleteurl.Response
			require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))

			require.Equal(t, tc.respErr, resp.Error)

		})
	}
}
