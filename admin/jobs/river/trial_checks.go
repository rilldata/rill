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

type TrialEndingSoonArgs struct{}

func (TrialEndingSoonArgs) Kind() string { return "trial_ending_soon" }

type TrialEndingSoonWorker struct {
	river.WorkerDefaults[TrialEndingSoonArgs]
	admin  *admin.Service
	logger *zap.Logger
}

func (w *TrialEndingSoonWorker) Work(ctx context.Context, job *river.Job[TrialEndingSoonArgs]) error {
	return work(ctx, w.admin.Logger, job.Kind, w.trialEndingSoon)
}

func (w *TrialEndingSoonWorker) trialEndingSoon(ctx context.Context) error {
	onTrialOrgs, err := w.admin.DB.FindBillingIssueByTypeNotOverdueProcessed(ctx, database.BillingIssueTypeOnTrial)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			// no orgs have this billing issue
			return nil
		}
		return fmt.Errorf("failed to find organization with billing issue: %w", err)
	}

	for _, o := range onTrialOrgs {
		m := o.Metadata.(*database.BillingIssueMetadataOnTrial)
		if time.Now().UTC().Before(m.EndDate.AddDate(0, 0, -7)) {
			// trial end date is more than 7 days away, move to next org
			continue
		}

		// trial period ending soon, log warn and send email
		org, err := w.admin.DB.FindOrganization(ctx, o.OrgID)
		if err != nil {
			return fmt.Errorf("failed to find organization: %w", err)
		}

		w.logger.Warn("trial ending soon", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.Time("trial_end_date", m.EndDate))

		err = w.admin.Email.SendTrialEndingSoon(&email.TrialEndingSoon{
			ToEmail:      org.BillingEmail,
			ToName:       org.Name,
			OrgName:      org.Name,
			TrialEndDate: m.EndDate,
		})
		if err != nil {
			return fmt.Errorf("failed to send trial ending soon email for org %q: %w", org.Name, err)
		}

		// mark the billing issue as processed
		err = w.admin.DB.UpdateBillingIssueOverdueAsProcessed(ctx, o.ID)
		if err != nil {
			return fmt.Errorf("failed to update billing issue as processed: %w", err)
		}
	}

	return nil
}

type TrialEndCheckArgs struct{}

func (TrialEndCheckArgs) Kind() string { return "trial_end_check" }

type TrialEndCheckWorker struct {
	river.WorkerDefaults[TrialEndCheckArgs]
	admin  *admin.Service
	logger *zap.Logger
}

func (w *TrialEndCheckWorker) Work(ctx context.Context, job *river.Job[TrialEndCheckArgs]) error {
	return work(ctx, w.admin.Logger, job.Kind, w.trialEndCheck)
}

func (w *TrialEndCheckWorker) trialEndCheck(ctx context.Context) error {
	onTrialOrgs, err := w.admin.DB.FindBillingIssueByType(ctx, database.BillingIssueTypeOnTrial)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			// no orgs have this billing issue
			return nil
		}
		return fmt.Errorf("failed to find organization with billing issue: %w", err)
	}

	for _, o := range onTrialOrgs {
		m := o.Metadata.(*database.BillingIssueMetadataOnTrial)
		if time.Now().UTC().Before(m.EndDate.AddDate(0, 0, 1)) {
			// trial end date is not finished yet, move to next org
			continue
		}

		// trial period has ended, log warn and send email
		org, err := w.admin.DB.FindOrganization(ctx, o.OrgID)
		if err != nil {
			return fmt.Errorf("failed to find organization: %w", err)
		}

		w.logger.Warn("trial period has ended", zap.String("org_id", org.ID), zap.String("org_name", org.Name))

		gracePeriodEndDate := m.EndDate.AddDate(0, 0, gracePeriodDays)

		cctx, tx, err := w.admin.DB.NewTx(ctx)
		if err != nil {
			return fmt.Errorf("failed to start transaction: %w", err)
		}

		_, err = w.admin.DB.UpsertBillingIssue(cctx, &database.UpsertBillingIssueOptions{
			OrgID: org.ID,
			Type:  database.BillingIssueTypeTrialEnded,
			Metadata: &database.BillingIssueMetadataTrialEnded{
				GracePeriodEndDate: gracePeriodEndDate,
			},
			EventTime: m.EndDate.AddDate(0, 0, 1),
		})
		if err != nil {
			err = tx.Rollback()
			if err != nil {
				return fmt.Errorf("failed to rollback transaction: %w", err)
			}
			return fmt.Errorf("failed to add billing error: %w", err)
		}

		// delete the on-trial billing issue
		err = w.admin.DB.DeleteBillingIssue(cctx, o.ID)
		if err != nil {
			err = tx.Rollback()
			if err != nil {
				return fmt.Errorf("failed to rollback transaction: %w", err)
			}
			return fmt.Errorf("failed to delete billing issue: %w", err)
		}

		// send email
		err = w.admin.Email.SendTrialEnded(&email.TrialEnded{
			ToEmail:            org.BillingEmail,
			ToName:             org.Name,
			OrgName:            org.Name,
			GracePeriodEndDate: gracePeriodEndDate,
		})
		if err != nil {
			err = tx.Rollback()
			if err != nil {
				return fmt.Errorf("failed to rollback transaction: %w", err)
			}
			return fmt.Errorf("failed to send trial period ended email for org %q: %w", org.Name, err)
		}

		err = tx.Commit()
		if err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}
	}

	return nil
}

type TrialGracePeriodCheckArgs struct{}

func (TrialGracePeriodCheckArgs) Kind() string { return "trial_grace_period_check" }

type TrialGracePeriodCheckWorker struct {
	river.WorkerDefaults[TrialGracePeriodCheckArgs]
	admin  *admin.Service
	logger *zap.Logger
}

func (w *TrialGracePeriodCheckWorker) Work(ctx context.Context, job *river.Job[TrialGracePeriodCheckArgs]) error {
	return work(ctx, w.admin.Logger, job.Kind, w.trialGracePeriodCheck)
}

func (w *TrialGracePeriodCheckWorker) trialGracePeriodCheck(ctx context.Context) error {
	trailEndedOrgs, err := w.admin.DB.FindBillingIssueByTypeNotOverdueProcessed(ctx, database.BillingIssueTypeTrialEnded)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			// no orgs have this billing issue
			return nil
		}
		return fmt.Errorf("failed to find organization with billing issue: %w", err)
	}

	for _, o := range trailEndedOrgs {
		m := o.Metadata.(*database.BillingIssueMetadataTrialEnded)
		if time.Now().UTC().Before(m.GracePeriodEndDate.AddDate(0, 0, 1)) {
			// grace period end date is not finished yet, move to next org
			continue
		}

		org, err := w.admin.DB.FindOrganization(ctx, o.OrgID)
		if err != nil {
			return fmt.Errorf("failed to find organization: %w", err)
		}

		// double check - get active subscriptions for the org
		sub, err := w.admin.Biller.GetSubscriptionsForCustomer(ctx, org.BillingCustomerID)
		if err != nil {
			return fmt.Errorf("failed to get subscriptions for org %q: %w", org.Name, err)
		}

		if len(sub) == 0 {
			w.logger.Warn("trial grace period end check - no active subscription found for the org, please check manually", zap.String("org_id", org.ID), zap.String("org_name", org.Name))
			continue
		}

		if len(sub) > 1 {
			w.logger.Warn("trial grace period end check - multiple active subscriptions found for the org, please check manually", zap.String("org_id", org.ID), zap.String("org_name", org.Name))
			continue
		}

		if sub[0].ID != m.SubID || sub[0].Plan.ID != m.PlanID {
			w.logger.Warn("trial grace period end check - subscription or plan changed, but billing issue not updated, doing nothing, please check manually", zap.String("org_id", org.ID), zap.String("org_name", org.Name))
			// subscription or plan have changed, mark the billing issue as processed
			err = w.admin.DB.UpdateBillingIssueOverdueAsProcessed(ctx, o.ID)
			if err != nil {
				return fmt.Errorf("failed to update billing issue as processed: %w", err)
			}
			continue
		}

		// trial grace period has ended, log warn and hibernate projects
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
		w.logger.Warn("projects hibernated due to trial grace period ended", zap.String("org_id", org.ID), zap.String("org_name", org.Name))

		// send email
		err = w.admin.Email.SendTrialGracePeriodEnded(&email.TrialGracePeriodEnded{
			ToEmail: org.BillingEmail,
			ToName:  org.Name,
			OrgName: org.Name,
		})
		if err != nil {
			return fmt.Errorf("failed to send trial grace period ended email for org %q: %w", org.Name, err)
		}

		// mark the billing issue as processed
		err = w.admin.DB.UpdateBillingIssueOverdueAsProcessed(ctx, o.ID)
		if err != nil {
			return fmt.Errorf("failed to update billing issue as processed: %w", err)
		}
	}

	return nil
}
