package server

import (
	"context"
	"errors"
	"io"
	"slices"

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
	// Validate permissions
	claims := auth.GetClaims(ctx, req.InstanceId)
	switch req.Resolver {
	case "metrics", "metrics_sql":
		// As a special case, we allow metrics resolvers for users with ReadMetrics permission (i.e. all users)
		if !claims.Can(runtime.ReadMetrics) {
			return nil, status.Error(codes.PermissionDenied, "not allowed to query metrics")
		}
	default:
		// Other resolvers require ReadResolvers permission (i.e. project admin)
		if !claims.Can(runtime.ReadResolvers) {
			return nil, status.Error(codes.PermissionDenied, "only project admins can query resolvers")
		}
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

	// Determine who wrote the resolver properties, which decides whether templates in them may access project variables.
	// They come from the caller, unless the request names the component whose saved properties they were taken from.
	origin := runtime.PropertiesFromCaller
	if req.Component != "" {
		var err error
		origin, err = s.componentQueryOrigin(ctx, req.InstanceId, req.Component, req.Resolver, props, claims)
		if err != nil {
			return nil, err
		}
	}

	// Initialize the resolver.
	resolver, err := initializer(ctx, &runtime.ResolverOptions{
		Runtime:          s.runtime,
		InstanceID:       req.InstanceId,
		Properties:       props,
		PropertiesOrigin: origin,
		Args:             req.ResolverArgs.AsMap(),
		Claims:           claims,
		ForExport:        false,
	})
	if err != nil {
		return nil, mapGRPCErrorWithFallback(err, codes.InvalidArgument)
	}
	defer resolver.Close()

	// Query the resolver
	res, err := resolver.ResolveInteractive(ctx)
	if err != nil {
		return nil, mapGRPCErrorWithFallback(err, codes.InvalidArgument)
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
			return nil, err
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

// componentQueryOrigin checks that the resolver properties were taken from the saved properties of the given component,
// which the caller must be able to access, and returns who wrote the component's properties (see runtime.PropertiesOrigin).
// Only the metrics_sql resolver is supported: the SQL must match one of the component's metrics_sql queries.
// This lets canvas charts run the queries declared for them with the templating of a project-declared component, without the runtime ever returning the resolved SQL.
func (s *Server) componentQueryOrigin(ctx context.Context, instanceID, component, resolver string, props map[string]any, claims *runtime.SecurityClaims) (runtime.PropertiesOrigin, error) {
	if resolver != "metrics_sql" {
		return 0, status.Errorf(codes.InvalidArgument, "a component can only be referenced by the metrics_sql resolver, not %q", resolver)
	}
	res, origin, err := s.accessibleComponent(ctx, instanceID, component, claims)
	if err != nil {
		return 0, err
	}
	spec := res.GetComponent().State.ValidSpec
	if spec == nil {
		return 0, status.Errorf(codes.FailedPrecondition, "component %q is invalid", component)
	}

	var queries []string
	switch v := spec.RendererProperties.AsMap()["metrics_sql"].(type) {
	case string:
		queries = []string{v}
	case []any:
		for _, q := range v {
			if q, ok := q.(string); ok {
				queries = append(queries, q)
			}
		}
	}
	sql, _ := props["sql"].(string)
	if sql == "" || !slices.Contains(queries, sql) {
		return 0, status.Errorf(codes.InvalidArgument, "the SQL does not match a metrics_sql query of component %q", component)
	}
	return origin, nil
}
