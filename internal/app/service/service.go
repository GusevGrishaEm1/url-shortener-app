package service

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/config"
	customerrors "github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/errors"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/models"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/storage"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/util"
)

type ShortenerServiceImpl struct {
	config  *config.Config
	storage storage.Storage
	ch      chan models.URLToDelete
}

func New(ctx context.Context, config *config.Config) (*ShortenerServiceImpl, error) {
	storage, err := storage.New(storage.GetStorageTypeByConfig(config), config)
	if err != nil {
		return nil, err
	}
	service := &ShortenerServiceImpl{
		config:  config,
		storage: storage,
		ch:      make(chan models.URLToDelete, 1024),
	}
	go service.deleteURLBatch(ctx)
	return service, nil
}

func (service *ShortenerServiceImpl) CreateShortURL(ctx context.Context, userInfo models.UserInfo, originalURL string) (string, error) {
	if originalURL == "" {
		return "", customerrors.NewCustomErrorBadRequest(errors.New("original url is empty"))
	}
	shortURL, err := service.generateShortURL(ctx)
	if err != nil {
		return "", err
	}
	err = service.storage.Save(ctx, models.URL{
		ShortURL:    shortURL,
		OriginalURL: originalURL,
		CreatedBy:   userInfo.UserID,
	})
	return shortURL, err
}

func (service *ShortenerServiceImpl) generateShortURL(ctx context.Context) (string, error) {
	shortURL := util.GetShortURL()
	ok, err := service.storage.IsShortURLExists(ctx, shortURL)
	if err != nil {
		return "", err
	}
	for ok {
		shortURL := util.GetShortURL()
		ok, err = service.storage.IsShortURLExists(ctx, shortURL)
		if err != nil {
			return "", err
		}
	}
	return shortURL, nil
}

func (service *ShortenerServiceImpl) GetByShortURL(ctx context.Context, shortURL string) (string, error) {
	url, err := service.storage.FindByShortURL(ctx, shortURL)
	if err != nil {
		return "", err
	}
	if url.IsDeleted {
		err := customerrors.NewCustomError(errors.New("original url is deleted"))
		err.Status = http.StatusGone
		return "", err
	}
	return url.OriginalURL, nil
}

func (service *ShortenerServiceImpl) PingStorage(ctx context.Context) bool {
	return service.storage.Ping(ctx)
}

func (service *ShortenerServiceImpl) CreateBatchShortURL(ctx context.Context, userInfo models.UserInfo, arr []models.OriginalURLInfoBatch) ([]models.ShortURLInfoBatch, error) {
	if len(arr) == 0 {
		return nil, customerrors.NewCustomErrorBadRequest(errors.New("original url is empty"))
	}
	arrayToSave := make([]models.URL, len(arr))
	arrayToReturn := make([]models.ShortURLInfoBatch, len(arr))
	for i, url := range arr {
		shortURL, err := service.generateShortURL(ctx)
		if err != nil {
			return nil, customerrors.NewCustomErrorInternal(err)
		}
		arrayToSave[i] = models.URL{
			ShortURL:    shortURL,
			OriginalURL: url.OriginalURL,
			CreatedBy:   userInfo.UserID,
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
	return service.storage.GetUserID(ctx)
}

func (service *ShortenerServiceImpl) GetUrlsByUser(ctx context.Context, userInfo models.UserInfo) ([]models.URLByUser, error) {
	urls, err := service.storage.FindByUser(ctx, userInfo.UserID)
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

func (service *ShortenerServiceImpl) DeleteUrlsByUser(ctx context.Context, userInfo models.UserInfo, urls []string) {
	go func() {
		urlsToDelete := make([]models.URLToDelete, len(urls))
		for i, el := range urls {
			urlsToDelete[i].ShortURL = el
			urlsToDelete[i].UserID = userInfo.UserID
		}
		for _, el := range urlsToDelete {
			service.ch <- el
		}
	}()
}

func (service *ShortenerServiceImpl) deleteURLBatch(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	urlsToDelete := make([]models.URLToDelete, 0)
loop:
	for {
		select {
		case url := <-service.ch:
			urlsToDelete = append(urlsToDelete, url)
			if len(urlsToDelete) >= 1000 {
				err := service.storage.DeleteUrls(ctx, urlsToDelete)
				if err != nil {
					continue
				}
				urlsToDelete = make([]models.URLToDelete, 0)
			}
		case <-ctx.Done():
			if len(urlsToDelete) == 0 {
				break loop
			}
			service.storage.DeleteUrls(ctx, urlsToDelete)
			break loop
		case <-ticker.C:
			if len(urlsToDelete) == 0 {
				continue
			}
			err := service.storage.DeleteUrls(ctx, urlsToDelete)
			if err != nil {
				continue
			}
			urlsToDelete = make([]models.URLToDelete, 0)
		}
	}
}
