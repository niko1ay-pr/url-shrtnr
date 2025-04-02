package login

import (
	"context"
	"log/slog"
	"net/http"
	"niko1ay-pr/url-shrtnr/internal/lib/api/response"
	"niko1ay-pr/url-shrtnr/internal/lib/hash"
	"niko1ay-pr/url-shrtnr/internal/lib/logger/sl"
	"niko1ay-pr/url-shrtnr/internal/models"
	"strconv"

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
	Token string `json:"token"`
	Email string `json:"email"`
}

//go:generate go run github.com/vektra/mockery/v2@latest --name=UserHashGetter
type UserHashGetter interface {
	GetUserHash(context.Context, string) (*models.User, error)
}

func New(ctx context.Context, log *slog.Logger, hashGetter UserHashGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.users.login.New"
		log = log.With(slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())))

		var req Request
		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			render.JSON(w, r, response.Error("failed to decode request body"))
		}
		log.Debug("request body decoded", slog.Any("request", req))
		err = validator.New().Struct(req)
		if err != nil {
			log.Error("invalid request", sl.Err(err))
			validateErr := err.(validator.ValidationErrors)
			render.JSON(w, r, response.ValidationError(validateErr))
			return
		}

		user, err := hashGetter.GetUserHash(ctx, req.Email)
		if err != nil {
			log.Error("failed to login user", sl.Err(err))
			render.JSON(w, r, response.Error("failed to login user"))
			return
		}

		matches, err := hash.CheckPasswordHash(user.Password, req.Password+user.Salt)
		if !matches {
			if err != nil {
				log.Info("check password hash error", sl.Err(err))
			}
			render.JSON(w, r, response.Error("failed to login user"))
			return
		}

		log.Debug("password verified successfully")
		// TODO: if password accepted then generate new JWT-token

		token := strconv.Itoa(user.Id) // for sub claim
		log.Debug("user is logged in", slog.String("email", user.Email), slog.String("token", token))

		render.JSON(w, r, Response{
			Response: response.Ok(),
			Email:    req.Email,
			Token:    token,
		})
	}

}
