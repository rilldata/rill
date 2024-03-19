package worker

import (
	"context"

	"github.com/rilldata/rill/admin/metrics"
	"go.uber.org/zap"
)

func (w *Worker) runAutoscaler(ctx context.Context) error {
	recs, ok, err := w.allRecommendations(ctx)
	if err != nil {
		return err
	}
	if !ok {
		w.logger.Debug("skipping autoscaler: no metrics project configured")
		return nil
	}

	for _, rec := range recs {
		// TODO: Add autoscaling logic based on the recommendation here.
		// Consider checking rec.UpdatedOn to avoid making autoscaling decision on stale data!
		w.logger.Info("autoscaler recommendation", zap.String("project_id", rec.ProjectID), zap.Int("recommended_slots", rec.RecommendedSlots))
		break
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
		offset += limit
	}

	return recs, true, nil
}
