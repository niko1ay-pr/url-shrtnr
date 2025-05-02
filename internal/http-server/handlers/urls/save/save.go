package save

import (
	"context"
	"log/slog"
	"net/http"
	"niko1ay-pr/url-shrtnr/internal/lib/api/response"
	"niko1ay-pr/url-shrtnr/internal/lib/logger/sl"
	"niko1ay-pr/url-shrtnr/internal/lib/random"
	"niko1ay-pr/url-shrtnr/internal/repository"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

//go:generate go run github.com/vektra/mockery/v2@latest --name=URLSaver
type URLSaver interface {
	SaveURL(ctx context.Context, url string, alias string, ownerId int) error
}

type Entity struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}
type Request struct {
	Entities []Entity `json:"entities"`
}

type Response struct {
	response.Response
	Entities []Entity `json:"entities"`
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

		var resp Response
		for _, entity := range req.Entities {
			alias := entity.Alias
			if alias == "" {
				// TODO: check if genereted alias is already exists
				alias, err = random.NewRandomString(repository.AliasLength)
				if err != nil {
					log.Error("failed to generate new alias", sl.Err(err))
					render.JSON(w, r, response.Error("failed to generate new alias"))
					return
				}
			}
			// TODO: owner_id receive from header (bearer). Decode from JWT-token
			ownerId := 0
			err = urlSaver.SaveURL(ctx, entity.URL, alias, ownerId)
			if err == repository.ErrUrlExists {
				log.Info("url already exists", slog.String("url", entity.URL))
				render.JSON(w, r, response.Error("url already exists"))
				return

			}
			if err != nil {
				log.Error("failed to store url", sl.Err(err))
				render.JSON(w, r, response.Error("failed to store url"))
				return
			}
			resp.Entities = append(resp.Entities, Entity{Alias: alias, URL: entity.URL})

			log.Info("url added", slog.String("url", entity.URL), slog.String("alias", alias))
		}

		resp.Response = response.Ok()
		render.JSON(w, r, resp)
	}

}
