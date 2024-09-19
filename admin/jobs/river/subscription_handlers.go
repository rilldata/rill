package river

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/runtime/pkg/email"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
)

type PlanChangeByAPIArgs struct {
	OrgID     string
	SubID     string
	PlanID    string
	StartDate time.Time // just for deduplication
}

func (PlanChangeByAPIArgs) Kind() string { return "plan_change_by_api" }

type PlanChangeByAPIWorker struct {
	river.WorkerDefaults[PlanChangeByAPIArgs]
	admin  *admin.Service
	logger *zap.Logger
}

// Work This worker handle plan changes when upgrading plan or when we manually assign a new trial plan through admin APIs, does not handle changes done directly in the billing system
func (w *PlanChangeByAPIWorker) Work(ctx context.Context, job *river.Job[PlanChangeByAPIArgs]) error {
	org, err := w.admin.DB.FindOrganization(ctx, job.Args.OrgID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			// org got deleted, ignore
			return nil
		}
		return fmt.Errorf("failed to find organization: %w", err)
	}

	// check if the org has any active subscription
	sub, err := w.admin.Biller.GetSubscriptionsForCustomer(ctx, org.BillingCustomerID)
	if err != nil {
		return fmt.Errorf("failed to get subscriptions for org %q: %w", org.Name, err)
	}

	if len(sub) == 0 {
		return fmt.Errorf("no active subscription found for the org %q", org.Name)
	}

	if len(sub) > 1 {
		return fmt.Errorf("multiple active subscriptions found for the org %q", org.Name)
	}

	if sub[0].ID != job.Args.SubID || sub[0].Plan.ID != job.Args.PlanID {
		// subscription or plan have changed, ignore
		w.logger.Warn("plan change api worker - subscription or plan changed before job could run, doing nothing, please check manually", zap.String("org_id", org.ID), zap.String("org_name", org.Name))
		return nil
	}

	// delete any trial related billing errors and warnings, irrespective of the new plan.
	err = w.admin.CleanupTrialBillingIssues(ctx, org.ID)
	if err != nil {
		return fmt.Errorf("failed to cleanup trial billing errors and warnings: %w", err)
	}

	// delete any subscription cancellation billing error
	err = w.admin.CleanupBillingErrorSubCancellation(ctx, org.ID)
	if err != nil {
		return fmt.Errorf("failed to cleanup subscription cancellation errors: %w", err)
	}

	// if the new plan is still a trial plan, raise on-trial billing issue. Can happen if manually assigned new trial plan for example to extend trial period for a customer
	if sub[0].TrialEndDate.After(time.Now().AddDate(0, 0, 1)) {
		_, err = w.admin.DB.UpsertBillingIssue(ctx, &database.UpsertBillingIssueOptions{
			OrgID: org.ID,
			Type:  database.BillingIssueTypeOnTrial,
			Metadata: &database.BillingIssueMetadataOnTrial{
				SubID:   sub[0].ID,
				PlanID:  sub[0].Plan.ID,
				EndDate: sub[0].TrialEndDate,
			},
			EventTime: sub[0].StartDate,
		})
		if err != nil {
			return fmt.Errorf("failed to upsert billing warning: %w", err)
		}
	}

	// update quotas
	_, err = w.admin.DB.UpdateOrganization(ctx, org.ID, &database.UpdateOrganizationOptions{
		Name:                                org.Name,
		DisplayName:                         org.DisplayName,
		Description:                         org.Description,
		CustomDomain:                        org.CustomDomain,
		QuotaProjects:                       valOrDefault(sub[0].Plan.Quotas.NumProjects, org.QuotaProjects),
		QuotaDeployments:                    valOrDefault(sub[0].Plan.Quotas.NumDeployments, org.QuotaDeployments),
		QuotaSlotsTotal:                     valOrDefault(sub[0].Plan.Quotas.NumSlotsTotal, org.QuotaSlotsTotal),
		QuotaSlotsPerDeployment:             valOrDefault(sub[0].Plan.Quotas.NumSlotsPerDeployment, org.QuotaSlotsPerDeployment),
		QuotaOutstandingInvites:             valOrDefault(sub[0].Plan.Quotas.NumOutstandingInvites, org.QuotaOutstandingInvites),
		QuotaStorageLimitBytesPerDeployment: valOrDefault(sub[0].Plan.Quotas.StorageLimitBytesPerDeployment, org.QuotaStorageLimitBytesPerDeployment),
		BillingCustomerID:                   org.BillingCustomerID,
		PaymentCustomerID:                   org.PaymentCustomerID,
		BillingEmail:                        org.BillingEmail,
	})
	if err != nil {
		return err
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
	cancelled, err := w.admin.DB.FindBillingIssueByTypeNotOverdueProcessed(ctx, database.BillingIssueTypeSubscriptionCancelled)
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
		sub, err := w.admin.Biller.GetSubscriptionsForCustomer(ctx, org.BillingCustomerID)
		if err != nil {
			return fmt.Errorf("failed to get subscriptions for org %q: %w", org.Name, err)
		}

		if len(sub) == 0 {
			return fmt.Errorf("no active subscription found for the org %q", org.Name)
		}

		if len(sub) > 1 {
			return fmt.Errorf("multiple active subscriptions found for the org %q", org.Name)
		}

		// update quotas to the default plan and hibernate all projects
		_, err = w.admin.DB.UpdateOrganization(ctx, org.ID, &database.UpdateOrganizationOptions{
			Name:                                org.Name,
			DisplayName:                         org.DisplayName,
			Description:                         org.Description,
			CustomDomain:                        org.CustomDomain,
			QuotaProjects:                       valOrDefault(sub[0].Plan.Quotas.NumProjects, org.QuotaProjects),
			QuotaDeployments:                    valOrDefault(sub[0].Plan.Quotas.NumDeployments, org.QuotaDeployments),
			QuotaSlotsTotal:                     valOrDefault(sub[0].Plan.Quotas.NumSlotsTotal, org.QuotaSlotsTotal),
			QuotaSlotsPerDeployment:             valOrDefault(sub[0].Plan.Quotas.NumSlotsPerDeployment, org.QuotaSlotsPerDeployment),
			QuotaOutstandingInvites:             valOrDefault(sub[0].Plan.Quotas.NumOutstandingInvites, org.QuotaOutstandingInvites),
			QuotaStorageLimitBytesPerDeployment: valOrDefault(sub[0].Plan.Quotas.StorageLimitBytesPerDeployment, org.QuotaStorageLimitBytesPerDeployment),
			BillingCustomerID:                   org.BillingCustomerID,
			PaymentCustomerID:                   org.PaymentCustomerID,
			BillingEmail:                        org.BillingEmail,
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

func valOrDefault[T any](ptr *T, def T) T {
	if ptr != nil {
		return *ptr
	}
	return def
}
