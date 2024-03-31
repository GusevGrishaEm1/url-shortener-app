package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/config"
	customerrors "github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/errors"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/models"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/security"
)

type ShortenerService interface {
	CreateShortURL(ctx context.Context, userInfo models.UserInfo, shortURL string) (string, error)
	CreateBatchShortURL(ctx context.Context, userInfo models.UserInfo, arr []models.OriginalURLInfoBatch) ([]models.ShortURLInfoBatch, error)
	GetByShortURL(ctx context.Context, shortURL string) (string, error)
	PingStorage(ctx context.Context) bool
	GetUrlsByUser(ctx context.Context, userInfo models.UserInfo) ([]models.URLByUser, error)
	DeleteUrlsByUser(ctx context.Context, userInfo models.UserInfo, urls []string)
}

type ShortenerHandlerImpl struct {
	service      ShortenerService
	serverConfig config.Config
}

func New(config config.Config, service ShortenerService) *ShortenerHandlerImpl {
	return &ShortenerHandlerImpl{
		service:      service,
		serverConfig: config,
	}
}

func (handler *ShortenerHandlerImpl) ShortenHandler(res http.ResponseWriter, req *http.Request) {
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

func (handler *ShortenerHandlerImpl) validateShortenHandlerResult(err error, res http.ResponseWriter) bool {
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

func (handler *ShortenerHandlerImpl) ShortenJSONHandler(res http.ResponseWriter, req *http.Request) {
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

func (handler *ShortenerHandlerImpl) validateShortenJSONHandlerResult(err error, res http.ResponseWriter) bool {
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

func (handler *ShortenerHandlerImpl) ExpandHandler(res http.ResponseWriter, req *http.Request) {
	originalURL, err := handler.service.GetByShortURL(req.Context(), req.URL.Path[1:])
	shouldReturn := handler.validateExpandHandlerResult(err, res)
	if shouldReturn {
		return
	}
	res.Header().Add("Location", originalURL)
	res.WriteHeader(http.StatusTemporaryRedirect)
}

func (*ShortenerHandlerImpl) validateExpandHandlerResult(err error, res http.ResponseWriter) bool {
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

func (handler *ShortenerHandlerImpl) PingStorageHandler(res http.ResponseWriter, req *http.Request) {
	ok := handler.service.PingStorage(req.Context())
	if !ok {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	res.WriteHeader(http.StatusOK)
}

func (handler *ShortenerHandlerImpl) ShortenJSONBatchHandler(res http.ResponseWriter, req *http.Request) {
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

func (handler *ShortenerHandlerImpl) validateShortenJSONBatchHandlerResult(err error, res http.ResponseWriter) bool {
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

func (handler *ShortenerHandlerImpl) UrlsByUserHandler(res http.ResponseWriter, req *http.Request) {
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

func (handler *ShortenerHandlerImpl) DeleteUrlsHandler(res http.ResponseWriter, req *http.Request) {
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

func (ShortenerHandlerImpl) getUserInfo(ctx context.Context) models.UserInfo {
	userInfo := models.UserInfo{}
	if user := ctx.Value(security.UserID); user != nil {
		userID, ok := user.(int)
		if ok {
			userInfo.UserID = userID
		}
	}
	return userInfo
}
