package worker

import (
	"context"
	"time"

	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/metrics"
	"go.uber.org/zap"
)

const legacyRecommendTime = 24 * time.Hour

const scaleThreshold = 0.10

func (w *Worker) runAutoscaler(ctx context.Context) error {
	recs, ok, err := w.allRecommendations(ctx)
	if err != nil {
		w.logger.Error("failed to autoscale: unable to fetch recommended slots", zap.Error(err))
		return err
	}
	if !ok {
		w.logger.Debug("skipping autoscaler: no metrics project configured")
		return nil
	}

	for _, rec := range recs {
		// if UpdatedOn is too old, the recommendation is stale and may not be trusted.
		if time.Since(rec.UpdatedOn) >= legacyRecommendTime {
			w.logger.Debug("skipping autoscaler: the recommendation is stale", zap.String("project_id", rec.ProjectID), zap.Time("recommendation_updated_on", rec.UpdatedOn))
			continue
		}

		if rec.RecommendedSlots <= 0 {
			w.logger.Debug("skipping autoscaler: the recommend slot is <= 0", zap.String("project_id", rec.ProjectID), zap.Int("recommendation_slots", rec.RecommendedSlots))
			continue
		}

		targetProject, err := w.admin.DB.FindProject(ctx, rec.ProjectID)
		if err != nil {
			w.logger.Debug("failed to find project:", zap.String("project_id", rec.ProjectID), zap.Error(err))
			continue
		}

		if !shouldScale(targetProject.ProdSlots, rec.RecommendedSlots) {
			w.logger.Debug("skipping autoscaler: target slots are within threshold of original slots",
				zap.Int("project_slots", targetProject.ProdSlots),
				zap.Int("recommend_slots", rec.RecommendedSlots),
				zap.Float64("scale_threshold_percentage", scaleThreshold),
				zap.String("project_id", targetProject.ID),
			)
			continue
		}

		updatedProject, err := w.admin.UpdateProject(ctx, targetProject, &database.UpdateProjectOptions{
			Name:                 targetProject.Name,
			Description:          targetProject.Description,
			Public:               targetProject.Public,
			GithubURL:            targetProject.GithubURL,
			GithubInstallationID: targetProject.GithubInstallationID,
			ProdVersion:          targetProject.ProdVersion,
			ProdBranch:           targetProject.ProdBranch,
			ProdVariables:        targetProject.ProdVariables,
			ProdDeploymentID:     targetProject.ProdDeploymentID,
			ProdSlots:            rec.RecommendedSlots,
			ProdTTLSeconds:       targetProject.ProdTTLSeconds,
			Provisioner:          targetProject.Provisioner,
			Annotations:          targetProject.Annotations,
		})

		if err != nil {
			w.logger.Error("failed to autoscale:", zap.String("project_id", rec.ProjectID), zap.Error(err))
			continue
		}

		w.logger.Info("succeeded in autoscaling:", zap.String("project_id", updatedProject.Name), zap.Int("project_slots", updatedProject.ProdSlots))
	}

	return nil
}

func (w *Worker) allRecommendations(ctx context.Context) ([]metrics.AutoscalerSlotsRecommendation, bool, error) {
	client, ok, err := w.admin.OpenMetricsProject(ctx)
	if err != nil {
		return nil, false, err
	}
	if !ok {
		return nil, false, nil
	}

	var recs []metrics.AutoscalerSlotsRecommendation
	limit := 1000
	offset := 0
	for {
		batch, err := client.AutoscalerSlotsRecommendations(ctx, limit, offset)
		if err != nil {
			return nil, false, err
		}
		if len(batch) == 0 {
			break
		}
		recs = append(recs, batch...)

		if len(batch) < limit {
			break
		}

		offset += limit
	}

	return recs, true, nil
}

func shouldScale(originSlots, recommendSlots int) bool {
	lowerBound := float64(originSlots) * (1 - scaleThreshold)
	upperBound := float64(originSlots) * (1 + scaleThreshold)
	return float64(recommendSlots) < lowerBound || float64(recommendSlots) > upperBound
}
