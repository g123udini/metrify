package main

import "flag"

type flags struct {
	RunAddr string `env:"ADDRESS"`
}

func parseFlags() *flags {
	f := &flags{
		RunAddr: ":8080",
	}

	flag.StringVar(&f.RunAddr, "a", f.RunAddr, "address and port to run server")
	flag.Parse()

	return f
}
