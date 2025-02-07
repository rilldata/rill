package server

import (
	"context"
	"fmt"

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

	// Resolver should exist
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

	// Convert the results to a proto response
	schema := res.Schema()
	data := make([]*structpb.Struct, 0)
	for {
		raw, err := res.Next()
		if err != nil {
			return nil, err
		}
		if raw == nil {
			break
		}
		row, err := structpb.NewStruct(raw)
		if err != nil {
			return nil, fmt.Errorf("failed to convert row to proto: %w", err)
		}
		data = append(data, row)
	}

	return &runtimev1.QueryResolverResponse{
		Schema: schema,
		Data:   data,
	}, nil
}
