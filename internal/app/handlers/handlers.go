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
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/service"
)

type ShortenerHandler interface {
	ShortenHandler(res http.ResponseWriter, req *http.Request)
	ShortenJSONHandler(res http.ResponseWriter, req *http.Request)
	ShortenJSONBatchHandler(res http.ResponseWriter, req *http.Request)
	ExpandHandler(res http.ResponseWriter, req *http.Request)
	PingDBHandler(res http.ResponseWriter, req *http.Request)
}

type ShortenerHandlerImpl struct {
	service      service.ShortenerService
	serverConfig *config.Config
}

func New(config *config.Config) (ShortenerHandler, error) {
	service, err := service.New(config)
	return &ShortenerHandlerImpl{
		service:      service,
		serverConfig: config,
	}, err
}

func (handler *ShortenerHandlerImpl) ShortenHandler(res http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	body, err := io.ReadAll(req.Body)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	shortURL, err := handler.service.CreateShortURL(ctx, string(body))
	if err != nil {
		var cusErr *customerrors.OriginalURLAlreadyExists
		if errors.As(err, &cusErr) {
			res.Header().Add("content-type", "text/plain")
			res.WriteHeader(http.StatusConflict)
			res.Write([]byte(handler.serverConfig.BaseReturnURL + "/" + err.(*customerrors.OriginalURLAlreadyExists).ShortURL))
			return
		}
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	res.Header().Add("content-type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	res.Write([]byte(handler.serverConfig.BaseReturnURL + "/" + shortURL))
}

func (handler *ShortenerHandlerImpl) ShortenJSONHandler(res http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
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
	shortURL, err := handler.service.CreateShortURL(ctx, reqModel.URL)
	if err != nil {
		var cusErr *customerrors.OriginalURLAlreadyExists
		if errors.As(err, &cusErr) {
			resModel := models.Response{
				Result: handler.serverConfig.BaseReturnURL + "/" + err.(*customerrors.OriginalURLAlreadyExists).ShortURL,
			}
			body, err = json.Marshal(resModel)
			if err != nil {
				res.WriteHeader(http.StatusBadRequest)
				return
			}
			res.Header().Add("content-type", "application/json")
			res.WriteHeader(http.StatusConflict)
			res.Write(body)
			return
		}
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	resModel := models.Response{
		Result: handler.serverConfig.BaseReturnURL + "/" + shortURL,
	}
	body, err = json.Marshal(resModel)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	res.Header().Add("content-type", "application/json")
	res.WriteHeader(http.StatusCreated)
	res.Write(body)
}

func (handler *ShortenerHandlerImpl) ExpandHandler(res http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	originalURL, err := handler.service.GetByShortURL(ctx, req.URL.Path[1:])
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	res.Header().Add("Location", originalURL)
	res.WriteHeader(http.StatusTemporaryRedirect)
}

func (handler *ShortenerHandlerImpl) PingDBHandler(res http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ok := handler.service.PingDB(ctx)
	if !ok {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	res.WriteHeader(http.StatusOK)
}

func (handler *ShortenerHandlerImpl) ShortenJSONBatchHandler(res http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
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
	shortURLArray, err := handler.service.CreateBatchShortURL(ctx, urls)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	body, err = json.Marshal(shortURLArray)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	res.Header().Add("content-type", "application/json")
	res.WriteHeader(http.StatusCreated)
	res.Write(body)
}
