package utils

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

func InitLogger() {
	var config zap.Config

	env := os.Getenv("GIN_MODE")
	if env == "release" {
		config = zap.NewProductionConfig()
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	} else {
		config = zap.NewDevelopmentConfig()
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
		config.Development = true
	}

	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.MessageKey = "message"
	config.EncoderConfig.LevelKey = "level"
	config.EncoderConfig.CallerKey = "caller"
	config.EncoderConfig.StacktraceKey = "stacktrace"

	var err error
	Logger, err = config.Build(zap.AddCallerSkip(1))
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
}

func Info(msg string, fields ...zap.Field) {
	if Logger == nil {
		InitLogger()
	}
	Logger.Info(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	if Logger == nil {
		InitLogger()
	}
	Logger.Error(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	if Logger == nil {
		InitLogger()
	}
	Logger.Warn(msg, fields...)
}

func Debug(msg string, fields ...zap.Field) {
	if Logger == nil {
		InitLogger()
	}
	Logger.Debug(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	if Logger == nil {
		InitLogger()
	}
	Logger.Fatal(msg, fields...)
}

func Sync() {
	if Logger != nil {
		Logger.Sync()
	}
}