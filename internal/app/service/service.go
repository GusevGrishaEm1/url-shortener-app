// Package service предоставляет функциональность для работы с URL-сервисом, включая создание, получение и удаление URL-ов.
package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/config"
	customerrors "github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/errors"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/logger"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/models"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/storage"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/util"
)

type shortenerService struct {
	config  config.Config
	storage storage.ShortenerStorage
	ch      chan models.URLToDelete
}

// NewShortenerService создает новый экземпляр сервиса для работы с URL.
func NewShortenerService(ctx context.Context, config config.Config, storage storage.ShortenerStorage) (*shortenerService, error) {
	service := &shortenerService{
		config:  config,
		storage: storage,
		ch:      make(chan models.URLToDelete, 1024),
	}
	go service.deleteURLBatch(ctx)
	return service, nil
}

// CreateShortURL создает короткую ссылку на основе переданного URL.
func (service *shortenerService) CreateShortURL(ctx context.Context, userInfo models.UserInfo, originalURL string) (string, error) {
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

func (service *shortenerService) generateShortURL(ctx context.Context) (string, error) {
	shortURL := util.GenerateShortURL()
	ok, err := service.storage.IsShortURLExists(ctx, shortURL)
	if err != nil {
		return "", err
	}
	for ok {
		shortURL := util.GenerateShortURL()
		ok, err = service.storage.IsShortURLExists(ctx, shortURL)
		if err != nil {
			return "", err
		}
	}
	return shortURL, nil
}

// GetByShortURL возвращает оригинальный URL по короткой ссылке.
func (service *shortenerService) GetByShortURL(ctx context.Context, shortURL string) (string, error) {
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

// PingStorage выполняет ping хранилища.
func (service *shortenerService) PingStorage(ctx context.Context) bool {
	return service.storage.Ping(ctx)
}

// CreateBatchShortURL создает короткие ссылки для массива URL-ов.
func (service *shortenerService) CreateBatchShortURL(ctx context.Context, userInfo models.UserInfo, arr []models.OriginalURLInfoBatch) ([]models.ShortURLInfoBatch, error) {
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

// GetUserID возвращает идентификатор пользователя.
func (service *shortenerService) GetUserID(ctx context.Context) int {
	return service.storage.GetUserID(ctx)
}

// GetUrlsByUser возвращает URL-ы, созданные пользователем.
func (service *shortenerService) GetUrlsByUser(ctx context.Context, userInfo models.UserInfo) ([]models.URLByUser, error) {
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

// DeleteUrlsByUser удаляет URL-ы, созданные пользователем.
func (service *shortenerService) DeleteUrlsByUser(ctx context.Context, userInfo models.UserInfo, urls []string) {
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

func (service *shortenerService) deleteURLBatch(ctx context.Context) {
	tickerPeriod := 10 * time.Second
	ticker := time.NewTicker(tickerPeriod)
	maxSizeArray := 1000
	urlsToDelete := make([]models.URLToDelete, 0, maxSizeArray)
	defer func() {
		if len(urlsToDelete) > 0 {
			err := service.storage.DeleteUrls(ctx, urlsToDelete)
			if err != nil {
				service.logErrorWhenDeleteUrls(err, urlsToDelete)
			}
		}
	}()
	for {
		select {
		case url := <-service.ch:
			urlsToDelete = append(urlsToDelete, url)
			if len(urlsToDelete) >= maxSizeArray {
				err := service.storage.DeleteUrls(ctx, urlsToDelete)
				if err != nil {
					service.logErrorWhenDeleteUrls(err, urlsToDelete)
					continue
				}
				urlsToDelete = urlsToDelete[:0]
			}
		case <-ctx.Done():
			if len(urlsToDelete) == 0 {
				return
			}
			service.storage.DeleteUrls(ctx, urlsToDelete)
			return
		case <-ticker.C:
			if len(urlsToDelete) == 0 {
				continue
			}
			err := service.storage.DeleteUrls(ctx, urlsToDelete)
			if err != nil {
				service.logErrorWhenDeleteUrls(err, urlsToDelete)
				continue
			}
			urlsToDelete = urlsToDelete[:0]
		}
	}
}

func (*shortenerService) logErrorWhenDeleteUrls(err error, urlsToDelete []models.URLToDelete) {
	urlsJSON, errJSON := json.Marshal(urlsToDelete)
	if errJSON != nil {
		logger.Logger.Error("failed to marshal urls to delete", errJSON)
		return
	}
	logger.Logger.Error("delete urls error", err, urlsJSON)
}
