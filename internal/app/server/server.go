package server

import (
	"net/http"

	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/config"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/handler"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/service"
	"github.com/go-chi/chi/v5"
)

func StartServer(serverConfig *config.Config) {
	mux := initHandlers(serverConfig)
	err := http.ListenAndServe(serverConfig.GetServerURL(), mux)
	if err != nil {
		panic(err)
	}
}

func initHandlers(serverConfig *config.Config) *chi.Mux {
	service := service.ShortenerServiceImpl{
		Urls: make(map[string]string),
	}
	handler := handler.ShortHandlerImpl{
		Service:      &service,
		ServerConfig: serverConfig,
	}
	r := chi.NewRouter()
	r.Post("/", handler.ShortHandler)
	r.Get("/{shorturl}", handler.ExpandHandler)
	return r
}
