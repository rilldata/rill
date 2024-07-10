package payment

import (
	"context"

	"github.com/rilldata/rill/admin/database"
)

type Provider interface {
	CreateCustomer(ctx context.Context, organization *database.Organization) (*Customer, error)
	FindCustomer(ctx context.Context, customerID string) (*Customer, error)
	FindCustomerForOrg(ctx context.Context, organization *database.Organization) (*Customer, error)
	DeleteCustomer(ctx context.Context, customerID string) error
	// GetBillingPortalURL returns the payment portal URL to collect payment information from the customer.
	GetBillingPortalURL(ctx context.Context, customerID, returnURL string) (string, error)
}

type Customer struct {
	ID                 string
	Name               string
	Email              string
	ValidPaymentMethod bool
}
