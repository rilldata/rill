package worker

import (
	"context"

	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.uber.org/zap"
)

func (w *Worker) hibernateExpiredDeployments(ctx context.Context) error {
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

func (w *Worker) hibernateExpiredDeployment(ctx context.Context, depl *database.Deployment) error {
	proj, err := w.admin.DB.FindProject(ctx, depl.ProjectID)
	if err != nil {
		return err
	}

	if proj.ProdDeploymentID != nil && *proj.ProdDeploymentID == depl.ID {
		_, err = w.admin.DB.UpdateProject(ctx, proj.ID, &database.UpdateProjectOptions{
			Name:                 proj.Name,
			Description:          proj.Description,
			Public:               proj.Public,
			Provisioner:          proj.Provisioner,
			ArchiveAssetID:       proj.ArchiveAssetID,
			GithubURL:            proj.GithubURL,
			GithubInstallationID: proj.GithubInstallationID,
			ProdVersion:          proj.ProdVersion,
			ProdBranch:           proj.ProdBranch,
			Subpath:              proj.Subpath,
			ProdSlots:            proj.ProdSlots,
			ProdTTLSeconds:       proj.ProdTTLSeconds,
			ProdDeploymentID:     nil,
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

	return nil
}
