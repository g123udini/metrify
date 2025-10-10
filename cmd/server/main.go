package main

import (
	"metrify/internal/router"
	"net/http"
)

func main() {
	err := http.ListenAndServe(":8080", router.Metric())

	if err != nil {
		panic(err)
	}
}
