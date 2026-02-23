package main

import (
	"flag"
	"github.com/caarlos0/env"
	"log"
	"metrify/internal/agent"
	"metrify/internal/service"
	"os"
)

// generate:reset
type flags struct {
	Host           string `env:"ADDRESS"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	BatchUpdate    bool   `env:"BATCH_UPDATE"`
	Key            string `env:"KEY"`
	RateLimit      int    `env:"RATE_LIMIT"`
	CryptoKey      string `env:"CRYPTO_KEY"`
	Config         string `env:"CONFIG"`
}

func parseFlags() *flags {
	cnfPath := getConfigPath()
	f := flags{}

	if cnfPath != "" {
		config, err := service.FromFile[agent.Config](cnfPath)
		if err != nil {
			log.Fatalf("Could not load config file: %s", cnfPath)
		}

		f.Host = config.Address
		f.ReportInterval = config.ReportInterval
		f.PollInterval = config.PollInterval
		f.CryptoKey = config.CryptoKey
	} else {
		setDefaults(&f)
	}

	if err := env.Parse(&f); err != nil {
		log.Fatal(err)
	}

	flag.StringVar(&f.Host, "address", f.Host, "The address to listen on")
	flag.StringVar(&f.Host, "a", f.Host, "address and host to run server")
	flag.IntVar(&f.PollInterval, "p", f.PollInterval, "interval between polls")
	flag.IntVar(&f.ReportInterval, "r", f.ReportInterval, "interval between reports")
	flag.BoolVar(&f.BatchUpdate, "b", f.BatchUpdate, "send metrics in batches")
	flag.StringVar(&f.Key, "k", f.Key, "private key to use for authentication")
	flag.IntVar(&f.RateLimit, "l", f.RateLimit, "rate limit")
	flag.StringVar(&f.CryptoKey, "crypto-key", f.CryptoKey, "crypto key")
	flag.StringVar(&f.Config, "config", f.Config, "configuration file")

	flag.Parse()

	return &f
}

func getConfigPath() string {
	var cfg string

	fs := flag.NewFlagSet("config", flag.ContinueOnError)
	fs.StringVar(&cfg, "config", "", "configuration file")
	fs.StringVar(&cfg, "c", "", "configuration file")

	_ = fs.Parse(os.Args[1:])

	if cfg != "" {
		return cfg
	}

	if v := os.Getenv("CONFIG"); v != "" {
		return v
	}

	return ""
}

func setDefaults(f *flags) {
	f.Host = ":8080"
	f.PollInterval = 2
	f.ReportInterval = 10
	f.BatchUpdate = false
	f.Key = ""
	f.RateLimit = 1
	f.CryptoKey = ""
	f.Config = ""
}
