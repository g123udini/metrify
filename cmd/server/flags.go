package main

import "flag"

func parseFlags() string {
	var flagRunAddr string

	flag.StringVar(&flagRunAddr, "a", ":8080", "address and port to run server")
	flag.Parse()

	return flagRunAddr
}
