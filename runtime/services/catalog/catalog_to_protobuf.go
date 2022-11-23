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
	catalog := &api.CatalogObject{
		Name:        obj.Name,
		Path:        obj.Path,
		CreatedOn:   timestamppb.New(obj.CreatedOn),
		UpdatedOn:   timestamppb.New(obj.UpdatedOn),
		RefreshedOn: timestamppb.New(obj.RefreshedOn),
	}

	switch obj.Type {
	case drivers.CatalogObjectTypeSource:
		src, err := catalogObjectSourceToPB(obj)
		if err != nil {
			return nil, err
		}
		catalog.Type = api.CatalogObject_TYPE_SOURCE
		catalog.Source = src

	case drivers.CatalogObjectTypeModel:
		model, err := catalogObjectModelToPB(obj)
		if err != nil {
			return nil, err
		}
		catalog.Type = api.CatalogObject_TYPE_MODEL
		catalog.Model = model

	case drivers.CatalogObjectTypeMetricsView:
		metricsView, err := catalogObjectMetricsViewToPB(obj)
		if err != nil {
			return nil, err
		}
		catalog.Type = api.CatalogObject_TYPE_METRICS_VIEW
		catalog.MetricsView = metricsView

	default:
		fmt.Println("not implemented")
	}

	return catalog, nil
}

func catalogObjectSourceToPB(obj *drivers.CatalogObject) (*api.Source, error) {
	if obj.SQL == "" {
		source := &api.Source{}
		err := proto.Unmarshal(obj.Definition, source)
		source.Schema = obj.Schema
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
		Schema:     obj.Schema,
	}, nil
}

func catalogObjectModelToPB(obj *drivers.CatalogObject) (*api.Model, error) {
	return &api.Model{
		Name:    obj.Name,
		Sql:     obj.SQL,
		Dialect: api.Model_DIALECT_DUCKDB,
		Schema:  obj.Schema,
	}, nil
}

func catalogObjectMetricsViewToPB(obj *drivers.CatalogObject) (*api.MetricsView, error) {
	var metricsView api.MetricsView
	err := proto.Unmarshal(obj.Definition, &metricsView)
	return &metricsView, err
}
