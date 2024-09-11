package river

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/runtime/pkg/email"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
	"go.uber.org/zap"
)

type InvoicePaymentFailedArgs struct {
	BillingCustomerID string
	InvoiceID         string
	InvoiceNumber     string
	InvoiceURL        string
	Amount            string
	Currency          string
	DueDate           time.Time
	FailedAt          time.Time
}

func (InvoicePaymentFailedArgs) Kind() string { return "invoice_payment_failed" }

type InvoicePaymentFailedWorker struct {
	river.WorkerDefaults[InvoicePaymentFailedArgs]
	admin         *admin.Service
	billingLogger *zap.Logger
}

func (w *InvoicePaymentFailedWorker) Work(ctx context.Context, job *river.Job[InvoicePaymentFailedArgs]) error {
	org, err := w.admin.DB.FindOrganizationForBillingCustomerID(ctx, job.Args.BillingCustomerID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			// org got deleted, ignore
			return nil
		}
		return fmt.Errorf("failed to find organization of billing customer id %q: %w", job.Args.BillingCustomerID, err)
	}

	// schedule a job to check if the invoice is paid after end of grace period
	gracePeriodEndDate := job.Args.FailedAt.Truncate(24*time.Hour).AddDate(0, 0, gracePeriodDays)
	j, err := w.admin.Jobs.InvoicePaymentFailedGracePeriodCheck(ctx, org.ID, job.Args.InvoiceID, gracePeriodEndDate)
	if err != nil {
		return fmt.Errorf("failed to schedule invoice payment failed grace period check job: %w", err)
	}

	be, err := w.admin.DB.FindBillingErrorByType(ctx, org.ID, database.BillingErrorTypeInvoicePaymentFailed)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			return fmt.Errorf("failed to find billing errors: %w", err)
		}
	}
	var metadata *database.BillingErrorMetadataInvoicePaymentFailed
	if be != nil {
		metadata = be.Metadata.(*database.BillingErrorMetadataInvoicePaymentFailed)
	} else {
		metadata = &database.BillingErrorMetadataInvoicePaymentFailed{
			Invoices: make(map[string]database.InvoicePaymentFailedMeta),
		}
	}

	metadata.Invoices[job.Args.InvoiceID] = database.InvoicePaymentFailedMeta{
		ID:                  job.Args.InvoiceID,
		Number:              job.Args.InvoiceNumber,
		URL:                 job.Args.InvoiceURL,
		Amount:              job.Args.Amount,
		Currency:            job.Args.Currency,
		DueDate:             job.Args.DueDate,
		FailedOn:            job.Args.FailedAt,
		GracePeriodEndJobID: j.ID,
	}

	// insert billing error
	_, err = w.admin.DB.UpsertBillingError(ctx, &database.UpsertBillingErrorOptions{
		OrgID:     org.ID,
		Type:      database.BillingErrorTypeInvoicePaymentFailed,
		Metadata:  &database.BillingErrorMetadataInvoicePaymentFailed{Invoices: metadata.Invoices},
		EventTime: job.Args.FailedAt,
	})
	if err != nil {
		return fmt.Errorf("failed to add billing error: %w", err)
	}

	err = w.admin.Email.SendInvoicePaymentFailed(&email.InvoicePaymentFailed{
		ToEmail:            org.BillingEmail,
		ToName:             org.Name,
		OrgName:            org.Name,
		Currency:           job.Args.Currency,
		Amount:             job.Args.Amount,
		GracePeriodEndDate: gracePeriodEndDate,
	})
	if err != nil {
		return fmt.Errorf("failed to send invoice payment failed email for org %q: %w", org.Name, err)
	}
	w.billingLogger.Warn("invoice payment failed", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.String("amount", job.Args.Amount), zap.Time("due_date", job.Args.DueDate), zap.String("invoice_id", job.Args.InvoiceID), zap.String("invoice_url", job.Args.InvoiceURL))

	return nil
}

type InvoicePaymentSuccessArgs struct {
	BillingCustomerID string
	InvoiceID         string
}

func (InvoicePaymentSuccessArgs) Kind() string { return "invoice_payment_success" }

type InvoicePaymentSuccessWorker struct {
	river.WorkerDefaults[InvoicePaymentSuccessArgs]
	admin         *admin.Service
	billingLogger *zap.Logger
}

func (w *InvoicePaymentSuccessWorker) Work(ctx context.Context, job *river.Job[InvoicePaymentSuccessArgs]) error {
	org, err := w.admin.DB.FindOrganizationForBillingCustomerID(ctx, job.Args.BillingCustomerID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			// org got deleted, ignore
			return nil
		}
		return fmt.Errorf("failed to find organization of billing customer id %q: %w", job.Args.BillingCustomerID, err)
	}

	// check for existing billing error and delete it
	be, err := w.admin.DB.FindBillingErrorByType(ctx, org.ID, database.BillingErrorTypeInvoicePaymentFailed)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			// no billing error, ignore
			return nil
		}
		return fmt.Errorf("failed to find billing errors: %w", err)
	}

	failedInvoices := be.Metadata.(*database.BillingErrorMetadataInvoicePaymentFailed).Invoices
	failedInvoice, ok := failedInvoices[job.Args.InvoiceID]
	if !ok {
		// invoice not found in the failed invoices, do nothing
		return nil
	}

	// remove any scheduled job for invoice payment failed grace period check
	if failedInvoice.GracePeriodEndJobID > 0 { // river job ids starts from 1
		err = w.admin.Jobs.CancelJob(ctx, failedInvoice.GracePeriodEndJobID)
		if err != nil {
			if !errors.Is(err, rivertype.ErrNotFound) {
				w.billingLogger.Error("failed to cancel grace period check job", zap.Error(err), zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.String("invoice_id", job.Args.InvoiceID), zap.Error(err))
				// don't return error, continue as the grace period check worker will check the billing error and invoice status before doing anything else
			}
		}
	}
	delete(failedInvoices, job.Args.InvoiceID)
	w.billingLogger.Info("invoice payment success for a failed invoice", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.String("invoice_id", job.Args.InvoiceID))

	// if no more failed invoices, delete the billing error
	if len(failedInvoices) == 0 {
		err = w.admin.DB.DeleteBillingError(ctx, be.ID)
		if err != nil {
			return fmt.Errorf("failed to delete billing error: %w", err)
		}
	} else {
		// update the metadata
		_, err = w.admin.DB.UpsertBillingError(ctx, &database.UpsertBillingErrorOptions{
			OrgID:     org.ID,
			Type:      database.BillingErrorTypeInvoicePaymentFailed,
			Metadata:  &database.BillingErrorMetadataInvoicePaymentFailed{Invoices: failedInvoices},
			EventTime: be.EventTime,
		})
		if err != nil {
			return fmt.Errorf("failed to update billing error: %w", err)
		}
	}

	// send email
	err = w.admin.Email.SendInvoicePaymentSuccess(&email.InvoicePaymentSuccess{
		ToEmail:  org.BillingEmail,
		ToName:   org.Name,
		OrgName:  org.Name,
		Currency: failedInvoice.Currency,
		Amount:   failedInvoice.Amount,
	})
	if err != nil {
		// ignore email sending error
		w.billingLogger.Error("failed to send invoice payment success email", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.String("invoice_id", job.Args.InvoiceID), zap.Error(err))
	}

	return nil
}

type InvoicePaymentFailedGracePeriodCheckArgs struct {
	OrgID              string
	InvoiceID          string
	GracePeriodEndDate time.Time
}

func (InvoicePaymentFailedGracePeriodCheckArgs) Kind() string {
	return "invoice_payment_failed_grace_period_check"
}

type InvoicePaymentFailedGracePeriodCheckWorker struct {
	river.WorkerDefaults[InvoicePaymentFailedGracePeriodCheckArgs]
	admin         *admin.Service
	billingLogger *zap.Logger
}

func (w *InvoicePaymentFailedGracePeriodCheckWorker) Work(ctx context.Context, job *river.Job[InvoicePaymentFailedGracePeriodCheckArgs]) error {
	org, err := w.admin.DB.FindOrganization(ctx, job.Args.OrgID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			// org got deleted, ignore
			return nil
		}
		return fmt.Errorf("failed to find organization: %w", err)
	}

	if time.Now().UTC().Before(job.Args.GracePeriodEndDate.AddDate(0, 0, 1)) {
		return fmt.Errorf("grace period date %s not finished yet for org %q", job.Args.GracePeriodEndDate, org.Name) // will be retried later
	}

	// check if the org has still invoice failed billing error
	be, err := w.admin.DB.FindBillingErrorByType(ctx, org.ID, database.BillingErrorTypeInvoicePaymentFailed)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			// no billing error, ignore
			return nil
		}
		return fmt.Errorf("failed to find billing errors: %w", err)
	}

	failedInvoices := be.Metadata.(*database.BillingErrorMetadataInvoicePaymentFailed).Invoices
	// check if the invoice is still in the failed invoices
	if _, ok := failedInvoices[job.Args.InvoiceID]; !ok {
		// invoice is not in the failed invoices, do nothing
		return nil
	}

	// just to be very sure, check if the invoice is still unpaid directly from the biller
	invoice, err := w.admin.Biller.GetInvoice(ctx, job.Args.InvoiceID)
	if err != nil {
		return fmt.Errorf("failed to get invoice %q: %w", job.Args.InvoiceID, err)
	}

	if w.admin.Biller.IsInvoicePaid(ctx, invoice) || !w.admin.Biller.IsInvoiceValid(ctx, invoice) {
		w.admin.Logger.Warn("Invoice was already paid or invalid but billing error was not cleared", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.String("billing_customer_id", org.BillingCustomerID), zap.String("invoice_id", job.Args.InvoiceID), zap.String("invoice_status", invoice.Status))

		// clearing the billing error for this invoice
		delete(failedInvoices, job.Args.InvoiceID)

		// if no more failed invoices, delete the billing error
		if len(failedInvoices) == 0 {
			err = w.admin.DB.DeleteBillingError(ctx, be.ID)
			if err != nil {
				return fmt.Errorf("failed to delete billing error: %w", err)
			}
		} else {
			// update the metadata
			_, err = w.admin.DB.UpsertBillingError(ctx, &database.UpsertBillingErrorOptions{
				OrgID:     org.ID,
				Type:      database.BillingErrorTypeInvoicePaymentFailed,
				Metadata:  &database.BillingErrorMetadataInvoicePaymentFailed{Invoices: failedInvoices},
				EventTime: be.EventTime,
			})
			if err != nil {
				return fmt.Errorf("failed to update billing error: %w", err)
			}
		}
		return nil
	}

	// hibernate projects
	limit := 10
	afterProjectName := ""
	for {
		projs, err := w.admin.DB.FindProjectsForOrganization(ctx, org.ID, afterProjectName, limit)
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
	w.billingLogger.Warn("projects hibernated due to unpaid invoice", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.String("invoice_id", job.Args.InvoiceID))

	// send email
	err = w.admin.Email.SendInvoiceUnpaid(&email.InvoiceUnpaid{
		ToEmail:  org.BillingEmail,
		ToName:   org.Name,
		OrgName:  org.Name,
		Currency: invoice.Currency,
		Amount:   invoice.Amount,
	})
	if err != nil {
		return fmt.Errorf("failed to send payment method expired email for org %q: %w", org.Name, err)
	}

	return nil
}
