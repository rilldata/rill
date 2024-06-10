package billing

import (
	"context"
	"time"

	"github.com/rilldata/rill/admin/database"
)

const (
	DefaultPlanID   = "Starter"
	SupportEmail    = "support@rilldata.com"
	DefaultTimeZone = "UTC"
)

type Biller interface {
	GetDefaultPlan(ctx context.Context) (*Plan, error)
	GetPlans(ctx context.Context) ([]*Plan, error)
	// GetPublicPlans for listing purposes
	GetPublicPlans(ctx context.Context) ([]*Plan, error)
	// GetPlan returns the plan with the given Rill plan ID or biller plan ID.
	GetPlan(ctx context.Context, rillPlanID string, billerPlanID string) (*Plan, error)

	// CreateCustomer creates a customer for the given organization in the billing system and returns the external customer ID.
	CreateCustomer(ctx context.Context, organization *database.Organization) (string, error)

	// CreateSubscription creates a subscription for the given organization.
	// The subscription starts immediately. TODO - support starting at a future date.
	CreateSubscription(ctx context.Context, customerID string, plan *Plan) (*Subscription, error)
	CancelSubscription(ctx context.Context, subscriptionID string, cancelOption SubscriptionCancellationOption, cancellationDate time.Time) error
	GetSubscriptionsForCustomer(ctx context.Context, customerID string) ([]*Subscription, error)
	// CancelSubscriptionsForCustomer deletes the subscription for the given organization.
	// cancellationDate only applicable if option is SubscriptionCancellationOptionRequestedDate
	CancelSubscriptionsForCustomer(ctx context.Context, customerID string, cancelOption SubscriptionCancellationOption, cancellationDate time.Time) error

	ReportUsage(ctx context.Context, customerID string, usage []*Usage) error

	GetReportingGranularity() UsageReportingGranularity
	GetReportingWorkerCron() string
}

type Plan struct {
	BillerID          string // ID of the plan in the external billing system
	RillID            string // ID of the plan in Rill, can be empty if biller does not support it
	Name              string
	Description       string
	TrialPeriodDays   int
	Quota             Quota
	ReportableMetrics []string // list of metric names that are reported to the billing system
	Metadata          map[string]string
}

type Quota struct {
	ManagedDataBytes *int64
	NumUsers         *int

	// Existing quotas
	NumProjects           *int
	NumDeployments        *int
	NumSlotsTotal         *int
	NumSlotsPerDeployment *int
	NumOutstandingInvites *int
}

type Subscription struct {
	ID                      string
	CustomerID              string
	Plan                    *Plan
	StartDate               time.Time
	EndDate                 time.Time
	CurrentBillingStartDate time.Time
	CurrentBillingEndDate   time.Time
	TrialEndDate            time.Time
	Metadata                map[string]string
}

type Usage struct {
	MetricName    string
	Amount        float64
	ReportingGran UsageReportingGranularity
	StartTime     time.Time // Start time of the usage period
	EndTime       time.Time // End time of the usage period
	Metadata      map[string]interface{}
}

type UsageReportingGranularity string

const (
	UsageReportingGranularityNone UsageReportingGranularity = "none"
	UsageReportingGranularityHour UsageReportingGranularity = "hour"
)

type SubscriptionCancellationOption int

const (
	SubscriptionCancellationOptionEndOfSubscriptionTerm SubscriptionCancellationOption = iota
	SubscriptionCancellationOptionImmediate
	SubscriptionCancellationOptionRequestedDate
)
