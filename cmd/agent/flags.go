package main

import (
	"flag"
)

var (
	port           string
	pollInterval   int
	reportInterval int
)

func parsesFlags() {
	flag.StringVar(&port, "a", ":8080", "address and port to run server")
	flag.IntVar(&pollInterval, "p", 2, "interval between polls")
	flag.IntVar(&reportInterval, "r", 10, "interval between reports")
	flag.Parse()
}
