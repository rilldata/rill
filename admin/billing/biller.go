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
	// The subscription starts immediately.
	CreateSubscription(ctx context.Context, customerID string, plan *Plan) (*Subscription, error)
	CancelSubscription(ctx context.Context, subscriptionID string, cancelOption SubscriptionCancellationOption) error
	GetSubscriptionsForCustomer(ctx context.Context, customerID string) ([]*Subscription, error)
	ChangeSubscriptionPlan(ctx context.Context, subscriptionID string, plan *Plan, changeOption SubscriptionChangeOption) (*Subscription, error)
	// CancelSubscriptionsForCustomer deletes the subscription for the given organization.
	// cancellationDate only applicable if option is SubscriptionCancellationOptionRequestedDate
	CancelSubscriptionsForCustomer(ctx context.Context, customerID string, cancelOption SubscriptionCancellationOption) error

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
	Quotas            Quotas
	ReportableMetrics []string // list of metric names that are reported to the billing system
	Metadata          map[string]string
	// TODO do we need to expose pricing information
}

type Quotas struct {
	StorageLimitBytesPerDeployment *int64

	// Existing quotas
	NumProjects           *int
	NumDeployments        *int
	NumSlotsTotal         *int
	NumSlotsPerDeployment *int
	NumOutstandingInvites *int
}

type Subscription struct {
	ID                           string
	CustomerID                   string
	Plan                         *Plan
	StartDate                    time.Time
	EndDate                      time.Time
	CurrentBillingCycleStartDate time.Time
	CurrentBillingCycleEndDate   time.Time
	TrialEndDate                 time.Time
	Metadata                     map[string]string
}

type Usage struct {
	CustomerID     string
	MetricName     string
	Amount         float64
	ReportingGrain UsageReportingGranularity
	StartTime      time.Time // Start time of the usage period
	EndTime        time.Time // End time of the usage period
	Metadata       map[string]interface{}
}

type UsageReportingGranularity string

const (
	UsageReportingGranularityNone UsageReportingGranularity = ""
	UsageReportingGranularityHour UsageReportingGranularity = "hour"
)

type SubscriptionCancellationOption int

const (
	SubscriptionCancellationOptionEndOfSubscriptionTerm SubscriptionCancellationOption = iota
	SubscriptionCancellationOptionImmediate
)

type SubscriptionChangeOption int

const (
	SubscriptionChangeOptionEndOfSubscriptionTerm SubscriptionChangeOption = iota
	SubscriptionChangeOptionImmediate
)
