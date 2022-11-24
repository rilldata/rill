package sources

import (
	"context"
	"fmt"
	"strings"

	"github.com/rilldata/rill/runtime/connectors"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/services/catalog/migrator"
	sql "github.com/rilldata/rill/runtime/sql/pure"
)

func init() {
	migrator.Register(drivers.ObjectTypeSource, &sourceMigrator{})
}

type sourceMigrator struct{}

func (m *sourceMigrator) Create(ctx context.Context, olap drivers.OLAPStore, repo drivers.RepoStore, catalogObj *drivers.CatalogEntry) error {
	apiSource := catalogObj.GetSource()

	source := &connectors.Source{
		Name:       apiSource.Name,
		Connector:  apiSource.Connector,
		Properties: apiSource.Properties.AsMap(),
	}

	env := &connectors.Env{
		RepoDriver: repo.Driver(),
		RepoDSN:    repo.DSN(),
	}

	return olap.Ingest(ctx, env, source)
}

func (m *sourceMigrator) Update(ctx context.Context, olap drivers.OLAPStore, repo drivers.RepoStore, catalogObj *drivers.CatalogEntry) error {
	return m.Create(ctx, olap, repo, catalogObj)
}

func (m *sourceMigrator) Rename(ctx context.Context, olap drivers.OLAPStore, from string, catalogObj *drivers.CatalogEntry) error {
	rows, err := olap.Execute(ctx, &drivers.Statement{
		Query:    fmt.Sprintf("ALTER TABLE %s RENAME TO %s", from, catalogObj.Name),
		Priority: 100,
	})
	if err != nil {
		return err
	}
	return rows.Close()
}

func (m *sourceMigrator) Delete(ctx context.Context, olap drivers.OLAPStore, catalogObj *drivers.CatalogEntry) error {
	rows, err := olap.Execute(ctx, &drivers.Statement{
		Query:    fmt.Sprintf("DROP TABLE IF EXISTS %s", catalogObj.Name),
		Priority: 100,
	})
	if err != nil {
		return err
	}
	return rows.Close()
}

func (m *sourceMigrator) GetDependencies(ctx context.Context, olap drivers.OLAPStore, catalog *drivers.CatalogEntry) []string {
	return []string{}
}

func (m *sourceMigrator) Validate(ctx context.Context, olap drivers.OLAPStore, catalog *drivers.CatalogEntry) error {
	// TODO
	return nil
}

func (m *sourceMigrator) IsEqual(ctx context.Context, cat1 *drivers.CatalogEntry, cat2 *drivers.CatalogEntry) bool {
	if cat1.GetSource().Connector != cat2.GetSource().Connector {
		return false
	}
	s1 := &connectors.Source{
		Properties: cat1.GetSource().Properties.AsMap(),
	}
	s2 := &connectors.Source{
		Properties: cat2.GetSource().Properties.AsMap(),
	}
	return s1.PropertiesEquals(s2)
}

func (m *sourceMigrator) ExistsInOlap(ctx context.Context, olap drivers.OLAPStore, catalog *drivers.CatalogEntry) (bool, error) {
	_, err := olap.InformationSchema().Lookup(ctx, catalog.Name)
	if err == drivers.ErrNotFound {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
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
