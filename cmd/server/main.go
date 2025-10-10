package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"metrify/internal/handler"
	"net/http"
)

func main() {
	r := chi.NewRouter()

	r.Route("/update", func(r chi.Router) {
		r.Use(middleware.AllowContentType(handler.TextUpdateContentType))
		r.Use(handler.MetricTypeMiddleware)
		r.Post("/counter/{name}/{value}", handler.UpdateCounter)
		r.Post("/gauge/{name}/{value}", handler.UpdateGauge)
	})

	r.Get("/", handler.GetList)

	r.Route("/value", func(r chi.Router) {
		r.Use(middleware.AllowContentType(handler.TextUpdateContentType))
		r.Use(handler.MetricTypeMiddleware)
		r.Get("/counter/{name}", handler.GetCounter)
		r.Get("/gauge/{name}", handler.GetGauge)
	})

	err := http.ListenAndServe(":8080", r)

	if err != nil {
		panic(err)
	}
}
