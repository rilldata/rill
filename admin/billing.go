package admin

import (
	"context"

	"github.com/rilldata/rill/admin/billing"
	"github.com/rilldata/rill/admin/database"
	"go.uber.org/zap"
)

func (s *Service) InitOrganizationBilling(ctx context.Context, org *database.Organization, plan *billing.Plan) (*database.Organization, []*billing.Subscription, error) {
	update := false
	newPayment := false
	if org.PaymentCustomerID == "" {
		// payment customer not created yet
		cust, err := s.Payment.CreateCustomer(ctx, org)
		if err != nil {
			return nil, nil, err
		}
		s.Logger.Info("created payment customer", zap.String("org", org.Name), zap.String("customer_id", cust.ID))
		newPayment = true
		update = true
		org.PaymentCustomerID = cust.ID
	}

	if org.BillingCustomerID == "" {
		// billing customer not created yet
		cust, err := s.Biller.CreateCustomer(ctx, org)
		if err != nil {
			return nil, nil, err
		}
		s.Logger.Info("created billing customer", zap.String("org", org.Name), zap.String("customer_id", cust.ID))
		update = true
		org.BillingCustomerID = cust.ID
	} else if newPayment {
		// update payment customer id in billing system
		err := s.Biller.UpdateCustomerPaymentID(ctx, org.BillingCustomerID, org.PaymentCustomerID)
		if err != nil {
			return nil, nil, err
		}
	}

	subs, err := s.Biller.GetSubscriptionsForCustomer(ctx, org.BillingCustomerID)
	if err != nil {
		return nil, nil, err
	}

	if len(subs) == 0 {
		sub, err := s.Biller.CreateSubscription(ctx, org.BillingCustomerID, plan)
		if err != nil {
			return nil, nil, err
		}
		s.Logger.Info("created subscription", zap.String("org", org.Name), zap.String("subscription_id", sub.ID))
		subs = append(subs, sub)
		update = true
	}

	if update {
		org, err = s.DB.UpdateOrganization(ctx, org.ID, &database.UpdateOrganizationOptions{
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
			return nil, nil, err
		}
	}

	return org, subs, nil
}

func valOrDefault[T any](ptr *T, def T) T {
	if ptr != nil {
		return *ptr
	}
	return def
}
