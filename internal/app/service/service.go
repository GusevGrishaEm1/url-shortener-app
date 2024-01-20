package service

import (
	"errors"
	"sync"

	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/config"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/models"
	repository "github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/storage"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/util"
)

type ShortenerService interface {
	CreateShortURL(originalURL string) (string, bool)
	GetByShortURL(shortURL string) (string, bool)
	PingDB() bool
}

type ShortenerServiceImpl struct {
	config *config.Config
	mu     sync.Mutex
	repo   repository.URLRepository
}

func New(config *config.Config) (ShortenerService, error) {
	repo, err := repository.New(config)
	return &ShortenerServiceImpl{
		config: config,
		repo:   repo,
	}, err
}

func (service *ShortenerServiceImpl) CreateShortURL(originalURL string) (string, bool) {
	service.mu.Lock()
	defer service.mu.Unlock()
	if originalURL == "" {
		return "", false
	}
	shortURL := util.GetShortURL()
	_, err := service.repo.FindByShortURL(shortURL)
	if err != nil && !errors.Is(err, repository.OriginalURLNotFound) {
		return "", false
	}
	for err == nil {
		shortURL = util.GetShortURL()
		_, err = service.repo.FindByShortURL(shortURL)
		if err != nil && !errors.Is(err, repository.OriginalURLNotFound) {
			return "", false
		}
	}
	service.repo.Save(models.URLInfo{
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	})
	return shortURL, true
}

func (service *ShortenerServiceImpl) GetByShortURL(shortURL string) (string, bool) {
	service.mu.Lock()
	defer service.mu.Unlock()
	url, err := service.repo.FindByShortURL(shortURL)
	if err != nil {
		return "", false
	}
	return url.OriginalURL, true
}

func (service *ShortenerServiceImpl) PingDB() bool {
	service.mu.Lock()
	defer service.mu.Unlock()
	return service.repo.PingDB()
}
