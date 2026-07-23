package river

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/riverqueue/river"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

type ReconcileDeploymentArgs struct {
	DeploymentID string
}

func (ReconcileDeploymentArgs) Kind() string { return "reconcile_deployment" }

type ReconcileDeploymentWorker struct {
	river.WorkerDefaults[ReconcileDeploymentArgs]
	admin                  *admin.Service
	findDeployment         func(context.Context, string) (*database.Deployment, error)
	updateDeploymentStatus func(context.Context, string, database.DeploymentStatus, string) (*database.Deployment, error)
	startDeployment        func(context.Context, *database.Deployment) error
	updateDeployment       func(context.Context, *database.Deployment) error
	stopDeployment         func(context.Context, *database.Deployment) error
	deleteDeployment       func(context.Context, *database.Deployment) error
	enqueueReconcile       func(context.Context, string) (int64, error)
}

// NewReconcileDeploymentWorker creates a new ReconcileDeploymentWorker. Only to be used in tests to trigger the worker directly.
func NewReconcileDeploymentWorker(admin *admin.Service) *ReconcileDeploymentWorker {
	return &ReconcileDeploymentWorker{
		admin: admin,
	}
}

// ReconcileDeploymentWorker is a state machine, it reconciles the state of a deployment based on its desired and current status.
// This job is responsible for transitioning deployments through their lifecycle states,
// such as starting, updating, stopping, and deleting deployments.
// We handle all deployment state transitions in this job to ensure consistency and to avoid concurrent conflicting operations on the same deployment.
func (w *ReconcileDeploymentWorker) Work(ctx context.Context, job *river.Job[ReconcileDeploymentArgs]) error {
	observability.AddRequestAttributes(ctx, attribute.String("args.deployment_id", job.Args.DeploymentID))
	depl, err := w.loadDeployment(ctx, job.Args.DeploymentID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			// If the deployment doesn't exist, we can just finish the job and do nothing more.
			w.logger().Info("reconcile deployment: deployment not found, job succeeded", observability.ZapCtx(ctx))
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
			// Reconcile the running deployment towards its desired configuration.
			// UpdateDeploymentInner is change-aware: it only reprovisions when the provisioning args have
			// changed, and otherwise performs a lightweight drift-aware resource check. We therefore keep the
			// deployment in the Running status here rather than flipping it to Updating, which would otherwise flap
			// on every periodic reconciliation.
			err := w.reconcileRunningDeployment(ctx, depl)
			if err != nil {
				return err
			}
		} else {
			// Update the deployment status to pending
			depl, err = w.persistDeploymentStatus(ctx, depl.ID, database.DeploymentStatusPending, "Provisioning...")
			if err != nil {
				return err
			}

			// Initialize the deployment (by provisioning a runtime and creating an instance on it)
			err := w.initializeDeployment(ctx, depl)
			if err != nil {
				return err
			}
		}

		newStatus = database.DeploymentStatusRunning

	case database.DeploymentStatusStopped:
		if depl.Status == database.DeploymentStatusStopped {
			// No action needed, already stopped
			return nil
		}
		// Update the deployment status to stopping
		depl, err = w.persistDeploymentStatus(ctx, depl.ID, database.DeploymentStatusStopping, "Stopping...")
		if err != nil {
			return err
		}

		// Stop the deployment by tearing down its runtime instance and resources.
		err = w.hibernateDeployment(ctx, depl)
		if err != nil {
			return err
		}

		newStatus = database.DeploymentStatusStopped

	case database.DeploymentStatusDeleted:
		if depl.Status == database.DeploymentStatusDeleted {
			// No action needed, already deleted
			return nil
		}
		// Update the deployment status to deleting
		depl, err = w.persistDeploymentStatus(ctx, depl.ID, database.DeploymentStatusDeleting, "Deleting...")
		if err != nil {
			return err
		}

		// Delete the deployment and all its resources.
		err := w.removeDeployment(ctx, depl)
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
	depl, err = w.persistDeploymentStatus(ctx, depl.ID, newStatus, "")
	if err != nil {
		return err
	}

	// If current depl.DesiredStatusUpdatedOn != desiredStatusUpdatedOn when job started, then the deployment changed while we were working and we should reschedule another job.
	if !depl.DesiredStatusUpdatedOn.Equal(desiredStatusUpdatedOn) {
		// Deployment changed while we were working, reschedule another job to reconcile again.
		newJobID, err := w.scheduleReconcile(ctx, job.Args.DeploymentID)
		if err != nil {
			return err
		}
		w.logger().Info("reconcile deployment: changes to deployment detected since job started, rescheduling job", observability.ZapCtx(ctx), zap.Int64("new_job_id", newJobID))
	}

	return nil
}

func (w *ReconcileDeploymentWorker) loadDeployment(ctx context.Context, deploymentID string) (*database.Deployment, error) {
	if w.findDeployment != nil {
		return w.findDeployment(ctx, deploymentID)
	}
	return w.admin.DB.FindDeployment(ctx, deploymentID)
}

func (w *ReconcileDeploymentWorker) persistDeploymentStatus(ctx context.Context, deploymentID string, status database.DeploymentStatus, message string) (*database.Deployment, error) {
	if w.updateDeploymentStatus != nil {
		return w.updateDeploymentStatus(ctx, deploymentID, status, message)
	}
	return w.admin.DB.UpdateDeploymentStatus(ctx, deploymentID, status, message)
}

func (w *ReconcileDeploymentWorker) initializeDeployment(ctx context.Context, deployment *database.Deployment) error {
	if w.startDeployment != nil {
		return w.startDeployment(ctx, deployment)
	}
	return w.admin.StartDeploymentInner(ctx, deployment)
}

func (w *ReconcileDeploymentWorker) reconcileRunningDeployment(ctx context.Context, deployment *database.Deployment) error {
	if w.updateDeployment != nil {
		return w.updateDeployment(ctx, deployment)
	}
	return w.admin.UpdateDeploymentInner(ctx, deployment)
}

func (w *ReconcileDeploymentWorker) hibernateDeployment(ctx context.Context, deployment *database.Deployment) error {
	if w.stopDeployment != nil {
		return w.stopDeployment(ctx, deployment)
	}
	return w.admin.StopDeploymentInner(ctx, deployment)
}

func (w *ReconcileDeploymentWorker) removeDeployment(ctx context.Context, deployment *database.Deployment) error {
	if w.deleteDeployment != nil {
		return w.deleteDeployment(ctx, deployment)
	}
	return w.admin.DeleteDeploymentInner(ctx, deployment)
}

func (w *ReconcileDeploymentWorker) scheduleReconcile(ctx context.Context, deploymentID string) (int64, error) {
	if w.enqueueReconcile != nil {
		return w.enqueueReconcile(ctx, deploymentID)
	}
	c := river.ClientFromContext[pgx.Tx](ctx)
	res, err := c.Insert(ctx, ReconcileDeploymentArgs{DeploymentID: deploymentID}, nil)
	if err != nil {
		return 0, err
	}
	return res.Job.ID, nil
}

func (w *ReconcileDeploymentWorker) logger() *zap.Logger {
	if w.admin != nil && w.admin.Logger != nil {
		return w.admin.Logger
	}
	return zap.NewNop()
}
