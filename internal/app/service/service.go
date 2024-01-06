package service

import (
	"sync"

	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/config"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/models"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/storage"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/util"
)

type ShortenerService interface {
	CreateShortURL(originalURL string) (string, bool)
	GetByShortURL(shortURL string) (string, bool)
}

type ShortenerServiceImpl struct {
	mu      sync.Mutex
	urls    map[string]string
	storage storage.URLStorage
	uuidSeq int
}

func New(config *config.Config) ShortenerService {
	storage, err := storage.NewFileStorage(config.FileStoragePath)
	urls, uuidSeq := initUrlsFromStorage(storage, err == nil)
	return &ShortenerServiceImpl{
		storage: storage,
		urls:    urls,
		uuidSeq: uuidSeq,
	}
}

func initUrlsFromStorage(storage storage.URLStorage, isFileOpen bool) (map[string]string, int) {
	uuidSeq := 1
	urls := make(map[string]string)
	if isFileOpen {
		array := storage.LoadFromStorage()
		for _, el := range array {
			if uuidSeq <= el.UUID {
				uuidSeq = el.UUID + 1
			}
			urls[el.ShortURL] = el.OriginalURL
		}
	}
	return urls, uuidSeq
}

func (service *ShortenerServiceImpl) CreateShortURL(originalURL string) (string, bool) {
	service.mu.Lock()
	defer service.mu.Unlock()
	if originalURL == "" {
		return "", false
	}
	shortURL := util.GetShortURL()
	for _, ok := service.urls[shortURL]; ok; {
		shortURL = util.GetShortURL()
	}
	service.urls[shortURL] = originalURL
	err := saveToStorage(service, service.uuidSeq, shortURL, originalURL)
	if err != nil {
		return "", false
	}
	return shortURL, true
}

func saveToStorage(service *ShortenerServiceImpl, uuid int, shortURL string, originalURL string) error {
	err := service.storage.SaveToStorage(models.StorageURLInfo{
		UUID:        uuid,
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	})
	if err == nil {
		service.uuidSeq++
	}
	return err
}

func (service *ShortenerServiceImpl) GetByShortURL(shortURL string) (string, bool) {
	service.mu.Lock()
	defer service.mu.Unlock()
	originalURL, ok := service.urls[shortURL]
	return originalURL, ok
}
