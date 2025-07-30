package river

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
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
	"go.uber.org/zap/exp/zapslog"
)

var (
	tracer              = otel.Tracer("github.com/rilldata/rill/admin/jobs/river")
	meter               = otel.Meter("github.com/rilldata/rill/admin/jobs/river")
	jobLatencyHistogram = observability.Must(meter.Int64Histogram("job_latency", metric.WithUnit("ms")))
)

type Client struct {
	logger      *zap.Logger
	dbPool      *pgxpool.Pool
	riverClient *river.Client[pgx.Tx]
}

func New(ctx context.Context, dsn string, adm *admin.Service) (jobs.Client, error) {
	dbPool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}

	migrator, err := rivermigrate.New(riverpgxv5.New(dbPool), nil)
	if err != nil {
		return nil, err
	}

	res, err := migrator.Migrate(ctx, rivermigrate.DirectionUp, nil)
	if err != nil {
		return nil, err
	}

	for _, version := range res.Versions {
		adm.Logger.Info("river database migrated", zap.String("direction", string(res.Direction)), zap.Int("version", version.Version))
	}

	billingLogger := adm.Logger.Named("billing")

	workers := river.NewWorkers()
	// NOTE: Register new job workers here
	river.AddWorker(workers, &ValidateDeploymentsWorker{admin: adm})
	river.AddWorker(workers, &ResetAllDeploymentsWorker{admin: adm})

	// payment provider event handlers
	river.AddWorker(workers, &PaymentMethodAddedWorker{admin: adm})
	river.AddWorker(workers, &PaymentMethodRemovedWorker{admin: adm})
	river.AddWorker(workers, &CustomerAddressUpdatedWorker{admin: adm})

	// biller event handlers
	river.AddWorker(workers, &PaymentFailedWorker{admin: adm, logger: billingLogger})
	river.AddWorker(workers, &PaymentSuccessWorker{admin: adm, logger: billingLogger})
	river.AddWorker(workers, &PaymentFailedGracePeriodCheckWorker{admin: adm, logger: billingLogger})
	river.AddWorker(workers, &PlanChangedWorker{admin: adm})

	// trial checks worker
	river.AddWorker(workers, &TrialEndingSoonWorker{admin: adm, logger: billingLogger})
	river.AddWorker(workers, &TrialEndCheckWorker{admin: adm, logger: billingLogger})
	river.AddWorker(workers, &TrialGracePeriodCheckWorker{admin: adm, logger: billingLogger})

	// subscription related workers
	river.AddWorker(workers, &SubscriptionCancellationCheckWorker{admin: adm, logger: billingLogger})

	// org related workers
	river.AddWorker(workers, &InitOrgBillingWorker{admin: adm, logger: billingLogger})
	river.AddWorker(workers, &RepairOrgBillingWorker{admin: adm, logger: billingLogger})
	river.AddWorker(workers, &StartTrialWorker{admin: adm, logger: billingLogger})
	river.AddWorker(workers, &DeleteOrgWorker{admin: adm, logger: billingLogger})
	river.AddWorker(workers, &HibernateInactiveOrgsWorker{admin: adm, logger: billingLogger})

	// managed github repo cleanup
	river.AddWorker(workers, &deleteUnusedGithubReposWorker{admin: adm, logger: adm.Logger})

	// token cleanup workers
	river.AddWorker(workers, &DeleteUnusedUserTokenWorker{admin: adm, logger: adm.Logger})
	river.AddWorker(workers, &DeleteUnusedServiceTokenWorker{admin: adm, logger: adm.Logger})

	river.AddWorker(workers, &CheckProvisionersWorker{admin: adm, logger: adm.Logger})
	river.AddWorker(workers, &BillingReporterWorker{admin: adm, logger: billingLogger})
	river.AddWorker(workers, &DeleteExpiredAuthCodesWorker{admin: adm, logger: adm.Logger})
	river.AddWorker(workers, &DeleteExpiredDeviceAuthCodesWorker{admin: adm, logger: adm.Logger})
	river.AddWorker(workers, &DeleteExpiredTokensWorker{admin: adm})
	river.AddWorker(workers, &DeleteExpiredVirtualFilesWorker{admin: adm})
	river.AddWorker(workers, &DeleteUnusedAssetsWorker{admin: adm})
	river.AddWorker(workers, &DeploymentsHealthCheckWorker{admin: adm, logger: adm.Logger})
	river.AddWorker(workers, &HibernateExpiredDeploymentsWorker{admin: adm, logger: adm.Logger})
	river.AddWorker(workers, &RunAutoscalerWorker{admin: adm, logger: adm.Logger})

	periodicJobs := []*river.PeriodicJob{
		// NOTE: Add new periodic jobs here
		newPeriodicJob(&ValidateDeploymentsArgs{}, "*/30 * * * *", true),                  // half-hourly
		newPeriodicJob(&PaymentFailedGracePeriodCheckArgs{}, "0 1 * * *", true),           // daily at 1am UTC
		newPeriodicJob(&TrialEndingSoonArgs{}, "5 1 * * *", true),                         // daily at 1:05am UTC
		newPeriodicJob(&TrialEndCheckArgs{}, "10 1 * * *", true),                          // daily at 1:10am UTC
		newPeriodicJob(&TrialGracePeriodCheckArgs{}, "15 1 * * *", true),                  // daily at 1:15am UTC
		newPeriodicJob(&SubscriptionCancellationCheckArgs{}, "20 1 * * *", true),          // daily at 1:20am UTC
		newPeriodicJob(&DeleteUnusedUserTokenArgs{}, "0 */12 * * *", true),                // every 12 hours
		newPeriodicJob(&DeleteUnusedServiceTokenArgs{}, "0 */12 * * *", true),             // every 12 hours
		newPeriodicJob(&deleteUnusedGithubReposArgs{}, "0 */6 * * *", true),               // every 6 hours
		newPeriodicJob(&HibernateInactiveOrgsArgs{}, "0 7 * * 1", true),                   // Monday at 7:00am UTC
		newPeriodicJob(&CheckProvisionersArgs{}, "0 */15 * * *", true),                    // every 15 minutes
		newPeriodicJob(&BillingReporterArgs{}, adm.Biller.GetReportingWorkerCron(), true), // configured by the admin billing service
		newPeriodicJob(&DeleteExpiredAuthCodesArgs{}, "0 */6 * * *", true),                // every 6 hours
		newPeriodicJob(&DeleteExpiredDeviceAuthCodesArgs{}, "0 */6 * * *", true),          // every 6 hours
		newPeriodicJob(&DeleteExpiredTokensArgs{}, "0 */6 * * *", true),                   // every 6 hours
		newPeriodicJob(&DeleteExpiredVirtualFilesArgs{}, "0 */6 * * *", true),             // every 6 hours
		newPeriodicJob(&DeleteUnusedAssetsArgs{}, "0 */6 * * *", true),                    // every 6 hours
		newPeriodicJob(&DeploymentsHealthCheckArgs{}, "0 */10 * * *", true),               // every 10 minutes
		newPeriodicJob(&HibernateExpiredDeploymentsArgs{}, "0 */15 * * *", true),          // every 15 minutes
		newPeriodicJob(&RunAutoscalerArgs{}, adm.AutoscalerCron, true),                    // configured by the admin autoscaler service
	}

	// Wire our zap logger to a slog logger for the river client
	logger := slog.New(zapslog.NewHandler(adm.Logger.Core(), &zapslog.HandlerOptions{
		AddSource: true,
	}))

	riverClient, err := river.NewClient(riverpgxv5.New(dbPool), &river.Config{
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: 10},
		},
		Workers:      workers,
		PeriodicJobs: periodicJobs,
		Logger:       logger,
		JobTimeout:   time.Hour,
		MaxAttempts:  3, // default retry policy with backoff of attempt^4 seconds
		ErrorHandler: &ErrorHandler{logger: adm.Logger},
	})
	if err != nil {
		return nil, err
	}

	return &Client{
		logger:      adm.Logger,
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

func (c *Client) EnqueueByKind(ctx context.Context, kind string) (*jobs.InsertResult, error) {
	// Jobs that have their own methods with custom logic
	switch kind {
	case "reset_all_deployments":
		return c.ResetAllDeployments(ctx)
	case "check_provisioners":
		return c.CheckProvisioners(ctx)
	case "deployments_health_check":
		return c.DeploymentsHealthCheck(ctx)
	case "hibernate_expired_deployments":
		return c.HibernateExpiredDeployments(ctx)
	case "run_autoscaler":
		return c.RunAutoscaler(ctx)
	case "hibernate_inactive_orgs":
		return c.HibernateInactiveOrgs(ctx)
	case "billing_reporter":
		return c.BillingReporter(ctx)
	case "delete_expired_auth_codes":
		return c.DeleteExpiredAuthCodes(ctx)
	case "delete_expired_device_auth_codes":
		return c.DeleteExpiredDeviceAuthCodes(ctx)
	case "delete_expired_tokens":
		return c.DeleteExpiredTokens(ctx)
	case "delete_expired_virtual_files":
		return c.DeleteExpiredVirtualFiles(ctx)
	case "delete_unused_assets":
		return c.DeleteUnusedAssets(ctx)
	}

	// Jobs that just need simple insertion with empty args
	var jobArgs river.JobArgs
	switch kind {
	case "validate_deployments":
		jobArgs = ValidateDeploymentsArgs{}
	case "payment_failed_grace_period_check":
		jobArgs = PaymentFailedGracePeriodCheckArgs{}
	case "trial_ending_soon":
		jobArgs = TrialEndingSoonArgs{}
	case "trial_end_check":
		jobArgs = TrialEndCheckArgs{}
	case "trial_grace_period_check":
		jobArgs = TrialGracePeriodCheckArgs{}
	case "subscription_cancellation_check":
		jobArgs = SubscriptionCancellationCheckArgs{}
	case "delete_unused_user_tokens":
		jobArgs = DeleteUnusedUserTokenArgs{}
	case "delete_unused_service_tokens":
		jobArgs = DeleteUnusedServiceTokenArgs{}
	case "delete_unused_github_repos":
		jobArgs = deleteUnusedGithubReposArgs{}
	default:
		return nil, fmt.Errorf("unknown job kind: %s", kind)
	}

	res, err := c.riverClient.Insert(ctx, jobArgs, nil)
	if err != nil {
		return nil, err
	}
	return &jobs.InsertResult{ID: res.Job.ID, Duplicate: res.UniqueSkippedAsDuplicate}, nil
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

func (c *Client) PaymentMethodAdded(ctx context.Context, paymentMethodID, paymentCustomerID, paymentType string, eventTime time.Time) (*jobs.InsertResult, error) {
	res, err := c.riverClient.Insert(ctx, PaymentMethodAddedArgs{
		PaymentMethodID:   paymentMethodID,
		PaymentCustomerID: paymentCustomerID,
		PaymentType:       paymentType,
		EventTime:         eventTime,
	}, &river.InsertOpts{
		UniqueOpts: river.UniqueOpts{
			ByArgs: true,
		},
	})
	if err != nil {
		return nil, err
	}

	if res.UniqueSkippedAsDuplicate {
		c.logger.Debug("PaymentMethodAdded job skipped as duplicate", zap.String("payment_method_id", paymentMethodID), zap.String("payment_customer_id", paymentCustomerID), zap.String("payment_type", paymentType), zap.Time("event_time", eventTime))
	}

	return &jobs.InsertResult{
		ID:        res.Job.ID,
		Duplicate: res.UniqueSkippedAsDuplicate,
	}, nil
}

func (c *Client) PaymentMethodRemoved(ctx context.Context, paymentMethodID, paymentCustomerID string, eventTime time.Time) (*jobs.InsertResult, error) {
	res, err := c.riverClient.Insert(ctx, PaymentMethodRemovedArgs{
		PaymentMethodID:   paymentMethodID,
		PaymentCustomerID: paymentCustomerID,
		EventTime:         eventTime,
	}, &river.InsertOpts{
		UniqueOpts: river.UniqueOpts{
			ByArgs: true,
		},
	})
	if err != nil {
		return nil, err
	}

	if res.UniqueSkippedAsDuplicate {
		c.logger.Debug("PaymentMethodRemoved job skipped as duplicate", zap.String("payment_method_id", paymentMethodID), zap.String("payment_customer_id", paymentCustomerID), zap.Time("event_time", eventTime))
	}

	return &jobs.InsertResult{
		ID:        res.Job.ID,
		Duplicate: res.UniqueSkippedAsDuplicate,
	}, nil
}

func (c *Client) CustomerAddressUpdated(ctx context.Context, paymentCustomerID string, eventTime time.Time) (*jobs.InsertResult, error) {
	res, err := c.riverClient.Insert(ctx, CustomerAddressUpdatedArgs{
		PaymentCustomerID: paymentCustomerID,
		EventTime:         eventTime,
	}, &river.InsertOpts{
		UniqueOpts: river.UniqueOpts{
			ByArgs: true,
		},
	})
	if err != nil {
		return nil, err
	}

	if res.UniqueSkippedAsDuplicate {
		c.logger.Debug("CustomerAddressUpdated job skipped as duplicate", zap.String("payment_customer_id", paymentCustomerID), zap.Time("event_time", eventTime))
	}

	return &jobs.InsertResult{
		ID:        res.Job.ID,
		Duplicate: res.UniqueSkippedAsDuplicate,
	}, nil
}

func (c *Client) PaymentFailed(ctx context.Context, billingCustomerID, invoiceID, invoiceNumber, invoiceURL, amount, currency string, dueDate, failedAt time.Time) (*jobs.InsertResult, error) {
	res, err := c.riverClient.Insert(ctx, PaymentFailedArgs{
		BillingCustomerID: billingCustomerID,
		InvoiceID:         invoiceID,
		InvoiceNumber:     invoiceNumber,
		InvoiceURL:        invoiceURL,
		Amount:            amount,
		Currency:          currency,
		DueDate:           dueDate,
		FailedAt:          failedAt,
	}, &river.InsertOpts{
		UniqueOpts: river.UniqueOpts{
			ByArgs: true,
		},
	})
	if err != nil {
		return nil, err
	}

	if res.UniqueSkippedAsDuplicate {
		c.logger.Debug("PaymentFailed job skipped as duplicate", zap.String("billing_customer_id", billingCustomerID), zap.String("invoice_id", invoiceID), zap.String("invoice_number", invoiceNumber), zap.String("invoice_url", invoiceURL), zap.String("amount", amount), zap.String("currency", currency), zap.Time("due_date", dueDate), zap.Time("failed_at", failedAt))
	}

	return &jobs.InsertResult{
		ID:        res.Job.ID,
		Duplicate: res.UniqueSkippedAsDuplicate,
	}, nil
}

func (c *Client) PaymentSuccess(ctx context.Context, billingCustomerID, invoiceID string) (*jobs.InsertResult, error) {
	res, err := c.riverClient.Insert(ctx, PaymentSuccessArgs{
		BillingCustomerID: billingCustomerID,
		InvoiceID:         invoiceID,
	}, &river.InsertOpts{
		UniqueOpts: river.UniqueOpts{
			ByArgs: true,
		},
	})
	if err != nil {
		return nil, err
	}

	if res.UniqueSkippedAsDuplicate {
		c.logger.Debug("PaymentSuccess job skipped as duplicate", zap.String("billing_customer_id", billingCustomerID), zap.String("invoice_id", invoiceID))
	}

	return &jobs.InsertResult{
		ID:        res.Job.ID,
		Duplicate: res.UniqueSkippedAsDuplicate,
	}, nil
}

func (c *Client) InitOrgBilling(ctx context.Context, orgID string) (*jobs.InsertResult, error) {
	res, err := c.riverClient.Insert(ctx, InitOrgBillingArgs{
		OrgID: orgID,
	}, &river.InsertOpts{
		UniqueOpts: river.UniqueOpts{
			ByArgs: true,
		},
	})
	if err != nil {
		return nil, err
	}

	if res.UniqueSkippedAsDuplicate {
		c.logger.Debug("InitOrgBilling job skipped as duplicate", zap.String("org_id", orgID))
	}

	return &jobs.InsertResult{
		ID:        res.Job.ID,
		Duplicate: res.UniqueSkippedAsDuplicate,
	}, nil
}

func (c *Client) RepairOrgBilling(ctx context.Context, orgID string) (*jobs.InsertResult, error) {
	res, err := c.riverClient.Insert(ctx, RepairOrgBillingArgs{
		OrgID: orgID,
	}, &river.InsertOpts{
		UniqueOpts: river.UniqueOpts{
			ByArgs:  true,
			ByState: []rivertype.JobState{rivertype.JobStateAvailable, rivertype.JobStateRunning, rivertype.JobStateRetryable, rivertype.JobStateScheduled}, // to prevent concurrent run but still allow retries
		},
	})
	if err != nil {
		return nil, err
	}

	if res.UniqueSkippedAsDuplicate {
		c.logger.Debug("RepairOrgBilling job skipped as duplicate", zap.String("org_id", orgID))
	}

	return &jobs.InsertResult{
		ID:        res.Job.ID,
		Duplicate: res.UniqueSkippedAsDuplicate,
	}, nil
}

func (c *Client) StartOrgTrial(ctx context.Context, orgID string) (*jobs.InsertResult, error) {
	res, err := c.riverClient.Insert(ctx, StartTrialArgs{
		OrgID: orgID,
	}, &river.InsertOpts{
		UniqueOpts: river.UniqueOpts{
			ByArgs: true,
		},
		MaxAttempts: 5, // override default retries as init org billing job should complete before this if org creation and project deployment were done in single flow
	})
	if err != nil {
		return nil, err
	}

	if res.UniqueSkippedAsDuplicate {
		c.logger.Debug("StartTrial job skipped as duplicate", zap.String("org_id", orgID))
	}

	return &jobs.InsertResult{
		ID:        res.Job.ID,
		Duplicate: res.UniqueSkippedAsDuplicate,
	}, nil
}

func (c *Client) DeleteOrg(ctx context.Context, orgID string) (*jobs.InsertResult, error) {
	res, err := c.riverClient.Insert(ctx, DeleteOrgArgs{
		OrgID: orgID,
	}, &river.InsertOpts{})
	if err != nil {
		return nil, err
	}

	return &jobs.InsertResult{
		ID:        res.Job.ID,
		Duplicate: res.UniqueSkippedAsDuplicate,
	}, nil
}

func (c *Client) PlanChanged(ctx context.Context, billingCustomerID string) (*jobs.InsertResult, error) {
	res, err := c.riverClient.Insert(ctx, PlanChangedArgs{
		BillingCustomerID: billingCustomerID,
	}, &river.InsertOpts{})
	if err != nil {
		return nil, err
	}

	return &jobs.InsertResult{
		ID:        res.Job.ID,
		Duplicate: res.UniqueSkippedAsDuplicate,
	}, nil
}

func (c *Client) HibernateInactiveOrgs(ctx context.Context) (*jobs.InsertResult, error) {
	res, err := c.riverClient.Insert(ctx, HibernateInactiveOrgsArgs{}, nil)
	if err != nil {
		return nil, err
	}
	return &jobs.InsertResult{
		ID:        res.Job.ID,
		Duplicate: res.UniqueSkippedAsDuplicate,
	}, nil
}

func (c *Client) CheckProvisioners(ctx context.Context) (*jobs.InsertResult, error) {
	res, err := c.riverClient.Insert(ctx, CheckProvisionersArgs{}, &river.InsertOpts{
		UniqueOpts: river.UniqueOpts{
			ByArgs: true,
		},
	})
	if err != nil {
		return nil, err
	}

	if res.UniqueSkippedAsDuplicate {
		c.logger.Debug("CheckProvisioners job skipped as duplicate")
	}

	return &jobs.InsertResult{
		ID:        res.Job.ID,
		Duplicate: res.UniqueSkippedAsDuplicate,
	}, nil
}

func (c *Client) BillingReporter(ctx context.Context) (*jobs.InsertResult, error) {
	res, err := c.riverClient.Insert(ctx, BillingReporterArgs{}, &river.InsertOpts{
		UniqueOpts: river.UniqueOpts{
			ByArgs: true,
		},
	})
	if err != nil {
		return nil, err
	}

	if res.UniqueSkippedAsDuplicate {
		c.logger.Debug("BillingReporter job skipped as duplicate")
	}

	return &jobs.InsertResult{
		ID:        res.Job.ID,
		Duplicate: res.UniqueSkippedAsDuplicate,
	}, nil
}

func (c *Client) DeleteExpiredAuthCodes(ctx context.Context) (*jobs.InsertResult, error) {
	res, err := c.riverClient.Insert(ctx, DeleteExpiredAuthCodesArgs{}, &river.InsertOpts{
		UniqueOpts: river.UniqueOpts{
			ByArgs: true,
		},
	})
	if err != nil {
		return nil, err
	}

	if res.UniqueSkippedAsDuplicate {
		c.logger.Debug("DeleteExpiredAuthCodes job skipped as duplicate")
	}

	return &jobs.InsertResult{
		ID:        res.Job.ID,
		Duplicate: res.UniqueSkippedAsDuplicate,
	}, nil
}

func (c *Client) DeleteExpiredDeviceAuthCodes(ctx context.Context) (*jobs.InsertResult, error) {
	res, err := c.riverClient.Insert(ctx, DeleteExpiredDeviceAuthCodesArgs{}, &river.InsertOpts{
		UniqueOpts: river.UniqueOpts{
			ByArgs: true,
		},
	})
	if err != nil {
		return nil, err
	}

	if res.UniqueSkippedAsDuplicate {
		c.logger.Debug("DeleteExpiredDeviceAuthCodes job skipped as duplicate")
	}

	return &jobs.InsertResult{
		ID:        res.Job.ID,
		Duplicate: res.UniqueSkippedAsDuplicate,
	}, nil
}

func (c *Client) DeleteExpiredTokens(ctx context.Context) (*jobs.InsertResult, error) {
	res, err := c.riverClient.Insert(ctx, DeleteExpiredTokensArgs{}, &river.InsertOpts{
		UniqueOpts: river.UniqueOpts{
			ByArgs: true,
		},
	})
	if err != nil {
		return nil, err
	}

	if res.UniqueSkippedAsDuplicate {
		c.logger.Debug("DeleteExpiredTokens job skipped as duplicate")
	}

	return &jobs.InsertResult{
		ID:        res.Job.ID,
		Duplicate: res.UniqueSkippedAsDuplicate,
	}, nil
}

func (c *Client) DeleteExpiredVirtualFiles(ctx context.Context) (*jobs.InsertResult, error) {
	res, err := c.riverClient.Insert(ctx, DeleteExpiredVirtualFilesArgs{}, &river.InsertOpts{
		UniqueOpts: river.UniqueOpts{
			ByArgs: true,
		},
	})
	if err != nil {
		return nil, err
	}

	if res.UniqueSkippedAsDuplicate {
		c.logger.Debug("DeleteExpiredVirtualFiles job skipped as duplicate")
	}

	return &jobs.InsertResult{
		ID:        res.Job.ID,
		Duplicate: res.UniqueSkippedAsDuplicate,
	}, nil
}

func (c *Client) DeleteUnusedAssets(ctx context.Context) (*jobs.InsertResult, error) {
	res, err := c.riverClient.Insert(ctx, DeleteUnusedAssetsArgs{}, &river.InsertOpts{
		UniqueOpts: river.UniqueOpts{
			ByArgs: true,
		},
	})
	if err != nil {
		return nil, err
	}

	if res.UniqueSkippedAsDuplicate {
		c.logger.Debug("DeleteUnusedAssets job skipped as duplicate")
	}

	return &jobs.InsertResult{
		ID:        res.Job.ID,
		Duplicate: res.UniqueSkippedAsDuplicate,
	}, nil
}

func (c *Client) DeploymentsHealthCheck(ctx context.Context) (*jobs.InsertResult, error) {
	res, err := c.riverClient.Insert(ctx, DeploymentsHealthCheckArgs{}, &river.InsertOpts{
		UniqueOpts: river.UniqueOpts{
			ByArgs: true,
		},
	})
	if err != nil {
		return nil, err
	}

	if res.UniqueSkippedAsDuplicate {
		c.logger.Debug("DeploymentsHealthCheck job skipped as duplicate")
	}

	return &jobs.InsertResult{
		ID:        res.Job.ID,
		Duplicate: res.UniqueSkippedAsDuplicate,
	}, nil
}

func (c *Client) HibernateExpiredDeployments(ctx context.Context) (*jobs.InsertResult, error) {
	res, err := c.riverClient.Insert(ctx, HibernateExpiredDeploymentsArgs{}, &river.InsertOpts{
		UniqueOpts: river.UniqueOpts{
			ByArgs: true,
		},
	})
	if err != nil {
		return nil, err
	}

	if res.UniqueSkippedAsDuplicate {
		c.logger.Debug("HibernateExpiredDeployments job skipped as duplicate")
	}

	return &jobs.InsertResult{
		ID:        res.Job.ID,
		Duplicate: res.UniqueSkippedAsDuplicate,
	}, nil
}

func (c *Client) RunAutoscaler(ctx context.Context) (*jobs.InsertResult, error) {
	res, err := c.riverClient.Insert(ctx, RunAutoscalerArgs{}, &river.InsertOpts{
		UniqueOpts: river.UniqueOpts{
			ByArgs: true,
		},
	})
	if err != nil {
		return nil, err
	}

	if res.UniqueSkippedAsDuplicate {
		c.logger.Debug("RunAutoscaler job skipped as duplicate")
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
	if job.Attempt >= job.MaxAttempts {
		var args string
		_ = json.Unmarshal(job.EncodedArgs, &args) // ignore parse errors
		h.logger.Error("Job failed, max attempts reached", zap.Int64("job_id", job.ID), zap.Int("num_attempt", job.Attempt), zap.Int("max_attempts", job.MaxAttempts), zap.String("kind", job.Kind), zap.String("args", args), zap.Error(err))
	}
	return nil
}

func (h *ErrorHandler) HandlePanic(ctx context.Context, job *rivertype.JobRow, panicVal any, trace string) *river.ErrorHandlerResult {
	var args string
	_ = json.Unmarshal(job.EncodedArgs, &args) // ignore parse errors
	h.logger.Error("Job panicked", zap.Int64("job_id", job.ID), zap.String("kind", job.Kind), zap.String("args", args), zap.Any("panic_val", panicVal), zap.String("trace", trace))
	// Set the job to be immediately cancelled
	return &river.ErrorHandlerResult{SetCancelled: true}
}

func newPeriodicJob(jobArgs river.JobArgs, cronExpr string, runOnStart bool) *river.PeriodicJob { // nolint:unparam // runOnStart may be used in the future
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
