package service

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/config"
	customerrors "github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/errors"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/models"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/security"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/storage"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/util"
)

type ShortenerServiceImpl struct {
	config *config.Config
	sync.Mutex
	storage storage.Storage
	ch      chan models.URLToDelete
}

func New(config *config.Config) (*ShortenerServiceImpl, error) {
	storage, err := storage.New(storage.GetStorageTypeByConfig(config), config)
	if err != nil {
		return nil, err
	}
	service := &ShortenerServiceImpl{
		config:  config,
		storage: storage,
		ch:      make(chan models.URLToDelete, 1024),
	}
	go service.deleteURLBatch()
	return service, nil
}

func (service *ShortenerServiceImpl) CreateShortURL(ctx context.Context, originalURL string) (string, error) {
	service.Lock()
	defer service.Unlock()
	if originalURL == "" {
		return "", customerrors.ErrOriginalIsEmpty
	}
	shortURL, err := generateShortURL(ctx, service)
	if err != nil {
		return "", err
	}
	userID := 0
	if user := ctx.Value(security.UserID); user != nil {
		userID = user.(int)
	}
	err = service.storage.Save(ctx, models.URL{
		ShortURL:    shortURL,
		OriginalURL: originalURL,
		CreatedBy:   userID,
	})
	return shortURL, err
}

func generateShortURL(ctx context.Context, service *ShortenerServiceImpl) (string, error) {
	shortURL := util.GetShortURL()
	_, err := service.storage.FindByShortURL(ctx, shortURL)
	if err != nil && !errors.Is(err, customerrors.ErrOriginalURLNotFound) {
		return "", err
	}
	for err == nil {
		shortURL = util.GetShortURL()
		_, err = service.storage.FindByShortURL(ctx, shortURL)
		if err != nil && !errors.Is(err, customerrors.ErrOriginalURLNotFound) {
			return "", err
		}
	}
	return shortURL, nil
}

func (service *ShortenerServiceImpl) GetByShortURL(ctx context.Context, shortURL string) (string, error) {
	service.Lock()
	defer service.Unlock()
	url, err := service.storage.FindByShortURL(ctx, shortURL)
	if err != nil {
		return "", err
	}
	if url.IsDeleted {
		return "", customerrors.ErrOriginalURLIsDeleted
	}
	return url.OriginalURL, nil
}

func (service *ShortenerServiceImpl) PingStorage(ctx context.Context) bool {
	service.Lock()
	defer service.Unlock()
	return service.storage.Ping(ctx)
}

func (service *ShortenerServiceImpl) CreateBatchShortURL(ctx context.Context, arr []models.OriginalURLInfoBatch) ([]models.ShortURLInfoBatch, error) {
	service.Lock()
	defer service.Unlock()
	if len(arr) == 0 {
		return nil, customerrors.ErrOriginalIsEmpty
	}
	arrayToSave := make([]models.URL, len(arr))
	arrayToReturn := make([]models.ShortURLInfoBatch, len(arr))
	userID := service.getUserID(ctx)
	for i, url := range arr {
		shortURL, err := generateShortURL(ctx, service)
		if err != nil {
			return nil, err
		}
		arrayToSave[i] = models.URL{
			ShortURL:    shortURL,
			OriginalURL: url.OriginalURL,
			CreatedBy:   userID,
		}
		arrayToReturn[i] = models.ShortURLInfoBatch{
			CorrelationID: url.CorrelationID,
			ShortURL:      service.config.BaseReturnURL + "/" + shortURL,
		}
	}
	err := service.storage.SaveBatch(ctx, arrayToSave)
	if err != nil {
		return nil, err
	}
	return arrayToReturn, nil
}

func (service *ShortenerServiceImpl) GetUserID(ctx context.Context) int {
	service.Lock()
	defer service.Unlock()
	return service.storage.GetUserID(ctx)
}

func (service *ShortenerServiceImpl) GetUrlsByUser(ctx context.Context) ([]models.URLByUser, error) {
	service.Lock()
	defer service.Unlock()
	userID := service.getUserID(ctx)
	urls, err := service.storage.FindByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	urlsForUser := make([]models.URLByUser, len(urls))
	for i, el := range urls {
		urlsForUser[i] = models.URLByUser{
			ShortURL:    service.config.BaseReturnURL + "/" + el.ShortURL,
			OriginalURL: el.OriginalURL,
		}
	}
	return urlsForUser, nil
}

func (*ShortenerServiceImpl) getUserID(ctx context.Context) int {
	userID := 0
	if user := ctx.Value(security.UserID); user != nil {
		userID = user.(int)
	}
	return userID
}

func (service *ShortenerServiceImpl) DeleteUrlsByUser(ctx context.Context, urls []models.URLToDelete) error {
	service.Lock()
	defer service.Unlock()
	for _, el := range urls {
		service.ch <- el
	}
	return nil
}

func (service *ShortenerServiceImpl) deleteURLBatch() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ticker := time.NewTicker(10 * time.Second)
	urlsToDelete := make([]models.URLToDelete, 0)
	for {
		select {
		case url := <-service.ch:
			urlsToDelete = append(urlsToDelete, url)
		case <-ticker.C:
			if len(urlsToDelete) == 0 {
				continue
			}
			err := service.storage.DeleteUrls(ctx, urlsToDelete, service.getUserID(ctx))
			urlsToDelete = make([]models.URLToDelete, 0)
			if err != nil {
				continue
			}
		}
	}
}
