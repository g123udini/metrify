package main

import (
	"flag"
	"github.com/caarlos0/env"
	"log"
)

type flags struct {
	Host           string `env:"ADDRESS"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
}

func parseFlags() *flags {
	f := flags{
		Host:           ":8080",
		PollInterval:   2,
		ReportInterval: 10,
	}

	flag.StringVar(&f.Host, "a", f.Host, "address and host to run server")
	flag.IntVar(&f.PollInterval, "p", f.PollInterval, "interval between polls")
	flag.IntVar(&f.ReportInterval, "r", f.ReportInterval, "interval between reports")

	flag.Parse()

	err := env.Parse(&f)

	if err != nil {
		log.Fatal(err)
	}

	return &f
}
