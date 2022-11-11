package migrator

import (
	"context"

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
		return
		// no panic here. to make sure migrators are registered we could load them multiple times
		//panic(fmt.Errorf("already registered migrator type with name '%s'", name))
	}
	Migrators[name] = artifact
}

type EntityMigrator interface {
	Create(ctx context.Context, olap drivers.OLAPStore, catalog *api.CatalogObject) error
	Update(ctx context.Context, olap drivers.OLAPStore, catalog *api.CatalogObject) error
	Rename(ctx context.Context, olap drivers.OLAPStore, from string, catalog *api.CatalogObject) error
	Delete(ctx context.Context, olap drivers.OLAPStore, catalog *api.CatalogObject) error
	GetDependencies(ctx context.Context, olap drivers.OLAPStore, catalog *api.CatalogObject) []string
	Validate(ctx context.Context, olap drivers.OLAPStore, catalog *api.CatalogObject) error
	// IsEqual checks everything but the name
	IsEqual(ctx context.Context, cat1 *api.CatalogObject, cat2 *api.CatalogObject) bool
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

func GetDependencies(ctx context.Context, olap drivers.OLAPStore, catalog *api.CatalogObject) []string {
	migrator, ok := getMigrator(catalog)
	if !ok {
		// no error here. not all migrators are needed
		return []string{}
	}
	return migrator.GetDependencies(ctx, olap, catalog)
}

// Validate also returns list of dependents
func Validate(ctx context.Context, olap drivers.OLAPStore, catalog *api.CatalogObject) error {
	migrator, ok := getMigrator(catalog)
	if !ok {
		// no error here. not all migrators are needed
		return nil
	}
	return migrator.Validate(ctx, olap, catalog)
}

// IsEqual checks everything but the name
func IsEqual(ctx context.Context, cat1 *api.CatalogObject, cat2 *api.CatalogObject) bool {
	if cat1.Type != cat2.Type {
		return false
	}
	migrator, ok := getMigrator(cat1)
	if !ok {
		// no error here. not all migrators are needed
		return false
	}
	return migrator.IsEqual(ctx, cat1, cat2)
}

func getMigrator(catalog *api.CatalogObject) (EntityMigrator, bool) {
	var objType drivers.CatalogObjectType
	// TODO: temporary for the merge with main
	switch catalog.Type {
	case api.CatalogObject_TYPE_SOURCE:
		objType = drivers.CatalogObjectTypeSource
	case api.CatalogObject_TYPE_MODEL:
		objType = drivers.CatalogObjectTypeModel
	case api.CatalogObject_TYPE_METRICS_VIEW:
		objType = drivers.CatalogObjectTypeMetricsView
	}

	migrator, ok := Migrators[string(objType)]
	return migrator, ok
}
