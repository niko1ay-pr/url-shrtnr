package main

import (
	"fmt"
	"net/http"
	redisstore "niko1ay-pr/url-shrtnr/internal/storage/redisstore"
	"niko1ay-pr/url-shrtnr/utils"
	"os"
)

type Server struct {
	db Storage
}

func NewServer(db Storage) *Server {
	return &Server{db}
}

type Storage interface {
	Save(string) error
	Get(string) (string, error)
}

func main() {

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
