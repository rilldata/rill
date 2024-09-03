package riverutils

import (
	"database/sql"
	"time"

	"github.com/riverqueue/river"
)

// InsertOnlyRiverClient is a river client that only supports inserting jobs, actual river worker is started in the worker service
var InsertOnlyRiverClient *river.Client[*sql.Tx]

type PaymentMethodAddedArgs struct {
	PaymentMethodID   string
	PaymentCustomerID string
	PaymentType       string
	EventTime         time.Time
}

func (PaymentMethodAddedArgs) Kind() string { return "payment_method_added" }

type PaymentMethodRemovedArgs struct {
	PaymentMethodID   string
	PaymentCustomerID string
	EventTime         time.Time
}

func (PaymentMethodRemovedArgs) Kind() string { return "payment_method_removed" }

type CustomerAddressUpdatedArgs struct {
	PaymentCustomerID string
	EventTime         time.Time
}

func (CustomerAddressUpdatedArgs) Kind() string { return "customer_address_updated" }

type TrialEndingSoonArgs struct {
	OrgID  string
	SubID  string
	PlanID string
}

func (TrialEndingSoonArgs) Kind() string { return "trial_ending_soon" }

type TrialEndCheckArgs struct {
	OrgID  string
	SubID  string
	PlanID string
}

func (TrialEndCheckArgs) Kind() string { return "trial_end_check" }

type TrialGracePeriodCheckArgs struct {
	OrgID  string
	SubID  string
	PlanID string
}

func (TrialGracePeriodCheckArgs) Kind() string { return "trial_grace_period_check" }

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

type InvoicePaymentSuccessArgs struct {
	BillingCustomerID string
	InvoiceID         string
}

func (InvoicePaymentSuccessArgs) Kind() string { return "invoice_payment_success" }

type InvoicePaymentFailedGracePeriodCheckArgs struct {
	OrgID     string
	InvoiceID string
}

func (InvoicePaymentFailedGracePeriodCheckArgs) Kind() string {
	return "invoice_payment_failed_grace_period_check"
}

type HandlePlanChangeByAPIArgs struct {
	OrgID  string
	SubID  string
	PlanID string
}

func (HandlePlanChangeByAPIArgs) Kind() string { return "handle_plan_change_by_api" }

type HandleSubscriptionCancellationArgs struct {
	OrgID  string
	SubID  string
	PlanID string
}

func (HandleSubscriptionCancellationArgs) Kind() string { return "handle_subscription_cancellation" }
