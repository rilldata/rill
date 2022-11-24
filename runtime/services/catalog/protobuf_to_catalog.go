package catalog

import (
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func pbToObjectType(in runtimev1.ObjectType) drivers.ObjectType {
	switch in {
	case runtimev1.ObjectType_OBJECT_TYPE_UNSPECIFIED:
		return drivers.ObjectTypeUnspecified
	case runtimev1.ObjectType_OBJECT_TYPE_TABLE:
		return drivers.ObjectTypeTable
	case runtimev1.ObjectType_OBJECT_TYPE_SOURCE:
		return drivers.ObjectTypeSource
	case runtimev1.ObjectType_OBJECT_TYPE_MODEL:
		return drivers.ObjectTypeModel
	case runtimev1.ObjectType_OBJECT_TYPE_METRICS_VIEW:
		return drivers.ObjectTypeMetricsView
	}
	panic(fmt.Errorf("unhandled object type %s", in))
}

func catalogObjectToPB(obj *drivers.CatalogEntry) (*runtimev1.CatalogEntry, error) {
	catalog := &runtimev1.CatalogEntry{
		Name:        obj.Name,
		Path:        obj.Path,
		CreatedOn:   timestamppb.New(obj.CreatedOn),
		UpdatedOn:   timestamppb.New(obj.UpdatedOn),
		RefreshedOn: timestamppb.New(obj.RefreshedOn),
	}

	switch obj.Type {
	case drivers.ObjectTypeTable:
		catalog.Object = &runtimev1.CatalogEntry_Table{
			Table: obj.GetTable(),
		}
	case drivers.ObjectTypeSource:
		catalog.Object = &runtimev1.CatalogEntry_Source{
			Source: obj.GetSource(),
		}
	case drivers.ObjectTypeModel:
		catalog.Object = &runtimev1.CatalogEntry_Model{
			Model: obj.GetModel(),
		}
	case drivers.ObjectTypeMetricsView:
		catalog.Object = &runtimev1.CatalogEntry_MetricsView{
			MetricsView: obj.GetMetricsView(),
		}
	default:
		panic("not implemented")
	}

	return catalog, nil
}
