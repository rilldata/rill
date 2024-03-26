package worker

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/runtime/pkg/graceful"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

var (
	tracer              = otel.Tracer("github.com/rilldata/rill/admin/worker")
	meter               = otel.Meter("github.com/rilldata/rill/admin/worker")
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
	group.Go(func() error {
		return w.schedule(ctx, "check_provisioner_capacity", w.checkProvisionerCapacity, 15*time.Minute)
	})
	group.Go(func() error {
		return w.schedule(ctx, "delete_expired_tokens", w.deleteExpiredAuthTokens, 6*time.Hour)
	})
	group.Go(func() error {
		return w.schedule(ctx, "delete_expired_device_auth_codes", w.deleteExpiredDeviceAuthCodes, 6*time.Hour)
	})
	group.Go(func() error {
		return w.schedule(ctx, "delete_expired_virtual_files", w.deleteExpiredVirtualFiles, 6*time.Hour)
	})
	group.Go(func() error {
		return w.schedule(ctx, "hibernate_expired_deployments", w.hibernateExpiredDeployments, 15*time.Minute)
	})
	group.Go(func() error {
		return w.schedule(ctx, "upgrade_latest_version_projects", w.upgradeLatestVersionProjects, 6*time.Hour)
	})
	group.Go(func() error {
		return w.schedule(ctx, "run_autoscaler", w.runAutoscaler, 6*time.Hour)
	})

	// NOTE: Add new scheduled jobs here

	w.logger.Info("worker started")
	defer w.logger.Info("worker stopped")
	return group.Wait()
}

func (w *Worker) RunJob(ctx context.Context, name string) error {
	switch name {
	case "check_provisioner_capacity":
		return w.runJob(ctx, name, w.checkProvisionerCapacity)
	case "reset_all_deployments":
		return w.runJob(ctx, name, w.resetAllDeployments)
	case "upgrade_latest_version_projects":
		return w.runJob(ctx, name, w.upgradeLatestVersionProjects)
	// NOTE: Add new ad-hoc jobs here
	default:
		return fmt.Errorf("unknown job: %s", name)
	}
}

func (w *Worker) schedule(ctx context.Context, name string, fn func(context.Context) error, every time.Duration) error {
	for {
		err := w.runJob(ctx, name, fn)
		if err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(every):
		}
	}
}

func (w *Worker) runJob(ctx context.Context, name string, fn func(context.Context) error) error {
	ctx, span := tracer.Start(ctx, fmt.Sprintf("runJob %s", name), trace.WithAttributes(attribute.String("name", name)))
	defer span.End()

	start := time.Now()
	w.logger.Info("job started", zap.String("name", name), observability.ZapCtx(ctx))
	err := fn(ctx)
	jobLatencyHistogram.Record(ctx, time.Since(start).Milliseconds(), metric.WithAttributes(attribute.String("name", name), attribute.Bool("failed", err != nil)))
	if err != nil {
		w.logger.Error("job failed", zap.String("name", name), zap.Error(err), zap.Duration("duration", time.Since(start)), observability.ZapCtx(ctx))
		return err
	}
	w.logger.Info("job completed", zap.String("name", name), zap.Duration("duration", time.Since(start)), observability.ZapCtx(ctx))
	return nil
}

type pingHandler struct{}

func (h *pingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("pong"))
	if err != nil {
		panic(err)
	}
}

// StartPingServer starts a http server that returns 200 OK on /ping
func StartPingServer(ctx context.Context, port int) error {
	httpMux := http.NewServeMux()
	httpMux.Handle("/ping", &pingHandler{})
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: httpMux,
	}

	return graceful.ServeHTTP(ctx, srv, graceful.ServeOptions{
		Port: port,
	})
}
