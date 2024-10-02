package river

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/database"
	"github.com/riverqueue/river"
)

type PaymentMethodAddedArgs struct {
	PaymentMethodID   string
	PaymentCustomerID string
	PaymentType       string
	EventTime         time.Time
}

func (PaymentMethodAddedArgs) Kind() string { return "payment_method_added" }

type PaymentMethodAddedWorker struct {
	river.WorkerDefaults[PaymentMethodAddedArgs]
	admin *admin.Service
}

func (w *PaymentMethodAddedWorker) Work(ctx context.Context, job *river.Job[PaymentMethodAddedArgs]) error {
	org, err := w.admin.DB.FindOrganizationForPaymentCustomerID(ctx, job.Args.PaymentCustomerID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			// org got deleted, ignore
			return nil
		}
		return fmt.Errorf("failed to find organization for payment customer id: %w", err)
	}

	// check for no payment method billing error
	be, err := w.admin.DB.FindBillingIssueByTypeForOrg(ctx, org.ID, database.BillingIssueTypeNoPaymentMethod)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			return fmt.Errorf("failed to find billing errors: %w", err)
		}
	}

	// delete the no payment method error if any payment method found for customer
	if be != nil {
		c, err := w.admin.PaymentProvider.FindCustomer(ctx, job.Args.PaymentCustomerID)
		if err != nil {
			return fmt.Errorf("failed to find customer: %w", err)
		}

		if c.HasPaymentMethod {
			err = w.admin.DB.DeleteBillingIssue(ctx, be.ID)
			if err != nil {
				return fmt.Errorf("failed to delete billing error: %w", err)
			}
		}
	}

	return nil
}

type PaymentMethodRemovedArgs struct {
	PaymentMethodID   string
	PaymentCustomerID string
	EventTime         time.Time
}

func (PaymentMethodRemovedArgs) Kind() string { return "payment_method_removed" }

type PaymentMethodRemovedWorker struct {
	river.WorkerDefaults[PaymentMethodRemovedArgs]
	admin *admin.Service
}

func (w *PaymentMethodRemovedWorker) Work(ctx context.Context, job *river.Job[PaymentMethodRemovedArgs]) error {
	org, err := w.admin.DB.FindOrganizationForPaymentCustomerID(ctx, job.Args.PaymentCustomerID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			// org got deleted, ignore
			return nil
		}
		return fmt.Errorf("failed to find organization for payment customer id: %w", err)
	}

	// check payment provider if the customer has any payment method
	c, err := w.admin.PaymentProvider.FindCustomer(ctx, job.Args.PaymentCustomerID)
	if err != nil {
		return fmt.Errorf("failed to find customer: %w", err)
	}

	if !c.HasPaymentMethod {
		_, err = w.admin.DB.UpsertBillingIssue(ctx, &database.UpsertBillingIssueOptions{
			OrgID:     org.ID,
			Type:      database.BillingIssueTypeNoPaymentMethod,
			Metadata:  &database.BillingIssueMetadataNoPaymentMethod{},
			EventTime: job.Args.EventTime,
		})
		if err != nil {
			return fmt.Errorf("failed to add billing error: %w", err)
		}
	}

	return nil
}

type CustomerAddressUpdatedArgs struct {
	PaymentCustomerID string
	EventTime         time.Time
}

func (CustomerAddressUpdatedArgs) Kind() string { return "customer_address_updated" }

type CustomerAddressUpdatedWorker struct {
	river.WorkerDefaults[CustomerAddressUpdatedArgs]
	admin *admin.Service
}

func (w *CustomerAddressUpdatedWorker) Work(ctx context.Context, job *river.Job[CustomerAddressUpdatedArgs]) error {
	org, err := w.admin.DB.FindOrganizationForPaymentCustomerID(ctx, job.Args.PaymentCustomerID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			// org got deleted, ignore
			return nil
		}
		return fmt.Errorf("failed to find organization for payment customer id: %w", err)
	}

	// look for no billable address billing error and remove it
	be, err := w.admin.DB.FindBillingIssueByTypeForOrg(ctx, org.ID, database.BillingIssueTypeNoBillableAddress)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			return fmt.Errorf("failed to find billing errors: %w", err)
		}
	}

	if be != nil {
		err = w.admin.DB.DeleteBillingIssue(ctx, be.ID)
		if err != nil {
			return fmt.Errorf("failed to delete billing error: %w", err)
		}
	}

	return nil
}
