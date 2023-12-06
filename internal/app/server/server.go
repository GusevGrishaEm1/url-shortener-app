package server

import (
	"io"
	"math/rand"
	"net/http"
	"strings"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var urls map[string]string = make(map[string]string)

func Init() {
	mux := initHandlers()
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}

func initHandlers() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", ShortenerHandler)
	return mux
}

func ShortenerHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
		} else {
			bodyStr := string(body)
			shortURL := GetShortURL()
			urls[shortURL] = bodyStr
			res.Header().Add("content-type", "text/plain")
			res.WriteHeader(http.StatusCreated)
			res.Write([]byte("http://localhost:8080/" + shortURL))
		}
	} else if req.Method == http.MethodGet {
		urlsParts := strings.Split(req.URL.Path, "/")
		shortURL := urlsParts[len(urlsParts)-1]
		originalURL, ok := urls[shortURL]
		if !ok {
			res.WriteHeader(http.StatusBadRequest)
		} else {
			res.Header().Add("Location", originalURL)
			res.WriteHeader(http.StatusTemporaryRedirect)
		}
	} else {
		res.WriteHeader(http.StatusBadRequest)
	}
}

func GetShortURL() string {
	shortURL := make([]byte, 5)
	var uniqueShortURL string
	for {
		for i := range shortURL {
			shortURL[i] = letterBytes[rand.Intn(len(letterBytes))]
		}
		uniqueShortURL = string(shortURL)
		if _, ok := urls[uniqueShortURL]; !ok {
			return uniqueShortURL
		}
	}
}
