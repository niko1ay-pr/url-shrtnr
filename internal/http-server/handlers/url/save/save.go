package save

import (
	"context"
	"log/slog"
	"net/http"
	"niko1ay-pr/url-shrtnr/internal/lib/api/response"
	"niko1ay-pr/url-shrtnr/internal/lib/logger/sl"
	"niko1ay-pr/url-shrtnr/internal/lib/random"
	"niko1ay-pr/url-shrtnr/internal/storage"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

// TODO: move to config

//go:generate go run github.com/vektra/mockery/v2@latest --name=URLSaver
type URLSaver interface {
	SaveURL(ctx context.Context, url string, alias string) error
}

type Request struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

type Response struct {
	response.Response
	Alias string `json:"alias,omitempty"`
}

func New(ctx context.Context, log *slog.Logger, urlSaver URLSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.save.New"
		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

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
		alias := req.Alias
		if alias == "" {
			// TODO: check if genereted alias is already exists
			alias, err = random.NewRandomString(storage.AliasLength)
			if err != nil {
				log.Error("failed to generate new alias", sl.Err(err))
				render.JSON(w, r, response.Error("failed to generate new alias"))
				return
			}
		}

		err = urlSaver.SaveURL(ctx, req.URL, alias)
		if err == storage.ErrUrlExists {
			log.Info("url already exists", slog.String("url", req.URL))
			render.JSON(w, r, response.Error("url already exists"))
			return

		}
		if err != nil {
			log.Error("failed to store url", sl.Err(err))
			render.JSON(w, r, response.Error("failed to store url"))
			return
		}

		log.Info("url added", slog.String("url", req.URL), slog.String("alias", alias))

		render.JSON(w, r, Response{
			Response: response.Ok(),
			Alias:    alias,
		})
	}

}
