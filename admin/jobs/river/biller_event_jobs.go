package river

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/billing"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/runtime/pkg/email"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
)

type PaymentFailedArgs struct {
	BillingCustomerID string
	InvoiceID         string
	InvoiceNumber     string
	InvoiceURL        string
	Amount            string
	Currency          string
	DueDate           time.Time
	FailedAt          time.Time
}

func (PaymentFailedArgs) Kind() string { return "payment_failed" }

type PaymentFailedWorker struct {
	river.WorkerDefaults[PaymentFailedArgs]
	admin  *admin.Service
	logger *zap.Logger
}

func (w *PaymentFailedWorker) Work(ctx context.Context, job *river.Job[PaymentFailedArgs]) error {
	org, err := w.admin.DB.FindOrganizationForBillingCustomerID(ctx, job.Args.BillingCustomerID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			// org got deleted, ignore
			return nil
		}
		return fmt.Errorf("failed to find organization of billing customer id %q: %w", job.Args.BillingCustomerID, err)
	}

	be, err := w.admin.DB.FindBillingIssueByTypeForOrg(ctx, org.ID, database.BillingIssueTypePaymentFailed)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			return fmt.Errorf("failed to find billing errors: %w", err)
		}
	}
	var metadata *database.BillingIssueMetadataPaymentFailed
	if be != nil {
		metadata = be.Metadata.(*database.BillingIssueMetadataPaymentFailed)
	} else {
		metadata = &database.BillingIssueMetadataPaymentFailed{
			Invoices: make(map[string]*database.BillingIssueMetadataPaymentFailedMeta),
		}
	}

	gracePeriodEndDate := job.Args.DueDate.AddDate(0, 0, database.BillingGracePeriodDays)
	metadata.Invoices[job.Args.InvoiceID] = &database.BillingIssueMetadataPaymentFailedMeta{
		ID:                 job.Args.InvoiceID,
		Number:             job.Args.InvoiceNumber,
		URL:                job.Args.InvoiceURL,
		Amount:             job.Args.Amount,
		Currency:           job.Args.Currency,
		DueDate:            job.Args.DueDate,
		FailedOn:           job.Args.FailedAt,
		GracePeriodEndDate: gracePeriodEndDate,
	}

	// insert billing error
	_, err = w.admin.DB.UpsertBillingIssue(ctx, &database.UpsertBillingIssueOptions{
		OrgID:     org.ID,
		Type:      database.BillingIssueTypePaymentFailed,
		Metadata:  &database.BillingIssueMetadataPaymentFailed{Invoices: metadata.Invoices},
		EventTime: job.Args.FailedAt,
	})
	if err != nil {
		return fmt.Errorf("failed to add billing error: %w", err)
	}

	w.logger.Warn("invoice payment failed", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.String("amount", job.Args.Amount), zap.Time("due_date", job.Args.DueDate), zap.String("invoice_id", job.Args.InvoiceID), zap.String("invoice_url", job.Args.InvoiceURL))

	err = w.admin.Email.SendInvoicePaymentFailed(&email.InvoicePaymentFailed{
		ToEmail:            org.BillingEmail,
		ToName:             org.Name,
		OrgName:            org.Name,
		Currency:           job.Args.Currency,
		Amount:             job.Args.Amount,
		PaymentURL:         w.admin.URLs.PaymentPortal(org.Name),
		GracePeriodEndDate: gracePeriodEndDate,
	})
	if err != nil {
		return fmt.Errorf("failed to send invoice payment failed email for org %q: %w", org.Name, err)
	}

	return nil
}

type PaymentSuccessArgs struct {
	BillingCustomerID string
	InvoiceID         string
}

func (PaymentSuccessArgs) Kind() string { return "payment_success" }

type PaymentSuccessWorker struct {
	river.WorkerDefaults[PaymentSuccessArgs]
	admin  *admin.Service
	logger *zap.Logger
}

func (w *PaymentSuccessWorker) Work(ctx context.Context, job *river.Job[PaymentSuccessArgs]) error {
	org, err := w.admin.DB.FindOrganizationForBillingCustomerID(ctx, job.Args.BillingCustomerID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			// org got deleted, ignore
			return nil
		}
		return fmt.Errorf("failed to find organization of billing customer id %q: %w", job.Args.BillingCustomerID, err)
	}

	// check for existing billing error and delete it
	be, err := w.admin.DB.FindBillingIssueByTypeForOrg(ctx, org.ID, database.BillingIssueTypePaymentFailed)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			// no billing error, ignore
			return nil
		}
		return fmt.Errorf("failed to find billing errors: %w", err)
	}

	failedInvoices := be.Metadata.(*database.BillingIssueMetadataPaymentFailed).Invoices
	_, ok := failedInvoices[job.Args.InvoiceID]
	if !ok {
		// invoice not found in the failed invoices, do nothing
		return nil
	}

	delete(failedInvoices, job.Args.InvoiceID)
	w.logger.Info("invoice payment success for a failed invoice", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.String("invoice_id", job.Args.InvoiceID))

	// if no more failed invoices, delete the billing error
	if len(failedInvoices) == 0 {
		err = w.admin.DB.DeleteBillingIssue(ctx, be.ID)
		if err != nil {
			return fmt.Errorf("failed to delete billing error: %w", err)
		}
	} else {
		// update the metadata
		_, err = w.admin.DB.UpsertBillingIssue(ctx, &database.UpsertBillingIssueOptions{
			OrgID:     org.ID,
			Type:      database.BillingIssueTypePaymentFailed,
			Metadata:  &database.BillingIssueMetadataPaymentFailed{Invoices: failedInvoices},
			EventTime: be.EventTime,
		})
		if err != nil {
			return fmt.Errorf("failed to update billing error: %w", err)
		}
	}

	// send email
	err = w.admin.Email.SendInvoicePaymentSuccess(&email.InvoicePaymentSuccess{
		ToEmail:        org.BillingEmail,
		ToName:         org.Name,
		OrgName:        org.Name,
		PaymentDate:    time.Now(),
		BillingPageURL: w.admin.URLs.Billing(org.Name, false),
	})
	if err != nil {
		// ignore email sending error
		w.logger.Error("failed to send invoice payment success email", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.String("invoice_id", job.Args.InvoiceID), zap.String("billing_email", org.BillingEmail), zap.Error(err))
	}

	return nil
}

type PaymentFailedGracePeriodCheckArgs struct{}

func (PaymentFailedGracePeriodCheckArgs) Kind() string {
	return "payment_failed_grace_period_check"
}

type PaymentFailedGracePeriodCheckWorker struct {
	river.WorkerDefaults[PaymentFailedGracePeriodCheckArgs]
	admin  *admin.Service
	logger *zap.Logger
}

func (w *PaymentFailedGracePeriodCheckWorker) Work(ctx context.Context, job *river.Job[PaymentFailedGracePeriodCheckArgs]) error {
	return work(ctx, w.admin.Logger, job.Kind, w.paymentFailedGracePeriodCheck)
}

func (w *PaymentFailedGracePeriodCheckWorker) paymentFailedGracePeriodCheck(ctx context.Context) error {
	failures, err := w.admin.DB.FindBillingIssueByTypeAndOverdueProcessed(ctx, database.BillingIssueTypePaymentFailed, false)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			// no orgs have this billing error
			return nil
		}
		return fmt.Errorf("failed to find organization with billing issue: %w", err)
	}

	// failures are per org
	for _, f := range failures {
		overdue, err := w.checkFailedInvoicesForOrg(ctx, f)
		if err != nil {
			w.logger.Error("failed to check failed invoices for org", zap.String("org_id", f.OrgID), zap.Error(err))
			continue // continue to next org
		}

		if !overdue {
			continue // continue to next org
		}

		// hibernate projects
		limit := 10
		afterProjectName := ""
		for {
			projs, err := w.admin.DB.FindProjectsForOrganization(ctx, f.OrgID, afterProjectName, limit)
			if err != nil {
				return err
			}

			for _, proj := range projs {
				_, err = w.admin.HibernateProject(ctx, proj)
				if err != nil {
					return fmt.Errorf("failed to hibernate project %q: %w", proj.Name, err)
				}
				afterProjectName = proj.Name
			}

			if len(projs) < limit {
				break
			}
		}
		org, err := w.admin.DB.FindOrganization(ctx, f.OrgID)
		if err != nil {
			return fmt.Errorf("failed to find organization: %w", err)
		}

		w.logger.Warn("projects hibernated due to unpaid invoice", zap.String("org_id", org.ID), zap.String("org_name", org.Name))

		// send email
		err = w.admin.Email.SendInvoiceUnpaid(&email.InvoiceUnpaid{
			ToEmail:    org.BillingEmail,
			ToName:     org.Name,
			OrgName:    org.Name,
			PaymentURL: w.admin.URLs.PaymentPortal(org.Name),
		})
		if err != nil {
			return fmt.Errorf("failed to send project hibernated due to payment overdue email for org %q: %w", org.Name, err)
		}

		// mark the billing issue as processed
		err = w.admin.DB.UpdateBillingIssueOverdueAsProcessed(ctx, f.ID)
		if err != nil {
			return fmt.Errorf("failed to mark billing issue as processed: %w", err)
		}
	}
	return nil
}

// reconciles failed payments for the org and returns true if any is overdue
func (w *PaymentFailedGracePeriodCheckWorker) checkFailedInvoicesForOrg(ctx context.Context, orgPaymentFailures *database.BillingIssue) (bool, error) {
	hasOverdue := false
	for invoiceID, failedInvoice := range orgPaymentFailures.Metadata.(*database.BillingIssueMetadataPaymentFailed).Invoices {
		if time.Now().UTC().Before(failedInvoice.GracePeriodEndDate.AddDate(0, 0, 1)) {
			continue
		}

		// just to be very sure, check if the invoice is still unpaid directly from the biller
		invoice, err := w.admin.Biller.GetInvoice(ctx, invoiceID)
		if err != nil {
			return false, fmt.Errorf("failed to get invoice %q: %w", invoiceID, err)
		}

		// if invoice is valid and not paid
		if w.admin.Biller.IsInvoiceValid(ctx, invoice) && !w.admin.Biller.IsInvoicePaid(ctx, invoice) {
			hasOverdue = true
			continue
		}

		w.logger.Warn("invoice was already paid or invalid but billing issue was not cleared", zap.String("org_id", orgPaymentFailures.OrgID), zap.String("invoice_id", invoiceID), zap.String("invoice_status", invoice.Status))

		// clearing the billing error for this invoice
		delete(orgPaymentFailures.Metadata.(*database.BillingIssueMetadataPaymentFailed).Invoices, invoiceID)

		// if no more failed invoices, delete the billing error
		if len(orgPaymentFailures.Metadata.(*database.BillingIssueMetadataPaymentFailed).Invoices) == 0 {
			err = w.admin.DB.DeleteBillingIssue(ctx, orgPaymentFailures.ID)
			if err != nil {
				return false, fmt.Errorf("failed to delete billing error: %w", err)
			}
		} else {
			// update the metadata
			_, err = w.admin.DB.UpsertBillingIssue(ctx, &database.UpsertBillingIssueOptions{
				OrgID:     orgPaymentFailures.OrgID,
				Type:      database.BillingIssueTypePaymentFailed,
				Metadata:  orgPaymentFailures.Metadata,
				EventTime: orgPaymentFailures.EventTime,
			})
			if err != nil {
				return false, fmt.Errorf("failed to update billing error: %w", err)
			}
		}
	}
	return hasOverdue, nil
}

type PlanChangedArgs struct {
	BillingCustomerID string
}

func (PlanChangedArgs) Kind() string { return "plan_changed" }

type PlanChangedWorker struct {
	river.WorkerDefaults[PlanChangedArgs]
	admin *admin.Service
}

func (w *PlanChangedWorker) Work(ctx context.Context, job *river.Job[PlanChangedArgs]) error {
	org, err := w.admin.DB.FindOrganizationForBillingCustomerID(ctx, job.Args.BillingCustomerID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			// org got deleted, ignore
			return nil
		}
		return fmt.Errorf("failed to find organization of billing customer id %q: %w", job.Args.BillingCustomerID, err)
	}

	orgName := org.Name
	// something related to plan changed, just fetch the latest plan from the biller
	sub, err := w.admin.Biller.GetActiveSubscription(ctx, org.BillingCustomerID)
	if err != nil && !errors.Is(err, billing.ErrNotFound) {
		return fmt.Errorf("failed to get subscriptions for org %q: %w", orgName, err)
	}

	var planDisplayName string
	var planName string
	if sub == nil {
		planDisplayName = ""
		planName = ""
	} else {
		planDisplayName = sub.Plan.DisplayName
		planName = sub.Plan.Name
	}

	if org.BillingPlanName == nil || *org.BillingPlanName != planName {
		_, err = w.admin.DB.UpdateOrganization(ctx, org.ID, &database.UpdateOrganizationOptions{
			Name:                                org.Name,
			DisplayName:                         org.DisplayName,
			Description:                         org.Description,
			LogoAssetID:                         org.LogoAssetID,
			FaviconAssetID:                      org.FaviconAssetID,
			ThumbnailAssetID:                    org.ThumbnailAssetID,
			CustomDomain:                        org.CustomDomain,
			DefaultProjectRoleID:                org.DefaultProjectRoleID,
			QuotaProjects:                       org.QuotaProjects,
			QuotaDeployments:                    org.QuotaDeployments,
			QuotaSlotsTotal:                     org.QuotaSlotsTotal,
			QuotaSlotsPerDeployment:             org.QuotaSlotsPerDeployment,
			QuotaOutstandingInvites:             org.QuotaOutstandingInvites,
			QuotaStorageLimitBytesPerDeployment: org.QuotaStorageLimitBytesPerDeployment,
			BillingCustomerID:                   org.BillingCustomerID,
			PaymentCustomerID:                   org.PaymentCustomerID,
			BillingEmail:                        org.BillingEmail,
			BillingPlanName:                     &planName,
			BillingPlanDisplayName:              &planDisplayName,
			CreatedByUserID:                     org.CreatedByUserID,
		})
		if err != nil {
			return fmt.Errorf("failed to update plan cache for org %q: %w", orgName, err)
		}
	}

	return nil
}

type BillingReporterArgs struct{}

func (BillingReporterArgs) Kind() string { return "billing_reporter" }

type BillingReporterWorker struct {
	river.WorkerDefaults[BillingReporterArgs]
	admin  *admin.Service
	logger *zap.Logger
}

func NewBillingReporterWorker(admin *admin.Service, logger *zap.Logger) *BillingReporterWorker {
	return &BillingReporterWorker{
		admin:  admin,
		logger: logger,
	}
}

func (w *BillingReporterWorker) Work(ctx context.Context, job *river.Job[BillingReporterArgs]) error {
	// Get reporting granularity
	var granularity time.Duration
	var sqlGrainIdentifier string
	var gracePeriod time.Duration
	switch w.admin.Biller.GetReportingGranularity() {
	case billing.UsageReportingGranularityHour:
		granularity = time.Hour
		gracePeriod = time.Hour
		sqlGrainIdentifier = "hour"
	case billing.UsageReportingGranularityNone:
		w.logger.Debug("skipping usage reporting: no reporting granularity configured")
		return nil
	default:
		return fmt.Errorf("unsupported reporting granularity: %s", w.admin.Biller.GetReportingGranularity())
	}

	t, err := w.admin.DB.FindBillingUsageReportedOn(ctx)
	if err != nil {
		return fmt.Errorf("failed to get last usage reporting time: %w", err)
	}

	endTime := time.Now().UTC().Add(-gracePeriod).Truncate(granularity)
	var startTime time.Time
	if t.IsZero() {
		startTime = endTime.Add(-granularity)
	} else {
		startTime = t.UTC()
	}

	if !startTime.Before(endTime) {
		w.logger.Debug("skipping usage reporting: no new usage data available", zap.Time("start_time", startTime), zap.Time("end_time", endTime))
		return nil
	}

	client, ok, err := w.admin.OpenMetricsProject(ctx)
	if err != nil {
		w.logger.Error("failed to report usage: unable to get metrics client", zap.Error(err))
		return err
	}
	if !ok {
		w.logger.Debug("skipping usage reporting: no metrics project configured")
		return nil
	}

	reportedOrgs := make(map[string]struct{})
	stop := false
	limit := 10000
	afterTime := time.Time{}
	afterOrgID := ""
	afterProjectID := ""
	afterEventName := ""

	checkPoint := startTime
	maxEndTime := time.Time{}
	for !stop {
		u, err := client.GetUsageMetrics(ctx, startTime, endTime, afterTime, afterOrgID, afterProjectID, afterEventName, sqlGrainIdentifier, limit)
		if err != nil {
			return fmt.Errorf("failed to get usage metrics: %w", err)
		}

		if len(u) == 0 {
			break
		}

		if len(u) < limit {
			stop = true
		} else {
			afterTime = u[len(u)-1].StartTime
			afterOrgID = u[len(u)-1].OrgID
			afterProjectID = u[len(u)-1].ProjectID
			afterEventName = u[len(u)-1].EventName
		}
		maxEndTime = u[len(u)-1].EndTime

		var usage []*billing.Usage
		for _, m := range u {
			reportedOrgs[m.OrgID] = struct{}{}
			customerID := m.OrgID
			if m.BillingCustomerID != nil && *m.BillingCustomerID != "" {
				customerID = *m.BillingCustomerID
			}
			usage = append(usage, &billing.Usage{
				CustomerID:     customerID,
				MetricName:     m.EventName,
				Value:          m.MaxValue,
				ReportingGrain: w.admin.Biller.GetReportingGranularity(),
				StartTime:      m.StartTime,
				EndTime:        m.EndTime,
				Metadata: map[string]interface{}{
					"org_id":          m.OrgID,
					"project_id":      m.ProjectID,
					"project_name":    m.ProjectName,
					"billing_service": m.BillingService,
				},
			})
		}

		err = w.admin.Biller.ReportUsage(ctx, usage)
		if err != nil {
			return fmt.Errorf("failed to report usage: %w", err)
		}

		if afterTime.After(checkPoint) {
			checkPoint = afterTime
			err = w.admin.DB.UpdateBillingUsageReportedOn(ctx, checkPoint)
			if err != nil {
				return fmt.Errorf("failed to update last usage reporting time: %w", err)
			}
		}
	}

	if len(reportedOrgs) == 0 {
		w.logger.Named("billing").Warn("skipping usage reporting: no usage data available", zap.Time("start_time", startTime), zap.Time("end_time", endTime))
		return nil
	}

	if maxEndTime.IsZero() {
		return fmt.Errorf("failed to update last usage reporting time: max end time not updated after reporting usage data")
	}

	err = w.admin.DB.UpdateBillingUsageReportedOn(ctx, maxEndTime)
	if err != nil {
		return fmt.Errorf("failed to update last usage reporting time: %w", err)
	}

	orgs, err := w.admin.DB.FindOrganizationIDsWithBilling(ctx)
	if err != nil {
		return fmt.Errorf("failed to report usage: unable to fetch orgs: %w", err)
	}

	for _, org := range orgs {
		if _, ok := reportedOrgs[org]; !ok {
			count, err := w.admin.DB.CountBillingProjectsForOrganization(ctx, org, endTime)
			if err != nil {
				w.logger.Warn("failed to validate active projects for org", zap.String("org_id", org), zap.Error(err))
				continue
			}
			if count > 0 {
				w.logger.Warn("skipped usage reporting for org as no usage data was available", zap.String("org_id", org), zap.Time("start_time", startTime), zap.Time("end_time", endTime))
			}
		}
	}
	return nil
}
