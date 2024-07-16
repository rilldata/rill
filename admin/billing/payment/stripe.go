package payment

import (
	"context"
	"errors"
	"time"

	"github.com/rilldata/rill/admin/billing"
	"github.com/rilldata/rill/admin/database"
	"github.com/stripe/stripe-go/v79"
	"github.com/stripe/stripe-go/v79/billingportal/session"
	"github.com/stripe/stripe-go/v79/customer"
)

var _ Provider = &Stripe{}

type Stripe struct{}

func NewStripe(stripeKey string) *Stripe {
	stripe.Key = stripeKey
	return &Stripe{}
}

func (s *Stripe) CreateCustomer(ctx context.Context, organization *database.Organization) (*Customer, error) {
	// Create a new customer
	params := &stripe.CustomerParams{
		Email: stripe.String(billing.SupportEmail), // TODO capture email for the org or use admin's email
		Name:  stripe.String(organization.ID),
	}

	c, err := customer.New(params)
	if err != nil {
		return nil, err
	}

	return &Customer{
		ID:    c.ID,
		Name:  c.Name,
		Email: c.Email,
	}, nil
}

func (s *Stripe) FindCustomer(ctx context.Context, customerID string) (*Customer, error) {
	c, err := customer.Get(customerID, nil)
	if err != nil {
		var stripeErr *stripe.Error
		if errors.As(err, &stripeErr) && stripeErr.Code == stripe.ErrorCodeResourceMissing {
			return nil, billing.ErrNotFound
		}
		return nil, err
	}

	i := customer.ListPaymentMethods(&stripe.CustomerListPaymentMethodsParams{
		Customer: stripe.String(c.ID),
	})

	return &Customer{
		ID:                 c.ID,
		Name:               c.Name,
		Email:              c.Email,
		ValidPaymentMethod: i.Next(), // very basic check if the customer has a payment method // TODO improve this
	}, nil
}

func (s *Stripe) FindCustomerForOrg(ctx context.Context, organization *database.Organization) (*Customer, error) {
	// TODO once we capture billing email then we can use that to list customers
	searchStart := organization.CreatedOn.Add(-5 * time.Minute) // search 5 minutes before the org creation time
	searchEnd := organization.CreatedOn.Add(5 * time.Minute)    // search 5 minutes after the org creation time
	params := &stripe.CustomerListParams{
		CreatedRange: &stripe.RangeQueryParams{
			GreaterThanOrEqual: searchStart.Unix(),
			LesserThanOrEqual:  searchEnd.Unix(),
		},
	}

	i := customer.List(params)
	for i.Next() {
		c := i.Customer()
		if c.Name == organization.ID {
			it := customer.ListPaymentMethods(&stripe.CustomerListPaymentMethodsParams{
				Customer: stripe.String(c.ID),
			})
			return &Customer{
				ID:                 c.ID,
				Name:               c.Name,
				Email:              c.Email,
				ValidPaymentMethod: it.Next(), // very basic check if the customer has a payment method // TODO improve this
			}, nil
		}
	}

	return nil, billing.ErrNotFound
}

func (s *Stripe) DeleteCustomer(ctx context.Context, customerID string) error {
	_, err := customer.Del(customerID, nil)
	return err
}

func (s *Stripe) GetBillingPortalURL(ctx context.Context, customerID, returnURL string) (string, error) {
	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(customerID),
		ReturnURL: stripe.String(returnURL),
	}
	sess, err := session.New(params)
	if err != nil {
		return "", err
	}

	return sess.URL, nil
}
