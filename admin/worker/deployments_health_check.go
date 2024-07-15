package worker

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rilldata/rill/admin/database"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (w *Worker) deploymentsHealthCheck(ctx context.Context) error {
	afterID := ""
	limit := 100
	seenHosts := map[string]bool{}
	for {
		deployments, err := w.admin.DB.FindDeployments(ctx, afterID, limit)
		if err != nil {
			return fmt.Errorf("deploymentsHealthCheck: failed to get deployments: %w", err)
		}
		if len(deployments) == 0 {
			return nil
		}

		group, cctx := errgroup.WithContext(ctx)
		group.SetLimit(8)
		for _, d := range deployments {
			d := d
			if d.Status != database.DeploymentStatusOK {
				if time.Since(d.UpdatedOn) > time.Hour {
					w.logger.Error("deploymentsHealthCheck: failing deployment", zap.String("id", d.ID), zap.String("status", d.Status.String()), zap.Duration("duration", time.Since(d.UpdatedOn)))
				}
				continue
			}
			if seenHosts[d.RuntimeHost] {
				continue
			}
			seenHosts[d.RuntimeHost] = true
			group.Go(func() error {
				return w.deploymentHealthCheck(cctx, d)
			})
		}
		if err := group.Wait(); err != nil {
			return err
		}
		if len(deployments) < limit {
			return nil
		}
		afterID = deployments[len(deployments)-1].ID
		// fetch again
	}
}

func (w *Worker) deploymentHealthCheck(ctx context.Context, d *database.Deployment) error {
	client, err := w.admin.OpenRuntimeClient(d.RuntimeHost, d.RuntimeAudience)
	if err != nil {
		w.logger.Error("deploymentsHealthCheck: failed to open runtime client", zap.String("host", d.RuntimeHost), zap.Error(err))
		return nil
	}
	defer client.Close()

	resp, err := client.Health(ctx, &runtimev1.HealthRequest{})
	if err != nil {
		if status.Code(err) != codes.Unavailable {
			w.logger.Error("deploymentsHealthCheck: health check call failed", zap.String("host", d.RuntimeHost), zap.Error(err))
			return nil
		}
		// an unavailable error could also be because the deployment got deleted
		d, dbErr := w.admin.DB.FindDeployment(ctx, d.ID)
		if dbErr != nil {
			if errors.Is(dbErr, database.ErrNotFound) {
				// Deployment was deleted
				return nil
			}
			w.logger.Error("deploymentsHealthCheck: failed to find deployment", zap.String("deployment", d.ID), zap.Error(dbErr))
			return nil
		}
		if d.Status == database.DeploymentStatusOK {
			w.logger.Error("deploymentsHealthCheck: health check call failed", zap.String("host", d.RuntimeHost), zap.Error(err))
		}
		// Deployment status changed (probably being deleted)
		return nil
	}

	if runtimeUnhealthy(resp) {
		f := []zap.Field{zap.String("host", d.RuntimeHost)}
		if resp.LimiterError != "" {
			f = append(f, zap.String("limiter_error", resp.LimiterError))
		}
		if resp.ConnCacheError != "" {
			f = append(f, zap.String("conn_cache_error", resp.ConnCacheError))
		}
		if resp.MetastoreError != "" {
			f = append(f, zap.String("metastore_error", resp.MetastoreError))
		}
		if resp.NetworkError != "" {
			f = append(f, zap.String("network_error", resp.NetworkError))
		}
		w.logger.Error("deploymentsHealthCheck: runtime is unhealthy", f...)
		return nil
	}
	for id, i := range resp.InstancesHealth {
		if !instanceUnhealthy(i) {
			continue
		}
		f := []zap.Field{zap.String("host", d.RuntimeHost), zap.String("instance_id", id)}
		if i.OlapError != "" {
			f = append(f, zap.String("olap_error", i.OlapError))
		}
		if i.ControllerError != "" {
			f = append(f, zap.String("controller_error", i.ControllerError))
		}
		if i.RepoError != "" {
			f = append(f, zap.String("repo_error", i.RepoError))
		}
		w.logger.Error("deploymentsHealthCheck: runtime instance is unhealthy", f...)
	}
	return nil
}

func runtimeUnhealthy(r *runtimev1.HealthResponse) bool {
	return r.LimiterError != "" || r.ConnCacheError != "" || r.MetastoreError != "" || r.NetworkError != ""
}

func instanceUnhealthy(i *runtimev1.InstanceHealth) bool {
	return i.OlapError != "" || i.ControllerError != "" || i.RepoError != ""
}
