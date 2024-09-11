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

const gracePeriodDays = 9

type TrialEndingSoonArgs struct {
	OrgID  string
	SubID  string
	PlanID string
}

func (TrialEndingSoonArgs) Kind() string { return "trial_ending_soon" }

type TrialEndingSoonWorker struct {
	river.WorkerDefaults[TrialEndingSoonArgs]
	admin         *admin.Service
	billingLogger *zap.Logger
}

func (w *TrialEndingSoonWorker) Work(ctx context.Context, job *river.Job[TrialEndingSoonArgs]) error {
	org, err := w.admin.DB.FindOrganization(ctx, job.Args.OrgID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			// org got deleted, ignore
			return nil
		}
		return fmt.Errorf("failed to find organization: %w", err)
	}

	// check if org has subscription cancelled billing error, if yes ignore the job as we put the customer back on default plan on cancellation
	// this is just a cautionary check as we cancel this job on subscription cancellation if scheduled
	be, err := w.admin.DB.FindBillingErrorByType(ctx, org.ID, database.BillingErrorTypeSubscriptionCancelled)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			return fmt.Errorf("failed to find billing errors: %w", err)
		}
	}

	if be != nil {
		return nil
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

	if time.Now().UTC().After(sub[0].TrialEndDate.AddDate(0, 0, 1)) {
		// trial end date has already passed, ignore
		return nil
	}

	// trial period ending soon, log warn and send email
	w.billingLogger.Warn("trial ending soon", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.Time("trial_end_date", sub[0].TrialEndDate))

	// send email
	err = w.admin.Email.SendTrialEndingSoon(&email.TrialEndingSoon{
		ToEmail:      org.BillingEmail,
		ToName:       org.Name,
		OrgName:      org.Name,
		TrialEndDate: sub[0].TrialEndDate,
	})
	if err != nil {
		return fmt.Errorf("failed to send trial ending soon email for org %q: %w", org.Name, err)
	}

	return nil
}

type TrialEndCheckArgs struct {
	OrgID  string
	SubID  string
	PlanID string
}

func (TrialEndCheckArgs) Kind() string { return "trial_end_check" }

type TrialEndCheckWorker struct {
	river.WorkerDefaults[TrialEndCheckArgs]
	admin         *admin.Service
	billingLogger *zap.Logger
}

func (w *TrialEndCheckWorker) Work(ctx context.Context, job *river.Job[TrialEndCheckArgs]) error {
	org, err := w.admin.DB.FindOrganization(ctx, job.Args.OrgID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			// org got deleted, ignore
			return nil
		}
		return fmt.Errorf("failed to find organization: %w", err)
	}

	// check if org has subscription cancelled billing error, if yes ignore the job as we put the customer back on default plan on cancellation
	// this is just a cautionary check as we cancel this job on subscription cancellation if scheduled
	be, err := w.admin.DB.FindBillingErrorByType(ctx, org.ID, database.BillingErrorTypeSubscriptionCancelled)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			return fmt.Errorf("failed to find billing errors: %w", err)
		}
	}

	if be != nil {
		return nil
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

	if time.Now().UTC().Before(sub[0].TrialEndDate.AddDate(0, 0, 1)) {
		return fmt.Errorf("trial end date %s not finished yet for org %q", sub[0].TrialEndDate, org.Name) // will be retried later
	}

	// trial period has ended, log warn, send email and schedule a job to hibernate projects after grace period days if still on trial
	w.billingLogger.Warn("trial period has ended", zap.String("org_id", org.ID), zap.String("org_name", org.Name))

	gracePeriodEndDate := sub[0].TrialEndDate.AddDate(0, 0, gracePeriodDays)
	// schedule a job to check if the org is still on trial after end of grace period date
	j, err := w.admin.Jobs.TrialGracePeriodCheck(ctx, org.ID, sub[0].ID, sub[0].Plan.ID, gracePeriodEndDate)
	if err != nil {
		return fmt.Errorf("failed to schedule trial grace period check job: %w", err)
	}

	_, err = w.admin.DB.UpsertBillingError(ctx, &database.UpsertBillingErrorOptions{
		OrgID: org.ID,
		Type:  database.BillingErrorTypeTrialEnded,
		Metadata: &database.BillingErrorMetadataTrialEnded{
			GracePeriodEndDate:  gracePeriodEndDate,
			GracePeriodEndJobID: j.ID,
		},
		EventTime: sub[0].TrialEndDate.AddDate(0, 0, 1),
	})
	if err != nil {
		return fmt.Errorf("failed to add billing error: %w", err)
	}

	// send email
	err = w.admin.Email.SendTrialEnded(&email.TrialEnded{
		ToEmail:            org.BillingEmail,
		ToName:             org.Name,
		OrgName:            org.Name,
		GracePeriodEndDate: gracePeriodEndDate,
	})
	if err != nil {
		return fmt.Errorf("failed to send trial period ended email for org %q: %w", org.Name, err)
	}

	return nil
}

type TrialGracePeriodCheckArgs struct {
	OrgID              string
	SubID              string
	PlanID             string
	GracePeriodEndDate time.Time
}

func (TrialGracePeriodCheckArgs) Kind() string { return "trial_grace_period_check" }

type TrialGracePeriodCheckWorker struct {
	river.WorkerDefaults[TrialGracePeriodCheckArgs]
	admin         *admin.Service
	billingLogger *zap.Logger
}

func (w *TrialGracePeriodCheckWorker) Work(ctx context.Context, job *river.Job[TrialGracePeriodCheckArgs]) error {
	org, err := w.admin.DB.FindOrganization(ctx, job.Args.OrgID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			// org got deleted, ignore
			return nil
		}
		return fmt.Errorf("failed to find organization: %w", err)
	}

	// check if org has subscription cancelled billing error, if yes ignore the job as we put the customer back on default plan on cancellation
	// this is just a cautionary check as we cancel this job on subscription cancellation if scheduled
	be, err := w.admin.DB.FindBillingErrorByType(ctx, org.ID, database.BillingErrorTypeSubscriptionCancelled)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) {
			return fmt.Errorf("failed to find billing errors: %w", err)
		}
	}

	if be != nil {
		return nil
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
		// subscription or plan have changed, delete the billing error if this was for this job
		be, err := w.admin.DB.FindBillingErrorByType(ctx, org.ID, database.BillingErrorTypeTrialEnded)
		if err != nil {
			if !errors.Is(err, database.ErrNotFound) {
				return fmt.Errorf("failed to find billing errors: %w", err)
			}
		}

		if be != nil {
			meta, ok := be.Metadata.(*database.BillingErrorMetadataTrialEnded)
			if ok && meta.GracePeriodEndJobID == job.ID {
				err = w.admin.DB.DeleteBillingError(ctx, be.ID)
				if err != nil {
					return fmt.Errorf("failed to delete billing error: %w", err)
				}
			}
		}
		return nil
	}

	if time.Now().UTC().Before(job.Args.GracePeriodEndDate.AddDate(0, 0, 1)) {
		return fmt.Errorf("grace period end date %s not finished yet for org %q", sub[0].TrialEndDate.AddDate(0, 0, gracePeriodDays), org.Name) // will be retried later
	}

	// trial grace period has ended, log warn, send email and hibernate projects
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
	w.billingLogger.Warn("projects hibernated due to trial grace period ended", zap.String("org_id", org.ID), zap.String("org_name", org.Name))

	// send email
	err = w.admin.Email.SendTrialGracePeriodEnded(&email.TrialGracePeriodEnded{
		ToEmail: org.BillingEmail,
		ToName:  org.Name,
		OrgName: org.Name,
	})
	if err != nil {
		return fmt.Errorf("failed to send trial grace period ended email for org %q: %w", org.Name, err)
	}

	return nil
}
