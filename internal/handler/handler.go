package handler

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"metrify/internal/service"
	"net/http"
	"strconv"
)

const TextUpdateContentType = "text/plain"

var ms = service.NewMemStorage()

func GetList(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(ms)
}

func GetGauge(w http.ResponseWriter, r *http.Request) {
	metricName := chi.URLParam(r, "name")

	val, ok := ms.GetGauge(metricName)

	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	data := strconv.FormatFloat(val, 'f', -1, 64)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(data))
}

func GetCounter(w http.ResponseWriter, r *http.Request) {
	metricName := chi.URLParam(r, "name")

	val, ok := ms.GetCounter(metricName)

	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	data := strconv.FormatInt(val, 10)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(data))
}

func UpdateGauge(w http.ResponseWriter, r *http.Request) {
	metricName := chi.URLParam(r, "name")
	metricValue, err := strconv.ParseFloat(chi.URLParam(r, "value"), 64)

	if err != nil {
		http.Error(w, "Invalid metric value", http.StatusBadRequest)
		return
	}

	ms.UpdateGauge(metricName, metricValue)
}

func UpdateCounter(w http.ResponseWriter, r *http.Request) {
	metricName := chi.URLParam(r, "name")
	metricValue, err := strconv.ParseInt(chi.URLParam(r, "value"), 10, 64)

	if err != nil {
		http.Error(w, "Invalid metric value", http.StatusBadRequest)
		return
	}

	ms.UpdateCounter(metricName, metricValue)

	w.WriteHeader(http.StatusOK)
}

func MetricTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		metricType, err := service.ExtractType(path)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		if !service.ValidateMetricType(metricType) {
			http.Error(w, "Invalid metric type", http.StatusBadRequest)
			return
		}
		next.ServeHTTP(w, r)
	})
}
