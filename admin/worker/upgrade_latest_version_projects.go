package worker

import (
	"context"

	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.uber.org/zap"
)

func (w *Worker) upgradeLatestVersionProjects(ctx context.Context) error {
	latestVersion := "latest"

	// Try to resolve 'latest' version
	if w.admin.VersionNumber != "" {
		latestVersion = w.admin.VersionNumber
	} else if w.admin.VersionCommit != "" {
		latestVersion = w.admin.VersionCommit
	}

	// Verify version is valid
	err := w.admin.ValidateRuntimeVersion(latestVersion)
	if err != nil {
		return err
	}

	// Iterate over batches of projects with 'latest' version
	limit := 100
	afterName := ""
	stop := false
	for !stop {
		// Get batch and update iterator variables
		projs, err := w.admin.DB.FindProjectsByVersion(ctx, "latest", afterName, limit)
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
			err := w.upgradeAllDeploymentsForProject(ctx, proj, latestVersion)
			if err != nil {
				// We log the error, but continues to the next deployment
				w.logger.Error("upgrade latest version projects: failed to upgrade project deployments", zap.String("project_id", proj.ID), zap.String("version", latestVersion), observability.ZapCtx(ctx))
			}
		}
	}

	return nil
}

func (w *Worker) upgradeAllDeploymentsForProject(ctx context.Context, proj *database.Project, latestVersion string) error {
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
		if depl.RuntimeVersion != latestVersion {
			w.logger.Info("upgrade latest version projects: upgrading deployment", zap.String("deployment_id", depl.ID), zap.String("provision_id", depl.ProvisionID), zap.String("instance_id", depl.RuntimeInstanceID), zap.String("version", latestVersion), observability.ZapCtx(ctx))

			// Update deployment to latest version
			err = w.admin.UpdateDeployment(ctx, depl, &admin.UpdateDeploymentOptions{
				Version:         latestVersion,
				Branch:          depl.Branch,
				Variables:       proj.ProdVariables,
				Annotations:     w.admin.NewDeploymentAnnotations(org, proj),
				EvictCachedRepo: false,
			})
			if err != nil {
				w.logger.Error("upgrade latest version projects: failed to upgrade deployment", zap.String("deployment_id", depl.ID), zap.String("version", latestVersion), observability.ZapCtx(ctx))
				return err
			}

			w.logger.Info("upgrade latest version projects: upgraded deployment", zap.String("deployment_id", depl.ID), zap.String("version", latestVersion), observability.ZapCtx(ctx))
		}
	}

	return nil
}
