package sources

import (
	"context"
	"fmt"
	"strings"

	"github.com/rilldata/rill/runtime/api"
	"github.com/rilldata/rill/runtime/connectors"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/services/catalog/migrator"
	sql "github.com/rilldata/rill/runtime/sql/pure"
)

func init() {
	migrator.Register(drivers.CatalogObjectTypeSource, &sourceMigrator{})
}

type sourceMigrator struct{}

func (m *sourceMigrator) Create(ctx context.Context, olap drivers.OLAPStore, catalogObj *api.CatalogObject) error {
	apiSource := catalogObj.Type.(*api.CatalogObject_Source).Source
	var source *connectors.Source
	var err error
	if apiSource.Sql != "" {
		source, err = SqlToSource(apiSource.Sql)
		if err != nil {
			return err
		}
	} else {
		source = &connectors.Source{
			Name:       apiSource.Name,
			Connector:  apiSource.Connector,
			Properties: apiSource.Properties.AsMap(),
		}
	}
	return olap.Ingest(ctx, source)
}

func (m *sourceMigrator) Update(ctx context.Context, olap drivers.OLAPStore, catalogObj *api.CatalogObject) error {
	return m.Create(ctx, olap, catalogObj)
}

func (m *sourceMigrator) Rename(ctx context.Context, olap drivers.OLAPStore, from string, catalogObj *api.CatalogObject) error {
	rows, err := olap.Execute(ctx, &drivers.Statement{
		Query:    fmt.Sprintf("ALTER TABLE %s RENAME TO %s", catalogObj.Name, from),
		Priority: 100,
	})
	if err != nil {
		return err
	}
	return rows.Close()
}

func (m *sourceMigrator) Delete(ctx context.Context, olap drivers.OLAPStore, catalogObj *api.CatalogObject) error {
	rows, err := olap.Execute(ctx, &drivers.Statement{
		Query:    fmt.Sprintf("DROP TABLE IF EXISTS %s", catalogObj.Name),
		Priority: 100,
	})
	if err != nil {
		return err
	}
	return rows.Close()
}

func SqlToSource(sqlStr string) (*connectors.Source, error) {
	astStmt, err := sql.Parse(sqlStr)
	if err != nil {
		return nil, fmt.Errorf("parse error: %s", err.Error())
	}

	if astStmt.CreateSource == nil {
		return nil, fmt.Errorf("refresh error: object cannot be refreshed")
	}

	ast := astStmt.CreateSource

	s := &connectors.Source{
		Name:       ast.Name,
		Properties: make(map[string]any),
	}

	for _, prop := range ast.With.Properties {
		if strings.ToLower(prop.Key) == "connector" {
			s.Connector = safePtrToStr(prop.Value.String)
			continue
		}
		if prop.Value.Number != nil {
			s.Properties[prop.Key] = *prop.Value.Number
		} else if prop.Value.String != nil {
			s.Properties[prop.Key] = *prop.Value.String
		} else if prop.Value.Boolean != nil {
			s.Properties[prop.Key] = *prop.Value.Boolean
		}
	}

	err = s.Validate()
	if err != nil {
		return nil, err
	}

	return s, nil
}

func safePtrToStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
