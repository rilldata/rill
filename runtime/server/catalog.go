package server

import (
	"context"
	"fmt"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/connectors"
	"github.com/rilldata/rill/runtime/drivers"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// ListCatalogEntries implements RuntimeService
func (s *Server) ListCatalogEntries(ctx context.Context, req *runtimev1.ListCatalogEntriesRequest) (*runtimev1.ListCatalogEntriesResponse, error) {
	service, err := s.serviceCache.createCatalogService(ctx, s, req.InstanceId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	pbs, err := service.ListObjects(ctx, req.Type)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &runtimev1.ListCatalogEntriesResponse{Entries: pbs}, nil
}

// GetCatalogEntry implements RuntimeService
func (s *Server) GetCatalogEntry(ctx context.Context, req *runtimev1.GetCatalogEntryRequest) (*runtimev1.GetCatalogEntryResponse, error) {
	service, err := s.serviceCache.createCatalogService(ctx, s, req.InstanceId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	pb, err := service.GetCatalogObject(ctx, req.Name)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &runtimev1.GetCatalogEntryResponse{Entry: pb}, nil
}

// TriggerRefresh implements RuntimeService
func (s *Server) TriggerRefresh(ctx context.Context, req *runtimev1.TriggerRefreshRequest) (*runtimev1.TriggerRefreshResponse, error) {
	registry, _ := s.metastore.RegistryStore()
	inst, found, err := registry.FindInstance(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, status.Error(codes.InvalidArgument, "instance not found")
	}

	catalog, err := s.openCatalog(ctx, inst)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Find object
	obj, found := catalog.FindEntry(ctx, req.InstanceId, req.Name)
	if !found {
		return nil, status.Error(codes.InvalidArgument, "object not found")
	}

	// Check that it's a refreshable object
	switch obj.Type {
	case drivers.ObjectTypeSource:
	default:
		return nil, status.Error(codes.InvalidArgument, "object is not refreshable")
	}

	// Parse SQL
	source := &connectors.Source{
		Name:       obj.GetSource().Name,
		Connector:  obj.GetSource().Connector,
		Properties: obj.GetSource().Properties.AsMap(),
	}

	// Get olap
	conn, err := s.connCache.openAndMigrate(ctx, inst.ID, inst.OLAPDriver, inst.OLAPDSN)
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}
	olap, _ := conn.OLAPStore()

	// Make connector env
	// Since we're deprecating this code soon, this is just a hack to ingest sources from paths relative to pwd
	env := &connectors.Env{
		RepoDriver: inst.RepoDriver,
		RepoDSN:    inst.RepoDSN,
	}

	// Ingest the source
	err = olap.Ingest(ctx, env, source)
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}

	// Update object
	obj.RefreshedOn = time.Now()
	err = catalog.UpdateEntry(ctx, req.InstanceId, obj)
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}

	return &runtimev1.TriggerRefreshResponse{}, nil
}

// TriggerSync implements RuntimeService
func (s *Server) TriggerSync(ctx context.Context, req *runtimev1.TriggerSyncRequest) (*runtimev1.TriggerSyncResponse, error) {
	// TODO: move to using reconcile
	// Get instance
	registry, _ := s.metastore.RegistryStore()
	inst, found, err := registry.FindInstance(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, status.Error(codes.InvalidArgument, "instance not found")
	}

	// Get OLAP
	conn, err := s.connCache.openAndMigrate(ctx, inst.ID, inst.OLAPDriver, inst.OLAPDSN)
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
	objs := catalogStore.FindEntries(ctx, req.InstanceId, drivers.ObjectTypeUnspecified)

	// Get information schema
	tables, err := olap.InformationSchema().All(ctx)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, err.Error())
	}

	// Index objects for lookup
	objMap := make(map[string]*drivers.CatalogEntry)
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
		if ok && obj.Type == drivers.ObjectTypeTable && !obj.GetTable().Managed {
			// If the table has already been synced, update the schema if it has changed
			tbl := obj.GetTable()
			if !proto.Equal(t.Schema, tbl.Schema) {
				tbl.Schema = t.Schema
				err := catalogStore.UpdateEntry(ctx, inst.ID, obj)
				if err != nil {
					return nil, status.Errorf(codes.FailedPrecondition, err.Error())
				}
				updated++
			}
		} else if !ok {
			// If we haven't seen this table before, add it
			err := catalogStore.CreateEntry(ctx, inst.ID, &drivers.CatalogEntry{
				Name: t.Name,
				Type: drivers.ObjectTypeTable,
				Object: &runtimev1.Table{
					Name:    t.Name,
					Schema:  t.Schema,
					Managed: false,
				},
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
		if !seen && obj.Type == drivers.ObjectTypeTable && !obj.GetTable().Managed {
			err := catalogStore.DeleteEntry(ctx, inst.ID, name)
			if err != nil {
				return nil, status.Errorf(codes.FailedPrecondition, err.Error())
			}
			removed++
		}
	}

	// Done
	return &runtimev1.TriggerSyncResponse{
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

	conn, err := s.connCache.openAndMigrate(ctx, inst.ID, inst.OLAPDriver, inst.OLAPDSN)
	if err != nil {
		return nil, err
	}

	catalogStore, ok := conn.CatalogStore()
	if !ok {
		return nil, fmt.Errorf("instance cannot embed catalog")
	}

	return catalogStore, nil
}
