package main

import (
	"metrify/internal/agent"
	models "metrify/internal/model"
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

	for {
		time.Sleep(time.Duration(pollInterval) * time.Second)

		gauges = agent.CollectGauge()
		pollCount++

		if time.Since(lastReport) >= time.Duration(reportInterval)*time.Second {
			for key, metric := range gauges {
				val := strconv.FormatFloat(metric, 'f', -1, 64)
				if err := agent.UpdateMetric(host, models.Gauge, key, val); err != nil {
					panic(err)
				}
			}

			val := strconv.FormatInt(pollCount, 10)
			if err := agent.UpdateMetric(host, models.Counter, "PollCount", val); err != nil {
				panic(err)
			}

			lastReport = time.Now()
		}
	}
}
