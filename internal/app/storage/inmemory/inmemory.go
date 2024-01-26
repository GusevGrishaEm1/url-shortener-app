package inmemory

import (
	"context"

	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/errors"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/models"
)

func NewInMemoryStorage() *StorageInMemory {
	return &StorageInMemory{
		urls: make(map[string]string),
	}
}

type StorageInMemory struct {
	urls map[string]string
}

func (storage *StorageInMemory) FindByShortURL(_ context.Context, shortURL string) (*models.URLInfo, error) {
	originalURL, ok := storage.urls[shortURL]
	if !ok {
		return nil, errors.ErrOriginalURLNotFound
	}
	return &models.URLInfo{
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	}, nil
}

func (storage *StorageInMemory) Save(_ context.Context, url models.URLInfo) error {
	storage.urls[url.ShortURL] = url.OriginalURL
	return nil
}

func (storage *StorageInMemory) Ping(_ context.Context) bool {
	return true
}

func (storage *StorageInMemory) SaveBatch(ctx context.Context, urls []models.URLInfo) error {
	for _, url := range urls {
		err := storage.Save(ctx, url)
		if err != nil {
			return err
		}
	}
	return nil
}
