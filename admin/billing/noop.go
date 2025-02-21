package billing

import (
	"context"
	"time"

	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/jobs"
	"github.com/rilldata/rill/runtime/pkg/httputil"
)

var _ Biller = &noop{}

type noop struct{}

func NewNoop() Biller {
	return noop{}
}

func (n noop) Name() string {
	return "noop"
}

func (n noop) GetDefaultPlan(ctx context.Context) (*Plan, error) {
	return &Plan{Quotas: Quotas{}}, nil
}

func (n noop) GetPlans(ctx context.Context) ([]*Plan, error) {
	return nil, nil
}

func (n noop) GetPlan(ctx context.Context, id string) (*Plan, error) {
	return nil, nil
}

func (n noop) GetPlanByName(ctx context.Context, name string) (*Plan, error) {
	return nil, nil
}

func (n noop) GetPlanTypeForExternalId(externalID string) PlanType {
	return TrailPlanType
}

func (n noop) GetPublicPlans(ctx context.Context) ([]*Plan, error) {
	return nil, nil
}

func (n noop) CreateCustomer(ctx context.Context, organization *database.Organization, provider PaymentProvider) (*Customer, error) {
	return &Customer{}, nil
}

func (n noop) FindCustomer(ctx context.Context, customerID string) (*Customer, error) {
	return &Customer{}, nil
}

func (n noop) UpdateCustomerPaymentID(ctx context.Context, customerID string, provider PaymentProvider, paymentProviderID string) error {
	return nil
}

func (n noop) UpdateCustomerEmail(ctx context.Context, customerID, email string) error {
	return nil
}

func (n noop) DeleteCustomer(ctx context.Context, customerID string) error {
	return nil
}

func (n noop) CreateSubscription(ctx context.Context, customerID string, plan *Plan) (*Subscription, error) {
	return &Subscription{Customer: &Customer{}, Plan: &Plan{Quotas: Quotas{}}}, nil
}

func (n noop) GetActiveSubscription(ctx context.Context, customerID string) (*Subscription, error) {
	return &Subscription{Customer: &Customer{}, Plan: &Plan{Quotas: Quotas{}}}, nil
}

func (n noop) ChangeSubscriptionPlan(ctx context.Context, subscriptionID string, plan *Plan) (*Subscription, error) {
	return &Subscription{Customer: &Customer{}, Plan: &Plan{Quotas: Quotas{}}}, nil
}

func (n noop) UnscheduleCancellation(ctx context.Context, subscriptionID string) (*Subscription, error) {
	return &Subscription{Customer: &Customer{}, Plan: &Plan{Quotas: Quotas{}}}, nil
}

func (n noop) CancelSubscriptionsForCustomer(ctx context.Context, customerID string, cancelOption SubscriptionCancellationOption) (time.Time, error) {
	return time.Time{}, nil
}

func (n noop) GetInvoice(ctx context.Context, invoiceID string) (*Invoice, error) {
	return nil, nil
}

func (n noop) IsInvoiceValid(ctx context.Context, invoice *Invoice) bool {
	return true
}

func (n noop) IsInvoicePaid(ctx context.Context, invoice *Invoice) bool {
	return true
}

func (n noop) MarkCustomerTaxExempt(ctx context.Context, customerID string) error {
	return nil
}

func (n noop) UnmarkCustomerTaxExempt(ctx context.Context, customerID string) error {
	return nil
}

func (n noop) ReportUsage(ctx context.Context, usage []*Usage) error {
	return nil
}

func (n noop) GetReportingGranularity() UsageReportingGranularity {
	return UsageReportingGranularityNone
}

func (n noop) GetReportingWorkerCron() string {
	return ""
}

func (n noop) WebhookHandlerFunc(ctx context.Context, jc jobs.Client) httputil.Handler {
	return nil
}
