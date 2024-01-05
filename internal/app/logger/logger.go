package logger

import (
	"log/slog"
	"net/http"
	"time"

	slogzap "github.com/samber/slog-zap/v2"
	"go.uber.org/zap"
)

var Logger = slog.New(&slogzap.ZapHandler{})

func Init(level slog.Level) error {
	zapLogger, err := zap.NewProduction()
	if err != nil {
		return err
	}
	Logger = slog.New(slogzap.Option{Level: level, Logger: zapLogger}.NewZapHandler())
	return nil
}

func RequestLogger(h http.HandlerFunc) http.HandlerFunc {
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
		h(&lw, r)
		Logger.
			With(
				slog.Group("request",
					slog.String("uri", r.RequestURI),
					slog.String("method", r.Method),
					slog.Duration("duration", time.Since(start)),
				),
				slog.Group("response",
					slog.Int("status", responseData.status),
					slog.Int("size", responseData.size),
				),
			).
			Info("request info")
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

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.rw.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.rw.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func (r *loggingResponseWriter) Header() http.Header {
	return r.rw.Header()
}
