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
	admin *admin.Service
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
		return nil
	}

	// delete any trial related billing errors and warnings, irrespective of the new plan.
	err = w.admin.CleanupTrialBillingErrorsAndWarnings(ctx, org.ID)
	if err != nil {
		return fmt.Errorf("failed to cleanup trial billing errors and warnings: %w", err)
	}

	// delete any subscription cancellation billing error
	err = w.admin.CleanupBillingErrorSubCancellation(ctx, org.ID)
	if err != nil {
		return fmt.Errorf("failed to cleanup subscription cancellation errors: %w", err)
	}

	// if the new plan is still a trial plan, schedule trial checks. Can happen if manually assigned new trial plan for example to extend trial period for a customer
	if sub[0].TrialEndDate.After(time.Now().Add(time.Hour * 1)) {
		err = w.admin.ScheduleTrialEndCheckJobs(ctx, org.ID, sub[0].ID, sub[0].Plan.ID, sub[0].StartDate, sub[0].TrialEndDate)
		if err != nil {
			return fmt.Errorf("failed to schedule trial end check job: %w", err)
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

type SubscriptionCancellationArgs struct {
	OrgID      string
	SubID      string
	PlanID     string
	SubEndDate time.Time
}

func (SubscriptionCancellationArgs) Kind() string { return "subscription_cancellation" }

type SubscriptionCancellationWorker struct {
	river.WorkerDefaults[SubscriptionCancellationArgs]
	admin *admin.Service
}

// Work This worker runs at end of the current subscription term after subscription cancellation
func (w *SubscriptionCancellationWorker) Work(ctx context.Context, job *river.Job[SubscriptionCancellationArgs]) error {
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
		return nil
	}

	if time.Now().UTC().Before(job.Args.SubEndDate.AddDate(0, 0, 1)) {
		return fmt.Errorf("subscription end date %s is not finished yet for org %q", job.Args.SubEndDate, org.Name) // will be retried
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

	err = w.admin.Email.SendInformational(&email.Informational{
		ToEmail: org.BillingEmail,
		ToName:  org.Name,
		Subject: "Subscription ended",
		Title:   "",
		Body:    "Thank you for using Rill, all your projects have been hibernated as subscription has ended.",
	})
	if err != nil {
		return fmt.Errorf("failed to send payment method expired email for org %q: %w", org.Name, err)
	}

	w.admin.Logger.Info("projects hibernated due to subscription cancellation", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.String("billing_customer_id", org.BillingCustomerID))
	return nil
}

func valOrDefault[T any](ptr *T, def T) T {
	if ptr != nil {
		return *ptr
	}
	return def
}
