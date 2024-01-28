package server

import (
	"net/http"

	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/config"
	gzipreq "github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/gzip"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/handlers"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/logger"
	"github.com/go-chi/chi/v5"
)

type ShortenerHandler interface {
	ShortenHandler(res http.ResponseWriter, req *http.Request)
	ShortenJSONHandler(res http.ResponseWriter, req *http.Request)
	ShortenJSONBatchHandler(res http.ResponseWriter, req *http.Request)
	ExpandHandler(res http.ResponseWriter, req *http.Request)
	PingStorageHandler(res http.ResponseWriter, req *http.Request)
}

func StartServer(serverConfig *config.Config) error {
	handlers, err := handlers.New(serverConfig)
	if err != nil {
		return err
	}
	mux := initHandlers(serverConfig, handlers)
	err = http.ListenAndServe(serverConfig.ServerURL, mux)
	return err
}

func initHandlers(serverConfig *config.Config, handlers ShortenerHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Post("/", gzipreq.RequestZipper(logger.RequestLogger(handlers.ShortenHandler)))
	r.Get("/{shorturl}", gzipreq.RequestZipper(logger.RequestLogger(handlers.ExpandHandler)))
	r.Post("/api/shorten", gzipreq.RequestZipper(logger.RequestLogger(handlers.ShortenJSONHandler)))
	r.Post("/api/shorten/batch", gzipreq.RequestZipper(logger.RequestLogger(handlers.ShortenJSONBatchHandler)))
	r.Get("/ping", gzipreq.RequestZipper(logger.RequestLogger(handlers.PingStorageHandler)))
	return r
}
