package save

import (
	"context"
	"log/slog"
	"net/http"
	"niko1ay-pr/url-shrtnr/internal/http-server/middleware/jwtauth"
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
	SaveURL(ctx context.Context, url string, alias string, ownerId int64) error
}

type Entity struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}
type Request struct {
	Entities []Entity `json:"entities" validate:"required,dive"`
}

type Response struct {
	response.Response
	Entities []Entity `json:"entities"`
}

func New(log *slog.Logger, urlSaver URLSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.save.New"
		const maxGeneratingRetries = 3
		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		ownerID := jwtauth.GetUserID(r.Context())
		if ownerID == 0 {
			log.Error("owner ID cannot be empty")
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("failed to decode request body"))
			return
		}
		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to decode request body"))
			return
		}
		log.Debug("request body decoded", slog.Any("request", req))

		err = validator.New().Struct(req)
		if err != nil {
			log.Error("invalid request", sl.Err(err))
			validateErr := err.(validator.ValidationErrors)
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.ValidationError(validateErr))
			return
		}

		var resp Response
		for _, entity := range req.Entities {
			alias := entity.Alias
			newAlias := ""
			retries := 1
			if alias == "" {
				retries = maxGeneratingRetries
			} else {
				newAlias = alias
			}

			for i := range retries {
				if alias == "" {
					// generate alias if not presented
					newAlias, err = random.NewRandomString(repository.AliasLength)
					if err != nil {
						log.Error("failed to generate new alias", sl.Err(err))
						render.Status(r, http.StatusInternalServerError)
						render.JSON(w, r, response.Error("failed to generate new alias"))
						return
					}
				}
				err = urlSaver.SaveURL(r.Context(), entity.URL, newAlias, ownerID)
				if err == repository.ErrAliasExists {
					// while tries left keep generating alias
					if i < retries-1 {
						log.Info("alias already exists. Trying to generate new one", slog.String("alias", newAlias))
						continue
					}
					log.Info("alias already exists", slog.String("alias", newAlias))
					render.Status(r, http.StatusInternalServerError)
					render.JSON(w, r, response.Error("alias already exists"))
					return
				}
				if err != nil {
					log.Error("failed to store url", sl.Err(err))
					render.Status(r, http.StatusInternalServerError)
					render.JSON(w, r, response.Error("failed to store url"))
					return
				}
				break
			}
			resp.Entities = append(resp.Entities, Entity{Alias: newAlias, URL: entity.URL})

			log.Info("url added", slog.String("url", entity.URL), slog.String("alias", newAlias))
		}

		resp.Response = response.Ok()
		render.Status(r, http.StatusOK)
		render.JSON(w, r, resp)
	}

}
