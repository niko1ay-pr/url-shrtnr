package all

import (
	"context"
	"log/slog"
	"net/http"
	"niko1ay-pr/url-shrtnr/internal/lib/api/response"
	"niko1ay-pr/url-shrtnr/internal/lib/logger/sl"
	"niko1ay-pr/url-shrtnr/internal/storage"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type Response struct {
	response.Response
	Entities []storage.UrlEntity
}

type URLGetterAll interface {
	GetAll(context.Context) ([]storage.UrlEntity, error)
}

func New(ctx context.Context, log *slog.Logger, urlGetterAll URLGetterAll) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.all.New"
		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())))

		entities, err := urlGetterAll.GetAll(ctx)
		if err != nil {
			log.Error("failed to get all urls", sl.Err(err))
			render.JSON(w, r, response.Error("internal error"))
			return
		}

		if len(entities) == 0 {
			log.Debug("list of urls is empty")
		}

		render.JSON(w, r, Response{
			Response: response.Ok(),
			Entities: entities,
		})
	}
}
