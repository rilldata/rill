package admin

import (
	"context"
	"fmt"
	"time"

	"github.com/rilldata/rill/admin/database"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TriggerReconcileAndAwaitResource triggers a reconcile and polls the runtime until the given resource's spec version has been updated (or ctx is canceled).
func (s *Service) TriggerReconcileAndAwaitResource(ctx context.Context, depl *database.Deployment, name, kind string) error {
	rt, err := s.openRuntimeClientForDeployment(depl)
	if err != nil {
		return err
	}
	defer rt.Close()

	resourceReq := &runtimev1.GetResourceRequest{
		InstanceId: depl.RuntimeInstanceID,
		Name: &runtimev1.ResourceName{
			Kind: kind,
			Name: name,
		},
	}

	// Get old spec version
	var oldSpecVersion *int64
	r, err := rt.GetResource(ctx, resourceReq)
	if err == nil {
		oldSpecVersion = &r.Resource.Meta.SpecVersion
	}

	// Trigger reconcile
	_, err = rt.CreateTrigger(ctx, &runtimev1.CreateTriggerRequest{
		InstanceId: depl.RuntimeInstanceID,
		Trigger: &runtimev1.CreateTriggerRequest_PullTriggerSpec{
			PullTriggerSpec: &runtimev1.PullTriggerSpec{},
		},
	})
	if err != nil {
		return err
	}

	// Poll every 1 seconds until the resource is found or the ctx is cancelled or times out
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}

		r, err := rt.GetResource(ctx, resourceReq)
		if err != nil {
			if s, ok := status.FromError(err); !ok || s.Code() != codes.NotFound {
				return fmt.Errorf("failed to poll for resource: %w", err)
			}
			if oldSpecVersion != nil {
				// Success - previously the resource was found, now we cannot find it anymore
				return nil
			}
			// Continue polling
			continue
		}
		if oldSpecVersion == nil {
			// Success - previously the resource was not found, now we found one
			return nil
		}
		if *oldSpecVersion != r.Resource.Meta.SpecVersion {
			// Success - the spec version has changed
			return nil
		}
	}
}
