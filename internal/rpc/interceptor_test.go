package rpc

import (
	"context"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestNewTrustedSubnetInterceptor_InvalidCIDR(t *testing.T) {
	_, err := NewTrustedSubnetInterceptor("not-a-cidr")
	if err == nil {
		t.Fatal("expected error for invalid cidr")
	}
}

func TestTrustedSubnetInterceptor_EmptySubnet_AllowsRequest(t *testing.T) {
	interceptor, err := NewTrustedSubnetInterceptor("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	called := false
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		called = true
		return "ok", nil
	}

	resp, err := interceptor(
		context.Background(),
		"req",
		&grpc.UnaryServerInfo{FullMethod: "/metrics.Metrics/UpdateMetrics"},
		handler,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected handler to be called")
	}
	if resp != "ok" {
		t.Fatalf("unexpected response: %v", resp)
	}
}

func TestTrustedSubnetInterceptor_MetadataMissing(t *testing.T) {
	interceptor, err := NewTrustedSubnetInterceptor("192.168.1.0/24")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		t.Fatal("handler must not be called")
		return nil, nil
	}

	_, err = interceptor(
		context.Background(),
		"req",
		&grpc.UnaryServerInfo{FullMethod: "/metrics.Metrics/UpdateMetrics"},
		handler,
	)
	if err == nil {
		t.Fatal("expected permission denied error")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected grpc status error, got: %v", err)
	}
	if st.Code() != codes.PermissionDenied {
		t.Fatalf("expected PermissionDenied, got %v", st.Code())
	}
	if st.Message() != "metadata is missing" {
		t.Fatalf("unexpected message: %q", st.Message())
	}
}

func TestTrustedSubnetInterceptor_RealIPMissing(t *testing.T) {
	interceptor, err := NewTrustedSubnetInterceptor("192.168.1.0/24")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx := metadata.NewIncomingContext(context.Background(), metadata.MD{})

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		t.Fatal("handler must not be called")
		return nil, nil
	}

	_, err = interceptor(
		ctx,
		"req",
		&grpc.UnaryServerInfo{FullMethod: "/metrics.Metrics/UpdateMetrics"},
		handler,
	)
	if err == nil {
		t.Fatal("expected permission denied error")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected grpc status error, got: %v", err)
	}
	if st.Code() != codes.PermissionDenied {
		t.Fatalf("expected PermissionDenied, got %v", st.Code())
	}
	if st.Message() != "x-real-ip metadata is missing" {
		t.Fatalf("unexpected message: %q", st.Message())
	}
}

func TestTrustedSubnetInterceptor_InvalidIP(t *testing.T) {
	interceptor, err := NewTrustedSubnetInterceptor("192.168.1.0/24")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx := metadata.NewIncomingContext(
		context.Background(),
		metadata.Pairs("x-real-ip", "not-an-ip"),
	)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		t.Fatal("handler must not be called")
		return nil, nil
	}

	_, err = interceptor(
		ctx,
		"req",
		&grpc.UnaryServerInfo{FullMethod: "/metrics.Metrics/UpdateMetrics"},
		handler,
	)
	if err == nil {
		t.Fatal("expected permission denied error")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected grpc status error, got: %v", err)
	}
	if st.Code() != codes.PermissionDenied {
		t.Fatalf("expected PermissionDenied, got %v", st.Code())
	}
	if st.Message() != "invalid x-real-ip" {
		t.Fatalf("unexpected message: %q", st.Message())
	}
}

func TestTrustedSubnetInterceptor_IPNotInSubnet(t *testing.T) {
	interceptor, err := NewTrustedSubnetInterceptor("192.168.1.0/24")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx := metadata.NewIncomingContext(
		context.Background(),
		metadata.Pairs("x-real-ip", "10.0.0.1"),
	)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		t.Fatal("handler must not be called")
		return nil, nil
	}

	_, err = interceptor(
		ctx,
		"req",
		&grpc.UnaryServerInfo{FullMethod: "/metrics.Metrics/UpdateMetrics"},
		handler,
	)
	if err == nil {
		t.Fatal("expected permission denied error")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected grpc status error, got: %v", err)
	}
	if st.Code() != codes.PermissionDenied {
		t.Fatalf("expected PermissionDenied, got %v", st.Code())
	}
	if st.Message() != "agent ip is not in trusted subnet" {
		t.Fatalf("unexpected message: %q", st.Message())
	}
}

func TestTrustedSubnetInterceptor_IPInSubnet_AllowsRequest(t *testing.T) {
	interceptor, err := NewTrustedSubnetInterceptor("192.168.1.0/24")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx := metadata.NewIncomingContext(
		context.Background(),
		metadata.Pairs("x-real-ip", "192.168.1.42"),
	)

	called := false
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		called = true
		return "ok", nil
	}

	resp, err := interceptor(
		ctx,
		"req",
		&grpc.UnaryServerInfo{FullMethod: "/metrics.Metrics/UpdateMetrics"},
		handler,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected handler to be called")
	}
	if resp != "ok" {
		t.Fatalf("unexpected response: %v", resp)
	}
}
