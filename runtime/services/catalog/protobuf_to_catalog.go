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
		catalog.Definition, err = proto.Marshal(apiCatalog.Source)
	case api.CatalogObject_TYPE_MODEL:
		catalog.SQL = apiCatalog.Model.Sql
		catalog.Definition, err = proto.Marshal(apiCatalog.Model)
	case api.CatalogObject_TYPE_METRICS_VIEW:
		catalog.Definition, err = proto.Marshal(apiCatalog.MetricsView)
	}

	return catalog, err
}
