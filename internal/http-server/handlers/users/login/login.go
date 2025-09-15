package login

import (
	"context"
	"log/slog"
	"net/http"
	"niko1ay-pr/url-shrtnr/internal/lib/api/response"
	"niko1ay-pr/url-shrtnr/internal/lib/hash"
	"niko1ay-pr/url-shrtnr/internal/lib/jwt"
	"niko1ay-pr/url-shrtnr/internal/lib/logger/sl"
	"niko1ay-pr/url-shrtnr/internal/lib/random"
	"niko1ay-pr/url-shrtnr/internal/models"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	Email    string `json:"email" validate:"email,required"`
	Password string `json:"password" validate:"required"`
}

type Response struct {
	response.Response
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	Email        string `json:"email"`
}

//go:generate go run github.com/vektra/mockery/v2@latest --name=LoginStorage
type LoginStorage interface {
	UserHashGetter
	RefreshTokenStorer
}

//go:generate go run github.com/vektra/mockery/v2@latest --name=UserHashGetter
type UserHashGetter interface {
	GetUserHash(context.Context, string) (*models.User, error)
}

//go:generate go run github.com/vektra/mockery/v2@latest --name=RefreshTokenStorer
type RefreshTokenStorer interface {
	StoreRefreshToken(context.Context, *models.RefreshToken) error
}

func New(ctx context.Context, log *slog.Logger, loginStorage LoginStorage, secret []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.users.login.New"
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

		user, err := loginStorage.GetUserHash(ctx, req.Email)
		if err != nil {
			log.Error("failed to login user", sl.Err(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to login user"))
			return
		}

		matches, err := hash.CheckPasswordHash(user.Password, req.Password+user.Salt)
		if !matches {
			if err != nil {
				log.Info("check password hash error", sl.Err(err))
			}
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to login user"))
			return
		}

		log.Debug("password verified successfully")

		signedString, err := jwt.GenerateToken(log, secret, user.Id, user.Email)
		if err != nil {
			log.Error("failed to login user", sl.Err(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to login user"))
			return
		}

		// Getting refresh-token and store it in Users table
		rtString, err := random.NewRandomString(32)
		if err != nil {
			log.Error("failed to login user", sl.Err(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to get refresh token"))
			return
		}
		refreshToken := &models.RefreshToken{
			Id:    user.Id,
			Token: rtString,
		}
		err = loginStorage.StoreRefreshToken(ctx, refreshToken)
		if err != nil {
			log.Error("failed to login user", sl.Err(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to get refresh token"))
			return
		}

		log.Debug("user is logged in", slog.String("email", user.Email), slog.String("token", signedString))

		render.JSON(w, r, Response{
			Response:     response.Ok(),
			Email:        req.Email,
			RefreshToken: refreshToken.Token,
			Token:        signedString,
		})
	}

}
