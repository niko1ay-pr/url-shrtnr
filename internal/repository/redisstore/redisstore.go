package redisstore

import (
	"context"
	"fmt"
	"niko1ay-pr/url-shrtnr/internal/repository"

	"github.com/go-redis/redis/v8"
)

type Storage struct {
	db *redis.Client
}

func New(ctx context.Context) (*Storage, error) {

	const op = "repository.redissstore.New"

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

func (s *Storage) SaveURL(ctx context.Context, urlToSave string, alias string) error {
	const op = "repository.redisstore.SaveURL"

	_, err := s.db.Get(context.Background(), urlToSave).Result()
	if err == redis.Nil {
		return fmt.Errorf("%s: %v", op, repository.ErrUrlExists)
	} else if err != nil {
		return fmt.Errorf("%s: %v", op, err)
	}

	err = s.db.Set(context.Background(), urlToSave, alias, 0).Err()
	if err != nil {
		return fmt.Errorf("%s: %v", op, err)
	}

	return nil
}

func (s *Storage) GetURL(ctx context.Context, urlToGet string) (string, error) {
	const op = "repository.redisstore.GetURL"
	url, err := s.db.Get(context.Background(), urlToGet).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("%s: %v", op, repository.ErrUrlNotFound)
	} else if err != nil {
		return "", fmt.Errorf("%s: %v", op, err)
	}
	return url, nil
}
