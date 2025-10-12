package redisstore

import (
	"context"
	"fmt"
	"niko1ay-pr/url-shrtnr/internal/repository"
	"time"

	"github.com/go-redis/redis/v8"
)

type Storage struct {
	db  *redis.Client
	ttl time.Duration
}

func New(ctx context.Context, addr string, password string, dbID int, ttl time.Duration) (*Storage, error) {

	const op = "repository.redissstore.New"

	db := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       dbID,
	})
	err := db.Ping(context.Background()).Err()
	if err != nil {
		return nil, fmt.Errorf("%s: %v", op, err)
	}
	return &Storage{db: db, ttl: ttl}, nil
}

func (s *Storage) SaveURL(ctx context.Context, urlToSave string, alias string) error {
	const op = "repository.redisstore.SaveURL"

	_, err := s.db.Get(ctx, urlToSave).Result()
	if err != nil && err != redis.Nil {
		return fmt.Errorf("%s: %v", op, err)
	}

	err = s.db.Set(ctx, alias, urlToSave, s.ttl).Err()
	if err != nil {
		return fmt.Errorf("%s: %v", op, err)
	}

	return nil
}

func (s *Storage) GetURL(ctx context.Context, alias string) (string, error) {
	const op = "repository.redisstore.GetURL"

	url, err := s.db.Get(ctx, alias).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("%s: %v", op, repository.ErrUrlNotFound)
	} else if err != nil {
		return "", fmt.Errorf("%s: %v", op, err)
	}
	return url, nil
}

func (s *Storage) DeleteURL(ctx context.Context, alias string) error {

	err := s.db.Del(ctx, alias).Err()
	if err != nil {
		return err
	}
	return nil
}
func (s *Storage) UpdateURL(ctx context.Context, url string, alias string) error {
	err := s.db.Set(ctx, alias, url, s.ttl).Err()
	return err
}
