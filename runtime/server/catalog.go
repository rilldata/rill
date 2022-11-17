package server

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/runtime/api"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/services/catalog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// InitCatalogService implements RuntimeService
func (s *Server) InitCatalogService(ctx context.Context, req *api.InitCatalogServiceRequest) (*api.InitCatalogServiceResponse, error) {
	instResp, err := s.CreateInstance(ctx, req.Instance)
	if err != nil {
		return nil, err
	}

	repoResp, err := s.CreateRepo(ctx, req.Repo)
	if err != nil {
		return nil, err
	}

	service, err := s.serviceCache.createCatalogService(ctx, s, instResp.Instance.InstanceId, repoResp.Repo.RepoId)
	if err != nil {
		return nil, err
	}
	resp, err := service.Init(ctx)
	if err != nil {
		return nil, err
	}
	if len(resp.Errors) > 0 {
		// TODO: send more errors
		return nil, status.Error(codes.Unknown, resp.Errors[0].Message)
	}

	return &api.InitCatalogServiceResponse{
		Instance: instResp.Instance,
		Repo:     repoResp.Repo,
	}, nil
}

// ListCatalogObjects implements RuntimeService
func (s *Server) ListCatalogObjects(ctx context.Context, req *api.ListCatalogObjectsRequest) (*api.ListCatalogObjectsResponse, error) {
	service, err := s.serviceCache.createCatalogService(ctx, s, req.InstanceId, "")
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
	service, err := s.serviceCache.createCatalogService(ctx, s, req.InstanceId, "")
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
	service, err := s.serviceCache.createCatalogService(ctx, s, req.InstanceId, "")
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	path, ok := service.NameToPath[req.Name]
	if !ok {
		return nil, status.Error(codes.NotFound, "artifact not found")
	}

	resp, err := service.Migrate(ctx, catalog.MigrationConfig{
		ChangedPaths: []string{path},
		ForcedPaths:  []string{path},
		Strict:       true,
	})
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}
	if len(resp.Errors) > 0 {
		return nil, status.Error(codes.Unknown, resp.Errors[0].Message)
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

	// Reset catalog cache
	s.catalogCache.reset(req.InstanceId)

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
