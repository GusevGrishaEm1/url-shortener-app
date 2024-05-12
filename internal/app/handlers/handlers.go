package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/config"
	customerrors "github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/errors"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/middlewares/security"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/models"
)

// ShortenerService определяет методы для взаимодействия с сервисом сокращения URL.
type ShortenerService interface {
	// CreateShortURL создает сокращенный URL на основе исходного URL.
	CreateShortURL(ctx context.Context, userInfo models.UserInfo, shortURL string) (string, error)
	// CreateBatchShortURL создает несколько сокращенных URL на основе списка исходных URL.
	CreateBatchShortURL(ctx context.Context, userInfo models.UserInfo, arr []models.OriginalURLInfoBatch) ([]models.ShortURLInfoBatch, error)
	// GetByShortURL возвращает исходный URL по сокращенному URL.
	GetByShortURL(ctx context.Context, shortURL string) (string, error)
	// PingStorage проверяет доступность хранилища данных.
	PingStorage(ctx context.Context) bool
	// GetUrlsByUser возвращает список URL, созданных пользователем.
	GetUrlsByUser(ctx context.Context, userInfo models.UserInfo) ([]models.URLByUser, error)
	// DeleteUrlsByUser удаляет список URL, созданных пользователем.
	DeleteUrlsByUser(ctx context.Context, userInfo models.UserInfo, urls []string)
	// GetStats возвращающий в ответ объект статистики.
	GetStats(ctx context.Context) (models.Stats, error)
}

type shortenerHandler struct {
	service      ShortenerService
	serverConfig config.Config
}

// NewShortenerHandler создает новый экземпляр обработчика
func NewShortenerHandler(config config.Config, service ShortenerService) *shortenerHandler {
	return &shortenerHandler{
		service:      service,
		serverConfig: config,
	}
}

// ShortenHandler создает сокращенный URL на основе исходного URL
func (handler *shortenerHandler) ShortenHandler(res http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	userInfo := handler.getUserInfo(req.Context())
	shortURL, err := handler.service.CreateShortURL(req.Context(), userInfo, string(body))
	shouldReturn := handler.validateShortenHandlerResult(err, res)
	if shouldReturn {
		return
	}
	res.Header().Add("content-type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	res.Write([]byte(handler.serverConfig.BaseReturnURL + "/" + shortURL))
}

func (handler *shortenerHandler) validateShortenHandlerResult(err error, res http.ResponseWriter) bool {
	if err != nil {
		var customerr *customerrors.CustomError
		if errors.As(err, &customerr) {
			if customerr.Status == http.StatusConflict {
				customerr.ContentType = "text/plain"
				customerr.Body = []byte(handler.serverConfig.BaseReturnURL + "/" + customerr.ShortURL)
			}
			if customerr.ContentType != "" {
				res.Header().Add("content-type", customerr.ContentType)
			}
			res.WriteHeader(customerr.Status)
			if customerr.Body != nil {
				res.Write(customerr.Body)
			}
			return true
		}
		res.WriteHeader(http.StatusInternalServerError)
		return true
	}
	return false
}

// ShortenJSONHandler создает сокращенный URL на основе исходного
func (handler *shortenerHandler) ShortenJSONHandler(res http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	var reqModel models.Request
	err = json.Unmarshal(body, &reqModel)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	userInfo := handler.getUserInfo(req.Context())
	shortURL, err := handler.service.CreateShortURL(req.Context(), userInfo, reqModel.URL)
	shouldReturn := handler.validateShortenJSONHandlerResult(err, res)
	if shouldReturn {
		return
	}
	resModel := models.Response{
		Result: handler.serverConfig.BaseReturnURL + "/" + shortURL,
	}
	body, err = json.Marshal(resModel)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	res.Header().Add("content-type", "application/json")
	res.WriteHeader(http.StatusCreated)
	res.Write(body)
}

func (handler *shortenerHandler) validateShortenJSONHandlerResult(err error, res http.ResponseWriter) bool {
	if err != nil {
		var customerr *customerrors.CustomError
		if errors.As(err, &customerr) {
			if customerr.Status == http.StatusConflict {
				customerr.ContentType = "application/json"
				body, err := json.Marshal(&models.Response{
					Result: handler.serverConfig.BaseReturnURL + "/" + customerr.ShortURL,
				})
				if err != nil {
					res.WriteHeader(http.StatusInternalServerError)
					return true
				}
				customerr.Body = body
			}
			if customerr.ContentType != "" {
				res.Header().Add("content-type", customerr.ContentType)
			}
			res.WriteHeader(customerr.Status)
			if customerr.Body != nil {
				res.Write(customerr.Body)
			}
			return true
		}
		res.WriteHeader(http.StatusInternalServerError)
		return true
	}
	return false
}

// ExpandHandler возвращает исходный URL по сокращенному URL.
func (handler *shortenerHandler) ExpandHandler(res http.ResponseWriter, req *http.Request) {
	originalURL, err := handler.service.GetByShortURL(req.Context(), req.URL.Path[1:])
	shouldReturn := handler.validateExpandHandlerResult(err, res)
	if shouldReturn {
		return
	}
	res.Header().Add("Location", originalURL)
	res.WriteHeader(http.StatusTemporaryRedirect)
}

func (*shortenerHandler) validateExpandHandlerResult(err error, res http.ResponseWriter) bool {
	if err != nil {
		var customerr *customerrors.CustomError
		if errors.As(err, &customerr) {
			if customerr.ContentType != "" {
				res.Header().Add("content-type", customerr.ContentType)
			}
			res.WriteHeader(customerr.Status)
			if customerr.Body != nil {
				res.Write(customerr.Body)
			}
			return true
		}
		res.WriteHeader(http.StatusInternalServerError)
		return true
	}
	return false
}

// PingStorageHandler проверяет доступность хранилища данных
func (handler *shortenerHandler) PingStorageHandler(res http.ResponseWriter, req *http.Request) {
	ok := handler.service.PingStorage(req.Context())
	if !ok {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	res.WriteHeader(http.StatusOK)
}

// ShortenJSONBatchHandler создает сокращенный URL на основе исход
func (handler *shortenerHandler) ShortenJSONBatchHandler(res http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	var urls []models.OriginalURLInfoBatch
	err = json.Unmarshal(body, &urls)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	userInfo := handler.getUserInfo(req.Context())
	shortURLArray, err := handler.service.CreateBatchShortURL(req.Context(), userInfo, urls)
	shouldReturn := handler.validateShortenJSONBatchHandlerResult(err, res)
	if shouldReturn {
		return
	}
	body, err = json.Marshal(shortURLArray)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	res.Header().Add("content-type", "application/json")
	res.WriteHeader(http.StatusCreated)
	res.Write(body)
}

func (handler *shortenerHandler) validateShortenJSONBatchHandlerResult(err error, res http.ResponseWriter) bool {
	if err != nil {
		var customerr *customerrors.CustomError
		if errors.As(err, &customerr) {
			if customerr.Status == http.StatusConflict {
				customerr.ContentType = "application/json"
				body, err := json.Marshal(&models.Response{
					Result: handler.serverConfig.BaseReturnURL + "/" + customerr.ShortURL,
				})
				if err != nil {
					res.WriteHeader(http.StatusInternalServerError)
					return true
				}
				customerr.Body = body
			}
			if customerr.ContentType != "" {
				res.Header().Add("content-type", customerr.ContentType)
			}
			res.WriteHeader(customerr.Status)
			if customerr.Body != nil {
				res.Write(customerr.Body)
			}
			return true
		}
		res.WriteHeader(http.StatusInternalServerError)
		return true
	}
	return false
}

// UrlsByUserHandler возвращает все сокращенные URL для пользователя
func (handler *shortenerHandler) UrlsByUserHandler(res http.ResponseWriter, req *http.Request) {
	userInfo := handler.getUserInfo(req.Context())
	urls, err := handler.service.GetUrlsByUser(req.Context(), userInfo)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(urls) == 0 {
		res.WriteHeader(http.StatusNoContent)
		return
	}
	body, err := json.Marshal(urls)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	res.Header().Add("content-type", "application/json")
	res.WriteHeader(http.StatusOK)
	res.Write(body)
}

// DeleteUrlsHandler удаляет все сокращенные URL
func (handler *shortenerHandler) DeleteUrlsHandler(res http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	var urls []string
	if err := json.Unmarshal(body, &urls); err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	userInfo := handler.getUserInfo(req.Context())
	handler.service.DeleteUrlsByUser(req.Context(), userInfo, urls)
	res.WriteHeader(http.StatusAccepted)
}

func (handler *shortenerHandler) StatsHandler(res http.ResponseWriter, req *http.Request) {
	result, err := handler.service.GetStats(req.Context())
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	body, err := json.Marshal(result)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	res.Header().Add("content-type", "application/json")
	res.WriteHeader(http.StatusOK)
	res.Write(body)
}

func (*shortenerHandler) getUserInfo(ctx context.Context) models.UserInfo {
	userInfo := models.UserInfo{}
	if user := ctx.Value(security.UserID); user != nil {
		userID, ok := user.(int)
		if ok {
			userInfo.UserID = userID
		}
	}
	return userInfo
}
