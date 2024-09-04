package jobs

import (
	"context"
	"time"
)

type Client interface {
	Close(ctx context.Context) error
	Work(ctx context.Context) error
	CancelJob(ctx context.Context, jobID int64) error

	// NOTE: Add new job trigger functions here
	ResetAllDeployments(ctx context.Context) (*InsertResult, error)

	// payment provider related workers
	PaymentMethodAdded(ctx context.Context, methodID, paymentCustomerID, typ string, eventTime time.Time) (*InsertResult, error)
	PaymentMethodRemoved(ctx context.Context, methodID, paymentCustomerID string, eventTime time.Time) (*InsertResult, error)
	CustomerAddressUpdated(ctx context.Context, paymentCustomerID string, eventTime time.Time) (*InsertResult, error)

	// biller related workers
	InvoicePaymentFailed(ctx context.Context, billingCustomerID, invoiceID, invoiceNumber, invoiceURL, amount, currency string, dueDate, failedAt time.Time) (*InsertResult, error)
	InvoicePaymentSuccess(ctx context.Context, billingCustomerID, invoiceID string) (*InsertResult, error)
	InvoicePaymentFailedGracePeriodCheck(ctx context.Context, orgID, invoiceID string, gracePeriodEndDate time.Time) (*InsertResult, error)

	// trial checks worker
	TrialEndingSoon(ctx context.Context, orgID, subID, planID string, trialEndDate time.Time) (*InsertResult, error)
	TrialEndCheck(ctx context.Context, orgID, subID, planID string, trialEndDate time.Time) (*InsertResult, error)
	TrialGracePeriodCheck(ctx context.Context, orgID, subID, planID string, gracePeriodEndDate time.Time) (*InsertResult, error)

	// subscription related workers
	PlanChangeByAPI(ctx context.Context, orgID, subID, planID string) (*InsertResult, error)
	SubscriptionCancellation(ctx context.Context, orgID, subID, planID string, subEndDate time.Time) (*InsertResult, error)
}

type InsertResult struct {
	ID        int64
	Duplicate bool
}
