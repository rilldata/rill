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

func (n noop) GetDefaultPlan(ctx context.Context) (*Plan, error) {
	return nil, nil
}

func (n noop) GetPlans(ctx context.Context) ([]*Plan, error) {
	return nil, nil
}

func (n noop) GetPlan(ctx context.Context, rillPlanID, billerPlanID string) (*Plan, error) {
	return nil, nil
}

func (n noop) GetPublicPlans(ctx context.Context) ([]*Plan, error) {
	return nil, nil
}

func (n noop) CreateCustomer(ctx context.Context, organization *database.Organization) (string, error) {
	return "", nil
}

func (n noop) CreateSubscription(ctx context.Context, customerID string, plan *Plan) (*Subscription, error) {
	return nil, nil
}

func (n noop) CancelSubscription(ctx context.Context, subscriptionID string, cancelOption SubscriptionCancellationOption) error {
	return nil
}

func (n noop) GetSubscriptionsForCustomer(ctx context.Context, customerID string) ([]*Subscription, error) {
	return nil, nil
}

func (n noop) ChangeSubscriptionPlan(ctx context.Context, subscriptionID string, plan *Plan, changeOption SubscriptionChangeOption) (*Subscription, error) {
	return nil, nil
}

func (n noop) CancelSubscriptionsForCustomer(ctx context.Context, customerID string, cancelOption SubscriptionCancellationOption) error {
	return nil
}

func (n noop) ReportUsage(ctx context.Context, customerID string, usage []*Usage) error {
	return nil
}

func (n noop) GetReportingGranularity() UsageReportingGranularity {
	return UsageReportingGranularityNone
}

func (n noop) GetReportingWorkerCron() string {
	return ""
}
