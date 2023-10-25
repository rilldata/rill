package logutil

import (
	"context"

	"github.com/rilldata/rill/runtime/pkg/logbuffer"
	"go.uber.org/zap/exp/zapslog"
	"golang.org/x/exp/slog"
)

type DuplicatingHandler struct {
	zapHandler *zapslog.Handler
	logs       *logbuffer.Buffer
}

func NewDuplicatingHandler(zapHandler *zapslog.Handler, logs *logbuffer.Buffer) *DuplicatingHandler {
	return &DuplicatingHandler{
		zapHandler: zapHandler,
		logs:       logs,
	}
}

func (d *DuplicatingHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return d.zapHandler.Enabled(ctx, level)
}

func (d *DuplicatingHandler) Handle(ctx context.Context, record slog.Record) error {
	err := d.zapHandler.Handle(ctx, record)
	if err != nil {
		return err
	}
	return d.logs.Add(record)
}

func (d *DuplicatingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandler := *d
	newHandler.zapHandler = d.zapHandler.WithAttrs(attrs).(*zapslog.Handler)
	return &newHandler
}

func (d *DuplicatingHandler) WithGroup(name string) slog.Handler {
	newHandler := *d
	newHandler.zapHandler = d.zapHandler.WithGroup(name).(*zapslog.Handler)
	return &newHandler
}
