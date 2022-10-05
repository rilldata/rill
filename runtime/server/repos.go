package server

import (
	"context"

	"github.com/rilldata/rill/runtime/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ListRepos implements RuntimeService
func (s *Server) ListRepos(ctx context.Context, req *api.ListReposRequest) (*api.ListReposResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")
}

// GetRepo implements RuntimeService
func (s *Server) GetRepo(ctx context.Context, req *api.GetRepoRequest) (*api.GetRepoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")
}

// CreateRepo implements RuntimeService
func (s *Server) CreateRepo(ctx context.Context, req *api.CreateRepoRequest) (*api.CreateRepoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")
}

// DeleteRepo implements RuntimeService
func (s *Server) DeleteRepo(ctx context.Context, req *api.DeleteRepoRequest) (*api.DeleteRepoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")
}

// ListRepoObjects implements RuntimeService
func (s *Server) ListRepoObjects(ctx context.Context, req *api.ListRepoObjectsRequest) (*api.ListRepoObjectsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")
}

// GetRepoObject implements RuntimeService
func (s *Server) GetRepoObject(ctx context.Context, req *api.GetRepoObjectRequest) (*api.GetRepoObjectResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")
}

// PutRepoObject implements RuntimeService
func (s *Server) PutRepoObject(ctx context.Context, req *api.PutRepoObjectRequest) (*api.PutRepoObjectResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")
}
