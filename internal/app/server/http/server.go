// Package http реализует веб-сервер для обработки запросов по сокращению URL.
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
package http

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/config"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/logger"
	gzipreq "github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/server/http/middlewares/gzip"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/server/http/middlewares/security"
	"github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/server/http/middlewares/trustedsubnet"

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
	// StatsHandler возвращающий в ответ объект статистики
	StatsHandler(res http.ResponseWriter, req *http.Request)
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

// InternalMiddleware определяет middleware для обработки запросов по trusted subnet.
type TrustedSubnetMiddleware interface {
	// Internal ограничивает обработку запросов по trusted subnet.
	TrustedSubnet(h http.Handler) http.Handler
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
	ctx, cancel := context.WithCancel(ctx)
	var wg sync.WaitGroup
	service, err := service.NewShortenerServiceWithWorkers(ctx, config, storage, &wg)
	if err != nil {
		cancel()
		return err
	}
	compress := gzipreq.NewCompressionMiddleware()
	security := security.NewSecurityMiddleware(service)
	subnet := trustedsubnet.NewTrustedSubnetMiddleware(config)
	handler := NewShortenerHandler(config, service)
	handlersAndMiddlewares := handlersAndMiddlewares{
		handler,
		security,
		compress,
		subnet,
	}

	mux := getMux(handlersAndMiddlewares)
	srv := &http.Server{Handler: mux, Addr: config.ServerURL}
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	idleConnsClosed := make(chan struct{})
	go func() {
		<-sigs
		err := srv.Shutdown(ctx)
		if err != nil {
			log.Fatal(err)
		}
		close(idleConnsClosed)
	}()

	if config.EnableHTTPS {
		err = srv.ListenAndServeTLS("server.crt", "server.key")
		<-idleConnsClosed
		cancel()
		wg.Wait()
		return err
	}
	err = srv.ListenAndServe()
	<-idleConnsClosed
	cancel()
	wg.Wait()
	return err
}

type handlersAndMiddlewares struct {
	ShortenerHandler
	SecurityMiddleware
	CompressionMiddleware
	TrustedSubnetMiddleware
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
		r.Use(ham.TrustedSubnet)
		r.Get("/api/internal/stats", ham.StatsHandler)
	})

	r.Group(func(r chi.Router) {
		r.Use(ham.RequiredUserID)
		r.Get("/api/user/urls", ham.UrlsByUserHandler)
		r.Delete("/api/user/urls", ham.DeleteUrlsHandler)
	})

	return r
}
