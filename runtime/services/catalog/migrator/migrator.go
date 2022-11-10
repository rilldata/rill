package migrator

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/runtime/api"
	"github.com/rilldata/rill/runtime/drivers"
)

/**
 * Any entity specific actions when a catalog is deleted will be here.
 * EG: on delete sources will drop the table and models will drop the view
 * Any future entity specific cache invalidation will go here as well.
 */

var Migrators = make(map[string]EntityMigrator)

func Register(name string, artifact EntityMigrator) {
	if Migrators[name] != nil {
		panic(fmt.Errorf("already registered migrator type with name '%s'", name))
	}
	Migrators[name] = artifact
}

type EntityMigrator interface {
	Create(ctx context.Context, olap drivers.OLAPStore, catalog *api.CatalogObject) error
	Update(ctx context.Context, olap drivers.OLAPStore, catalog *api.CatalogObject) error
	Rename(ctx context.Context, olap drivers.OLAPStore, from string, catalog *api.CatalogObject) error
	Delete(ctx context.Context, olap drivers.OLAPStore, catalog *api.CatalogObject) error
}

func Create(ctx context.Context, olap drivers.OLAPStore, catalog *api.CatalogObject) error {
	migrator, ok := getMigrator(catalog)
	if !ok {
		// no error here. not all migrators are needed
		return nil
	}
	return migrator.Create(ctx, olap, catalog)
}

func Update(ctx context.Context, olap drivers.OLAPStore, catalog *api.CatalogObject) error {
	migrator, ok := getMigrator(catalog)
	if !ok {
		// no error here. not all migrators are needed
		return nil
	}
	return migrator.Update(ctx, olap, catalog)
}

func Rename(ctx context.Context, olap drivers.OLAPStore, from string, catalog *api.CatalogObject) error {
	migrator, ok := getMigrator(catalog)
	if !ok {
		// no error here. not all migrators are needed
		return nil
	}
	return migrator.Rename(ctx, olap, from, catalog)
}

func Delete(ctx context.Context, olap drivers.OLAPStore, catalog *api.CatalogObject) error {
	migrator, ok := getMigrator(catalog)
	if !ok {
		// no error here. not all migrators are needed
		return nil
	}
	return migrator.Delete(ctx, olap, catalog)
}

func getMigrator(catalog *api.CatalogObject) (EntityMigrator, bool) {
	var objType string
	switch catalog.Type.(type) {
	case *api.CatalogObject_Source:
		objType = drivers.CatalogObjectTypeSource
	case *api.CatalogObject_Model:
		objType = drivers.CatalogObjectTypeModel
	}

	migrator, ok := Migrators[objType]
	return migrator, ok
}
