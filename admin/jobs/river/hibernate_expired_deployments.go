package river

import (
	"context"

	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
)

type HibernateExpiredDeploymentsArgs struct{}

func (HibernateExpiredDeploymentsArgs) Kind() string { return "hibernate_expired_deployments" }

type HibernateExpiredDeploymentsWorker struct {
	river.WorkerDefaults[HibernateExpiredDeploymentsArgs]
	admin  *admin.Service
	logger *zap.Logger
}

func (w *HibernateExpiredDeploymentsWorker) Work(ctx context.Context, job *river.Job[HibernateExpiredDeploymentsArgs]) error {
	depls, err := w.admin.DB.FindExpiredDeployments(ctx)
	if err != nil {
		return err
	}
	if len(depls) == 0 {
		return nil
	}

	w.logger.Info("hibernate: starting", zap.Int("deployments", len(depls)))

	for _, depl := range depls {
		w.logger.Info("hibernate: deleting deployment", zap.String("project_id", depl.ProjectID), zap.String("deployment_id", depl.ID))
		err := w.hibernateExpiredDeployment(ctx, depl)
		if err != nil {
			w.logger.Error("hibernate: failed to delete deployment", zap.String("project_id", depl.ProjectID), zap.String("deployment_id", depl.ID), zap.Error(err), observability.ZapCtx(ctx))
			continue
		}
		w.logger.Info("hibernate: deleted deployment", zap.String("project_id", depl.ProjectID), zap.String("deployment_id", depl.ID))
	}

	w.logger.Info("hibernate: completed", zap.Int("deployments", len(depls)))

	return nil
}

func (w *HibernateExpiredDeploymentsWorker) hibernateExpiredDeployment(ctx context.Context, depl *database.Deployment) error {
	proj, err := w.admin.DB.FindProject(ctx, depl.ProjectID)
	if err != nil {
		return err
	}

	if depl.Environment == "prod" {
		// Tear down prod deployments on hibernation
		// TODO: update this to stop deployment instead of tearing it down when the frontend supports it
		if proj.ProdDeploymentID != nil && *proj.ProdDeploymentID == depl.ID {
			_, err = w.admin.DB.UpdateProject(ctx, proj.ID, &database.UpdateProjectOptions{
				Name:                 proj.Name,
				Description:          proj.Description,
				Public:               proj.Public,
				DirectoryName:        proj.DirectoryName,
				Provisioner:          proj.Provisioner,
				ArchiveAssetID:       proj.ArchiveAssetID,
				GitRemote:            proj.GitRemote,
				GithubInstallationID: proj.GithubInstallationID,
				GithubRepoID:         proj.GithubRepoID,
				ManagedGitRepoID:     proj.ManagedGitRepoID,
				ProdVersion:          proj.ProdVersion,
				ProdBranch:           proj.ProdBranch,
				Subpath:              proj.Subpath,
				ProdSlots:            proj.ProdSlots,
				ProdTTLSeconds:       proj.ProdTTLSeconds,
				ProdDeploymentID:     nil,
				DevSlots:             proj.DevSlots,
				DevTTLSeconds:        proj.DevTTLSeconds,
				Annotations:          proj.Annotations,
			})
			if err != nil {
				return err
			}
		}

		err = w.admin.TeardownDeployment(ctx, depl)
		if err != nil {
			return err
		}
	} else if depl.Environment == "dev" {
		// For dev deployments we stop the deployment
		err = w.admin.StopDeployment(ctx, depl)
		if err != nil {
			return err
		}
	}

	return nil
}
