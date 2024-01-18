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
	ExpandHandler(res http.ResponseWriter, req *http.Request)
	PingDBHandler(res http.ResponseWriter, req *http.Request)
}

type ShortenerHandlerImpl struct {
	service      service.ShortenerService
	serverConfig *config.Config
}

func New(config *config.Config) ShortenerHandler {
	return &ShortenerHandlerImpl{
		service:      service.New(config),
		serverConfig: config,
	}
}

func (handler *ShortenerHandlerImpl) ShortenHandler(res http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	shortURL, ok := handler.service.CreateShortURL(string(body))
	if !ok {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	res.Header().Add("content-type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	res.Write([]byte(handler.serverConfig.BaseReturnURL + "/" + shortURL))
}

func (handler *ShortenerHandlerImpl) ShortenJSONHandler(res http.ResponseWriter, req *http.Request) {
	var reqModel models.Request
	dec := json.NewDecoder(req.Body)
	if err := dec.Decode(&reqModel); err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	shortURL, ok := handler.service.CreateShortURL(reqModel.URL)
	if !ok {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	resModel := models.Response{
		Result: handler.serverConfig.BaseReturnURL + "/" + shortURL,
	}
	res.Header().Add("content-type", "application/json")
	res.WriteHeader(http.StatusCreated)
	enc := json.NewEncoder(res)
	if err := enc.Encode(resModel); err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
}

func (handler *ShortenerHandlerImpl) ExpandHandler(res http.ResponseWriter, req *http.Request) {
	originalURL, ok := handler.service.GetByShortURL(req.URL.Path[1:])
	if !ok {
		res.WriteHeader(http.StatusBadRequest)
		return
	}
	res.Header().Add("Location", originalURL)
	res.WriteHeader(http.StatusTemporaryRedirect)
}

func (handler *ShortenerHandlerImpl) PingDBHandler(res http.ResponseWriter, req *http.Request) {
	ok := handler.service.PingDB()
	if !ok {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
	res.WriteHeader(http.StatusOK)
}
