package river

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/jobs"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivermigrate"
	"github.com/riverqueue/river/rivertype"
	"github.com/robfig/cron/v3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	oteltrace "go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

var (
	tracer              = otel.Tracer("github.com/rilldata/rill/admin/jobs/river")
	meter               = otel.Meter("github.com/rilldata/rill/admin/jobs/river")
	jobLatencyHistogram = observability.Must(meter.Int64Histogram("job_latency", metric.WithUnit("ms")))
)

type Client struct {
	dbPool      *pgxpool.Pool
	riverClient *river.Client[pgx.Tx]
}

func New(ctx context.Context, dsn string, adm *admin.Service) (jobs.Client, error) {
	dbPool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}

	tx, err := dbPool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	migrator := rivermigrate.New(riverpgxv5.New(dbPool), nil)

	res, err := migrator.MigrateTx(ctx, tx, rivermigrate.DirectionUp, nil)
	if err != nil {
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	for _, version := range res.Versions {
		adm.Logger.Info("river database migrated", zap.String("direction", string(res.Direction)), zap.Int("version", version.Version))
	}

	workers := river.NewWorkers()
	// NOTE: Register new job workers here
	river.AddWorker(workers, &ValidateDeploymentsWorker{admin: adm})
	river.AddWorker(workers, &ResetAllDeploymentsWorker{admin: adm})

	periodicJobs := []*river.PeriodicJob{
		// NOTE: Add new periodic jobs here
		newPeriodicJob(&ValidateDeploymentsArgs{}, "* */6 * * *", true),
	}

	riverClient, err := river.NewClient(riverpgxv5.New(dbPool), &river.Config{
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: 100},
		},
		Workers:      workers,
		PeriodicJobs: periodicJobs,
		JobTimeout:   time.Hour,
		MaxAttempts:  3,
		ErrorHandler: &ErrorHandler{logger: adm.Logger},
	})
	if err != nil {
		return nil, err
	}

	return &Client{
		dbPool:      dbPool,
		riverClient: riverClient,
	}, nil
}

func (c *Client) Close(ctx context.Context) error {
	err := c.riverClient.Stop(ctx)
	c.dbPool.Close()
	return err
}

func (c *Client) Work(ctx context.Context) error {
	err := c.riverClient.Start(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) CancelJob(ctx context.Context, jobID int64) error {
	_, err := c.riverClient.JobCancel(ctx, jobID)
	if err != nil {
		return err
	}
	return nil
}

// NOTE: Add new job trigger functions here
func (c *Client) ResetAllDeployments(ctx context.Context) (*jobs.InsertResult, error) {
	res, err := c.riverClient.Insert(ctx, ResetAllDeploymentsArgs{}, nil)
	if err != nil {
		return nil, err
	}
	return &jobs.InsertResult{
		ID:        res.Job.ID,
		Duplicate: res.UniqueSkippedAsDuplicate,
	}, nil
}

type ErrorHandler struct {
	logger *zap.Logger
}

func (h *ErrorHandler) HandleError(ctx context.Context, job *rivertype.JobRow, err error) *river.ErrorHandlerResult {
	var args string
	_ = json.Unmarshal(job.EncodedArgs, &args) // ignore parse errors
	h.logger.Error("Job errored", zap.Int64("job_id", job.ID), zap.Int("num_attempt", job.Attempt), zap.String("kind", job.Kind), zap.String("args", args), zap.Error(err))
	return nil
}

func (h *ErrorHandler) HandlePanic(ctx context.Context, job *rivertype.JobRow, panicVal any, trace string) *river.ErrorHandlerResult {
	var args string
	_ = json.Unmarshal(job.EncodedArgs, &args) // ignore parse errors
	h.logger.Error("Job panicked", zap.Int64("job_id", job.ID), zap.String("kind", job.Kind), zap.String("args", args), zap.Any("panic_val", panicVal), zap.String("trace", trace))
	// Set the job to be immediately cancelled
	return &river.ErrorHandlerResult{SetCancelled: true}
}

func newPeriodicJob(jobArgs river.JobArgs, cronExpr string, runOnStart bool) *river.PeriodicJob {
	schedule, err := cron.ParseStandard(cronExpr)
	if err != nil {
		panic(err)
	}

	periodicJob := river.NewPeriodicJob(
		schedule,
		func() (river.JobArgs, *river.InsertOpts) {
			return jobArgs, nil
		},
		&river.PeriodicJobOpts{RunOnStart: runOnStart},
	)

	return periodicJob
}

// Observability work wrapper for the job workers
func work(ctx context.Context, logger *zap.Logger, name string, fn func(context.Context) error) error {
	ctx, span := tracer.Start(ctx, fmt.Sprintf("runJob %s", name), oteltrace.WithAttributes(attribute.String("name", name)))
	defer span.End()

	start := time.Now()
	logger.Info("job started", zap.String("name", name), observability.ZapCtx(ctx))
	err := fn(ctx)
	jobLatencyHistogram.Record(ctx, time.Since(start).Milliseconds(), metric.WithAttributes(attribute.String("name", name), attribute.Bool("failed", err != nil)))
	if err != nil {
		logger.Error("job failed", zap.String("name", name), zap.Error(err), zap.Duration("duration", time.Since(start)), observability.ZapCtx(ctx))
		return err
	}
	logger.Info("job completed", zap.String("name", name), zap.Duration("duration", time.Since(start)), observability.ZapCtx(ctx))
	return nil
}
