package inmemory

import (
	"context"

	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/errors"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/models"
)

func NewInMemoryStorage() *StorageInMemory {
	return &StorageInMemory{
		urls:        make(map[string]*models.URL),
		urlsOfUsers: make(map[int][]*models.URL),
		userIDSeq:   1,
	}
}

type StorageInMemory struct {
	urls        map[string]*models.URL
	urlsOfUsers map[int][]*models.URL
	userIDSeq   int
}

func (storage *StorageInMemory) FindByShortURL(_ context.Context, shortURL string) (*models.URL, error) {
	url, ok := storage.urls[shortURL]
	if !ok {
		return nil, errors.ErrOriginalURLNotFound
	}
	return &models.URL{
		ShortURL:    shortURL,
		OriginalURL: url.OriginalURL,
	}, nil
}

func (storage *StorageInMemory) Save(ctx context.Context, url models.URL) error {
	storage.saveSingleURLForUser(ctx, url)
	storage.urls[url.ShortURL] = &url
	return nil
}

func (storage *StorageInMemory) saveSingleURLForUser(ctx context.Context, url models.URL) {
	if url.CreatedBy != 0 {
		urls, ok := storage.urlsOfUsers[url.CreatedBy]
		storage.userIDSeq++
		if ok {
			urls = append(urls, &url)
			storage.urlsOfUsers[url.CreatedBy] = urls
			return
		}
		storage.urlsOfUsers[url.CreatedBy] = make([]*models.URL, 1)
		storage.urlsOfUsers[url.CreatedBy][0] = &url
	}
}

func (storage *StorageInMemory) Ping(_ context.Context) bool {
	return storage.urls != nil
}

func (storage *StorageInMemory) SaveBatch(ctx context.Context, urls []models.URL) error {
	for _, url := range urls {
		err := storage.Save(ctx, url)
		if err != nil {
			return err
		}
	}
	return nil
}

func (storage *StorageInMemory) GetUserID(context.Context) int {
	userID := storage.userIDSeq
	storage.userIDSeq++
	return userID
}

func (storage *StorageInMemory) FindByUser(ctx context.Context, userID int) ([]*models.URL, error) {
	urls, ok := storage.urlsOfUsers[userID]
	if !ok {
		return nil, errors.ErrOriginalURLNotFound
	}
	return urls, nil
}

func (storage *StorageInMemory) DeleteUrls(_ context.Context, urls []models.URLToDelete, userID int) error {
	for _, url := range urls {
		el, ok := storage.urls[string(url)]
		if ok && el.CreatedBy == userID {
			el.IsDeleted = true
			storage.urls[string(url)] = el
		}
	}
	return nil
}
