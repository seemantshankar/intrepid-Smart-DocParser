package logger

import (
	"contract-analysis-service/configs"
	"go.uber.org/zap"
)

// NewLogger creates a new zap logger based on config
func NewLogger(cfg configs.LoggerConfig) *zap.Logger {
	level := zap.InfoLevel
	switch cfg.Level {
	case "debug":
		level = zap.DebugLevel
	case "warn":
		level = zap.WarnLevel
	case "error":
		level = zap.ErrorLevel
	}
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(level)
	logger, err := config.Build()
	if err != nil {
		panic(err)
	}
	return logger
}
