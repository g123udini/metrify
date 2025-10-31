package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"metrify/internal/handler"
)

func Metric(handler *handler.Handler) chi.Router {
	r := chi.NewRouter()

	r.Use(handler.WithLogging)
	r.Use(handler.WithRequestCompress)
	r.Use(handler.WithResponseCompress)
	update(r, handler)
	get(r, handler)

	return r
}

func update(r chi.Router, handler *handler.Handler) {
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
