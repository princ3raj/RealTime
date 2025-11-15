package main

import (
	"RealTime/internal/config"
	"RealTime/internal/log"
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

	log.InitLogger()
	defer func(Logger *zap.Logger) {
		err := Logger.Sync()
		if err != nil {
			log.Logger.Fatal("Logger sync failed", zap.Error(err))
		}
	}(log.Logger)

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
		log.Logger.Info("Server starting", zap.String("port", cfg.WSPort))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Logger.Fatal("ListenAndServe failed", zap.Error(err), zap.String("port", cfg.WSPort))
		}
	}(log.Logger)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Logger.Info("Server shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Logger.Fatal("Server shutdown failed:", zap.Error(err))
	}

	log.Logger.Info("Server exited gracefully")
}
