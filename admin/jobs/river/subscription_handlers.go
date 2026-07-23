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

type SubscriptionCancellationCheckArgs struct{}

func (SubscriptionCancellationCheckArgs) Kind() string { return "subscription_cancellation_check" }

type SubscriptionCancellationCheckWorker struct {
	river.WorkerDefaults[SubscriptionCancellationCheckArgs]
	admin  *admin.Service
	logger *zap.Logger
	now    func() time.Time
}

// Work This worker runs at end of the current subscription term after subscription cancellation
func (w *SubscriptionCancellationCheckWorker) Work(ctx context.Context, job *river.Job[SubscriptionCancellationCheckArgs]) error {
	cancelled, err := w.admin.DB.FindBillingIssueByTypeAndOverdueProcessed(ctx, database.BillingIssueTypeSubscriptionCancelled, false)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			// no orgs have this billing issue
			return nil
		}
		return fmt.Errorf("failed to find orgs with subscription cancellation billing issue: %w", err)
	}

	now := billingLifecycleNow(w.now)
	var workErr error
	for _, issue := range cancelled {
		m := issue.Metadata.(*database.BillingIssueMetadataSubscriptionCancelled)
		if now.Before(m.EndDate.AddDate(0, 0, 1)) {
			// subscription end date is not finished yet, continue to next org
			continue
		}

		if err := w.processIssue(ctx, issue); err != nil {
			workErr = errors.Join(workErr, fmt.Errorf("process subscription cancellation for org %q: %w", issue.OrgID, err))
		}
	}

	return workErr
}

func (w *SubscriptionCancellationCheckWorker) processIssue(ctx context.Context, issue *database.BillingIssue) error {
	org, err := w.admin.DB.FindOrganization(ctx, issue.OrgID)
	if err != nil {
		return fmt.Errorf("failed to find organization: %w", err)
	}

	sub, err := w.admin.Biller.GetActiveSubscription(ctx, org.BillingCustomerID)
	if err != nil && !errors.Is(err, billing.ErrNotFound) {
		return fmt.Errorf("failed to get subscriptions for org %q: %w", org.Name, err)
	}
	if sub != nil {
		// A replacement subscription makes this cancellation issue stale. Retire it
		// and continue so it cannot block cancellation checks for later organizations.
		w.logger.Warn("active replacement subscription found after cancellation; skipping hibernation", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.String("subscription_id", sub.ID))
		if err := w.admin.DB.UpdateBillingIssueOverdueAsProcessed(ctx, issue.ID); err != nil {
			return fmt.Errorf("failed to retire stale subscription cancellation issue: %w", err)
		}
		return nil
	}

	if err := zeroOrganizationQuotas(ctx, w.admin, org); err != nil {
		return err
	}
	projectCount, err := hibernateProjectsForOrganization(ctx, w.admin, org.ID)
	if err != nil {
		return err
	}

	w.logger.Warn("projects hibernated due to subscription cancellation",
		zap.String("org_id", org.ID),
		zap.String("org_name", org.Name),
		zap.Int("count_of_projects", projectCount),
		zap.String("user_email", org.BillingEmail),
	)

	// Persist completion before SMTP to prevent a failed processed-marker update
	// from causing a duplicate email on retry.
	if err := w.admin.DB.UpdateBillingIssueOverdueAsProcessed(ctx, issue.ID); err != nil {
		return fmt.Errorf("failed to update billing issue as processed: %w", err)
	}
	err = w.admin.Email.SendSubscriptionEnded(&email.SubscriptionEnded{
		ToEmail:    org.BillingEmail,
		ToName:     org.Name,
		OrgName:    org.Name,
		BillingURL: w.admin.URLs.Billing(org.Name, false),
	})
	if err != nil {
		w.logger.Error("failed to send subscription ended email after marking it attempted", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.String("billing_email", org.BillingEmail), zap.Error(err))
	}
	return nil
}
