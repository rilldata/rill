package server

import (
	"context"
	"errors"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/canvas"
	"github.com/rilldata/rill/runtime/drivers"
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
		return nil, status.Error(codes.PermissionDenied, "does not have access to canvas data")
	}

	res, err := s.runtime.ResolveCanvas(ctx, req.InstanceId, req.Canvas, claims, req.Unsafe)
	if err != nil {
		return nil, err
	}

	return &runtimev1.ResolveCanvasResponse{
		Canvas:                 res.Canvas,
		ResolvedComponents:     res.ResolvedComponents,
		ReferencedMetricsViews: res.ReferencedMetricsViews,
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
		return nil, status.Error(codes.PermissionDenied, "does not have access to component data")
	}

	// Find the component spec
	ctrl, err := s.runtime.Controller(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}
	res, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindComponent, Name: req.Component}, false)
	if err != nil {
		return nil, err
	}
	spec := res.GetComponent().State.ValidSpec
	if spec == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "component %q is invalid", req.Component)
	}

	// Get current instance metadata
	inst, err := s.runtime.Instance(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}

	// Parse args and validate them against the component's declared params.
	// Components without declared params accept arbitrary args for backwards compatibility.
	args := req.Args.AsMap()
	if len(spec.Params) > 0 {
		if err := canvas.ValidateParamBindings(spec.Params, args, nil); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}

	// Merge the declared param defaults (and legacy input variable defaults) under the provided args.
	effectiveArgs := canvas.EffectiveArgs(spec, args)

	// Resolve the metrics view metadata of the bound fields, so a spec can title an axis with a
	// measure's display name or format it with the measure's formatter. See canvas.FieldTemplateData.
	metricsViews, err := s.boundMetricsViews(ctx, req.InstanceId, claims, spec.Params, effectiveArgs)
	if err != nil {
		return nil, err
	}

	// Setup templating data.
	// The effective args are exposed both as .params (canonical, matching the YAML key) and .args (legacy alias).
	td := parser.TemplateData{
		Environment: inst.Environment,
		User:        claims.UserAttributes,
		Variables:   inst.ResolveVariables(false),
		ExtraProps: map[string]any{
			"args":   effectiveArgs,
			"params": effectiveArgs,
			"fields": canvas.FieldTemplateData(spec.Params, effectiveArgs, metricsViews),
		},
	}

	// Resolve templating in the renderer properties.
	// Numeric and boolean params are substituted before templating so they keep their type;
	// see canvas.CoerceScalarParams.
	var rendererProps *structpb.Struct
	if spec.RendererProperties != nil {
		props := canvas.CoerceScalarParams(spec.RendererProperties.AsMap(), spec.Params, effectiveArgs)

		v, err := parser.ResolveTemplateRecursively(props, td, false)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "failed to resolve renderer properties: %s", err.Error())
		}

		props, ok := v.(map[string]any)
		if !ok {
			return nil, status.Errorf(codes.Internal, "failed to convert resolved renderer properties to map: %v", v)
		}

		// For custom charts, inject scalar param values as native Vega-Lite params into the vega_spec.
		if spec.Renderer == "custom_chart" {
			if vegaSpec, ok := props["vega_spec"].(string); ok {
				injected, err := canvas.InjectVegaParams(vegaSpec, spec.Params, effectiveArgs)
				if err != nil {
					return nil, status.Errorf(codes.InvalidArgument, "failed to inject params into vega_spec: %s", err.Error())
				}
				props["vega_spec"] = injected
			}
		}

		rendererProps, err = structpb.NewStruct(props)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to convert renderer properties to struct: %s", err.Error())
		}
	}

	resolvedArgs, err := structpb.NewStruct(effectiveArgs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to convert resolved args to struct: %s", err.Error())
	}

	// Return the response
	return &runtimev1.ResolveComponentResponse{
		RendererProperties: rendererProps,
		ResolvedArgs:       resolvedArgs,
	}, nil
}

// boundMetricsViews loads the valid specs of the metrics views bound to a component's
// metrics_view params, with the caller's security policy applied.
// Metrics views that don't exist or that the caller cannot access are omitted.
func (s *Server) boundMetricsViews(ctx context.Context, instanceID string, claims *runtime.SecurityClaims, params []*runtimev1.ComponentParam, args map[string]any) (map[string]*runtimev1.MetricsViewSpec, error) {
	ctrl, err := s.runtime.Controller(ctx, instanceID)
	if err != nil {
		return nil, err
	}

	res := make(map[string]*runtimev1.MetricsViewSpec)
	for _, p := range params {
		if p.Type != "metrics_view" {
			continue
		}
		name, ok := args[p.Name].(string)
		if !ok || name == "" || res[name] != nil {
			continue
		}
		mvRes, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: name}, false)
		if err != nil {
			if errors.Is(err, drivers.ErrResourceNotFound) {
				continue
			}
			return nil, err
		}
		mvRes, access, err := s.runtime.ApplySecurityPolicy(ctx, instanceID, claims, mvRes)
		if err != nil {
			return nil, err
		}
		if !access {
			continue
		}
		if spec := mvRes.GetMetricsView().State.ValidSpec; spec != nil {
			res[name] = spec
		}
	}
	return res, nil
}
