package river

import (
	"context"
	"math"
	"time"

	"github.com/rilldata/rill/admin"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/metrics"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
)

const (
	legacyRecommendTime   = 24 * time.Hour
	scaleThreshold        = 0.10 // 10%
	smallServiceThreshold = 10
	minScalingSlots       = 5.0
	disableAutoscaler     = true

	// Reasons for not scaling
	scaledown      = "scaling down is temporarily disabled due to constraint"
	scaleMatch     = "current scale equals recommendation"
	belowThreshold = "scaling change is below the threshold"
)

type RunAutoscalerArgs struct{}

func (RunAutoscalerArgs) Kind() string { return "run_autoscaler" }

type RunAutoscalerWorker struct {
	river.WorkerDefaults[RunAutoscalerArgs]
	admin  *admin.Service
	logger *zap.Logger
}

func (w *RunAutoscalerWorker) Work(ctx context.Context, job *river.Job[RunAutoscalerArgs]) error {
	if disableAutoscaler {
		w.logger.Info("skipping autoscaler: disabled by configuration")
		return nil
	}

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
		targetProject, err := w.admin.DB.FindProject(ctx, rec.ProjectID)
		if err != nil {
			w.logger.Debug("failed to find project", zap.String("project_name", targetProject.Name), zap.Error(err))
			continue
		}

		projectOrg, err := w.admin.DB.FindOrganization(ctx, targetProject.OrganizationID)
		if err != nil {
			w.logger.Error("failed to autoscale: unable to find org for the project", zap.String("organization_name", projectOrg.Name), zap.String("project_name", targetProject.Name), zap.Error(err))
			continue
		}

		// if UpdatedOn is too old, the recommendation is stale and may not be trusted.
		if time.Since(rec.UpdatedOn) >= legacyRecommendTime {
			w.logger.Debug("skipping autoscaler: the recommendation is stale", zap.String("project_name", targetProject.Name), zap.Time("recommendation_updated_on", rec.UpdatedOn))
			continue
		}

		if rec.RecommendedSlots <= 0 {
			w.logger.Debug("skipping autoscaler: the recommend slot is <= 0", zap.String("project_name", targetProject.Name), zap.Int("recommended_slots", rec.RecommendedSlots))
			continue
		}

		// If it's proposing to scale up, make sure we don't scale beyond the quota
		if rec.RecommendedSlots > targetProject.ProdSlots {
			usage, err := w.admin.DB.CountProjectsQuotaUsage(ctx, projectOrg.ID)
			if err != nil {
				return err
			}

			var overshoot int
			if projectOrg.QuotaSlotsPerDeployment >= 0 {
				overshoot = max(overshoot, rec.RecommendedSlots-projectOrg.QuotaSlotsPerDeployment)
			}
			if projectOrg.QuotaSlotsTotal >= 0 {
				overshoot = max(overshoot, usage.Slots-targetProject.ProdSlots+rec.RecommendedSlots-projectOrg.QuotaSlotsTotal)
			}

			// If the recommendation would exceed a quota, change it to scale to the limit of the quota.
			if overshoot > 0 {
				if rec.RecommendedSlots-overshoot < targetProject.ProdSlots {
					w.logger.Debug("skipping autoscaler: already scaled to or beyond the quota", zap.String("organization_name", projectOrg.Name), zap.String("project_name", targetProject.Name), zap.Int("recommended_slots", rec.RecommendedSlots))
					continue
				}

				rec.RecommendedSlots -= overshoot
			}
		}

		if shouldScale, reason := shouldScale(targetProject.ProdSlots, rec.RecommendedSlots, w.admin.ScaleDownConstraint); !shouldScale {
			logMessage := "skipping autoscaler: " + reason

			logFields := []zap.Field{
				zap.String("organization_name", projectOrg.Name),
				zap.String("project_name", targetProject.Name),
				zap.Int("current_slots", targetProject.ProdSlots),
				zap.Int("recommended_slots", rec.RecommendedSlots),
				zap.Float64("scale_threshold_percentage", scaleThreshold),
			}

			if reason == scaledown {
				w.logger.Info(logMessage, logFields...)
			} else {
				w.logger.Debug(logMessage, logFields...)
			}
			continue
		}

		updatedProject, err := w.admin.UpdateProject(ctx, targetProject, &database.UpdateProjectOptions{
			Name:                 targetProject.Name,
			Description:          targetProject.Description,
			Public:               targetProject.Public,
			DirectoryName:        targetProject.DirectoryName,
			ArchiveAssetID:       targetProject.ArchiveAssetID,
			GitRemote:            targetProject.GitRemote,
			GithubInstallationID: targetProject.GithubInstallationID,
			GithubRepoID:         targetProject.GithubRepoID,
			ManagedGitRepoID:     targetProject.ManagedGitRepoID,
			ProdVersion:          targetProject.ProdVersion,
			ProdBranch:           targetProject.ProdBranch,
			Subpath:              targetProject.Subpath,
			ProdDeploymentID:     targetProject.ProdDeploymentID,
			ProdSlots:            rec.RecommendedSlots,
			ProdTTLSeconds:       targetProject.ProdTTLSeconds,
			DevSlots:             targetProject.DevSlots,
			DevTTLSeconds:        targetProject.DevTTLSeconds,
			Provisioner:          targetProject.Provisioner,
			Annotations:          targetProject.Annotations,
		})
		if err != nil {
			w.logger.Error("failed to autoscale: error updating the project", zap.String("project_name", targetProject.Name), zap.String("organization_name", projectOrg.Name), zap.Error(err))
			continue
		}

		scaleMsg := "succeeded in autoscaling "
		if updatedProject.ProdSlots > targetProject.ProdSlots {
			scaleMsg += "up"
		} else {
			scaleMsg += "down"
		}

		w.logger.Info(scaleMsg,
			zap.String("project_name", updatedProject.Name),
			zap.Int("updated_slots", updatedProject.ProdSlots),
			zap.Int("prev_slots", targetProject.ProdSlots),
			zap.String("organization_name", projectOrg.Name),
		)
	}

	return nil
}

func (w *RunAutoscalerWorker) allRecommendations(ctx context.Context) ([]metrics.AutoscalerSlotsRecommendation, bool, error) {
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

// shouldScale determines whether scaling operations should be initiated
// based on the comparison of the current number of slots (originSlots)
// and the recommended number of slots (recommendSlots).
func shouldScale(originSlots, recommendSlots, scaleDownConstraint int) (bool, string) {
	if recommendSlots == originSlots {
		return false, scaleMatch
	}

	// NOTE(2024-10-15): Disable scale down if breaking the constraints
	if recommendSlots < originSlots {
		if scaleDownConstraint != -1 && originSlots > scaleDownConstraint {
			return false, scaledown
		}
	}

	// Always allow scaling for small services
	if originSlots < smallServiceThreshold {
		return true, ""
	}

	// Calculate the absolute difference in slots
	scalingSlots := math.Abs(float64(recommendSlots - originSlots))

	// Avoid scaling if increase/decrease is less than 10%
	if scalingSlots <= float64(originSlots)*scaleThreshold {
		return false, belowThreshold
	}

	// Avoid scaling if increase/decrease is less than 5 slots
	if scalingSlots < minScalingSlots {
		return false, belowThreshold
	}

	return true, ""
}
