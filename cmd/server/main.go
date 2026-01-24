package main

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
	"log"
	"metrify/internal/audit"
	"metrify/internal/handler"
	"metrify/internal/router"
	"metrify/internal/service"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func main() {
	f := parseFlags()
	db := initDB(f.Dsn)
	ms := service.NewMemStorage(f.FileStorePath, db)
	logger := service.NewLogger()

	if f.Restore {
		err := ms.ReadFromFile(f.FileStorePath)

		if err != nil {
			log.Printf("could not read from file store: %v", err)
		}
	}

	go runMetricDumper(ms, f)
	err := run(ms, db, logger, f)

	if err != nil {
		log.Fatal(err.Error())
	}
}

func run(ms *service.MemStorage, db *sql.DB, logger *zap.SugaredLogger, f *flags) error {
	f.RunAddr = normalizeAddr(f.RunAddr)
	fmt.Println("Running server on", f.RunAddr)

	auditPublisher := initAuditPublisher(f)

	h := handler.NewHandler(
		ms,
		logger,
		db,
		auditPublisher,
		f.StoreInterval == 0,
		f.Key,
	)

	return http.ListenAndServe(f.RunAddr, router.Metric(h))
}

func runMetricDumper(ms *service.MemStorage, f *flags) {
	ticker := time.NewTicker(time.Duration(f.StoreInterval) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		err := ms.FlushToFile()

		if err != nil {
			log.Printf("cannot save metrics: %v", err)
		}
	}
}

func initDB(DSN string) *sql.DB {
	if !isValidPostgresDSN(DSN) {
		return nil
	}

	db, err := sql.Open("pgx", DSN)

	if err != nil {
		log.Fatal(err)
	}

	initMigrations(db)

	return db
}

func initMigrations(db *sql.DB) {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatal("postgres driver error: ", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres",
		driver,
	)
	if err != nil {
		log.Fatal("migrate init error: ", err)
	}

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("migrate up error: %v", err)
	}
}

func isValidPostgresDSN(dsn string) bool {
	dsn = strings.TrimSpace(dsn)
	if dsn == "" {
		return false
	}

	u, err := url.Parse(dsn)
	if err != nil {
		return false
	}

	if u.Scheme != "postgres" && u.Scheme != "postgresql" {
		return false
	}

	if u.Host == "" {
		return false
	}

	if u.Path == "" || u.Path == "/" {
		return false
	}

	return true
}

func normalizeAddr(addr string) string {
	if h, p, err := net.SplitHostPort(addr); err == nil {
		if h == "localhost" || h == "" {
			return ":" + p
		}
	}
	return addr
}

func initAuditPublisher(f *flags) *audit.Publisher {
	p := audit.NewPublisher()

	if f.AuditFile != "" {
		p.Add(audit.NewFileReceiver(f.AuditFile))
	}

	if f.AuditURL != "" {
		p.Add(audit.NewHTTPReceiver(
			f.AuditURL,
			&http.Client{Timeout: 3 * time.Second},
		))
	}

	return p
}
