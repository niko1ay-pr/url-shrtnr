package main

import (
	"context"
	"log/slog"
	"net/http"
	"niko1ay-pr/url-shrtnr/internal/config"
	"niko1ay-pr/url-shrtnr/internal/http-server/handlers/url/all"
	"niko1ay-pr/url-shrtnr/internal/http-server/handlers/url/delete"
	"niko1ay-pr/url-shrtnr/internal/http-server/handlers/url/redirect"
	"niko1ay-pr/url-shrtnr/internal/http-server/handlers/url/save"
	"niko1ay-pr/url-shrtnr/internal/http-server/handlers/url/update"
	"niko1ay-pr/url-shrtnr/internal/http-server/handlers/users/register"
	mwlogger "niko1ay-pr/url-shrtnr/internal/http-server/middleware/logger"
	sl "niko1ay-pr/url-shrtnr/internal/lib/logger/sl"
	"niko1ay-pr/url-shrtnr/internal/storage/sqlite"
	"os"
	"os/signal"
	"syscall"

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

	ctx, done := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGKILL)
	defer done()

	storage, err := sqlite.New(ctx, config.StoragePath)

	if err != nil {
		log.Error("Error storage init:", sl.Err(err))
		os.Exit(1)
	}

	router := chi.NewRouter()
	router.Use(mwlogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.RequestID)
	router.Use(middleware.URLFormat)

	router.Route("/url", func(r chi.Router) {
		r.Use(middleware.BasicAuth("url-shrtnr", map[string]string{
			config.HTTPServer.User: config.HTTPServer.Password,
		}))
		r.Delete("/{alias}", delete.New(ctx, log, storage))
		r.Post("/", save.New(ctx, log, storage))
		r.Put("/{alias}", update.New(ctx, log, storage))
	})
	router.Get("/urls", all.New(ctx, log, storage))
	router.Get("/url/{alias}",
		redirect.New(ctx, log, storage))
	router.Post("/users/register", register.New(ctx, log, storage))

	s := http.Server{
		Handler:     router,
		Addr:        config.HTTPServer.Address,
		IdleTimeout: config.IdleTimeout,
		ReadTimeout: config.Timeout,
	}

	go func() {
		err = s.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Error("Server error: ", sl.Err(err))
			os.Exit(1)
		}

	}()
	for {
		select {
		case <-ctx.Done():
			log.Info("Server is shutting down...")
			err = s.Shutdown(ctx)
			if err != nil {
				log.Error("Shutdown error: ", sl.Err(err))
			}
			log.Info("Server shut down gracefully")
			os.Exit(0)
		}
	}
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
