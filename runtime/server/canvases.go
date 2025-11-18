package server

import (
	"context"
	"errors"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/metricsview/metricssql"
	"github.com/rilldata/rill/runtime/parser"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

func (s *Server) ResolveCanvas(ctx context.Context, req *runtimev1.ResolveCanvasRequest) (*runtimev1.ResolveCanvasResponse, error) {
	// Add observability attributes
	s.addInstanceRequestAttributes(ctx, req.InstanceId)
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.canvas", req.Canvas),
	)

	// Check if user has access to query for canvas data (we use the ReadAPI permission for this for now)
	claims := auth.GetClaims(ctx, req.InstanceId)
	if !claims.Can(runtime.ReadAPI) {
		return nil, status.Errorf(codes.FailedPrecondition, "does not have access to canvas data")
	}

	// Find the canvas resource
	ctrl, err := s.runtime.Controller(ctx, req.InstanceId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	res, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindCanvas, Name: req.Canvas}, false)
	if err != nil {
		if errors.Is(err, drivers.ErrResourceNotFound) {
			return nil, status.Errorf(codes.NotFound, "canvas with name %q not found", req.Canvas)
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Check if the user has access to the canvas
	res, access, err := s.runtime.ApplySecurityPolicy(ctx, req.InstanceId, claims, res)
	if err != nil {
		return nil, fmt.Errorf("failed to apply security policy: %w", err)
	}
	if !access {
		return nil, status.Errorf(codes.PermissionDenied, "user does not have access to canvas %q", req.Canvas)
	}

	// Exit early if the canvas is not valid
	spec := res.GetCanvas().State.ValidSpec
	if spec == nil {
		return &runtimev1.ResolveCanvasResponse{
			Canvas: res,
		}, nil
	}

	// Setup templating data
	inst, err := s.runtime.Instance(ctx, req.InstanceId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	templateData := parser.TemplateData{
		Environment: inst.Environment,
		User:        claims.UserAttributes,
		Variables:   inst.ResolveVariables(false),
		ExtraProps: map[string]any{
			"args": req.Args.AsMap(),
		},
	}

	components := make(map[string]*runtimev1.Resource)

	for _, row := range spec.Rows {
		for _, item := range row.Items {
			// Skip if already resolved.
			if _, ok := components[item.Component]; ok {
				continue
			}

			// Get component resource.
			// NOTE: By passing true, we get a cloned object that is safe to modify in-place.
			cmp, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindComponent, Name: item.Component}, true)
			if err != nil {
				if errors.Is(err, drivers.ErrResourceNotFound) {
					return nil, status.Errorf(codes.Internal, "component %q in valid spec not found", item.Component)
				}
				return nil, err
			}

			// Resolve the renderer properties in the valid_spec.
			validSpec := cmp.GetComponent().State.ValidSpec
			if validSpec != nil && validSpec.RendererProperties != nil {
				v, err := parser.ResolveTemplateRecursively(validSpec.RendererProperties.AsMap(), templateData, false)
				if err != nil {
					return nil, status.Errorf(codes.InvalidArgument, "component %q: failed to resolve templating: %s", item.Component, err.Error())
				}

				props, ok := v.(map[string]any)
				if !ok {
					return nil, status.Errorf(codes.Internal, "component %q: failed to convert resolved renderer properties to map: %v", item.Component, v)
				}

				propsPB, err := structpb.NewStruct(props)
				if err != nil {
					return nil, status.Errorf(codes.Internal, "component %q: failed to convert renderer properties to struct: %s", item.Component, err.Error())
				}

				validSpec.RendererProperties = propsPB
			}

			// Add to map.
			components[item.Component] = cmp
		}
	}

	// Extract metrics view names from components
	var msqlParser *metricssql.Compiler
	metricsViews := make(map[string]bool)
	for _, cmp := range components {
		validSpec := cmp.GetComponent().State.ValidSpec
		if validSpec == nil || validSpec.RendererProperties == nil {
			continue
		}

		for k, v := range validSpec.RendererProperties.Fields {
			switch k {
			case "metrics_view":
				if name := v.GetStringValue(); name != "" {
					metricsViews[name] = true
				}
			case "metrics_sql":
				// Instantiate a metrics SQL parser
				if msqlParser == nil {
					msqlParser = metricssql.New(&metricssql.CompilerOptions{
						GetMetricsView: func(ctx context.Context, name string) (*runtimev1.Resource, error) {
							mv, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: name}, false)
							if err != nil {
								return nil, err
							}
							sec, err := s.runtime.ResolveSecurity(ctx, ctrl.InstanceID, claims, mv)
							if err != nil {
								return nil, err
							}
							if !sec.CanAccess() {
								return nil, runtime.ErrForbidden
							}
							return mv, nil
						},
					})
				}

				// Create list of queries to analyze
				var queries []string
				if s := v.GetStringValue(); s != "" {
					queries = append(queries, s)
				} else if vals := v.GetListValue(); vals != nil {
					for _, val := range vals.Values {
						if s := val.GetStringValue(); s != "" {
							queries = append(queries, s)
						}
					}
				}

				// Analyze each query
				for _, sql := range queries {
					q, err := msqlParser.Parse(ctx, sql)
					if err == nil && q.MetricsView != "" {
						metricsViews[q.MetricsView] = true
					}
				}
			}
		}
	}

	// Lookup metrics view resources
	referencedMetricsViews := make(map[string]*runtimev1.Resource)
	for mvName := range metricsViews {
		mv, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: mvName}, false)
		if err != nil {
			if errors.Is(err, drivers.ErrResourceNotFound) {
				return nil, status.Errorf(codes.Internal, "metrics view %q in valid spec not found", mvName)
			}
			return nil, err
		}

		// Add to map.
		referencedMetricsViews[mvName] = mv
	}

	// Apply security policies to the metrics views.
	for name, mv := range referencedMetricsViews {
		mv, access, err := s.runtime.ApplySecurityPolicy(ctx, req.InstanceId, claims, mv)
		if err != nil {
			return nil, fmt.Errorf("failed to apply security policy: %w", err)
		}
		if !access {
			delete(referencedMetricsViews, name)
			continue
		}
		referencedMetricsViews[name] = mv
	}

	// Return the response
	return &runtimev1.ResolveCanvasResponse{
		Canvas:                 res,
		ResolvedComponents:     components,
		ReferencedMetricsViews: referencedMetricsViews,
	}, nil
}

func (s *Server) ResolveComponent(ctx context.Context, req *runtimev1.ResolveComponentRequest) (*runtimev1.ResolveComponentResponse, error) {
	// Add observability attributes
	s.addInstanceRequestAttributes(ctx, req.InstanceId)
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.component", req.Component),
	)

	// Check if user has access to query for component data (we use the ReadAPI permission for this for now)
	claims := auth.GetClaims(ctx, req.InstanceId)
	if !claims.Can(runtime.ReadAPI) {
		return nil, status.Errorf(codes.FailedPrecondition, "does not have access to component data")
	}

	// Find the component spec
	ctrl, err := s.runtime.Controller(ctx, req.InstanceId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	res, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindComponent, Name: req.Component}, false)
	if err != nil {
		if errors.Is(err, drivers.ErrResourceNotFound) {
			return nil, status.Errorf(codes.NotFound, "component with name %q not found", req.Component)
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	spec := res.GetComponent().State.ValidSpec
	if spec == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "component %q is invalid", req.Component)
	}

	// Get current instance metadata
	inst, err := s.runtime.Instance(ctx, req.InstanceId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Parse args
	args := req.Args.AsMap()

	// Setup templating data
	td := parser.TemplateData{
		Environment: inst.Environment,
		User:        claims.UserAttributes,
		Variables:   inst.ResolveVariables(false),
		ExtraProps: map[string]any{
			"args": args,
		},
	}

	// Resolve templating in the renderer properties
	var rendererProps *structpb.Struct
	if spec.RendererProperties != nil {
		v, err := parser.ResolveTemplateRecursively(spec.RendererProperties.AsMap(), td, false)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		props, ok := v.(map[string]any)
		if !ok {
			return nil, status.Errorf(codes.Internal, "failed to convert resolved renderer properties to map: %v", v)
		}

		rendererProps, err = structpb.NewStruct(props)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to convert renderer properties to struct: %s", err.Error())
		}
	}

	// Return the response
	return &runtimev1.ResolveComponentResponse{
		RendererProperties: rendererProps,
	}, nil
}
