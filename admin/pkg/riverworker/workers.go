package riverworker

import (
	"context"
	"encoding/json"
	"fmt"
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

func NewAddBillingErrorWorker(adm *admin.Service) *AddBillingErrorWorker {
	return &AddBillingErrorWorker{admin: adm}
}

// AddBillingErrorWorker worker to add billing error of payment failed in the billing_error table for an org
type AddBillingErrorWorker struct {
	river.WorkerDefaults[riverutils.AddBillingErrorArgs]
	admin *admin.Service
}

func (w *AddBillingErrorWorker) Work(ctx context.Context, job *river.Job[riverutils.AddBillingErrorArgs]) error {
	org, err := w.admin.DB.FindOrganizationForPaymentCustomerID(ctx, job.Args.CustomerID)
	if err != nil {
		return fmt.Errorf("failed to find organization for payment customer id: %w", err)
	}
	message := "Recent payment failed, please fix your payment method by visiting the billing portal."
	// check if there is charge id and amount in the metadata and add it to the message
	if chargeID, ok := job.Args.Metadata["charge_id"]; ok {
		message += fmt.Sprintf(" Charge ID: %s", chargeID)
	}
	if amount, ok := job.Args.Metadata["amount"]; ok {
		message += fmt.Sprintf(" Amount: %s", amount)
	}

	_, err = w.admin.DB.InsertBillingError(ctx, &database.InsertBillingErrorOptions{
		OrgID:   org.ID,
		Type:    job.Args.ErrorType,
		Message: message,
	})
	if err != nil {
		return fmt.Errorf("failed to add billing error: %w", err)
	}
	return nil
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
	// check for billing errors and delete them
	errs, err := w.admin.DB.FindBillingErrorsByType(ctx, org.ID, database.BillingErrorTypePaymentFailed)
	if err != nil {
		return fmt.Errorf("failed to find billing errors: %w", err)
	}
	// for now delete all the payment failed errors, this should be ok as we don't expect payment failures across subsequent billing cycles, and we don't insert duplicates
	// but if we want to be more specific then we can use charge amount to compare and delete
	for _, e := range errs {
		err = w.admin.DB.DeleteBillingError(ctx, e.ID)
		if err != nil {
			return fmt.Errorf("failed to delete billing error: %w", err)
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
