package handler

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	models "metrify/internal/model"
	"metrify/internal/service"
)

func newTestStorage() *service.MemStorage {
	f, _ := os.CreateTemp("", "memstorage-test-*.json")
	path := f.Name()
	f.Close()
	return service.NewMemStorage(path, nil)
}

func newTestHandler() (*Handler, *service.MemStorage) {
	ms := newTestStorage()
	logger := zap.NewNop().Sugar()
	h := NewHandler(ms, logger, nil, false, "")
	return h, ms
}

func ptrF(v float64) *float64 { return &v }
func ptrI(v int64) *int64     { return &v }

func TestHandler_GetGauge(t *testing.T) {
	h, ms := newTestHandler()
	ms.UpdateGauge("temp", 36.6)

	r := chi.NewRouter()
	r.Get("/value/gauge/{name}", h.GetGauge)

	req := httptest.NewRequest("GET", "/value/gauge/temp", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}

	if strings.TrimSpace(rr.Body.String()) != "36.6" {
		t.Fatalf("body = %q, want %q", rr.Body.String(), "36.6")
	}
}

func TestHandler_GetGauge_NotFound(t *testing.T) {
	h, _ := newTestHandler()

	r := chi.NewRouter()
	r.Get("/value/gauge/{name}", h.GetGauge)

	req := httptest.NewRequest("GET", "/value/gauge/none", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rr.Code)
	}
}

func TestHandler_GetCounter(t *testing.T) {
	h, ms := newTestHandler()
	ms.UpdateCounter("hits", 5)
	ms.UpdateCounter("hits", 3)

	r := chi.NewRouter()
	r.Get("/value/counter/{name}", h.GetCounter)

	req := httptest.NewRequest("GET", "/value/counter/hits", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != 200 {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	if strings.TrimSpace(rr.Body.String()) != "8" {
		t.Fatalf("body = %q, want %q", rr.Body.String(), "8")
	}
}

func TestHandler_GetMetrics_Gauge(t *testing.T) {
	h, ms := newTestHandler()
	ms.UpdateGauge("load", 0.99)

	body, _ := json.Marshal(models.Metrics{
		ID:    "load",
		MType: models.Gauge,
	})

	req := httptest.NewRequest("POST", "/value", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	h.GetMetrics(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}

	var resp models.Metrics
	_ = json.NewDecoder(rr.Body).Decode(&resp)

	if resp.Value == nil || *resp.Value != 0.99 {
		t.Fatalf("Value = %v, want 0.99", resp.Value)
	}
}

func TestHandler_UpdateMetrics_Gauge(t *testing.T) {
	h, ms := newTestHandler()

	body, _ := json.Marshal(models.Metrics{
		ID:    "temp",
		MType: models.Gauge,
		Value: ptrF(123.4),
	})

	req := httptest.NewRequest("POST", "/update", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	h.UpdateMetrics(rr, req)

	if rr.Code != 200 {
		t.Fatalf("status = %d, want 200", rr.Code)
	}

	val, ok := ms.GetGauge("temp")
	if !ok || val != 123.4 {
		t.Fatalf("gauge not updated, val=%v ok=%v", val, ok)
	}
}

func TestHandler_UpdateMetricsBatch(t *testing.T) {
	h, ms := newTestHandler()

	metrics := []models.Metrics{
		{ID: "temp", MType: models.Gauge, Value: ptrF(7.7)},
		{ID: "hits", MType: models.Counter, Delta: ptrI(10)},
	}

	body, _ := json.Marshal(metrics)

	req := httptest.NewRequest("POST", "/updates", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	h.UpdateMetricsBatch(rr, req)

	if rr.Code != 200 {
		t.Fatalf("status = %d, want 200", rr.Code)
	}

	v1, _ := ms.GetGauge("temp")
	if v1 != 7.7 {
		t.Fatalf("temp gauge = %v, want 7.7", v1)
	}

	v2, _ := ms.GetCounter("hits")
	if v2 != 10 {
		t.Fatalf("hits counter = %v, want 10", v2)
	}
}

func TestHandler_UpdateGauge_InvalidValue(t *testing.T) {
	h, _ := newTestHandler()

	r := chi.NewRouter()
	r.Post("/update/gauge/{name}/{value}", h.UpdateGauge)

	req := httptest.NewRequest("POST", "/update/gauge/test/notnumber", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d want 400", rr.Code)
	}
}

func TestHandler_UpdateCounter_OK(t *testing.T) {
	h, ms := newTestHandler()

	r := chi.NewRouter()
	r.Post("/update/counter/{name}/{value}", h.UpdateCounter)

	req := httptest.NewRequest("POST", "/update/counter/calls/5", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != 200 {
		t.Fatalf("status = %d, want 200", rr.Code)
	}

	v, _ := ms.GetCounter("calls")
	if v != 5 {
		t.Fatalf("counter = %d, want 5", v)
	}
}

func TestHandler_WithRequestCompress(t *testing.T) {
	h, _ := newTestHandler()

	var received string

	base := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, _ := io.ReadAll(r.Body)
		received = string(data)
	})

	wrapped := h.WithRequestCompress(base)

	// compress body
	var buf bytes.Buffer
	gzw := gzip.NewWriter(&buf)
	gzw.Write([]byte("hello"))
	gzw.Close()

	req := httptest.NewRequest("POST", "/comp", bytes.NewReader(buf.Bytes()))
	req.Header.Set("Content-Encoding", "gzip")

	rr := httptest.NewRecorder()
	wrapped.ServeHTTP(rr, req)

	if received != "hello" {
		t.Fatalf("received=%q want %q", received, "hello")
	}
}

func TestHandler_WithResponseCompress(t *testing.T) {
	h, _ := newTestHandler()

	base := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	wrapped := h.WithResponseCompress(base)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	if rr.Header().Get("Content-Encoding") != "gzip" {
		t.Fatalf("expected gzip header")
	}

	gzr, _ := gzip.NewReader(bytes.NewReader(rr.Body.Bytes()))
	out, _ := io.ReadAll(gzr)

	if string(out) != "pong" {
		t.Fatalf("decoded=%q want pong", string(out))
	}
}

func TestHandler_Ping_DBNil(t *testing.T) {
	h, _ := newTestHandler() // db=nil

	req := httptest.NewRequest("GET", "/ping", nil)
	rr := httptest.NewRecorder()

	h.Ping(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d want 500", rr.Code)
	}
}
