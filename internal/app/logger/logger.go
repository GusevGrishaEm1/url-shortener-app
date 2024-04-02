package logger

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	slogzap "github.com/samber/slog-zap/v2"
	"go.uber.org/zap"
)

// Logger представляет собой экземпляр логгера, используемого для записи логов.
var Logger = slog.New(&slogzap.ZapHandler{})

// Init инициализирует логгер с указанным уровнем логирования.
// Возвращает ошибку, если инициализация не удалась.
func Init(level slog.Level) error {
	zapLogger, err := zap.NewProduction()
	if err != nil {
		return err
	}
	Logger = slog.New(slogzap.Option{Level: level, Logger: zapLogger}.NewZapHandler())
	return nil
}

// RequestLogger возвращает middleware для логирования информации о HTTP-запросах.
func RequestLogger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			rw:           w,
			responseData: responseData,
		}
		h.ServeHTTP(&lw, r)
		duration := time.Since(start).Milliseconds()
		Logger.Info("request info",
			slog.Group("request",
				slog.String("uri", r.RequestURI),
				slog.String("method", r.Method),
				slog.String("duration", fmt.Sprintf("%d milliseconds", duration)),
				slog.String("client_ip", r.RemoteAddr),
			),
			slog.Group("response",
				slog.Int("status", lw.responseData.status),
				slog.Int("size", lw.responseData.size),
			),
		)
	})
}

type (
	responseData struct {
		status int
		size   int
	}
	loggingResponseWriter struct {
		rw           http.ResponseWriter
		responseData *responseData
	}
)

// WriteHeader записывает статус и размер ответа.
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.rw.Write(b)
	r.responseData.size += size
	return size, err
}

// WriteHeader записывает статус и размер ответа.
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.rw.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

// Header возвращает заголовки HTTP-запроса.
func (r *loggingResponseWriter) Header() http.Header {
	return r.rw.Header()
}
