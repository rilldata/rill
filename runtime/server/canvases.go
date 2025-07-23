package server

import (
	"context"
	"errors"
	"fmt"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/parser"
	metricssqlparser "github.com/rilldata/rill/runtime/pkg/metricssql"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

func (s *Server) ResolveCanvas(ctx context.Context, req *runtimev1.ResolveCanvasRequest) (*runtimev1.ResolveCanvasResponse, error) {
	canvas, ctrl, templateData, err := s.canvasResolution(ctx, req)
	if err != nil {
		return nil, err
	}

	components, err := s.resolveCanvasComponents(ctx, ctrl, canvas, templateData)
	if err != nil {
		return nil, err
	}

	metricsViews, err := s.resolveMetricsViews(ctx, ctrl, req.InstanceId, components)
	if err != nil {
		return nil, err
	}

	return &runtimev1.ResolveCanvasResponse{
		Canvas:                 canvas,
		ResolvedComponents:     components,
		ReferencedMetricsViews: metricsViews,
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
	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.ReadAPI) {
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
		User:        auth.GetClaims(ctx).SecurityClaims().UserAttributes,
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

func (s *Server) canvasResolution(ctx context.Context, req *runtimev1.ResolveCanvasRequest) (*runtimev1.Resource, *runtime.Controller, parser.TemplateData, error) {
	// Add observability attributes
	s.addInstanceRequestAttributes(ctx, req.InstanceId)
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.canvas", req.Canvas),
	)

	// Check if user has access to query for canvas data (we use the ReadAPI permission for this for now)
	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.ReadAPI) {
		return nil, nil, parser.TemplateData{}, status.Errorf(codes.FailedPrecondition, "does not have access to canvas data")
	}

	// Find the canvas resource
	ctrl, err := s.runtime.Controller(ctx, req.InstanceId)
	if err != nil {
		return nil, nil, parser.TemplateData{}, status.Error(codes.Internal, err.Error())
	}
	canvas, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindCanvas, Name: req.Canvas}, false)
	if err != nil {
		if errors.Is(err, drivers.ErrResourceNotFound) {
			return nil, nil, parser.TemplateData{}, status.Errorf(codes.NotFound, "canvas with name %q not found", req.Canvas)
		}
		return nil, nil, parser.TemplateData{}, status.Error(codes.Internal, err.Error())
	}
	spec := canvas.GetCanvas().State.ValidSpec
	if spec == nil {
		return canvas, ctrl, parser.TemplateData{}, nil
	}

	// Setup templating data
	inst, err := s.runtime.Instance(ctx, req.InstanceId)
	if err != nil {
		return nil, nil, parser.TemplateData{}, status.Error(codes.Internal, err.Error())
	}
	td := parser.TemplateData{
		Environment: inst.Environment,
		User:        auth.GetClaims(ctx).SecurityClaims().UserAttributes,
		Variables:   inst.ResolveVariables(false),
		ExtraProps: map[string]any{
			"args": req.Args.AsMap(),
		},
	}

	return canvas, ctrl, td, nil
}

func (s *Server) resolveCanvasComponents(ctx context.Context, ctrl *runtime.Controller, canvas *runtimev1.Resource, td parser.TemplateData) (map[string]*runtimev1.Resource, error) {
	components := make(map[string]*runtimev1.Resource)
	spec := canvas.GetCanvas().State.ValidSpec

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
				v, err := parser.ResolveTemplateRecursively(validSpec.RendererProperties.AsMap(), td, false)
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

	return components, nil
}

func (s *Server) resolveMetricsViews(ctx context.Context, ctrl *runtime.Controller, instanceID string, components map[string]*runtimev1.Resource) (map[string]*runtimev1.Resource, error) {
	// Extract metrics view names from components
	allMVNames := s.extractMetricsViewNames(ctx, ctrl, instanceID, components)

	// Lookup metrics view resources
	metricsViews := make(map[string]*runtimev1.Resource)
	for _, mvName := range allMVNames {
		mv, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: mvName}, false)
		if err != nil {
			if errors.Is(err, drivers.ErrResourceNotFound) {
				return nil, status.Errorf(codes.Internal, "metrics view %q in valid spec not found", mvName)
			}
			return nil, err
		}

		// Add to map.
		metricsViews[mvName] = mv
	}

	return metricsViews, nil
}

func (s *Server) extractMetricsViewNames(ctx context.Context, ctrl *runtime.Controller, instanceID string, components map[string]*runtimev1.Resource) []string {
	allMVNames := make(map[string]bool)

	for cmpName, cmp := range components {
		mvNames := s.getMetricsViewNamesFromComponent(ctx, ctrl, instanceID, cmpName, cmp)

		for _, mvName := range mvNames {
			if mvName != "" {
				allMVNames[mvName] = true
			}
		}
	}

	result := make([]string, 0, len(allMVNames))
	for mvName := range allMVNames {
		result = append(result, mvName)
	}

	return result
}

func (s *Server) getMetricsViewNamesFromComponent(ctx context.Context, ctrl *runtime.Controller, instanceID, cmpName string, cmp *runtimev1.Resource) []string {
	var mvNames []string
	validSpec := cmp.GetComponent().State.ValidSpec
	if validSpec == nil || validSpec.RendererProperties == nil {
		return mvNames
	}

	for k, v := range validSpec.RendererProperties.Fields {
		switch k {
		case "metrics_view":
			if name := v.GetStringValue(); name != "" {
				mvNames = append(mvNames, name)
			}
		case "metrics_sql":
			mvName, err := parseMetricsViewsFromSQL(ctx, ctrl, instanceID, v.GetStringValue())
			if err != nil {
				observability.AddRequestAttributes(ctx,
					attribute.String("sql_parse_error.component", cmpName),
					attribute.String("sql_parse_error.message", err.Error()),
				)
				// Don't fail the entire canvas resolution if one component has an invalid SQL
				continue
			}
			if mvName != "" {
				mvNames = append(mvNames, mvName)
			}
		}
	}

	return mvNames
}

func parseMetricsViewsFromSQL(ctx context.Context, ctrl *runtime.Controller, instanceID, sql string) (string, error) {
	// basic validations
	sql = strings.TrimSpace(sql)
	if sql == "" {
		return "", fmt.Errorf("SQL query cannot be empty")
	}

	if len(sql) > 10000 {
		return "", fmt.Errorf("SQL query too large")
	}

	if !strings.HasPrefix(strings.ToUpper(sql), "SELECT") {
		return "", fmt.Errorf("only SELECT statements are allowed")
	}

	// Parse the SQL to extract metrics view
	claims := auth.GetClaims(ctx).SecurityClaims()
	compiler := metricssqlparser.New(ctrl, instanceID, claims, 0)
	q, err := compiler.Rewrite(ctx, sql)
	if err != nil {
		return "", fmt.Errorf("failed to parse metrics SQL: %w", err)
	}

	return q.MetricsView, nil
}
