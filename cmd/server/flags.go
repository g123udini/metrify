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
	Dsn           string `env:"DATABASE_DSN"`
	Key           string `env:"KEY"`
	AuditFile     string `env:"AUDIT_FILE"`
	AuditURL      string `env:"AUDIT_URL"`
}

func parseFlags() *flags {
	f := flags{
		RunAddr:       ":8080",
		StoreInterval: 5,
		FileStorePath: "./metrics.json",
		Restore:       true,
		Dsn:           "",
		Key:           "",
		AuditFile:     "",
		AuditURL:      "",
	}

	flag.StringVar(&f.RunAddr, "a", f.RunAddr, "address and port to run server")
	flag.IntVar(&f.StoreInterval, "i", f.StoreInterval, "number of iterations")
	flag.StringVar(&f.FileStorePath, "f", f.FileStorePath, "path to store files")
	flag.BoolVar(&f.Restore, "r", f.Restore, "restore metrics")
	flag.StringVar(&f.Dsn, "d", f.Dsn, "database connection string")
	flag.StringVar(&f.Key, "k", f.Key, "key to use for encryption")
	flag.StringVar(&f.AuditFile, "audit-file", f.AuditFile, "path to audit log file (disables audit if empty)")
	flag.StringVar(&f.AuditURL, "audit-url", f.AuditURL, "audit receiver URL (POST, disables audit if empty)")

	flag.Parse()

	err := env.Parse(&f)

	if err != nil {
		log.Fatal(err)
	}

	return &f
}
