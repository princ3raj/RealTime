package main

import (
	"RealTime/internal/config"
	"RealTime/internal/logger"
	"RealTime/internal/wiring"
	"context"
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
			logger.Logger.Fatal("Logger sync failed", zap.Error(err))
		}
	}(logger.Logger)

	wsApp, err := wiring.BuildWsServer(&cfg)
	if err != nil {
		return
	}

	go wsApp.ChatHub.Run()
	go wsApp.NotifyHub.Run()

	server := &http.Server{
		Addr:              "0.0.0.0:" + cfg.WSPort,
		Handler:           wsApp.Handler,
		ReadTimeout:       cfg.ReadTimeout,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       60 * time.Second,
	}

	go func(logger *zap.Logger) {
		logger.Info("Server starting", zap.String("port", cfg.WSPort))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("ListenAndServe failed", zap.Error(err), zap.String("port", cfg.WSPort))
		}
	}(logger.Logger)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Logger.Info("Server shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Logger.Fatal("Server shutdown failed:", zap.Error(err))
	}

	logger.Logger.Info("Server exited gracefully")
}
