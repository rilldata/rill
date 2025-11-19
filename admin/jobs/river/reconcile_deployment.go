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

// ReconcileDeploymentWorker is a state machine, it reconciles the state of a deployment based on its current status.
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

	var deplOpts *database.UpdateDeploymentOptions

	switch depl.Status {
	case database.DeploymentStatusPending:
		// Initialize the deployment (by provisioning a runtime and creating an instance on it)
		rtCfg, err := w.admin.StartDeploymentInner(ctx, depl)
		if err != nil {
			return err
		}
		deplOpts = &database.UpdateDeploymentOptions{
			Branch:            depl.Branch,
			RuntimeHost:       rtCfg.Host,
			RuntimeInstanceID: rtCfg.InstanceID,
			RuntimeAudience:   rtCfg.Audience,
			Status:            database.DeploymentStatusOK,
			StatusMessage:     "",
		}

	case database.DeploymentStatusStopping:
		// Stop the deployment by tearing down its runtime instance and resources.
		err := w.admin.StopDeploymentInner(ctx, depl)
		if err != nil {
			return err
		}
		deplOpts = &database.UpdateDeploymentOptions{
			Branch:            depl.Branch,
			RuntimeHost:       depl.RuntimeHost,
			RuntimeInstanceID: depl.RuntimeInstanceID,
			RuntimeAudience:   depl.RuntimeAudience,
			Status:            database.DeploymentStatusStopped,
			StatusMessage:     "",
		}

	case database.DeploymentStatusUpdating:
		// Update the deployment by updating its runtime instance and resources.
		err := w.admin.UpdateDeploymentInner(ctx, depl)
		if err != nil {
			return err
		}
		deplOpts = &database.UpdateDeploymentOptions{
			Branch:            depl.Branch,
			RuntimeHost:       depl.RuntimeHost,
			RuntimeInstanceID: depl.RuntimeInstanceID,
			RuntimeAudience:   depl.RuntimeAudience,
			Status:            database.DeploymentStatusOK,
			StatusMessage:     "",
		}

	case database.DeploymentStatusDeleting:
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

	updatedOn := depl.UpdatedOn
	reschedule := false

	// If current depl.UpdatedOn != updatedOn when job started, then the deployment changed while we were working and we should reschedule another job.
	// Otherwise, we can just update the status and finish the job, we do this in a transaction to prevent other updates to the deployment in between.
	txCtx, tx, err := w.admin.DB.NewTx(ctx, false)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	depl, err = w.admin.DB.FindDeployment(txCtx, job.Args.DeploymentID)
	if err != nil {
		return err
	}

	if depl.UpdatedOn.Equal(updatedOn) {
		// Update the deployment
		_, err = w.admin.DB.UpdateDeployment(txCtx, depl.ID, deplOpts)
		if err != nil {
			return err
		}
	} else {
		reschedule = true
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	if reschedule {
		// Deployment changed while we were working, reschedule another job to reconcile again.
		c := river.ClientFromContext[pgx.Tx](ctx)
		res, err := c.Insert(ctx, ReconcileDeploymentArgs{
			DeploymentID: job.Args.DeploymentID,
		}, nil)
		if err != nil {
			return err
		}
		w.admin.Logger.Info("reconcile deployment: changes to deployment detected since job started, rescheduling job", observability.ZapCtx(ctx), zap.Int64("new_job_id", res.Job.ID))
		return nil
	}

	return nil
}
