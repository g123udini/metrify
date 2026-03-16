package rpc

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	models "metrify/internal/model"
	"metrify/internal/proto"
	"metrify/internal/service"
	"net"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func RunGRPCServer(addr string, storage service.Storage, logger *zap.SugaredLogger) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer()
	proto.RegisterMetricsServer(grpcServer, NewMetricsService(storage))

	logger.Infow("grpc server started", "addr", addr)

	return grpcServer.Serve(lis)
}

type MetricsService struct {
	proto.UnimplementedMetricsServer
	storage service.Storage
}

func NewMetricsService(storage service.Storage) *MetricsService {
	return &MetricsService{
		storage: storage,
	}
}

func (s *MetricsService) UpdateMetrics(
	ctx context.Context,
	req *proto.UpdateMetricsRequest,
) (*proto.UpdateMetricsResponse, error) {
	_ = ctx

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}

	for _, grpcMetric := range req.GetMetrics() {
		metric, err := metricFromProto(grpcMetric)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		switch metric.MType {
		case models.Gauge:
			if err := s.storage.UpdateGauge(metric.ID, *metric.Value); err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
		case models.Counter:
			if err := s.storage.UpdateCounter(metric.ID, *metric.Delta); err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
		default:
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("unknown metric type %q", metric.MType))
		}
	}

	return &proto.UpdateMetricsResponse{}, nil
}

func metricFromProto(m *proto.Metric) (*models.Metrics, error) {
	if m == nil {
		return nil, fmt.Errorf("metric is nil")
	}

	if m.GetId() == "" {
		return nil, fmt.Errorf("metric id is empty")
	}

	switch m.GetType() {
	case proto.Metric_GAUGE:
		value := m.GetValue()

		return &models.Metrics{
			ID:    m.GetId(),
			MType: models.Gauge,
			Value: &value,
		}, nil

	case proto.Metric_COUNTER:
		delta := m.GetDelta()

		return &models.Metrics{
			ID:    m.GetId(),
			MType: models.Counter,
			Delta: &delta,
		}, nil

	default:
		return nil, fmt.Errorf("unknown metric type %v", m.GetType())
	}
}
