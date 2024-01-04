package service

import (
	"sync"

	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/util"
)

type ShortenerService interface {
	CreateShortURL(originalURL string) (string, bool)
	GetByShortURL(shortURL string) (string, bool)
}

type ShortenerServiceImpl struct {
	mu   sync.Mutex
	urls map[string]string
}

func New() ShortenerService {
	return &ShortenerServiceImpl{
		urls: make(map[string]string),
	}
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
	return shortURL, true
}

func (service *ShortenerServiceImpl) GetByShortURL(shortURL string) (string, bool) {
	service.mu.Lock()
	defer service.mu.Unlock()
	originalURL, ok := service.urls[shortURL]
	return originalURL, ok
}
