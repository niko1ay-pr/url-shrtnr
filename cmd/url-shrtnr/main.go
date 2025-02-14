package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"niko1ay-pr/url-shrtnr/internal/config"
	redisstore "niko1ay-pr/url-shrtnr/internal/storage/redisstore"
	"niko1ay-pr/url-shrtnr/utils"
	"os"
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

	storage, err := redisstore.New()
	if err != nil {
		fmt.Println("Error creating Redis store:", err)
		os.Exit(1)
	}

	_ = storage

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
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envDev:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return log
}
