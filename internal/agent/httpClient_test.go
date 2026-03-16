package agent

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
	models "metrify/internal/model"
)

func newTestHTTPClient(host string) *HTTPClient {
	client := NewHTTPClient(host, zap.NewNop().Sugar(), "", nil)
	client.maxRetry = 1
	client.resty.SetTimeout(2 * time.Second)

	return client
}

func TestNewHTTPClient(t *testing.T) {
	client := NewHTTPClient("localhost:8080", zap.NewNop().Sugar(), "key", nil)

	if client == nil {
		t.Fatal("expected client, got nil")
	}
	if client.host != "localhost:8080" {
		t.Fatalf("unexpected host: %q", client.host)
	}
	if client.hashKey != "key" {
		t.Fatalf("unexpected hash key: %q", client.hashKey)
	}
	if client.maxRetry != 3 {
		t.Fatalf("unexpected maxRetry: %d", client.maxRetry)
	}
	if client.resty == nil {
		t.Fatal("expected resty client to be initialized")
	}
}

func TestHTTPClient_Close(t *testing.T) {
	client := &HTTPClient{}

	if err := client.Close(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHTTPClient_UpdateMetric(t *testing.T) {
	var gotPath string
	var gotContentType string
	var gotBody []byte

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotContentType = r.Header.Get("Content-Type")

		var err error
		gotBody, err = io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read body: %v", err)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	host := strings.TrimPrefix(ts.URL, "http://")
	client := newTestHTTPClient(host)

	v := 12.34
	err := client.UpdateMetric(models.Metrics{
		ID:    "Alloc",
		MType: models.Gauge,
		Value: &v,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/update" {
		t.Fatalf("got path %q, want /update", gotPath)
	}
	if gotContentType != "application/json" {
		t.Fatalf("got content-type %q, want application/json", gotContentType)
	}

	var metric models.Metrics
	if err := json.Unmarshal(gotBody, &metric); err != nil {
		t.Fatalf("failed to unmarshal body: %v", err)
	}

	if metric.ID != "Alloc" {
		t.Fatalf("got metric id %q, want Alloc", metric.ID)
	}
	if metric.MType != models.Gauge {
		t.Fatalf("got metric type %q, want %q", metric.MType, models.Gauge)
	}
	if metric.Value == nil || *metric.Value != 12.34 {
		t.Fatalf("unexpected metric value: %+v", metric.Value)
	}
}

func TestHTTPClient_UpdateMetrics(t *testing.T) {
	var gotPath string
	var gotContentType string
	var gotBody []byte

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotContentType = r.Header.Get("Content-Type")

		var err error
		gotBody, err = io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read body: %v", err)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	host := strings.TrimPrefix(ts.URL, "http://")
	client := newTestHTTPClient(host)

	v := 1.5
	d := int64(3)

	err := client.UpdateMetrics([]models.Metrics{
		{
			ID:    "Alloc",
			MType: models.Gauge,
			Value: &v,
		},
		{
			ID:    "PollCount",
			MType: models.Counter,
			Delta: &d,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/updates" {
		t.Fatalf("got path %q, want /updates", gotPath)
	}
	if gotContentType != "application/json" {
		t.Fatalf("got content-type %q, want application/json", gotContentType)
	}

	var metrics []models.Metrics
	if err := json.Unmarshal(gotBody, &metrics); err != nil {
		t.Fatalf("failed to unmarshal body: %v", err)
	}

	if len(metrics) != 2 {
		t.Fatalf("got %d metrics, want 2", len(metrics))
	}
}

func TestHTTPClient_SendRequest_Non200(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer ts.Close()

	host := strings.TrimPrefix(ts.URL, "http://")
	client := newTestHTTPClient(host)

	err := client.sendRequest("/update", []byte(`{}`), 1)
	if err == nil {
		t.Fatal("expected error")
	}

	if !strings.Contains(err.Error(), "unexpected status 500") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHTTPClient_EncryptBody_NoPublicKey(t *testing.T) {
	client := &HTTPClient{}

	plain := []byte("hello")
	got, err := client.encryptBody(plain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !bytes.Equal(got, plain) {
		t.Fatalf("expected same bytes, got %q", string(got))
	}
}

func TestHTTPClient_EncryptBody_WithPublicKey(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatalf("failed to generate rsa key: %v", err)
	}

	client := &HTTPClient{
		publicKey: &privateKey.PublicKey,
	}

	plain := []byte("hello")
	got, err := client.encryptBody(plain)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if bytes.Equal(got, plain) {
		t.Fatal("encrypted data must differ from plain data")
	}

	decoded := make([]byte, base64.StdEncoding.DecodedLen(len(got)))
	n, err := base64.StdEncoding.Decode(decoded, got)
	if err != nil {
		t.Fatalf("encrypted body is not valid base64: %v", err)
	}
	if n == 0 {
		t.Fatal("decoded encrypted body is empty")
	}
}
