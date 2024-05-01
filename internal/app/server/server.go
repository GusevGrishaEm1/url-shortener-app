// Package server реализует веб-сервер для обработки запросов по сокращению URL.
//
// В данном пакете определены интерфейсы для обработчиков URL-ов,
// промежуточных middleware и функция для запуска веб-сервера.
//
// Пример использования:
//
//	import "context"
//	import "github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/config"
//	import "github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/server"
//
//	func main() {
//		ctx := context.Background()
//		config := config.GetDefaultConfig() // Получение конфигурации приложения
//		err := server.StartServer(ctx, config)
//		if err != nil {
//			log.Fatal("Server startup failed: ", err)
//		}
//	}
package server

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/config"
	gzipreq "github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/gzip"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/handlers"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/logger"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/security"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/service"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/storage"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

// ShortenerHandler определяет методы для обработки запросов по сокращению URL.
type ShortenerHandler interface {
	// ShortenHandler обрабатывает запрос на сокращение URL и возвращает сокращенный URL.
	ShortenHandler(res http.ResponseWriter, req *http.Request)
	// ShortenJSONHandler обрабатывает запрос на сокращение URL в формате JSON и возвращает сокращенный URL.
	ShortenJSONHandler(res http.ResponseWriter, req *http.Request)
	// ShortenJSONBatchHandler обрабатывает запрос на сокращение нескольких URL в формате JSON и возвращает сокращенные URL.
	ShortenJSONBatchHandler(res http.ResponseWriter, req *http.Request)
	// ExpandHandler обрабатывает запрос на расширение сокращенного URL и выполняет редирект на исходный URL.
	ExpandHandler(res http.ResponseWriter, req *http.Request)
	// PingStorageHandler обрабатывает запрос на проверку доступности хранилища.
	PingStorageHandler(res http.ResponseWriter, req *http.Request)
	// UrlsByUserHandler обрабатывает запрос на получение списка URL, созданных пользователем.
	UrlsByUserHandler(res http.ResponseWriter, req *http.Request)
	// DeleteUrlsHandler обрабатывает запрос на удаление списка URL, созданных пользователем.
	DeleteUrlsHandler(res http.ResponseWriter, req *http.Request)
}

// SecurityMiddleware определяет middleware для обеспечения безопасности.
type SecurityMiddleware interface {
	// RequiredUserID проверяет наличие идентификатора пользователя в запросе.
	RequiredUserID(h http.Handler) http.Handler
	// Security обеспечивает безопасность обработки HTTP-запросов.
	Security(h http.Handler) http.Handler
}

// CompressionMiddleware определяет middleware для decode/encode HTTP-ответов
type CompressionMiddleware interface {
	// Compression выполняет encode HTTP-ответов.
	// и decode HTTP-запросов.
	Compression(h http.Handler) http.Handler
}

// StartServer запускает веб-сервер для обработки http запросов.
// Он инициализирует необходимые хранилища, сервисы и middleware, а также определяет маршруты для хендлеров.
func StartServer(ctx context.Context, config config.Config) error {
	if err := logger.Init(slog.LevelInfo); err != nil {
		return err
	}

	storage, err := storage.NewShortenerStorage(storage.GetStorageTypeByConfig(config), config)
	if err != nil {
		return err
	}
	service, err := service.NewShortenerService(ctx, config, storage)
	if err != nil {
		return err
	}
	compress := gzipreq.NewCompressionMiddleware()
	security := security.NewSecurityMiddleware(service)
	handler := handlers.NewShortenerHandler(config, service)
	handlersAndMiddlewares := handlersAndMiddlewares{
		handler,
		security,
		compress,
	}

	mux := getMux(handlersAndMiddlewares)
	if config.EnableHTTPS {
		http.ListenAndServeTLS(config.ServerURL, "conf.pem", "key.pem", mux)
		return err
	}
	err = http.ListenAndServe(config.ServerURL, mux)
	return err
}

type handlersAndMiddlewares struct {
	ShortenerHandler
	SecurityMiddleware
	CompressionMiddleware
}

func getMux(ham handlersAndMiddlewares) *chi.Mux {
	r := chi.NewRouter()

	r.Use(ham.Security)
	r.Use(ham.Compression)
	r.Use(logger.RequestLogger)

	r.Mount("/", middleware.Profiler())

	r.Post("/", ham.ShortenHandler)
	r.Get("/{shorturl}", ham.ExpandHandler)
	r.Post("/api/shorten", ham.ShortenJSONHandler)
	r.Post("/api/shorten/batch", ham.ShortenJSONBatchHandler)
	r.Get("/ping", ham.PingStorageHandler)

	r.Group(func(r chi.Router) {
		r.Use(ham.RequiredUserID)
		r.Get("/api/user/urls", ham.UrlsByUserHandler)
		r.Delete("/api/user/urls", ham.DeleteUrlsHandler)
	})

	return r
}
