package main

import (
	"fmt"
	"metrify/internal/router"
	"net"
	"net/http"
)

func main() {
	parseFlags()

	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	fmt.Println("Running server on", flagRunAddr)
	if h, p, err := net.SplitHostPort(flagRunAddr); err == nil {
		if h == "localhost" || h == "" {
			flagRunAddr = ":" + p // нормализуем к ":38849"
		}
	}
	return http.ListenAndServe(flagRunAddr, router.Metric())
}
