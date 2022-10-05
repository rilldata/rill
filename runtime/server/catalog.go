package server

import (
	"context"

	"github.com/rilldata/rill/runtime/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ListCatalogObjects implements RuntimeService
func (s *Server) ListCatalogObjects(ctx context.Context, req *api.ListCatalogObjectsRequest) (*api.ListCatalogObjectsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")
}

// GetCatalogObject implements RuntimeService
func (s *Server) GetCatalogObject(ctx context.Context, req *api.GetCatalogObjectRequest) (*api.GetCatalogObjectResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")
}

// TriggerRefresh implements RuntimeService
func (s *Server) TriggerRefresh(ctx context.Context, req *api.TriggerRefreshRequest) (*api.TriggerRefreshResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")
}
