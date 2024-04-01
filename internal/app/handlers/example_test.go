package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
)

func ExampleShortenHandler() {
	handler := &shortenerHandler{}

	// Создание запроса на сокращение URL
	reqBody := strings.NewReader(`{"url": "https://example.com"}`)
	req := httptest.NewRequest(http.MethodPost, "/", reqBody)
	resp := httptest.NewRecorder()

	// Обработка запроса
	handler.ShortenHandler(resp, req)

	// Вывод результата
	fmt.Println("Response:", resp.Body.String())
}

func ExampleExpandHandler() {
	handler := &shortenerHandler{}

	// Создание запроса на расширение URL
	req := httptest.NewRequest(http.MethodGet, "/abc123", nil)
	resp := httptest.NewRecorder()

	// Обработка запроса
	handler.ExpandHandler(resp, req)

	// Вывод результата
	fmt.Println("Response:", resp.Body.String())
}
