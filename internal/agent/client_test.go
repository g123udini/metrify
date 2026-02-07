package agent

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/zap"
	models "metrify/internal/model"
	"metrify/internal/service"
)

func TestClient_UpdateMetric_SendsRequestWithHeaders(t *testing.T) {
	var gotMethod, gotPath, gotCT, gotHash string
	var gotBody []byte

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotCT = r.Header.Get("Content-Type")
		gotHash = r.Header.Get("HashSHA256")
		b, _ := io.ReadAll(r.Body)
		gotBody = b

		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	host := strings.TrimPrefix(srv.URL, "http://")
	logger := zap.NewNop().Sugar()

	c := NewClient(host, logger, "secret")
	c.maxRetry = 1

	var metric models.Metrics
	wantBody, err := json.Marshal(metric)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	if err := c.UpdateMetric(metric); err != nil {
		t.Fatalf("UpdateMetric err: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Fatalf("method=%q want=%q", gotMethod, http.MethodPost)
	}
	if gotPath != "/update" {
		t.Fatalf("path=%q want=%q", gotPath, "/update")
	}
	if !strings.Contains(gotCT, "application/json") {
		t.Fatalf("Content-Type=%q want contains application/json", gotCT)
	}
	if string(gotBody) != string(wantBody) {
		t.Fatalf("body=%s want=%s", string(gotBody), string(wantBody))
	}

	wantHash := service.SignData(wantBody, "secret")
	if gotHash != wantHash {
		t.Fatalf("HashSHA256=%q want=%q", gotHash, wantHash)
	}
}

func TestClient_UpdateMetric_WithoutHashHeaderWhenNoKey(t *testing.T) {
	var gotHash string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHash = r.Header.Get("HashSHA256")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	host := strings.TrimPrefix(srv.URL, "http://")
	logger := zap.NewNop().Sugar()

	c := NewClient(host, logger, "")
	c.maxRetry = 1

	var metric models.Metrics
	if err := c.UpdateMetric(metric); err != nil {
		t.Fatalf("UpdateMetric err: %v", err)
	}

	if gotHash != "" {
		t.Fatalf("HashSHA256=%q want empty", gotHash)
	}
}

func TestClient_UpdateMetrics_SendsToUpdates(t *testing.T) {
	var gotPath string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	host := strings.TrimPrefix(srv.URL, "http://")
	logger := zap.NewNop().Sugar()

	c := NewClient(host, logger, "k")
	c.maxRetry = 1

	var m models.Metrics
	if err := c.UpdateMetrics([]models.Metrics{m, m}); err != nil {
		t.Fatalf("UpdateMetrics err: %v", err)
	}

	if gotPath != "/updates" {
		t.Fatalf("path=%q want=%q", gotPath, "/updates")
	}
}

func TestClient_SendRequest_Non200ReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("bad"))
	}))
	defer srv.Close()

	host := strings.TrimPrefix(srv.URL, "http://")
	logger := zap.NewNop().Sugar()

	c := NewClient(host, logger, "")
	c.maxRetry = 1

	err := c.sendRequest("/update", []byte(`{"x":1}`), 1)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "unexpected status 400") {
		t.Fatalf("err=%q want contains %q", err.Error(), "unexpected status 400")
	}
	if !strings.Contains(err.Error(), "bad") {
		t.Fatalf("err=%q want contains %q", err.Error(), "bad")
	}
}
