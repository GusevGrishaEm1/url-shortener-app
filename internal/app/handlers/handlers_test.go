package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShortenHandler(t *testing.T) {
	config := config.GetDefault()
	handlers := New(config)
	tests := []struct {
		name           string
		url            string
		expectedStatus int
	}{
		{
			name:           "test#1",
			url:            "https://practicum.yandex.ru/",
			expectedStatus: 201,
		},
		{
			name:           "test#2",
			url:            "",
			expectedStatus: 400,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(test.url)))
			w := httptest.NewRecorder()
			handlers.ShortenHandler(w, request)
			res := w.Result()
			defer res.Body.Close()
			equals := assert.Equal(t, test.expectedStatus, res.StatusCode)
			if equals && test.expectedStatus == 201 {
				resBody, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				partsURL := strings.Split(string(resBody), "/")
				shortURL := partsURL[len(partsURL)-1]
				assert.Equal(t, config.GetBaseReturnURL()+"/"+shortURL, string(resBody))
			}
		})
	}
}

func TestExpandHandler(t *testing.T) {
	config := config.GetDefault()
	handlers := New(config)
	originalURL := "https://gophercises.com/#signup"
	savedShortURL := initShortURLSForExpandHandler(handlers, originalURL)
	tests := []struct {
		name           string
		shortURL       string
		expectedStatus int
	}{
		{
			name:           "test#1",
			shortURL:       savedShortURL,
			expectedStatus: 307,
		},
		{
			name:           "test#2",
			shortURL:       "",
			expectedStatus: 400,
		},
	}
	for _, test := range tests {
		request := httptest.NewRequest(http.MethodGet, "/"+test.shortURL, nil)
		w := httptest.NewRecorder()
		handlers.ExpandHandler(w, request)
		res := w.Result()
		defer res.Body.Close()
		equals := assert.Equal(t, test.expectedStatus, res.StatusCode)
		if equals && test.expectedStatus == 307 {
			assert.Equal(t, originalURL, res.Header.Get("Location"))
		}
	}
}

func initShortURLSForExpandHandler(handlers ShortenerHandler, originalURL string) string {
	request := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(originalURL)))
	w := httptest.NewRecorder()
	handlers.ShortenHandler(w, request)
	res := w.Result()
	defer res.Body.Close()
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return ""
	}
	partsURL := strings.Split(string(resBody), "/")
	return partsURL[len(partsURL)-1]
}

func TestShortenJSONHandler(t *testing.T) {
	config := config.GetDefault()
	handlers := New(config)
	tests := []struct {
		name            string
		reqBody         []byte
		expectedResBody string
		expectedStatus  int
	}{
		{
			name:            "test#1",
			reqBody:         []byte(`{"url":"https://practicum.yandex.ru/"}`),
			expectedResBody: `{"result":%s}`,
			expectedStatus:  201,
		},
		{
			name:           "test#2",
			reqBody:        []byte(`{"ggg":"https://practicum.yandex.ru/"}`),
			expectedStatus: 400,
		},
		{
			name:           "test#3",
			reqBody:        []byte(``),
			expectedStatus: 400,
		},
		{
			name:           "test#4",
			reqBody:        []byte(`{"url":""}`),
			expectedStatus: 400,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewReader(test.reqBody))
			request.Header.Set("content-type", "application/json")
			w := httptest.NewRecorder()
			handlers.ShortenJSONHandler(w, request)
			res := w.Result()
			defer res.Body.Close()
			equals := assert.Equal(t, test.expectedStatus, res.StatusCode)
			if equals && test.expectedStatus == 201 {
				var err error
				resBody, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				jsonMap := make(map[string]json.RawMessage)
				err = json.Unmarshal(resBody, &jsonMap)
				require.NoError(t, err)
				assert.Equal(t, fmt.Sprintf(test.expectedResBody, string(jsonMap["result"]))+"\n", string(resBody))
			}
		})
	}
}

// func TestShortenerHandlers(t *testing.T) {
// 	handlers := New(config.GetDefault())
// 	tests := []struct {
// 		name               string
// 		originalURL        string
// 		expectedStatusPost int
// 		expectedStatusGet  int
// 		shortURL           string
// 	}{
// 		{
// 			name:               "test#1",
// 			originalURL:        "https://practicum.yandex.ru/",
// 			expectedStatusPost: 201,
// 			expectedStatusGet:  307,
// 		},
// 		{
// 			name:               "test#2",
// 			expectedStatusPost: 400,
// 			expectedStatusGet:  400,
// 		},
// 	}
// 	host := "http://localhost:8080/"
// 	for _, test := range tests {
// 		t.Run(test.name+" POST", func(t *testing.T) {
// 			request := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(test.originalURL)))
// 			w := httptest.NewRecorder()
// 			handlers.ShortenHandler(w, request)
// 			res := w.Result()
// 			defer res.Body.Close()
// 			assert.Equal(t, test.expectedStatusPost, res.StatusCode)
// 			if test.expectedStatusPost == 201 {
// 				resBody, err := io.ReadAll(res.Body)
// 				require.NoError(t, err)
// 				partsURL := strings.Split(string(resBody), "/")
// 				test.shortURL = partsURL[len(partsURL)-1]
// 				assert.Equal(t, host+test.shortURL, string(resBody))
// 			}
// 		})
// 		t.Run(test.name+" GET", func(t *testing.T) {
// 			request := httptest.NewRequest(http.MethodGet, "/"+test.shortURL, nil)
// 			w := httptest.NewRecorder()
// 			handlers.ExpandHandler(w, request)
// 			res := w.Result()
// 			defer res.Body.Close()
// 			assert.Equal(t, test.expectedStatusGet, res.StatusCode)
// 			if test.expectedStatusGet == 307 {
// 				assert.Equal(t, test.originalURL, res.Header.Get("Location"))
// 			}
// 		})
// 	}
// }
