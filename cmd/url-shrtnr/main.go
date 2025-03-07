package main

import (
	"context"
	"fmt"
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

	storage, err := sqlite.New(config.StoragePath)

	if err != nil {
		log.Error("Error storage init:", sl.Err(err))
		os.Exit(1)
	}

	router := chi.NewRouter()
	router.Use(mwlogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Route("/url", func(r chi.Router) {
		r.Use(middleware.BasicAuth("url-shrtnr", map[string]string{
			config.HTTPServer.User: config.HTTPServer.Password,
		}))
		r.Post("/", save.New(log, storage))
		r.Delete("/{alias}", delete.New(log, storage))
	})
	router.Get("/url/{alias}",
		redirect.New(log, storage))

	shutdown, done := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGKILL)
	defer done()

	s := http.Server{
		Handler:     router,
		Addr:        config.HTTPServer.Address,
		IdleTimeout: config.IdleTimeout,
	}

	ctx, close := context.WithTimeout(context.Background(), config.Timeout)
	defer close()
	go func() {
		err = s.ListenAndServe()
		fmt.Printf("ListenAndServe err: %v\n", err)
		if err != nil && err != http.ErrServerClosed {
			log.Error("Server error: ", sl.Err(err))
			os.Exit(1)
		}

	}()
	for {
		select {
		case <-shutdown.Done():
			fmt.Println("context.Done() invoked")
			err = s.Shutdown(ctx)
			if err != nil {
				log.Error("Shutdown err:", sl.Err(err))
			}
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
