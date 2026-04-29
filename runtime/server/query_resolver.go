package server

import (
	"context"
	"errors"
	"io"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/pbutil"
	"github.com/rilldata/rill/runtime/server/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

// QueryResolver enables superusers and project admins to query a resolver within a project
func (s *Server) QueryResolver(ctx context.Context, req *runtimev1.QueryResolverRequest) (*runtimev1.QueryResolverResponse, error) {
	// Validate the caller has the ReadResolvers permission
	claims := auth.GetClaims(ctx, req.InstanceId)
	if !claims.Can(runtime.ReadResolvers) {
		return nil, status.Error(codes.PermissionDenied, "only project admins can query resolvers")
	}

	// Resolver should exist
	initializer, ok := runtime.ResolverInitializers[req.Resolver]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "no resolver found of type %q", req.Resolver)
	}

	// Inject limit into the props.
	// Note: Not all resolvers support `limit` being passed here, but it's better than nothing.
	// In case the resolver does not apply the limit, we fall back to applying it when reading the results later in this handler.
	props := req.ResolverProperties.AsMap()
	if req.Limit != 0 {
		props["limit"] = req.Limit
	}

	// Initialize the resolver
	resolver, err := initializer(ctx, &runtime.ResolverOptions{
		Runtime:    s.runtime,
		InstanceID: req.InstanceId,
		Properties: props,
		Args:       req.ResolverArgs.AsMap(),
		Claims:     claims,
		ForExport:  false,
	})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	defer resolver.Close()

	// Query the resolver
	res, err := resolver.ResolveInteractive(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	defer res.Close()

	data := make([]*structpb.Struct, 0)
	count := 0
	for {
		// Break if we've reached the limit (when limit > 0)
		if req.Limit > 0 && count >= int(req.Limit) {
			break
		}

		// Next returns the next row of data. It returns io.EOF when there are no more rows.
		row, err := res.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		rowStruct, err := pbutil.ToStruct(row, res.Schema())
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		data = append(data, rowStruct)
		count++
	}

	// Return the response
	metaPB, err := pbutil.ToStruct(res.Meta(), nil)
	if err != nil {
		metaPB = &structpb.Struct{Fields: make(map[string]*structpb.Value)}
	}
	return &runtimev1.QueryResolverResponse{
		Meta:   metaPB,
		Schema: res.Schema(),
		Data:   data,
	}, nil
}
