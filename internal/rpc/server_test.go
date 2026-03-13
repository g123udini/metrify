package rpc

import (
	"context"
	"errors"
	"testing"

	models "metrify/internal/model"
	"metrify/internal/proto"
)

type storageMock struct {
	gauges           map[string]float64
	counters         map[string]int64
	updateGaugeErr   error
	updateCounterErr error
}

func newStorageMock() *storageMock {
	return &storageMock{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

func (m *storageMock) GetCounter(key string) (int64, bool) {
	v, ok := m.counters[key]
	return v, ok
}

func (m *storageMock) GetGauge(key string) (float64, bool) {
	v, ok := m.gauges[key]
	return v, ok
}

func (m *storageMock) UpdateGauge(name string, value float64) error {
	if m.updateGaugeErr != nil {
		return m.updateGaugeErr
	}
	m.gauges[name] = value
	return nil
}

func (m *storageMock) UpdateCounter(name string, delta int64) error {
	if m.updateCounterErr != nil {
		return m.updateCounterErr
	}
	m.counters[name] += delta
	return nil
}

func (m *storageMock) FlushToFile() error {
	return nil
}

func TestNewMetricsService(t *testing.T) {
	st := newStorageMock()

	svc := NewMetricsService(st)
	if svc == nil {
		t.Fatal("expected service, got nil")
	}
	if svc.storage != st {
		t.Fatal("expected storage to be set")
	}
}

func TestMetricFromProto(t *testing.T) {
	t.Run("nil metric", func(t *testing.T) {
		_, err := metricFromProto(nil)
		if err == nil {
			t.Fatal("expected error")
		}
		if err.Error() != "metric is nil" {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("empty id", func(t *testing.T) {
		metric := &proto.Metric{}
		metric.SetType(proto.Metric_GAUGE)
		_, err := metricFromProto(metric)
		if err == nil {
			t.Fatal("expected error")
		}
		if err.Error() != "metric id is empty" {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("gauge metric", func(t *testing.T) {
		metric := &proto.Metric{}
		metric.SetId("Alloc")
		metric.SetType(proto.Metric_GAUGE)
		metric.SetValue(12.5)

		got, err := metricFromProto(metric)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ID != "Alloc" {
			t.Fatalf("unexpected id: %s", got.ID)
		}
		if got.MType != models.Gauge {
			t.Fatalf("unexpected type: %s", got.MType)
		}
		if got.Value == nil || *got.Value != 12.5 {
			t.Fatalf("unexpected value: %+v", got.Value)
		}
	})

	t.Run("counter metric", func(t *testing.T) {
		metric := &proto.Metric{}
		metric.SetId("PollCount")
		metric.SetType(proto.Metric_COUNTER)
		metric.SetDelta(7)
		got, err := metricFromProto(metric)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ID != "PollCount" {
			t.Fatalf("unexpected id: %s", got.ID)
		}
		if got.MType != models.Counter {
			t.Fatalf("unexpected type: %s", got.MType)
		}
		if got.Delta == nil || *got.Delta != 7 {
			t.Fatalf("unexpected delta: %+v", got.Delta)
		}
	})

	t.Run("unknown metric type", func(t *testing.T) {
		metric := &proto.Metric{}
		metric.SetId("Broken")
		metric.SetType(proto.Metric_MType(99))

		_, err := metricFromProto(metric)
		if err == nil {
			t.Fatal("expected error")
		}
		if err.Error() != "unknown metric type 99" {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestMetricsService_UpdateMetrics(t *testing.T) {
	t.Run("nil request", func(t *testing.T) {
		st := newStorageMock()
		svc := NewMetricsService(st)

		_, err := svc.UpdateMetrics(context.Background(), nil)
		if err == nil {
			t.Fatal("expected error")
		}

		if got := err.Error(); got != "rpc error: code = InvalidArgument desc = request is nil" {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("invalid metric in request", func(t *testing.T) {
		st := newStorageMock()
		svc := NewMetricsService(st)
		req := &proto.UpdateMetricsRequest{}
		req.SetMetrics([]*proto.Metric{nil})

		_, err := svc.UpdateMetrics(context.Background(), req)
		if err == nil {
			t.Fatal("expected error")
		}

		if got := err.Error(); got != "rpc error: code = InvalidArgument desc = metric is nil" {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("update gauge and counter", func(t *testing.T) {
		st := newStorageMock()
		svc := NewMetricsService(st)

		gaugeMetric := &proto.Metric{}
		gaugeMetric.SetId("Alloc")
		gaugeMetric.SetType(proto.Metric_GAUGE)
		gaugeMetric.SetValue(42.5)

		counterMetric := &proto.Metric{}
		counterMetric.SetId("PollCount")
		counterMetric.SetType(proto.Metric_COUNTER)
		counterMetric.SetDelta(3)

		req := &proto.UpdateMetricsRequest{}
		req.SetMetrics([]*proto.Metric{gaugeMetric, counterMetric})

		resp, err := svc.UpdateMetrics(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp == nil {
			t.Fatal("expected response, got nil")
		}

		gauge, ok := st.GetGauge("Alloc")
		if !ok {
			t.Fatal("expected gauge to be saved")
		}
		if gauge != 42.5 {
			t.Fatalf("unexpected gauge value: %v", gauge)
		}

		counter, ok := st.GetCounter("PollCount")
		if !ok {
			t.Fatal("expected counter to be saved")
		}
		if counter != 3 {
			t.Fatalf("unexpected counter value: %v", counter)
		}
	})

	t.Run("storage update gauge error", func(t *testing.T) {
		st := newStorageMock()
		st.updateGaugeErr = errors.New("db failed")
		svc := NewMetricsService(st)
		req := &proto.UpdateMetricsRequest{}
		gaugeMetric := &proto.Metric{}
		gaugeMetric.SetId("Alloc")
		gaugeMetric.SetType(proto.Metric_GAUGE)
		gaugeMetric.SetValue(42.5)
		req.SetMetrics([]*proto.Metric{gaugeMetric})

		_, err := svc.UpdateMetrics(context.Background(), req)
		if err == nil {
			t.Fatal("expected error")
		}

		if got := err.Error(); got != "rpc error: code = Internal desc = db failed" {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("storage update counter error", func(t *testing.T) {
		st := newStorageMock()
		st.updateCounterErr = errors.New("counter failed")
		svc := NewMetricsService(st)

		req := &proto.UpdateMetricsRequest{}
		metric := &proto.Metric{}
		metric.SetId("PollCount")
		metric.SetType(proto.Metric_COUNTER)
		metric.SetDelta(7)
		req.SetMetrics([]*proto.Metric{metric})

		_, err := svc.UpdateMetrics(context.Background(), req)
		if err == nil {
			t.Fatal("expected error")
		}

		if got := err.Error(); got != "rpc error: code = Internal desc = counter failed" {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("empty metrics list", func(t *testing.T) {
		st := newStorageMock()
		svc := NewMetricsService(st)

		req := &proto.UpdateMetricsRequest{}
		req.SetMetrics([]*proto.Metric{})

		resp, err := svc.UpdateMetrics(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp == nil {
			t.Fatal("expected response, got nil")
		}
	})
}
