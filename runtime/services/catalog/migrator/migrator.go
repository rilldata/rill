package migrator

import (
	"context"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
)

/**
 * Any entity specific actions when a catalog is deleted will be here.
 * EG: on delete sources will drop the table and models will drop the view
 * Any future entity specific cache invalidation will go here as well.
 *
 * TODO: does migrator name fit this?
 * TODO: is this in the right place?
 */

var Migrators = make(map[string]EntityMigrator)

func Register(name string, artifact EntityMigrator) {
	if Migrators[name] != nil {
		panic(fmt.Errorf("already registered migrator type with name '%s'", name))
	}
	Migrators[name] = artifact
}

type EntityMigrator interface {
	Create(ctx context.Context, olap drivers.OLAPStore, repo drivers.RepoStore, catalog *runtimev1.CatalogObject) error
	Update(ctx context.Context, olap drivers.OLAPStore, repo drivers.RepoStore, catalog *runtimev1.CatalogObject) error
	Rename(ctx context.Context, olap drivers.OLAPStore, from string, catalog *runtimev1.CatalogObject) error
	Delete(ctx context.Context, olap drivers.OLAPStore, catalog *runtimev1.CatalogObject) error
	GetDependencies(ctx context.Context, olap drivers.OLAPStore, catalog *runtimev1.CatalogObject) []string
	Validate(ctx context.Context, olap drivers.OLAPStore, catalog *runtimev1.CatalogObject) error
	// IsEqual checks everything but the name
	IsEqual(ctx context.Context, cat1 *runtimev1.CatalogObject, cat2 *runtimev1.CatalogObject) bool
	ExistsInOlap(ctx context.Context, olap drivers.OLAPStore, catalog *runtimev1.CatalogObject) (bool, error)
}

func Create(ctx context.Context, olap drivers.OLAPStore, repo drivers.RepoStore, catalog *runtimev1.CatalogObject) error {
	migrator, ok := getMigrator(catalog)
	if !ok {
		// no error here. not all migrators are needed
		return nil
	}
	return migrator.Create(ctx, olap, repo, catalog)
}

func Update(ctx context.Context, olap drivers.OLAPStore, repo drivers.RepoStore, catalog *runtimev1.CatalogObject) error {
	migrator, ok := getMigrator(catalog)
	if !ok {
		// no error here. not all migrators are needed
		return nil
	}
	return migrator.Update(ctx, olap, repo, catalog)
}

func Rename(ctx context.Context, olap drivers.OLAPStore, from string, catalog *runtimev1.CatalogObject) error {
	migrator, ok := getMigrator(catalog)
	if !ok {
		// no error here. not all migrators are needed
		return nil
	}
	return migrator.Rename(ctx, olap, from, catalog)
}

func Delete(ctx context.Context, olap drivers.OLAPStore, catalog *runtimev1.CatalogObject) error {
	migrator, ok := getMigrator(catalog)
	if !ok {
		// no error here. not all migrators are needed
		return nil
	}
	return migrator.Delete(ctx, olap, catalog)
}

func GetDependencies(ctx context.Context, olap drivers.OLAPStore, catalog *runtimev1.CatalogObject) []string {
	migrator, ok := getMigrator(catalog)
	if !ok {
		// no error here. not all migrators are needed
		return []string{}
	}
	return migrator.GetDependencies(ctx, olap, catalog)
}

// Validate also returns list of dependents
func Validate(ctx context.Context, olap drivers.OLAPStore, catalog *runtimev1.CatalogObject) error {
	migrator, ok := getMigrator(catalog)
	if !ok {
		// no error here. not all migrators are needed
		return nil
	}
	return migrator.Validate(ctx, olap, catalog)
}

// IsEqual checks everything but the name
func IsEqual(ctx context.Context, cat1 *runtimev1.CatalogObject, cat2 *runtimev1.CatalogObject) bool {
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

func ExistsInOlap(ctx context.Context, olap drivers.OLAPStore, catalog *runtimev1.CatalogObject) (bool, error) {
	migrator, ok := getMigrator(catalog)
	if !ok {
		// no error here. not all migrators are needed
		return false, nil
	}
	return migrator.ExistsInOlap(ctx, olap, catalog)
}

func getMigrator(catalog *runtimev1.CatalogObject) (EntityMigrator, bool) {
	var objType drivers.CatalogObjectType
	// TODO: temporary for the merge with main
	switch catalog.Type {
	case runtimev1.CatalogObject_TYPE_SOURCE:
		objType = drivers.CatalogObjectTypeSource
	case runtimev1.CatalogObject_TYPE_MODEL:
		objType = drivers.CatalogObjectTypeModel
	case runtimev1.CatalogObject_TYPE_METRICS_VIEW:
		objType = drivers.CatalogObjectTypeMetricsView
	}

	migrator, ok := Migrators[string(objType)]
	return migrator, ok
}
