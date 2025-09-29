package refresh

import (
	"context"
	"log/slog"
	"net/http"
	"niko1ay-pr/url-shrtnr/internal/lib/api/response"
	"niko1ay-pr/url-shrtnr/internal/lib/jwt"
	"niko1ay-pr/url-shrtnr/internal/lib/logger/sl"
	"niko1ay-pr/url-shrtnr/internal/lib/random"
	"niko1ay-pr/url-shrtnr/internal/models"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	Id           int64  `json:"id" validate:"required,gt=0"`
	Email        string `json:"email" validate:"email,required"`
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type Response struct {
	response.Response
	Id           int64  `json:"id" validate:"required"`
	Token        string `json:"token" validate:"required"`
	RefreshToken string `json:"refresh_token" validate:"required"`
}

//go:generate go run github.com/vektra/mockery/v2@latest --name=RefreshTokenGetStorer
type RefreshTokenGetStorer interface {
	RefreshTokenGetter
	RefreshTokenStorer
}

//go:generate go run github.com/vektra/mockery/v2@latest --name=RefreshTokenGetter
type RefreshTokenGetter interface {
	GetRefreshToken(context.Context, int64) (*models.RefreshToken, error)
}

//go:generate go run github.com/vektra/mockery/v2@latest --name=RefreshTokenStorer
type RefreshTokenStorer interface {
	StoreRefreshToken(context.Context, *models.RefreshToken) error
}

func New(ctx context.Context, log *slog.Logger, storage RefreshTokenGetStorer, secret []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.users.refresh.New"

		log = log.With(slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())))

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

		// get refresh token from db
		token, err := storage.GetRefreshToken(ctx, req.Id)
		if err != nil {
			log.Error("failed to get refresh token", sl.Err(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to refresh token"))
			return
		}
		// compare with one from request
		if token.Token != req.RefreshToken {
			log.Error("stored and posted tokens not matching")
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, response.Error("failed to refresh token"))
			return
		}
		// generate new JWT-token
		newJWT, err := jwt.GenerateToken(log, secret, req.Id, req.Email)
		if err != nil {
			log.Error("failed to generate token", sl.Err(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to refresh token"))
			return
		}

		// getting refresh-token and store it in Users table
		// generate new refresh token for next use
		newRefreshToken, err := random.NewRandomString(32)
		if err != nil {
			log.Error("failed to generate new refresh token", sl.Err(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to refresh token"))
			return
		}
		// store new refresh token in DB
		err = storage.StoreRefreshToken(ctx, &models.RefreshToken{Id: req.Id, Token: newRefreshToken})
		if err != nil {
			log.Error("failed to store new refresh token", sl.Err(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to refresh token"))
			return
		}

		// send response with new tokens
		log.Debug("token successfully refreshed", slog.Int64("id", req.Id), slog.String("email", req.Email))
		render.JSON(w, r, Response{
			Id:           req.Id,
			Token:        newJWT,
			RefreshToken: newRefreshToken,
		})
	}
}
