// purpose of main.go is to start the server after all the modules are implemented completely.
package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Naitik2411/stockit/internal/apiclient"
	"github.com/Naitik2411/stockit/internal/config"
	"github.com/Naitik2411/stockit/internal/database"
	"github.com/Naitik2411/stockit/internal/handler"
	"github.com/Naitik2411/stockit/internal/logger"
	"github.com/Naitik2411/stockit/internal/repository"
	"github.com/Naitik2411/stockit/internal/router"
	"github.com/Naitik2411/stockit/internal/server"
	"github.com/Naitik2411/stockit/internal/service"
	"github.com/Naitik2411/stockit/internal/worker"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}
	loggerService := logger.NewLoggerService(cfg.Observability)
	log := logger.NewLoggerWithService(cfg.Observability, loggerService)
	appLogger := &log

	appLogger.Info().
		Str("env", cfg.Primary.Env).
		Str("port", cfg.Server.Port).
		Msg("Starting stockit backend")

	srv, err := server.New(cfg, appLogger, loggerService)
	if err != nil {
		appLogger.Fatal().Err(err).Msg("Failed to initialize server")
	}

	migrateCtx, migrateCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer migrateCancel()

	if err := database.Migrate(migrateCtx, appLogger, cfg); err != nil {
		appLogger.Fatal().Err(err).Msg("failed to run database migration")
	}

	repos := repository.NewRepositories(srv)
	services, err := service.NewServices(srv, repos)
	if err != nil {
		appLogger.Fatal().Err(err).Msg("failed to initialize services")
	}

	handlers := handler.NewHandlers(srv, services)
	e := router.NewRouter(srv, handlers, services)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	alphaClient := apiclient.NewAlphaVantageClient(cfg.Integration.AlphaVantageKey)
	priceWorker := worker.NewPriceSyncWorker(srv, alphaClient, repos.Stock, []string{"AAPL", "MSFT", "GOOGL", "AMZN", "TSLA"}, cfg.Integration.PriceSyncInterval)

	go priceWorker.Start(ctx)

	srv.SetupHTTPServer(e)

	errCh := make(chan error, 1)
	go func() {
		if err := srv.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()
	// 8. Wait for shutdown signal or startup error
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		appLogger.Fatal().Err(err).Msg("server failed")
	case sig := <-quit:
		appLogger.Info().Str("signal", sig.String()).Msg("shutdown signal received")
		cancel()
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		appLogger.Error().Err(err).Msg("server shutdown failed")
	}
	loggerService.Shutdown()
	appLogger.Info().Msg("server stopped")
}
