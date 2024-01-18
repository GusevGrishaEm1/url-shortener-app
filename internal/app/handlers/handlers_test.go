package handlers

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	gzipreq "github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/gzip"

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
				assert.Equal(t, config.BaseReturnURL+"/"+shortURL, string(resBody))
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
		statusValid := assert.Equal(t, test.expectedStatus, res.StatusCode)
		if statusValid && test.expectedStatus == 307 {
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
			request := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(test.reqBody))
			request.Header.Set("content-type", "application/json")
			w := httptest.NewRecorder()
			handlers.ShortenJSONHandler(w, request)
			res := w.Result()
			defer res.Body.Close()
			statusValid := assert.Equal(t, test.expectedStatus, res.StatusCode)
			if statusValid && test.expectedStatus == http.StatusCreated {
				resBody, jsonMap := readJSON(res, t)
				test.expectedResBody = fmt.Sprintf(test.expectedResBody, string(jsonMap["result"]))
				assert.JSONEq(t, test.expectedResBody, string(resBody))
			}
		})
	}
}

func TestGzipCompression(t *testing.T) {
	handlers := New(config.GetDefault())
	handler := gzipreq.RequestZipper(handlers.ShortenJSONHandler)
	reqBody := `{"url":"https://practicum.yandex.ru/"}`
	expectedResBody := `{"result":%s}`
	t.Run("gzip_send", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		zb := gzip.NewWriter(buf)
		_, err := zb.Write([]byte(reqBody))
		require.NoError(t, err)
		err = zb.Close()
		require.NoError(t, err)
		r := httptest.NewRequest("POST", "/api/shorten", buf)
		r.Header.Set("Content-Encoding", "gzip")
		w := httptest.NewRecorder()
		handler(w, r)
		res := w.Result()
		defer res.Body.Close()
		statusValid := assert.Equal(t, http.StatusCreated, res.StatusCode)
		if statusValid {
			resBody, jsonMap := readJSON(res, t)
			expectedResBody = fmt.Sprintf(expectedResBody, string(jsonMap["result"]))
			assert.JSONEq(t, expectedResBody, string(resBody))
		}
	})
}

func readJSON(res *http.Response, t *testing.T) ([]byte, map[string]json.RawMessage) {
	var err error
	resBody, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	jsonMap := make(map[string]json.RawMessage)
	err = json.Unmarshal(resBody, &jsonMap)
	require.NoError(t, err)
	return resBody, jsonMap
}

func TestGzipDecompression(t *testing.T) {
	handlers := New(config.GetDefault())
	handler := gzipreq.RequestZipper(handlers.ShortenJSONHandler)
	reqBody := `{"url":"https://practicum.yandex.ru/"}`
	expectedResBody := `{"result":%s}`
	t.Run("gzip_recive", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader([]byte(reqBody)))
		r.Header.Set("Accept-Encoding", "gzip")
		w := httptest.NewRecorder()
		handler(w, r)
		res := w.Result()
		defer res.Body.Close()
		statusValid := assert.Equal(t, http.StatusCreated, res.StatusCode)
		if statusValid {
			resBody, jsonMap := readJSONGzip(res, t)
			expectedResBody = fmt.Sprintf(expectedResBody, string(jsonMap["result"]))
			assert.JSONEq(t, expectedResBody, string(resBody))
		}
	})
}

func readJSONGzip(res *http.Response, t *testing.T) ([]byte, map[string]json.RawMessage) {
	var err error
	gzipResBody, err := gzip.NewReader(res.Body)
	require.NoError(t, err)
	resBody, err := io.ReadAll(gzipResBody)
	require.NoError(t, err)
	jsonMap := make(map[string]json.RawMessage)
	err = json.Unmarshal(resBody, &jsonMap)
	require.NoError(t, err)
	return resBody, jsonMap
}

// func TestPingDBHandler(t *testing.T) {
// 	handlers := New(config.GetDefault())
// 	handler := handlers.PingDBHandler
// 	t.Run("ping db", func(t *testing.T) {
// 		r := httptest.NewRequest(http.MethodGet, "/ping", nil)
// 		w := httptest.NewRecorder()
// 		handler(w, r)
// 		res := w.Result()
// 		defer res.Body.Close()
// 		assert.Equal(t, http.StatusOK, res.StatusCode)
// 	})
// }
