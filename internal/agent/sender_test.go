package agent

import (
	"go.uber.org/zap"
	"testing"
)

func TestNewSender_HTTP_Default(t *testing.T) {
	logger := zap.NewNop().Sugar()

	s, err := NewSender("", "localhost:8080", logger, "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if s == nil {
		t.Fatal("expected sender")
	}

	if _, ok := s.(*HTTPClient); !ok {
		t.Fatalf("expected *HTTPClient, got %T", s)
	}
}

func TestNewSender_HTTP(t *testing.T) {
	logger := zap.NewNop().Sugar()

	s, err := NewSender("http", "localhost:8080", logger, "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := s.(*HTTPClient); !ok {
		t.Fatalf("expected *HTTPClient, got %T", s)
	}
}

func TestNewSender_GRPC(t *testing.T) {
	logger := zap.NewNop().Sugar()

	s, err := NewSender("grpc", "localhost:9090", logger, "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := s.(*GRPCClient); !ok {
		t.Fatalf("expected *GRPCClient, got %T", s)
	}
}

func TestNewSender_UnknownProtocol(t *testing.T) {
	logger := zap.NewNop().Sugar()

	_, err := NewSender("ws", "localhost:8080", logger, "", nil)
	if err == nil {
		t.Fatal("expected error")
	}

	want := `unknown protocol "ws"`
	if err.Error() != want {
		t.Fatalf("got %q want %q", err.Error(), want)
	}
}

func TestGetOutboundIP(t *testing.T) {
	ip, err := getOutboundIP()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ip == "" {
		t.Fatal("expected non empty ip")
	}
}
