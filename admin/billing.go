package admin

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rilldata/rill/admin/billing"
	"github.com/rilldata/rill/admin/database"
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

	// scheduling it before the org update as repair billing job can take care of it if update fails
	err = s.ScheduleTrialEndCheckJobs(ctx, org.ID, sub.ID, plan.ID, sub.StartDate, sub.TrialEndDate)
	if err != nil {
		return nil, nil, err
	}

	// raise no payment method billing issue
	_, err = s.DB.UpsertBillingIssue(ctx, &database.UpsertBillingIssueOptions{
		OrgID:     org.ID,
		Type:      database.BillingIssueTypeNoPaymentMethod,
		Metadata:  &database.BillingIssueMetadataNoPaymentMethod{},
		EventTime: org.CreatedOn,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to upsert billing error: %w", err)
	}

	// raise no billable address billing issue
	_, err = s.DB.UpsertBillingIssue(ctx, &database.UpsertBillingIssueOptions{
		OrgID:     org.ID,
		Type:      database.BillingIssueTypeNoBillableAddress,
		Metadata:  &database.BillingIssueMetadataNoBillableAddress{},
		EventTime: org.CreatedOn,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to upsert billing error: %w", err)
	}

	org, err = s.DB.UpdateOrganization(ctx, org.ID, &database.UpdateOrganizationOptions{
		Name:                                org.Name,
		DisplayName:                         org.DisplayName,
		Description:                         org.Description,
		CustomDomain:                        org.CustomDomain,
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
			err = s.ScheduleTrialEndCheckJobs(ctx, org.ID, sub.ID, plan.ID, sub.StartDate, sub.TrialEndDate)
			if err != nil {
				return nil, nil, err
			}

			// raise no payment method billing issue
			_, err = s.DB.UpsertBillingIssue(ctx, &database.UpsertBillingIssueOptions{
				OrgID:     org.ID,
				Type:      database.BillingIssueTypeNoPaymentMethod,
				Metadata:  &database.BillingIssueMetadataNoPaymentMethod{},
				EventTime: org.CreatedOn,
			})
			if err != nil {
				return nil, nil, fmt.Errorf("failed to upsert billing error: %w", err)
			}

			// raise no billable address billing issue
			_, err = s.DB.UpsertBillingIssue(ctx, &database.UpsertBillingIssueOptions{
				OrgID:     org.ID,
				Type:      database.BillingIssueTypeNoBillableAddress,
				Metadata:  &database.BillingIssueMetadataNoBillableAddress{},
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
		err = s.ScheduleTrialEndCheckJobs(ctx, org.ID, sub.ID, plan.ID, sub.StartDate, sub.TrialEndDate)
		if err != nil {
			return nil, nil, err
		}

		// raise no payment method billing issue
		_, err = s.DB.UpsertBillingIssue(ctx, &database.UpsertBillingIssueOptions{
			OrgID:     org.ID,
			Type:      database.BillingIssueTypeNoPaymentMethod,
			Metadata:  &database.BillingIssueMetadataNoPaymentMethod{},
			EventTime: org.CreatedOn,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to upsert billing error: %w", err)
		}

		// raise no billable address billing issue
		_, err = s.DB.UpsertBillingIssue(ctx, &database.UpsertBillingIssueOptions{
			OrgID:     org.ID,
			Type:      database.BillingIssueTypeNoBillableAddress,
			Metadata:  &database.BillingIssueMetadataNoBillableAddress{},
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
		DisplayName:                         org.DisplayName,
		Description:                         org.Description,
		CustomDomain:                        org.CustomDomain,
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

func (s *Service) ScheduleTrialEndCheckJobs(ctx context.Context, orgID, subID, planID string, trialStartDate, trialEndDate time.Time) error {
	if trialEndDate.Before(time.Now()) {
		return nil
	}

	// schedule trial ending soon job
	tes, err := s.Jobs.TrialEndingSoon(ctx, orgID, subID, planID, trialEndDate)
	if err != nil {
		return fmt.Errorf("failed to schedule trial ending soon job: %w", err)
	}

	// schedule trial end check job
	tec, err := s.Jobs.TrialEndCheck(ctx, orgID, subID, planID, trialEndDate)
	if err != nil {
		return fmt.Errorf("failed to schedule trial end check job: %w", err)
	}

	// raise on-trial billing warning
	_, err = s.DB.UpsertBillingIssue(ctx, &database.UpsertBillingIssueOptions{
		OrgID: orgID,
		Type:  database.BillingIssueTypeOnTrial,
		Metadata: &database.BillingIssueMetadataOnTrial{
			EndDate:              trialEndDate,
			TrialEndingSoonJobID: tes.ID,
			TrialEndCheckJobID:   tec.ID,
		},
		EventTime: trialStartDate,
	})
	if err != nil {
		return fmt.Errorf("failed to upsert billing warning: %w", err)
	}

	return nil
}

// CleanupTrialBillingIssues removes trial related billing issues and cancel associated jobs
func (s *Service) CleanupTrialBillingIssues(ctx context.Context, orgID string) error {
	bitt, err := s.DB.FindBillingIssueByType(ctx, orgID, database.BillingIssueTypeTrialEnded)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			return fmt.Errorf("failed to find billing errors: %w", err)
		}
	}

	if bitt != nil {
		metadata, ok := bitt.Metadata.(*database.BillingIssueMetadataTrialEnded)
		if ok && metadata.GracePeriodEndJobID > 0 {
			// cancel the trial end grace check job, ignore errors.
			_ = s.Jobs.CancelJob(ctx, metadata.GracePeriodEndJobID)
		}
		err = s.DB.DeleteBillingIssue(ctx, bitt.ID)
		if err != nil {
			return fmt.Errorf("failed to delete billing error: %w", err)
		}
	}

	bwit, err := s.DB.FindBillingIssueByType(ctx, orgID, database.BillingIssueTypeOnTrial)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			return fmt.Errorf("failed to find billing warnings: %w", err)
		}
	}

	if bwit != nil {
		metadata, ok := bwit.Metadata.(*database.BillingIssueMetadataOnTrial)
		if ok && metadata.TrialEndingSoonJobID > 0 {
			// cancel the trial ending soon job, ignore errors.
			_ = s.Jobs.CancelJob(ctx, metadata.TrialEndingSoonJobID)
		}

		if ok && metadata.TrialEndCheckJobID > 0 {
			// cancel the trial end check job, ignore errors.
			_ = s.Jobs.CancelJob(ctx, metadata.TrialEndCheckJobID)
		}

		err = s.DB.DeleteBillingIssue(ctx, bwit.ID)
		if err != nil {
			return fmt.Errorf("failed to delete billing warning: %w", err)
		}
	}

	return nil
}

// CleanupBillingErrorSubCancellation removes subscription cancellation related billing error and cancel associated job
func (s *Service) CleanupBillingErrorSubCancellation(ctx context.Context, orgID string) error {
	bisc, err := s.DB.FindBillingIssueByType(ctx, orgID, database.BillingIssueTypeSubscriptionCancelled)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			return fmt.Errorf("failed to find billing errors: %w", err)
		}
	}

	if bisc != nil {
		metadata, ok := bisc.Metadata.(*database.BillingIssueMetadataSubscriptionCancelled)
		if ok && metadata.SubEndJobID > 0 {
			// cancel the subscription end check job, ignore errors.
			_ = s.Jobs.CancelJob(ctx, metadata.SubEndJobID)
		}
		err = s.DB.DeleteBillingIssue(ctx, bisc.ID)
		if err != nil {
			return fmt.Errorf("failed to delete billing error: %w", err)
		}
	}

	return nil
}

func (s *Service) CheckBillingErrors(ctx context.Context, orgID string) error {
	be, err := s.DB.FindBillingIssueByType(ctx, orgID, database.BillingIssueTypeTrialEnded)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			return err
		}
	}

	if be != nil {
		return fmt.Errorf("trial has ended")
	}

	be, err = s.DB.FindBillingIssueByType(ctx, orgID, database.BillingIssueTypeInvoicePaymentFailed)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			return err
		}
	}

	if be != nil { // should we allow any grace period here?
		return fmt.Errorf("invoice payment failed")
	}

	be, err = s.DB.FindBillingIssueByType(ctx, orgID, database.BillingIssueTypeSubscriptionCancelled)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			return err
		}
	}

	if be != nil && be.Metadata.(*database.BillingIssueMetadataSubscriptionCancelled).EndDate.AddDate(0, 0, 1).After(time.Now()) {
		return fmt.Errorf("subscription cancelled")
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
