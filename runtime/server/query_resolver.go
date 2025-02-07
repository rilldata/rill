package server

import (
	"context"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/server/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

// QueryResolver enables admins to query a resolver within a project
func (s *Server) QueryResolver(ctx context.Context, req *runtimev1.QueryResolverRequest) (*runtimev1.QueryResolverResponse, error) {
	// Validate the caller is an admin of the project
	claims := auth.GetClaims(ctx)
	if !claims.CanInstance(req.InstanceId, auth.ManageInstances) {
		return nil, status.Error(codes.PermissionDenied, "must be an admin of the project to query a resolver")
	}

	// Validate the resolver exists
	initializer, ok := runtime.ResolverInitializers[req.Resolver]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "no resolver found of type %q", req.Resolver)
	}

	// Initialize the resolver
	resolver, err := initializer(ctx, &runtime.ResolverOptions{
		Runtime:    s.runtime,
		InstanceID: req.InstanceId,
		Properties: req.ResolverProperties.AsMap(),
		Args:       req.ResolverArgs.AsMap(),
		Claims:     claims.SecurityClaims(),
		ForExport:  false,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer resolver.Close()

	// Query the resolver
	res, err := resolver.ResolveInteractive(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer res.Close()

	var data []*structpb.Struct

	// TODO: Read the resolver's output and return it as a response

	return &runtimev1.QueryResolverResponse{
		Data: data,
	}, nil
}
