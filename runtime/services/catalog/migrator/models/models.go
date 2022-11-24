package models

import (
	"context"
	"fmt"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/services/catalog/migrator"
)

func init() {
	migrator.Register(string(drivers.CatalogObjectTypeModel), &modelMigrator{})
}

type modelMigrator struct{}

func (m *modelMigrator) Create(ctx context.Context, olap drivers.OLAPStore, repo drivers.RepoStore, catalogObj *runtimev1.CatalogObject) error {
	rows, err := olap.Execute(ctx, &drivers.Statement{
		Query:    fmt.Sprintf("CREATE OR REPLACE VIEW %s AS (%s)", catalogObj.Name, catalogObj.Model.Sql),
		Priority: 100,
	})
	if err != nil {
		return err
	}
	return rows.Close()
}

func (m *modelMigrator) Update(ctx context.Context, olap drivers.OLAPStore, repo drivers.RepoStore, catalogObj *runtimev1.CatalogObject) error {
	return m.Create(ctx, olap, repo, catalogObj)
}

func (m *modelMigrator) Rename(ctx context.Context, olap drivers.OLAPStore, from string, catalogObj *runtimev1.CatalogObject) error {
	rows, err := olap.Execute(ctx, &drivers.Statement{
		Query:    fmt.Sprintf("ALTER VIEW %s RENAME TO %s", from, catalogObj.Name),
		Priority: 100,
	})
	if err != nil {
		return err
	}
	return rows.Close()
}

func (m *modelMigrator) Delete(ctx context.Context, olap drivers.OLAPStore, catalogObj *runtimev1.CatalogObject) error {
	rows, err := olap.Execute(ctx, &drivers.Statement{
		Query:    fmt.Sprintf("DROP VIEW IF EXISTS %s", catalogObj.Name),
		Priority: 100,
	})
	if err != nil {
		return err
	}
	return rows.Close()
}

func (m *modelMigrator) GetDependencies(ctx context.Context, olap drivers.OLAPStore, catalog *runtimev1.CatalogObject) []string {
	return ExtractTableNames(catalog.Model.Sql)
}

func (m *modelMigrator) Validate(ctx context.Context, olap drivers.OLAPStore, catalog *runtimev1.CatalogObject) error {
	_, err := olap.Execute(ctx, &drivers.Statement{
		Query:    catalog.Model.Sql,
		Priority: 100,
		DryRun:   true,
	})
	return err
}

func (m *modelMigrator) IsEqual(ctx context.Context, cat1 *runtimev1.CatalogObject, cat2 *runtimev1.CatalogObject) bool {
	return cat1.Model.Dialect == cat2.Model.Dialect &&
		// TODO: handle same queries but different text
		strings.TrimSpace(cat1.Model.Sql) == strings.TrimSpace(cat2.Model.Sql)
}

func (m *modelMigrator) ExistsInOlap(ctx context.Context, olap drivers.OLAPStore, catalog *runtimev1.CatalogObject) (bool, error) {
	_, err := olap.InformationSchema().Lookup(ctx, catalog.Name)
	if err == drivers.ErrNotFound {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}
