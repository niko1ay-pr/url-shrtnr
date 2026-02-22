package delete

import (
	"context"
	"log/slog"
	"net/http"
	"niko1ay-pr/url-shrtnr/internal/http-server/middleware/jwtauth"
	"niko1ay-pr/url-shrtnr/internal/lib/api/response"
	"niko1ay-pr/url-shrtnr/internal/lib/logger/sl"
	"niko1ay-pr/url-shrtnr/internal/repository"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type Response struct {
	response.Response
}

//go:generate go run github.com/vektra/mockery/v2@latest --name=URLDeleter
type URLDeleter interface {
	DeleteURL(context.Context, string, int64) error
}

func New(log *slog.Logger, urlDeleter URLDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.delete.New"

		log = log.With(slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())))

		ownerID := jwtauth.GetUserID(r.Context())
		if ownerID == 0 {
			log.Error("owner ID cannot be empty")
			render.JSON(w, r, response.Error("failed to decode request body"))
			return
		}

		alias := chi.URLParam(r, "alias")
		if alias == "" {
			log.Info("not found", sl.Err(repository.ErrUrlNotFound))
			render.JSON(w, r, response.Error("url not found"))

			return
		}

		log.Info("received alias", slog.String("alias", alias))

		err := urlDeleter.DeleteURL(r.Context(), alias, ownerID)

		if err == repository.ErrUrlNotFound {
			log.Info("url not found", slog.String("alias", alias))
			render.JSON(w, r, response.Error("url not found"))

			return
		}
		if err != nil {
			log.Error("failed to delete url", slog.String("alias", alias), sl.Err(err))
			render.JSON(w, r, response.Error("internal error"))

			return
		}
		log.Info("url deleted", slog.String("alias", alias))
		render.JSON(w, r, Response{Response: response.Ok()})
	}
}
