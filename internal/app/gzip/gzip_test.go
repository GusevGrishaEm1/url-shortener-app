package gzip

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCompression(t *testing.T) {
	// Создаем HTTP запрос и ответ для тестирования
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	respRecorder := httptest.NewRecorder()

	// Создаем экземпляр промежуточного ПО и обрабатываем запрос
	middleware := NewCompressionMiddleware()
	handler := middleware.Compression(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("test data"))
		w.WriteHeader(http.StatusOK)
	}))
	handler.ServeHTTP(respRecorder, req)

	// Проверяем, что ответ был сжат с помощью gzip
	if respRecorder.Header().Get("Content-Encoding") != "gzip" {
		t.Errorf("Expected Content-Encoding: gzip, got: %s", respRecorder.Header().Get("Content-Encoding"))
	}

	// Декодируем сжатые данные
	gz, err := gzip.NewReader(respRecorder.Body)
	if err != nil {
		t.Fatal("Error creating gzip reader:", err)
	}
	defer gz.Close()

	// Считываем и проверяем расжатые данные
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(gz); err != nil {
		t.Fatal("Error reading from gzip reader:", err)
	}
	if !strings.Contains(buf.String(), "test data") {
		t.Errorf("Expected response body to contain 'test data', got: %s", buf.String())
	}
}

func TestDecompression(t *testing.T) {
	// Создаем сжатые данные
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	_, err := gz.Write([]byte("test data"))
	if err != nil {
		t.Fatal("Error writing to gzip writer:", err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal("Error closing gzip writer:", err)
	}

	// Создаем HTTP запрос с сжатыми данными и обрабатываем его
	req := httptest.NewRequest(http.MethodPost, "/", &buf)
	req.Header.Set("Content-Encoding", "gzip")
	respRecorder := httptest.NewRecorder()

	// Создаем экземпляр промежуточного ПО и обрабатываем запрос
	middleware := NewCompressionMiddleware()
	handler := middleware.Compression(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body", http.StatusInternalServerError)
			return
		}
		if !strings.Contains(string(data), "test data") {
			http.Error(w, "Request body does not contain 'test data'", http.StatusBadRequest)
			return
		}
	}))
	handler.ServeHTTP(respRecorder, req)

	// Проверяем, что обработчик получил расжатые данные
	if respRecorder.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got: %d", http.StatusOK, respRecorder.Code)
	}
}
