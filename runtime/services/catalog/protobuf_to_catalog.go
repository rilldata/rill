package catalog

import (
	"github.com/rilldata/rill/runtime/api"
	"github.com/rilldata/rill/runtime/drivers"
	"google.golang.org/protobuf/proto"
)

func pbToCatalogObject(apiCatalog *api.CatalogObject) (*drivers.CatalogObject, error) {
	catalog := &drivers.CatalogObject{
		Name: apiCatalog.Name,
		Path: apiCatalog.Path,
	}

	var err error
	switch apiCatalog.Type {
	case api.CatalogObject_TYPE_SOURCE:
		catalog.SQL = apiCatalog.Source.Sql
		catalog.Type = drivers.CatalogObjectTypeSource
		catalog.Definition, err = proto.Marshal(apiCatalog.Source)
	case api.CatalogObject_TYPE_MODEL:
		catalog.SQL = apiCatalog.Model.Sql
		catalog.Type = drivers.CatalogObjectTypeModel
		catalog.Definition, err = proto.Marshal(apiCatalog.Model)
	case api.CatalogObject_TYPE_METRICS_VIEW:
		catalog.Type = drivers.CatalogObjectTypeMetricsView
		catalog.Definition, err = proto.Marshal(apiCatalog.MetricsView)
	}

	return catalog, err
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
