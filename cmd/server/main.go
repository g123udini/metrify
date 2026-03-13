package main

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"database/sql"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"log"
	"metrify/internal/audit"
	"metrify/internal/handler"
	"metrify/internal/pprof"
	"metrify/internal/proto"
	"metrify/internal/router"
	"metrify/internal/rpc"
	"metrify/internal/service"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var (
	BuildVersion = "N/A"
	BuildTime    = "N/A"
	BuildCommit  = "N/A"
)

// @title           Metrify API
// @version         1.0
// @description     Metrics collection service API.
// @BasePath        /
// @schemes         http
func main() {
	fmt.Printf("version=%s, time=%s\n, commit=%s\n", BuildVersion, BuildTime, BuildCommit)

	f := parseFlags()

	db := initDB(f.Dsn)
	ms := service.NewMemStorage(f.FileStorePath, db)
	logger := service.NewLogger()

	rootCtx := context.Background()
	ctx, cancel := signal.NotifyContext(rootCtx, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()

	if f.Restore {
		if err := ms.ReadFromFile(f.FileStorePath); err != nil {
			log.Printf("could not read from file store: %v", err)
		}
	}

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return runMetricDumper(ctx, ms, f)
	})

	g.Go(func() error {
		pprof.ListenSignals(ctx, logger, f.CPUProfileFile, f.CPUProfileDuration, f.MemProfileFile)
		return nil
	})

	g.Go(func() error {
		if f.Protocol == "http" {
			return runHTTPServer(ctx, ms, db, logger, f)
		} else {
			interceptor, err := rpc.NewTrustedSubnetInterceptor(f.TrustedSubnet)
			if err != nil {
				logger.Error(err)
			}

			grpcServer := grpc.NewServer(
				grpc.UnaryInterceptor(interceptor),
			)

			proto.RegisterMetricsServer(grpcServer, rpc.NewMetricsService(ms))

			return err
		}
	})

	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatal(err)
	}
}

func runHTTPServer(ctx context.Context, ms *service.MemStorage, db *sql.DB, logger *zap.SugaredLogger, f *flags) error {
	f.RunAddr = normalizeAddr(f.RunAddr)
	fmt.Println("Running server on", f.RunAddr)

	auditPublisher := initAuditPublisher(f)

	var privKey *rsa.PrivateKey
	if f.CryptoKey != "" {
		key, err := readPrivateKeyFromFile(f.CryptoKey)
		if err != nil {
			logger.Fatal(err)
		}
		privKey = key
	}

	h := handler.NewHandler(
		ms,
		logger,
		db,
		auditPublisher,
		f.StoreInterval == 0,
		f.Key,
		privKey,
		f.TrustedSubnet,
	)

	srv := &http.Server{
		Addr:    f.RunAddr,
		Handler: router.Metric(h),
	}

	shutdownErr := make(chan error, 1)
	go func() {
		<-ctx.Done()

		ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		shutdownErr <- srv.Shutdown(ctxTimeout)
	}()

	err := srv.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	select {
	case err := <-shutdownErr:
		return err
	default:
		return nil
	}
}

func runMetricDumper(ctx context.Context, ms *service.MemStorage, f *flags) error {
	if f.StoreInterval <= 0 {
		return nil
	}

	ticker := time.NewTicker(time.Duration(f.StoreInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			_ = ms.FlushToFile()
			return ctx.Err()
		case <-ticker.C:
			if err := ms.FlushToFile(); err != nil {
				log.Printf("cannot save metrics: %v", err)
			}
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

func readPrivateKeyFromFile(filepath string) (*rsa.PrivateKey, error) {
	keyBytes, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(keyBytes)
	if block == nil {
		return nil, fmt.Errorf("can`t decode PEM")
	}

	return x509.ParsePKCS1PrivateKey(block.Bytes)
}
