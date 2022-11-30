package server

import (
	"context"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ListInstances implements RuntimeService
func (s *Server) ListInstances(ctx context.Context, req *runtimev1.ListInstancesRequest) (*runtimev1.ListInstancesResponse, error) {
	registry, _ := s.metastore.RegistryStore()
	instances := registry.FindInstances(ctx)

	pbs := make([]*runtimev1.Instance, len(instances))
	for i, inst := range instances {
		pbs[i] = instanceToPB(inst)
	}

	return &runtimev1.ListInstancesResponse{Instances: pbs}, nil
}

// GetInstance implements RuntimeService
func (s *Server) GetInstance(ctx context.Context, req *runtimev1.GetInstanceRequest) (*runtimev1.GetInstanceResponse, error) {
	registry, _ := s.metastore.RegistryStore()
	inst, found := registry.FindInstance(ctx, req.InstanceId)
	if !found {
		return nil, status.Error(codes.NotFound, "instance not found")
	}

	return &runtimev1.GetInstanceResponse{
		Instance: instanceToPB(inst),
	}, nil
}

// CreateInstance implements RuntimeService
func (s *Server) CreateInstance(ctx context.Context, req *runtimev1.CreateInstanceRequest) (*runtimev1.CreateInstanceResponse, error) {
	inst := &drivers.Instance{
		ID:           req.InstanceId,
		OLAPDriver:   req.OlapDriver,
		OLAPDSN:      req.OlapDsn,
		RepoDriver:   req.RepoDriver,
		RepoDSN:      req.RepoDsn,
		EmbedCatalog: req.EmbedCatalog,
	}

	// Check OLAP connection
	olap, err := drivers.Open(inst.OLAPDriver, inst.OLAPDSN)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "could not connect to driver '%s': %s", inst.OLAPDriver, err.Error())
	}
	_, ok := olap.OLAPStore()
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "not a valid OLAP driver")
	}

	// Check repo connection
	repo, err := drivers.Open(inst.RepoDriver, inst.RepoDSN)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	_, ok = repo.RepoStore()
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "not a valid repo driver")
	}

	// Check that it's a driver that supports embedded catalogs
	if inst.EmbedCatalog {
		_, ok := olap.CatalogStore()
		if !ok {
			return nil, status.Error(codes.InvalidArgument, "driver does not support embedded catalogs")
		}
	}

	// Prepare connections for use
	err = olap.Migrate(ctx)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to prepare instance: %s", err.Error())
	}
	err = repo.Migrate(ctx)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to prepare instance: %s", err.Error())
	}

	registry, _ := s.metastore.RegistryStore()
	err = registry.CreateInstance(ctx, inst)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &runtimev1.CreateInstanceResponse{
		Instance: instanceToPB(inst),
	}, nil
}

// DeleteInstance implements RuntimeService
func (s *Server) DeleteInstance(ctx context.Context, req *runtimev1.DeleteInstanceRequest) (*runtimev1.DeleteInstanceResponse, error) {
	registry, _ := s.metastore.RegistryStore()
	err := registry.DeleteInstance(ctx, req.InstanceId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &runtimev1.DeleteInstanceResponse{}, nil
}

func instanceToPB(inst *drivers.Instance) *runtimev1.Instance {
	return &runtimev1.Instance{
		InstanceId:   inst.ID,
		OlapDriver:   inst.OLAPDriver,
		OlapDsn:      inst.OLAPDSN,
		RepoDriver:   inst.RepoDriver,
		RepoDsn:      inst.RepoDSN,
		EmbedCatalog: inst.EmbedCatalog,
	}
}
