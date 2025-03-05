package river

import (
	"context"
	"sync"
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

func (w *ValidateDeploymentsWorker) Work(ctx context.Context, job *river.Job[ValidateDeploymentsArgs]) error {
	return work(ctx, w.admin.Logger, job.Kind, w.validateDeployments)
}

const validateDeploymentsForProjectTimeout = 5 * time.Minute

func (w *ValidateDeploymentsWorker) validateDeployments(ctx context.Context) error {
	// Resolve batch size from config
	limit := 100
	if w.admin.ProvisionerMaxConcurrency <= 100 {
		limit = w.admin.ProvisionerMaxConcurrency
	} else {
		w.admin.Logger.Warn("validate deployments: provisioner max concurrency set too high, using maximum concurrency of 100", zap.Int("provisioner_max_concurrency", w.admin.ProvisionerMaxConcurrency), observability.ZapCtx(ctx))
	}

	// Iterate over batches of projects
	var wg sync.WaitGroup
	afterName := ""
	stop := false
	for !stop {
		// Get batch and update iterator variables
		projs, err := w.admin.DB.FindProjects(ctx, afterName, limit)
		if err != nil {
			return err
		}
		if len(projs) < limit {
			stop = true
		}
		if len(projs) != 0 {
			afterName = projs[len(projs)-1].Name
		}

		// Process batch concurrently
		for _, proj := range projs {
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := w.validateDeploymentsForProject(ctx, proj)
				if err != nil {
					// We log the error, but continue to the next project
					w.admin.Logger.Error("validate deployments: failed to validate project deployments", zap.String("project_id", proj.ID), zap.Error(err), observability.ZapCtx(ctx))
				}
			}()
		}

		wg.Wait()
	}

	return nil
}

func (w *ValidateDeploymentsWorker) validateDeploymentsForProject(ctx context.Context, proj *database.Project) error {
	// Apply timeout
	ctx, cancel := context.WithTimeout(ctx, validateDeploymentsForProjectTimeout)
	defer cancel()

	// Get all project deployments
	depls, err := w.admin.DB.FindDeploymentsForProject(ctx, proj.ID)
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
	if proj.ProdDeploymentID != nil {
		prodDeplID = *proj.ProdDeploymentID
	}

	for _, depl := range depls {
		// If it appears to be an orphaned deployment, we tear it down.
		// This might for example happen if a redeploy failed after switching to the new deployment.
		// We consider a deployment orphaned if it is not the prod deployment and has not been updated in 3 hours.
		// The 3 hour delay is to ensure we don't tear down a deployment that is in the process of being created and is to become the new prod deployment.
		if depl.ID != prodDeplID && depl.UpdatedOn.Add(3*time.Hour).Before(time.Now()) {
			w.admin.Logger.Info("validate deployments: removing deployment", zap.String("organization_id", org.ID), zap.String("project_id", proj.ID), zap.String("deployment_id", depl.ID), zap.String("instance_id", depl.RuntimeInstanceID), observability.ZapCtx(ctx))
			err = w.admin.TeardownDeployment(ctx, depl)
			if err != nil {
				w.admin.Logger.Error("validate deployments: failed to remove deployment", zap.String("organization_id", org.ID), zap.String("project_id", proj.ID), zap.String("deployment_id", depl.ID), zap.String("instance_id", depl.RuntimeInstanceID), observability.ZapCtx(ctx), zap.Error(err))
				continue
			}
			w.admin.Logger.Info("validate deployments: removed deployment", zap.String("organization_id", org.ID), zap.String("project_id", proj.ID), zap.String("deployment_id", depl.ID), zap.String("instance_id", depl.RuntimeInstanceID), observability.ZapCtx(ctx))
			continue
		}

		// Retrieve the deployment's provisioned resources
		prs, err := w.admin.DB.FindProvisionerResourcesForDeployment(ctx, depl.ID)
		if err != nil {
			return err
		}
		if len(prs) == 0 {
			continue
		}

		// Build annotations for the deployment
		annotations := w.admin.NewDeploymentAnnotations(org, proj)

		// Validate each provisioned resource
		for _, pr := range prs {
			w.admin.Logger.Info("validate deployments: checking resource", zap.String("organization_id", org.ID), zap.String("project_id", proj.ID), zap.String("deployment_id", depl.ID), zap.String("instance_id", depl.RuntimeInstanceID), zap.String("resource_id", pr.ID), zap.String("provisioner", pr.Provisioner), observability.ZapCtx(ctx))
			err := w.admin.CheckProvisionerResource(ctx, pr, annotations)
			if err != nil {
				w.admin.Logger.Error("validate deployments: failed to check resource", zap.String("organization_id", org.ID), zap.String("project_id", proj.ID), zap.String("deployment_id", depl.ID), zap.String("instance_id", depl.RuntimeInstanceID), zap.String("resource_id", pr.ID), zap.String("provisioner", pr.Provisioner), zap.Error(err), observability.ZapCtx(ctx))
				continue
			}
			w.admin.Logger.Info("validate deployments: checked resource", zap.String("organization_id", org.ID), zap.String("project_id", proj.ID), zap.String("deployment_id", depl.ID), zap.String("instance_id", depl.RuntimeInstanceID), zap.String("resource_id", pr.ID), zap.String("provisioner", pr.Provisioner), observability.ZapCtx(ctx))
		}
	}

	return nil
}
