package service

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// --- Вспомогательный мок для проверки Close у CompressReader ---

type mockReadCloser struct {
	io.Reader
	closed bool
}

func (m *mockReadCloser) Read(p []byte) (int, error) {
	return m.Reader.Read(p)
}

func (m *mockReadCloser) Close() error {
	m.closed = true
	return nil
}

// --- CompressWriter ---

func TestNewCompressWriter_Basic(t *testing.T) {
	rr := httptest.NewRecorder()

	cw := NewCompressWriter(rr)

	if cw.ResponseWriter != rr {
		t.Errorf("NewCompressWriter: ResponseWriter not set correctly")
	}
	if cw.gz == nil {
		t.Fatalf("NewCompressWriter: gz writer is nil")
	}
}

func TestCompressWriter_WriteHeader_SetsGzipOn2xx(t *testing.T) {
	rr := httptest.NewRecorder()
	cw := NewCompressWriter(rr)

	cw.WriteHeader(http.StatusOK)

	if got := rr.Header().Get("Content-Encoding"); got != "gzip" {
		t.Errorf("Content-Encoding = %q, want %q", got, "gzip")
	}
	if rr.Code != http.StatusOK {
		t.Errorf("status code = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestCompressWriter_WriteHeader_DoesNotSetGzipOnErrorStatus(t *testing.T) {
	rr := httptest.NewRecorder()
	cw := NewCompressWriter(rr)

	cw.WriteHeader(http.StatusBadRequest)

	if got := rr.Header().Get("Content-Encoding"); got != "" {
		t.Errorf("Content-Encoding = %q, want empty", got)
	}
	if rr.Code != http.StatusBadRequest {
		t.Errorf("status code = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestCompressWriter_FullFlow_WriteAndClose(t *testing.T) {
	rr := httptest.NewRecorder()
	cw := NewCompressWriter(rr)

	body := []byte("hello gzip")

	cw.WriteHeader(http.StatusOK)
	if _, err := cw.Write(body); err != nil {
		t.Fatalf("Write error: %v", err)
	}
	if err := cw.Close(); err != nil {
		t.Fatalf("Close error: %v", err)
	}

	// Проверяем, что контент реально gzipped и раскодируется обратно
	gzr, err := gzip.NewReader(bytes.NewReader(rr.Body.Bytes()))
	if err != nil {
		t.Fatalf("gzip.NewReader error: %v", err)
	}
	defer gzr.Close()

	decoded, err := io.ReadAll(gzr)
	if err != nil {
		t.Fatalf("io.ReadAll error: %v", err)
	}

	if string(decoded) != string(body) {
		t.Errorf("decoded body = %q, want %q", decoded, body)
	}
}

func TestCompressWriter_Close(t *testing.T) {
	rr := httptest.NewRecorder()
	cw := NewCompressWriter(rr)

	if err := cw.Close(); err != nil {
		t.Fatalf("Close() error = %v, want nil", err)
	}
}

// --- CompressReader ---

func TestNewCompressReader_Success(t *testing.T) {
	// готовим gzipped данные
	var buf bytes.Buffer
	gzw := gzip.NewWriter(&buf)
	_, err := gzw.Write([]byte("hello"))
	if err != nil {
		t.Fatalf("gzip.Write error: %v", err)
	}
	if err := gzw.Close(); err != nil {
		t.Fatalf("gzip.Close error: %v", err)
	}

	r := io.NopCloser(bytes.NewReader(buf.Bytes()))

	cr, err := NewCompressReader(r)
	if err != nil {
		t.Fatalf("NewCompressReader error: %v", err)
	}
	if cr == nil {
		t.Fatalf("NewCompressReader returned nil reader")
	}

	out, err := io.ReadAll(cr)
	if err != nil {
		t.Fatalf("ReadAll error: %v", err)
	}

	if string(out) != "hello" {
		t.Errorf("decoded = %q, want %q", out, "hello")
	}

	if err := cr.Close(); err != nil {
		t.Fatalf("Close error: %v", err)
	}
}

func TestNewCompressReader_InvalidData(t *testing.T) {
	r := io.NopCloser(bytes.NewReader([]byte("not a gzip stream")))
	cr, err := NewCompressReader(r)
	if err == nil {
		t.Fatalf("expected error for invalid gzip data, got nil")
	}
	if cr != nil {
		t.Fatalf("expected nil reader on error, got %v", cr)
	}
}

func TestCompressReader_Read(t *testing.T) {
	// gzipped "test-data"
	var buf bytes.Buffer
	gzw := gzip.NewWriter(&buf)
	if _, err := gzw.Write([]byte("test-data")); err != nil {
		t.Fatalf("gzip.Write error: %v", err)
	}
	if err := gzw.Close(); err != nil {
		t.Fatalf("gzip.Close error: %v", err)
	}

	r := io.NopCloser(bytes.NewReader(buf.Bytes()))
	cr, err := NewCompressReader(r)
	if err != nil {
		t.Fatalf("NewCompressReader error: %v", err)
	}
	defer cr.Close()

	p := make([]byte, 4)
	n, err := cr.Read(p)
	if err != nil && err != io.EOF {
		t.Fatalf("Read error: %v", err)
	}

	if n == 0 {
		t.Fatalf("Read returned 0 bytes, expected > 0")
	}
}

func TestCompressReader_Close_ClosesBothReaders(t *testing.T) {
	// gzipped пустое тело, нам важен только Close
	var buf bytes.Buffer
	gzw := gzip.NewWriter(&buf)
	if err := gzw.Close(); err != nil {
		t.Fatalf("gzip.Close error: %v", err)
	}

	rc := &mockReadCloser{Reader: bytes.NewReader(buf.Bytes())}
	gzr, err := gzip.NewReader(rc)
	if err != nil {
		t.Fatalf("gzip.NewReader error: %v", err)
	}

	cr := &CompressReader{
		ReadCloser: rc,
		gz:         gzr,
	}

	if err := cr.Close(); err != nil {
		t.Fatalf("CompressReader.Close error: %v", err)
	}

	if !rc.closed {
		t.Fatalf("underlying ReadCloser was not closed")
	}
}

func TestNewCompressWriter_IntegrationWithHTTP(t *testing.T) {
	// мини-интеграционный тест: handler, который пишет gzipped ответ
	handler := func(w http.ResponseWriter, r *http.Request) {
		cw := NewCompressWriter(w)
		defer cw.Close()

		cw.WriteHeader(http.StatusOK)
		_, _ = cw.Write([]byte("ping"))
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	handler(rr, req)

	if rr.Header().Get("Content-Encoding") != "gzip" {
		t.Errorf("Content-Encoding = %q, want gzip", rr.Header().Get("Content-Encoding"))
	}

	gzr, err := gzip.NewReader(bytes.NewReader(rr.Body.Bytes()))
	if err != nil {
		t.Fatalf("gzip.NewReader error: %v", err)
	}
	defer gzr.Close()

	data, err := io.ReadAll(gzr)
	if err != nil {
		t.Fatalf("io.ReadAll error: %v", err)
	}
	if string(data) != "ping" {
		t.Errorf("decoded body = %q, want %q", data, "ping")
	}
}
