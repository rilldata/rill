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

type HandlePlanChangeBillingIssuesArgs struct {
	OrgID     string
	SubID     string
	PlanID    string
	StartDate time.Time // just for deduplication
}

func (HandlePlanChangeBillingIssuesArgs) Kind() string { return "handle_plan_change_billing_issues" }

type HandlePlanChangeBillingIssues struct {
	river.WorkerDefaults[HandlePlanChangeBillingIssuesArgs]
	admin  *admin.Service
	logger *zap.Logger
}

// Work handles the billing issues for the org after a plan change
func (w *HandlePlanChangeBillingIssues) Work(ctx context.Context, job *river.Job[HandlePlanChangeBillingIssuesArgs]) error {
	org, err := w.admin.DB.FindOrganization(ctx, job.Args.OrgID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			// org got deleted, ignore
			return nil
		}
		return fmt.Errorf("failed to find organization: %w", err)
	}

	// check if the org has any active subscription
	sub, err := w.admin.Biller.GetActiveSubscription(ctx, org.BillingCustomerID)
	if err != nil {
		return fmt.Errorf("failed to get subscriptions for org %q: %w", org.Name, err)
	}

	if sub.ID != job.Args.SubID || sub.Plan.ID != job.Args.PlanID {
		// subscription or plan have changed, ignore
		w.logger.Warn("plan change api worker - subscription or plan changed before job could run, doing nothing, please check manually", zap.String("org_id", org.ID), zap.String("org_name", org.Name))
		return nil
	}

	// delete any trial related billing issues, irrespective of the new plan.
	err = w.admin.CleanupTrialBillingIssues(ctx, org.ID)
	if err != nil {
		w.logger.Error("failed to cleanup trial billing issues", zap.String("org_id", org.ID), zap.Error(err))
		return fmt.Errorf("failed to cleanup trial billing errors and warnings: %w", err)
	}

	// delete any subscription related billing issues
	err = w.admin.CleanupSubscriptionBillingIssues(ctx, org.ID)
	if err != nil {
		w.logger.Error("failed to cleanup subscription billing issues", zap.String("org_id", org.ID), zap.Error(err))
		return fmt.Errorf("failed to cleanup subscription cancellation errors: %w", err)
	}

	// if the new plan is still a trial plan, raise on-trial billing issue. Can happen if manually assigned new trial plan for example to extend trial period for a customer
	if sub.TrialEndDate.After(time.Now().Truncate(24*time.Hour).AddDate(0, 0, 1)) {
		_, err = w.admin.DB.UpsertBillingIssue(ctx, &database.UpsertBillingIssueOptions{
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
			return fmt.Errorf("failed to upsert billing warning: %w", err)
		}
	}

	return nil
}

type SubscriptionCancellationCheckArgs struct{}

func (SubscriptionCancellationCheckArgs) Kind() string { return "subscription_cancellation_check" }

type SubscriptionCancellationCheckWorker struct {
	river.WorkerDefaults[SubscriptionCancellationCheckArgs]
	admin  *admin.Service
	logger *zap.Logger
}

// Work This worker runs at end of the current subscription term after subscription cancellation
func (w *SubscriptionCancellationCheckWorker) Work(ctx context.Context, job *river.Job[SubscriptionCancellationCheckArgs]) error {
	return work(ctx, w.admin.Logger, job.Kind, w.subscriptionCancellationCheck)
}

func (w *SubscriptionCancellationCheckWorker) subscriptionCancellationCheck(ctx context.Context) error {
	cancelled, err := w.admin.DB.FindBillingIssueByTypeAndOverdueProcessed(ctx, database.BillingIssueTypeSubscriptionCancelled, false)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			// no orgs have this billing issue
			return nil
		}
		return fmt.Errorf("failed to find orgs with subscription cancellation billing issue: %w", err)
	}

	for _, issue := range cancelled {
		m := issue.Metadata.(*database.BillingIssueMetadataSubscriptionCancelled)
		if time.Now().UTC().Before(m.EndDate.AddDate(0, 0, 1)) {
			// subscription end date is not finished yet, continue to next org
			continue
		}

		org, err := w.admin.DB.FindOrganization(ctx, issue.OrgID)
		if err != nil {
			return fmt.Errorf("failed to find organization: %w", err)
		}

		// check if the org has any active subscription
		sub, err := w.admin.Biller.GetActiveSubscription(ctx, org.BillingCustomerID)
		if err != nil {
			if !errors.Is(err, billing.ErrNotFound) {
				return fmt.Errorf("failed to get subscriptions for org %q: %w", org.Name, err)
			}
		}

		if sub != nil {
			w.logger.Warn("active subscription found for the org even after sub cancellation, skipping hibernation", zap.String("org_id", org.ID), zap.String("org_name", org.Name))
			return fmt.Errorf("active subscription found for the org %q", org.Name)
		}

		// update quotas to 0 and hibernate all projects
		_, err = w.admin.DB.UpdateOrganization(ctx, org.ID, &database.UpdateOrganizationOptions{
			Name:                                org.Name,
			DisplayName:                         org.DisplayName,
			Description:                         org.Description,
			CustomDomain:                        org.CustomDomain,
			QuotaProjects:                       0,
			QuotaDeployments:                    0,
			QuotaSlotsTotal:                     0,
			QuotaSlotsPerDeployment:             0,
			QuotaOutstandingInvites:             0,
			QuotaStorageLimitBytesPerDeployment: 0,
			BillingCustomerID:                   org.BillingCustomerID,
			PaymentCustomerID:                   org.PaymentCustomerID,
			BillingEmail:                        org.BillingEmail,
			CreatedByUserID:                     org.CreatedByUserID,
		})
		if err != nil {
			return err
		}

		// hibernate projects
		limit := 10
		afterProjectName := ""
		for {
			projs, err := w.admin.DB.FindProjectsForOrganization(ctx, org.ID, afterProjectName, limit)
			if err != nil {
				return err
			}

			for _, proj := range projs {
				_, err = w.admin.HibernateProject(ctx, proj)
				if err != nil {
					return fmt.Errorf("failed to hibernate project %q: %w", proj.Name, err)
				}
				afterProjectName = proj.Name
			}

			if len(projs) < limit {
				break
			}
		}

		err = w.admin.Email.SendSubscriptionEnded(&email.SubscriptionEnded{
			ToEmail: org.BillingEmail,
			ToName:  org.Name,
			OrgName: org.Name,
		})
		if err != nil {
			w.logger.Error("failed to send subscription ended email", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.Error(err))
		}

		w.logger.Warn("projects hibernated due to subscription cancellation", zap.String("org_id", org.ID), zap.String("org_name", org.Name))

		// mark the billing issue as processed
		err = w.admin.DB.UpdateBillingIssueOverdueAsProcessed(ctx, issue.ID)
		if err != nil {
			return fmt.Errorf("failed to update billing issue as processed: %w", err)
		}
	}

	return nil
}
