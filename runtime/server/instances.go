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

	if !auth.GetClaims(ctx).Can(auth.ManageInstances) {
		return nil, ErrForbidden
	}

	inst := &drivers.Instance{
		ID:                  req.InstanceId,
		OLAPDriver:          req.OlapDriver,
		OLAPDSN:             req.OlapDsn,
		RepoDriver:          req.RepoDriver,
		RepoDSN:             req.RepoDsn,
		EmbedCatalog:        req.EmbedCatalog,
		Variables:           req.Variables,
		IngestionLimitBytes: req.IngestionLimitBytes,
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
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.olap_driver", req.GetOlapDriver()),
		attribute.String("args.repo_driver", req.GetRepoDriver()),
	)

	if !auth.GetClaims(ctx).Can(auth.ManageInstances) {
		return nil, ErrForbidden
	}

	olderInst, err := s.runtime.FindInstance(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}

	inst := &drivers.Instance{
		ID:                  req.InstanceId,
		OLAPDriver:          valOrDefault(req.OlapDriver, olderInst.OLAPDriver),
		OLAPDSN:             valOrDefault(req.OlapDsn, olderInst.OLAPDSN),
		RepoDriver:          valOrDefault(req.RepoDriver, olderInst.RepoDriver),
		RepoDSN:             valOrDefault(req.RepoDsn, olderInst.RepoDSN),
		EmbedCatalog:        valOrDefault(req.EmbedCatalog, olderInst.EmbedCatalog),
		Variables:           req.Variables,
		IngestionLimitBytes: valOrDefault(req.IngestionLimitBytes, olderInst.IngestionLimitBytes),
	}

	err = s.runtime.EditInstance(ctx, inst)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &runtimev1.EditInstanceResponse{
		Instance: instanceToPB(inst),
	}, nil
}

// DeleteInstance implements RuntimeService.
func (s *Server) DeleteInstance(ctx context.Context, req *runtimev1.DeleteInstanceRequest) (*runtimev1.DeleteInstanceResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.Bool("args.drop_db", req.DropDb),
	)

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
		OlapDsn:             inst.OLAPDSN,
		RepoDriver:          inst.RepoDriver,
		RepoDsn:             inst.RepoDSN,
		EmbedCatalog:        inst.EmbedCatalog,
		Variables:           inst.Variables,
		ProjectVariables:    inst.ProjectVariables,
		IngestionLimitBytes: inst.IngestionLimitBytes,
	}
}

func valOrDefault[T any](ptr *T, def T) T {
	if ptr != nil {
		return *ptr
	}
	return def
}
