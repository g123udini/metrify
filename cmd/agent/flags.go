package main

import (
	"flag"
	"time"
)

var (
	port           string
	pollInterval   time.Duration
	reportInterval time.Duration
)

func parsesFlags() {
	flag.StringVar(&port, "a", ":8080", "address and port to run server")
	flag.DurationVar(&pollInterval, "p", 2*time.Second, "interval between polls")
	flag.DurationVar(&reportInterval, "r", 10*time.Second, "interval between reports")
	flag.Parse()
}
