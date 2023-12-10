package main

import (
	"net/http"

	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/config"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/handlers"
	"github.com/go-chi/chi/v5"
)

var urls map[string]string = make(map[string]string)

var serverConfig *config.Config = config.GetDefault()

func Init(config *config.Config) {
	serverConfig = config
	mux := initHandlers()
	err := http.ListenAndServe(serverConfig.GetServerURL(), mux)
	if err != nil {
		panic(err)
	}
}

func initHandlers() *chi.Mux {
	r := chi.NewRouter()
	r.Post("/", func(res http.ResponseWriter, req *http.Request) {
		handlers.ShortHandler(res, req, urls, serverConfig)
	})
	r.Get("/{shorturl}", func(res http.ResponseWriter, req *http.Request) {
		handlers.ExpandHandler(res, req, urls)
	})
	return r
}

func main() {
	Init(parseFlagsAndEnv())
}
