package register

import (
	"context"
	"log/slog"
	"net/http"
	"niko1ay-pr/url-shrtnr/internal/lib/api/response"
	"niko1ay-pr/url-shrtnr/internal/lib/hash"
	"niko1ay-pr/url-shrtnr/internal/lib/logger/sl"
	"niko1ay-pr/url-shrtnr/internal/models"
	"niko1ay-pr/url-shrtnr/internal/repository"

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
	Email string `json:"email"`
}

//go:generate go run github.com/vektra/mockery/v2@latest --name=UserSaver
type UserSaver interface {
	UserSave(context.Context, models.User) error
}

func New(ctx context.Context, log *slog.Logger, userSaver UserSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.auth.register.New"
		log = log.With(slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())))

		var req Request
		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, "failed decode request body")
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
		//add personal generated salt for each password
		salt, err := hash.RandomSalt(hash.SaltSize)
		if err != nil {
			log.Error("salt generation error", sl.Err(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to register"))
			return
		}

		hashedPass, err := hash.HashPassword(req.Password, salt)
		if err != nil {
			log.Error("hashing password error", sl.Err(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to register"))
			return
		}
		//store salt near password
		err = userSaver.UserSave(ctx, models.User{Email: req.Email,
			Password: hashedPass, Salt: salt})
		if err == repository.ErrUserExists {
			log.Info("already exists", sl.Err(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to register"))
			return
		}
		if err != nil {
			log.Error("failed to register", sl.Err(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.Error("failed to register"))
			return
		}

		log.Info("user registered", slog.String("email", req.Email))
		// TODO: return created status
		render.Status(r, http.StatusCreated)
		render.JSON(w, r, Response{Response: response.Ok(),
			Email: req.Email,
		})
	}
}
