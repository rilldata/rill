package river

import (
	"context"
	"errors"

	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/billing"
	"github.com/rilldata/rill/admin/database"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
)

type PurgeOrgArgs struct {
	OrgID string
}

func (PurgeOrgArgs) Kind() string { return "purge_org" }

type PurgeOrgWorker struct {
	river.WorkerDefaults[PurgeOrgArgs]
	admin *admin.Service
}

// Work This worker handles the deletion of an organization and all its associated data
func (w *PurgeOrgWorker) Work(ctx context.Context, job *river.Job[PurgeOrgArgs]) error {
	org, err := w.admin.DB.FindOrganization(ctx, job.Args.OrgID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			// org got deleted, ignore
			return nil
		}
		return err
	}

	// cancel subscription
	if org.BillingCustomerID != "" {
		err = w.admin.Biller.CancelSubscriptionsForCustomer(ctx, org.BillingCustomerID, billing.SubscriptionCancellationOptionImmediate)
		if err != nil {
			w.admin.Logger.Error("failed to cancel subscriptions", zap.String("org_id", org.ID), zap.String("org_name", org.Name), zap.Error(err))
		}
		w.admin.Logger.Warn("canceled subscriptions", zap.String("org_id", org.ID), zap.String("org_name", org.Name))
	}

	// clean billing errors and warnings and cancel associated scheduled jobs
	err = w.admin.CleanupTrialBillingErrorsAndWarnings(ctx, org.ID)
	if err != nil {
		return err
	}

	err = w.admin.CleanupBillingErrorSubCancellation(ctx, org.ID)
	if err != nil {
		return err
	}

	// now delete org, other errors and warnings will be cascade deleted
	err = w.admin.DB.DeleteOrganization(ctx, org.Name)
	if err != nil {
		return err
	}

	return nil
}
