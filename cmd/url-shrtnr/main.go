package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"niko1ay-pr/url-shrtnr/internal/config"
	mwlogger "niko1ay-pr/url-shrtnr/internal/http-server/middleware/logger"
	sl "niko1ay-pr/url-shrtnr/internal/lib/logger/sl"
	"niko1ay-pr/url-shrtnr/internal/storage/sqlite"
	"niko1ay-pr/url-shrtnr/utils"
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

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("It's URL shortener"))
	})

	http.HandleFunc("/shorten", func(w http.ResponseWriter, r *http.Request) {
		// Get user's URL from request via FormValue
		url := r.FormValue("url")
		fmt.Println("Payload: ", url)

		shortURL := utils.GetShortCode()
		fullShortenURL := fmt.Sprintf("http://localhost:8080/r/%s", shortURL)
		fmt.Println("Generated shortened URL: ", fullShortenURL)
	})

	fmt.Println("Server running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)

}

func setupLogger(env string) *slog.Logger {

	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(prettyslog.NewHandler(&slog.HandlerOptions{Level: slog.LevelDebug}))
	case envDev:
		log = slog.New(prettyslog.NewHandler(&slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		log = slog.New(prettyslog.NewHandler(&slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return log
}
