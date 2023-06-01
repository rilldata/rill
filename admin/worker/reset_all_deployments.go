package worker

import (
	"context"

	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.uber.org/zap"
)

func (w *Worker) resetAllDeployments(ctx context.Context) error {
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
				return err
			}
		}
	}

	// Wait for all the background reconciles to finish.
	// We can remove this when the runtime supports async reconciles.
	w.admin.UnsafeWaitForReconciles()

	return nil
}

func (w *Worker) resetAllDeploymentsForProject(ctx context.Context, proj *database.Project) error {
	depls, err := w.admin.DB.FindDeploymentsForProject(ctx, proj.ID)
	if err != nil {
		return err
	}

	for _, depl := range depls {
		w.logger.Info("reset all deployments: redeploying deployment", zap.String("deployment", depl.ID), observability.ZapCtx(ctx))
		err := w.admin.TriggerRedeploy(ctx, proj, depl)
		if err != nil {
			return err
		}
		w.logger.Info("reset all deployments: redeployed deployment", zap.String("deployment", depl.ID), observability.ZapCtx(ctx))
	}

	return nil
}
