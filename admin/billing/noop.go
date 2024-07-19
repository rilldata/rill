package billing

import (
	"context"

	"github.com/rilldata/rill/admin/database"
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

func (n noop) CreateSubscription(ctx context.Context, customerID string, plan *Plan) (*Subscription, error) {
	return &Subscription{Customer: &Customer{}, Plan: &Plan{Quotas: Quotas{}}}, nil
}

func (n noop) CancelSubscription(ctx context.Context, subscriptionID string, cancelOption SubscriptionCancellationOption) error {
	return nil
}

func (n noop) GetSubscriptionsForCustomer(ctx context.Context, customerID string) ([]*Subscription, error) {
	return []*Subscription{{Customer: &Customer{}, Plan: &Plan{Quotas: Quotas{}}}}, nil
}

func (n noop) ChangeSubscriptionPlan(ctx context.Context, subscriptionID string, plan *Plan) (*Subscription, error) {
	return &Subscription{Customer: &Customer{}, Plan: &Plan{Quotas: Quotas{}}}, nil
}

func (n noop) CancelSubscriptionsForCustomer(ctx context.Context, customerID string, cancelOption SubscriptionCancellationOption) error {
	return nil
}

func (n noop) FindSubscriptionsPastTrialPeriod(ctx context.Context) ([]*Subscription, error) {
	return []*Subscription{}, nil
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
