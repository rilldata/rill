package models

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/runtime/api"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/services/catalog/migrator"
)

func init() {
	migrator.Register(string(drivers.CatalogObjectTypeModel), &modelMigrator{})
}

type modelMigrator struct{}

func (m *modelMigrator) Create(ctx context.Context, olap drivers.OLAPStore, catalogObj *api.CatalogObject) error {
	rows, err := olap.Execute(ctx, &drivers.Statement{
		Query:    fmt.Sprintf("CREATE OR REPLACE TEMPORARY VIEW %s AS (%s)", catalogObj.Name, catalogObj.Model.Sql),
		Priority: 100,
	})
	if err != nil {
		return err
	}
	return rows.Close()
}

func (m *modelMigrator) Update(ctx context.Context, olap drivers.OLAPStore, catalogObj *api.CatalogObject) error {
	return m.Create(ctx, olap, catalogObj)
}

func (m *modelMigrator) Rename(ctx context.Context, olap drivers.OLAPStore, from string, catalogObj *api.CatalogObject) error {
	rows, err := olap.Execute(ctx, &drivers.Statement{
		Query:    fmt.Sprintf("ALTER VIEW %s RENAME TO %s", catalogObj.Name, from),
		Priority: 100,
	})
	if err != nil {
		return err
	}
	return rows.Close()
}

func (m *modelMigrator) Delete(ctx context.Context, olap drivers.OLAPStore, catalogObj *api.CatalogObject) error {
	rows, err := olap.Execute(ctx, &drivers.Statement{
		Query:    fmt.Sprintf("DROP VIEW IF EXISTS %s", catalogObj.Name),
		Priority: 100,
	})
	if err != nil {
		return err
	}
	return rows.Close()
}
