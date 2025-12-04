package main

import (
	"context"
	"fmt"
	"metrify/internal/agent"
	models "metrify/internal/model"
	"metrify/internal/service"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	f := parseFlags()
	metricChan := make(chan models.Metrics, 1000) // буфер, чтобы коллекторы не стопорились
	normalizedHost := normalizeHost(f.Host)
	logger := service.NewLogger()
	client := agent.NewClient(normalizedHost, logger, f.Key)

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		defer wg.Done()
		runRuntimeCollector(ctx, time.Duration(f.PollInterval)*time.Second, metricChan)
	}()
	go func() {
		defer wg.Done()
		runGopsutilCollector(ctx, time.Duration(f.PollInterval)*time.Second, metricChan)
	}()

	for i := 0; i < f.RateLimit; i++ {
		go runSender(ctx, client, f, metricChan)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	<-sigCh
	cancel()
	wg.Wait()
	close(metricChan)
}

func runRuntimeCollector(
	ctx context.Context,
	pollInterval time.Duration,
	metricChan chan<- models.Metrics,
) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	pollCounter := int64(0)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m := agent.CollectGauge()
			for name, value := range m {
				metricChan <- models.Metrics{
					ID:    name,
					Value: &value,
					MType: models.Gauge,
				}
			}
			pollCounter++
			metricChan <- models.Metrics{
				ID:    "PollCount",
				Delta: &pollCounter,
				MType: models.Counter,
			}
		}
	}
}

func runGopsutilCollector(
	ctx context.Context,
	pollInterval time.Duration,
	metricChan chan<- models.Metrics,
) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m := agent.CollectGauge() // или отдельная CollectGopsutilGauge
			for name, value := range m {
				metricChan <- models.Metrics{
					ID:    name,
					Value: &value,
					MType: models.Gauge,
				}
			}
		}
	}
}

func runSender(
	ctx context.Context,
	client *agent.Client,
	f *flags,
	metricChan <-chan models.Metrics,
) {
	ticker := time.NewTicker(time.Duration(f.ReportInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			if f.BatchUpdate {
				var batch []models.Metrics

			drainBatch:
				for {
					select {
					case m, ok := <-metricChan:
						if !ok {
							if len(batch) > 0 {
								client.UpdateMetrics(batch)
							}
							return
						}
						batch = append(batch, m)

					default:
						if len(batch) > 0 {
							client.UpdateMetrics(batch)
						}
						break drainBatch
					}
				}
			} else {
			drainSingle:
				for {
					select {
					case m, ok := <-metricChan:
						if !ok {
							return
						}
						client.UpdateMetric(m)

					default:
						break drainSingle
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
