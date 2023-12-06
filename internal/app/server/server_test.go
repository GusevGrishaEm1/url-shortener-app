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

func TestShortenerHandler(t *testing.T) {
	host := "http://localhost:8080/"
	originalURL := "https://practicum.yandex.ru/"
	var shortURL string
	t.Run("test POST", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(originalURL)))
		w := httptest.NewRecorder()
		ShortenerHandler(w, request)
		res := w.Result()
		assert.Equal(t, 201, res.StatusCode)
		defer res.Body.Close()
		resBody, err := io.ReadAll(res.Body)
		require.NoError(t, err)
		partsURL := strings.Split(string(resBody), "/")
		shortURL = partsURL[len(partsURL)-1]
		assert.Equal(t, string(resBody), host+shortURL)
	})
	t.Run("test GET", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/"+shortURL, nil)
		w := httptest.NewRecorder()
		ShortenerHandler(w, request)
		res := w.Result()
		assert.Equal(t, 307, res.StatusCode)
		assert.Equal(t, res.Header.Get("Location"), originalURL)
	})
}
