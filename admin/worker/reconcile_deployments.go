package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.uber.org/zap"
)

const reconcileAllDeploymentsForProjectTimeout = 5 * time.Minute

func (w *Worker) reconcileDeployments(ctx context.Context) error {
	// Resolve 'latest' version
	latestVersion := w.admin.ResolveLatestRuntimeVersion()

	// Verify version is valid
	err := w.admin.ValidateRuntimeVersion(latestVersion)
	if err != nil {
		return err
	}

	// Iterate over batches of projects
	limit := 100
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

		// Process batch
		for _, proj := range projs {
			err := w.reconcileAllDeploymentsForProject(ctx, proj, latestVersion)
			if err != nil {
				// We log the error, but continues to the next project
				w.logger.Error("reconcile deployments: failed to reconcile project deployments", zap.String("project_id", proj.ID), zap.String("version", latestVersion), observability.ZapCtx(ctx), zap.Error(err))
			}
		}
	}

	return nil
}

func (w *Worker) reconcileAllDeploymentsForProject(ctx context.Context, proj *database.Project, latestVersion string) error {
	// Apply timeout
	ctx, cancel := context.WithTimeout(ctx, reconcileAllDeploymentsForProjectTimeout)
	defer cancel()

	// Get all project deployments
	depls, err := w.admin.DB.FindDeploymentsForProject(ctx, proj.ID)
	if err != nil {
		return err
	}

	// Get project organization, we need this to create the deployment annotations
	org, err := w.admin.DB.FindOrganization(ctx, proj.OrganizationID)
	if err != nil {
		return err
	}

	for _, depl := range depls {
		if depl.ID == *proj.ProdDeploymentID {
			// Get deployment provisioner
			p, ok := w.admin.ProvisionerSet[depl.Provisioner]
			if !ok {
				return fmt.Errorf("reconcile deployments: %q is not in the provisioner set", depl.Provisioner)
			}

			v, err := p.ValidateConfig(ctx, depl.ProvisionID)
			if err != nil {
				w.logger.Warn("reconcile deployments: error validating provisioner config", zap.String("organization_id", org.ID), zap.String("project_id", proj.ID), zap.String("deployment_id", depl.ID), zap.String("provisioner", depl.Provisioner), zap.String("provision_id", depl.ProvisionID), zap.Error(err), observability.ZapCtx(ctx))
				return err
			}

			// Trigger a redeploy if config is no longer valid
			if !v {
				w.logger.Info("reconcile deployments: config no longer valid, triggering redeploy", zap.String("organization_id", org.ID), zap.String("project_id", proj.ID), zap.String("deployment_id", depl.ID), observability.ZapCtx(ctx))
				_, err = w.admin.TriggerRedeploy(ctx, proj, depl)
				if err != nil {
					return err
				}
				w.logger.Info("reconcile deployments: redeployed", zap.String("organization_id", org.ID), zap.String("project_id", proj.ID), observability.ZapCtx(ctx))
				continue
			}

			// If project is running 'latest' version then update if needed, skip if 'static' provisioner type
			if p.GetType() != "static" && proj.ProdVersion == "latest" && depl.RuntimeVersion != latestVersion {
				w.logger.Info("reconcile deployments: upgrading deployment", zap.String("organization_id", org.ID), zap.String("project_id", proj.ID), zap.String("deployment_id", depl.ID), zap.String("provisioner", depl.Provisioner), zap.String("provision_id", depl.ProvisionID), zap.String("instance_id", depl.RuntimeInstanceID), zap.String("version", latestVersion), observability.ZapCtx(ctx))

				// Update deployment to latest version
				err = w.admin.UpdateDeployment(ctx, depl, &admin.UpdateDeploymentOptions{
					Version:         latestVersion,
					Branch:          depl.Branch,
					Variables:       proj.ProdVariables,
					Annotations:     w.admin.NewDeploymentAnnotations(org, proj),
					EvictCachedRepo: false,
				})
				if err != nil {
					w.logger.Error("reconcile deployments: failed to upgrade deployment", zap.String("organization_id", org.ID), zap.String("project_id", proj.ID), zap.String("deployment_id", depl.ID), zap.String("provisioner", depl.Provisioner), zap.String("provision_id", depl.ProvisionID), zap.String("instance_id", depl.RuntimeInstanceID), zap.String("version", latestVersion), observability.ZapCtx(ctx), zap.Error(err))
					return err
				}
				w.logger.Info("reconcile deployments: upgraded deployment", zap.String("organization_id", org.ID), zap.String("project_id", proj.ID), zap.String("deployment_id", depl.ID), zap.String("provisioner", depl.Provisioner), zap.String("provision_id", depl.ProvisionID), zap.String("instance_id", depl.RuntimeInstanceID), zap.String("version", latestVersion), observability.ZapCtx(ctx))
			}
		} else if depl.UpdatedOn.Add(3 * time.Hour).After(time.Now()) {
			// Teardown old orphan non-prod deployment if more than 3 hours since last update
			err = w.admin.TeardownDeployment(ctx, depl)
			if err != nil {
				w.logger.Error("reconcile deployments: teardown deployment error", zap.String("organization_id", org.ID), zap.String("project_id", proj.ID), zap.String("deployment_id", depl.ID), zap.String("provisioner", depl.Provisioner), zap.String("provision_id", depl.ProvisionID), zap.String("instance_id", depl.RuntimeInstanceID), observability.ZapCtx(ctx), zap.Error(err))
				continue
			}
		}
	}

	return nil
}
