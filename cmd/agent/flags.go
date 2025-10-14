package main

import (
	"flag"
)

type flags struct {
	Host           string
	PollInterval   int
	ReportInterval int
}

func parseFlags() *flags {
	var f flags

	flag.StringVar(&f.Host, "a", ":8080", "address and host to run server")
	flag.IntVar(&f.PollInterval, "p", 2, "interval between polls")
	flag.IntVar(&f.ReportInterval, "r", 10, "interval between reports")
	flag.Parse()

	return &f
}
