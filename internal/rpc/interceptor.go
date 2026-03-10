package rpc

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func NewTrustedSubnetInterceptor(trustedSubnet string) (grpc.UnaryServerInterceptor, error) {
	var ipNet *net.IPNet

	if trustedSubnet != "" {
		_, parsedNet, err := net.ParseCIDR(trustedSubnet)
		if err != nil {
			return nil, fmt.Errorf("invalid trusted subnet %q: %w", trustedSubnet, err)
		}
		ipNet = parsedNet
	}

	interceptor := func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		_ = info

		if ipNet == nil {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.PermissionDenied, "metadata is missing")
		}

		values := md.Get("x-real-ip")
		if len(values) == 0 || values[0] == "" {
			return nil, status.Error(codes.PermissionDenied, "x-real-ip metadata is missing")
		}

		ip := net.ParseIP(values[0])
		if ip == nil {
			return nil, status.Error(codes.PermissionDenied, "invalid x-real-ip")
		}

		if !ipNet.Contains(ip) {
			return nil, status.Error(codes.PermissionDenied, "agent ip is not in trusted subnet")
		}

		return handler(ctx, req)
	}

	return interceptor, nil
}
