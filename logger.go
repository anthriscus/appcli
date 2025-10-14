package main

import (
	"context"
	"log/slog"
	"os"
)

type ContextHandler struct {
	slog.Handler // interface
}

type appLogger struct {
	log *slog.Logger
}

func setupLogger(fi *os.File, options slog.HandlerOptions) *slog.Logger {
	baseHandler := slog.NewJSONHandler(fi, &options)
	// add in the context handler
	customHandler := &ContextHandler{Handler: baseHandler}
	logger := slog.New(customHandler)
	return logger
}

func (h *ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	if traceId, ok := ctx.Value(traceIdKey).(string); ok {
		r.AddAttrs(slog.String(string(traceIdKey), traceId))
	}
	return h.Handler.Handle(ctx, r)
}
