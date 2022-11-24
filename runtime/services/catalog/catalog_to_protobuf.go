package catalog

import (
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/services/catalog/migrator/sources"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func catalogObjectToPB(obj *drivers.CatalogObject) (*runtimev1.CatalogObject, error) {
	catalog := &runtimev1.CatalogObject{
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
		catalog.Type = runtimev1.CatalogObject_TYPE_SOURCE
		catalog.Source = src

	case drivers.CatalogObjectTypeModel:
		model, err := catalogObjectModelToPB(obj)
		if err != nil {
			return nil, err
		}
		catalog.Type = runtimev1.CatalogObject_TYPE_MODEL
		catalog.Model = model

	case drivers.CatalogObjectTypeMetricsView:
		metricsView, err := catalogObjectMetricsViewToPB(obj)
		if err != nil {
			return nil, err
		}
		catalog.Type = runtimev1.CatalogObject_TYPE_METRICS_VIEW
		catalog.MetricsView = metricsView

	default:
		fmt.Println("not implemented")
	}

	return catalog, nil
}

func catalogObjectSourceToPB(obj *drivers.CatalogObject) (*runtimev1.Source, error) {
	if obj.SQL == "" {
		source := &runtimev1.Source{}
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

	return &runtimev1.Source{
		Sql:        obj.SQL,
		Name:       obj.Name,
		Connector:  source.Connector,
		Properties: propsPB,
	}, nil
}

func catalogObjectModelToPB(obj *drivers.CatalogObject) (*runtimev1.Model, error) {
	return &runtimev1.Model{
		Name:    obj.Name,
		Sql:     obj.SQL,
		Dialect: runtimev1.Model_DIALECT_DUCKDB,
	}, nil
}

func catalogObjectMetricsViewToPB(obj *drivers.CatalogObject) (*runtimev1.MetricsView, error) {
	var metricsView runtimev1.MetricsView
	err := proto.Unmarshal(obj.Definition, &metricsView)
	return &metricsView, err
}
