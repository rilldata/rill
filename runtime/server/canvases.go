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

	res, err := s.runtime.ResolveCanvas(ctx, req.InstanceId, req.Canvas, claims, &runtime.ResolveCanvasOptions{
		Unsafe:        req.Unsafe,
		IncludeHidden: req.IncludeHidden,
	})
	if err != nil {
		return nil, err
	}

	return &runtimev1.ResolveCanvasResponse{
		Canvas:                 res.Canvas,
		ResolvedComponents:     res.ResolvedComponents,
		ReferencedMetricsViews: res.ReferencedMetricsViews,
		HiddenPaths:            res.HiddenPaths,
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

	// Find the component and check the user can access it
	res, canvasRes, err := s.accessibleComponent(ctx, req.InstanceId, req.Component, claims)
	if err != nil {
		return nil, err
	}

	spec := res.GetComponent().State.ValidSpec
	if spec == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "component %q is invalid", req.Component)
	}

	// Templates in the renderer properties may only access project variables if the component was declared in the project's source files.
	// A component of an admin-managed canvas (e.g. a personal canvas) was written by the canvas's creator, so it is not trusted.
	trusted := canvasRes == nil || !runtime.IsAdminManaged(canvasRes.GetCanvas().State.ValidSpec.Annotations)

	// Get current instance metadata
	inst, err := s.runtime.Instance(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}

	// Parse args and merge defaults before loading the metrics views referenced by the bindings.
	args := req.Args.AsMap()
	effectiveArgs := canvas.EffectiveArgs(spec, args)

	// Resolve metrics view metadata both to validate field bindings at this API boundary and to
	// expose display names and formatters to renderer-property templates.
	metricsViews, err := s.boundMetricsViews(ctx, req.InstanceId, claims, spec.Params, effectiveArgs)
	if err != nil {
		return nil, err
	}
	// Components without declared params accept arbitrary args for backwards compatibility.
	if len(spec.Params) > 0 {
		if err := canvas.ValidateParamBindings(spec.Params, args, metricsViews); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}

	variables, err := runtime.TemplateVariables(inst, claims, trusted, spec.RendererProperties.AsMap())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Setup templating data.
	// The effective args are exposed both as .params (canonical, matching the YAML key) and .args (legacy alias).
	td := parser.TemplateData{
		Environment: inst.Environment,
		User:        claims.UserAttributes,
		Variables:   variables,
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

		// A custom chart's vega_spec is a JSON document held in a string, so it is set aside and
		// resolved below with JSON-escaped template data; resolving it with the plain data would
		// splice a display name or string param containing a quote into the JSON unescaped.
		var vegaSpec string
		var hasVegaSpec bool
		if spec.Renderer == "custom_chart" {
			if vegaSpec, hasVegaSpec = props["vega_spec"].(string); hasVegaSpec {
				delete(props, "vega_spec")
			}
		}

		v, err := parser.ResolveTemplateRecursively(props, td, false)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "failed to resolve renderer properties: %s", err.Error())
		}

		props, ok := v.(map[string]any)
		if !ok {
			return nil, status.Errorf(codes.Internal, "failed to convert resolved renderer properties to map: %v", v)
		}

		// For custom charts, resolve the vega_spec and inject scalar param values into it as native Vega-Lite params.
		// The params are injected from the unescaped args; InjectVegaParams encodes them itself.
		if hasVegaSpec {
			vegaTD := td
			vegaTD.User = canvas.EscapeJSONStrings(td.User)
			vegaTD.Variables = canvas.EscapeJSONStrings(td.Variables)
			vegaTD.ExtraProps = canvas.EscapeJSONStrings(td.ExtraProps)
			resolved, err := parser.ResolveTemplate(vegaSpec, vegaTD, false)
			if err != nil {
				return nil, status.Errorf(codes.InvalidArgument, "failed to resolve vega_spec: %s", err.Error())
			}
			injected, err := canvas.InjectVegaParams(resolved, spec.Params, effectiveArgs)
			if err != nil {
				return nil, status.Errorf(codes.InvalidArgument, "failed to inject params into vega_spec: %s", err.Error())
			}
			props["vega_spec"] = injected
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

// accessibleComponent returns the component with the given name after checking that the claims may access it.
// Components have no access rules of their own (see the security engine's resolveRules),
// so a component defined inline in a canvas is only accessible if the claims can access that canvas and the component is not hidden from them on it.
// For such a component it also returns the canvas that defines it, so callers can tell whether the component was declared in the project's source files
// or in an admin-managed canvas (see runtime.IsAdminManaged). The canvas is nil for a component defined in its own file.
func (s *Server) accessibleComponent(ctx context.Context, instanceID, name string, claims *runtime.SecurityClaims) (*runtimev1.Resource, *runtimev1.Resource, error) {
	ctrl, err := s.runtime.Controller(ctx, instanceID)
	if err != nil {
		return nil, nil, err
	}
	res, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindComponent, Name: name}, false)
	if err != nil {
		if errors.Is(err, drivers.ErrResourceNotFound) {
			return nil, nil, status.Errorf(codes.NotFound, "component %q not found", name)
		}
		return nil, nil, err
	}

	// Check the component is not hidden from the claims by an `if` condition on its canvas
	res, access, err := s.runtime.ApplySecurityPolicy(ctx, instanceID, claims, res)
	if err != nil {
		return nil, nil, err
	}
	if !access {
		return nil, nil, status.Errorf(codes.PermissionDenied, "does not have access to component %q", name)
	}

	spec := res.GetComponent().State.ValidSpec
	if spec == nil {
		spec = res.GetComponent().Spec
	}
	if !spec.DefinedInCanvas {
		return res, nil, nil
	}

	// Find the canvas that defines the component.
	// Applying the security policy to a canvas denies it if the claims cannot access it and prunes the content hidden from them,
	// so the component is accessible only if an accessible canvas still contains it.
	canvases, err := ctrl.List(ctx, runtime.ResourceKindCanvas, "", false)
	if err != nil {
		return nil, nil, err
	}
	for _, c := range canvases {
		c, access, err := s.runtime.ApplySecurityPolicy(ctx, instanceID, claims, c)
		if err != nil {
			return nil, nil, err
		}
		if !access || c.GetCanvas().State.ValidSpec == nil {
			continue
		}
		names := make(map[string]bool)
		runtime.CollectCanvasComponentNames(c.GetCanvas().State.ValidSpec.Rows, names)
		if names[name] {
			return res, c, nil
		}
	}
	return nil, nil, status.Errorf(codes.PermissionDenied, "does not have access to component %q", name)
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
	for _, name := range canvas.BoundMetricsViewNames(params, args) {
		if res[name] != nil {
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
