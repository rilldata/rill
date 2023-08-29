package server

import (
	"context"
	"fmt"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"golang.org/x/sync/errgroup"
)

// GetLogs implements runtimev1.RuntimeServiceServer
func (s *Server) GetLogs(ctx context.Context, req *runtimev1.GetLogsRequest) (*runtimev1.GetLogsResponse, error) {
	panic("not implemented")
}

// WatchLogs implements runtimev1.RuntimeServiceServer
func (s *Server) WatchLogs(req *runtimev1.WatchLogsRequest, srv runtimev1.RuntimeService_WatchLogsServer) error {
	panic("not implemented")
}

// ListResources implements runtimev1.RuntimeServiceServer
func (s *Server) ListResources(ctx context.Context, req *runtimev1.ListResourcesRequest) (*runtimev1.ListResourcesResponse, error) {
	panic("not implemented")
}

// WatchResources implements runtimev1.RuntimeServiceServer
func (s *Server) WatchResources(req *runtimev1.WatchResourcesRequest, srv runtimev1.RuntimeService_WatchResourcesServer) error {
	// Temporary code for testing
	ctx := srv.Context()

	catalog, catalogRelease, err := s.runtime.Catalog(ctx, req.InstanceId)
	if err != nil {
		return err
	}
	defer catalogRelease()

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		lastRun := time.Now()

		for {
			entries, err := catalog.FindEntries(ctx, drivers.ObjectTypeUnspecified)
			if err != nil {
				return err
			}

			for _, entry := range entries {
				if entry.UpdatedOn.Before(lastRun) {
					continue
				}

				r := &runtimev1.Resource{
					Meta: &runtimev1.ResourceMeta{
						Name: &runtimev1.ResourceName{
							Kind: "catalog",
							Name: entry.Name,
						},
						FilePaths: []string{entry.Path},
					},
				}
				err = srv.Send(&runtimev1.WatchResourcesResponse{
					Resource: r,
				})
				if err != nil {
					return err
				}
			}

			if ctx.Err() != nil {
				fmt.Println(ctx.Err())
				return ctx.Err()
			}
			lastRun = time.Now()
			time.Sleep(time.Second)
		}
	})

	return g.Wait()
}

// GetResource implements runtimev1.RuntimeServiceServer
func (s *Server) GetResource(ctx context.Context, req *runtimev1.GetResourceRequest) (*runtimev1.GetResourceResponse, error) {
	panic("not implemented")
}

// CreateTrigger implements runtimev1.RuntimeServiceServer
func (s *Server) CreateTrigger(ctx context.Context, req *runtimev1.CreateTriggerRequest) (*runtimev1.CreateTriggerResponse, error) {
	panic("not implemented")
}
