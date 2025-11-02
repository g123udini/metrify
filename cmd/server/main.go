package main

import (
	"fmt"
	"log"
	"metrify/internal/handler"
	"metrify/internal/router"
	"metrify/internal/service"
	"net"
	"net/http"
	"time"
)

func main() {
	f := parseFlags()
	ms := service.NewMemStorage(f.FileStorePath)

	if !f.Restore {
		err := ms.ReadFromFile(f.FileStorePath)

		if err != nil {
			log.Printf("could not read from file store: %v", err)
		}
	}

	go runMetricDumper(ms, f)
	err := run(ms, f)

	if err != nil {
		log.Fatal(err.Error())
	}
}

func run(ms *service.MemStorage, f *flags) error {
	fmt.Println("Running server on", f.RunAddr)
	if h, p, err := net.SplitHostPort(f.RunAddr); err == nil {
		if h == "localhost" || h == "" {
			f.RunAddr = ":" + p
		}
	}
	h := handler.NewHandler(ms, f.StoreIterval == 0)

	return http.ListenAndServe(f.RunAddr, router.Metric(h))
}

func runMetricDumper(ms *service.MemStorage, f *flags) {
	ticker := time.NewTicker(time.Duration(f.StoreIterval) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		err := ms.FlushToFile()

		if err != nil {
			log.Printf("cannot save metrics: %v", err)
		}
	}
}
