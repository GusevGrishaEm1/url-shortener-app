package server

import (
	"io"
	"math/rand"
	"net/http"

	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/server/config"
	"github.com/go-chi/chi/v5"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var urls map[string]string = make(map[string]string)

var returnURL string = "http://localhost:8080/"

func Init(config *config.Config) {
	mux := initHandlers()
	returnURL = config.GetBaseReturnURL()
	err := http.ListenAndServe(config.GetServerURL(), mux)
	if err != nil {
		panic(err)
	}
}

func initHandlers() *chi.Mux {
	r := chi.NewRouter()
	r.Post("/", ShortHandler)
	r.Get("/{shorturl}", ExpandHandler)
	return r
}

func ShortHandler(res http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
	} else {
		bodyStr := string(body)
		if bodyStr != "" {
			shortURL := getShortURL()
			for _, ok := urls[shortURL]; ok; {
				shortURL = getShortURL()
			}
			urls[shortURL] = bodyStr
			res.Header().Add("content-type", "text/plain")
			res.WriteHeader(http.StatusCreated)
			res.Write([]byte(returnURL + shortURL))
		} else {
			res.WriteHeader(http.StatusBadRequest)
		}
	}
}

func ExpandHandler(res http.ResponseWriter, req *http.Request) {
	shortURL := req.URL.Path[1:]
	originalURL, ok := urls[shortURL]
	if !ok {
		res.WriteHeader(http.StatusBadRequest)
	} else {
		res.Header().Add("Location", originalURL)
		res.WriteHeader(http.StatusTemporaryRedirect)
	}
}

func getShortURL() string {
	shortURL := make([]byte, 5)
	for i := range shortURL {
		shortURL[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(shortURL)
}
