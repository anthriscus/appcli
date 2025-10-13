package main

import (
	"log/slog"
	"os"
)

type appLogger struct {
	log *slog.Logger
}

func setupLogger(fi *os.File, options slog.HandlerOptions) *slog.Logger {
	logger := slog.New(slog.NewJSONHandler(fi, &options))
	return logger
}
