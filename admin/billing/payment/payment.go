package payment

import (
	"context"

	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/jobs"
	"github.com/rilldata/rill/runtime/pkg/httputil"
)

type Provider interface {
	CreateCustomer(ctx context.Context, organization *database.Organization) (*Customer, error)
	FindCustomer(ctx context.Context, customerID string) (*Customer, error)
	// FindCustomerForOrg Use with caution - This should only be used if we don't have payment customer ID in the org and we want to check if the customer already exists. Use FindCustomer instead if payment customer ID is available.
	// Stripe implementation for this is not optimal and list all customers around org creation time to find this customer.
	FindCustomerForOrg(ctx context.Context, organization *database.Organization) (*Customer, error)
	UpdateCustomerEmail(ctx context.Context, customerID, email string) error
	DeleteCustomer(ctx context.Context, customerID string) error
	// GetBillingPortalURL returns the payment portal URL to collect payment information from the customer.
	GetBillingPortalURL(ctx context.Context, customerID, returnURL string) (string, error)

	// WebhookHandlerFunc returns a http.HandlerFunc that can be used to handle incoming webhooks from the payment provider. Return nil if you don't want to register any webhook handlers. jobs is used to enqueue jobs for processing the webhook events.
	WebhookHandlerFunc(ctx context.Context, jobs jobs.Client) httputil.Handler
}

type Customer struct {
	ID                 string
	Name               string
	Email              string
	HasPaymentMethod   bool
	HasBillableAddress bool
}
