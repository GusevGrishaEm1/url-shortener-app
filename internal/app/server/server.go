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

	mux.HandleFunc("/", mainHandler)

	return mux
}

func getShortUrl() string {
	shortUrl := make([]byte, 5)
	var uniqueShortUrl string
	for {
		for i := range shortUrl {
			shortUrl[i] = letterBytes[rand.Intn(len(letterBytes))]
		}
		uniqueShortUrl = string(shortUrl)
		if _, ok := urls[uniqueShortUrl]; !ok {
			return uniqueShortUrl
		}
	}
}

func mainHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" && req.Header.Get("content-type") == "text/plain" {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
		}
		bodyStr := string(body)
		shortUrl := getShortUrl()
		urls[shortUrl] = bodyStr
		res.WriteHeader(http.StatusCreated)
		res.Header().Set("content-type", "text/plain")
		res.Write([]byte("http://localhost:8080/" + shortUrl))
	} else if req.Method == "GET" {
		urlsParts := strings.Split(req.URL.Path, "/")
		shortUrl := urlsParts[len(urlsParts)-1]
		originalUrl, ok := urls[shortUrl]
		if !ok {
			res.WriteHeader(http.StatusBadRequest)
		} else {
			res.Header().Add("Location", originalUrl)
			res.WriteHeader(http.StatusTemporaryRedirect)
		}
	} else {
		res.WriteHeader(http.StatusBadRequest)
	}
}
