package river

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
)

type ReconcileDeploymentArgs struct {
	DeploymentID string
}

func (ReconcileDeploymentArgs) Kind() string { return "reconcile_deployment" }

type ReconcileDeploymentWorker struct {
	river.WorkerDefaults[ReconcileDeploymentArgs]
	admin *admin.Service
}

// ReconcileDeploymentWorker is a state machine, it reconciles the state of a deployment based on its desired and current status.
// This job is responsible for transitioning deployments through their lifecycle states,
// such as starting, updating, stopping, and deleting deployments.
// We handle all deployment state transitions in this job to ensure consistency and to avoid concurrent conflicting operations on the same deployment.
func (w *ReconcileDeploymentWorker) Work(ctx context.Context, job *river.Job[ReconcileDeploymentArgs]) error {
	depl, err := w.admin.DB.FindDeployment(ctx, job.Args.DeploymentID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			// If the deployment doesn't exist, we can just finish the job and do nothing more.
			w.admin.Logger.Info("reconcile deployment: deployment not found, job succeeded", observability.ZapCtx(ctx))
			return nil
		}
		return err
	}

	// Capture the DesiredStatusUpdatedOn at the start of the job
	desiredStatusUpdatedOn := depl.DesiredStatusUpdatedOn

	var newStatus database.DeploymentStatus
	switch depl.DesiredStatus {
	case database.DeploymentStatusRunning:
		// Check current status to either start or update the deployment
		if depl.Status == database.DeploymentStatusRunning {
			// Update the deployment status to updating
			depl, err = w.admin.DB.UpdateDeploymentStatus(ctx, depl.ID, database.DeploymentStatusUpdating, "Updating...")
			if err != nil {
				return err
			}

			// Update the deployment by updating its runtime instance and resources.
			err := w.admin.UpdateDeploymentInner(ctx, depl)
			if err != nil {
				return err
			}
		} else {
			// Update the deployment status to pending
			depl, err = w.admin.DB.UpdateDeploymentStatus(ctx, depl.ID, database.DeploymentStatusPending, "Provisioning...")
			if err != nil {
				return err
			}

			// Initialize the deployment (by provisioning a runtime and creating an instance on it)
			err := w.admin.StartDeploymentInner(ctx, depl)
			if err != nil {
				return err
			}
		}

		newStatus = database.DeploymentStatusRunning

	case database.DeploymentStatusStopped:
		// Update the deployment status to stopping
		depl, err = w.admin.DB.UpdateDeploymentStatus(ctx, depl.ID, database.DeploymentStatusStopping, "Stopping...")
		if err != nil {
			return err
		}

		// Stop the deployment by tearing down its runtime instance and resources.
		err = w.admin.StopDeploymentInner(ctx, depl)
		if err != nil {
			return err
		}

		newStatus = database.DeploymentStatusStopped

	case database.DeploymentStatusDeleted:
		// Update the deployment status to deleting
		depl, err = w.admin.DB.UpdateDeploymentStatus(ctx, depl.ID, database.DeploymentStatusDeleting, "Deleting...")
		if err != nil {
			return err
		}

		// Delete the deployment and all its resources.
		err := w.admin.StopDeploymentInner(ctx, depl)
		if err != nil {
			return err
		}

		// Delete the deployment
		err = w.admin.DB.DeleteDeployment(ctx, depl.ID)
		if err != nil {
			return err
		}

		// Nothing more to do, the job is complete.
		return nil

	default:
		// No action needed for other statuses
		return nil
	}

	// Update the deployment status
	depl, err = w.admin.DB.UpdateDeploymentStatus(ctx, depl.ID, newStatus, "")
	if err != nil {
		return err
	}

	// If current depl.DesiredStatusUpdatedOn != desiredStatusUpdatedOn when job started, then the deployment changed while we were working and we should reschedule another job.
	if !depl.DesiredStatusUpdatedOn.Equal(desiredStatusUpdatedOn) {
		// Deployment changed while we were working, reschedule another job to reconcile again.
		c := river.ClientFromContext[pgx.Tx](ctx)
		res, err := c.Insert(ctx, ReconcileDeploymentArgs{
			DeploymentID: job.Args.DeploymentID,
		}, nil)
		if err != nil {
			return err
		}
		w.admin.Logger.Info("reconcile deployment: changes to deployment detected since job started, rescheduling job", observability.ZapCtx(ctx), zap.Int64("new_job_id", res.Job.ID))
	}

	return nil
}
