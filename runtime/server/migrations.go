package server

import (
	"context"

	"github.com/rilldata/rill/runtime/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Migrate implements RuntimeService
func (s *Server) Migrate(ctx context.Context, req *api.MigrateRequest) (*api.MigrateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")
}

// MigrateSingle implements RuntimeService
func (s *Server) MigrateSingle(ctx context.Context, req *api.MigrateSingleRequest) (*api.MigrateSingleResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")
}

// MigrateDelete implements RuntimeService
func (s *Server) MigrateDelete(ctx context.Context, req *api.MigrateDeleteRequest) (*api.MigrateDeleteResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")
}
