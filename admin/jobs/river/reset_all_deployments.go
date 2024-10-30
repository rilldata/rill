package river

import (
	"context"

	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
)

type ResetAllDeploymentsArgs struct{}

func (ResetAllDeploymentsArgs) Kind() string { return "reset_all_deployments" }

type ResetAllDeploymentsWorker struct {
	river.WorkerDefaults[ResetAllDeploymentsArgs]
	admin *admin.Service
}

func (w *ResetAllDeploymentsWorker) Work(ctx context.Context, job *river.Job[ResetAllDeploymentsArgs]) error {
	return work(ctx, w.admin.Logger, job.Kind, w.resetAllDeployments)
}

func (w *ResetAllDeploymentsWorker) resetAllDeployments(ctx context.Context) error {
	// Iterate over batches of projects to redeploy
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
			err := w.resetAllDeploymentsForProject(ctx, proj)
			if err != nil {
				// We log the error, but continues to the next project
				w.admin.Logger.Error("reset all deployments: failed to reset project deployments", zap.String("project_id", proj.ID), observability.ZapCtx(ctx), zap.Error(err))
			}
		}
	}

	return nil
}

func (w *ResetAllDeploymentsWorker) resetAllDeploymentsForProject(ctx context.Context, proj *database.Project) error {
	depls, err := w.admin.DB.FindDeploymentsForProject(ctx, proj.ID)
	if err != nil {
		return err
	}

	for _, depl := range depls {
		// Make sure the deployment provisioner is in the current provisioner set
		_, ok := w.admin.ProvisionerSet[depl.Provisioner]
		if !ok {
			w.admin.Logger.Error("reset all deployments: provisioner is not in the provisioner set", zap.String("provisioner", depl.Provisioner), zap.String("deployment_id", depl.ID), observability.ZapCtx(ctx))
			continue
		}

		w.admin.Logger.Info("reset all deployments: redeploying deployment", zap.String("deployment_id", depl.ID), observability.ZapCtx(ctx))
		_, err = w.admin.RedeployProject(ctx, proj, depl)
		if err != nil {
			return err
		}
		w.admin.Logger.Info("reset all deployments: redeployed deployment", zap.String("deployment_id", depl.ID), observability.ZapCtx(ctx))
	}

	return nil
}
