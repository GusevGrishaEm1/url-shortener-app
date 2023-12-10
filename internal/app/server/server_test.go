package server

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShortenerHandlers(t *testing.T) {
	tests := []struct {
		name               string
		originalURL        string
		expectedStatusPost int
		expectedStatusGet  int
		shortURL           string
	}{
		{
			name:               "test#1",
			originalURL:        "https://practicum.yandex.ru/",
			expectedStatusPost: 201,
			expectedStatusGet:  307,
		},
		{
			name:               "test#2",
			expectedStatusPost: 400,
			expectedStatusGet:  400,
		},
	}
	host := "http://localhost:8080/"
	for _, test := range tests {
		t.Run(test.name+" POST", func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(test.originalURL)))
			w := httptest.NewRecorder()
			ShortHandler(w, request)
			res := w.Result()
			defer res.Body.Close()
			assert.Equal(t, test.expectedStatusPost, res.StatusCode)
			if test.expectedStatusPost == 201 {
				resBody, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				partsURL := strings.Split(string(resBody), "/")
				test.shortURL = partsURL[len(partsURL)-1]
				assert.Equal(t, host+test.shortURL, string(resBody))
			}
		})
		t.Run(test.name+" GET", func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/"+test.shortURL, nil)
			w := httptest.NewRecorder()
			ExpandHandler(w, request)
			res := w.Result()
			defer res.Body.Close()
			assert.Equal(t, test.expectedStatusGet, res.StatusCode)
			if test.expectedStatusGet == 307 {
				assert.Equal(t, test.originalURL, res.Header.Get("Location"))
			}
		})
	}
}
