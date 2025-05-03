package worker

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/jobs"
	"github.com/rilldata/rill/runtime/pkg/graceful"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/robfig/cron/v3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

const jobTimeout = 60 * time.Minute

var (
	tracer              = otel.Tracer("github.com/rilldata/rill/admin/worker")
	meter               = otel.Meter("github.com/rilldata/rill/admin/worker")
	jobLatencyHistogram = observability.Must(meter.Int64Histogram("job_latency", metric.WithUnit("ms")))
)

type Worker struct {
	logger *zap.Logger
	admin  *admin.Service
	jobs   jobs.Client
}

func New(logger *zap.Logger, adm *admin.Service, jobsClient jobs.Client) *Worker {
	return &Worker{
		logger: logger,
		admin:  adm,
		jobs:   jobsClient,
	}
}

func (w *Worker) Run(ctx context.Context) error {
	// Start jobs client workers
	err := w.jobs.Work(ctx)
	if err != nil {
		panic(err)
	}
	w.logger.Info("jobs client worker started")

	group, ctx := errgroup.WithContext(ctx)
	group.Go(func() error {
		return w.schedule(ctx, "check_provisioners", w.checkProvisioners, 15*time.Minute)
	})
	group.Go(func() error {
		return w.schedule(ctx, "delete_expired_tokens", w.deleteExpiredAuthTokens, 6*time.Hour)
	})
	group.Go(func() error {
		return w.schedule(ctx, "delete_expired_device_auth_codes", w.deleteExpiredDeviceAuthCodes, 6*time.Hour)
	})
	group.Go(func() error {
		return w.schedule(ctx, "delete_expired_auth_codes", w.deleteExpiredAuthCodes, 6*time.Hour)
	})
	group.Go(func() error {
		return w.schedule(ctx, "delete_expired_virtual_files", w.deleteExpiredVirtualFiles, 6*time.Hour)
	})
	group.Go(func() error {
		return w.schedule(ctx, "hibernate_expired_deployments", w.hibernateExpiredDeployments, 15*time.Minute)
	})
	group.Go(func() error {
		return w.scheduleCron(ctx, "run_autoscaler", w.runAutoscaler, w.admin.AutoscalerCron)
	})
	group.Go(func() error {
		return w.schedule(ctx, "delete_unused_assets", w.deleteUnusedAssets, 6*time.Hour)
	})
	group.Go(func() error {
		return w.schedule(ctx, "deployments_health_check", w.deploymentsHealthCheck, 10*time.Minute)
	})

	if w.admin.Biller.GetReportingWorkerCron() != "" {
		group.Go(func() error {
			return w.scheduleCron(ctx, "run_billing_reporter", w.reportUsage, w.admin.Biller.GetReportingWorkerCron())
		})
	}

	// NOTE: Add new scheduled jobs here

	w.logger.Info("worker started")
	defer w.logger.Info("worker stopped")
	return group.Wait()
}

func (w *Worker) RunJob(ctx context.Context, name string) error {
	switch name {
	case "check_provisioners":
		return w.runJob(ctx, name, w.checkProvisioners)
	case "reset_all_deployments":
		_, err := w.jobs.ResetAllDeployments(ctx)
		return err
	// NOTE: Add new ad-hoc jobs here
	default:
		return fmt.Errorf("unknown job: %s", name)
	}
}

func (w *Worker) schedule(ctx context.Context, name string, fn func(context.Context) error, every time.Duration) error {
	for {
		err := w.runJob(ctx, name, fn)
		if err != nil && !errors.Is(err, context.Canceled) {
			w.logger.Error("Failed to run the job", zap.String("job_name", name), zap.Error(err))
		}
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(every):
		}
	}
}

func (w *Worker) scheduleCron(ctx context.Context, name string, fn func(context.Context) error, cronExpr string) error {
	schedule, err := cron.ParseStandard(cronExpr)
	if err != nil {
		return err
	}

	for {
		nextRun := schedule.Next(time.Now())
		waitDuration := time.Until(nextRun)

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(waitDuration):
			err := w.runJob(ctx, name, fn)
			if err != nil && !errors.Is(err, context.Canceled) {
				w.logger.Error("Failed to run the cronjob", zap.String("cronjob_name", name), zap.Error(err))
			}
		}
	}
}

func (w *Worker) runJob(ctx context.Context, name string, fn func(context.Context) error) error {
	ctx, cancel := context.WithTimeout(ctx, jobTimeout)
	defer cancel()

	ctx, span := tracer.Start(ctx, fmt.Sprintf("runJob %s", name), trace.WithAttributes(attribute.String("name", name)))
	defer span.End()

	start := time.Now()
	w.logger.Info("job started", zap.String("name", name), observability.ZapCtx(ctx))
	err := fn(ctx)
	jobLatencyHistogram.Record(ctx, time.Since(start).Milliseconds(), metric.WithAttributes(attribute.String("name", name), attribute.Bool("failed", err != nil)))
	if err != nil {
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
