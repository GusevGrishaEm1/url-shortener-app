package service

import (
	"context"
	"sync"

	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/config"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/models"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/storage"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/util"
	"github.com/jackc/pgx/v5"
)

type ShortenerService interface {
	CreateShortURL(originalURL string) (string, bool)
	GetByShortURL(shortURL string) (string, bool)
	PingDB() bool
}

type ShortenerServiceImpl struct {
	config  *config.Config
	mu      sync.Mutex
	urls    map[string]string
	storage storage.URLStorage
	uuidSeq int
}

func New(config *config.Config) ShortenerService {
	storage, err := storage.New(config.FileStoragePath)
	urls, uuidSeq := initUrlsFromStorage(storage, err == nil)
	return &ShortenerServiceImpl{
		urls:    urls,
		config:  config,
		storage: storage,
		uuidSeq: uuidSeq,
	}
}

func initUrlsFromStorage(storage storage.URLStorage, isFileOpen bool) (map[string]string, int) {
	uuidSeq := 1
	urls := make(map[string]string)
	if isFileOpen {
		urlsFromFile := storage.LoadFromStorage()
		for _, el := range urlsFromFile {
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
	if service.storage != nil {
		err := saveToStorage(service, service.uuidSeq, shortURL, originalURL)
		if err != nil {
			return "", false
		}
	}
	return shortURL, true
}

func saveToStorage(service *ShortenerServiceImpl, uuid int, shortURL string, originalURL string) error {
	err := service.storage.SaveToStorage(models.URLInfo{
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

func (service *ShortenerServiceImpl) PingDB() bool {
	conn, err := pgx.Connect(context.Background(), service.config.DatabaseURL)
	if err != nil {
		return false
	}
	defer conn.Close(context.Background())
	return conn.Ping(context.Background()) == nil
}
