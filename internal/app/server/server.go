package server

import (
	"net/http"

	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/config"
	gzipreq "github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/gzip"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/handlers"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/logger"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/security"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/service"
	"github.com/go-chi/chi/v5"
)

type ShortenerHandler interface {
	ShortenHandler(res http.ResponseWriter, req *http.Request)
	ShortenJSONHandler(res http.ResponseWriter, req *http.Request)
	ShortenJSONBatchHandler(res http.ResponseWriter, req *http.Request)
	ExpandHandler(res http.ResponseWriter, req *http.Request)
	PingStorageHandler(res http.ResponseWriter, req *http.Request)
	UrlsByUserHandler(res http.ResponseWriter, req *http.Request)
	DeleteUrlsHandler(res http.ResponseWriter, req *http.Request)
}

type SecurityHandler interface {
	RequestSecurityOnlyUserID(h http.HandlerFunc) http.HandlerFunc
	RequestSecurity(h http.HandlerFunc) http.HandlerFunc
}

func StartServer(serverConfig *config.Config) error {
	service, err := service.New(serverConfig)
	if err != nil {
		return err
	}
	securityHandler := security.New(serverConfig, service)
	handlers := handlers.New(serverConfig, service)
	mux := initHandlers(serverConfig, handlers, securityHandler)
	err = http.ListenAndServe(serverConfig.ServerURL, mux)
	return err
}

func initHandlers(serverConfig *config.Config, handlers ShortenerHandler, secHandler SecurityHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Post("/", secHandler.RequestSecurity(gzipreq.RequestZipper(logger.RequestLogger(handlers.ShortenHandler))))
	r.Get("/{shorturl}", secHandler.RequestSecurity(gzipreq.RequestZipper(logger.RequestLogger(handlers.ExpandHandler))))
	r.Post("/api/shorten", secHandler.RequestSecurity(gzipreq.RequestZipper(logger.RequestLogger(handlers.ShortenJSONHandler))))
	r.Post("/api/shorten/batch", secHandler.RequestSecurity(gzipreq.RequestZipper(logger.RequestLogger(handlers.ShortenJSONBatchHandler))))
	r.Get("/ping", secHandler.RequestSecurity(gzipreq.RequestZipper(logger.RequestLogger(handlers.PingStorageHandler))))
	r.Get("/api/user/urls", secHandler.RequestSecurityOnlyUserID(gzipreq.RequestZipper(logger.RequestLogger(handlers.UrlsByUserHandler))))
	r.Delete("/api/user/urls", secHandler.RequestSecurityOnlyUserID(gzipreq.RequestZipper(logger.RequestLogger(handlers.DeleteUrlsHandler))))
	return r
}
