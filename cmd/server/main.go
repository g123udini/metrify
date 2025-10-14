package main

import (
	"fmt"
	"log"
	"metrify/internal/handler"
	"metrify/internal/router"
	"metrify/internal/service"
	"net"
	"net/http"
)

func main() {
	parseFlags()

	if err := run(); err != nil {
		log.Fatal(err.Error())
	}
}

func run() error {
	fmt.Println("Running server on", flagRunAddr)
	if h, p, err := net.SplitHostPort(flagRunAddr); err == nil {
		if h == "localhost" || h == "" {
			flagRunAddr = ":" + p
		}
	}

	ms := service.NewMemStorage()
	h := handler.NewHandler(ms)

	return http.ListenAndServe(flagRunAddr, router.Metric(h))
}
