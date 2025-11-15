package main

import (
	"RealTime/internal/config"
	"RealTime/internal/log"
	"RealTime/internal/store"
	"RealTime/internal/wiring"
	"context"
	"database/sql"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

func main() {

	cfg := config.LoadConfig()

	log.InitLogger()
	defer func(Logger *zap.Logger) {
		err := Logger.Sync()
		if err != nil {
			log.Logger.Error("Logger sync failed", zap.Error(err))
		}
	}(log.Logger)
	log.Logger.Info("Starting auth REST server...", zap.String("dbUrl", cfg.DBUrl))
	db, err := store.InitDB(cfg.DBUrl)
	if err != nil {
		log.Logger.Fatal("Failed to initialize database", zap.Error(err))
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Logger.Error("Failed to close database connection", zap.Error(err))
		}
	}(db)
	log.Logger.Info("Successfully connected to database")

	app, err := wiring.BuildRestApi(db, &cfg)

	if err != nil {
		log.Logger.Fatal("Failed to build rest api", zap.Error(err))
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
		log.Logger.Info("Auth REST Server starting", zap.String("port", cfg.APIPort))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Logger.Fatal("Auth server ListenAndServe failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Logger.Info("Auth server shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Logger.Fatal("Auth server shutdown failed", zap.Error(err))
	}

	log.Logger.Info("Auth server exited gracefully")
}
