package handler

import (
	"github.com/go-chi/chi/v5"
	"metrify/internal/logger"
	"metrify/internal/service"
	"net/http"
	"strconv"
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
