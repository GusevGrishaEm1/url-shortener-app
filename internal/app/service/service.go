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

var (
	ErrOriginalIsEmpty = errors.New("original url is empty")
)

type ShortenerService interface {
	CreateShortURL(ctx context.Context, shortURL string) (string, error)
	CreateBatchShortURL(ctx context.Context, arr []models.OriginalURLInfoBatch) ([]models.ShortURLInfoBatch, error)
	GetByShortURL(ctx context.Context, shortURL string) (string, error)
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

func (service *ShortenerServiceImpl) CreateShortURL(ctx context.Context, originalURL string) (string, error) {
	service.mu.Lock()
	defer service.mu.Unlock()
	if originalURL == "" {
		return "", ErrOriginalIsEmpty
	}
	shortURL, err := generateShortURL(ctx, service)
	if err != nil {
		return "", err
	}
	service.repo.Save(ctx, models.URLInfo{
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	})
	return shortURL, nil
}

func generateShortURL(ctx context.Context, service *ShortenerServiceImpl) (string, error) {
	shortURL := util.GetShortURL()
	_, err := service.repo.FindByShortURL(ctx, shortURL)
	if err != nil && !errors.Is(err, repository.ErrOriginalURLNotFound) {
		return "", err
	}
	for err == nil {
		shortURL = util.GetShortURL()
		_, err = service.repo.FindByShortURL(ctx, shortURL)
		if err != nil && !errors.Is(err, repository.ErrOriginalURLNotFound) {
			return "", err
		}
	}
	return shortURL, nil
}

func (service *ShortenerServiceImpl) GetByShortURL(ctx context.Context, shortURL string) (string, error) {
	service.mu.Lock()
	defer service.mu.Unlock()
	url, err := service.repo.FindByShortURL(ctx, shortURL)
	if err != nil {
		return "", err
	}
	return url.OriginalURL, nil
}

func (service *ShortenerServiceImpl) PingDB(ctx context.Context) bool {
	service.mu.Lock()
	defer service.mu.Unlock()
	return service.repo.PingDB(ctx)
}

func (service *ShortenerServiceImpl) CreateBatchShortURL(ctx context.Context, arr []models.OriginalURLInfoBatch) ([]models.ShortURLInfoBatch, error) {
	service.mu.Lock()
	defer service.mu.Unlock()
	if len(arr) == 0 {
		return nil, ErrOriginalIsEmpty
	}
	arrayToSave := make([]models.URLInfo, len(arr))
	arrayToReturn := make([]models.ShortURLInfoBatch, len(arr))
	for i, url := range arr {
		shortURL, err := generateShortURL(ctx, service)
		if err != nil {
			return nil, err
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
		return nil, err
	}
	return arrayToReturn, nil
}
