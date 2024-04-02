// Пакет gzip предоставляет промежуточное ПО для сжатия и разжатия данных с использованием gzip.
//
// gzipMiddleware является промежуточным ПО для сжатия и разжатия данных. Он применяет сжатие gzip к ответам HTTP,
// если заголовок запроса Accept-Encoding содержит "gzip", и разжимает тело запроса, если заголовок запроса Content-Encoding содержит "gzip".
// Пакет также предоставляет структуры compressWriter и decompressReader для обеспечения записи сжатых данных и чтения разжатых данных соответственно.
//
// Пример использования:
//
//	import (
//	    "net/http"
//	    "github.com/GusevGrishaEm1/url-shortener-app.git/internal/app/gzip"
//	)
//
//	func main() {
//	    mux := http.NewServeMux()
//	    mux.HandleFunc("/", myHandler)
//
//	    compressionMiddleware := gzip.NewCompressionMiddleware()
//	    compressedMux := compressionMiddleware.Compression(mux)
//
//	    http.ListenAndServe(":8080", compressedMux)
//	}
package gzip

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type gzipMiddleware struct{}

// Compression возвращает обработчик HTTP, который сжимает или разжимает данные в зависимости от заголовков запроса и ответа.
func (*gzipMiddleware) Compression(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		responseWriter := w
		acceptEncoding := r.Header.Get("Accept-Encoding")
		if strings.Contains(acceptEncoding, "gzip") {
			cw := newCompressWriter(w)
			responseWriter = cw
			defer cw.Close()
		}
		contentEncoding := r.Header.Get("Content-Encoding")
		if strings.Contains(contentEncoding, "gzip") {
			dr, err := newDecompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = dr
		}
		h.ServeHTTP(responseWriter, r)
	})
}

// NewCompressionMiddleware создает новый экземпляр промежуточного ПО для сжатия и разжатия данных.
func NewCompressionMiddleware() *gzipMiddleware {
	return &gzipMiddleware{}
}

type compressWriter struct {
	rw  http.ResponseWriter
	gzw *gzip.Writer
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		rw:  w,
		gzw: gzip.NewWriter(w),
	}
}

// Header возвращает заголовки HTTP-ответа.
func (c *compressWriter) Header() http.Header {
	return c.rw.Header()
}

// Write записывает данные в HTTP-ответ.
func (c *compressWriter) Write(data []byte) (int, error) {
	return c.gzw.Write(data)
}

// WriteHeader записывает статус HTTP-ответа.
func (c *compressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.rw.Header().Set("Content-Encoding", "gzip")
	}
	c.rw.WriteHeader(statusCode)
}

// Close закрывает HTTP-ответ.
func (c *compressWriter) Close() error {
	return c.gzw.Close()
}

type decompressReader struct {
	rc  io.ReadCloser
	gzr *gzip.Reader
}

func newDecompressReader(r io.ReadCloser) (*decompressReader, error) {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	return &decompressReader{
		rc:  r,
		gzr: gzr,
	}, nil
}

// Read возвращает данные из HTTP-запроса.
func (d *decompressReader) Read(p []byte) (n int, err error) {
	return d.gzr.Read(p)
}

// Close закрывает HTTP-запрос.
func (d *decompressReader) Close() error {
	if err := d.rc.Close(); err != nil {
		return err
	}
	return d.gzr.Close()
}
