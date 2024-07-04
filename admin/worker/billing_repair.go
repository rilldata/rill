package worker

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rilldata/rill/admin/billing"
	"github.com/rilldata/rill/admin/database"
	"go.uber.org/zap"
)

func (w *Worker) repairOrgBilling(ctx context.Context) error {
	startTime := time.Now().UTC()
	t, err := w.admin.DB.FindBillingRepairedOn(ctx)
	if err != nil {
		return fmt.Errorf("failed to get last billing repair time: %w", err)
	}

	ids, err := w.admin.DB.FindOrganizationIDsWithoutPaymentCreatedOnOrAfter(ctx, t)
	if err != nil {
		return fmt.Errorf("failed to get organizations without payment created after %s: %w", t, err)
	}

	for _, id := range ids {
		org, err := w.admin.DB.FindOrganization(ctx, id)
		if err != nil {
			w.logger.Error("failed to get organization", zap.String("organization_id", id), zap.Error(err))
			continue
		}
		// check if customer exits in the billing system
		billingCustomer, err := w.admin.Biller.FindCustomer(ctx, org.ID)
		if err != nil && !errors.Is(err, billing.ErrNotFound) {
			w.logger.Error("failed to find billing customer", zap.String("organization_id", id), zap.Error(err))
			continue
		}

		if billingCustomer != nil {
			org.BillingCustomerID = billingCustomer.ID
			if billingCustomer.PaymentID != "" {
				org.PaymentCustomerID = billingCustomer.PaymentID
			}
		}

		if org.PaymentCustomerID == "" {
			cust, err := w.admin.Payment.FindCustomerForOrg(ctx, org)
			if err != nil {
				if errors.Is(err, billing.ErrNotFound) {
					// Create a new customer
					cust, err = w.admin.Payment.CreateCustomer(ctx, org)
					if err != nil {
						w.logger.Error("failed to create customer", zap.String("organization_id", id), zap.Error(err))
						continue
					}
				} else {
					w.logger.Error("failed to find customer", zap.String("organization_id", id), zap.Error(err))
					continue
				}
			}
			org.PaymentCustomerID = cust.ID
		}

		if billingCustomer == nil {
			// create a new customer
			cust, err := w.admin.Biller.CreateCustomer(ctx, org)
			if err != nil {
				w.logger.Error("failed to create billing customer", zap.String("organization_id", id), zap.Error(err))
				continue
			}
			w.logger.Info("created billing customer", zap.String("org", org.Name), zap.String("customer_id", cust.ID))
			org.BillingCustomerID = cust.ID
		} else if billingCustomer.PaymentID == "" {
			// update payment customer id in billing system
			err := w.admin.Biller.UpdateCustomerPaymentID(ctx, org.BillingCustomerID, org.PaymentCustomerID)
			if err != nil {
				w.logger.Error("failed to update payment customer id", zap.String("organization_id", id), zap.Error(err))
				continue
			}
		}

		subs, err := w.admin.Biller.GetSubscriptionsForCustomer(ctx, org.BillingCustomerID)
		if err != nil {
			w.logger.Error("failed to get subscriptions for customer", zap.String("organization_id", id), zap.Error(err))
			continue
		}

		var plan *billing.Plan
		if len(subs) == 0 {
			// get default plan
			plan, err = w.admin.Biller.GetDefaultPlan(ctx)
			if err != nil {
				w.logger.Error("failed to get default plan", zap.String("organization_id", id), zap.Error(err))
				continue
			}
			sub, err := w.admin.Biller.CreateSubscription(ctx, org.BillingCustomerID, plan)
			if err != nil {
				w.logger.Error("failed to create subscription", zap.String("organization_id", id), zap.Error(err))
				continue
			}
			w.logger.Info("created subscription", zap.String("org", org.Name), zap.String("subscription_id", sub.ID))
		} else {
			// get the latest subscription
			plan = subs[0].Plan
		}

		_, err = w.admin.DB.UpdateOrganization(ctx, org.ID, &database.UpdateOrganizationOptions{
			Name:                                org.Name,
			Description:                         org.Description,
			QuotaProjects:                       valOrDefault(plan.Quotas.NumProjects, org.QuotaProjects),
			QuotaDeployments:                    valOrDefault(plan.Quotas.NumDeployments, org.QuotaDeployments),
			QuotaSlotsTotal:                     valOrDefault(plan.Quotas.NumSlotsTotal, org.QuotaSlotsTotal),
			QuotaSlotsPerDeployment:             valOrDefault(plan.Quotas.NumSlotsPerDeployment, org.QuotaSlotsPerDeployment),
			QuotaOutstandingInvites:             valOrDefault(plan.Quotas.NumOutstandingInvites, org.QuotaOutstandingInvites),
			QuotaStorageLimitBytesPerDeployment: valOrDefault(plan.Quotas.StorageLimitBytesPerDeployment, org.QuotaStorageLimitBytesPerDeployment),
			BillingCustomerID:                   org.BillingCustomerID,
			PaymentCustomerID:                   org.PaymentCustomerID,
		})
		if err != nil {
			w.logger.Error("failed to update organization", zap.String("organization_id", id), zap.Error(err))
			continue
		}
		w.logger.Info("repaired billing for organization", zap.String("organization_id", id))
	}

	err = w.admin.DB.UpdateBillingRepairedOn(ctx, startTime)
	if err != nil {
		return fmt.Errorf("failed to update last billing repair time: %w", err)
	}
	return nil
}

func valOrDefault[T any](ptr *T, def T) T {
	if ptr != nil {
		return *ptr
	}
	return def
}
