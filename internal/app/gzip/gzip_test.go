package gzip

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCompression(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	respRecorder := httptest.NewRecorder()

	middleware := NewCompressionMiddleware()
	handler := middleware.Compression(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("test data"))
		w.WriteHeader(http.StatusOK)
	}))
	handler.ServeHTTP(respRecorder, req)

	if respRecorder.Header().Get("Content-Encoding") != "gzip" {
		t.Errorf("Expected Content-Encoding: gzip, got: %s", respRecorder.Header().Get("Content-Encoding"))
	}

	gz, err := gzip.NewReader(respRecorder.Body)
	require.NoError(t, err)
	defer gz.Close()

	var buf bytes.Buffer
	_, err = buf.ReadFrom(gz)
	require.NoError(t, err)
	if !strings.Contains(buf.String(), "test data") {
		t.Errorf("Expected response body to contain 'test data', got: %s", buf.String())
	}
}

func TestDecompression(t *testing.T) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	_, err := gz.Write([]byte("test data"))
	require.NoError(t, err)
	err = gz.Close()
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/", &buf)
	req.Header.Set("Content-Encoding", "gzip")
	respRecorder := httptest.NewRecorder()

	middleware := NewCompressionMiddleware()
	handler := middleware.Compression(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		if !strings.Contains(string(data), "test data") {
			http.Error(w, "Request body does not contain 'test data'", http.StatusBadRequest)
			return
		}
	}))
	handler.ServeHTTP(respRecorder, req)

	if respRecorder.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got: %d", http.StatusOK, respRecorder.Code)
	}
}
