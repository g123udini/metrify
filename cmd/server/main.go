package main

import (
	"fmt"
	"log"
	"metrify/internal/router"
	"net"
	"net/http"
)

func main() {
	parseFlags()

	if err := run(); err != nil {
		log.Fatalf(err.Error())
	}
}

func run() error {
	fmt.Println("Running server on", flagRunAddr)
	if h, p, err := net.SplitHostPort(flagRunAddr); err == nil {
		if h == "localhost" || h == "" {
			flagRunAddr = ":" + p
		}
	}
	return http.ListenAndServe(flagRunAddr, router.Metric())
}
