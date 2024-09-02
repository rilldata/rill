package admin

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rilldata/rill/admin/billing"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/pkg/riverworker/riverutils"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
)

func (s *Service) InitOrganizationBilling(ctx context.Context, org *database.Organization) (*database.Organization, *billing.Subscription, error) {
	// TODO This can be moved to a background job and repair org billing job can be removed in the next version. We need repair job to fix existing orgs but afterwards background job wil ensure that all orgs are in sync with billing system
	// create payment customer
	pc, err := s.PaymentProvider.CreateCustomer(ctx, org)
	if err != nil {
		return nil, nil, err
	}
	s.Logger.Info("created payment customer", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.String("payment_customer_id", pc.ID))
	org.PaymentCustomerID = pc.ID

	// create billing customer
	bc, err := s.Biller.CreateCustomer(ctx, org, billing.PaymentProviderStripe)
	if err != nil {
		return nil, nil, err
	}
	s.Logger.Info("created billing customer", zap.String("org", org.Name), zap.String("billing_customer_id", bc.ID))
	org.BillingCustomerID = bc.ID

	// create subscription with the default plan
	plan, err := s.Biller.GetDefaultPlan(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get default plan: %w", err)
	}

	sub, err := s.Biller.CreateSubscription(ctx, org.BillingCustomerID, plan)
	if err != nil {
		return nil, nil, err
	}
	s.Logger.Info("created subscription", zap.String("org", org.Name), zap.String("subscription_id", sub.ID))

	// can be done in same tx as UpdateOrganization using InsertTx but river driver expects sql.Tx and does not work with the Tx implementation we have
	// scheduling it before the update as repair billing job can take care of the update if update fails
	err = s.ScheduleTrialEndCheckJobs(ctx, org.ID, sub.ID, plan.ID, sub.TrialEndDate)
	if err != nil {
		return nil, nil, err
	}

	// raise no payment method billing error
	_, err = s.DB.UpsertBillingError(ctx, &database.UpsertBillingErrorOptions{
		OrgID:     org.ID,
		Type:      database.BillingErrorTypeNoPaymentMethod,
		EventTime: org.CreatedOn,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to upsert billing error: %w", err)
	}

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
		BillingEmail:                        org.BillingEmail,
	})
	if err != nil {
		return nil, nil, err
	}

	return org, sub, nil
}

func (s *Service) RepairOrgBilling(ctx context.Context, org *database.Organization) (*database.Organization, []*billing.Subscription, error) {
	if org.BillingCustomerID != "" && org.PaymentCustomerID != "" {
		// get subscriptions for the customer
		subs, err := s.Biller.GetSubscriptionsForCustomer(ctx, org.BillingCustomerID)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get subscriptions for customer: %w", err)
		}
		// should not happen
		if len(subs) == 0 {
			// create a new subscription
			plan, err := s.Biller.GetDefaultPlan(ctx)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to get default plan: %w", err)
			}
			sub, err := s.Biller.CreateSubscription(ctx, org.BillingCustomerID, plan)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to create subscription: %w", err)
			}
			s.Logger.Info("created subscription", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.String("subscription_id", sub.ID))
			subs = append(subs, sub)

			// schedule trial end check job to the river queue, if it was already scheduled it will be ignored
			err = s.ScheduleTrialEndCheckJobs(ctx, org.ID, sub.ID, plan.ID, sub.TrialEndDate)
			if err != nil {
				return nil, nil, err
			}

			// raise no payment method billing error
			_, err = s.DB.UpsertBillingError(ctx, &database.UpsertBillingErrorOptions{
				OrgID:     org.ID,
				Type:      database.BillingErrorTypeNoPaymentMethod,
				EventTime: org.CreatedOn,
			})
			if err != nil {
				return nil, nil, fmt.Errorf("failed to upsert billing error: %w", err)
			}
		}
		if len(subs) > 1 {
			s.Logger.Warn("multiple subscriptions found for the customer", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.Int("num_subscriptions", len(subs)))
		}
		return org, subs, nil
	}

	// check if customer exits in the billing system
	billingCustomer, err := s.Biller.FindCustomer(ctx, org.ID)
	if err != nil && !errors.Is(err, billing.ErrNotFound) {
		return nil, nil, fmt.Errorf("error finding billing customer: %w", err)
	}

	if billingCustomer != nil {
		org.BillingCustomerID = billingCustomer.ID
		if billingCustomer.PaymentProviderID != "" {
			org.PaymentCustomerID = billingCustomer.PaymentProviderID
		}
	}

	if org.PaymentCustomerID == "" {
		cust, err := s.PaymentProvider.FindCustomerForOrg(ctx, org)
		if err != nil {
			if errors.Is(err, billing.ErrNotFound) {
				// Create a new customer
				cust, err = s.PaymentProvider.CreateCustomer(ctx, org)
				if err != nil {
					return nil, nil, fmt.Errorf("failed to create payment customer: %w", err)
				}
				s.Logger.Info("created payment customer", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.String("payment_customer_id", cust.ID))
			} else {
				return nil, nil, fmt.Errorf("error finding payment customer: %w", err)
			}
		}
		org.PaymentCustomerID = cust.ID
	}

	if billingCustomer == nil {
		// create a new customer
		cust, err := s.Biller.CreateCustomer(ctx, org, billing.PaymentProviderStripe)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create billing customer: %w", err)
		}
		s.Logger.Info("created billing customer", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.String("billing_customer_id", cust.ID))
		org.BillingCustomerID = cust.ID
	} else if billingCustomer.PaymentProviderID == "" {
		// update payment customer id in billing system
		err = s.Biller.UpdateCustomerPaymentID(ctx, org.BillingCustomerID, billing.PaymentProviderStripe, org.PaymentCustomerID)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to update payment customer id: %w", err)
		}
	}

	subs, err := s.Biller.GetSubscriptionsForCustomer(ctx, org.BillingCustomerID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get subscriptions for customer: %w", err)
	}

	var plan *billing.Plan
	if len(subs) == 0 {
		// get default plan
		plan, err = s.Biller.GetDefaultPlan(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get default plan: %w", err)
		}
		sub, err := s.Biller.CreateSubscription(ctx, org.BillingCustomerID, plan)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create subscription: %w", err)
		}
		s.Logger.Info("created subscription", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.String("subscription_id", sub.ID))
		subs = append(subs, sub)

		// schedule trial end check job to the river queue
		err = s.ScheduleTrialEndCheckJobs(ctx, org.ID, sub.ID, plan.ID, sub.TrialEndDate)
		if err != nil {
			return nil, nil, err
		}

		// raise no payment method billing error
		_, err = s.DB.UpsertBillingError(ctx, &database.UpsertBillingErrorOptions{
			OrgID:     org.ID,
			Type:      database.BillingErrorTypeNoPaymentMethod,
			EventTime: org.CreatedOn,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to upsert billing error: %w", err)
		}
	} else if len(subs) > 1 {
		s.Logger.Warn("multiple subscriptions found for the customer", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.Int("num_subscriptions", len(subs)))
	}
	// get the latest subscription
	plan = subs[0].Plan

	org, err = s.DB.UpdateOrganization(ctx, org.ID, &database.UpdateOrganizationOptions{
		Name:                                org.Name,
		Description:                         org.Description,
		QuotaProjects:                       biggerOfInt(plan.Quotas.NumProjects, org.QuotaProjects),
		QuotaDeployments:                    biggerOfInt(plan.Quotas.NumDeployments, org.QuotaDeployments),
		QuotaSlotsTotal:                     biggerOfInt(plan.Quotas.NumSlotsTotal, org.QuotaSlotsTotal),
		QuotaSlotsPerDeployment:             biggerOfInt(plan.Quotas.NumSlotsPerDeployment, org.QuotaSlotsPerDeployment),
		QuotaOutstandingInvites:             biggerOfInt(plan.Quotas.NumOutstandingInvites, org.QuotaOutstandingInvites),
		QuotaStorageLimitBytesPerDeployment: biggerOfInt64(plan.Quotas.StorageLimitBytesPerDeployment, org.QuotaStorageLimitBytesPerDeployment),
		BillingCustomerID:                   org.BillingCustomerID,
		PaymentCustomerID:                   org.PaymentCustomerID,
		BillingEmail:                        org.BillingEmail,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to update organization: %w", err)
	}
	return org, subs, nil
}

func (s *Service) ScheduleTrialEndCheckJobs(ctx context.Context, orgID, subID, planID string, trialEndDate time.Time) error {
	if trialEndDate.After(time.Now()) {
		return nil
	}

	// schedule trial ending soon job 7 days before trial end date
	_, err := riverutils.InsertOnlyRiverClient.Insert(ctx, riverutils.TrialEndingSoonArgs{
		OrgID:  orgID,
		SubID:  subID,
		PlanID: planID,
	}, &river.InsertOpts{
		ScheduledAt: trialEndDate.AddDate(0, 0, -7),
		UniqueOpts: river.UniqueOpts{
			ByArgs: true,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to schedule trial ending soon job: %w", err)
	}

	// schedule trial end check job on trial end date
	_, err = riverutils.InsertOnlyRiverClient.Insert(ctx, riverutils.TrialEndCheckArgs{
		OrgID:  orgID,
		SubID:  subID,
		PlanID: planID,
	}, &river.InsertOpts{
		ScheduledAt: trialEndDate.Add(time.Hour * 25), // add buffer of 1 hour to ensure the job runs after trial period days
		UniqueOpts: river.UniqueOpts{
			ByArgs: true,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to schedule trial end check job: %w", err)
	}

	return nil
}

func valOrDefault[T any](ptr *T, def T) T {
	if ptr != nil {
		return *ptr
	}
	return def
}

func biggerOfInt(ptr *int, def int) int {
	if ptr != nil {
		if *ptr < 0 || *ptr > def {
			return *ptr
		}
	}
	return def
}

func biggerOfInt64(ptr *int64, def int64) int64 {
	if ptr != nil {
		if *ptr < 0 || *ptr > def {
			return *ptr
		}
	}
	return def
}
