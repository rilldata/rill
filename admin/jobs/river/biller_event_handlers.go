package river

import (
	"context"
	"errors"
	"fmt"
	"html/template"
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
	admin *admin.Service
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

	// schedule a job to check if the invoice is paid after grace period days
	// add a buffer of 1 hour to ensure the job runs after grace period days
	j, err := w.admin.Jobs.InvoicePaymentFailedGracePeriodCheck(ctx, org.ID, job.Args.InvoiceID, time.Now().Truncate(24*time.Hour).AddDate(0, 0, gracePeriodDays+1).Add(time.Hour*1))
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

	// send email
	err = w.admin.Email.SendInformational(&email.Informational{
		ToEmail: org.BillingEmail,
		ToName:  org.Name,
		Subject: "Your invoice payment failed",
		Title:   "",
		Body:    template.HTML(fmt.Sprintf("Your invoice payment failed, please visit the billing portal to fix issues or contact support to continue using Rill. Your projects will be hibernated after %d days, if invoice still not paid.", gracePeriodDays)),
	})
	if err != nil {
		return fmt.Errorf("failed to send invoice payment failed email for org %q: %w", org.Name, err)
	}
	w.admin.Logger.Info("email sent for invoice payment failed", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.String("billing_customer_id", org.BillingCustomerID))

	return nil
}

type InvoicePaymentSuccessArgs struct {
	BillingCustomerID string
	InvoiceID         string
}

func (InvoicePaymentSuccessArgs) Kind() string { return "invoice_payment_success" }

type InvoicePaymentSuccessWorker struct {
	river.WorkerDefaults[InvoicePaymentSuccessArgs]
	admin *admin.Service
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
		if !errors.Is(err, database.ErrNotFound) {
			return fmt.Errorf("failed to find billing errors: %w", err)
		}
	}

	// delete the invoice failed error
	if be != nil {
		failedInvoices := be.Metadata.(*database.BillingErrorMetadataInvoicePaymentFailed).Invoices
		// if found remove the invoice from the failed invoices
		if i, ok := failedInvoices[job.Args.InvoiceID]; ok {
			// remove any scheduled job for invoice payment failed grace period check
			if i.GracePeriodEndJobID > 0 { // river job ids starts from 1
				err = w.admin.Jobs.CancelJob(ctx, i.GracePeriodEndJobID)
				if err != nil {
					if !errors.Is(err, rivertype.ErrNotFound) {
						return fmt.Errorf("failed to cancel invoice payment failed grace period check job: %w", err)
					}
				}
			}
			delete(failedInvoices, job.Args.InvoiceID)
		}

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
	}

	return nil
}

type InvoicePaymentFailedGracePeriodCheckArgs struct {
	OrgID     string
	InvoiceID string
}

func (InvoicePaymentFailedGracePeriodCheckArgs) Kind() string {
	return "invoice_payment_failed_grace_period_check"
}

type InvoicePaymentFailedGracePeriodCheckWorker struct {
	river.WorkerDefaults[InvoicePaymentFailedGracePeriodCheckArgs]
	admin *admin.Service
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

	// check if the org has still invoice failed billing error
	be, err := w.admin.DB.FindBillingErrorByType(ctx, org.ID, database.BillingErrorTypeInvoicePaymentFailed)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			return fmt.Errorf("failed to find billing errors: %w", err)
		}
	}

	if be != nil {
		// just to be very sure, check if the invoice is still unpaid directly from the biller
		invoice, err := w.admin.Biller.GetInvoice(ctx, job.Args.InvoiceID)
		if err != nil {
			return fmt.Errorf("failed to get invoice %q: %w", job.Args.InvoiceID, err)
		}

		if w.admin.Biller.IsInvoicePaid(ctx, invoice) || !w.admin.Biller.IsInvoiceValid(ctx, invoice) {
			w.admin.Logger.Warn("Invoice was already paid or invalid but billing error was not cleared", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.String("billing_customer_id", org.BillingCustomerID), zap.String("invoice_id", job.Args.InvoiceID), zap.String("invoice_status", invoice.Status))

			// clearing the billing error for this invoice
			failedInvoices := be.Metadata.(*database.BillingErrorMetadataInvoicePaymentFailed).Invoices
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
		w.admin.Logger.Info("Projects hibernated due to no valid payment method", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.String("billing_customer_id", org.BillingCustomerID))

		// send email
		err = w.admin.Email.SendInformational(&email.Informational{
			ToEmail: org.BillingEmail,
			ToName:  org.Name,
			Subject: "No valid payment method found",
			Title:   "",
			Body:    "We did not receive any payment for the unpaid invoice, your projects have been hibernated. Please visit the billing portal to enter payment method and then run `rill project reconcile` to unhibernate each project.",
		})
		if err != nil {
			return fmt.Errorf("failed to send payment method expired email for org %q: %w", org.Name, err)
		}
		w.admin.Logger.Info("email sent for projects hibernated", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.String("billing_customer_id", org.BillingCustomerID))
	}

	return nil
}
