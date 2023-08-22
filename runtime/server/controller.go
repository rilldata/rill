package server

import (
	"context"

	"github.com/bufbuild/connect-go"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

// GetLogs implements runtimev1.RuntimeServiceServer
func (s *Server) GetLogs(ctx context.Context, req *connect.Request[runtimev1.GetLogsRequest]) (*connect.Response[runtimev1.GetLogsResponse], error) {
	panic("not implemented")
}

// WatchLogs implements runtimev1.RuntimeServiceServer
func (s *Server) WatchLogs(ctx context.Context, req *connect.Request[runtimev1.WatchLogsRequest], srv *connect.ServerStream[runtimev1.WatchLogsResponse]) error {
	panic("not implemented")
}

// ListResources implements runtimev1.RuntimeServiceServer
func (s *Server) ListResources(ctx context.Context, req *connect.Request[runtimev1.ListResourcesRequest]) (*connect.Response[runtimev1.ListResourcesResponse], error) {
	panic("not implemented")
}

// WatchResources implements runtimev1.RuntimeServiceServer
func (s *Server) WatchResources(ctx context.Context, req *connect.Request[runtimev1.WatchResourcesRequest], srv *connect.ServerStream[runtimev1.WatchResourcesResponse]) error {
	panic("not implemented")
}

// GetResource implements runtimev1.RuntimeServiceServer
func (s *Server) GetResource(ctx context.Context, req *connect.Request[runtimev1.GetResourceRequest]) (*connect.Response[runtimev1.GetResourceResponse], error) {
	panic("not implemented")
}

// CreateTrigger implements runtimev1.RuntimeServiceServer
func (s *Server) CreateTrigger(ctx context.Context, req *connect.Request[runtimev1.CreateTriggerRequest]) (*connect.Response[runtimev1.CreateTriggerResponse], error) {
	panic("not implemented")
}
