package worker

import (
	"context"
	"time"

	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

var (
	meter               = global.Meter("admin/worker")
	jobLatencyHistogram = observability.Must(meter.Int64Histogram("job_latency", metric.WithUnit("ms")))
)

type Worker struct {
	logger *zap.Logger
	admin  *admin.Service
}

func New(logger *zap.Logger, adm *admin.Service) *Worker {
	return &Worker{
		logger: logger,
		admin:  adm,
	}
}

func (w *Worker) Run(ctx context.Context) error {
	group, ctx := errgroup.WithContext(ctx)
	group.Go(func() error { return w.schedule(ctx, "check_slots", w.checkSlots, 15*time.Minute) })
	group.Go(func() error { return w.schedule(ctx, "delete_expired_tokens", w.deleteExpiredTokens, 15*time.Minute) })
	// NOTE: Add new jobs here

	w.logger.Info("worker started")
	defer w.logger.Info("worker stopped")
	return group.Wait()
}

func (w *Worker) schedule(ctx context.Context, name string, fn func(context.Context) error, every time.Duration) error {
	for {
		start := time.Now()
		w.logger.Info("job started", zap.String("name", name))
		err := fn(ctx)
		jobLatencyHistogram.Record(ctx, time.Since(start).Milliseconds(), metric.WithAttributes(attribute.String("name", name), attribute.Bool("failed", err != nil)))
		if err != nil {
			w.logger.Error("job failed", zap.String("name", name), zap.Error(err), zap.Duration("duration", time.Since(start)))
			return err
		}
		w.logger.Info("job completed", zap.String("name", name), zap.Duration("duration", time.Since(start)))
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(every):
		}
	}
}
