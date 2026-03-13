package agent

import (
	"context"
	"crypto/rsa"
	"fmt"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	models "metrify/internal/model"
	"metrify/internal/proto"
)

// generate:reset
type GRPCClient struct {
	logger    *zap.SugaredLogger
	host      string
	hashKey   string
	maxRetry  int
	publicKey *rsa.PublicKey

	conn   *grpc.ClientConn
	client proto.MetricsClient
}

func NewGRPCClient(host string, logger *zap.SugaredLogger, hashKey string, publicKey *rsa.PublicKey) *GRPCClient {
	conn, err := grpc.NewClient(
		host,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		logger.Fatalw("failed to dial grpc server", "host", host, "error", err)
	}

	return &GRPCClient{
		logger:    logger,
		host:      host,
		hashKey:   hashKey,
		maxRetry:  3,
		publicKey: publicKey,
		conn:      conn,
		client:    proto.NewMetricsClient(conn),
	}
}

func (client *GRPCClient) Close() error {
	if client.conn == nil {
		return nil
	}

	return client.conn.Close()
}

func (client *GRPCClient) UpdateMetric(metric models.Metrics) error {
	return client.UpdateMetrics([]models.Metrics{metric})
}

func (client *GRPCClient) UpdateMetrics(metrics []models.Metrics) error {
	protoMetrics := make([]*proto.Metric, 0, len(metrics))

	for _, metric := range metrics {
		protoMetric, err := transformMetricToProto(metric)
		if err != nil {
			return err
		}

		protoMetrics = append(protoMetrics, protoMetric)
	}

	req := &proto.UpdateMetricsRequest{}
	req.SetMetrics(protoMetrics)

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	ctx = client.withRealIP(ctx)

	_, err := client.client.UpdateMetrics(ctx, req)
	if err != nil {
		return fmt.Errorf("grpc update metrics failed: %w", err)
	}

	return nil
}

func (client *GRPCClient) withRealIP(ctx context.Context) context.Context {
	ip, err := getOutboundIP()
	if err != nil {
		client.logger.Warnw("failed to detect outbound ip for grpc metadata", "error", err)
		return ctx
	}

	if ip == "" {
		return ctx
	}

	return metadata.AppendToOutgoingContext(ctx, "x-real-ip", ip)
}

func transformMetricToProto(metric models.Metrics) (*proto.Metric, error) {
	m := &proto.Metric{}
	m.SetId(metric.ID)

	switch metric.MType {
	case models.Gauge:
		if metric.Value == nil {
			return nil, fmt.Errorf("gauge metric %q has nil value", metric.ID)
		}

		m.SetType(proto.Metric_GAUGE)
		m.SetValue(*metric.Value)

		return m, nil

	case models.Counter:
		if metric.Delta == nil {
			return nil, fmt.Errorf("counter metric %q has nil delta", metric.ID)
		}

		m.SetType(proto.Metric_COUNTER)
		m.SetDelta(*metric.Delta)

		return m, nil

	default:
		return nil, fmt.Errorf("unknown metric type %q", metric.MType)
	}
}
