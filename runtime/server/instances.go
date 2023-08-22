package server

import (
	"context"
	"errors"

	"github.com/bufbuild/connect-go"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ListInstances implements RuntimeService.
func (s *Server) ListInstances(ctx context.Context, req *connect.Request[runtimev1.ListInstancesRequest]) (*connect.Response[runtimev1.ListInstancesResponse], error) {
	if !auth.GetClaims(ctx).Can(auth.ManageInstances) {
		return nil, ErrForbidden
	}

	instances, err := s.runtime.FindInstances(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	pbs := make([]*runtimev1.Instance, len(instances))
	for i, inst := range instances {
		pbs[i] = instanceToPB(inst)
	}

	return connect.NewResponse(&runtimev1.ListInstancesResponse{Instances: pbs}), nil
}

// GetInstance implements RuntimeService.
func (s *Server) GetInstance(ctx context.Context, req *connect.Request[runtimev1.GetInstanceRequest]) (*connect.Response[runtimev1.GetInstanceResponse], error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.Msg.InstanceId),
	)

	s.addInstanceRequestAttributes(ctx, req.Msg.InstanceId)

	if !auth.GetClaims(ctx).CanInstance(req.Msg.InstanceId, auth.ReadInstance) {
		return nil, ErrForbidden
	}

	inst, err := s.runtime.FindInstance(ctx, req.Msg.InstanceId)
	if err != nil {
		if errors.Is(err, drivers.ErrNotFound) {
			return nil, status.Error(codes.InvalidArgument, "instance not found")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return connect.NewResponse(&runtimev1.GetInstanceResponse{
		Instance: instanceToPB(inst),
	}), nil
}

// CreateInstance implements RuntimeService.
func (s *Server) CreateInstance(ctx context.Context, req *connect.Request[runtimev1.CreateInstanceRequest]) (*connect.Response[runtimev1.CreateInstanceResponse], error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.Msg.InstanceId),
		attribute.String("args.olap_driver", req.Msg.OlapDriver),
		attribute.String("args.repo_driver", req.Msg.RepoDriver),
	)

	s.addInstanceRequestAttributes(ctx, req.Msg.InstanceId)

	if !auth.GetClaims(ctx).Can(auth.ManageInstances) {
		return nil, ErrForbidden
	}

	inst := &drivers.Instance{
		ID:                  req.Msg.InstanceId,
		OLAPDriver:          req.Msg.OlapDriver,
		RepoDriver:          req.Msg.RepoDriver,
		EmbedCatalog:        req.Msg.EmbedCatalog,
		Variables:           req.Msg.Variables,
		IngestionLimitBytes: req.Msg.IngestionLimitBytes,
		Annotations:         req.Msg.Annotations,
		Connectors:          req.Msg.Connectors,
	}

	err := s.runtime.CreateInstance(ctx, inst)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return connect.NewResponse(&runtimev1.CreateInstanceResponse{
		Instance: instanceToPB(inst),
	}), nil
}

// EditInstance implements RuntimeService.
func (s *Server) EditInstance(ctx context.Context, req *connect.Request[runtimev1.EditInstanceRequest]) (*connect.Response[runtimev1.EditInstanceResponse], error) {
	observability.AddRequestAttributes(ctx, attribute.String("args.instance_id", req.Msg.InstanceId))
	if req.Msg.OlapDriver != nil {
		observability.AddRequestAttributes(ctx, attribute.String("args.olap_driver", *req.Msg.OlapDriver))
	}
	if req.Msg.RepoDriver != nil {
		observability.AddRequestAttributes(ctx, attribute.String("args.repo_driver", *req.Msg.RepoDriver))
	}

	s.addInstanceRequestAttributes(ctx, req.Msg.InstanceId)

	if !auth.GetClaims(ctx).Can(auth.ManageInstances) {
		return nil, ErrForbidden
	}

	oldInst, err := s.runtime.FindInstance(ctx, req.Msg.InstanceId)
	if err != nil {
		return nil, err
	}

	annotations := req.Msg.Annotations
	if len(annotations) == 0 { // annotations not changed
		annotations = oldInst.Annotations
	}

	inst := &drivers.Instance{
		ID:                  req.Msg.InstanceId,
		OLAPDriver:          valOrDefault(req.Msg.OlapDriver, oldInst.OLAPDriver),
		RepoDriver:          valOrDefault(req.Msg.RepoDriver, oldInst.RepoDriver),
		EmbedCatalog:        valOrDefault(req.Msg.EmbedCatalog, oldInst.EmbedCatalog),
		Variables:           oldInst.Variables,
		IngestionLimitBytes: valOrDefault(req.Msg.IngestionLimitBytes, oldInst.IngestionLimitBytes),
		Annotations:         annotations,
	}
	if len(req.Msg.Connectors) == 0 {
		inst.Connectors = oldInst.Connectors
	} else {
		inst.Connectors = req.Msg.Connectors
	}

	err = s.runtime.EditInstance(ctx, inst)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return connect.NewResponse(&runtimev1.EditInstanceResponse{
		Instance: instanceToPB(inst),
	}), nil
}

// EditInstanceVariables implements RuntimeService.
func (s *Server) EditInstanceVariables(ctx context.Context, req *connect.Request[runtimev1.EditInstanceVariablesRequest]) (*connect.Response[runtimev1.EditInstanceVariablesResponse], error) {
	observability.AddRequestAttributes(ctx, attribute.String("args.instance_id", req.Msg.InstanceId))

	s.addInstanceRequestAttributes(ctx, req.Msg.InstanceId)

	if !auth.GetClaims(ctx).Can(auth.ManageInstances) {
		return nil, ErrForbidden
	}

	oldInst, err := s.runtime.FindInstance(ctx, req.Msg.InstanceId)
	if err != nil {
		return nil, err
	}

	inst := &drivers.Instance{
		ID:                  req.Msg.InstanceId,
		OLAPDriver:          oldInst.OLAPDriver,
		RepoDriver:          oldInst.RepoDriver,
		EmbedCatalog:        oldInst.EmbedCatalog,
		IngestionLimitBytes: oldInst.IngestionLimitBytes,
		Variables:           req.Msg.Variables,
		Annotations:         oldInst.Annotations,
		Connectors:          oldInst.Connectors,
	}

	err = s.runtime.EditInstance(ctx, inst)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return connect.NewResponse(&runtimev1.EditInstanceVariablesResponse{
		Instance: instanceToPB(inst),
	}), nil
}

// EditInstanceAnnotations implements RuntimeService.
func (s *Server) EditInstanceAnnotations(ctx context.Context, req *runtimev1.EditInstanceAnnotationsRequest) (*runtimev1.EditInstanceAnnotationsResponse, error) {
	observability.AddRequestAttributes(ctx, attribute.String("args.instance_id", req.InstanceId))

	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	if !auth.GetClaims(ctx).Can(auth.ManageInstances) {
		return nil, ErrForbidden
	}

	oldInst, err := s.runtime.FindInstance(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}

	inst := &drivers.Instance{
		ID:                  req.InstanceId,
		OLAPDriver:          oldInst.OLAPDriver,
		RepoDriver:          oldInst.RepoDriver,
		EmbedCatalog:        oldInst.EmbedCatalog,
		IngestionLimitBytes: oldInst.IngestionLimitBytes,
		Variables:           oldInst.Variables,
		Annotations:         req.Annotations,
	}

	err = s.runtime.EditInstance(ctx, inst)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &runtimev1.EditInstanceAnnotationsResponse{
		Instance: instanceToPB(inst),
	}, nil
}

// DeleteInstance implements RuntimeService.
func (s *Server) DeleteInstance(ctx context.Context, req *connect.Request[runtimev1.DeleteInstanceRequest]) (*connect.Response[runtimev1.DeleteInstanceResponse], error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.Msg.InstanceId),
		attribute.Bool("args.drop_db", req.Msg.DropDb),
	)

	s.addInstanceRequestAttributes(ctx, req.Msg.InstanceId)

	if !auth.GetClaims(ctx).Can(auth.ManageInstances) {
		return nil, ErrForbidden
	}

	err := s.runtime.DeleteInstance(ctx, req.Msg.InstanceId, req.Msg.DropDb)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return connect.NewResponse(&runtimev1.DeleteInstanceResponse{}), nil
}

func instanceToPB(inst *drivers.Instance) *runtimev1.Instance {
	return &runtimev1.Instance{
		InstanceId:          inst.ID,
		OlapDriver:          inst.OLAPDriver,
		RepoDriver:          inst.RepoDriver,
		EmbedCatalog:        inst.EmbedCatalog,
		Variables:           inst.Variables,
		ProjectVariables:    inst.ProjectVariables,
		IngestionLimitBytes: inst.IngestionLimitBytes,
		Connectors:          inst.Connectors,
	}
}

func valOrDefault[T any](ptr *T, def T) T {
	if ptr != nil {
		return *ptr
	}
	return def
}
