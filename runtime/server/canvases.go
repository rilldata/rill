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
	res, origin, err := s.accessibleComponent(ctx, req.InstanceId, req.Component, claims)
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

	// Who wrote the renderer properties decides whether templates in them may access project variables (see runtime.PropertiesOrigin).
	variables, err := runtime.TemplateVariables(inst, claims, origin, spec.RendererProperties.AsMap())
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

// accessibleComponent returns the component with the given name after checking that the claims may access it,
// along with who wrote its properties (see runtime.PropertiesOrigin).
// Components have no access rules of their own (see the security engine's resolveRules).
// A component defined inline in a canvas is therefore accessible only if the claims can access a canvas that shows it,
// usually the canvas that defines it but possibly another canvas that references it by name (e.g. through a shared link that grants only that canvas),
// and if they meet the `if` conditions the component inherits from its item, which hide it wherever it's referenced (matching ResolveCanvas).
// The canvas that defines the component decides the origin: a personal canvas is written by its creator, not declared in the project's source files.
// A component defined in its own file is declared in the project.
func (s *Server) accessibleComponent(ctx context.Context, instanceID, name string, claims *runtime.SecurityClaims) (*runtimev1.Resource, runtime.PropertiesOrigin, error) {
	ctrl, err := s.runtime.Controller(ctx, instanceID)
	if err != nil {
		return nil, 0, err
	}
	res, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindComponent, Name: name}, false)
	if err != nil {
		if errors.Is(err, drivers.ErrResourceNotFound) {
			return nil, 0, status.Errorf(codes.NotFound, "component %q not found", name)
		}
		return nil, 0, err
	}
	denied := status.Errorf(codes.PermissionDenied, "does not have access to component %q", name)

	spec := res.GetComponent().State.ValidSpec
	if spec == nil {
		spec = res.GetComponent().Spec
	}
	if !spec.DefinedInCanvas {
		res, access, err := s.runtime.ApplySecurityPolicy(ctx, instanceID, claims, res)
		if err != nil {
			return nil, 0, err
		}
		if !access {
			return nil, 0, denied
		}
		return res, runtime.PropertiesFromProject, nil
	}

	definingCanvas, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindCanvas, Name: spec.ParentCanvas}, false)
	if err != nil {
		if errors.Is(err, drivers.ErrResourceNotFound) {
			return nil, 0, denied
		}
		return nil, 0, err
	}

	// Check the defining canvas first, since it almost always is the canvas the component is accessed through.
	shown, err := s.canvasShowsComponent(ctx, instanceID, claims, definingCanvas, name)
	if err != nil {
		return nil, 0, err
	}
	if !shown {
		canvases, err := ctrl.List(ctx, runtime.ResourceKindCanvas, "", false)
		if err != nil {
			return nil, 0, err
		}
		for _, c := range canvases {
			if c.Meta.Name.Name == spec.ParentCanvas {
				continue
			}
			shown, err = s.canvasShowsComponent(ctx, instanceID, claims, c, name)
			if err != nil {
				return nil, 0, err
			}
			if shown {
				break
			}
		}
	}
	if !shown {
		return nil, 0, denied
	}

	visibility, err := s.runtime.VisibilityConditions(ctx, instanceID, claims, res)
	if err != nil {
		return nil, 0, err
	}
	if !visibility.ComponentMet(res) {
		return nil, 0, denied
	}

	canvasSpec := definingCanvas.GetCanvas().State.ValidSpec
	if canvasSpec == nil {
		canvasSpec = definingCanvas.GetCanvas().Spec
	}
	return res, runtime.PropertiesOriginFromAnnotations(canvasSpec.Annotations), nil
}

// canvasShowsComponent reports whether the claims can access the canvas and the canvas shows the component to them.
// Applying the security policy prunes the content hidden from the claims by `if` conditions, so a hidden item does not count.
func (s *Server) canvasShowsComponent(ctx context.Context, instanceID string, claims *runtime.SecurityClaims, canvasRes *runtimev1.Resource, component string) (bool, error) {
	canvasRes, access, err := s.runtime.ApplySecurityPolicy(ctx, instanceID, claims, canvasRes)
	if err != nil {
		return false, err
	}
	if !access || canvasRes.GetCanvas().State.ValidSpec == nil {
		return false, nil
	}
	names := make(map[string]bool)
	runtime.CollectCanvasComponentNames(canvasRes.GetCanvas().State.ValidSpec.Rows, names)
	return names[component], nil
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
