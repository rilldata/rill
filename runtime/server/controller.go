package server

import (
	"context"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
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
	panic("not implemented")
}

// GetResource implements runtimev1.RuntimeServiceServer
func (s *Server) GetResource(ctx context.Context, req *runtimev1.GetResourceRequest) (*runtimev1.GetResourceResponse, error) {
	panic("not implemented")
}

// CreateTrigger implements runtimev1.RuntimeServiceServer
func (s *Server) CreateTrigger(ctx context.Context, req *runtimev1.CreateTriggerRequest) (*runtimev1.CreateTriggerResponse, error) {
	panic("not implemented")
}
