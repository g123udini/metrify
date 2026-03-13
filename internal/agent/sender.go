package agent

import (
	"crypto/rsa"
	"fmt"
	"go.uber.org/zap"
	models "metrify/internal/model"
	"net"
)

type Sender interface {
	UpdateMetric(metric models.Metrics) error
	UpdateMetrics(metrics []models.Metrics) error
	Close() error
}

func NewSender(protocol, host string, logger *zap.SugaredLogger, hashKey string, publicKey *rsa.PublicKey) (Sender, error) {
	switch protocol {
	case "", "http":
		return NewHTTPClient(host, logger, hashKey, publicKey), nil
	case "grpc":
		return NewGRPCClient(host, logger, hashKey, publicKey), nil
	default:
		return nil, fmt.Errorf("unknown protocol %q", protocol)
	}
}

func getOutboundIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddr, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok {
		return "", fmt.Errorf("local address is not UDP")
	}

	return localAddr.IP.String(), nil
}
