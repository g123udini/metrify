package main

import (
	"flag"
	"github.com/caarlos0/env"
	"log"
	"metrify/internal/config"
	"metrify/internal/service"
	"os"
	"time"
)

type flags struct {
	RunAddr            string        `env:"ADDRESS"`
	StoreInterval      int           `env:"STORE_INTERVAL"`
	FileStorePath      string        `env:"FILE_STORE_PATH"`
	Restore            bool          `env:"RESTORE"`
	Dsn                string        `env:"DATABASE_DSN"`
	Key                string        `env:"KEY"`
	AuditFile          string        `env:"AUDIT_FILE"`
	AuditURL           string        `env:"AUDIT_URL"`
	CPUProfileFile     string        `env:"CPU_PROFILE_FILE"`
	CPUProfileDuration time.Duration `env:"CPU_PROFILE_DURATION"`
	MemProfileFile     string        `env:"MEM_PROFILE_FILE"`
	CryptoKey          string        `env:"CRYPTO_KEY"`
}

func parseFlags() *flags {
	cnfPath := getConfigPath()
	f := flags{}

	if cnfPath != "" {
		servConfig, err := service.FromFile[config.Config](cnfPath)
		if err != nil {
			log.Fatalf("Could not load servConfig file: %s", cnfPath)
		}

		f.RunAddr = servConfig.Address
		f.Restore = servConfig.Restore
		f.FileStorePath = servConfig.StoreFile
		f.Dsn = servConfig.DatabaseDsn
		f.CryptoKey = servConfig.CryptoKey

		if servConfig.StoreInterval != "" {
			d, err := time.ParseDuration(servConfig.StoreInterval)
			if err != nil {
				log.Fatalf("invalid store_interval in servConfig: %v", err)
			}
			f.StoreInterval = int(d.Seconds())
		}
	} else {
		setDefaults(&f)
	}

	if err := env.Parse(&f); err != nil {
		log.Fatal(err)
	}

	flag.StringVar(&f.RunAddr, "a", f.RunAddr, "address and port to run server")
	flag.IntVar(&f.StoreInterval, "i", f.StoreInterval, "number of iterations")
	flag.StringVar(&f.FileStorePath, "f", f.FileStorePath, "path to store files")
	flag.BoolVar(&f.Restore, "r", f.Restore, "restore metrics")
	flag.StringVar(&f.Dsn, "d", f.Dsn, "database connection string")
	flag.StringVar(&f.Key, "k", f.Key, "key to use for encryption")
	flag.StringVar(&f.AuditFile, "audit-file", f.AuditFile, "path to audit log file (disables audit if empty)")
	flag.StringVar(&f.AuditURL, "audit-url", f.AuditURL, "audit receiver URL (POST, disables audit if empty)")
	flag.StringVar(&f.CPUProfileFile, "cpu-profile-file", f.CPUProfileFile, "path to CPU profile file")
	flag.DurationVar(&f.CPUProfileDuration, "cpu-profile-duration", f.CPUProfileDuration, "path to CPU profile duration")
	flag.StringVar(&f.MemProfileFile, "mem-profile-file", f.MemProfileFile, "path to memory profile file")
	flag.StringVar(&f.CryptoKey, "crypto-key", f.CryptoKey, "crypto key")

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
	f.RunAddr = ":8080"
	f.StoreInterval = 5
	f.FileStorePath = "./metrics.json"
	f.Restore = true
	f.Dsn = ""
	f.Key = ""
	f.AuditFile = ""
	f.AuditURL = ""
	f.CPUProfileFile = ""
	f.CPUProfileDuration = 0
	f.MemProfileFile = ""
	f.CryptoKey = ""
}
