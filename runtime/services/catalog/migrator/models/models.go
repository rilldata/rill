package models

import (
	"context"
	"fmt"
	"strings"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/services/catalog/migrator"
)

func init() {
	migrator.Register(drivers.ObjectTypeModel, &modelMigrator{})
}

type modelMigrator struct{}

func (m *modelMigrator) Create(ctx context.Context, olap drivers.OLAPStore, repo drivers.RepoStore, catalogObj *drivers.CatalogEntry) error {
	rows, err := olap.Execute(ctx, &drivers.Statement{
		Query:    fmt.Sprintf("CREATE OR REPLACE VIEW %s AS (%s)", catalogObj.Name, catalogObj.GetModel().Sql),
		Priority: 100,
	})
	if err != nil {
		return err
	}
	return rows.Close()
}

func (m *modelMigrator) Update(ctx context.Context, olap drivers.OLAPStore, repo drivers.RepoStore, catalogObj *drivers.CatalogEntry) error {
	return m.Create(ctx, olap, repo, catalogObj)
}

func (m *modelMigrator) Rename(ctx context.Context, olap drivers.OLAPStore, from string, catalogObj *drivers.CatalogEntry) error {
	rows, err := olap.Execute(ctx, &drivers.Statement{
		Query:    fmt.Sprintf("ALTER VIEW %s RENAME TO %s", from, catalogObj.Name),
		Priority: 100,
	})
	if err != nil {
		return err
	}
	return rows.Close()
}

func (m *modelMigrator) Delete(ctx context.Context, olap drivers.OLAPStore, catalogObj *drivers.CatalogEntry) error {
	rows, err := olap.Execute(ctx, &drivers.Statement{
		Query:    fmt.Sprintf("DROP VIEW IF EXISTS %s", catalogObj.Name),
		Priority: 100,
	})
	if err != nil {
		return err
	}
	return rows.Close()
}

func (m *modelMigrator) GetDependencies(ctx context.Context, olap drivers.OLAPStore, catalog *drivers.CatalogEntry) []string {
	return ExtractTableNames(catalog.GetModel().Sql)
}

func (m *modelMigrator) Validate(ctx context.Context, olap drivers.OLAPStore, catalog *drivers.CatalogEntry) error {
	_, err := olap.Execute(ctx, &drivers.Statement{
		Query:    catalog.GetModel().Sql,
		Priority: 100,
		DryRun:   true,
	})
	return err
}

func (m *modelMigrator) IsEqual(ctx context.Context, cat1 *drivers.CatalogEntry, cat2 *drivers.CatalogEntry) bool {
	return cat1.GetModel().Dialect == cat2.GetModel().Dialect &&
		// TODO: handle same queries but different text
		strings.TrimSpace(cat1.GetModel().Sql) == strings.TrimSpace(cat2.GetModel().Sql)
}

func (m *modelMigrator) ExistsInOlap(ctx context.Context, olap drivers.OLAPStore, catalog *drivers.CatalogEntry) (bool, error) {
	_, err := olap.InformationSchema().Lookup(ctx, catalog.Name)
	if err == drivers.ErrNotFound {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}
