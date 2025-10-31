package handler

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"metrify/internal/compresser"
	"metrify/internal/logger"
	models "metrify/internal/model"
	"metrify/internal/service"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Handler struct {
	ms                 service.Storage
	AllowedContentType string
}

func NewHandler(ms service.Storage) *Handler {
	return &Handler{
		ms:                 ms,
		AllowedContentType: "text/plain",
	}
}

func (handler *Handler) GetGauge(w http.ResponseWriter, r *http.Request) {
	metricName := chi.URLParam(r, "name")

	val, ok := handler.ms.GetGauge(metricName)

	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	data := strconv.FormatFloat(val, 'f', -1, 64)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(data))
}

func (handler *Handler) GetCounter(w http.ResponseWriter, r *http.Request) {
	metricName := chi.URLParam(r, "name")

	val, ok := handler.ms.GetCounter(metricName)

	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	data := strconv.FormatInt(val, 10)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(data))
}

func (handler *Handler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	dec := json.NewDecoder(r.Body)
	metric := models.Metrics{}
	sugar := logger.NewLogger()
	defer r.Body.Close()

	if err := dec.Decode(&metric); err != nil {
		sugar.Debug("Error decoding JSON", zap.Error(err))
	}

	if metric.MType == models.Gauge {
		val, ok := handler.ms.GetGauge(metric.ID)

		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		metric.Value = &val
	} else {
		val, ok := handler.ms.GetCounter(metric.ID)

		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		metric.Delta = &val
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(metric); err != nil {
		sugar.Error("Error encoding JSON", zap.Error(err))
	}
}

func (handler *Handler) UpdateMetrics(w http.ResponseWriter, r *http.Request) {
	dec := json.NewDecoder(r.Body)
	metric := models.Metrics{}
	sugar := logger.NewLogger()
	defer r.Body.Close()

	if err := dec.Decode(&metric); err != nil {
		sugar.Debug("Error decoding JSON", zap.Error(err))
	}

	if metric.MType == models.Gauge {
		handler.ms.UpdateGauge(metric.ID, *metric.Value)
	} else {
		handler.ms.UpdateCounter(metric.ID, *metric.Delta)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "ok"}`))
}

func (handler *Handler) UpdateGauge(w http.ResponseWriter, r *http.Request) {
	metricName := chi.URLParam(r, "name")
	metricValue, err := strconv.ParseFloat(chi.URLParam(r, "value"), 64)

	if err != nil {
		http.Error(w, "Invalid metric value", http.StatusBadRequest)
		return
	}

	handler.ms.UpdateGauge(metricName, metricValue)
}

func (handler *Handler) UpdateCounter(w http.ResponseWriter, r *http.Request) {
	metricName := chi.URLParam(r, "name")
	metricValue, err := strconv.ParseInt(chi.URLParam(r, "value"), 10, 64)

	if err != nil {
		http.Error(w, "Invalid metric value", http.StatusBadRequest)
		return
	}

	handler.ms.UpdateCounter(metricName, metricValue)

	w.WriteHeader(http.StatusOK)
}

func (handler *Handler) InvalidMetricHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "invalid metric type (expect counter|gauge)", http.StatusBadRequest)
}

func (handler *Handler) WithLogging(h http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		sugar := logger.NewLogger()
		loggingWriter := logger.NewLoggingResponseWriter(w)
		start := time.Now()
		uri := r.RequestURI
		method := r.Method

		h.ServeHTTP(loggingWriter, r) // обслуживание оригинального запроса
		end := time.Now()
		duration := end.Sub(start)

		sugar.Infoln(
			"uri", uri,
			"method", method,
			"duration", duration,
			"size", loggingWriter.ResponseData.Size,
			"status", loggingWriter.ResponseData.Status,
		)

	}

	return http.HandlerFunc(logFn)
}

func (handler *Handler) WithCompress(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			h.ServeHTTP(w, r)
			return
		}

		dr, err := compresser.NewCompressReader(r.Body)

		if err != nil {
			http.Error(w, "invalid gzip", http.StatusBadRequest)
			return
		}
		r.Body = dr

		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			w = compresser.NewCompressWriter(w)
		}

		h.ServeHTTP(w, r)
	})
}
