package main

import (
	"metrify/internal/handler"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.Handle("/update/", handler.Middleware(http.HandlerFunc(handler.Update)))

	err := http.ListenAndServe(":8080", mux)

	if err != nil {
		panic(err)
	}
}
