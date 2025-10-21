package logging

import (
	"context"
	"io"
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

var (
	logger AppLogger
)

func Log() *slog.Logger {
	return logger.Log
}

func (h *ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	if traceId, ok := ctx.Value(appcontext.TraceIdKey).(string); ok {
		r.AddAttrs(slog.String(string(appcontext.TraceIdKey), traceId))
	}
	return h.Handler.Handle(ctx, r)
}

func Setup(w io.Writer, options slog.HandlerOptions) {
	baseHandler := slog.NewJSONHandler(w, &options)
	// add in the context handler
	customHandler := &ContextHandler{Handler: baseHandler}
	logger.Log = slog.New(customHandler)
}

func Default() {
	logger.Log = slog.New(slog.NewTextHandler(os.Stdout, nil))
}
