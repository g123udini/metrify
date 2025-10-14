package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"metrify/internal/handler"
)

func Metric(handler *handler.Handler) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.AllowContentType(handler.UpdateContentType))
	update(r, handler)
	get(r, handler)

	return r
}

func update(r chi.Router, handler *handler.Handler) {
	r.Route("/update", func(r chi.Router) {
		r.Post("/counter/{name}/{value}", handler.UpdateCounter)
		r.Post("/gauge/{name}/{value}", handler.UpdateGauge)

		r.Post("/{type}/{name}/{value}", handler.InvalidMetricHandler)
	})
}

func get(r chi.Router, handler *handler.Handler) {
	r.Route("/value", func(r chi.Router) {
		r.Get("/counter/{name}", handler.GetCounter)
		r.Get("/gauge/{name}", handler.GetGauge)

		r.Get("/{type}/{name}", handler.InvalidMetricHandler)
	})
}
