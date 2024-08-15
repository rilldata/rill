package riverworker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/pkg/riverworker/riverutils"
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
	river.WorkerDefaults[riverutils.PaymentMethodAdded]
	admin *admin.Service
}

func (w *PaymentMethodAddedWorker) Work(ctx context.Context, job *river.Job[riverutils.PaymentMethodAdded]) error {
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
	river.WorkerDefaults[riverutils.PaymentMethodRemoved]
	admin *admin.Service
}

func (w *PaymentMethodRemovedWorker) Work(ctx context.Context, job *river.Job[riverutils.PaymentMethodRemoved]) error {
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
		OrgID:   org.ID,
		Type:    database.BillingErrorTypeNoPaymentMethod,
		Message: "No payment method attached, please add a payment method by visiting the billing portal",
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
