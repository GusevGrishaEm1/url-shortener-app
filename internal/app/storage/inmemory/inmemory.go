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

// StorageInMemory представляет хранилище URL-ов в памяти.
type StorageInMemory struct {
	urls        map[string]models.URL
	urlsOfUsers map[int][]models.URL
	sync.RWMutex
	userIDSeq atomic.Int64
	config    config.Config
}

// NewInMemoryStorage создает новый экземпляр хранилища URL-ов в памяти.
func NewInMemoryStorage(config config.Config) *StorageInMemory {
	return &StorageInMemory{
		urls:        make(map[string]models.URL),
		urlsOfUsers: make(map[int][]models.URL),
		config:      config,
	}
}

// FindByShortURL находит оригинальный URL по сокращенному URL.
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

// Save сохраняет URL в хранилище.
func (storage *StorageInMemory) Save(ctx context.Context, url models.URL) error {
	storage.Lock()
	defer storage.Unlock()
	storage.saveURLForUser(ctx, url)
	storage.urls[url.ShortURL] = url
	return nil
}

func (storage *StorageInMemory) saveURLForUser(_ context.Context, url models.URL) {
	if url.CreatedBy == 0 {
		return
	}
	urls, ok := storage.urlsOfUsers[url.CreatedBy]
	if ok {
		urls = append(urls, url)
		storage.urlsOfUsers[url.CreatedBy] = urls
		return
	}
	storage.userIDSeq.Add(1)
	storage.urlsOfUsers[url.CreatedBy] = make([]models.URL, 1)
	storage.urlsOfUsers[url.CreatedBy][0] = url
}

// Ping проверяет доступность хранилища.
func (storage *StorageInMemory) Ping(_ context.Context) bool {
	return true
}

// SaveBatch сохраняет список URL в хранилище.
func (storage *StorageInMemory) SaveBatch(ctx context.Context, urls []models.URL) error {
	storage.Lock()
	defer storage.Unlock()
	for _, url := range urls {
		storage.saveURLForUser(ctx, url)
		storage.urls[url.ShortURL] = url
	}
	return nil
}

// GetUserID возвращает идентификатор пользователя из хранилища.
func (storage *StorageInMemory) GetUserID(context.Context) int {
	userID := storage.userIDSeq.Load()
	storage.userIDSeq.Add(1)
	return int(userID)
}

// FindByUser находит URL, созданные конкретным пользователем.
func (storage *StorageInMemory) FindByUser(ctx context.Context, userID int) ([]models.URL, error) {
	storage.RLock()
	defer storage.RUnlock()
	urls, ok := storage.urlsOfUsers[userID]
	if !ok {
		return nil, customerrors.NewCustomErrorBadRequest(errors.New("original url isn't found"))
	}
	return urls, nil
}

// DeleteUrls удаляет список URL из хранилища.
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

// IsShortURLExists проверяет, существует ли указанный сокращенный URL в хранилище.
func (storage *StorageInMemory) IsShortURLExists(_ context.Context, shortURL string) (bool, error) {
	storage.RLock()
	defer storage.RUnlock()
	_, ok := storage.urls[shortURL]
	return ok, nil
}

// GetStats возвращает статистику по хранилищу.
func (storage *StorageInMemory) GetStats(ctx context.Context) (models.Stats, error) {
	var stats models.Stats
	countURLS := 0
	for _, val := range storage.urls {
		if !val.IsDeleted {
			countURLS++
		}
	}
	stats.URLS = countURLS
	stats.Users = len(storage.urlsOfUsers)
	return stats, nil
}
