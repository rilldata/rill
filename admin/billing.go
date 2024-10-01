package admin

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rilldata/rill/admin/billing"
	"github.com/rilldata/rill/admin/billing/payment"
	"github.com/rilldata/rill/admin/database"
	"go.uber.org/zap"
)

func (s *Service) InitOrganizationBilling(ctx context.Context, org *database.Organization) (*database.Organization, error) {
	// create payment customer
	pc, err := s.PaymentProvider.CreateCustomer(ctx, org)
	if err != nil {
		return nil, err
	}
	s.Logger.Info("created payment customer", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.String("payment_customer_id", pc.ID))
	org.PaymentCustomerID = pc.ID

	// create billing customer
	bc, err := s.Biller.CreateCustomer(ctx, org, billing.PaymentProviderStripe)
	if err != nil {
		return nil, err
	}
	s.Logger.Info("created billing customer", zap.String("org", org.Name), zap.String("billing_customer_id", bc.ID))
	org.BillingCustomerID = bc.ID

	org, err = s.DB.UpdateOrganization(ctx, org.ID, &database.UpdateOrganizationOptions{
		Name:                                org.Name,
		DisplayName:                         org.DisplayName,
		Description:                         org.Description,
		CustomDomain:                        org.CustomDomain,
		QuotaProjects:                       org.QuotaProjects,
		QuotaDeployments:                    org.QuotaDeployments,
		QuotaSlotsTotal:                     org.QuotaSlotsTotal,
		QuotaSlotsPerDeployment:             org.QuotaSlotsPerDeployment,
		QuotaOutstandingInvites:             org.QuotaOutstandingInvites,
		QuotaStorageLimitBytesPerDeployment: org.QuotaStorageLimitBytesPerDeployment,
		BillingCustomerID:                   org.BillingCustomerID,
		PaymentCustomerID:                   org.PaymentCustomerID,
		BillingEmail:                        org.BillingEmail,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update organization: %w", err)
	}

	err = s.RaiseNewOrgBillingIssues(ctx, org.ID, org.CreatedOn, pc.HasPaymentMethod, bc.HasBillableAddress)
	if err != nil {
		return nil, err
	}

	return org, nil
}

// RepairOrganizationBilling repairs billing for an organization by checking if customer exists in billing systems, if not creating new. Useful for migrating existing orgs to billing system and in rare case when InitOrganizationBilling fails in the middle
func (s *Service) RepairOrganizationBilling(ctx context.Context, org *database.Organization, initSub bool) (*database.Organization, []*billing.Subscription, error) {
	var bc *billing.Customer
	var pc *payment.Customer
	var err error

	bc, err = s.Biller.FindCustomer(ctx, org.ID)
	if err != nil && !errors.Is(err, billing.ErrNotFound) {
		return nil, nil, fmt.Errorf("error finding billing customer: %w", err)
	}

	if bc != nil {
		org.BillingCustomerID = bc.ID
		if bc.PaymentProviderID != "" {
			org.PaymentCustomerID = bc.PaymentProviderID
		}
	}

	if org.PaymentCustomerID == "" {
		pc, err = s.PaymentProvider.FindCustomerForOrg(ctx, org)
		if err != nil {
			if errors.Is(err, billing.ErrNotFound) {
				// Create a new customer
				pc, err = s.PaymentProvider.CreateCustomer(ctx, org)
				if err != nil {
					return nil, nil, fmt.Errorf("failed to create payment customer: %w", err)
				}
				s.Logger.Info("created payment customer", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.String("payment_customer_id", pc.ID))
			} else {
				return nil, nil, fmt.Errorf("error finding payment customer: %w", err)
			}
		}
		org.PaymentCustomerID = pc.ID
	}

	if pc == nil {
		pc, err = s.PaymentProvider.FindCustomer(ctx, org.PaymentCustomerID)
		if err != nil {
			return nil, nil, fmt.Errorf("error finding payment customer: %w", err)
		}
	}

	if bc == nil {
		bc, err = s.Biller.CreateCustomer(ctx, org, billing.PaymentProviderStripe)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create billing customer: %w", err)
		}
		s.Logger.Info("created billing customer", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.String("billing_customer_id", bc.ID))
		org.BillingCustomerID = bc.ID
	} else if bc.PaymentProviderID == "" {
		// update payment customer id in billing system
		err = s.Biller.UpdateCustomerPaymentID(ctx, org.BillingCustomerID, billing.PaymentProviderStripe, org.PaymentCustomerID)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to update payment customer id: %w", err)
		}
	}

	// update billing and payment customer id
	org, err = s.DB.UpdateOrganization(ctx, org.ID, &database.UpdateOrganizationOptions{
		Name:                                org.Name,
		DisplayName:                         org.DisplayName,
		Description:                         org.Description,
		CustomDomain:                        org.CustomDomain,
		QuotaProjects:                       org.QuotaProjects,
		QuotaDeployments:                    org.QuotaDeployments,
		QuotaSlotsTotal:                     org.QuotaSlotsTotal,
		QuotaSlotsPerDeployment:             org.QuotaSlotsPerDeployment,
		QuotaOutstandingInvites:             org.QuotaOutstandingInvites,
		QuotaStorageLimitBytesPerDeployment: org.QuotaStorageLimitBytesPerDeployment,
		BillingCustomerID:                   org.BillingCustomerID,
		PaymentCustomerID:                   org.PaymentCustomerID,
		BillingEmail:                        org.BillingEmail,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to update organization: %w", err)
	}

	err = s.RaiseNewOrgBillingIssues(ctx, org.ID, org.CreatedOn, pc.HasPaymentMethod, bc.HasBillableAddress)
	if err != nil {
		return nil, nil, err
	}

	if !initSub {
		return org, nil, nil
	}

	subs, err := s.Biller.GetActiveSubscriptionsForCustomer(ctx, org.BillingCustomerID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get subscriptions for customer: %w", err)
	}

	if len(subs) > 1 {
		s.Logger.Named("billing").Warn("multiple subscriptions found for org, please check manually", zap.String("org_id", org.ID), zap.String("org_name", org.Name))
		return org, subs, nil
	}

	var updatedOrg *database.Organization
	if len(subs) == 0 {
		var sub *billing.Subscription
		updatedOrg, sub, err = s.StartTrial(ctx, org)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to start trial: %w", err)
		}
		subs = append(subs, sub)
	} else {
		s.Logger.Named("billing").Warn("subscription already exists for org", zap.String("org_id", org.ID), zap.String("org_name", org.Name))
		// update org quotas, this subscription might have been manually created
		updatedOrg, err = s.DB.UpdateOrganization(ctx, org.ID, &database.UpdateOrganizationOptions{
			Name:                                org.Name,
			DisplayName:                         org.DisplayName,
			Description:                         org.Description,
			CustomDomain:                        org.CustomDomain,
			QuotaProjects:                       biggerOfInt(subs[0].Plan.Quotas.NumProjects, org.QuotaProjects),
			QuotaDeployments:                    biggerOfInt(subs[0].Plan.Quotas.NumDeployments, org.QuotaDeployments),
			QuotaSlotsTotal:                     biggerOfInt(subs[0].Plan.Quotas.NumSlotsTotal, org.QuotaSlotsTotal),
			QuotaSlotsPerDeployment:             biggerOfInt(subs[0].Plan.Quotas.NumSlotsPerDeployment, org.QuotaSlotsPerDeployment),
			QuotaOutstandingInvites:             biggerOfInt(subs[0].Plan.Quotas.NumOutstandingInvites, org.QuotaOutstandingInvites),
			QuotaStorageLimitBytesPerDeployment: biggerOfInt64(subs[0].Plan.Quotas.StorageLimitBytesPerDeployment, org.QuotaStorageLimitBytesPerDeployment),
			BillingCustomerID:                   org.BillingCustomerID,
			PaymentCustomerID:                   org.PaymentCustomerID,
			BillingEmail:                        org.BillingEmail,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to update organization: %w", err)
		}
	}

	return updatedOrg, subs, nil
}

func (s *Service) StartTrial(ctx context.Context, org *database.Organization) (*database.Organization, *billing.Subscription, error) {
	// get default plan
	plan, err := s.Biller.GetDefaultPlan(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get default plan: %w", err)
	}

	subs, err := s.Biller.GetActiveSubscriptionsForCustomer(ctx, org.BillingCustomerID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get subscriptions for customer: %w", err)
	}

	if len(subs) > 1 {
		s.Logger.Named("billing").Warn("multiple subscriptions found for org, please check manually", zap.String("org_id", org.ID), zap.String("org_name", org.Name))
		return nil, nil, nil
	}

	if len(subs) == 1 && subs[0].Plan.ID != plan.ID {
		s.Logger.Named("billing").Warn("subscription already exists for org with different plan, please check manually", zap.String("org_id", org.ID), zap.String("org_name", org.Name))
		return nil, nil, nil
	}

	var sub *billing.Subscription
	if len(subs) == 0 {
		sub, err = s.Biller.CreateSubscription(ctx, org.BillingCustomerID, plan)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create subscription: %w", err)
		}
	} else {
		sub = subs[0]
	}

	if sub.ID == "" || sub.Plan.ID == "" {
		// happens with noop biller
		s.Logger.Named("billing").Warn("no subscription or plan ID provided, skipping org and billing issues update", zap.String("org_id", org.ID))
		return org, sub, nil
	}

	s.Logger.Named("billing").Info("started trial", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.String("subscription_id", sub.ID))

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
		return nil, nil, fmt.Errorf("failed to update organization: %w", err)
	}

	// delete never subscribed billing issue
	bins, err := s.DB.FindBillingIssueByTypeForOrg(ctx, org.ID, database.BillingIssueTypeNeverSubscribed)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			return nil, nil, fmt.Errorf("failed to find billing issue: %w", err)
		}
	}
	if bins != nil {
		err = s.DB.DeleteBillingIssue(ctx, bins.ID)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to delete billing issue: %w", err)
		}
	}

	// raise on-trial billing warning
	_, err = s.DB.UpsertBillingIssue(ctx, &database.UpsertBillingIssueOptions{
		OrgID: org.ID,
		Type:  database.BillingIssueTypeOnTrial,
		Metadata: &database.BillingIssueMetadataOnTrial{
			SubID:   sub.ID,
			PlanID:  sub.Plan.ID,
			EndDate: sub.TrialEndDate,
		},
		EventTime: sub.StartDate,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to upsert billing warning: %w", err)
	}

	return org, sub, nil
}

// RaiseNewOrgBillingIssues raises billing issues for a new organization
func (s *Service) RaiseNewOrgBillingIssues(ctx context.Context, orgID string, creationTime time.Time, hasPaymentMethod, hasBillableAddress bool) error {
	if !hasPaymentMethod {
		_, err := s.DB.UpsertBillingIssue(ctx, &database.UpsertBillingIssueOptions{
			OrgID:     orgID,
			Type:      database.BillingIssueTypeNoPaymentMethod,
			Metadata:  &database.BillingIssueMetadataNoPaymentMethod{},
			EventTime: creationTime,
		})
		if err != nil {
			return fmt.Errorf("failed to upsert billing error: %w", err)
		}
	}

	if !hasBillableAddress {
		_, err := s.DB.UpsertBillingIssue(ctx, &database.UpsertBillingIssueOptions{
			OrgID:     orgID,
			Type:      database.BillingIssueTypeNoBillableAddress,
			Metadata:  &database.BillingIssueMetadataNoBillableAddress{},
			EventTime: creationTime,
		})
		if err != nil {
			return fmt.Errorf("failed to upsert billing error: %w", err)
		}
	}

	return nil
}

// CleanupTrialBillingIssues removes trial related billing issues
func (s *Service) CleanupTrialBillingIssues(ctx context.Context, orgID string) error {
	bite, err := s.DB.FindBillingIssueByTypeForOrg(ctx, orgID, database.BillingIssueTypeTrialEnded)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			return fmt.Errorf("failed to find billing issue: %w", err)
		}
	}

	if bite != nil {
		err = s.DB.DeleteBillingIssue(ctx, bite.ID)
		if err != nil {
			return fmt.Errorf("failed to delete billing issue: %w", err)
		}
	}

	biot, err := s.DB.FindBillingIssueByTypeForOrg(ctx, orgID, database.BillingIssueTypeOnTrial)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			return fmt.Errorf("failed to find billing issue: %w", err)
		}
	}

	if biot != nil {
		err = s.DB.DeleteBillingIssue(ctx, biot.ID)
		if err != nil {
			return fmt.Errorf("failed to delete billing issue: %w", err)
		}
	}

	return nil
}

// CleanupSubscriptionBillingIssues removes subscription related billing issues
func (s *Service) CleanupSubscriptionBillingIssues(ctx context.Context, orgID string) error {
	bins, err := s.DB.FindBillingIssueByTypeForOrg(ctx, orgID, database.BillingIssueTypeNeverSubscribed)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			return fmt.Errorf("failed to find billing issue: %w", err)
		}
	}

	if bins != nil {
		err = s.DB.DeleteBillingIssue(ctx, bins.ID)
		if err != nil {
			return fmt.Errorf("failed to delete billing issue: %w", err)
		}
	}

	bisc, err := s.DB.FindBillingIssueByTypeForOrg(ctx, orgID, database.BillingIssueTypeSubscriptionCancelled)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			return fmt.Errorf("failed to find billing errors: %w", err)
		}
	}

	if bisc != nil {
		err = s.DB.DeleteBillingIssue(ctx, bisc.ID)
		if err != nil {
			return fmt.Errorf("failed to delete billing error: %w", err)
		}
	}

	return nil
}

func (s *Service) CheckBlockingBillingErrors(ctx context.Context, orgID string) error {
	be, err := s.DB.FindBillingIssueByTypeForOrg(ctx, orgID, database.BillingIssueTypeTrialEnded)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			return err
		}
	}

	if be != nil {
		return fmt.Errorf("trial has ended")
	}

	be, err = s.DB.FindBillingIssueByTypeForOrg(ctx, orgID, database.BillingIssueTypePaymentFailed)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			return err
		}
	}

	if be != nil {
		earliestGracePeriodEndDate := time.Time{}
		invoices := be.Metadata.(*database.BillingIssueMetadataPaymentFailed).Invoices
		for _, inv := range invoices {
			if inv.GracePeriodEndDate.Before(earliestGracePeriodEndDate) || earliestGracePeriodEndDate.IsZero() {
				earliestGracePeriodEndDate = inv.GracePeriodEndDate
			}
		}

		if earliestGracePeriodEndDate.AddDate(0, 0, 1).After(time.Now()) || earliestGracePeriodEndDate.IsZero() {
			return fmt.Errorf("payment overdue")
		}
	}

	be, err = s.DB.FindBillingIssueByTypeForOrg(ctx, orgID, database.BillingIssueTypeSubscriptionCancelled)
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
