// client_test.go
package agent

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/zap"

	"metrify/internal/service"
)

func TestClient_sendRequest_NoCrypto_SendsPlainBodyAndHash(t *testing.T) {
	const (
		path    = "/update"
		hashKey = "secret"
	)

	wantBody := []byte(`{"hello":"world"}`)

	var gotContentType string
	var gotEnc string
	var gotHash string
	var gotBody []byte

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != path {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		gotContentType = r.Header.Get("Content-Type")
		gotEnc = r.Header.Get("Content-Encryption")
		gotHash = r.Header.Get("HashSHA256")

		b, _ := io.ReadAll(r.Body)
		_ = r.Body.Close()
		gotBody = b

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	t.Cleanup(srv.Close)

	host := strings.TrimPrefix(srv.URL, "http://")

	c := NewClient(host, zap.NewNop().Sugar(), hashKey, nil)

	if err := c.sendRequest(path, wantBody, 1); err != nil {
		t.Fatalf("sendRequest error: %v", err)
	}

	if gotContentType != "application/json" {
		t.Fatalf("Content-Type=%q want %q", gotContentType, "application/json")
	}
	if gotEnc != "" {
		t.Fatalf("Content-Encryption=%q want empty", gotEnc)
	}
	if string(gotBody) != string(wantBody) {
		t.Fatalf("body=%q want %q", string(gotBody), string(wantBody))
	}

	wantHash := service.SignData(wantBody, hashKey)
	if gotHash != wantHash {
		t.Fatalf("HashSHA256=%q want %q", gotHash, wantHash)
	}
}

func TestClient_sendRequest_WithCrypto_EncryptsBodyAndSetsHeaders(t *testing.T) {
	const (
		path    = "/updates"
		hashKey = "secret"
	)

	// RSA keys for test
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	pub := &priv.PublicKey

	plainBody := []byte(`{"k":"v"}`)

	var gotContentType string
	var gotEnc string
	var gotHash string
	var gotDecrypted []byte

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != path {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		gotContentType = r.Header.Get("Content-Type")
		gotEnc = r.Header.Get("Content-Encryption")
		gotHash = r.Header.Get("HashSHA256")

		encB64, _ := io.ReadAll(r.Body)
		_ = r.Body.Close()

		cipher, err := base64.StdEncoding.DecodeString(string(encB64))
		if err != nil {
			t.Fatalf("base64 decode: %v", err)
		}

		dec, err := rsa.DecryptPKCS1v15(rand.Reader, priv, cipher)
		if err != nil {
			t.Fatalf("rsa decrypt: %v", err)
		}
		gotDecrypted = dec

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	t.Cleanup(srv.Close)

	host := strings.TrimPrefix(srv.URL, "http://")

	c := NewClient(host, zap.NewNop().Sugar(), hashKey, pub)

	if err := c.sendRequest(path, plainBody, 1); err != nil {
		t.Fatalf("sendRequest error: %v", err)
	}

	if gotContentType != "application/octet-stream" {
		t.Fatalf("Content-Type=%q want %q", gotContentType, "application/octet-stream")
	}
	if gotEnc != "RSA-PKCS1v15" {
		t.Fatalf("Content-Encryption=%q want %q", gotEnc, "RSA-PKCS1v15")
	}

	// hash должен считаться по исходному (нешифрованному) body
	wantHash := service.SignData(plainBody, hashKey)
	if gotHash != wantHash {
		t.Fatalf("HashSHA256=%q want %q", gotHash, wantHash)
	}

	if string(gotDecrypted) != string(plainBody) {
		t.Fatalf("decrypted=%q want %q", string(gotDecrypted), string(plainBody))
	}
}

func TestClient_sendRequest_Non200_ReturnsError(t *testing.T) {
	const path = "/update"
	wantBody := []byte("x")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("fail"))
	}))
	t.Cleanup(srv.Close)

	host := strings.TrimPrefix(srv.URL, "http://")
	c := NewClient(host, zap.NewNop().Sugar(), "", nil)

	err := c.sendRequest(path, wantBody, 1)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "unexpected status 500") {
		t.Fatalf("error=%q, want contains %q", err.Error(), "unexpected status 500")
	}
	if !strings.Contains(err.Error(), "fail") {
		t.Fatalf("error=%q, want contains %q", err.Error(), "fail")
	}
}
