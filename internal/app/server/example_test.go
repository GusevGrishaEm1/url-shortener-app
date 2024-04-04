package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/config"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/handlers"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/service"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/storage"
)

func ExampleStartServer() {
	handler, _ := getHandler()

	// Создание запроса на сокращение URL
	reqBody := strings.NewReader(`{"url": "https://example.com"}`)
	req := httptest.NewRequest(http.MethodPost, "/", reqBody)
	resp := httptest.NewRecorder()

	// Обработка запроса
	handler.ShortenJSONHandler(resp, req)
	var data map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &data)
	if err != nil {
		fmt.Println("Ошибка при декодировании JSON:", err)
		return
	}
	url := data["result"].(string)
	tokens := strings.Split(url, "/")
	token := tokens[len(tokens)-1]

	req = httptest.NewRequest(http.MethodGet, "/"+token, nil)
	resp = httptest.NewRecorder()
	handler.ExpandHandler(resp, req)
	fmt.Println(resp.Header().Get("Location"))
	// Output: https://example.com
}

func getHandler() (ShortenerHandler, error) {
	config := config.GetDefault()
	ctx := context.Background()
	storage, err := storage.NewShortenerStorage(storage.GetStorageTypeByConfig(config), config)
	if err != nil {
		return nil, err
	}
	service, err := service.NewShortenerService(ctx, config, storage)
	if err != nil {
		return nil, err
	}
	handler := handlers.NewShortenerHandler(config, service)
	return handler, nil
}
