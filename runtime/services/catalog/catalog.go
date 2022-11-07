package catalog

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/rilldata/rill/runtime/api"
	"github.com/rilldata/rill/runtime/connectors"
	"github.com/rilldata/rill/runtime/drivers"
	sql "github.com/rilldata/rill/runtime/sql/pure"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Service struct {
}

func (s *Service) ListObjects(
	ctx context.Context,
	inst *drivers.Instance,
	catalog drivers.CatalogStore,
) ([]*api.CatalogObject, error) {
	objs := catalog.FindObjects(ctx, inst.ID)
	pbs := make([]*api.CatalogObject, len(objs))
	var err error
	for i, obj := range objs {
		pbs[i], err = catalogObjectToPB(obj)
		if err != nil {
			return nil, status.Error(codes.Unknown, err.Error())
		}
	}

	return pbs, nil
}

func (s *Service) GetCatalogObject(
	ctx context.Context,
	inst *drivers.Instance,
	name string,
	catalog drivers.CatalogStore,
) (*api.CatalogObject, error) {
	obj, found := catalog.FindObject(ctx, inst.ID, name)
	if !found {
		return nil, status.Error(codes.InvalidArgument, "object not found")
	}

	pb, err := catalogObjectToPB(obj)
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}

	return pb, nil
}

func (s *Service) TriggerRefresh(
	ctx context.Context,
	inst *drivers.Instance,
	name string,
	catalog drivers.CatalogStore,
	olap drivers.OLAPStore,
) error {
	// Find object
	obj, found := catalog.FindObject(ctx, inst.ID, name)
	if !found {
		return status.Error(codes.InvalidArgument, "object not found")
	}

	switch obj.Type {
	case drivers.CatalogObjectTypeSource:
		// Parse SQL
		source, err := sqlToSource(obj.SQL)
		if err != nil {
			return status.Error(codes.InvalidArgument, err.Error())
		}
		// Ingest the source
		err = olap.Ingest(ctx, source)
		if err != nil {
			return status.Error(codes.Unknown, err.Error())
		}

		// Update object
		obj.RefreshedOn = time.Now()
		err = catalog.UpdateObject(ctx, inst.ID, obj)

	case drivers.CatalogObjectTypeModel:
		//TODO
	}

	return nil
}

func catalogObjectToPB(obj *drivers.CatalogObject) (*api.CatalogObject, error) {
	switch obj.Type {
	case drivers.CatalogObjectTypeSource:
		src, err := catalogObjectSourceToPB(obj)
		if err != nil {
			return nil, err
		}
		return &api.CatalogObject{
			Type: &api.CatalogObject_Source{
				Source: src,
			},
			CreatedOn:   timestamppb.New(obj.CreatedOn),
			UpdatedOn:   timestamppb.New(obj.UpdatedOn),
			RefreshedOn: timestamppb.New(obj.RefreshedOn),
		}, nil

	case drivers.CatalogObjectTypeModel:
		model, err := catalogObjectModelToPB(obj)
		if err != nil {
			return nil, err
		}
		return &api.CatalogObject{
			Type: &api.CatalogObject_Model{
				Model: model,
			},
			CreatedOn:   timestamppb.New(obj.CreatedOn),
			UpdatedOn:   timestamppb.New(obj.UpdatedOn),
			RefreshedOn: timestamppb.New(obj.RefreshedOn),
		}, nil

	case drivers.CatalogObjectTypeMetricsView:
		metricsView, err := catalogObjectMetricsViewToPB(obj)
		if err != nil {
			return nil, err
		}
		return &api.CatalogObject{
			Type: &api.CatalogObject_MetricsView{
				MetricsView: metricsView,
			},
			CreatedOn:   timestamppb.New(obj.CreatedOn),
			UpdatedOn:   timestamppb.New(obj.UpdatedOn),
			RefreshedOn: timestamppb.New(obj.RefreshedOn),
		}, nil

	default:
		panic(fmt.Errorf("not implemented"))
	}
}

func catalogObjectSourceToPB(obj *drivers.CatalogObject) (*api.Source, error) {
	source, err := sqlToSource(obj.SQL)
	if err != nil {
		return nil, err
	}

	propsPB, err := structpb.NewStruct(source.Properties)
	if err != nil {
		panic(err) // TODO: Should never happen, but maybe handle defensively?
	}

	return &api.Source{
		Sql:        obj.SQL,
		Name:       obj.Name,
		Connector:  source.Connector,
		Properties: propsPB,
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
		Name:       ast.Name,
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

func catalogObjectModelToPB(obj *drivers.CatalogObject) (*api.Model, error) {
	return &api.Model{
		Name:    obj.Name,
		Sql:     obj.SQL,
		Dialect: api.Model_DuckDB,
	}, nil
}

func catalogObjectMetricsViewToPB(obj *drivers.CatalogObject) (*api.MetricsView, error) {
	var metricsView api.MetricsView
	err := proto.Unmarshal(obj.Definition, &metricsView)
	return &metricsView, err
}

func safePtrToStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
