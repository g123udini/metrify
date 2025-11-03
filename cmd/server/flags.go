package main

import (
	"flag"
	"github.com/caarlos0/env"
	"log"
)

type flags struct {
	RunAddr       string `env:"ADDRESS"`
	StoreInterval int    `env:"STORE_ITERVAL"`
	FileStorePath string `env:"FILE_STORE_PATH"`
	Restore       bool   `env:"RESTORE"`
}

func parseFlags() *flags {
	f := flags{
		RunAddr:       ":8080",
		StoreInterval: 5,
		FileStorePath: "./metrics.json",
		Restore:       true,
	}

	flag.StringVar(&f.RunAddr, "a", f.RunAddr, "address and port to run server")
	flag.IntVar(&f.StoreInterval, "i", f.StoreInterval, "number of iterations")
	flag.StringVar(&f.FileStorePath, "f", f.FileStorePath, "path to store files")
	flag.BoolVar(&f.Restore, "r", f.Restore, "restore metrics")

	flag.Parse()

	err := env.Parse(&f)

	if err != nil {
		log.Fatal(err)
	}

	return &f
}
