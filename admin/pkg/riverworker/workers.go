package riverworker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/pkg/riverworker/riverutils"
	"github.com/rilldata/rill/runtime/pkg/email"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
	"go.uber.org/zap"
)

var gracePeriodDays = 9

var Workers = river.NewWorkers()

func AddWorker[T river.JobArgs](worker river.Worker[T]) {
	river.AddWorker[T](Workers, worker)
}

func NewChargeSuccessWorker(adm *admin.Service) *ChargeSuccessWorker {
	return &ChargeSuccessWorker{admin: adm}
}

type ChargeSuccessWorker struct {
	river.WorkerDefaults[riverutils.ChargeSuccessArgs]
	admin *admin.Service
}

func (w *ChargeSuccessWorker) Work(ctx context.Context, job *river.Job[riverutils.ChargeSuccessArgs]) error {
	org, err := w.admin.DB.FindOrganizationForPaymentCustomerID(ctx, job.Args.PaymentCustomerID)
	if err != nil {
		return fmt.Errorf("failed to find organization for payment customer id: %w", err)
	}

	// check for existing billing error and delete if it is older than the event time
	be, err := w.admin.DB.FindBillingErrorByType(ctx, org.ID, database.BillingErrorTypePaymentFailed)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			return fmt.Errorf("failed to find billing errors: %w", err)
		}
	}

	// TODO may be do updates in a transaction
	// delete the payment failed error that are older than the event time
	if be != nil && job.Args.EventTime.After(be.EventTime) {
		err = w.admin.DB.DeleteBillingError(ctx, be.ID)
		if err != nil {
			return fmt.Errorf("failed to delete billing error: %w", err)
		}
	}

	// update latest event time for the org
	_, err = w.admin.DB.UpsertWebhookEventWatermark(ctx, &database.UpsertWebhookEventOptions{
		OrgID:          org.ID,
		Type:           database.StripeWebhookEventTypeChargeSucceeded,
		LastOccurrence: job.Args.EventTime,
	})
	if err != nil {
		return fmt.Errorf("failed to update webhook event watermark: %w", err)
	}

	return nil
}

func NewChargeFailedWorker(adm *admin.Service) *ChargeFailedWorker {
	return &ChargeFailedWorker{admin: adm}
}

// ChargeFailedWorker worker to add billing error of payment failed in the billing_error table for an org
type ChargeFailedWorker struct {
	river.WorkerDefaults[riverutils.ChargeFailedArgs]
	admin *admin.Service
}

func (w *ChargeFailedWorker) Work(ctx context.Context, job *river.Job[riverutils.ChargeFailedArgs]) error {
	org, err := w.admin.DB.FindOrganizationForPaymentCustomerID(ctx, job.Args.PaymentCustomerID)
	if err != nil {
		return fmt.Errorf("failed to find organization for payment customer id: %w", err)
	}
	// check if there is any charge success event after this charge failed event, if yes then ignore this charge failed event
	event, err := w.admin.DB.FindWebhookEventWatermark(ctx, org.ID, database.StripeWebhookEventTypeChargeSucceeded)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			return fmt.Errorf("failed to find webhook event watermark: %w", err)
		}
	}
	if event != nil && event.LastOccurrence.After(job.Args.EventTime) {
		return nil
	}

	_, err = w.admin.DB.UpsertBillingError(ctx, &database.UpsertBillingErrorOptions{
		OrgID:     org.ID,
		Type:      database.BillingErrorTypePaymentFailed,
		Message:   fmt.Sprintf("Recent payment of %s %d failed, please fix by visiting the billing portal. Charge id:%s", strings.ToUpper(job.Args.Currency), job.Args.Amount, job.Args.ID),
		EventTime: job.Args.EventTime,
	})
	if err != nil {
		return fmt.Errorf("failed to add billing error: %w", err)
	}

	_, err = w.admin.DB.UpsertWebhookEventWatermark(ctx, &database.UpsertWebhookEventOptions{
		OrgID:          org.ID,
		Type:           database.StripeWebhookEventTypeChargeFailed,
		LastOccurrence: job.Args.EventTime,
	})
	if err != nil {
		return fmt.Errorf("failed to update webhook event watermark: %w", err)
	}

	return nil
}

func NewPaymentMethodAddedWorker(adm *admin.Service) *PaymentMethodAddedWorker {
	return &PaymentMethodAddedWorker{admin: adm}
}

type PaymentMethodAddedWorker struct {
	river.WorkerDefaults[riverutils.PaymentMethodAddedArgs]
	admin *admin.Service
}

func (w *PaymentMethodAddedWorker) Work(ctx context.Context, job *river.Job[riverutils.PaymentMethodAddedArgs]) error {
	org, err := w.admin.DB.FindOrganizationForPaymentCustomerID(ctx, job.Args.PaymentCustomerID)
	if err != nil {
		return fmt.Errorf("failed to find organization for payment customer id: %w", err)
	}

	// check for no payment method billing error and delete if it is older than the event time
	be, err := w.admin.DB.FindBillingErrorByType(ctx, org.ID, database.BillingErrorTypeNoPaymentMethod)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			return fmt.Errorf("failed to find billing errors: %w", err)
		}
	}

	// delete the no payment method error if older than the event time
	if be != nil && job.Args.EventTime.After(be.EventTime) {
		err = w.admin.DB.DeleteBillingError(ctx, be.ID)
		if err != nil {
			return fmt.Errorf("failed to delete billing error: %w", err)
		}
	}

	// update latest event time for the org
	_, err = w.admin.DB.UpsertWebhookEventWatermark(ctx, &database.UpsertWebhookEventOptions{
		OrgID:          org.ID,
		Type:           database.StripeWebhookEventTypePaymentMethodAttached,
		LastOccurrence: job.Args.EventTime,
	})
	if err != nil {
		return fmt.Errorf("failed to update webhook event watermark: %w", err)
	}

	return nil
}

func NewPaymentMethodRemovedWorker(adm *admin.Service) *PaymentMethodRemovedWorker {
	return &PaymentMethodRemovedWorker{admin: adm}
}

type PaymentMethodRemovedWorker struct {
	river.WorkerDefaults[riverutils.PaymentMethodRemovedArgs]
	admin *admin.Service
}

func (w *PaymentMethodRemovedWorker) Work(ctx context.Context, job *river.Job[riverutils.PaymentMethodRemovedArgs]) error {
	org, err := w.admin.DB.FindOrganizationForPaymentCustomerID(ctx, job.Args.PaymentCustomerID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			// org got deleted, ignore
			return nil
		}
		return fmt.Errorf("failed to find organization for payment customer id: %w", err)
	}

	// check if there is any payment added event after this event, if yes then ignore this event
	event, err := w.admin.DB.FindWebhookEventWatermark(ctx, org.ID, database.StripeWebhookEventTypePaymentMethodAttached)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			return fmt.Errorf("failed to find webhook event watermark: %w", err)
		}
	}
	if event != nil && event.LastOccurrence.After(job.Args.EventTime) {
		return nil
	}

	_, err = w.admin.DB.UpsertBillingError(ctx, &database.UpsertBillingErrorOptions{
		OrgID:     org.ID,
		Type:      database.BillingErrorTypeNoPaymentMethod,
		Message:   "No payment method attached, please add a payment method by visiting the billing portal",
		EventTime: job.Args.EventTime,
	})
	if err != nil {
		return fmt.Errorf("failed to add billing error: %w", err)
	}

	_, err = w.admin.DB.UpsertWebhookEventWatermark(ctx, &database.UpsertWebhookEventOptions{
		OrgID:          org.ID,
		Type:           database.StripeWebhookEventTypePaymentMethodDetached,
		LastOccurrence: job.Args.EventTime,
	})
	if err != nil {
		return fmt.Errorf("failed to update webhook event watermark: %w", err)
	}

	return nil
}

func NewTrialEndingSoonWorker(adm *admin.Service) *TrialEndingSoonWorker {
	return &TrialEndingSoonWorker{admin: adm}
}

type TrialEndingSoonWorker struct {
	river.WorkerDefaults[riverutils.TrialEndingSoonArgs]
	admin *admin.Service
}

func (w *TrialEndingSoonWorker) Work(ctx context.Context, job *river.Job[riverutils.TrialEndingSoonArgs]) error {
	org, err := w.admin.DB.FindOrganization(ctx, job.Args.OrgID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			// org got deleted, ignore
			return nil
		}
		return fmt.Errorf("failed to find organization: %w", err)
	}

	// check if the org has any active subscription
	sub, err := w.admin.Biller.GetSubscriptionsForCustomer(ctx, org.BillingCustomerID)
	if err != nil {
		return fmt.Errorf("failed to get subscriptions for org %q: %w", org.Name, err)
	}

	if len(sub) == 0 {
		return fmt.Errorf("no active subscription found for the org %q", org.Name)
	}

	if len(sub) > 1 {
		return fmt.Errorf("multiple active subscriptions found for the org %q", org.Name)
	}

	if sub[0].ID != job.Args.SubID || sub[0].Plan.ID != job.Args.PlanID || sub[0].TrialEndDate.IsZero() {
		// subscription or plan might have changed, ignore
		return nil
	}

	if sub[0].TrialEndDate.After(time.Now().UTC()) {
		// just relying on trial period days to check if still on trial plan, may need explicit flag for this in the plan
		// trial period is ending soon, log error, schedule trial end check job and send email
		w.admin.Logger.Info("Trial period is ending soon", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.String("billing_customer_id", org.BillingCustomerID), zap.Time("trial_end_date", sub[0].TrialEndDate))

		// this should not happen but just in case directly schedule the trial end check job
		_, err := riverutils.InsertOnlyRiverClient.Insert(ctx, riverutils.TrialEndCheckArgs{OrgID: org.ID}, &river.InsertOpts{
			ScheduledAt: sub[0].TrialEndDate.Add(time.Hour * 25), // add buffer of 1 hour to ensure the job runs after trial period days
			UniqueOpts: river.UniqueOpts{
				ByArgs: true,
			},
		})
		if err != nil {
			return fmt.Errorf("failed to schedule trial end check job: %w", err)
		}

		// send email
		err = w.admin.Email.SendInformational(&email.Informational{
			ToEmail: org.BillingEmail,
			ToName:  org.Name,
			Subject: "Your trial period is ending soon",
			Title:   "",
			Body:    template.HTML(fmt.Sprintf("Just a reminder that your trial period will end on %s. Reach out to us through various means listed at https://docs.rilldata.com/contact for any help.", sub[0].TrialEndDate.Format("2006-01-02"))),
		})
		if err != nil {
			return fmt.Errorf("failed to send trial ending soon email for org %q: %w", org.Name, err)
		}
		w.admin.Logger.Info("Email sent for trial period ending soon", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.String("billing_customer_id", org.BillingCustomerID))
	} else {
		// this cannot happen but woke up after schedule or some error in calculating scheduling time
		w.admin.Logger.Warn("Trial period has already ended before check was run, please check the org manually", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.String("billing_customer_id", org.BillingCustomerID))
	}

	return nil
}

func NewTrialEndCheckWorker(adm *admin.Service) *TrialEndCheckWorker {
	return &TrialEndCheckWorker{admin: adm}
}

type TrialEndCheckWorker struct {
	river.WorkerDefaults[riverutils.TrialEndCheckArgs]
	admin *admin.Service
}

func (w *TrialEndCheckWorker) Work(ctx context.Context, job *river.Job[riverutils.TrialEndCheckArgs]) error {
	org, err := w.admin.DB.FindOrganization(ctx, job.Args.OrgID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			// org got deleted, ignore
			return nil
		}
		return fmt.Errorf("failed to find organization: %w", err)
	}

	// check if the org has any active subscription
	sub, err := w.admin.Biller.GetSubscriptionsForCustomer(ctx, org.BillingCustomerID)
	if err != nil {
		return fmt.Errorf("failed to get subscriptions for org %q: %w", org.Name, err)
	}

	if len(sub) == 0 {
		return fmt.Errorf("no active subscription found for the org %q", org.Name)
	}

	if len(sub) > 1 {
		return fmt.Errorf("multiple active subscriptions found for the org %q", org.Name)
	}

	if sub[0].ID != job.Args.SubID || sub[0].Plan.ID != job.Args.PlanID || sub[0].TrialEndDate.IsZero() {
		// subscription or plan might have changed, ignore
		return nil
	}

	if sub[0].TrialEndDate.Before(time.Now().UTC()) {
		// trial period has ended, log error, send email and schedule a job to hibernate projects after grace period days if still on trial
		w.admin.Logger.Error("Trial period has ended", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.String("billing_customer_id", org.BillingCustomerID))

		// schedule a job to check if the org is still on trial after 7 days
		j, err := riverutils.InsertOnlyRiverClient.Insert(ctx, &riverutils.TrialGracePeriodCheckArgs{OrgID: org.ID}, &river.InsertOpts{
			ScheduledAt: time.Now().AddDate(0, 0, gracePeriodDays).Add(time.Hour * 1), // add buffer of 1 hour to ensure the job runs after grace period days
			UniqueOpts: river.UniqueOpts{
				ByArgs: true,
			},
		})
		if err != nil {
			return fmt.Errorf("failed to schedule trial grace period check job: %w", err)
		}

		// insert billing error
		_, err = w.admin.DB.UpsertBillingError(ctx, &database.UpsertBillingErrorOptions{
			OrgID:              org.ID,
			Type:               database.BillingErrorTypeTrialEnded,
			Message:            fmt.Sprintf("Trial period ended, please visit the billing portal to upgrade your plan to continue using Rill. After %d days, your projects will be hibernated if still on trial.", gracePeriodDays),
			TriggersRiverJobID: j.Job.ID,
			EventTime:          sub[0].TrialEndDate.AddDate(0, 0, 1),
		})
		if err != nil {
			return fmt.Errorf("failed to add billing error: %w", err)
		}

		// send email
		err = w.admin.Email.SendInformational(&email.Informational{
			ToEmail: org.BillingEmail,
			ToName:  org.Name,
			Subject: "Your trial period has ended",
			Title:   "",
			Body:    template.HTML(fmt.Sprintf("Your trial period has ended, please visit the billing portal to enter payment method and upgrade your plan to continue using Rill. After %d days, your projects will be hibernated if you are still on trial.", gracePeriodDays)),
		})
		if err != nil {
			return fmt.Errorf("failed to send trial period ended email for org %q: %w", org.Name, err)
		}
		w.admin.Logger.Info("Email sent for trial period ended", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.String("billing_customer_id", org.BillingCustomerID))
	} else {
		w.admin.Logger.Warn("Trial period has not ended when check was run, please check the org manually", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.String("billing_customer_id", org.BillingCustomerID))
	}

	return nil
}

func NewTrialGracePeriodCheckWorker(adm *admin.Service) *TrialGracePeriodCheckWorker {
	return &TrialGracePeriodCheckWorker{admin: adm}
}

type TrialGracePeriodCheckWorker struct {
	river.WorkerDefaults[riverutils.TrialGracePeriodCheckArgs]
	admin *admin.Service
}

func (w *TrialGracePeriodCheckWorker) Work(ctx context.Context, job *river.Job[riverutils.TrialGracePeriodCheckArgs]) error {
	org, err := w.admin.DB.FindOrganization(ctx, job.Args.OrgID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			// org got deleted, ignore
			return nil
		}
		return fmt.Errorf("failed to find organization: %w", err)
	}

	// check if the org has any active subscription
	sub, err := w.admin.Biller.GetSubscriptionsForCustomer(ctx, org.BillingCustomerID)
	if err != nil {
		return fmt.Errorf("failed to get subscriptions for org %q: %w", org.Name, err)
	}

	if len(sub) == 0 {
		return fmt.Errorf("no active subscription found for the org %q", org.Name)
	}

	if len(sub) > 1 {
		return fmt.Errorf("multiple active subscriptions found for the org %q", org.Name)
	}

	if sub[0].ID != job.Args.SubID || sub[0].Plan.ID != job.Args.PlanID || sub[0].TrialEndDate.IsZero() {
		// subscription might have changed, ignore, delete the billing error if this was for this job
		be, err := w.admin.DB.FindBillingErrorByType(ctx, org.ID, database.BillingErrorTypeTrialEnded)
		if err != nil {
			if !errors.Is(err, database.ErrNotFound) {
				return fmt.Errorf("failed to find billing errors: %w", err)
			}
		}

		// check the jobID to handle the case when the customer was put on a new trial plan manually
		if be != nil && be.TriggersRiverJobID == job.ID {
			err = w.admin.DB.DeleteBillingError(ctx, be.ID)
			if err != nil {
				return fmt.Errorf("failed to delete billing error: %w", err)
			}
		}
		return nil
	}

	if sub[0].TrialEndDate.AddDate(0, 0, gracePeriodDays).Before(time.Now().UTC()) {
		// trial grace period has ended, log error, send email and hibernate projects
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
		w.admin.Logger.Info("Projects hibernated due to trial grace period ended", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.String("billing_customer_id", org.BillingCustomerID))

		// send email
		err = w.admin.Email.SendInformational(&email.Informational{
			ToEmail: org.BillingEmail,
			ToName:  org.Name,
			Subject: "Your trial grace period has ended",
			Title:   "",
			Body:    "Your trial grace period has ended, your projects have been hibernated. Please visit the billing portal to enter payment method and upgrade your plan to continue using Rill.",
		})
		if err != nil {
			return fmt.Errorf("failed to send trial grace period ended email for org %q: %w", org.Name, err)
		}
		w.admin.Logger.Info("Email sent for projects hibernated", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.String("billing_customer_id", org.BillingCustomerID))
	} else {
		w.admin.Logger.Warn("Trial grace period has not ended when check was run, please check the org manually", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.String("billing_customer_id", org.BillingCustomerID))
	}
	return nil
}

func NewInvoicePaymentFailedWorker(adm *admin.Service) *InvoicePaymentFailedWorker {
	return &InvoicePaymentFailedWorker{admin: adm}
}

type InvoicePaymentFailedWorker struct {
	river.WorkerDefaults[riverutils.InvoicePaymentFailedArgs]
	admin *admin.Service
}

func (w *InvoicePaymentFailedWorker) Work(ctx context.Context, job *river.Job[riverutils.InvoicePaymentFailedArgs]) error {
	org, err := w.admin.DB.FindOrganizationForBillingCustomerID(ctx, job.Args.BillingCustomerID)
	if err != nil {
		return fmt.Errorf("failed to find organization of billing customer id %q: %w", job.Args.BillingCustomerID, err)
	}

	// schedule a job to check if the invoice is paid after grace period days
	jobres, err := riverutils.InsertOnlyRiverClient.Insert(ctx, &riverutils.InvoicePaymentFailedGracePeriodCheckArgs{OrgID: org.ID, InvoiceID: job.Args.InvoiceID}, &river.InsertOpts{
		ScheduledAt: time.Now().AddDate(0, 0, gracePeriodDays).Add(time.Hour * 1), // add buffer of 1 hour to ensure the job runs after grace period days
		UniqueOpts: river.UniqueOpts{
			ByArgs: true,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to schedule invoice payment failed grace period check job: %w", err)
	}

	// insert billing error
	_, err = w.admin.DB.UpsertBillingError(ctx, &database.UpsertBillingErrorOptions{
		OrgID:              org.ID,
		Type:               database.BillingErrorTypeInvoicePaymentFailed,
		Message:            fmt.Sprintf("Invoice %s payment failed, please fix by visiting the billing portal.", job.Args.InvoiceID),
		TriggersRiverJobID: jobres.Job.ID,
		EventTime:          job.Args.FailedAt,
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
	w.admin.Logger.Info("Email sent for invoice payment failed", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.String("billing_customer_id", org.BillingCustomerID))

	return nil
}

func NewInvoicePaymentSuccessWorker(adm *admin.Service) *InvoicePaymentSuccessWorker {
	return &InvoicePaymentSuccessWorker{admin: adm}
}

type InvoicePaymentSuccessWorker struct {
	river.WorkerDefaults[riverutils.InvoicePaymentSuccessArgs]
	admin *admin.Service
}

func (w *InvoicePaymentSuccessWorker) Work(ctx context.Context, job *river.Job[riverutils.InvoicePaymentSuccessArgs]) error {
	org, err := w.admin.DB.FindOrganizationForBillingCustomerID(ctx, job.Args.BillingCustomerID)
	if err != nil {
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
		// remove any scheduled job for invoice payment failed grace period check
		if be.TriggersRiverJobID > 0 { // river job ids starts from 1
			_, err := riverutils.InsertOnlyRiverClient.JobCancel(ctx, be.TriggersRiverJobID)
			if err != nil {
				if !errors.Is(err, rivertype.ErrNotFound) {
					return fmt.Errorf("failed to cancel invoice payment failed grace period check job: %w", err)
				}
			}
		}
		err = w.admin.DB.DeleteBillingError(ctx, be.ID)
		if err != nil {
			return fmt.Errorf("failed to delete billing error: %w", err)
		}
	}

	return nil
}

func NewInvoicePaymentFailedGracePeriodCheckWorker(adm *admin.Service) *InvoicePaymentFailedGracePeriodCheckWorker {
	return &InvoicePaymentFailedGracePeriodCheckWorker{admin: adm}
}

type InvoicePaymentFailedGracePeriodCheckWorker struct {
	river.WorkerDefaults[riverutils.InvoicePaymentFailedGracePeriodCheckArgs]
	admin *admin.Service
}

func (w *InvoicePaymentFailedGracePeriodCheckWorker) Work(ctx context.Context, job *river.Job[riverutils.InvoicePaymentFailedGracePeriodCheckArgs]) error {
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

		if !w.admin.Biller.IsInvoiceValid(ctx, invoice) {
			w.admin.Logger.Warn("Invoice is invalid clearing billing error", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.String("billing_customer_id", org.BillingCustomerID))
			// invoice is invalid, delete the billing error
			err = w.admin.DB.DeleteBillingError(ctx, be.ID)
			if err != nil {
				return fmt.Errorf("failed to delete billing error: %w", err)
			}
			return nil
		}

		if w.admin.Biller.IsInvoicePaid(ctx, invoice) {
			w.admin.Logger.Warn("Invoice was already paid but billing error was not cleared", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.String("billing_customer_id", org.BillingCustomerID))
			// invoice is paid, delete the billing error
			err = w.admin.DB.DeleteBillingError(ctx, be.ID)
			if err != nil {
				return fmt.Errorf("failed to delete billing error: %w", err)
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
			Body:    "Your account does not have any payment method, your projects have been hibernated. Please visit the billing portal to enter payment method and then run `rill project reconcile` to unhibernate each project.",
		})
		if err != nil {
			return fmt.Errorf("failed to send payment method expired email for org %q: %w", org.Name, err)
		}
		w.admin.Logger.Info("Email sent for projects hibernated", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.String("billing_customer_id", org.BillingCustomerID))
	}

	return nil
}

func NewHandlePlanChangeByAPIWorker(adm *admin.Service) *HandlePlanChangeByAPIWorker {
	return &HandlePlanChangeByAPIWorker{admin: adm}
}

type HandlePlanChangeByAPIWorker struct {
	river.WorkerDefaults[riverutils.HandlePlanChangeByAPIArgs]
	admin *admin.Service
}

func (w *HandlePlanChangeByAPIWorker) Work(ctx context.Context, job *river.Job[riverutils.HandlePlanChangeByAPIArgs]) error {
	org, err := w.admin.DB.FindOrganization(ctx, job.Args.OrgID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			// org got deleted, ignore
			return nil
		}
		return fmt.Errorf("failed to find organization: %w", err)
	}

	// check if the org has any active subscription
	sub, err := w.admin.Biller.GetSubscriptionsForCustomer(ctx, org.BillingCustomerID)
	if err != nil {
		return fmt.Errorf("failed to get subscriptions for org %q: %w", org.Name, err)
	}

	if len(sub) == 0 {
		return fmt.Errorf("no active subscription found for the org %q", org.Name)
	}

	if len(sub) > 1 {
		return fmt.Errorf("multiple active subscriptions found for the org %q", org.Name)
	}

	if sub[0].ID != job.Args.SubID || sub[0].Plan.ID != job.Args.PlanID {
		// subscription or plan might have changed, ignore
		return nil
	}

	// if the new plan is still a trial plan, can happen if manually assigned new trial plan for example to extend trial period, if yes then schedule trail job checks
	if !sub[0].TrialEndDate.IsZero() {
		err = w.admin.ScheduleTrialEndCheckJobs(ctx, org.ID, sub[0].TrialEndDate)
		if err != nil {
			return fmt.Errorf("failed to schedule trial end check job: %w", err)
		}
	} else {
		// if new plan is paid then delete any trial related billing errors and warnings if any
		be, err := w.admin.DB.FindBillingErrorByType(ctx, org.ID, database.BillingErrorTypeTrialEnded)
		if err != nil {
			if !errors.Is(err, database.ErrNotFound) {
				return fmt.Errorf("failed to find billing errors: %w", err)
			}
		}

		if be != nil {
			err = w.admin.DB.DeleteBillingError(ctx, be.ID)
			if err != nil {
				return fmt.Errorf("failed to delete billing error: %w", err)
			}
		}
	}

	return nil
}

type ErrorHandler struct {
	Logger *zap.Logger
}

func (h *ErrorHandler) HandleError(ctx context.Context, job *rivertype.JobRow, err error) *river.ErrorHandlerResult {
	var args string
	_ = json.Unmarshal(job.EncodedArgs, &args) // ignore errors
	h.Logger.Error("Job errored", zap.Int64("job_id", job.ID), zap.Int("num_attempt", job.Attempt), zap.String("kind", job.Kind), zap.String("args", args), zap.Error(err))
	return nil
}

func (h *ErrorHandler) HandlePanic(ctx context.Context, job *rivertype.JobRow, panicVal any, trace string) *river.ErrorHandlerResult {
	var args string
	_ = json.Unmarshal(job.EncodedArgs, &args) // ignore errors
	h.Logger.Error("Job panicked", zap.Int64("job_id", job.ID), zap.String("kind", job.Kind), zap.String("args", args), zap.Any("panic_val", panicVal), zap.String("trace", trace))
	// Set the job to be immediately cancelled.
	return &river.ErrorHandlerResult{SetCancelled: true}
}
