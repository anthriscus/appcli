package logging

import (
	"log/slog"
	"os"
)

// slog levels are
// logger.Debug("Debug message")
// logger.Info("Info message")
// logger.Warn("Warning message")
// logger.Error("Error message")

// type logOptions struct {
// 	option slog.HandlerOptions
// }

func ErrorOptions() slog.HandlerOptions {
	appEnv := os.Getenv("APP_ENV")
	var options slog.HandlerOptions
	if appEnv == "development" {
		options = slog.HandlerOptions{AddSource: true, Level: slog.LevelDebug}
	} else {
		options = slog.HandlerOptions{AddSource: false}
	}
	return options
}

func ActivityOptions() slog.HandlerOptions {
	options := slog.HandlerOptions{AddSource: false}
	return options
}

func ServerOptions() slog.HandlerOptions {
	appEnv := os.Getenv("APP_ENV")
	var options slog.HandlerOptions
	if appEnv == "development" {
		options = slog.HandlerOptions{AddSource: true, Level: slog.LevelDebug}
	} else {
		options = slog.HandlerOptions{AddSource: false}
	}
	return options
}
