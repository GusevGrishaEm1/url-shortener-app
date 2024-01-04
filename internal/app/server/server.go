package server

import (
	"net/http"

	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/config"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/handlers"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/logger"
	"github.com/go-chi/chi/v5"
)

func StartServer(serverConfig *config.Config) error {
	mux := initHandlers(serverConfig)
	err := http.ListenAndServe(serverConfig.GetServerURL(), mux)
	return err
}

func initHandlers(serverConfig *config.Config) *chi.Mux {
	handlers := handlers.New(serverConfig)
	r := chi.NewRouter()
	r.Post("/", logger.RequestLogger(handlers.ShortenHandler))
	r.Get("/{shorturl}", logger.RequestLogger(handlers.ExpandHandler))
	r.Post("/shorten", logger.RequestLogger(handlers.ShortenJSONHandler))
	return r
}
