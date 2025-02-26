package main

import (
	"log/slog"
	"net/http"
	"niko1ay-pr/url-shrtnr/internal/config"
	"niko1ay-pr/url-shrtnr/internal/http-server/handlers/url/delete"
	"niko1ay-pr/url-shrtnr/internal/http-server/handlers/url/redirect"
	"niko1ay-pr/url-shrtnr/internal/http-server/handlers/url/save"
	mwlogger "niko1ay-pr/url-shrtnr/internal/http-server/middleware/logger"
	sl "niko1ay-pr/url-shrtnr/internal/lib/logger/sl"
	"niko1ay-pr/url-shrtnr/internal/storage/sqlite"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/niko1ay-pr/prettyslog"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {

	config := config.MustLoad()

	log := setupLogger(config.Env)
	log.Info("Starting URL shortener service", slog.String("env", config.Env))
	log.Debug("Debug messages are shown in the logs")

	storage, err := sqlite.New(config.StoragePath)
	if err != nil {
		log.Error("Error storage init:", sl.Err(err))
		os.Exit(1)
	}

	_ = storage

	router := chi.NewRouter()
	router.Use(mwlogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Post("/url", save.New(log, storage))
	router.Get("/url/{alias}",
		redirect.New(log, storage))
	router.Delete("/url/{alias}", delete.New(log, storage))

	http.ListenAndServe(config.HttpServer.Address, router)
}

func setupLogger(env string) *slog.Logger {

	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(prettyslog.NewHandler(&slog.HandlerOptions{Level: slog.LevelDebug})) // use pretty log formatting for local environment
	case envDev:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return log
}
