package runtime

import (
	"context"
	"errors"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/metricsview/metricssql"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (r *Runtime) ResolveCanvas(ctx context.Context, instanceID, canvas string, claims *SecurityClaims) (*runtimev1.ResolveCanvasResponse, error) {
	// Find the canvas resource
	ctrl, err := r.Controller(ctx, instanceID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	res, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: ResourceKindCanvas, Name: canvas}, false)
	if err != nil {
		if errors.Is(err, drivers.ErrResourceNotFound) {
			return nil, status.Errorf(codes.NotFound, "canvas with name %q not found", canvas)
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Check if the user has access to the canvas
	res, access, err := r.ApplySecurityPolicy(ctx, instanceID, claims, res)
	if err != nil {
		return nil, fmt.Errorf("failed to apply security policy: %w", err)
	}
	if !access {
		return nil, status.Errorf(codes.PermissionDenied, "user does not have access to canvas %q", canvas)
	}

	// Exit early if the canvas is not valid
	spec := res.GetCanvas().State.ValidSpec
	if spec == nil {
		return &runtimev1.ResolveCanvasResponse{
			Canvas: res,
		}, nil
	}

	components := make(map[string]*runtimev1.Resource)

	for _, row := range spec.Rows {
		for _, item := range row.Items {
			// Skip if already resolved.
			if _, ok := components[item.Component]; ok {
				continue
			}

			// Get component resource.
			cmp, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: ResourceKindComponent, Name: item.Component}, false)
			if err != nil {
				if errors.Is(err, drivers.ErrResourceNotFound) {
					return nil, status.Errorf(codes.Internal, "component %q in valid spec not found", item.Component)
				}
				return nil, err
			}

			// Add to map without resolving templates. Use ResolveTemplatedString RPC for template resolution.
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
							mv, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: ResourceKindMetricsView, Name: name}, false)
							if err != nil {
								return nil, err
							}
							sec, err := r.ResolveSecurity(ctx, ctrl.InstanceID, claims, mv)
							if err != nil {
								return nil, err
							}
							if !sec.CanAccess() {
								return nil, ErrForbidden
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
		mv, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: ResourceKindMetricsView, Name: mvName}, false)
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
		mv, access, err := r.ApplySecurityPolicy(ctx, instanceID, claims, mv)
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
