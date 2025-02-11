package redisstore

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
)

type Storage struct {
	db *redis.Client
}

func New() (*Storage, error) {

	const op = "storage.redisstore.New"

	db := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})
	err := db.Ping(context.Background()).Err()
	if err != nil {
		return nil, fmt.Errorf("%s: %v", op, err)
	}
	return &Storage{db: db}, nil
}
