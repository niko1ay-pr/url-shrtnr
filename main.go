package main

import (
	"context"
	"fmt"
	"net/http"
	"niko1ay-pr/url-shrtnr/utils"
	"time"

	redis "github.com/go-redis/redis/v8"
)

func main() {

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

	fmt.Println("Server running on port :8080")
	http.ListenAndServe(":8080", nil)

	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	//err := client.Set(ctx, "key", "value", 0).Err()
	//if err != nil {
	//panic(err)
	//}

	val, err := client.Get(ctx, "key").Result()

	if err != nil {
		panic(err)
	}
	fmt.Println(val) // Output: valuE

}
