package main

import (
	"flag"
	"github.com/caarlos0/env"
	"log"
)

type flags struct {
	RunAddr string `env:"ADDRESS"`
}

func parseFlags() *flags {
	f := flags{
		RunAddr: ":8080",
	}

	flag.StringVar(&f.RunAddr, "a", f.RunAddr, "address and port to run server")
	flag.Parse()

	flag.Parse()

	err := env.Parse(&f)

	if err != nil {
		log.Fatal(err)
	}

	return &f
}
