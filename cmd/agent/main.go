package main

import (
	"fmt"
	"log"
	"metrify/internal/agent"
	models "metrify/internal/model"
	"net"
	"strconv"
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

	for {
		time.Sleep(time.Duration(f.PollInterval) * time.Second)

		gauges = agent.CollectGauge()
		pollCount++

		if time.Since(lastReport) >= time.Duration(f.ReportInterval)*time.Second {
			for key, metric := range gauges {
				val := strconv.FormatFloat(metric, 'f', -1, 64)
				if err := agent.UpdateMetric(normalizedHost, models.Gauge, key, val); err != nil {
					log.Printf("failed to update gauge %q: %v", key, err)
					continue
				}
			}

			val := strconv.FormatInt(pollCount, 10)
			counterName := "PollCount"

			if err := agent.UpdateMetric(normalizedHost, models.Counter, counterName, val); err != nil {
				log.Printf("failed to update counter %s: %v", counterName, err)
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
