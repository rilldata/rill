package river

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/billing"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/runtime/pkg/email"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
)

type TrialEndingSoonArgs struct{}

func (TrialEndingSoonArgs) Kind() string { return "trial_ending_soon" }

type TrialEndingSoonWorker struct {
	river.WorkerDefaults[TrialEndingSoonArgs]
	admin  *admin.Service
	logger *zap.Logger
	now    func() time.Time
}

func (w *TrialEndingSoonWorker) Work(ctx context.Context, job *river.Job[TrialEndingSoonArgs]) error {
	onTrialOrgs, err := w.admin.DB.FindBillingIssueByTypeAndOverdueProcessed(ctx, database.BillingIssueTypeOnTrial, false)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			// no orgs have this billing issue
			return nil
		}
		return fmt.Errorf("failed to find organization with billing issue: %w", err)
	}

	now := billingLifecycleNow(w.now)
	var workErr error
	for _, o := range onTrialOrgs {
		m := o.Metadata.(*database.BillingIssueMetadataOnTrial)
		if now.Before(m.EndDate.AddDate(0, 0, -7)) {
			// trial end date is more than 7 days away, move to next org
			continue
		}

		if err := w.processIssue(ctx, o, m, now); err != nil {
			workErr = errors.Join(workErr, fmt.Errorf("process trial ending soon for org %q: %w", o.OrgID, err))
		}
	}

	return workErr
}

func (w *TrialEndingSoonWorker) processIssue(ctx context.Context, issue *database.BillingIssue, metadata *database.BillingIssueMetadataOnTrial, now time.Time) error {
	org, err := w.admin.DB.FindOrganization(ctx, issue.OrgID)
	if err != nil {
		return fmt.Errorf("failed to find organization: %w", err)
	}

	daysRemaining := max(int(metadata.EndDate.Sub(now).Hours()/24), 0)
	projects, err := w.admin.DB.CountProjectsForOrganization(ctx, org.ID)
	if err != nil {
		return fmt.Errorf("failed to count projects for org %q: %w", org.Name, err)
	}

	w.logger.Warn("trial ending soon",
		zap.String("org_id", org.ID),
		zap.String("org_name", org.Name),
		zap.Time("trial_end_date", metadata.EndDate),
		zap.String("user_email", org.BillingEmail),
		zap.Int("count_of_projects", projects),
		zap.Int("count_of_days_remaining", daysRemaining),
	)

	// The current email sender has no idempotency key. Persisting the marker first
	// intentionally gives lifecycle notifications at-most-once attempt semantics.
	if err := w.admin.DB.UpdateBillingIssueOverdueAsProcessed(ctx, issue.ID); err != nil {
		return fmt.Errorf("failed to update billing issue as processed: %w", err)
	}

	err = w.admin.Email.SendTrialEndingSoon(&email.TrialEndingSoon{
		ToEmail:      org.BillingEmail,
		ToName:       org.Name,
		OrgName:      org.Name,
		UpgradeURL:   w.admin.URLs.Billing(org.Name, true),
		TrialEndDate: metadata.EndDate,
	})
	if err != nil {
		w.logger.Error("failed to send trial ending soon email after marking it attempted", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.Error(err))
	}
	return nil
}

type TrialEndCheckArgs struct{}

func (TrialEndCheckArgs) Kind() string { return "trial_end_check" }

type TrialEndCheckWorker struct {
	river.WorkerDefaults[TrialEndCheckArgs]
	admin  *admin.Service
	logger *zap.Logger
	now    func() time.Time
}

func (w *TrialEndCheckWorker) Work(ctx context.Context, job *river.Job[TrialEndCheckArgs]) error {
	onTrialOrgs, err := w.admin.DB.FindBillingIssueByTypeAndOverdueProcessed(ctx, database.BillingIssueTypeOnTrial, true)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			// no orgs have this billing issue
			return nil
		}
		return fmt.Errorf("failed to find organization with billing issue: %w", err)
	}

	now := billingLifecycleNow(w.now)
	var workErr error
	for _, o := range onTrialOrgs {
		m := o.Metadata.(*database.BillingIssueMetadataOnTrial)
		if now.Before(m.EndDate) {
			// trial end date is not finished yet, move to next org
			continue
		}

		if err := w.processIssue(ctx, o, m); err != nil {
			workErr = errors.Join(workErr, fmt.Errorf("process trial end for org %q: %w", o.OrgID, err))
		}
	}

	return workErr
}

func (w *TrialEndCheckWorker) processIssue(ctx context.Context, issue *database.BillingIssue, metadata *database.BillingIssueMetadataOnTrial) error {
	org, err := w.admin.DB.FindOrganization(ctx, issue.OrgID)
	if err != nil {
		return fmt.Errorf("failed to find organization: %w", err)
	}

	sub, err := w.admin.Biller.GetActiveSubscription(ctx, org.BillingCustomerID)
	if err != nil {
		return fmt.Errorf("failed to get subscriptions for org %q: %w", org.Name, err)
	}
	if sub == nil || sub.Plan == nil {
		return fmt.Errorf("active subscription for org %q is incomplete", org.Name)
	}
	if sub.ID != metadata.SubID || sub.Plan.ID != metadata.PlanID {
		w.logger.Warn("trial period has ended, but org has different active subscription; removing stale trial state", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.String("sub_id", sub.ID), zap.String("sub_plan_id", sub.Plan.ID), zap.String("expected_sub_id", metadata.SubID), zap.String("expected_sub_plan_id", metadata.PlanID))
		if err := w.admin.DB.DeleteBillingIssue(ctx, issue.ID); err != nil {
			return fmt.Errorf("failed to delete stale on-trial billing issue: %w", err)
		}
		return nil
	}

	projects, err := w.admin.DB.CountProjectsForOrganization(ctx, org.ID)
	if err != nil {
		return fmt.Errorf("failed to count projects for org %q: %w", org.Name, err)
	}
	w.logger.Warn("trial period has ended",
		zap.String("org_id", org.ID),
		zap.String("org_name", org.Name),
		zap.String("user_email", org.BillingEmail),
		zap.Int("count_of_projects", projects),
	)

	cctx, tx, err := w.admin.DB.NewTx(ctx, false)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	rollback := func(cause error) error {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return errors.Join(cause, fmt.Errorf("failed to rollback transaction: %w", rollbackErr))
		}
		return cause
	}

	_, err = w.admin.DB.UpsertBillingIssue(cctx, &database.UpsertBillingIssueOptions{
		OrgID: org.ID,
		Type:  database.BillingIssueTypeTrialEnded,
		Metadata: &database.BillingIssueMetadataTrialEnded{
			SubID:              metadata.SubID,
			PlanID:             metadata.PlanID,
			EndDate:            metadata.EndDate,
			GracePeriodEndDate: metadata.GracePeriodEndDate,
		},
		EventTime: metadata.EndDate,
	})
	if err != nil {
		return rollback(fmt.Errorf("failed to add billing issue: %w", err))
	}
	if err := w.admin.DB.DeleteBillingIssue(cctx, issue.ID); err != nil {
		return rollback(fmt.Errorf("failed to delete billing issue: %w", err))
	}

	// Commit the state transition before attempting email. This makes a commit
	// failure retryable without duplicating an externally visible notification.
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	err = w.admin.Email.SendTrialEnded(&email.TrialEnded{
		ToEmail:            org.BillingEmail,
		ToName:             org.Name,
		OrgName:            org.Name,
		UpgradeURL:         w.admin.URLs.Billing(org.Name, true),
		GracePeriodEndDate: metadata.GracePeriodEndDate,
	})
	if err != nil {
		w.logger.Error("failed to send trial ended email after committing it as attempted", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.Error(err))
	}
	return nil
}

type TrialGracePeriodCheckArgs struct{}

func (TrialGracePeriodCheckArgs) Kind() string { return "trial_grace_period_check" }

type TrialGracePeriodCheckWorker struct {
	river.WorkerDefaults[TrialGracePeriodCheckArgs]
	admin  *admin.Service
	logger *zap.Logger
	now    func() time.Time
}

func (w *TrialGracePeriodCheckWorker) Work(ctx context.Context, job *river.Job[TrialGracePeriodCheckArgs]) error {
	trailEndedOrgs, err := w.admin.DB.FindBillingIssueByTypeAndOverdueProcessed(ctx, database.BillingIssueTypeTrialEnded, false)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			// no orgs have this billing issue
			return nil
		}
		return fmt.Errorf("failed to find organization with billing issue: %w", err)
	}

	now := billingLifecycleNow(w.now)
	var workErr error
	for _, o := range trailEndedOrgs {
		m := o.Metadata.(*database.BillingIssueMetadataTrialEnded)
		if now.Before(m.GracePeriodEndDate) {
			// grace period end date is not finished yet, move to next org
			continue
		}

		if err := w.processIssue(ctx, o, m); err != nil {
			workErr = errors.Join(workErr, fmt.Errorf("process trial grace period end for org %q: %w", o.OrgID, err))
		}
	}

	return workErr
}

func (w *TrialGracePeriodCheckWorker) processIssue(ctx context.Context, issue *database.BillingIssue, metadata *database.BillingIssueMetadataTrialEnded) error {
	org, err := w.admin.DB.FindOrganization(ctx, issue.OrgID)
	if err != nil {
		return fmt.Errorf("failed to find organization: %w", err)
	}

	sub, err := w.admin.Biller.GetActiveSubscription(ctx, org.BillingCustomerID)
	if err != nil && !errors.Is(err, billing.ErrNotFound) {
		return fmt.Errorf("failed to get subscriptions for org %q: %w", org.Name, err)
	}
	if sub == nil {
		w.logger.Warn("trial grace period has ended, but org has no active subscription", zap.String("org_id", org.ID), zap.String("org_name", org.Name))
	} else {
		if sub.Plan == nil {
			return fmt.Errorf("active subscription for org %q has no plan", org.Name)
		}
		if sub.ID != metadata.SubID || sub.Plan.ID != metadata.PlanID {
			w.logger.Warn("trial grace period has ended, but org has a replacement subscription; retiring stale trial issue", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.String("sub_id", sub.ID), zap.String("sub_plan_id", sub.Plan.ID), zap.String("expected_sub_id", metadata.SubID), zap.String("expected_sub_plan_id", metadata.PlanID))
			if err := w.admin.DB.UpdateBillingIssueOverdueAsProcessed(ctx, issue.ID); err != nil {
				return fmt.Errorf("failed to retire stale trial-ended issue: %w", err)
			}
			return nil
		}
	}

	if _, err := w.admin.Biller.CancelSubscriptionsForCustomer(ctx, org.BillingCustomerID, billing.SubscriptionCancellationOptionImmediate); err != nil {
		return fmt.Errorf("failed to cancel subscription for org %q: %w", org.Name, err)
	}
	if err := zeroOrganizationQuotas(ctx, w.admin, org); err != nil {
		return err
	}
	if _, err := hibernateProjectsForOrganization(ctx, w.admin, org.ID); err != nil {
		return err
	}
	w.logger.Warn("projects hibernated due to trial grace period ended", zap.String("org_id", org.ID), zap.String("org_name", org.Name))

	// The processed marker is the durable at-most-once notification marker.
	if err := w.admin.DB.UpdateBillingIssueOverdueAsProcessed(ctx, issue.ID); err != nil {
		return fmt.Errorf("failed to update billing issue as processed: %w", err)
	}
	err = w.admin.Email.SendTrialGracePeriodEnded(&email.TrialGracePeriodEnded{
		ToEmail:    org.BillingEmail,
		ToName:     org.Name,
		OrgName:    org.Name,
		UpgradeURL: w.admin.URLs.Billing(org.Name, true),
	})
	if err != nil {
		w.logger.Error("failed to send trial grace period ended email after marking it attempted", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.Error(err))
	}
	return nil
}

func billingLifecycleNow(now func() time.Time) time.Time {
	if now != nil {
		return now().UTC()
	}
	return time.Now().UTC()
}

func zeroOrganizationQuotas(ctx context.Context, adm *admin.Service, org *database.Organization) error {
	_, err := adm.DB.UpdateOrganization(ctx, org.ID, &database.UpdateOrganizationOptions{
		Name:                                org.Name,
		DisplayName:                         org.DisplayName,
		Description:                         org.Description,
		LogoAssetID:                         org.LogoAssetID,
		LogoDarkAssetID:                     org.LogoDarkAssetID,
		FaviconAssetID:                      org.FaviconAssetID,
		ThumbnailAssetID:                    org.ThumbnailAssetID,
		CustomDomain:                        org.CustomDomain,
		DefaultProjectRoleID:                org.DefaultProjectRoleID,
		QuotaProjects:                       0,
		QuotaDeployments:                    0,
		QuotaSlotsTotal:                     0,
		QuotaSlotsPerDeployment:             0,
		QuotaOutstandingInvites:             0,
		QuotaStorageLimitBytesPerDeployment: 0,
		QuotaSeats:                          0,
		BillingCustomerID:                   org.BillingCustomerID,
		PaymentCustomerID:                   org.PaymentCustomerID,
		BillingEmail:                        org.BillingEmail,
		BillingPlanName:                     org.BillingPlanName,
		BillingPlanDisplayName:              org.BillingPlanDisplayName,
		CreatedByUserID:                     org.CreatedByUserID,
	})
	if err != nil {
		return fmt.Errorf("failed to zero quotas for org %q: %w", org.Name, err)
	}
	return nil
}

func hibernateProjectsForOrganization(ctx context.Context, adm *admin.Service, orgID string) (int, error) {
	const limit = 10
	afterProjectName := ""
	projectCount := 0
	for {
		projects, err := adm.DB.FindProjectsForOrganization(ctx, orgID, afterProjectName, limit)
		if err != nil {
			return projectCount, fmt.Errorf("failed to find projects after %q: %w", afterProjectName, err)
		}

		for _, project := range projects {
			if _, err := adm.HibernateProject(ctx, project); err != nil {
				return projectCount, fmt.Errorf("failed to hibernate project %q: %w", project.Name, err)
			}
			// Advance only after the project's durable hibernation update succeeds.
			// Retries may revisit earlier pages; HibernateProject is idempotent.
			afterProjectName = project.Name
			projectCount++
		}

		if len(projects) < limit {
			return projectCount, nil
		}
	}
}
