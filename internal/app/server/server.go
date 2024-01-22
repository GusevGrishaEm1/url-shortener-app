package server

import (
	"net/http"

	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/config"
	gzipreq "github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/gzip"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/handlers"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/logger"
	"github.com/go-chi/chi/v5"
)

func StartServer(serverConfig *config.Config) error {
	serverConfig.DatabaseURL = "postgres://grisha:grisha@localhost:5432/url"
	mux, err := initHandlers(serverConfig)
	if err != nil {
		return err
	}
	err = http.ListenAndServe(serverConfig.ServerURL, mux)
	return err
}

func initHandlers(serverConfig *config.Config) (*chi.Mux, error) {
	handlers, err := handlers.New(serverConfig)
	r := chi.NewRouter()
	r.Post("/", gzipreq.RequestZipper(logger.RequestLogger(handlers.ShortenHandler)))
	r.Get("/{shorturl}", gzipreq.RequestZipper(logger.RequestLogger(handlers.ExpandHandler)))
	r.Post("/api/shorten", gzipreq.RequestZipper(logger.RequestLogger(handlers.ShortenJSONHandler)))
	r.Post("/api/shorten/batch", gzipreq.RequestZipper(logger.RequestLogger(handlers.ShortenJSONBatchHandler)))
	r.Get("/ping", gzipreq.RequestZipper(logger.RequestLogger(handlers.PingDBHandler)))
	return r, err
}
