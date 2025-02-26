package delete

import (
	"log/slog"
	"net/http"
	"niko1ay-pr/url-shrtnr/internal/lib/api/response"
	"niko1ay-pr/url-shrtnr/internal/lib/logger/sl"
	"niko1ay-pr/url-shrtnr/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type Request struct {
}

type Response struct {
}

type UrlDeleter interface {
	DeleteURL(string) error
}

func New(log *slog.Logger, urlDeleter UrlDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.delete.New"

		log.With(slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())))

		alias := chi.URLParam(r, "alias")
		if alias == "" {
			log.Info("not found", sl.Err(storage.ErrUrlNotFound))
			render.JSON(w, r, "url not found")

			return
		}

		log.Info("received alias", slog.String("alias", alias))

		// TODO: delete url from storage

		render.JSON(w, r, response.Ok())
	}
}
