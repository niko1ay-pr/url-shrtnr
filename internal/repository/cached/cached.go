package cached

import (
	"context"
	"niko1ay-pr/url-shrtnr/internal/http-server/handlers/urls/all"
	"niko1ay-pr/url-shrtnr/internal/http-server/handlers/urls/delete"
	"niko1ay-pr/url-shrtnr/internal/http-server/handlers/urls/redirect"
	"niko1ay-pr/url-shrtnr/internal/http-server/handlers/urls/save"
	"niko1ay-pr/url-shrtnr/internal/http-server/handlers/urls/update"
	"niko1ay-pr/url-shrtnr/internal/models"
	"niko1ay-pr/url-shrtnr/internal/repository/redisstore"
	"time"
)

type Storage interface {
	save.URLSaver
	update.URLUpdater
	delete.URLDeleter
	redirect.URLGetter
	all.URLGetterAll
}

type CachedStorage struct {
	primary Storage
	cache   *redisstore.Storage
}

func New(ctx context.Context, storage Storage, redisStorage *redisstore.Storage, ttl time.Duration) *CachedStorage {
	return &CachedStorage{primary: storage, cache: redisStorage}
}

func (c *CachedStorage) SaveURL(ctx context.Context, urlToSave string, alias string, ownerID int64) error {
	err := c.primary.SaveURL(ctx, urlToSave, alias, ownerID)
	if err != nil {
		return err
	}
	err = c.cache.SaveURL(ctx, urlToSave, alias)

	return err
}

func (c *CachedStorage) UpdateURL(ctx context.Context, url string, alias string, ownerId int64) error {
	err := c.primary.UpdateURL(ctx, url, alias, ownerId)
	if err != nil {
		return err
	}
	err = c.cache.UpdateURL(ctx, alias, url)
	return err
}

func (c *CachedStorage) DeleteURL(ctx context.Context, alias string, ownerID int64) error {
	err := c.cache.DeleteURL(ctx, alias)
	if err != nil {
		return err
	}
	err = c.primary.DeleteURL(ctx, alias, ownerID)
	return err
}

func (c *CachedStorage) GetURL(ctx context.Context, alias string) (string, error) {
	url, err := c.cache.GetURL(ctx, alias)
	if err == nil {
		return url, nil
	}

	url, err = c.primary.GetURL(ctx, alias)
	if err != nil {
		return "", err
	}

	err = c.cache.SaveURL(ctx, url, alias)
	if err != nil {
		return "", err
	}

	return url, nil
}

func (c *CachedStorage) GetAll(ctx context.Context) ([]models.UrlEntity, error) {
	urls, err := c.primary.GetAll(ctx)
	return urls, err
}
