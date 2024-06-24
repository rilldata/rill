package billing

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/eapache/go-resiliency/retrier"
	"github.com/mitchellh/mapstructure"
	"github.com/orbcorp/orb-go"
	"github.com/orbcorp/orb-go/option"
	"github.com/rilldata/rill/admin/database"
)

const (
	requestMaxLimit = 500
	requestTimeout  = 10 * time.Second
)

var _ Biller = &Orb{}

type Orb struct {
	client *orb.Client
}

func NewOrb(orbKey string) Biller {
	c := orb.NewClient(option.WithAPIKey(orbKey), option.WithRequestTimeout(requestTimeout))

	return &Orb{client: c}
}

func (o *Orb) Name() string {
	return "orb"
}

func (o *Orb) GetDefaultPlan(ctx context.Context) (*Plan, error) {
	plans, err := o.GetPlans(ctx)
	if err != nil {
		return nil, err
	}
	for _, p := range plans {
		if p.Default {
			return p, nil
		}
	}
	return nil, ErrNotFound
}

func (o *Orb) GetPlans(ctx context.Context) ([]*Plan, error) {
	return o.getAllPlans(ctx)
}

func (o *Orb) GetPublicPlans(ctx context.Context) ([]*Plan, error) {
	all, err := o.getAllPlans(ctx)
	if err != nil {
		return nil, err
	}
	var publicPlans []*Plan
	for _, p := range all {
		if p.Public {
			publicPlans = append(publicPlans, p)
		}
	}
	return publicPlans, nil
}

func (o *Orb) GetPlan(ctx context.Context, id string) (*Plan, error) {
	plans, err := o.getAllPlans(ctx)
	if err != nil {
		return nil, err
	}
	for _, p := range plans {
		if p.ID == id {
			return p, nil
		}
	}
	return nil, ErrNotFound
}

func (o *Orb) GetPlanByName(ctx context.Context, name string) (*Plan, error) {
	plans, err := o.getAllPlans(ctx)
	if err != nil {
		return nil, err
	}
	for _, p := range plans {
		if p.Name == name {
			return p, nil
		}
	}
	return nil, ErrNotFound
}

func (o *Orb) CreateCustomer(ctx context.Context, organization *database.Organization) (string, error) {
	customer, err := o.client.Customers.New(ctx, orb.CustomerNewParams{
		Email:              orb.String(SupportEmail), // TODO use creators email or capture organization billing email
		Name:               orb.String(organization.Name),
		ExternalCustomerID: orb.String(organization.ID),
		Timezone:           orb.String(DefaultTimeZone),
	})
	if err != nil {
		return "", err
	}

	return customer.ExternalCustomerID, nil
}

func (o *Orb) CreateSubscription(ctx context.Context, customerID string, plan *Plan) (*Subscription, error) {
	sub, err := o.client.Subscriptions.New(ctx, orb.SubscriptionNewParams{
		ExternalCustomerID: orb.String(customerID),
		PlanID:             orb.String(plan.ID),
	})
	if err != nil {
		return nil, err
	}

	return &Subscription{
		ID:                           sub.ID,
		CustomerID:                   sub.Customer.ExternalCustomerID,
		Plan:                         plan,
		StartDate:                    sub.StartDate,
		EndDate:                      sub.EndDate,
		CurrentBillingCycleStartDate: sub.CurrentBillingPeriodStartDate,
		CurrentBillingCycleEndDate:   sub.CurrentBillingPeriodEndDate,
		TrialEndDate:                 sub.TrialInfo.EndDate,
		Metadata:                     sub.Metadata,
	}, nil
}

func (o *Orb) GetSubscriptionsForCustomer(ctx context.Context, customerID string) ([]*Subscription, error) {
	sub, err := o.client.Subscriptions.List(ctx, orb.SubscriptionListParams{
		ExternalCustomerID: orb.String(customerID),
		Status:             orb.F(orb.SubscriptionListParamsStatusActive),
	})
	if err != nil {
		return nil, err
	}

	var subscriptions []*Subscription
	for i := 0; i < len(sub.Data); i++ {
		s := sub.Data[i]
		plan, err := getBillingPlanFromOrbPlan(&s.Plan)
		if err != nil {
			return nil, err
		}

		subscriptions = append(subscriptions, &Subscription{
			ID:                           s.ID,
			CustomerID:                   s.Customer.ExternalCustomerID,
			Plan:                         plan,
			StartDate:                    s.StartDate,
			EndDate:                      s.EndDate,
			CurrentBillingCycleStartDate: s.CurrentBillingPeriodStartDate,
			CurrentBillingCycleEndDate:   s.CurrentBillingPeriodEndDate,
			TrialEndDate:                 s.TrialInfo.EndDate,
			Metadata:                     s.Metadata,
		})
	}
	return subscriptions, nil
}

func (o *Orb) ChangeSubscriptionPlan(ctx context.Context, subscriptionID string, plan *Plan) (*Subscription, error) {
	s, err := o.client.Subscriptions.SchedulePlanChange(ctx, subscriptionID, orb.SubscriptionSchedulePlanChangeParams{
		PlanID:       orb.String(plan.ID),
		ChangeOption: orb.F(orb.SubscriptionSchedulePlanChangeParamsChangeOptionImmediate),
	})
	if err != nil {
		return nil, err
	}
	return &Subscription{
		ID:                           s.ID,
		CustomerID:                   s.Customer.ExternalCustomerID,
		Plan:                         plan,
		StartDate:                    s.StartDate,
		EndDate:                      s.EndDate,
		CurrentBillingCycleStartDate: s.CurrentBillingPeriodStartDate,
		CurrentBillingCycleEndDate:   s.CurrentBillingPeriodEndDate,
		TrialEndDate:                 s.TrialInfo.EndDate,
		Metadata:                     s.Metadata,
	}, nil
}

func (o *Orb) CancelSubscription(ctx context.Context, subscriptionID string, cancelOption SubscriptionCancellationOption) error {
	var cancelParams orb.SubscriptionCancelParams
	switch cancelOption {
	case SubscriptionCancellationOptionEndOfSubscriptionTerm:
		cancelParams = orb.SubscriptionCancelParams{
			CancelOption: orb.F(orb.SubscriptionCancelParamsCancelOptionEndOfSubscriptionTerm),
		}
	case SubscriptionCancellationOptionImmediate:
		cancelParams = orb.SubscriptionCancelParams{
			CancelOption: orb.F(orb.SubscriptionCancelParamsCancelOptionImmediate),
		}
	}

	_, err := o.client.Subscriptions.Cancel(ctx, subscriptionID, cancelParams)
	if err != nil {
		return err
	}
	return nil
}

func (o *Orb) CancelSubscriptionsForCustomer(ctx context.Context, customerID string, cancelOption SubscriptionCancellationOption) error {
	var cancelParams orb.SubscriptionCancelParams
	switch cancelOption {
	case SubscriptionCancellationOptionEndOfSubscriptionTerm:
		cancelParams = orb.SubscriptionCancelParams{
			CancelOption: orb.F(orb.SubscriptionCancelParamsCancelOptionEndOfSubscriptionTerm),
		}
	case SubscriptionCancellationOptionImmediate:
		cancelParams = orb.SubscriptionCancelParams{
			CancelOption: orb.F(orb.SubscriptionCancelParamsCancelOptionImmediate),
		}
	}

	subs, err := o.GetSubscriptionsForCustomer(ctx, customerID)
	if err != nil {
		return err
	}
	for _, s := range subs {
		_, err := o.client.Subscriptions.Cancel(ctx, s.ID, cancelParams)
		if err != nil {
			return err
		}
	}
	return nil
}

func (o *Orb) ReportUsage(ctx context.Context, usage []*Usage) error {
	var orbUsage []orb.EventIngestParamsEvent
	// sync max 500 events at a time
	for _, u := range usage {
		eventName := u.MetricName + "_" + string(u.ReportingGrain)
		// use end time minus 1 second to make sure the event is attributed to the current time bucket
		eventTime := u.EndTime.Add(-1 * time.Second)
		// generate idempotency key using customer id, timestamp, event name and metadata
		idempotencyKey := fmt.Sprintf("%s_%d_%s_%v", u.CustomerID, eventTime.UnixMilli(), eventName, u.Metadata)

		props := make(map[string]interface{}, len(u.Metadata)+1)
		for k, v := range u.Metadata {
			props[k] = v
		}
		props["amount"] = u.Amount

		orbUsage = append(orbUsage, orb.EventIngestParamsEvent{
			ExternalCustomerID: orb.String(u.CustomerID),
			EventName:          orb.String(eventName),
			IdempotencyKey:     orb.String(idempotencyKey),
			Timestamp:          orb.F(eventTime),
			Properties:         orb.F[any](props),
		})

		if len(orbUsage) == requestMaxLimit {
			err := o.pushUsage(ctx, &orbUsage)
			if err != nil {
				return err
			}
			orbUsage = nil
		}
	}

	if len(orbUsage) > 0 {
		err := o.pushUsage(ctx, &orbUsage)
		if err != nil {
			return err
		}
	}
	return nil
}

func (o *Orb) pushUsage(ctx context.Context, usage *[]orb.EventIngestParamsEvent) error {
	re := retrier.New(retrier.ExponentialBackoff(5, 500*time.Millisecond), retryErrClassifier{})
	err := re.RunCtx(ctx, func(ctx context.Context) error {
		resp, err := o.client.Events.Ingest(ctx, orb.EventIngestParams{
			Events: orb.F(*usage),
		})
		if err != nil {
			return err
		}
		if len(resp.ValidationFailed) > 0 {
			errMsg := fmt.Sprintf("validation failure for %d events, showing first few:", len(resp.ValidationFailed))
			for i := 0; i < 5 && i < len(resp.ValidationFailed); i++ {
				errMsg += fmt.Sprintf("\n%s: %s", resp.ValidationFailed[i].IdempotencyKey, resp.ValidationFailed[i].ValidationErrors)
			}
			return errors.New(errMsg)
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (o *Orb) GetReportingGranularity() UsageReportingGranularity {
	return UsageReportingGranularityHour
}

func (o *Orb) GetReportingWorkerCron() string {
	// run every hour at around end of the hour
	return "55 * * * *"
}

func (o *Orb) getAllPlans(ctx context.Context) ([]*Plan, error) {
	plans, err := o.client.Plans.List(ctx, orb.PlanListParams{
		Limit:  orb.Int(requestMaxLimit), // TODO handle pagination, for now don't expect more than 500 plans
		Status: orb.F(orb.PlanListParamsStatusActive),
	})
	if err != nil {
		return nil, err
	}

	var billingPlans []*Plan
	for i := 0; i < len(plans.Data); i++ {
		billingPlan, err := getBillingPlanFromOrbPlan(&plans.Data[i])
		if err != nil {
			return nil, err
		}
		billingPlans = append(billingPlans, billingPlan)
	}
	return billingPlans, nil
}

func getBillingPlanFromOrbPlan(p *orb.Plan) (*Plan, error) {
	metadata := &planMetadata{}
	err := mapstructure.WeakDecode(p.Metadata, metadata)
	if err != nil {
		return nil, err
	}

	q := &Quotas{
		StorageLimitBytesPerDeployment: metadata.StorageLimitBytesPerDeployment,
		NumProjects:                    metadata.NumProjects,
		NumDeployments:                 metadata.NumDeployments,
		NumSlotsTotal:                  metadata.NumSlotsTotal,
		NumSlotsPerDeployment:          metadata.NumSlotsPerDeployment,
		NumOutstandingInvites:          metadata.NumOutstandingInvites,
	}

	trialPeriodDays := 0
	if p.TrialConfig.TrialPeriodUnit == orb.PlanTrialConfigTrialPeriodUnitDays {
		trialPeriodDays = int(p.TrialConfig.TrialPeriod)
	}

	billingPlan := &Plan{
		ID:              p.ID,
		Name:            p.ExternalPlanID,
		DisplayName:     p.Name,
		Description:     p.Description,
		TrialPeriodDays: trialPeriodDays,
		Default:         metadata.Default,
		Public:          metadata.Public,
		Quotas:          *q,
		Metadata:        p.Metadata,
	}
	return billingPlan, nil
}

// retryErrClassifier classifies 429 and 500 errors as retryable and all other errors as non retryable
type retryErrClassifier struct{}

func (retryErrClassifier) Classify(err error) retrier.Action {
	if err == nil {
		return retrier.Succeed
	}

	var orbErr *orb.Error
	if errors.As(err, &orbErr) {
		if orbErr.Status == orb.ErrorStatus500 || orbErr.Status == orb.ErrorStatus429 {
			return retrier.Retry
		}
	} else {
		return retrier.Fail
	}

	return retrier.Fail
}
