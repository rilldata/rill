package server

import (
	"context"
	"fmt"
	"time"

	"github.com/rilldata/rill/runtime/api"
	"github.com/rilldata/rill/runtime/connectors"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/services/catalog/migrator/sources"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// ListCatalogObjects implements RuntimeService
func (s *Server) ListCatalogObjects(ctx context.Context, req *api.ListCatalogObjectsRequest) (*api.ListCatalogObjectsResponse, error) {
	service, err := s.serviceCache.createCatalogService(ctx, s, req.InstanceId, "") // TODO: Remove repo ID
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	pbs, err := service.ListObjects(ctx, req.Type)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &api.ListCatalogObjectsResponse{Objects: pbs}, nil
}

// GetCatalogObject implements RuntimeService
func (s *Server) GetCatalogObject(ctx context.Context, req *api.GetCatalogObjectRequest) (*api.GetCatalogObjectResponse, error) {
	service, err := s.serviceCache.createCatalogService(ctx, s, req.InstanceId, "") // TODO: Remove repo ID
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	pb, err := service.GetCatalogObject(ctx, req.Name)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &api.GetCatalogObjectResponse{Object: pb}, nil
}

// TriggerRefresh implements RuntimeService
func (s *Server) TriggerRefresh(ctx context.Context, req *api.TriggerRefreshRequest) (*api.TriggerRefreshResponse, error) {
	registry, _ := s.metastore.RegistryStore()
	inst, found := registry.FindInstance(ctx, req.InstanceId)
	if !found {
		return nil, status.Error(codes.InvalidArgument, "instance not found")
	}

	catalog, err := s.openCatalog(ctx, inst)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Find object
	obj, found := catalog.FindObject(ctx, req.InstanceId, req.Name)
	if !found {
		return nil, status.Error(codes.InvalidArgument, "object not found")
	}

	// Check that it's a refreshable object
	switch obj.Type {
	case drivers.CatalogObjectTypeSource:
	default:
		return nil, status.Error(codes.InvalidArgument, "object is not refreshable")
	}

	// Parse SQL
	source, err := sources.SqlToSource(obj.SQL)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Get olap
	conn, err := s.connCache.openAndMigrate(ctx, inst.ID, inst.Driver, inst.DSN)
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}
	olap, _ := conn.OLAPStore()

	// Make connector env
	// Since we're deprecating this code soon, this is just a hack to ingest sources from paths relative to pwd
	env := &connectors.Env{
		RepoDriver: "file",
		RepoDSN:    ".",
	}

	// Ingest the source
	err = olap.Ingest(ctx, env, source)
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}

	// Update object
	obj.RefreshedOn = time.Now()
	err = catalog.UpdateObject(ctx, req.InstanceId, obj)
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}

	return &api.TriggerRefreshResponse{}, nil
}

// TriggerSync implements RuntimeService
func (s *Server) TriggerSync(ctx context.Context, req *api.TriggerSyncRequest) (*api.TriggerSyncResponse, error) {
	// TODO: move to using migrate
	// Get instance
	registry, _ := s.metastore.RegistryStore()
	inst, found := registry.FindInstance(ctx, req.InstanceId)
	if !found {
		return nil, status.Error(codes.InvalidArgument, "instance not found")
	}

	// Get OLAP
	conn, err := s.connCache.openAndMigrate(ctx, inst.ID, inst.Driver, inst.DSN)
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}
	olap, _ := conn.OLAPStore()

	// Get catalog
	catalogStore, err := s.openCatalog(ctx, inst)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Get full catalog
	objs := catalogStore.FindObjects(ctx, req.InstanceId, drivers.CatalogObjectTypeUnspecified)

	// Get information schema
	tables, err := olap.InformationSchema().All(ctx)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, err.Error())
	}

	// Index objects for lookup
	objMap := make(map[string]*drivers.CatalogObject)
	objSeen := make(map[string]bool)
	for _, obj := range objs {
		objMap[obj.Name] = obj
		objSeen[obj.Name] = false
	}

	// Process tables in information schema
	added := 0
	updated := 0
	for _, t := range tables {
		obj, ok := objMap[t.Name]

		// Track that the object still exists
		if ok {
			objSeen[t.Name] = true
		}

		// Create or update in catalog if relevant
		if ok && obj.Type == drivers.CatalogObjectTypeTable && !obj.Managed {
			// If the table has already been synced, update the schema if it has changed
			if !proto.Equal(t.Schema, obj.Schema) {
				obj.Schema = t.Schema
				err := catalogStore.UpdateObject(ctx, inst.ID, obj)
				if err != nil {
					return nil, status.Errorf(codes.FailedPrecondition, err.Error())
				}
				updated++
			}
		} else if !ok {
			// If we haven't seen this table before, add it
			err := catalogStore.CreateObject(ctx, inst.ID, &drivers.CatalogObject{
				Name:    t.Name,
				Type:    drivers.CatalogObjectTypeTable,
				Schema:  t.Schema,
				Managed: false,
			})
			if err != nil {
				return nil, status.Errorf(codes.FailedPrecondition, err.Error())
			}
			added++
		}
		// Defensively do nothing in all other cases
	}

	// Remove non-managed tables not found in information schema
	removed := 0
	for name, seen := range objSeen {
		obj := objMap[name]
		if !seen && obj.Type == drivers.CatalogObjectTypeTable && !obj.Managed {
			err := catalogStore.DeleteObject(ctx, inst.ID, name)
			if err != nil {
				return nil, status.Errorf(codes.FailedPrecondition, err.Error())
			}
			removed++
		}
	}

	// Done
	return &api.TriggerSyncResponse{
		ObjectsCount:        uint32(len(tables)),
		ObjectsAddedCount:   uint32(added),
		ObjectsUpdatedCount: uint32(updated),
		ObjectsRemovedCount: uint32(removed),
	}, nil
}

func (s *Server) openCatalog(ctx context.Context, inst *drivers.Instance) (drivers.CatalogStore, error) {
	if !inst.EmbedCatalog {
		catalogStore, ok := s.metastore.CatalogStore()
		if !ok {
			return nil, fmt.Errorf("metastore cannot serve as catalog")
		}
		return catalogStore, nil
	}

	conn, err := s.connCache.openAndMigrate(ctx, inst.ID, inst.Driver, inst.DSN)
	if err != nil {
		return nil, err
	}

	catalogStore, ok := conn.CatalogStore()
	if !ok {
		return nil, fmt.Errorf("instance cannot embed catalog")
	}

	return catalogStore, nil
}
