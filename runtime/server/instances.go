package server

import (
	"context"

	"github.com/rilldata/rill/runtime/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ListInstances implements RuntimeService
func (s *Server) ListInstances(ctx context.Context, req *api.ListInstancesRequest) (*api.ListInstancesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")
}

// GetInstance implements RuntimeService
func (s *Server) GetInstance(ctx context.Context, req *api.GetInstanceRequest) (*api.GetInstanceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")
}

// CreateInstance implements RuntimeService
func (s *Server) CreateInstance(ctx context.Context, req *api.CreateInstanceRequest) (*api.CreateInstanceResponse, error) {
	instance, err := s.runtime.CreateInstance(req.Driver, req.Dsn)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	resp := &api.CreateInstanceResponse{
		InstanceId: instance.ID.String(),
		Instance: &api.Instance{
			InstanceId: instance.ID.String(),
		},
	}

	return resp, nil

}

// DeleteInstance implements RuntimeService
func (s *Server) DeleteInstance(ctx context.Context, req *api.DeleteInstanceRequest) (*api.DeleteInstanceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")
}
