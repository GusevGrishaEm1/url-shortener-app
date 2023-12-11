package handler

import (
	"io"
	"net/http"

	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/config"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/service"
)

type ShortHandler interface {
	ShortHandler(originalURL string)
	ExpandHandler(shortURL string)
}

type ShortHandlerImpl struct {
	Service      service.ShortenerService
	ServerConfig *config.Config
}

func (handler *ShortHandlerImpl) ShortHandler(res http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
	} else {
		shortURL, ok := handler.Service.CreateShortURL(string(body))
		if ok {
			res.Header().Add("content-type", "text/plain")
			res.WriteHeader(http.StatusCreated)
			res.Write([]byte(handler.ServerConfig.GetBaseReturnURL() + "/" + shortURL))
		} else {
			res.WriteHeader(http.StatusBadRequest)
		}
	}
}

func (handler *ShortHandlerImpl) ExpandHandler(res http.ResponseWriter, req *http.Request) {
	originalURL, ok := handler.Service.GetByShortURL(req.URL.Path[1:])
	if ok {
		res.Header().Add("Location", originalURL)
		res.WriteHeader(http.StatusTemporaryRedirect)
	} else {
		res.WriteHeader(http.StatusBadRequest)
	}
}
