package server

import (
	"context"
	"errors"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CreateResource implements runtimev1.RuntimeServiceServer
func (s *Server) CreateResource(ctx context.Context, req *runtimev1.CreateResourceRequest) (*runtimev1.CreateResourceResponse, error) {
	s.addInstanceRequestAttributes(ctx, req.InstanceId)
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.resource.meta.name.kind", req.Resource.Meta.Name.Kind),
		attribute.String("args.resource.meta.name.name", req.Resource.Meta.Name.Name),
	)

	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.EditRepo) {
		return nil, ErrForbidden
	}

	ctrl, err := s.runtime.Controller(ctx, req.InstanceId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Extract resource details
	name := req.Resource.Meta.Name
	refs := req.Resource.Meta.Refs
	var owner *runtimev1.ResourceName
	if req.Resource.Meta.Owner != nil {
		owner = req.Resource.Meta.Owner
	}
	paths := req.Resource.Meta.FilePaths
	hidden := req.Resource.Meta.Hidden

	err = ctrl.Create(ctx, name, refs, owner, paths, hidden, req.Resource)
	if err != nil {
		if errors.Is(err, drivers.ErrResourceAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, "resource already exists")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Get the created resource to return
	r, err := ctrl.Get(ctx, name, false)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &runtimev1.CreateResourceResponse{Resource: r}, nil
}

// UpdateResource implements runtimev1.RuntimeServiceServer
func (s *Server) UpdateResource(ctx context.Context, req *runtimev1.UpdateResourceRequest) (*runtimev1.UpdateResourceResponse, error) {
	s.addInstanceRequestAttributes(ctx, req.InstanceId)
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.kind", req.Kind),
		attribute.String("args.name", req.Name),
	)

	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.EditRepo) {
		return nil, ErrForbidden
	}

	ctrl, err := s.runtime.Controller(ctx, req.InstanceId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	name := &runtimev1.ResourceName{Kind: req.Kind, Name: req.Name}

	err = ctrl.UpdateSpec(ctx, name, req.Resource)
	if err != nil {
		if errors.Is(err, drivers.ErrResourceNotFound) {
			return nil, status.Error(codes.NotFound, "resource not found")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Get the updated resource to return
	r, err := ctrl.Get(ctx, name, false)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &runtimev1.UpdateResourceResponse{Resource: r}, nil
}

// DeleteResource implements runtimev1.RuntimeServiceServer
func (s *Server) DeleteResource(ctx context.Context, req *runtimev1.DeleteResourceRequest) (*runtimev1.DeleteResourceResponse, error) {
	s.addInstanceRequestAttributes(ctx, req.InstanceId)
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.kind", req.Kind),
		attribute.String("args.name", req.Name),
	)

	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.EditRepo) {
		return nil, ErrForbidden
	}

	ctrl, err := s.runtime.Controller(ctx, req.InstanceId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	name := &runtimev1.ResourceName{Kind: req.Kind, Name: req.Name}

	err = ctrl.Delete(ctx, name)
	if err != nil {
		if errors.Is(err, drivers.ErrResourceNotFound) {
			return nil, status.Error(codes.NotFound, "resource not found")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &runtimev1.DeleteResourceResponse{}, nil
}
