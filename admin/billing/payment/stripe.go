package payment

import (
	"context"
	"errors"
	"time"

	"github.com/rilldata/rill/admin/billing"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/jobs"
	"github.com/rilldata/rill/runtime/pkg/httputil"
	"github.com/stripe/stripe-go/v79"
	"github.com/stripe/stripe-go/v79/billingportal/session"
	"github.com/stripe/stripe-go/v79/customer"
	"go.uber.org/zap"
)

var _ Provider = &Stripe{}

type Stripe struct {
	logger        *zap.Logger
	webhookSecret string
}

func NewStripe(logger *zap.Logger, stripeKey, stripeWebhookSecret string) *Stripe {
	stripe.Key = stripeKey
	return &Stripe{
		logger:        logger,
		webhookSecret: stripeWebhookSecret,
	}
}

func (s *Stripe) CreateCustomer(ctx context.Context, organization *database.Organization) (*Customer, error) {
	// Create a new customer
	params := &stripe.CustomerParams{
		Email: stripe.String(billing.Email(organization)),
		Name:  stripe.String(organization.ID),
	}

	c, err := customer.New(params)
	if err != nil {
		return nil, err
	}

	return getPaymentCustomerFromStripeCustomer(c), nil
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

	return getPaymentCustomerFromStripeCustomer(c), nil
}

func (s *Stripe) FindCustomerForOrg(ctx context.Context, organization *database.Organization) (*Customer, error) {
	searchStart := organization.CreatedOn.Add(-5 * time.Minute) // search 5 minutes before the org creation time
	searchEnd := organization.CreatedOn.Add(501 * time.Minute)  // search 15 minutes after the org creation time
	params := &stripe.CustomerListParams{
		Email: stripe.String(billing.Email(organization)),
		CreatedRange: &stripe.RangeQueryParams{
			GreaterThanOrEqual: searchStart.Unix(),
			LesserThanOrEqual:  searchEnd.Unix(),
		},
	}

	i := customer.List(params)
	for i.Next() {
		c := i.Customer()
		if c.Name == organization.ID {
			return getPaymentCustomerFromStripeCustomer(c), nil
		}
	}

	return nil, billing.ErrNotFound
}

func (s *Stripe) UpdateCustomerEmail(ctx context.Context, customerID, email string) error {
	params := &stripe.CustomerParams{
		Email: stripe.String(email),
	}

	_, err := customer.Update(customerID, params)
	return err
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

func (s *Stripe) WebhookHandlerFunc(ctx context.Context, jc jobs.Client) httputil.Handler {
	if s.webhookSecret == "" {
		return nil
	}
	sw := &stripeWebhook{stripe: s, jobs: jc}
	return sw.handleWebhook
}

func getPaymentCustomerFromStripeCustomer(c *stripe.Customer) *Customer {
	i := customer.ListPaymentMethods(&stripe.CustomerListPaymentMethodsParams{
		Customer: stripe.String(c.ID),
	})

	return &Customer{
		ID:                 c.ID,
		Name:               c.Name,
		Email:              c.Email,
		HasPaymentMethod:   i.Next(),
		HasBillableAddress: c.Address != nil && c.Address.PostalCode != "",
		TaxExempt:          c.Address != nil && c.Address.Country != "US" && c.Address.Country != "CA",
	}
}
