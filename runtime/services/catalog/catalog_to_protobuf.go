package catalog

import (
	"fmt"

	"github.com/rilldata/rill/runtime/api"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/services/catalog/migrator/sources"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func catalogObjectToPB(obj *drivers.CatalogObject) (*api.CatalogObject, error) {
	switch obj.Type {
	case drivers.CatalogObjectTypeSource:
		src, err := catalogObjectSourceToPB(obj)
		if err != nil {
			return nil, err
		}
		return &api.CatalogObject{
			Name:        src.Name,
			Type:        api.CatalogObject_TYPE_SOURCE,
			Source:      src,
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
			Name:        model.Name,
			Type:        api.CatalogObject_TYPE_MODEL,
			Model:       model,
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
			Name:        metricsView.Name,
			Type:        api.CatalogObject_TYPE_METRICS_VIEW,
			MetricsView: metricsView,
			CreatedOn:   timestamppb.New(obj.CreatedOn),
			UpdatedOn:   timestamppb.New(obj.UpdatedOn),
			RefreshedOn: timestamppb.New(obj.RefreshedOn),
		}, nil

	default:
		panic(fmt.Errorf("not implemented"))
	}
}

func catalogObjectSourceToPB(obj *drivers.CatalogObject) (*api.Source, error) {
	if obj.SQL == "" {
		source := &api.Source{}
		err := proto.Unmarshal(obj.Definition, source)
		if err != nil {
			return nil, err
		}
		return source, nil
	}

	source, err := sources.SqlToSource(obj.SQL)
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
