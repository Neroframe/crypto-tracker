// @title           Crypto Tracker API
// @version         1.0
// @description     API for tracking cryptocurrency prices
// @host      localhost:8080
// @BasePath  /
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/Neroframe/crypto-tracker/config"
	"github.com/Neroframe/crypto-tracker/internal/app"
	httpdelivery "github.com/Neroframe/crypto-tracker/internal/delivery/http"
	ext "github.com/Neroframe/crypto-tracker/internal/infra/external"
	pgrepo "github.com/Neroframe/crypto-tracker/internal/infra/postgres"
	"github.com/Neroframe/crypto-tracker/pkg/logger"

	_ "github.com/Neroframe/crypto-tracker/docs"
)

func main() {
	cfg, err := config.Load("config/dev.yaml")
	if err != nil {
		log.Fatal(err)
	}

	log := logger.New(logger.Config(cfg.Log))
	log.Info("config loaded", "version", cfg.Version)

	// Connect to Postgres via GORM
	pgCfg := cfg.Postgres
	log.Info("initializing Postgres connection", "host", pgCfg.Host, "port", pgCfg.Port)

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		pgCfg.User,
		pgCfg.Password,
		pgCfg.Host,
		pgCfg.Port,
		pgCfg.DBName,
	)
	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect to Postgres", "error", err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		log.Fatal("failed to get sql.DB", "error", err)
	}

	log.Info("sql.DB instance acquired")

	sqlDB.SetMaxOpenConns(pgCfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(pgCfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(pgCfg.ConnMaxLifetime)

	log.Info("Postgres connection initialized successfully")

	// Init CoinPaprika client
	cpClient, err := ext.NewCoinPaprikaClient(
		nil,
		cfg.External.CoinPaprikaURL,
		cfg.External.RateLimit,
		log,
	)
	if err != nil {
		log.Fatal("failed to init CoinPaprika client", "error", err)
	}

	// Wire up
	validate := validator.New() // init validator
	repo := pgrepo.NewGormRepo(gormDB, log)
	svc := app.NewCryptoService(repo, cpClient, log)
	handler := httpdelivery.NewCryptoHandler(validate, log, svc)

	// Build Gin router
	router := httpdelivery.SetupRouter(handler)

	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port),
		Handler:      router,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Start worker
	go startScheduler(ctx, svc, cfg.External.FetchInterval, log)

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.ListenAndServe()
	}()

	// Block until either error or shutdown signal
	select {
	case <-ctx.Done():
		log.Info("shutdown signal received via context")
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			log.Error("server error", "error", err)
		} else {
			log.Info("server closed normally")
		}
	}

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("HTTP server shutdown error", "error", err)
	}

	if err := sqlDB.Close(); err != nil {
		log.Error("DB close error", "error", err)
	}
}

// Runs a price fetch with configured interval
func startScheduler(
	ctx context.Context,
	svc app.CryptoService,
	interval time.Duration,
	log *logger.Logger,
) {
	for {
		// Check for shutdown before starting
		select {
		case <-ctx.Done():
			log.Info("scheduler: received shutdown signal, stopping")
			return
		default:
		}

		// Run the job with backoff
		log.Debug("scheduler: starting FetchAndStorePrices")
		const maxAttempts = 3
		for attempt := 1; attempt <= maxAttempts; attempt++ {
			if err := svc.FetchAndStorePrices(ctx); err != nil {
				log.Error("scheduler fetch error", "attempt", attempt, "error", err)
				if attempt == maxAttempts {
					log.Warn("scheduler: max retries reached, giving up this run")
					break
				}
				// exponential backoff: 2s, 4s, 6s…
				select {
				case <-ctx.Done():
					log.Info("scheduler: shutdown during backoff, stopping")
					return
				case <-time.After(time.Duration(attempt*2) * time.Second):
				}
				continue
			}
			log.Info("scheduler: FetchAndStorePrices succeeded")
			break
		}

		// Wait interval before next run or exit if shutting down
		select {
		case <-ctx.Done():
			log.Info("scheduler: received shutdown signal during wait, stopping")
			return
		case <-time.After(interval):
		}
	}
}
