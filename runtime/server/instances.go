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

// ListInstances implements RuntimeService.
func (s *Server) ListInstances(ctx context.Context, req *runtimev1.ListInstancesRequest) (*runtimev1.ListInstancesResponse, error) {
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

	return &runtimev1.ListInstancesResponse{Instances: pbs}, nil
}

// GetInstance implements RuntimeService.
func (s *Server) GetInstance(ctx context.Context, req *runtimev1.GetInstanceRequest) (*runtimev1.GetInstanceResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
	)

	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.ReadInstance) {
		return nil, ErrForbidden
	}

	inst, err := s.runtime.FindInstance(ctx, req.InstanceId)
	if err != nil {
		if errors.Is(err, drivers.ErrNotFound) {
			return nil, status.Error(codes.InvalidArgument, "instance not found")
		}
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &runtimev1.GetInstanceResponse{
		Instance: instanceToPB(inst),
	}, nil
}

// CreateInstance implements RuntimeService.
func (s *Server) CreateInstance(ctx context.Context, req *runtimev1.CreateInstanceRequest) (*runtimev1.CreateInstanceResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.olap_driver", req.OlapDriver),
		attribute.String("args.repo_driver", req.RepoDriver),
	)

	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	if !auth.GetClaims(ctx).Can(auth.ManageInstances) {
		return nil, ErrForbidden
	}

	inst := &drivers.Instance{
		ID:                  req.InstanceId,
		OLAPDriver:          req.OlapDriver,
		RepoDriver:          req.RepoDriver,
		EmbedCatalog:        req.EmbedCatalog,
		Variables:           req.Variables,
		IngestionLimitBytes: req.IngestionLimitBytes,
		Annotations:         req.Annotations,
		Connectors:          req.Connectors,
	}

	err := s.runtime.CreateInstance(ctx, inst)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &runtimev1.CreateInstanceResponse{
		Instance: instanceToPB(inst),
	}, nil
}

// EditInstance implements RuntimeService.
func (s *Server) EditInstance(ctx context.Context, req *runtimev1.EditInstanceRequest) (*runtimev1.EditInstanceResponse, error) {
	observability.AddRequestAttributes(ctx, attribute.String("args.instance_id", req.InstanceId))
	if req.OlapDriver != nil {
		observability.AddRequestAttributes(ctx, attribute.String("args.olap_driver", *req.OlapDriver))
	}
	if req.RepoDriver != nil {
		observability.AddRequestAttributes(ctx, attribute.String("args.repo_driver", *req.RepoDriver))
	}

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
		OLAPDriver:          valOrDefault(req.OlapDriver, oldInst.OLAPDriver),
		RepoDriver:          valOrDefault(req.RepoDriver, oldInst.RepoDriver),
		EmbedCatalog:        valOrDefault(req.EmbedCatalog, oldInst.EmbedCatalog),
		Variables:           oldInst.Variables,
		IngestionLimitBytes: valOrDefault(req.IngestionLimitBytes, oldInst.IngestionLimitBytes),
		Annotations:         oldInst.Annotations,
	}
	if len(req.Connectors) == 0 {
		inst.Connectors = oldInst.Connectors
	} else {
		inst.Connectors = req.Connectors
	}

	err = s.runtime.EditInstance(ctx, inst)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &runtimev1.EditInstanceResponse{
		Instance: instanceToPB(inst),
	}, nil
}

// EditInstanceVariables implements RuntimeService.
func (s *Server) EditInstanceVariables(ctx context.Context, req *runtimev1.EditInstanceVariablesRequest) (*runtimev1.EditInstanceVariablesResponse, error) {
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
		Variables:           req.Variables,
		Annotations:         oldInst.Annotations,
		Connectors:          oldInst.Connectors,
	}

	err = s.runtime.EditInstance(ctx, inst)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &runtimev1.EditInstanceVariablesResponse{
		Instance: instanceToPB(inst),
	}, nil
}

// DeleteInstance implements RuntimeService.
func (s *Server) DeleteInstance(ctx context.Context, req *runtimev1.DeleteInstanceRequest) (*runtimev1.DeleteInstanceResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.Bool("args.drop_db", req.DropDb),
	)

	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	if !auth.GetClaims(ctx).Can(auth.ManageInstances) {
		return nil, ErrForbidden
	}

	err := s.runtime.DeleteInstance(ctx, req.InstanceId, req.DropDb)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &runtimev1.DeleteInstanceResponse{}, nil
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
