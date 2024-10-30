package river

import (
	"context"
	"fmt"
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

const validateAllDeploymentsForProjectTimeout = 5 * time.Minute

func (w *ValidateDeploymentsWorker) validateDeployments(ctx context.Context) error {
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
				w.admin.Logger.Error("validate deployments: failed to reconcile project deployments", zap.String("project_id", proj.ID), zap.String("version", latestVersion), observability.ZapCtx(ctx), zap.Error(err))
			}
		}
	}

	return nil
}

func (w *ValidateDeploymentsWorker) reconcileAllDeploymentsForProject(ctx context.Context, proj *database.Project, latestVersion string) error {
	// Apply timeout
	ctx, cancel := context.WithTimeout(ctx, validateAllDeploymentsForProjectTimeout)
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

	var prodDeplID string
	if proj.ProdDeploymentID != nil {
		prodDeplID = *proj.ProdDeploymentID
	}

	for _, depl := range depls {
		if depl.ID == prodDeplID {
			// Get deployment provisioner
			p, ok := w.admin.ProvisionerSet[depl.Provisioner]
			if !ok {
				return fmt.Errorf("validate deployments: %q is not in the provisioner set", depl.Provisioner)
			}

			v, err := p.ValidateConfig(ctx, depl.ProvisionID)
			if err != nil {
				w.admin.Logger.Warn("validate deployments: error validating provisioner config", zap.String("organization_id", org.ID), zap.String("project_id", proj.ID), zap.String("deployment_id", depl.ID), zap.String("provisioner", depl.Provisioner), zap.String("provision_id", depl.ProvisionID), zap.Error(err), observability.ZapCtx(ctx))
				return err
			}

			// Trigger re-provision if config is no longer valid
			if !v {
				w.admin.Logger.Info("validate deployments: config no longer valid, triggering re-provision", zap.String("organization_id", org.ID), zap.String("project_id", proj.ID), zap.String("deployment_id", depl.ID), observability.ZapCtx(ctx))
				err = w.admin.UpdateDeployment(ctx, depl, &admin.UpdateDeploymentOptions{
					Version:         depl.RuntimeVersion,
					Branch:          depl.Branch,
					Variables:       nil,
					Annotations:     w.admin.NewDeploymentAnnotations(org, proj),
					EvictCachedRepo: false,
				})
				if err != nil {
					return err
				}
				w.admin.Logger.Info("validate deployments: re-provisioned", zap.String("organization_id", org.ID), zap.String("project_id", proj.ID), observability.ZapCtx(ctx))
				continue
			}

			// If project is running 'latest' version then update if needed, skip if 'static' provisioner type
			if p.Type() != "static" && proj.ProdVersion == "latest" && depl.RuntimeVersion != latestVersion {
				w.admin.Logger.Info("validate deployments: upgrading deployment", zap.String("organization_id", org.ID), zap.String("project_id", proj.ID), zap.String("deployment_id", depl.ID), zap.String("provisioner", depl.Provisioner), zap.String("provision_id", depl.ProvisionID), zap.String("instance_id", depl.RuntimeInstanceID), zap.String("version", latestVersion), observability.ZapCtx(ctx))

				// Update deployment to latest version
				err = w.admin.UpdateDeployment(ctx, depl, &admin.UpdateDeploymentOptions{
					Version:         latestVersion,
					Branch:          depl.Branch,
					Variables:       nil,
					Annotations:     w.admin.NewDeploymentAnnotations(org, proj),
					EvictCachedRepo: false,
				})
				if err != nil {
					w.admin.Logger.Error("validate deployments: failed to upgrade deployment", zap.String("organization_id", org.ID), zap.String("project_id", proj.ID), zap.String("deployment_id", depl.ID), zap.String("provisioner", depl.Provisioner), zap.String("provision_id", depl.ProvisionID), zap.String("instance_id", depl.RuntimeInstanceID), zap.String("version", latestVersion), observability.ZapCtx(ctx), zap.Error(err))
					return err
				}
				w.admin.Logger.Info("validate deployments: upgraded deployment", zap.String("organization_id", org.ID), zap.String("project_id", proj.ID), zap.String("deployment_id", depl.ID), zap.String("provisioner", depl.Provisioner), zap.String("provision_id", depl.ProvisionID), zap.String("instance_id", depl.RuntimeInstanceID), zap.String("version", latestVersion), observability.ZapCtx(ctx))
			}
		} else if depl.UpdatedOn.Add(3 * time.Hour).Before(time.Now()) {
			// Teardown old orphan non-prod deployment if more than 3 hours since last update
			w.admin.Logger.Info("validate deployments: teardown deployment", zap.String("organization_id", org.ID), zap.String("project_id", proj.ID), zap.String("deployment_id", depl.ID), zap.String("provisioner", depl.Provisioner), zap.String("provision_id", depl.ProvisionID), zap.String("instance_id", depl.RuntimeInstanceID), observability.ZapCtx(ctx))
			err = w.admin.TeardownDeployment(ctx, depl)
			if err != nil {
				w.admin.Logger.Error("validate deployments: teardown deployment error", zap.String("organization_id", org.ID), zap.String("project_id", proj.ID), zap.String("deployment_id", depl.ID), zap.String("provisioner", depl.Provisioner), zap.String("provision_id", depl.ProvisionID), zap.String("instance_id", depl.RuntimeInstanceID), observability.ZapCtx(ctx), zap.Error(err))
				continue
			}
		}
	}

	return nil
}
