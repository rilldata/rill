package payment

import (
	"context"

	"github.com/rilldata/rill/admin/database"
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
}

type Customer struct {
	ID               string
	Name             string
	Email            string
	HasPaymentMethod bool
}
