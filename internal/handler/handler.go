package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	models "metrify/internal/model"
	"metrify/internal/service"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Handler struct {
	ms                 service.Storage
	logger             *zap.SugaredLogger
	db                 *sql.DB
	dumpToFile         bool
	AllowedContentType string
}

func NewHandler(ms service.Storage, logger *zap.SugaredLogger, db *sql.DB, dump bool) *Handler {
	return &Handler{
		ms:                 ms,
		logger:             logger,
		db:                 db,
		dumpToFile:         dump,
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
	defer r.Body.Close()

	if err := dec.Decode(&metric); err != nil {
		handler.logger.Debug("Error decoding JSON", zap.Error(err))
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
		handler.logger.Error("Error encoding JSON", zap.Error(err))
	}
}

func (handler *Handler) UpdateMetrics(w http.ResponseWriter, r *http.Request) {
	dec := json.NewDecoder(r.Body)
	metric := models.Metrics{}
	defer r.Body.Close()

	if err := dec.Decode(&metric); err != nil {
		handler.logger.Debug("Error decoding JSON", zap.Error(err))
	}

	if metric.Value == nil && metric.Delta == nil {
		http.Error(w, "Error JSON format", http.StatusBadRequest)
		return
	}

	if metric.MType == models.Gauge {
		handler.ms.UpdateGauge(metric.ID, *metric.Value)
	} else {
		handler.ms.UpdateCounter(metric.ID, *metric.Delta)
	}

	if handler.dumpToFile {
		err := handler.ms.FlushToFile()

		if err != nil {
			handler.logger.Error("Error flushing to file", zap.Error(err))
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "ok"}`))
}

func (handler *Handler) UpdateMetricsBatch(w http.ResponseWriter, r *http.Request) {
	dec := json.NewDecoder(r.Body)
	var metrics []models.Metrics
	var err error = nil
	var errs []error
	defer r.Body.Close()

	if err = dec.Decode(&metrics); err != nil {
		handler.logger.Debug("Error decoding JSON", zap.Error(err))
	}

	for _, metric := range metrics {
		if metric.MType == models.Gauge {
			err = handler.ms.UpdateGauge(metric.ID, *metric.Value)

			if err != nil {
				errs = append(errs, err)
			}
		} else {
			err = handler.ms.UpdateCounter(metric.ID, *metric.Delta)

			if err != nil {
				errs = append(errs, err)
			}
		}
	}

	if len(errs) > 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"errors": errs,
		})
		return
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
		loggingWriter := service.NewLoggingResponseWriter(w)
		start := time.Now()
		uri := r.RequestURI
		method := r.Method

		h.ServeHTTP(loggingWriter, r) // обслуживание оригинального запроса
		end := time.Now()
		duration := end.Sub(start)

		handler.logger.Infoln(
			"uri", uri,
			"method", method,
			"duration", duration,
			"size", loggingWriter.ResponseData.Size,
			"status", loggingWriter.ResponseData.Status,
		)

	}

	return http.HandlerFunc(logFn)
}

func (handler *Handler) WithRequestCompress(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			h.ServeHTTP(w, r)
			return
		}

		dr, err := service.NewCompressReader(r.Body)

		if err != nil {
			http.Error(w, "invalid gzip", http.StatusBadRequest)
			return
		}
		r.Body = dr
		h.ServeHTTP(w, r)
	})
}

func (handler *Handler) WithResponseCompress(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			h.ServeHTTP(w, r)
			return
		}

		cw := service.NewCompressWriter(w)
		w = cw
		defer cw.Close()
		cw.WriteHeader(http.StatusOK)

		h.ServeHTTP(w, r)
	})
}

func (handler *Handler) GetInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}

func (handler *Handler) Ping(w http.ResponseWriter, r *http.Request) {
	if handler.db == nil {
		http.Error(w, "database is not initialized", http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := handler.db.PingContext(ctx); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "ok"}`))
}
