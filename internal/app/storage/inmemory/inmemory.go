package inmemory

import (
	"context"
	"sync"
	"sync/atomic"

	"errors"

	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/config"
	customerrors "github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/errors"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/models"
)

func NewInMemoryStorage(config config.Config) *StorageInMemory {
	return &StorageInMemory{
		urls:        make(map[string]models.URL),
		urlsOfUsers: make(map[int][]models.URL),
		config:      config,
	}
}

type StorageInMemory struct {
	urls        map[string]models.URL
	urlsOfUsers map[int][]models.URL
	sync.RWMutex
	userIDSeq atomic.Int64
	config    config.Config
}

func (storage *StorageInMemory) FindByShortURL(_ context.Context, shortURL string) (*models.URL, error) {
	storage.RLock()
	defer storage.RUnlock()
	url, ok := storage.urls[shortURL]
	if !ok {
		return nil, customerrors.NewCustomErrorBadRequest(errors.New("original url isn't found"))
	}
	return &models.URL{
		ShortURL:    shortURL,
		OriginalURL: url.OriginalURL,
	}, nil
}

func (storage *StorageInMemory) Save(ctx context.Context, url models.URL) error {
	storage.Lock()
	defer storage.Unlock()
	storage.saveURLForUser(ctx, url)
	storage.urls[url.ShortURL] = url
	return nil
}

func (storage *StorageInMemory) saveURLForUser(ctx context.Context, url models.URL) {
	if url.CreatedBy == 0 {
		return
	}
	urls, ok := storage.urlsOfUsers[url.CreatedBy]
	storage.userIDSeq.Add(1)
	if ok {
		urls = append(urls, url)
		storage.urlsOfUsers[url.CreatedBy] = urls
		return
	}
	storage.urlsOfUsers[url.CreatedBy] = make([]models.URL, 1)
	storage.urlsOfUsers[url.CreatedBy][0] = url
}

func (storage *StorageInMemory) Ping(_ context.Context) bool {
	return true
}

func (storage *StorageInMemory) SaveBatch(ctx context.Context, urls []models.URL) error {
	storage.Lock()
	defer storage.Unlock()
	for _, url := range urls {
		storage.saveURLForUser(ctx, url)
		storage.urls[url.ShortURL] = url
	}
	return nil
}

func (storage *StorageInMemory) GetUserID(context.Context) int {
	userID := storage.userIDSeq.Load()
	storage.userIDSeq.Add(1)
	return int(userID)
}

func (storage *StorageInMemory) FindByUser(ctx context.Context, userID int) ([]models.URL, error) {
	storage.RLock()
	defer storage.RUnlock()
	urls, ok := storage.urlsOfUsers[userID]
	if !ok {
		return nil, customerrors.NewCustomErrorBadRequest(errors.New("original url isn't found"))
	}
	return urls, nil
}

func (storage *StorageInMemory) DeleteUrls(_ context.Context, urls []models.URLToDelete) error {
	storage.Lock()
	defer storage.Unlock()
	for _, url := range urls {
		el, ok := storage.urls[url.ShortURL]
		if ok && el.CreatedBy == url.UserID {
			el.IsDeleted = true
			storage.urls[url.ShortURL] = el
		}
	}
	return nil
}

func (storage *StorageInMemory) IsShortURLExists(_ context.Context, shortURL string) (bool, error) {
	storage.RLock()
	defer storage.RUnlock()
	_, ok := storage.urls[shortURL]
	return ok, nil
}
