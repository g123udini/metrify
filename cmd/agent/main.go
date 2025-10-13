package main

import (
	"fmt"
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
	parsesFlags()
	normalizedHost := normalizeHost(host)

	for {
		time.Sleep(time.Duration(pollInterval) * time.Second)

		gauges = agent.CollectGauge()
		pollCount++

		if time.Since(lastReport) >= time.Duration(reportInterval)*time.Second {
			for key, metric := range gauges {
				val := strconv.FormatFloat(metric, 'f', -1, 64)
				if err := agent.UpdateMetric(normalizedHost, models.Gauge, key, val); err != nil {
					panic(err)
				}
			}

			val := strconv.FormatInt(pollCount, 10)
			if err := agent.UpdateMetric(normalizedHost, models.Counter, "PollCount", val); err != nil {
				panic(err)
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
