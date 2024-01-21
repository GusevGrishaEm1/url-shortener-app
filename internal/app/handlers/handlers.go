package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/config"
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
	body, err := io.ReadAll(req.Body)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	shortURL, ok := handler.service.CreateShortURL(req.Context(), string(body))
	if !ok {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	res.Header().Add("content-type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	res.Write([]byte(handler.serverConfig.BaseReturnURL + "/" + shortURL))
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
	shortURL, ok := handler.service.CreateShortURL(req.Context(), reqModel.URL)
	if !ok {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	resModel := models.Response{
		Result: handler.serverConfig.BaseReturnURL + "/" + shortURL,
	}
	res.Header().Add("content-type", "application/json")
	res.WriteHeader(http.StatusCreated)
	body, err = json.Marshal(resModel)
	res.Write(body)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
}

func (handler *ShortenerHandlerImpl) ExpandHandler(res http.ResponseWriter, req *http.Request) {
	originalURL, ok := handler.service.GetByShortURL(req.Context(), req.URL.Path[1:])
	if !ok {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	res.Header().Add("Location", originalURL)
	res.WriteHeader(http.StatusTemporaryRedirect)
}

func (handler *ShortenerHandlerImpl) PingDBHandler(res http.ResponseWriter, req *http.Request) {
	ok := handler.service.PingDB(req.Context())
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
	var reqModel models.URLInfoBatchRequest
	err = json.Unmarshal(body, &reqModel)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	shortURLArray, ok := handler.service.CreateBatchShortURL(req.Context(), reqModel.Array)
	if !ok {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	resModel := models.URLInfoBatchResponse{
		Array: shortURLArray,
	}
	res.Header().Add("content-type", "application/json")
	res.WriteHeader(http.StatusCreated)
	body, err = json.Marshal(resModel)
	res.Write(body)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
}
