package redirect

import (
	"context"
	"log/slog"
	"net/http"
	"niko1ay-pr/url-shrtnr/internal/lib/api/response"
	"niko1ay-pr/url-shrtnr/internal/lib/logger/sl"
	"niko1ay-pr/url-shrtnr/internal/repository"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

//go:generate go run github.com/vektra/mockery/v2@latest --name=URLGetter
type URLGetter interface {
	GetURL(ctx context.Context, alias string) (string, error)
}

func New(ctx context.Context, log *slog.Logger, urlGetter URLGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.redirect.New"

		log = log.With(slog.String("op", op), slog.String("request_id", middleware.GetReqID(r.Context())))

		alias := chi.URLParam(r, "alias")
		if alias == "" {
			log.Info("alias is empty")
			render.JSON(w, r, response.Error("alias is empty"))
			return
		}

		url, err := urlGetter.GetURL(ctx, alias)
		if err == repository.ErrUrlNotFound {
			log.Info("ulr not found", slog.String("alias", alias))
			render.JSON(w, r, response.Error("url not found"))
			return
		}

		if err != nil {
			log.Error("failed to get url", sl.Err(err))
			render.JSON(w, r, response.Error("Internal error"))
			return
		}

		log.Info("url found. Redirecting.", slog.String("url", url),
			slog.String("alias", alias))

		http.Redirect(w, r, url, http.StatusFound)
	}
}
