package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

func InitLogger() {
	var err error

	// Use Production configuration for JSON logs
	config := zap.NewProductionConfig()

	// Set a reasonable time format
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// Optional: Use development config for console output during local testing
	config = zap.NewDevelopmentConfig()

	Logger, err = config.Build()
	if err != nil {
		panic(err)
	}

	// Defer syncing the buffer to ensure all logs are flushed on exit
	zap.ReplaceGlobals(Logger)
	Logger.Info("Structured logger initialized.")
}
