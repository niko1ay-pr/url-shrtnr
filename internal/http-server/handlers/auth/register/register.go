package register

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

type Request struct {
	Email    string `json:"email" validate:"email,required"`
	Password string `json:"password" validate:"required"`
}

type Response struct {
	response.Response
	Email string `json:"email"`
}

//go:generate go run github.com/vektra/mockery/v2@latest --name=AuthRegisterer
type UserSaver interface {
	UserSave(context.Context, string, string) error
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
			render.JSON(w, r, "failed decode request body")
			return
		}
		err = userSaver.UserSave(ctx, req.Email, req.Password)
		if err == storage.ErrUrlExists {
			log.Info("already exists", sl.Err(err))
			render.JSON(w, r, response.Error("failed to register"))
			return
		}
		if err != nil {
			log.Error("failed to register", sl.Err(err))
			render.JSON(w, r, response.Error("failed to register"))
			return
		}
		render.JSON(w, r, Response{Response: response.Ok(),
			Email: req.Email,
		})
	}
}
