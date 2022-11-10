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
	switch apiCatalog.Type.(type) {
	case *api.CatalogObject_Source:
		catalog.Definition, err = sourcePbToCatalogObject(apiCatalog, catalog)
	case *api.CatalogObject_Model:
		catalog.Definition, err = modelPbToCatalogObject(apiCatalog, catalog)
	case *api.CatalogObject_MetricsView:
		catalog.Definition, err = metricsViewPbToCatalogObject(apiCatalog)
	}

	return catalog, err
}

func sourcePbToCatalogObject(apiCatalog *api.CatalogObject, catalog *drivers.CatalogObject) ([]byte, error) {
	source := apiCatalog.Type.(*api.CatalogObject_Source).Source
	catalog.SQL = source.Sql
	return proto.Marshal(source)
}

func modelPbToCatalogObject(apiCatalog *api.CatalogObject, catalog *drivers.CatalogObject) ([]byte, error) {
	model := apiCatalog.Type.(*api.CatalogObject_Model).Model
	catalog.SQL = model.Sql
	return proto.Marshal(model)
}

func metricsViewPbToCatalogObject(apiCatalog *api.CatalogObject) ([]byte, error) {
	metricsView := apiCatalog.Type.(*api.CatalogObject_MetricsView).MetricsView
	return proto.Marshal(metricsView)
}
