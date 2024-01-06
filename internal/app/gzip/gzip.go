package gzip

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

func RequestZipper(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
		h(responseWriter, r)
	}
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

func (c *compressWriter) Header() http.Header {
	return c.rw.Header()
}

func (c *compressWriter) Write(data []byte) (int, error) {
	return c.gzw.Write(data)
}

func (c *compressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.rw.Header().Set("Content-Encoding", "gzip")
	}
	c.rw.WriteHeader(statusCode)
}

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

func (d *decompressReader) Read(p []byte) (n int, err error) {
	return d.gzr.Read(p)
}

func (d *decompressReader) Close() error {
	if err := d.rc.Close(); err != nil {
		return err
	}
	return d.gzr.Close()
}
