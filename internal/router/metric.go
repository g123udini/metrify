// Package router содержит HTTP-роуты сервиса metrify.
package router

// Metric возвращает chi.Router с зарегистрированными middleware и эндпоинтами метрик.
//
// Маршруты:
//   GET  /           - info
//   GET  /ping       - db ping
//   POST /updates/   - batch update (JSON)
//   POST /update/    - update (JSON)
//   POST /update/counter/{name}/{value} - update counter (text/plain)
//   POST /update/gauge/{name}/{value}   - update gauge (text/plain)
//   POST /value/     - get metric by body (JSON)
//   GET  /value/counter/{name}          - get counter (text/plain)
//   GET  /value/gauge/{name}            - get gauge (text/plain)
import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
	_ "metrify/docs"
	"metrify/internal/handler"
)

func Metric(handler *handler.Handler) chi.Router {
	r := chi.NewRouter()

	r.Use(handler.WithLogging)
	r.Use(handler.WithRequestCompress)
	r.Use(handler.WithResponseCompress)
	r.Use(handler.WithHashedRequest)
	r.Use(handler.WithDecrypt)
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	update(r, handler)
	get(r, handler)

	return r
}

func update(r chi.Router, handler *handler.Handler) {
	r.Route("/updates", func(r chi.Router) {
		r.With(middleware.AllowContentType("application/json")).
			Post("/", handler.UpdateMetricsBatch)
	})

	r.Route("/update", func(r chi.Router) {
		r.With(middleware.AllowContentType("application/json")).
			Post("/", handler.UpdateMetrics)

		r.With(middleware.AllowContentType(handler.AllowedContentType)).
			Post("/counter/{name}/{value}", handler.UpdateCounter)
		r.With(middleware.AllowContentType(handler.AllowedContentType)).
			Post("/gauge/{name}/{value}", handler.UpdateGauge)
		r.With(middleware.AllowContentType(handler.AllowedContentType)).
			Post("/{type}/{name}/{value}", handler.InvalidMetricHandler)
	})
}

func get(r chi.Router, handler *handler.Handler) {
	r.Get("/", handler.GetInfo)
	r.Get("/ping", handler.Ping)

	r.Route("/value", func(r chi.Router) {
		r.With(middleware.AllowContentType("application/json")).
			Post("/", handler.GetMetrics)

		r.With(middleware.AllowContentType(handler.AllowedContentType)).
			Get("/counter/{name}", handler.GetCounter)
		r.With(middleware.AllowContentType(handler.AllowedContentType)).
			Get("/gauge/{name}", handler.GetGauge)
		r.With(middleware.AllowContentType(handler.AllowedContentType)).
			Get("/{type}/{name}", handler.InvalidMetricHandler)
	})
}
