package main

import (
	"fmt"
	"log"
	"metrify/internal/agent"
	models "metrify/internal/model"
	"net"
	"time"
)

func main() {
	var (
		pollCount  int64
		gauges     map[string]float64
		lastReport = time.Now()
	)
	f := parseFlags()
	normalizedHost := normalizeHost(f.Host)
	metric := models.Metrics{}

	for {
		time.Sleep(time.Duration(f.PollInterval) * time.Second)

		gauges = agent.CollectGauge()
		pollCount++

		if time.Since(lastReport) >= time.Duration(f.ReportInterval)*time.Second {
			for key, val := range gauges {
				metric.ID = key
				metric.Value = &val
				metric.MType = models.Gauge

				if err := agent.UpdateMetric(normalizedHost, metric); err != nil {
					log.Printf("failed to update gauge %q: %v", key, err)
					continue
				}
			}

			metric.ID = "PollCount"
			metric.Delta = &pollCount
			metric.MType = models.Counter

			if err := agent.UpdateMetric(normalizedHost, metric); err != nil {
				log.Printf("failed to update counter %s: %v", metric.ID, err)
				continue
			}

			lastReport = time.Now()
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
