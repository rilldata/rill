package server

import (
	"context"
	"errors"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/compilers/rillv1"
	"github.com/rilldata/rill/runtime/drivers"
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
	td := rillv1.TemplateData{
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
		v, err := rillv1.ResolveTemplateRecursively(spec.RendererProperties.AsMap(), td)
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
		Show:               true,
		RendererProperties: rendererProps,
	}, nil
}
