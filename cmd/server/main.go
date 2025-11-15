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
	"metrify/internal/handler"
	"metrify/internal/router"
	"metrify/internal/service"
	"net"
	"net/http"
	"time"
)

func main() {
	f := parseFlags()
	ms := service.NewMemStorage(f.FileStorePath)
	logger := service.NewLogger()

	if f.Restore {
		err := ms.ReadFromFile(f.FileStorePath)

		if err != nil {
			log.Printf("could not read from file store: %v", err)
		}
	}

	go runMetricDumper(ms, f)
	err := run(ms, logger, f)

	if err != nil {
		log.Fatal(err.Error())
	}
}

func run(ms *service.MemStorage, logger *zap.SugaredLogger, f *flags) error {
	fmt.Println("Running server on", f.RunAddr)
	if h, p, err := net.SplitHostPort(f.RunAddr); err == nil {
		if h == "localhost" || h == "" {
			f.RunAddr = ":" + p
		}
	}

	db := initDB(f.Dsn)
	h := handler.NewHandler(ms, logger, db, f.StoreInterval == 0)

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
	if DSN == "" {
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
	defer m.Close()

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("migrate up error: %v", err)
	}
}
