package handler

import (
	"encoding/json"
	models "metrify/internal/model"
	"metrify/internal/service"
	"net/http"
)

const UpdateContentType = "text/plain"

var ms = service.NewMemStorage()

func Get(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(ms)
}

func Update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if UpdateContentType != r.Header.Get("Content-Type") {
		http.Error(w, "Unsupported Media Type", http.StatusUnsupportedMediaType)
		return
	}

	path := r.URL.Path
	metricType, _ := service.ExtractType(path)
	metricName, err := service.ExtractName(path)

	if err != nil {
		http.Error(w, "Page not found", http.StatusNotFound)
		return
	}

	if !service.ValidateMetricType(metricType) {
		http.Error(w, "Invalid metric type", http.StatusBadRequest)
		return
	}

	if metricType == models.Gauge {
		metricValue, err := service.ExtractGaugeValue(path)

		if err != nil {
			http.Error(w, "Invalid metric value", http.StatusBadRequest)
			return
		}

		ms.UpdateGauge(metricName, metricValue)
	} else {
		metricValue, err := service.ExtractCounterValue(path)

		if err != nil {
			http.Error(w, "Invalid metric value", http.StatusBadRequest)
			return
		}

		ms.UpdateCounter(metricName, metricValue)
	}

	w.WriteHeader(http.StatusOK)
}

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}
