package server

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/runtime/api"
	"github.com/rilldata/rill/runtime/drivers"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Migrate implements RuntimeService
func (s *Server) Migrate(ctx context.Context, req *api.MigrateRequest) (*api.MigrateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method not implemented")
}

// MigrateSingle implements RuntimeService
// NOTE: This is an initial migration implementation with several flaws.
func (s *Server) MigrateSingle(ctx context.Context, req *api.MigrateSingleRequest) (*api.MigrateSingleResponse, error) {
	// Parse SQL
	source, err := sqlToSource(req.Sql)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Get instance
	registry, _ := s.metastore.RegistryStore()
	inst, found := registry.FindInstance(ctx, req.InstanceId)
	if !found {
		return nil, status.Error(codes.InvalidArgument, "instance not found")
	}

	// Get catalog
	catalog, err := s.openCatalog(ctx, inst)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Check if object exists and is not a source
	obj, found := catalog.FindObject(ctx, req.InstanceId, source.Name)
	if found && obj.Type != drivers.CatalogObjectTypeSource {
		return nil, status.Errorf(codes.FailedPrecondition, "an object of type '%s' already exists with name '%s'", obj.Type, obj.Name)
	}

	// Get olap
	conn, err := drivers.Open(inst.Driver, inst.DSN)
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}
	olap, _ := conn.OLAPStore()

	// Ingest the source
	err = olap.Ingest(ctx, source)
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}

	// Save definition
	obj = &drivers.CatalogObject{
		Name: source.Name,
		Type: drivers.CatalogObjectTypeSource,
		SQL:  req.Sql,
	}
	if found {
		err = catalog.UpdateObject(ctx, req.InstanceId, obj)
	} else {
		err = catalog.CreateObject(ctx, req.InstanceId, obj)
	}
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "error: could not insert source into catalog: %s (warning: watch out for corruptions)", err.Error())
	}

	return &api.MigrateSingleResponse{}, nil
}

// MigrateDelete implements RuntimeService
func (s *Server) MigrateDelete(ctx context.Context, req *api.MigrateDeleteRequest) (*api.MigrateDeleteResponse, error) {
	// Get instance
	registry, _ := s.metastore.RegistryStore()
	inst, found := registry.FindInstance(ctx, req.InstanceId)
	if !found {
		return nil, status.Error(codes.InvalidArgument, "instance not found")
	}

	// Get catalog
	catalog, err := s.openCatalog(ctx, inst)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Get object
	obj, found := catalog.FindObject(ctx, req.InstanceId, req.Name)
	if !found {
		return nil, status.Errorf(codes.InvalidArgument, "object not found")
	}

	// Delete from underlying if applicable
	switch obj.Type {
	case drivers.CatalogObjectTypeSource:
		// Get OLAP
		conn, err := drivers.Open(inst.Driver, inst.DSN)
		if err != nil {
			return nil, status.Error(codes.Unknown, err.Error())
		}
		olap, _ := conn.OLAPStore()

		// Drop table with source name
		rows, err := olap.Execute(ctx, &drivers.Statement{Query: fmt.Sprintf("DROP TABLE %s", obj.Name)})
		if err != nil {
			return nil, status.Error(codes.Unknown, err.Error())
		}
		if err = rows.Close(); err != nil {
			return nil, status.Error(codes.Unknown, err.Error())
		}
	case drivers.CatalogObjectTypeUnmanagedTable:
		// Don't allow deletion of tables created directly in DB
		return nil, status.Error(codes.InvalidArgument, "can not delete unmanaged table")
	case drivers.CatalogObjectTypeMetricsView:
		// Nothing to do
	default:
		panic(fmt.Errorf("unhandled catalog object type: %v", obj.Type))
	}

	// Remove from catalog
	err = catalog.DeleteObject(ctx, req.InstanceId, obj.Name)
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "could not delete object: %s", err.Error())
	}

	return &api.MigrateDeleteResponse{}, nil
}
