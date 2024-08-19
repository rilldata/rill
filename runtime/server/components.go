package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/compilers/rillv1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/httputil"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

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

	// Find the component resource
	spec, err := s.getComponent(ctx, req.InstanceId, req.Component)
	if err != nil {
		if errors.Is(err, drivers.ErrResourceNotFound) {
			return nil, status.Errorf(codes.NotFound, "component with name %q not found", req.Component)
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Get current instance metadata
	inst, err := s.runtime.Instance(ctx, req.InstanceId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Parse args
	args := req.Args.AsMap()

	// Resolve templating in the renderer properties
	var rendererProps *structpb.Struct
	if spec.RendererProperties != nil {
		v, err := rillv1.ResolveTemplateRecursively(spec.RendererProperties.AsMap(), rillv1.TemplateData{
			Environment: inst.Environment,
			User:        auth.GetClaims(ctx).SecurityClaims().UserAttributes,
			Variables:   inst.ResolveVariables(),
			ExtraProps: map[string]any{
				"args": args,
			},
		})
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

	// Call the component's data resolver
	var schema *runtimev1.StructType
	var data []*structpb.Struct
	if spec.Resolver != "" {
		res, err := s.runtime.Resolve(ctx, &runtime.ResolveOptions{
			InstanceID:         req.InstanceId,
			Resolver:           spec.Resolver,
			ResolverProperties: spec.ResolverProperties.AsMap(),
			Args:               req.Args.AsMap(),
			Claims:             auth.GetClaims(ctx).SecurityClaims(),
		})
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		defer res.Close()

		schema = res.Schema()
		for {
			row, err := res.Next()
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				return nil, status.Error(codes.Internal, err.Error())
			}

			pb, err := structpb.NewStruct(row)
			if err != nil {
				return nil, status.Errorf(codes.Internal, "failed to convert row to struct: %s", err.Error())
			}

			data = append(data, pb)
		}
	}

	// Return the response
	return &runtimev1.ResolveComponentResponse{
		Schema:             schema,
		Data:               data,
		RendererProperties: rendererProps,
	}, nil
}

// componentDataHandler handles requests to resolve a component's data.
// Deprecated: Use ResolveComponent instead.
func (s *Server) componentDataHandler(w http.ResponseWriter, req *http.Request) error {
	// Parse path parameters
	ctx := req.Context()
	instanceID := req.PathValue("instance_id")
	component := req.PathValue("name")

	// Check if user has access to query for component data (we use the ReadAPI permission for this for now)
	if !auth.GetClaims(req.Context()).CanInstance(instanceID, auth.ReadAPI) {
		return httputil.Errorf(http.StatusForbidden, "does not have access to component data")
	}

	// Parse args from the request body and URL query
	args := make(map[string]any)
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return httputil.Errorf(http.StatusBadRequest, "failed to read request body: %w", err)
	}
	if len(body) > 0 { // For POST requests
		if err := json.Unmarshal(body, &args); err != nil {
			return httputil.Errorf(http.StatusBadRequest, "failed to unmarshal request body: %w", err)
		}
	}
	for k, v := range req.URL.Query() {
		// Set only the first value so that client does need to put array accessors in templates.
		args[k] = v[0]
	}

	// Find the component resource
	componentSpec, err := s.getComponent(ctx, instanceID, component)
	if err != nil {
		if errors.Is(err, drivers.ErrResourceNotFound) {
			return httputil.Errorf(http.StatusNotFound, "component with name %q not found", component)
		}
		return httputil.Error(http.StatusInternalServerError, err)
	}

	// Call the component's data resolver
	res, err := s.runtime.Resolve(ctx, &runtime.ResolveOptions{
		InstanceID:         instanceID,
		Resolver:           componentSpec.Resolver,
		ResolverProperties: componentSpec.ResolverProperties.AsMap(),
		Args:               args,
		Claims:             auth.GetClaims(ctx).SecurityClaims(),
	})
	if err != nil {
		return httputil.Error(http.StatusBadRequest, err)
	}
	defer res.Close()

	// Write the response
	data, err := res.MarshalJSON()
	if err != nil {
		return httputil.Error(http.StatusInternalServerError, err)
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	if err != nil {
		return httputil.Error(http.StatusInternalServerError, err)
	}

	return nil
}

func (s *Server) getComponent(ctx context.Context, instanceID, name string) (*runtimev1.ComponentSpec, error) {
	ctrl, err := s.runtime.Controller(ctx, instanceID)
	if err != nil {
		return nil, err
	}

	res, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindComponent, Name: name}, false)
	if err != nil {
		return nil, err
	}

	ch := res.GetComponent()
	spec := ch.Spec
	if spec == nil {
		return nil, fmt.Errorf("component %q is invalid", name)
	}

	return spec, nil
}
