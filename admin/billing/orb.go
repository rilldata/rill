package billing

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/eapache/go-resiliency/retrier"
	"github.com/orbcorp/orb-go"
	"github.com/orbcorp/orb-go/option"
	"github.com/rilldata/rill/admin/database"
)

const requestTimeout = 10 * time.Second

var ErrNotFound = errors.New("not found")

var _ Biller = &Orb{}

type Orb struct {
	client *orb.Client
}

func NewOrb(orbKey string) Biller {
	c := orb.NewClient(option.WithAPIKey(orbKey), option.WithRequestTimeout(requestTimeout))

	return &Orb{client: c}
}

func (o *Orb) GetDefaultPlan(ctx context.Context) (*Plan, error) {
	plans, err := o.GetPlans(ctx)
	if err != nil {
		return nil, err
	}
	for _, p := range plans {
		if strings.EqualFold(p.RillID, DefaultPlanID) {
			return p, nil
		}

		if v, ok := p.Metadata["default"]; ok {
			def, err := strconv.ParseBool(v)
			if err != nil {
				return nil, err
			}
			if def {
				return p, nil
			}
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
		if v, ok := p.Metadata["public"]; ok {
			public, err := strconv.ParseBool(v)
			if err != nil {
				return nil, err
			}
			if public {
				publicPlans = append(publicPlans, p)
			}
		}
	}
	return publicPlans, nil
}

func (o *Orb) GetPlan(ctx context.Context, rillPlanID, billerPlanID string) (*Plan, error) {
	plans, err := o.getAllPlans(ctx)
	if err != nil {
		return nil, err
	}
	for _, p := range plans {
		if strings.EqualFold(p.RillID, rillPlanID) || strings.EqualFold(p.BillerID, billerPlanID) {
			return p, nil
		}
	}
	return nil, ErrNotFound
}

func (o *Orb) CreateCustomer(ctx context.Context, organization *database.Organization) (string, error) {
	customer, err := o.client.Customers.New(ctx, orb.CustomerNewParams{
		Email:              orb.String(SupportEmail),
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
		PlanID:             orb.String(plan.BillerID),
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
	s, err := o.client.Subscriptions.SchedulePlanChange(ctx, subscriptionID, orb.SubscriptionSchedulePlanChangeParams{ChangeOption: orb.F(orb.SubscriptionSchedulePlanChangeParamsChangeOptionImmediate)})
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

func (o *Orb) ReportUsage(ctx context.Context, customerID string, usage []*Usage) error {
	if len(usage) == 0 {
		return nil
	}

	var orbUsage []orb.EventIngestParamsEvent
	for _, u := range usage {
		eventName := u.MetricName + "_" + string(u.ReportingGrain)
		// use end time minus 1 second to make sure the event is attributed to the current time bucket
		eventTime := u.EndTime.Add(-1 * time.Second)
		// generate idempotency key using customer id, timestamp, event name and metadata
		idempotencyKey := fmt.Sprintf("%s_%d_%s_%v", customerID, eventTime.UnixMilli(), eventName, u.Metadata)

		if u.Metadata == nil {
			u.Metadata = make(map[string]interface{})
		}
		u.Metadata["amount"] = u.Amount

		orbUsage = append(orbUsage, orb.EventIngestParamsEvent{
			ExternalCustomerID: orb.String(customerID),
			EventName:          orb.String(eventName),
			IdempotencyKey:     orb.String(idempotencyKey),
			Timestamp:          orb.F(eventTime),
			Properties:         orb.F[any](u.Metadata),
		})
	}

	re := retrier.New(retrier.ExponentialBackoff(5, 500*time.Millisecond), retryErrClassifier{})
	err := re.RunCtx(ctx, func(ctx context.Context) error {
		var err error
		_, err = o.client.Events.Ingest(ctx, orb.EventIngestParams{
			Events: orb.F(orbUsage),
		})
		return err
	})
	if err != nil {
		// TODO if err status is 400 the response contains validation errors, check if this err is logged properly
		return err
	}
	return nil
}

func (o *Orb) GetReportingGranularity() UsageReportingGranularity {
	return UsageReportingGranularityHour
}

func (o *Orb) GetReportingWorkerCron() string {
	// run every hour in middle of the hour
	return "30 * * * *"
}

func (o *Orb) getAllPlans(ctx context.Context) ([]*Plan, error) {
	plans, _ := o.client.Plans.List(ctx, orb.PlanListParams{
		Limit:  orb.Int(1000), // TODO handle pagination, for now don't expect more than 1000 plans
		Status: orb.F(orb.PlanListParamsStatusActive),
	})

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
	// Convert orb.Plan to billing.Plan
	q := &Quotas{}
	v, ok := p.Metadata["storage_limit_bytes_per_deployment"]
	if ok {
		m, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, err
		}
		q.StorageLimitBytesPerDeployment = &m
	}

	v, ok = p.Metadata["num_projects"]
	if ok {
		m, err := strconv.Atoi(v)
		if err != nil {
			return nil, err
		}
		q.NumProjects = &m
	}

	v, ok = p.Metadata["num_deployments"]
	if ok {
		m, err := strconv.Atoi(v)
		if err != nil {
			return nil, err
		}
		q.NumDeployments = &m
	}

	v, ok = p.Metadata["num_slots_total"]
	if ok {
		m, err := strconv.Atoi(v)
		if err != nil {
			return nil, err
		}
		q.NumSlotsTotal = &m
	}

	v, ok = p.Metadata["num_slots_per_deployment"]
	if ok {
		m, err := strconv.Atoi(v)
		if err != nil {
			return nil, err
		}
		q.NumSlotsPerDeployment = &m
	}

	v, ok = p.Metadata["num_outstanding_invites"]
	if ok {
		m, err := strconv.Atoi(v)
		if err != nil {
			return nil, err
		}
		q.NumOutstandingInvites = &m
	}

	v, ok = p.Metadata["reportable_metrics"]
	var reportableMetrics []string
	if ok {
		reportableMetrics = strings.Split(v, ",")
		// trim spaces, remove empty strings, and convert to lower case
		for i := 0; i < len(reportableMetrics); i++ {
			m := strings.TrimSpace(reportableMetrics[i])
			if m == "" {
				continue
			}
			reportableMetrics[i] = strings.ToLower(m)
		}
	}

	trialPeriodDays := 0
	if p.TrialConfig.TrialPeriodUnit == orb.PlanTrialConfigTrialPeriodUnitDays {
		trialPeriodDays = int(p.TrialConfig.TrialPeriod)
	}

	billingPlan := &Plan{
		BillerID:          p.ID,
		RillID:            p.ExternalPlanID,
		Name:              p.Name,
		Description:       p.Description,
		TrialPeriodDays:   trialPeriodDays,
		Quotas:            *q,
		ReportableMetrics: reportableMetrics,
		Metadata:          p.Metadata,
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
