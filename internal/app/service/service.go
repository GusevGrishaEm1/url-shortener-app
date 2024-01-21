package service

import (
	"context"
	"errors"
	"sync"

	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/config"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/models"
	repository "github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/storage"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/util"
)

type ShortenerService interface {
	CreateShortURL(ctx context.Context, shortURL string) (string, bool)
	CreateBatchShortURL(ctx context.Context, arr []models.OriginalURLInfoBatch) ([]models.ShortURLInfoBatch, bool)
	GetByShortURL(ctx context.Context, shortURL string) (string, bool)
	PingDB(ctx context.Context) bool
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

func (service *ShortenerServiceImpl) CreateShortURL(ctx context.Context, originalURL string) (string, bool) {
	service.mu.Lock()
	defer service.mu.Unlock()
	if originalURL == "" {
		return "", false
	}
	shortURL, ok := generateShortURL(ctx, service)
	if !ok {
		return "", false
	}
	service.repo.Save(ctx, models.URLInfo{
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	})
	return shortURL, true
}

func generateShortURL(ctx context.Context, service *ShortenerServiceImpl) (string, bool) {
	shortURL := util.GetShortURL()
	_, err := service.repo.FindByShortURL(ctx, shortURL)
	if err != nil && !errors.Is(err, repository.ErrOriginalURLNotFound) {
		return "", false
	}
	for err == nil {
		shortURL = util.GetShortURL()
		_, err = service.repo.FindByShortURL(ctx, shortURL)
		if err != nil && !errors.Is(err, repository.ErrOriginalURLNotFound) {
			return "", false
		}
	}
	return shortURL, true
}

func (service *ShortenerServiceImpl) GetByShortURL(ctx context.Context, shortURL string) (string, bool) {
	service.mu.Lock()
	defer service.mu.Unlock()
	url, err := service.repo.FindByShortURL(ctx, shortURL)
	if err != nil {
		return "", false
	}
	return url.OriginalURL, true
}

func (service *ShortenerServiceImpl) PingDB(ctx context.Context) bool {
	service.mu.Lock()
	defer service.mu.Unlock()
	return service.repo.PingDB(ctx)
}

func (service *ShortenerServiceImpl) CreateBatchShortURL(ctx context.Context, arr []models.OriginalURLInfoBatch) ([]models.ShortURLInfoBatch, bool) {
	service.mu.Lock()
	defer service.mu.Unlock()
	if len(arr) == 0 {
		return nil, false
	}
	arrayToSave := make([]models.URLInfo, len(arr))
	arrayToReturn := make([]models.ShortURLInfoBatch, len(arr))
	for i, url := range arr {
		shortURL, ok := generateShortURL(ctx, service)
		if !ok {
			return nil, false
		}
		arrayToSave[i] = models.URLInfo{
			ShortURL:    shortURL,
			OriginalURL: url.OriginalURL,
		}
		arrayToReturn[i] = models.ShortURLInfoBatch{
			CorrelationID: url.CorrelationID,
			ShortURL:      service.config.BaseReturnURL + "/" + shortURL,
		}
	}
	err := service.repo.SaveBatch(ctx, arrayToSave)
	if err != nil {
		return nil, false
	}
	return arrayToReturn, true
}
