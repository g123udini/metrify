package main

import (
	"flag"
)

var (
	host           string
	pollInterval   int
	reportInterval int
)

func parsesFlags() {
	flag.StringVar(&host, "a", ":8080", "address and host to run server")
	flag.IntVar(&pollInterval, "p", 2, "interval between polls")
	flag.IntVar(&reportInterval, "r", 10, "interval between reports")
	flag.Parse()
}
