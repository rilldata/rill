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

// stoppedDeploymentRetention is how long after a deployment is stopped that we keep it around before fully deleting it.
// (This ensures that we eventually delete persistent state like PVCs for stopped deployments.)
const stoppedDeploymentRetention = 30 * 24 * time.Hour

type HibernateExpiredDeploymentsArgs struct{}

func (HibernateExpiredDeploymentsArgs) Kind() string { return "hibernate_expired_deployments" }

type HibernateExpiredDeploymentsWorker struct {
	river.WorkerDefaults[HibernateExpiredDeploymentsArgs]
	admin  *admin.Service
	logger *zap.Logger
}

func (w *HibernateExpiredDeploymentsWorker) Work(ctx context.Context, job *river.Job[HibernateExpiredDeploymentsArgs]) error {
	// Stop currently running deployments that are inactive.
	running, err := w.admin.DB.FindDeploymentsToStop(ctx)
	if err != nil {
		return err
	}
	w.logger.Info("hibernate: checking running deployments", zap.Int("deployments", len(running)))
	for _, depl := range running {
		w.logger.Info("hibernate: stopping deployment", zap.String("project_id", depl.ProjectID), zap.String("deployment_id", depl.ID))
		err := w.stopDeployment(ctx, depl)
		if err != nil {
			w.logger.Error("hibernate: failed to stop deployment", zap.String("project_id", depl.ProjectID), zap.String("deployment_id", depl.ID), zap.Error(err), observability.ZapCtx(ctx))
			continue
		}
		w.logger.Info("hibernate: stopped deployment", zap.String("project_id", depl.ProjectID), zap.String("deployment_id", depl.ID))
	}
	w.logger.Info("hibernate: completed running deployments check", zap.Int("deployments", len(running)))

	// Delete deployments that have been stopped for a long time.
	stopped, err := w.admin.DB.FindDeploymentsToDelete(ctx, stoppedDeploymentRetention)
	if err != nil {
		return err
	}
	w.logger.Info("hibernate: checking stopped deployments", zap.Int("deployments", len(stopped)))
	for _, depl := range stopped {
		w.logger.Info("hibernate: deleting deployment", zap.String("project_id", depl.ProjectID), zap.String("deployment_id", depl.ID))
		err := w.deleteStoppedDeployment(ctx, depl)
		if err != nil {
			w.logger.Error("hibernate: failed to delete stopped deployment", zap.String("project_id", depl.ProjectID), zap.String("deployment_id", depl.ID), zap.Error(err), observability.ZapCtx(ctx))
			continue
		}
		w.logger.Info("hibernate: deleted stopped deployment", zap.String("project_id", depl.ProjectID), zap.String("deployment_id", depl.ID))
	}
	w.logger.Info("hibernate: completed stopped deployments check", zap.Int("deployments", len(stopped)))

	return nil
}

func (w *HibernateExpiredDeploymentsWorker) stopDeployment(ctx context.Context, depl *database.Deployment) error {
	err := w.admin.StopDeployment(ctx, depl)
	if err != nil {
		return err
	}

	return nil
}

func (w *HibernateExpiredDeploymentsWorker) deleteStoppedDeployment(ctx context.Context, depl *database.Deployment) error {
	proj, err := w.admin.DB.FindProject(ctx, depl.ProjectID)
	if err != nil {
		return err
	}

	if proj.PrimaryDeploymentID != nil && *proj.PrimaryDeploymentID == depl.ID {
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
			PrimaryBranch:        proj.PrimaryBranch,
			Subpath:              proj.Subpath,
			ProdSlots:            proj.ProdSlots,
			ProdTTLSeconds:       proj.ProdTTLSeconds,
			PrimaryDeploymentID:  nil,
			DevSlots:             proj.DevSlots,
			DevTTLSeconds:        proj.DevTTLSeconds,
			Annotations:          proj.Annotations,
		})
		if err != nil {
			return err
		}
	}

	return w.admin.TeardownDeployment(ctx, depl)
}
