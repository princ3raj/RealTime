package main

import (
	"RealTime/internal/config"
	"RealTime/internal/logger"
	"RealTime/internal/repository/postgres"
	"RealTime/internal/wiring"
	"context"
	"database/sql"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

func main() {

	cfg := config.LoadConfig()

	logger.InitLogger()
	defer func(Logger *zap.Logger) {
		err := Logger.Sync()
		if err != nil {
			logger.Logger.Error("Logger sync failed", zap.Error(err))
		}
	}(logger.Logger)
	logger.Logger.Info("Starting auth REST realtime...", zap.String("dbUrl", cfg.DBUrl))
	db, err := postgres.InitDB(cfg.DBUrl)
	if err != nil {
		logger.Logger.Fatal("Failed to initialize database", zap.Error(err))
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.Logger.Error("Failed to close database connection", zap.Error(err))
		}
	}(db)
	logger.Logger.Info("Successfully connected to database")

	app, err := wiring.BuildRestApi(db, &cfg)

	if err != nil {
		logger.Logger.Fatal("Failed to build http transport", zap.Error(err))
	}

	server := &http.Server{
		Addr:              "0.0.0.0:" + cfg.APIPort,
		Handler:           app,
		ReadTimeout:       cfg.ReadTimeout,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		logger.Logger.Info("Auth REST Server starting", zap.String("port", cfg.APIPort))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Logger.Fatal("Auth realtime ListenAndServe failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Logger.Info("Auth realtime shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Logger.Fatal("Auth realtime shutdown failed", zap.Error(err))
	}

	logger.Logger.Info("Auth realtime exited gracefully")
}
