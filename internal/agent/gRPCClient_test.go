package agent

import (
	"context"
	"errors"
	"net"
	"testing"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	models "metrify/internal/model"
	"metrify/internal/proto"
)

type metricsClientMock struct {
	lastReq *proto.UpdateMetricsRequest
	lastMD  metadata.MD
	err     error
}

func (m *metricsClientMock) UpdateMetrics(
	ctx context.Context,
	in *proto.UpdateMetricsRequest,
	_ ...grpc.CallOption,
) (*proto.UpdateMetricsResponse, error) {
	m.lastReq = in

	md, ok := metadata.FromOutgoingContext(ctx)
	if ok {
		m.lastMD = md
	}

	if m.err != nil {
		return nil, m.err
	}

	return &proto.UpdateMetricsResponse{}, nil
}

func newTestGRPCClient(mock proto.MetricsClient) *GRPCClient {
	return &GRPCClient{
		logger: zap.NewNop().Sugar(),
		client: mock,
	}
}

func TestGRPCClient_Close_NilConn(t *testing.T) {
	client := &GRPCClient{}

	if err := client.Close(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTransformMetricToProto_Gauge(t *testing.T) {
	v := 12.5

	got, err := transformMetricToProto(models.Metrics{
		ID:    "Alloc",
		MType: models.Gauge,
		Value: &v,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.Id != "Alloc" {
		t.Fatalf("got id %q, want Alloc", got.Id)
	}
	if got.Type != proto.Metric_GAUGE {
		t.Fatalf("got type %v, want %v", got.Type, proto.Metric_GAUGE)
	}
	if got.Value != 12.5 {
		t.Fatalf("got value %v, want 12.5", got.Value)
	}
}

func TestTransformMetricToProto_Counter(t *testing.T) {
	d := int64(7)

	got, err := transformMetricToProto(models.Metrics{
		ID:    "PollCount",
		MType: models.Counter,
		Delta: &d,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.Id != "PollCount" {
		t.Fatalf("got id %q, want PollCount", got.Id)
	}
	if got.Type != proto.Metric_COUNTER {
		t.Fatalf("got type %v, want %v", got.Type, proto.Metric_COUNTER)
	}
	if got.Delta != 7 {
		t.Fatalf("got delta %v, want 7", got.Delta)
	}
}

func TestTransformMetricToProto_GaugeNilValue(t *testing.T) {
	_, err := transformMetricToProto(models.Metrics{
		ID:    "Alloc",
		MType: models.Gauge,
	})
	if err == nil {
		t.Fatal("expected error")
	}

	want := `gauge metric "Alloc" has nil value`
	if err.Error() != want {
		t.Fatalf("got error %q, want %q", err.Error(), want)
	}
}

func TestTransformMetricToProto_CounterNilDelta(t *testing.T) {
	_, err := transformMetricToProto(models.Metrics{
		ID:    "PollCount",
		MType: models.Counter,
	})
	if err == nil {
		t.Fatal("expected error")
	}

	want := `counter metric "PollCount" has nil delta`
	if err.Error() != want {
		t.Fatalf("got error %q, want %q", err.Error(), want)
	}
}

func TestTransformMetricToProto_UnknownType(t *testing.T) {
	_, err := transformMetricToProto(models.Metrics{
		ID:    "Broken",
		MType: "unknown",
	})
	if err == nil {
		t.Fatal("expected error")
	}

	want := `unknown metric type "unknown"`
	if err.Error() != want {
		t.Fatalf("got error %q, want %q", err.Error(), want)
	}
}

func TestGRPCClient_UpdateMetric(t *testing.T) {
	mock := &metricsClientMock{}
	client := newTestGRPCClient(mock)

	v := 42.25
	err := client.UpdateMetric(models.Metrics{
		ID:    "Alloc",
		MType: models.Gauge,
		Value: &v,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mock.lastReq == nil {
		t.Fatal("expected request to be sent")
	}
	if len(mock.lastReq.Metrics) != 1 {
		t.Fatalf("got %d metrics, want 1", len(mock.lastReq.Metrics))
	}

	got := mock.lastReq.Metrics[0]
	if got.Id != "Alloc" {
		t.Fatalf("got id %q, want Alloc", got.Id)
	}
	if got.Type != proto.Metric_GAUGE {
		t.Fatalf("got type %v, want %v", got.Type, proto.Metric_GAUGE)
	}
	if got.Value != 42.25 {
		t.Fatalf("got value %v, want 42.25", got.Value)
	}
}

func TestGRPCClient_UpdateMetrics(t *testing.T) {
	mock := &metricsClientMock{}
	client := newTestGRPCClient(mock)

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

	if mock.lastReq == nil {
		t.Fatal("expected request to be sent")
	}
	if len(mock.lastReq.Metrics) != 2 {
		t.Fatalf("got %d metrics, want 2", len(mock.lastReq.Metrics))
	}

	if mock.lastReq.Metrics[0].Id != "Alloc" {
		t.Fatalf("unexpected first metric id: %q", mock.lastReq.Metrics[0].Id)
	}
	if mock.lastReq.Metrics[1].Id != "PollCount" {
		t.Fatalf("unexpected second metric id: %q", mock.lastReq.Metrics[1].Id)
	}
}

func TestGRPCClient_UpdateMetrics_TransformError(t *testing.T) {
	mock := &metricsClientMock{}
	client := newTestGRPCClient(mock)

	err := client.UpdateMetrics([]models.Metrics{
		{
			ID:    "Alloc",
			MType: models.Gauge,
			Value: nil,
		},
	})
	if err == nil {
		t.Fatal("expected error")
	}

	want := `gauge metric "Alloc" has nil value`
	if err.Error() != want {
		t.Fatalf("got error %q, want %q", err.Error(), want)
	}
	if mock.lastReq != nil {
		t.Fatal("request must not be sent on transform error")
	}
}

func TestGRPCClient_UpdateMetrics_ClientError(t *testing.T) {
	mock := &metricsClientMock{
		err: errors.New("grpc boom"),
	}
	client := newTestGRPCClient(mock)

	v := 10.0
	err := client.UpdateMetrics([]models.Metrics{
		{
			ID:    "Alloc",
			MType: models.Gauge,
			Value: &v,
		},
	})
	if err == nil {
		t.Fatal("expected error")
	}

	want := "grpc update metrics failed: grpc boom"
	if err.Error() != want {
		t.Fatalf("got error %q, want %q", err.Error(), want)
	}
}

func TestGRPCClient_WithRealIP_ContextValid(t *testing.T) {
	mock := &metricsClientMock{}
	client := newTestGRPCClient(mock)

	ctx := context.Background()
	gotCtx := client.withRealIP(ctx)
	if gotCtx == nil {
		t.Fatal("expected non-nil context")
	}
}

func TestGRPCClient_UpdateMetrics_AddsRealIPMetadata_WhenAvailable(t *testing.T) {
	mock := &metricsClientMock{}
	client := newTestGRPCClient(mock)

	v := 11.0
	err := client.UpdateMetrics([]models.Metrics{
		{
			ID:    "Alloc",
			MType: models.Gauge,
			Value: &v,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mock.lastMD == nil {
		t.Fatal("expected outgoing metadata")
	}

	values := mock.lastMD.Get("x-real-ip")
	if len(values) > 0 {
		if net.ParseIP(values[0]) == nil {
			t.Fatalf("invalid x-real-ip metadata value: %q", values[0])
		}
	}
}
