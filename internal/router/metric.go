package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"metrify/internal/handler"
)

var r = chi.NewRouter()

func Metric() chi.Router {
	r.Use(middleware.AllowContentType(handler.TextUpdateContentType))
	update()
	get()

	return r
}

func update() {
	r.Route("/update", func(r chi.Router) {
		r.Post("/counter/{name}/{value}", handler.UpdateCounter)
		r.Post("/gauge/{name}/{value}", handler.UpdateGauge)

		r.Post("/{type}", handler.InvalidMetricHandler)
	})
}

func get() {
	r.Route("/value", func(r chi.Router) {
		r.Get("/counter/{name}", handler.GetCounter)
		r.Get("/gauge/{name}", handler.GetGauge)

		r.Get("/{type}", handler.InvalidMetricHandler)
	})
}
