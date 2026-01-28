package audit

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewFileReceiver_EmptyPath_ReturnsNil(t *testing.T) {
	if r := NewFileReceiver(""); r != nil {
		t.Fatalf("expected nil receiver for empty path")
	}
}

func TestFileReceiver_Receive_AppendsJSONLine(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.log")

	r := NewFileReceiver(path)
	if r == nil {
		t.Fatalf("expected non-nil receiver")
	}

	e1 := Event{TS: 111, Metrics: []string{"Alloc"}, IPAddress: "1.2.3.4"}
	e2 := Event{TS: 222, Metrics: []string{"Frees", "HeapAlloc"}, IPAddress: "5.6.7.8"}

	if err := r.Receive(context.Background(), e1); err != nil {
		t.Fatalf("Receive(e1) error: %v", err)
	}
	if err := r.Receive(context.Background(), e2); err != nil {
		t.Fatalf("Receive(e2) error: %v", err)
	}

	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open file: %v", err)
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	var lines []string
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	if err := sc.Err(); err != nil {
		t.Fatalf("scan: %v", err)
	}

	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %#v", len(lines), lines)
	}

	var got1, got2 Event
	if err := json.Unmarshal([]byte(lines[0]), &got1); err != nil {
		t.Fatalf("unmarshal line1: %v, line=%q", err, lines[0])
	}
	if err := json.Unmarshal([]byte(lines[1]), &got2); err != nil {
		t.Fatalf("unmarshal line2: %v, line=%q", err, lines[1])
	}

	if got1.TS != e1.TS || got1.IPAddress != e1.IPAddress || strings.Join(got1.Metrics, ",") != strings.Join(e1.Metrics, ",") {
		t.Fatalf("unexpected event1: got=%+v want=%+v", got1, e1)
	}
	if got2.TS != e2.TS || got2.IPAddress != e2.IPAddress || strings.Join(got2.Metrics, ",") != strings.Join(e2.Metrics, ",") {
		t.Fatalf("unexpected event2: got=%+v want=%+v", got2, e2)
	}
}

func TestNewHTTPReceiver_EmptyURL_ReturnsNil(t *testing.T) {
	if r := NewHTTPReceiver("", nil); r != nil {
		t.Fatalf("expected nil receiver for empty url")
	}
}

func TestHTTPReceiver_Receive_PostsJSONAndHeaders(t *testing.T) {
	var (
		gotMethod      string
		gotContentType string
		gotBody        []byte
		gotCalled      bool
	)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotCalled = true
		gotMethod = r.Method
		gotContentType = r.Header.Get("Content-Type")

		b, _ := io.ReadAll(r.Body)
		_ = r.Body.Close()
		gotBody = b

		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	recv := NewHTTPReceiver(srv.URL, srv.Client())
	if recv == nil {
		t.Fatalf("expected non-nil receiver")
	}

	e := Event{TS: time.Now().Unix(), Metrics: []string{"Alloc", "Frees"}, IPAddress: "192.168.0.42"}
	if err := recv.Receive(context.Background(), e); err != nil {
		t.Fatalf("Receive() error: %v", err)
	}

	if !gotCalled {
		t.Fatalf("expected server to be called")
	}
	if gotMethod != http.MethodPost {
		t.Fatalf("expected method POST, got %q", gotMethod)
	}
	if gotContentType != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", gotContentType)
	}

	var got Event
	if err := json.Unmarshal(gotBody, &got); err != nil {
		t.Fatalf("invalid json body: %v; body=%q", err, string(gotBody))
	}
	if got.TS != e.TS || got.IPAddress != e.IPAddress || strings.Join(got.Metrics, ",") != strings.Join(e.Metrics, ",") {
		t.Fatalf("unexpected event body: got=%+v want=%+v", got, e)
	}
}

func TestHTTPReceiver_Receive_Non2xx_ReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot) // 418
	}))
	defer srv.Close()

	recv := NewHTTPReceiver(srv.URL, srv.Client())
	err := recv.Receive(context.Background(), Event{TS: 1})

	if err == nil {
		t.Fatalf("expected error on non-2xx status")
	}
	if !strings.Contains(err.Error(), "non-2xx") {
		t.Fatalf("expected error to mention non-2xx, got: %v", err)
	}
}
