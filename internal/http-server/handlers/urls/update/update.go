package update

import (
	"context"
	"log/slog"
	"net/http"
	"niko1ay-pr/url-shrtnr/internal/http-server/middleware/jwtauth"
	"niko1ay-pr/url-shrtnr/internal/lib/api/response"
	"niko1ay-pr/url-shrtnr/internal/lib/logger/sl"
	"niko1ay-pr/url-shrtnr/internal/repository"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	// NewAlias string `json:"new_alias,omitempty"`
	URL string `json:"url" validate:"required,url"`
}

type Response struct {
	response.Response
	Alias string `json:"alias"`
	URL   string `json:"url"`
}

//go:generate go run github.com/vektra/mockery/v2@latest --name=URLUpdater
type URLUpdater interface {
	UpdateURL(ctx context.Context, url string, alias string, ownerId int64) error
}

func New(ctx context.Context, log *slog.Logger, urlUpdater URLUpdater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.update.New"
		log = log.With(slog.String("op", op),
			slog.String("requiest_id", middleware.GetReqID(r.Context())))

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

		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			render.JSON(w, r, response.Error("failed to decode request body"))
			return
		}
		log.Debug("request body decoded", slog.Any("request", req))

		err = validator.New().Struct(req)
		if err != nil {
			log.Error("invalid request", sl.Err(err))
			validateErr := err.(validator.ValidationErrors)

			render.JSON(w, r, response.ValidationError(validateErr))
			return
		}

		// newAlias := req.NewAlias
		newURL := req.URL

		// if newAlias == "" {
		// 	newAlias, err = random.NewRandomString(repository.AliasLength)
		// 	if err != nil {
		// 		log.Error("failed to generate new alias", sl.Err(err))
		// 		render.JSON(w, r, response.Error("internal error"))
		// 		return
		// 	}
		// }

		err = urlUpdater.UpdateURL(ctx, alias, newURL, ownerID)
		if err == repository.ErrUrlNotFound {
			log.Info("url not found", slog.String("alias", alias))
			render.JSON(w, r, response.Error("url not found"))
			return
		}
		if err != nil {
			log.Error("failed to update url", sl.Err(err))
			render.JSON(w, r, response.Error("failed to update url"))
			return
		}

		log.Info("url updated", slog.String("old_alias", alias),
			slog.String("url", newURL),
			slog.String("alias", alias),
		)

		render.JSON(w, r, Response{
			Response: response.Ok(),
			Alias:    alias,
			URL:      newURL,
		})
	}
}
