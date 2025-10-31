package compresser

import (
	"compress/gzip"
	"io"
	"net/http"
)

type CompressWriter struct {
	http.ResponseWriter
	gz *gzip.Writer
}

func NewCompressWriter(w http.ResponseWriter) *CompressWriter {
	gz := gzip.NewWriter(w)
	return &CompressWriter{
		ResponseWriter: w,
		gz:             gz,
	}
}

func (c *CompressWriter) Write(b []byte) (int, error) {
	return c.gz.Write(b)
}

func (c *CompressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.Header().Set("Content-Encoding", "gzip")
	}
	c.ResponseWriter.WriteHeader(statusCode)
}

func (c *CompressWriter) Close() error {
	return c.gz.Close()
}

type CompressReader struct {
	io.ReadCloser // исходный r.Body
	gz            *gzip.Reader
}

func NewCompressReader(r io.ReadCloser) (*CompressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &CompressReader{
		ReadCloser: r,
		gz:         zr,
	}, nil
}

func (z *CompressReader) Read(p []byte) (int, error) {
	return z.gz.Read(p)
}

func (z *CompressReader) Close() error {
	if err := z.gz.Close(); err != nil {
		return err
	}

	return z.ReadCloser.Close()
}
