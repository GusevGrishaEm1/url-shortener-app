package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
)

func (handler *shortenerHandler) ExampleShortenHandler() {
	// ctx := context.Background()
	// config := config.GetDefault()
	// // Инициализируем хранилище URL-ов.
	// storage, err := storage.NewShortenerStorage(storage.GetStorageTypeByConfig(config), config)
	// if err != nil {
	//     panic(err)
	// }
	// // Инициализируем сервис.
	// service, err := service.NewShortenerService(
	// 	ctx,
	// 	config,
	// 	storage,
	// )
	// if err != nil {
	//     panic(err)
	// }
	// // Инициализируем обработчик.
	// handler := NewShortenerHandler(
	// 	config,
	// 	service,
	// )

	// Создание запроса на сокращение URL
	reqBody := strings.NewReader(`{"url": "https://example.com"}`)
	req := httptest.NewRequest(http.MethodPost, "/", reqBody)
	resp := httptest.NewRecorder()

	// Обработка запроса
	handler.ShortenHandler(resp, req)

	// Вывод результата
	fmt.Println("Response:", resp.Body.String())
}

func (handler *shortenerHandler) ExampleExpandHandler() {
	// ctx := context.Background()
	// config := config.GetDefault()
	// // Инициализируем хранилище URL-ов.
	// storage, err := storage.NewShortenerStorage(storage.GetStorageTypeByConfig(config), config)
	// if err != nil {
	//     panic(err)
	// }
	// // Инициализируем сервис.
	// service, err := service.NewShortenerService(
	// 	ctx,
	// 	config,
	// 	storage,
	// )
	// if err != nil {
	//     panic(err)
	// }
	// // Инициализируем обработчик.
	// handler := NewShortenerHandler(
	// 	config,
	// 	service,
	// )

	// Создание запроса на расширение URL
	req := httptest.NewRequest(http.MethodGet, "/abc123", nil)
	resp := httptest.NewRecorder()

	// Обработка запроса
	handler.ExpandHandler(resp, req)

	// Вывод результата
	fmt.Println("Response:", resp.Body.String())
}
