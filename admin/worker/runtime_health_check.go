package worker

import (
	"context"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/encoding/protojson"
)

const _allocatedRuntimesPageSize = 20

func (w *Worker) runtimeHealthCheck(ctx context.Context) error {
	lastRuntime := ""
	for {
		runtimes, err := w.admin.DB.FindAllocatedRuntimes(ctx, lastRuntime, _allocatedRuntimesPageSize)
		if err != nil {
			return fmt.Errorf("failed to get runtimes: %w", err)
		}
		if len(runtimes) == 0 {
			return nil
		}
		lastRuntime = runtimes[len(runtimes)-1].Host

		group, cctx := errgroup.WithContext(ctx)
		group.SetLimit(8)
		for _, rt := range runtimes {
			rt := rt
			group.Go(func() error {
				client, err := w.admin.OpenRuntimeClient(rt.Host, rt.Audience)
				if err != nil {
					w.logger.Error("runtimeHealthCheck: failed to open runtime client", zap.String("host", rt.Host), zap.Error(err))
					return nil
				}

				resp, err := client.Health(cctx, &runtimev1.HealthRequest{})
				if err != nil {
					client.Close()
					w.logger.Error("runtimeHealthCheck: health check call failed", zap.String("host", rt.Host), zap.Error(err))
					return nil
				}

				if !isRuntimeHealthy(resp) {
					s, _ := protojson.Marshal(resp)
					w.logger.Error("runtimeHealthCheck: runtime is unhealthy", zap.String("host", rt.Host), zap.ByteString("health_response", s))
				}
				client.Close()
				return nil
			})
		}
		if err := group.Wait(); err != nil {
			return err
		}
		if len(runtimes) < _allocatedRuntimesPageSize {
			return nil
		}
		// fetch again
	}
}

func isRuntimeHealthy(r *runtimev1.HealthResponse) bool {
	if r.LimiterError != "" || r.ConnCacheError != "" || r.MetastoreError != "" || r.NetworkError != "" {
		return false
	}
	for _, v := range r.InstancesHealth {
		if v.ControllerError != "" || v.OlapError != "" || v.RepoError != "" {
			return false
		}
	}
	return true
}
