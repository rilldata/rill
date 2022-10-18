package server

import (
	"context"

	"github.com/rilldata/rill/runtime/api"
	"github.com/rilldata/rill/runtime/drivers"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ListInstances implements RuntimeService
func (s *Server) ListInstances(ctx context.Context, req *api.ListInstancesRequest) (*api.ListInstancesResponse, error) {
	registry, _ := s.metastore.RegistryStore()
	instances := registry.FindInstances(ctx)

	pbs := make([]*api.Instance, len(instances))
	for i, inst := range instances {
		pbs[i] = instanceToPB(inst)
	}

	return &api.ListInstancesResponse{Instances: pbs}, nil
}

// GetInstance implements RuntimeService
func (s *Server) GetInstance(ctx context.Context, req *api.GetInstanceRequest) (*api.GetInstanceResponse, error) {
	registry, _ := s.metastore.RegistryStore()
	inst, found := registry.FindInstance(ctx, req.InstanceId)
	if !found {
		return nil, status.Error(codes.NotFound, "instance not found")
	}

	return &api.GetInstanceResponse{
		Instance: instanceToPB(inst),
	}, nil
}

// CreateInstance implements RuntimeService
func (s *Server) CreateInstance(ctx context.Context, req *api.CreateInstanceRequest) (*api.CreateInstanceResponse, error) {
	inst := &drivers.Instance{
		Driver:       req.Driver,
		DSN:          req.Dsn,
		ObjectPrefix: req.ObjectPrefix,
		Exposed:      req.Exposed,
		EmbedCatalog: req.EmbedCatalog,
	}

	// Check that it's a valid driver for OLAP
	conn, err := drivers.Open(inst.Driver, inst.DSN)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	_, ok := conn.OLAPStore()
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "not a valid OLAP driver")
	}

	// Check that it's a driver that supports embedded catalogs
	if inst.EmbedCatalog {
		_, ok := conn.CatalogStore()
		if !ok {
			return nil, status.Error(codes.InvalidArgument, "driver does not support embedded catalogs")
		}
	}

	registry, _ := s.metastore.RegistryStore()
	err = registry.CreateInstance(ctx, inst)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &api.CreateInstanceResponse{
		InstanceId: inst.ID,
		Instance:   instanceToPB(inst),
	}, nil
}

func CreateLocalInstance(s *Server, driver string, dsn string) error {
	inst := &drivers.Instance{
		ID:           "default",
		Driver:       driver,
		DSN:          dsn,
		Exposed:      true,
		EmbedCatalog: true,
	}

	// Check that it's a valid driver for OLAP
	conn, err := drivers.Open(inst.Driver, inst.DSN)
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	_, ok := conn.OLAPStore()
	if !ok {
		return status.Error(codes.InvalidArgument, "not a valid OLAP driver")
	}

	// Check that it's a driver that supports embedded catalogs
	if inst.EmbedCatalog {
		_, ok := conn.CatalogStore()
		if !ok {
			return status.Error(codes.InvalidArgument, "driver does not support embedded catalogs")
		}
	}

	registry, _ := s.metastore.RegistryStore()
	err = registry.CreateInstance(context.Background(), inst)
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	return nil
}

// DeleteInstance implements RuntimeService
func (s *Server) DeleteInstance(ctx context.Context, req *api.DeleteInstanceRequest) (*api.DeleteInstanceResponse, error) {
	registry, _ := s.metastore.RegistryStore()
	err := registry.DeleteInstance(ctx, req.InstanceId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &api.DeleteInstanceResponse{}, nil
}

func instanceToPB(inst *drivers.Instance) *api.Instance {
	return &api.Instance{
		InstanceId:   inst.ID,
		Driver:       inst.Driver,
		Dsn:          inst.DSN,
		ObjectPrefix: inst.ObjectPrefix,
		Exposed:      inst.Exposed,
		EmbedCatalog: inst.EmbedCatalog,
	}
}
