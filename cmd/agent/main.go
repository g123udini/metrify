package main

import (
	"metrify/internal/agent"
	models "metrify/internal/model"
	"strconv"
	"time"
)

const (
	pollInterval   = 2 * time.Second
	reportInterval = 10 * time.Second
)

func main() {
	var (
		pollCount  int64
		gauges     map[string]float64
		lastReport = time.Now()
	)

	for {
		time.Sleep(pollInterval)

		gauges = agent.CollectGauge()
		pollCount++

		if time.Since(lastReport) >= reportInterval {
			for key, metric := range gauges {
				val := strconv.FormatFloat(metric, 'f', -1, 64)
				if err := agent.SendMetric(models.Gauge, key, val); err != nil {
					panic(err)
				}
			}

			val := strconv.FormatInt(pollCount, 10)
			if err := agent.SendMetric(models.Counter, "PollCount", val); err != nil {
				panic(err)
			}

			lastReport = time.Now()
		}
	}
}
