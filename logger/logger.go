package logger

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger initialize our golang logger
func Start(logLevel string) {

	var config zap.Config = zap.NewDevelopmentConfig()
	if strings.EqualFold(os.Getenv("ENVIRONMENT"), "PRODUCTION") {
		config = zap.NewProductionConfig()
	}

	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	switch strings.ToLower(logLevel) {
	case "debug":
		config.Level.SetLevel(zap.DebugLevel)
	case "info":
		config.Level.SetLevel(zap.InfoLevel)
	case "warning":
		config.Level.SetLevel(zap.WarnLevel)
	case "error":
		config.Level.SetLevel(zap.ErrorLevel)
	default:
		config.Level.SetLevel(zap.InfoLevel)
	}

	logger, err := config.Build()
	if err != nil {
		zap.L().Panic("Failed to init zap global logger, no zap log will be shown till zap is properly initialized", zap.Error(err))
	}

	zap.ReplaceGlobals(logger)
}
