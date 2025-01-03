package billing

import (
	"context"
	"errors"
	"time"

	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/jobs"
	"github.com/rilldata/rill/runtime/pkg/httputil"
)

const (
	SupportEmail    = "support@rilldata.com"
	DefaultTimeZone = "UTC"
)

var ErrNotFound = errors.New("not found")

type Biller interface {
	Name() string
	GetDefaultPlan(ctx context.Context) (*Plan, error)
	GetPlans(ctx context.Context) ([]*Plan, error)
	// GetPublicPlans for listing purposes
	GetPublicPlans(ctx context.Context) ([]*Plan, error)
	// GetPlan returns the plan with the given biller plan ID.
	GetPlan(ctx context.Context, id string) (*Plan, error)
	// GetPlanByName returns the plan with the given Rill plan name.
	GetPlanByName(ctx context.Context, name string) (*Plan, error)

	// CreateCustomer creates a customer for the given organization in the billing system and returns the external customer ID.
	CreateCustomer(ctx context.Context, organization *database.Organization, provider PaymentProvider) (*Customer, error)
	FindCustomer(ctx context.Context, customerID string) (*Customer, error)
	UpdateCustomerPaymentID(ctx context.Context, customerID string, provider PaymentProvider, paymentProviderID string) error
	UpdateCustomerEmail(ctx context.Context, customerID, email string) error
	DeleteCustomer(ctx context.Context, customerID string) error

	// CreateSubscription creates a subscription for the given organization. Subscription starts immediately.
	CreateSubscription(ctx context.Context, customerID string, plan *Plan) (*Subscription, error)
	// GetActiveSubscription returns the active subscription for the given organization
	GetActiveSubscription(ctx context.Context, customerID string) (*Subscription, error)
	// CancelSubscriptionsForCustomer cancels all the subscriptions for the given organization and returns the end date of the subscription
	CancelSubscriptionsForCustomer(ctx context.Context, customerID string, cancelOption SubscriptionCancellationOption) (time.Time, error)

	// ChangeSubscriptionPlan changes the plan of the given subscription immediately and returns the updated subscription
	ChangeSubscriptionPlan(ctx context.Context, subscriptionID string, plan *Plan) (*Subscription, error)
	// UnscheduleCancellation cancels the scheduled cancellation for the given subscription and returns the updated subscription
	UnscheduleCancellation(ctx context.Context, subscriptionID string) (*Subscription, error)

	GetInvoice(ctx context.Context, invoiceID string) (*Invoice, error)
	IsInvoiceValid(ctx context.Context, invoice *Invoice) bool
	IsInvoicePaid(ctx context.Context, invoice *Invoice) bool

	MarkCustomerTaxExempt(ctx context.Context, customerID string) error
	UnmarkCustomerTaxExempt(ctx context.Context, customerID string) error

	ReportUsage(ctx context.Context, usage []*Usage) error

	GetReportingGranularity() UsageReportingGranularity
	GetReportingWorkerCron() string

	// WebhookHandlerFunc returns a http.HandlerFunc that can be used to handle incoming webhooks from the payment provider. Return nil if you don't want to register any webhook handlers. jobs is used to enqueue jobs for processing the webhook events.
	WebhookHandlerFunc(ctx context.Context, jobs jobs.Client) httputil.Handler

	// GetCurrentPlanDisplayName this is specifically added for the UI to show the current plan name
	GetCurrentPlanDisplayName(ctx context.Context, customerID string) (string, error)
}

type PlanType int

const (
	TrailPlanType PlanType = iota
	TeamPlanType
	ManagedPlanType
	EnterprisePlanType
)

type Plan struct {
	ID              string // ID of the plan in the external billing system
	Name            string // Unique name of the plan in Rill, can be empty if biller does not support it
	PlanType        PlanType
	DisplayName     string
	Description     string
	TrialPeriodDays int
	Default         bool
	Public          bool
	Quotas          Quotas
	Metadata        map[string]string
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

type planMetadata struct {
	Default                        bool   `mapstructure:"default"`
	Public                         bool   `mapstructure:"public"`
	StorageLimitBytesPerDeployment *int64 `mapstructure:"storage_limit_bytes_per_deployment"`
	NumProjects                    *int   `mapstructure:"num_projects"`
	NumDeployments                 *int   `mapstructure:"num_deployments"`
	NumSlotsTotal                  *int   `mapstructure:"num_slots_total"`
	NumSlotsPerDeployment          *int   `mapstructure:"num_slots_per_deployment"`
	NumOutstandingInvites          *int   `mapstructure:"num_outstanding_invites"`
}

type Subscription struct {
	ID                           string
	Customer                     *Customer
	Plan                         *Plan
	StartDate                    time.Time
	EndDate                      time.Time
	CurrentBillingCycleStartDate time.Time
	CurrentBillingCycleEndDate   time.Time
	TrialEndDate                 time.Time
	Metadata                     map[string]string
}

type Customer struct {
	ID                string
	Email             string
	Name              string
	PaymentProviderID string
	PortalURL         string
}

type Usage struct {
	CustomerID     string
	MetricName     string
	Value          float64
	ReportingGrain UsageReportingGranularity
	StartTime      time.Time // Start time of the usage period
	EndTime        time.Time // End time of the usage period
	Metadata       map[string]interface{}
}

type Invoice struct {
	ID             string
	Status         string
	CustomerID     string
	Amount         string
	Currency       string
	DueDate        time.Time
	CreatedAt      time.Time
	SubscriptionID string
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

type PaymentProvider string

const (
	PaymentProviderStripe PaymentProvider = "stripe"
)

func Email(organization *database.Organization) string {
	if organization.BillingEmail != "" {
		return organization.BillingEmail
	}
	return SupportEmail
}
