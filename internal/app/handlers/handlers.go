package handlers

import (
	"io"
	"net/http"

	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/config"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/util"
)

func ShortHandler(res http.ResponseWriter, req *http.Request, urls map[string]string, serverConfig *config.Config) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
	} else {
		bodyStr := string(body)
		if bodyStr != "" {
			shortURL := util.GetShortURL()
			for _, ok := urls[shortURL]; ok; {
				shortURL = util.GetShortURL()
			}
			urls[shortURL] = bodyStr
			res.Header().Add("content-type", "text/plain")
			res.WriteHeader(http.StatusCreated)
			res.Write([]byte(serverConfig.GetBaseReturnURL() + "/" + shortURL))
		} else {
			res.WriteHeader(http.StatusBadRequest)
		}
	}
}

func ExpandHandler(res http.ResponseWriter, req *http.Request, urls map[string]string) {
	shortURL := req.URL.Path[1:]
	originalURL, ok := urls[shortURL]
	if !ok {
		res.WriteHeader(http.StatusBadRequest)
	} else {
		res.Header().Add("Location", originalURL)
		res.WriteHeader(http.StatusTemporaryRedirect)
	}
}
