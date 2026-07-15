package river

import (
	"context"
	"time"

	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
)

type ValidateDeploymentsArgs struct{}

func (ValidateDeploymentsArgs) Kind() string { return "validate_deployments" }

type ValidateDeploymentsWorker struct {
	river.WorkerDefaults[ValidateDeploymentsArgs]
	admin *admin.Service
}

const validateDeploymentsForProjectTimeout = 5 * time.Minute

func (w *ValidateDeploymentsWorker) Work(ctx context.Context, job *river.Job[ValidateDeploymentsArgs]) error {
	// Iterate over batches of projects and validate each project's deployments.
	// The per-project work is lightweight (it only schedules reconcile jobs), so we process projects sequentially.
	limit := 100
	afterID := ""
	for {
		projs, err := w.admin.DB.FindProjects(ctx, afterID, limit)
		if err != nil {
			return err
		}

		for _, proj := range projs {
			w.admin.Logger.Info("validate deployments: validating project deployments", zap.String("project_id", proj.ID), observability.ZapCtx(ctx))
			if err := w.validateDeploymentsForProject(ctx, proj); err != nil {
				// We log the error, but continue to the next project
				w.admin.Logger.Error("validate deployments: failed to validate project deployments", zap.String("project_id", proj.ID), zap.Error(err), observability.ZapCtx(ctx))
			}
		}

		if len(projs) < limit {
			break
		}
		afterID = projs[len(projs)-1].ID
	}

	return nil
}

func (w *ValidateDeploymentsWorker) validateDeploymentsForProject(ctx context.Context, proj *database.Project) error {
	// Apply timeout
	ctx, cancel := context.WithTimeout(ctx, validateDeploymentsForProjectTimeout)
	defer cancel()

	// Get all project deployments for prod environment
	depls, err := w.admin.DB.FindDeploymentsForProject(ctx, proj.ID, "prod", "")
	if err != nil {
		return err
	}
	if len(depls) == 0 {
		return nil
	}

	// Get project organization, we need this to create the deployment annotations
	org, err := w.admin.DB.FindOrganization(ctx, proj.OrganizationID)
	if err != nil {
		return err
	}

	// Determine the current production deployment, if any
	var prodDeplID string
	if proj.PrimaryDeploymentID != nil {
		prodDeplID = *proj.PrimaryDeploymentID
	}

	for _, depl := range depls {
		// If it appears to be an orphaned prod deployment, we tear it down.
		// This might for example happen if a redeploy failed after switching to the new deployment.
		// We consider a deployment orphaned if it is not the prod deployment, is not stopped and has not been updated in 3 hours.
		// The 3 hour delay is to ensure we don't tear down a deployment that is in the process of being created and is to become the new prod deployment.
		if depl.Environment == "prod" && depl.ID != prodDeplID && depl.Status != database.DeploymentStatusStopped && depl.UpdatedOn.Add(3*time.Hour).Before(time.Now()) {
			w.admin.Logger.Info("validate deployments: removing deployment", zap.String("organization_id", org.ID), zap.String("project_id", proj.ID), zap.String("deployment_id", depl.ID), zap.String("instance_id", depl.RuntimeInstanceID), observability.ZapCtx(ctx))
			err = w.admin.TeardownDeployment(ctx, depl)
			if err != nil {
				w.admin.Logger.Error("validate deployments: failed to remove deployment", zap.String("organization_id", org.ID), zap.String("project_id", proj.ID), zap.String("deployment_id", depl.ID), zap.String("instance_id", depl.RuntimeInstanceID), observability.ZapCtx(ctx), zap.Error(err))
				continue
			}
			w.admin.Logger.Info("validate deployments: removed deployment", zap.String("organization_id", org.ID), zap.String("project_id", proj.ID), zap.String("deployment_id", depl.ID), zap.String("instance_id", depl.RuntimeInstanceID), observability.ZapCtx(ctx))
			continue
		}

		// Schedule a reconcile job. We deliberately do not check the deployment's provisioner resources directly
		// here: all modifications to provisioner resources must happen inside the reconcile loop to avoid racing
		// with other deployment operations (see Service.CheckDeploymentInner / Service.UpdateDeploymentInner).
		// Reconcile only ever drives a deployment toward its own DesiredStatus, is a no-op for already-consistent
		// deployments (e.g. stopped/stopped, deleted/deleted), and is de-duplicated per deployment, so scheduling
		// for every deployment is safe and self-healing.
		_, err := w.admin.Jobs.ReconcileDeployment(ctx, depl.ID)
		if err != nil {
			w.admin.Logger.Error("validate deployments: failed to schedule reconcile", zap.String("organization_id", org.ID), zap.String("project_id", proj.ID), zap.String("deployment_id", depl.ID), zap.String("instance_id", depl.RuntimeInstanceID), zap.Error(err), observability.ZapCtx(ctx))
			continue
		}
		w.admin.Logger.Info("validate deployments: scheduled reconcile", zap.String("organization_id", org.ID), zap.String("project_id", proj.ID), zap.String("deployment_id", depl.ID), zap.String("instance_id", depl.RuntimeInstanceID), observability.ZapCtx(ctx))
	}

	return nil
}
