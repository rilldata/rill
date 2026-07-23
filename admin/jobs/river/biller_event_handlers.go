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
	invoice, err := w.admin.Biller.GetInvoice(ctx, job.Args.InvoiceID)
	if err != nil {
		return fmt.Errorf("failed to verify invoice %q after payment failure: %w", job.Args.InvoiceID, err)
	}
	if invoice != nil && (!w.admin.Biller.IsInvoiceValid(ctx, invoice) || w.admin.Biller.IsInvoicePaid(ctx, invoice)) {
		// A delayed failure event must not resurrect an issue after the invoice has
		// already been paid or voided.
		w.logger.Info("ignoring stale invoice payment failure", zap.String("org_id", org.ID), zap.String("invoice_id", job.Args.InvoiceID), zap.String("invoice_status", invoice.Status))
		return nil
	}

	be, err := w.admin.DB.FindBillingIssueByTypeForOrg(ctx, org.ID, database.BillingIssueTypePaymentFailed)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			return fmt.Errorf("failed to find billing errors: %w", err)
		}
	}
	invoices := make(map[string]*database.BillingIssueMetadataPaymentFailedMeta)
	eventTime := job.Args.FailedAt
	if be != nil {
		metadata := be.Metadata.(*database.BillingIssueMetadataPaymentFailed)
		for id, failedInvoice := range metadata.Invoices {
			copy := *failedInvoice
			invoices[id] = &copy
		}
		if existing := invoices[job.Args.InvoiceID]; existing != nil && !job.Args.FailedAt.After(existing.FailedOn) {
			// The invoice entry itself is the durable marker for the at-most-once
			// failure-email attempt. Duplicates and older deliveries are no-ops.
			return nil
		}
		if be.EventTime.After(eventTime) {
			eventTime = be.EventTime
		}
	}

	gracePeriodEndDate := job.Args.DueDate.AddDate(0, 0, database.BillingGracePeriodDays)
	invoices[job.Args.InvoiceID] = &database.BillingIssueMetadataPaymentFailedMeta{
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
		Metadata:  &database.BillingIssueMetadataPaymentFailed{Invoices: invoices},
		EventTime: eventTime,
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
	now    func() time.Time
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
	invoice, err := w.admin.Biller.GetInvoice(ctx, job.Args.InvoiceID)
	if err != nil {
		return fmt.Errorf("failed to verify invoice %q after payment success: %w", job.Args.InvoiceID, err)
	}
	if invoice != nil && w.admin.Biller.IsInvoiceValid(ctx, invoice) && !w.admin.Biller.IsInvoicePaid(ctx, invoice) {
		// A delayed success event from an earlier attempt must not clear a newer
		// failure while the provider still considers the invoice unpaid.
		w.logger.Info("ignoring stale invoice payment success", zap.String("org_id", org.ID), zap.String("invoice_id", job.Args.InvoiceID), zap.String("invoice_status", invoice.Status))
		return nil
	}

	remainingInvoices := make(map[string]*database.BillingIssueMetadataPaymentFailedMeta, len(failedInvoices)-1)
	for id, failedInvoice := range failedInvoices {
		if id == job.Args.InvoiceID {
			continue
		}
		copy := *failedInvoice
		remainingInvoices[id] = &copy
	}
	w.logger.Info("invoice payment success for a failed invoice", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.String("invoice_id", job.Args.InvoiceID))

	// if no more failed invoices, delete the billing error
	if len(remainingInvoices) == 0 {
		err = w.admin.DB.DeleteBillingIssue(ctx, be.ID)
		if err != nil {
			return fmt.Errorf("failed to delete billing error: %w", err)
		}
	} else {
		// update the metadata
		_, err = w.admin.DB.UpsertBillingIssue(ctx, &database.UpsertBillingIssueOptions{
			OrgID:     org.ID,
			Type:      database.BillingIssueTypePaymentFailed,
			Metadata:  &database.BillingIssueMetadataPaymentFailed{Invoices: remainingInvoices},
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
		PaymentDate:    billingLifecycleNow(w.now),
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
	now    func() time.Time
}

func (w *PaymentFailedGracePeriodCheckWorker) Work(ctx context.Context, job *river.Job[PaymentFailedGracePeriodCheckArgs]) error {
	failures, err := w.admin.DB.FindBillingIssueByTypeAndOverdueProcessed(ctx, database.BillingIssueTypePaymentFailed, false)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			// no orgs have this billing error
			return nil
		}
		return fmt.Errorf("failed to find organization with billing issue: %w", err)
	}

	now := billingLifecycleNow(w.now)
	var workErr error
	// failures are per org
	for _, f := range failures {
		if err := w.processIssue(ctx, f, now); err != nil {
			w.logger.Error("failed to process overdue invoices for org", zap.String("org_id", f.OrgID), zap.Error(err))
			workErr = errors.Join(workErr, fmt.Errorf("process overdue invoices for org %q: %w", f.OrgID, err))
		}
	}
	return workErr
}

func (w *PaymentFailedGracePeriodCheckWorker) processIssue(ctx context.Context, issue *database.BillingIssue, now time.Time) error {
	overdue, err := w.checkFailedInvoicesForOrg(ctx, issue, now)
	if err != nil || !overdue {
		return err
	}
	if _, err := hibernateProjectsForOrganization(ctx, w.admin, issue.OrgID); err != nil {
		return err
	}
	org, err := w.admin.DB.FindOrganization(ctx, issue.OrgID)
	if err != nil {
		return fmt.Errorf("failed to find organization: %w", err)
	}
	w.logger.Warn("projects hibernated due to unpaid invoice", zap.String("org_id", org.ID), zap.String("org_name", org.Name))

	// Mark first so a processed-marker failure cannot follow a successful email
	// and cause a duplicate on retry.
	if err := w.admin.DB.UpdateBillingIssueOverdueAsProcessed(ctx, issue.ID); err != nil {
		return fmt.Errorf("failed to mark billing issue as processed: %w", err)
	}
	err = w.admin.Email.SendInvoiceUnpaid(&email.InvoiceUnpaid{
		ToEmail:    org.BillingEmail,
		ToName:     org.Name,
		OrgName:    org.Name,
		PaymentURL: w.admin.URLs.PaymentPortal(org.Name),
	})
	if err != nil {
		w.logger.Error("failed to send invoice unpaid email after marking it attempted", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.Error(err))
	}
	return nil
}

// reconciles failed payments for the org and returns true if any is overdue
func (w *PaymentFailedGracePeriodCheckWorker) checkFailedInvoicesForOrg(ctx context.Context, orgPaymentFailures *database.BillingIssue, now time.Time) (bool, error) {
	hasOverdue := false
	for invoiceID, failedInvoice := range orgPaymentFailures.Metadata.(*database.BillingIssueMetadataPaymentFailed).Invoices {
		if now.Before(failedInvoice.GracePeriodEndDate.AddDate(0, 0, 1)) {
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
			LogoDarkAssetID:                     org.LogoDarkAssetID,
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
			QuotaSeats:                          org.QuotaSeats,
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
