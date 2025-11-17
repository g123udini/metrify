package main

import (
	"fmt"
	"metrify/internal/agent"
	models "metrify/internal/model"
	"metrify/internal/service"
	"net"
	"time"
)

func main() {
	var (
		pollCount   int64
		gauges      map[string]float64
		lastReport  = time.Now()
		metricBatch []models.Metrics
		maxRetry    = 3
	)
	f := parseFlags()
	normalizedHost := normalizeHost(f.Host)
	metric := models.Metrics{}
	logger := service.NewLogger()
	client := agent.NewClient(normalizedHost, logger)

	for {
		time.Sleep(time.Duration(f.PollInterval) * time.Second)

		gauges = agent.CollectGauge()
		pollCount++

		if time.Since(lastReport) >= time.Duration(f.ReportInterval)*time.Second {
			for key, val := range gauges {
				metric.ID = key
				metric.Value = &val
				metric.MType = models.Gauge

				err := service.Retry(maxRetry, 2*time.Second, func() error {
					return client.UpdateMetric(metric)
				})

				if err != nil {
					logger.Error(err.Error())
				}

				metricBatch = append(metricBatch, metric)
			}

			metric.ID = "PollCount"
			metric.Delta = &pollCount
			metric.MType = models.Counter

			metricBatch = append(metricBatch, metric)

			err := service.Retry(maxRetry, 2*time.Second, func() error {
				return client.UpdateMetric(metric)
			})

			if err != nil {
				logger.Error(err.Error())
			}

			lastReport = time.Now()
		}

		err := service.Retry(maxRetry, 2*time.Second, func() error {
			return client.UpdateMetrics(metricBatch)
		})

		if err != nil {
			logger.Error(err.Error())
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
