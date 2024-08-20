package riverworker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	org, err := w.admin.DB.FindOrganizationForPaymentCustomerID(ctx, job.Args.CustomerID)
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
	org, err := w.admin.DB.FindOrganizationForPaymentCustomerID(ctx, job.Args.CustomerID)
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
	org, err := w.admin.DB.FindOrganizationForPaymentCustomerID(ctx, job.Args.CustomerID)
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
	org, err := w.admin.DB.FindOrganizationForPaymentCustomerID(ctx, job.Args.CustomerID)
	if err != nil {
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

	if sub[0].Plan.TrialPeriodDays > 0 && time.Now().UTC().After(sub[0].StartDate.AddDate(0, 0, sub[0].Plan.TrialPeriodDays)) {
		// trial period has ended, log error, send email and schedule a job to hibernate projects after 7 days if still on trial
		w.admin.Logger.Error("Trial period has ended", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.String("billing_customer_id", org.BillingCustomerID))

		// schedule a job to check if the org is still on trial after 7 days
		_, err := riverutils.InsertOnlyRiverClient.Insert(ctx, &riverutils.TrialGracePeriodCheckArgs{OrgID: org.ID}, &river.InsertOpts{
			ScheduledAt: time.Now().AddDate(0, 0, 7).Add(time.Hour * 1), // add buffer of 1 hour to ensure the job runs after 7 days
			UniqueOpts: river.UniqueOpts{
				ByArgs: true,
			},
		})
		if err != nil {
			return fmt.Errorf("failed to schedule trial grace period check job: %w", err)
		}
		w.admin.Logger.Info("Scheduled trial grace period check job", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.String("billing_customer_id", org.BillingCustomerID))

		// send email
		err = w.admin.Email.SendInformational(&email.Informational{
			ToEmail: org.BillingEmail,
			ToName:  org.Name,
			Subject: "Your trial period has ended",
			Title:   "",
			Body:    "Your trial period has ended, please visit the billing portal to enter payment method and upgrade your plan to continue using Rill. After 7 days, your projects will be hibernated if you are still on trial.",
		})
		if err != nil {
			return fmt.Errorf("failed to send trial period ended email for org %q: %w", org.Name, err)
		}
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

	if sub[0].Plan.TrialPeriodDays > 0 && time.Now().UTC().After(sub[0].StartDate.AddDate(0, 0, sub[0].Plan.TrialPeriodDays+7)) {
		// trial grace period has ended, log error, send email and hibernate projects
		w.admin.Logger.Error("Trial grace period has ended", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.String("billing_customer_id", org.BillingCustomerID))
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
