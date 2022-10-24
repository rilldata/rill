package server

import (
	"context"
	"fmt"
	"strings"

	"github.com/rilldata/rill/runtime/api"
	"github.com/rilldata/rill/runtime/connectors"
	"github.com/rilldata/rill/runtime/drivers"
	sql "github.com/rilldata/rill/runtime/sql/pure"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

// ListCatalogObjects implements RuntimeService
func (s *Server) ListCatalogObjects(ctx context.Context, req *api.ListCatalogObjectsRequest) (*api.ListCatalogObjectsResponse, error) {
	registry, _ := s.metastore.RegistryStore()
	inst, found := registry.FindInstance(ctx, req.InstanceId)
	if !found {
		return nil, status.Error(codes.InvalidArgument, "instance not found")
	}

	catalog, err := s.openCatalog(ctx, inst)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	objs := catalog.FindObjects(ctx, req.InstanceId, catalogObjectTypeFromPB(req.Type))
	pbs := make([]*api.CatalogObject, len(objs))
	for i, obj := range objs {
		pbs[i], err = catalogObjectToPB(obj)
		if err != nil {
			return nil, status.Error(codes.Unknown, err.Error())
		}
	}

	return &api.ListCatalogObjectsResponse{Objects: pbs}, nil
}

// GetCatalogObject implements RuntimeService
func (s *Server) GetCatalogObject(ctx context.Context, req *api.GetCatalogObjectRequest) (*api.GetCatalogObjectResponse, error) {
	registry, _ := s.metastore.RegistryStore()
	inst, found := registry.FindInstance(ctx, req.InstanceId)
	if !found {
		return nil, status.Error(codes.InvalidArgument, "instance not found")
	}

	catalog, err := s.openCatalog(ctx, inst)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	obj, found := catalog.FindObject(ctx, req.InstanceId, strings.ToLower(req.Name))
	if !found {
		return nil, status.Error(codes.InvalidArgument, "object not found")
	}

	pb, err := catalogObjectToPB(obj)
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
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
	obj, found := catalog.FindObject(ctx, req.InstanceId, strings.ToLower(req.Name))
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
	source, err := sqlToSource(obj.SQL)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Get olap
	conn, err := s.cache.openAndMigrate(ctx, inst.ID, inst.Driver, inst.DSN)
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}
	olap, _ := conn.OLAPStore()

	// Ingest the source
	err = olap.Ingest(ctx, source)
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}

	return &api.TriggerRefreshResponse{}, nil
}

// TriggerSync implements RuntimeService
func (s *Server) TriggerSync(ctx context.Context, req *api.TriggerSyncRequest) (*api.TriggerSyncResponse, error) {
	// TODO:
	return nil, nil
}

func (s *Server) openCatalog(ctx context.Context, inst *drivers.Instance) (drivers.CatalogStore, error) {
	if !inst.EmbedCatalog {
		catalog, ok := s.metastore.CatalogStore()
		if !ok {
			return nil, fmt.Errorf("metastore cannot serve as catalog")
		}
		return catalog, nil
	}

	conn, err := s.cache.openAndMigrate(ctx, inst.ID, inst.Driver, inst.DSN)
	if err != nil {
		return nil, err
	}

	catalog, ok := conn.CatalogStore()
	if !ok {
		return nil, fmt.Errorf("instance cannot embed catalog")
	}

	return catalog, nil
}

func catalogObjectTypeFromPB(t api.CatalogObject_Type) drivers.CatalogObjectType {
	switch t {
	case api.CatalogObject_TYPE_UNSPECIFIED:
		return drivers.CatalogObjectTypeUnspecified
	case api.CatalogObject_TYPE_TABLE:
		return drivers.CatalogObjectTypeTable
	case api.CatalogObject_TYPE_SOURCE:
		return drivers.CatalogObjectTypeSource
	case api.CatalogObject_TYPE_METRICS_VIEW:
		return drivers.CatalogObjectTypeMetricsView
	default:
		// NOTE: Consider returning and handling an error instead
		return drivers.CatalogObjectTypeUnspecified
	}
}

func catalogObjectToPB(obj *drivers.CatalogObject) (*api.CatalogObject, error) {
	switch obj.Type {
	case drivers.CatalogObjectTypeTable:
		return catalogObjectTableToPB(obj)
	case drivers.CatalogObjectTypeSource:
		return catalogObjectSourceToPB(obj)
	case drivers.CatalogObjectTypeMetricsView:
		return catalogObjectMetricsViewToPB(obj)
	default:
		panic(fmt.Errorf("not implemented"))
	}
}

func catalogObjectTableToPB(obj *drivers.CatalogObject) (*api.CatalogObject, error) {
	return &api.CatalogObject{
		Type: api.CatalogObject_TYPE_TABLE,
		Table: &api.Table{
			Name:    obj.Name,
			Schema:  obj.Schema,
			Managed: obj.Managed,
		},
	}, nil
}

func catalogObjectSourceToPB(obj *drivers.CatalogObject) (*api.CatalogObject, error) {
	source, err := sqlToSource(obj.SQL)
	if err != nil {
		return nil, err
	}

	propsPB, err := structpb.NewStruct(source.Properties)
	if err != nil {
		panic(err) // TODO: Should never happen, but maybe handle defensively?
	}

	return &api.CatalogObject{
		Type: api.CatalogObject_TYPE_SOURCE,
		Source: &api.Source{
			Sql:        obj.SQL,
			Name:       obj.Name,
			Connector:  source.Connector,
			Properties: propsPB,
		},
	}, nil
}

func catalogObjectMetricsViewToPB(obj *drivers.CatalogObject) (*api.CatalogObject, error) {
	mv, err := sqlToMetricsView(obj.SQL)
	if err != nil {
		return nil, err
	}

	return &api.CatalogObject{
		Type:        api.CatalogObject_TYPE_METRICS_VIEW,
		MetricsView: mv,
	}, nil
}

func sqlToSource(sqlStr string) (*connectors.Source, error) {
	astStmt, err := sql.Parse(sqlStr)
	if err != nil {
		return nil, fmt.Errorf("parse error: %s", err.Error())
	}

	if astStmt.CreateSource == nil {
		return nil, fmt.Errorf("refresh error: object cannot be refreshed")
	}

	ast := astStmt.CreateSource

	s := &connectors.Source{
		Name:       strings.ToLower(ast.Name),
		Properties: make(map[string]any),
	}

	for _, prop := range ast.With.Properties {
		if strings.ToLower(prop.Key) == "connector" {
			s.Connector = safePtrToStr(prop.Value.String)
			continue
		}
		if prop.Value.Number != nil {
			s.Properties[prop.Key] = *prop.Value.Number
		} else if prop.Value.String != nil {
			s.Properties[prop.Key] = *prop.Value.String
		} else if prop.Value.Boolean != nil {
			s.Properties[prop.Key] = *prop.Value.Boolean
		}
	}

	err = s.Validate()
	if err != nil {
		return nil, err
	}

	return s, nil
}

func sqlToMetricsView(sqlStr string) (*api.MetricsView, error) {
	mv := &api.MetricsView{Sql: sqlStr}

	// TODO

	return mv, nil
}

func safePtrToStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
