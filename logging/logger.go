package logging

import (
	"context"
	"log/slog"
	"os"

	"github.com/anthriscus/appcli/appcontext"
)

type ContextHandler struct {
	slog.Handler // interface
}

type AppLogger struct {
	Log *slog.Logger
}

func SetupLogger(fi *os.File, options slog.HandlerOptions) *slog.Logger {
	baseHandler := slog.NewJSONHandler(fi, &options)
	// add in the context handler
	customHandler := &ContextHandler{Handler: baseHandler}
	logger := slog.New(customHandler)
	return logger
}

func (h *ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	if traceId, ok := ctx.Value(appcontext.TraceIdKey).(string); ok {
		r.AddAttrs(slog.String(string(appcontext.TraceIdKey), traceId))
	}
	return h.Handler.Handle(ctx, r)
}
