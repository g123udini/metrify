package main

import (
	"flag"
	"time"
)

var (
	host           = "http://localhost:8080"
	pollInterval   = 2 * time.Second
	reportInterval = 10 * time.Second
)

func parsesFlags() {
	flag.StringVar(&host, "a", "http://localhost:8080", "address and port to run server")
	flag.DurationVar(&pollInterval, "p", 2*time.Second, "interval between polls")
	flag.DurationVar(&reportInterval, "r", 10*time.Second, "interval between reports")
	flag.Parse()
}
