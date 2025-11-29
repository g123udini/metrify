package main

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"metrify/internal/agent"
	models "metrify/internal/model"
	"metrify/internal/service"
	"net"
	"sync"
	"time"
)

func main() {
	f := parseFlags()
	logger := service.NewLogger()
	normalizedHost := normalizeHost(f.Host)
	client := agent.NewClient(normalizedHost, logger, f.Key)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	gaugesCh := make(chan map[string]float64)
	jobs := make(chan []models.Metrics, 10)

	go runRuntimeCollector(ctx, time.Duration(f.PollInterval)*time.Second, gaugesCh)
	go runGopsutilCollector(ctx, time.Duration(f.PollInterval)*time.Second, gaugesCh)

	go runCollector(ctx, f, gaugesCh, jobs)

	workerCount := f.RateLimit

	var wg sync.WaitGroup
	wg.Add(workerCount)

	for i := 0; i <= workerCount; i++ {
		go func(id int) {
			defer wg.Done()
			runSender(ctx, id, jobs, client, f.BatchUpdate, logger)
		}(i + 1)
	}
}

func runCollector(
	ctx context.Context,
	f *flags,
	gaugesCh <-chan map[string]float64,
	jobs chan<- []models.Metrics,
) {
	defer close(jobs)

	reportTicker := time.NewTicker(time.Duration(f.ReportInterval) * time.Second)
	defer reportTicker.Stop()

	gauges := make(map[string]float64)
	var pollCount int64

	for {
		select {
		case <-ctx.Done():
			return

		case m := <-gaugesCh:
			for k, v := range m {
				gauges[k] = v
			}
			pollCount++

		case <-reportTicker.C:
			if len(gauges) == 0 {
				continue
			}

			var batch []models.Metrics

			for key, val := range gauges {
				v := val
				batch = append(batch, models.Metrics{
					ID:    key,
					Value: &v,
					MType: models.Gauge,
				})
			}

			pc := pollCount
			batch = append(batch, models.Metrics{
				ID:    "PollCount",
				Delta: &pc,
				MType: models.Counter,
			})

			select {
			case <-ctx.Done():
				return
			case jobs <- batch:
			}
		}
	}
}

func runRuntimeCollector(
	ctx context.Context,
	pollInterval time.Duration,
	out chan<- map[string]float64,
) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m := agent.CollectGauge()
			select {
			case <-ctx.Done():
				return
			case out <- m:
			}
		}
	}
}

func runGopsutilCollector(
	ctx context.Context,
	pollInterval time.Duration,
	out chan<- map[string]float64,
) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m := agent.CollectGopsutilGauges()
			select {
			case <-ctx.Done():
				return
			case out <- m:
			}
		}
	}
}
func runSender(
	ctx context.Context,
	id int,
	jobs <-chan []models.Metrics,
	client *agent.Client,
	batchUpdate bool,
	logger *zap.SugaredLogger,
) {
	for {
		select {
		case <-ctx.Done():
			return
		case batch, ok := <-jobs:
			if !ok {
				return
			}

			if batchUpdate {
				if err := client.UpdateMetrics(batch); err != nil {
					logger.Error(fmt.Sprintf("worker %d: UpdateMetrics error: %v", id, err))
				}
			} else {
				for _, m := range batch {
					if err := client.UpdateMetric(m); err != nil {
						logger.Error(fmt.Sprintf("worker %d: UpdateMetric error: %v", id, err))
					}
				}
			}
		}
	}
}

func normalizeHost(host string) string {
	if h, p, err := net.SplitHostPort(host); err == nil {
		if h == "" {
			host = fmt.Sprintf("localhost:%s", p)
		}
	}
	return host
}
