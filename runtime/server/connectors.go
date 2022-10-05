package server

import (
	"context"

	"github.com/rilldata/rill/runtime/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ListConnectors implements RuntimeService
func (s *Server) ListConnectors(ctx context.Context, req *api.ListConnectorsRequest) (*api.ListConnectorsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")
}
