package models

import (
	"context"
	"errors"
	"fmt"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/services/catalog/migrator"
	"github.com/rilldata/rill/runtime/services/catalog/migrator/sources"
)

func init() {
	migrator.Register(drivers.ObjectTypeModel, &modelMigrator{})
}

type modelMigrator struct{}

func (m *modelMigrator) Create(ctx context.Context, olap drivers.OLAPStore, repo drivers.RepoStore, catalogObj *drivers.CatalogEntry) error {
	sql := catalogObj.GetModel().Sql
	materialize := catalogObj.GetModel().Materialize
	materializeType, err := getMaterializeType(materialize)
	if err != nil {
		return err
	}
	return olap.Exec(ctx, &drivers.Statement{
		Query: fmt.Sprintf(
			"CREATE OR REPLACE %s %s AS (%s)",
			materializeType,
			catalogObj.Name,
			sql,
		),
		Priority: 100,
	})
}

func (m *modelMigrator) Update(ctx context.Context, olap drivers.OLAPStore, repo drivers.RepoStore, oldCatalogObj, newCatalogObj *drivers.CatalogEntry) error {
	if oldCatalogObj.Name != newCatalogObj.Name {
		// should not happen but just to be sure
		return errors.New("update is called but model name has changed")
	}
	oldModel := oldCatalogObj.GetModel()
	newModel := newCatalogObj.GetModel()
	// currently we are treating Model_MATERIALIZE_TRUE and Model_MATERIALIZE_INFERRED the same as both are TABLEs
	// and Model_MATERIALIZE_FALSE and Model_MATERIALIZE_UNSPECIFIED the same as both are VIEWs
	oldMaterializeType, err := getMaterializeType(oldModel.Materialize)
	if err != nil {
		return err
	}
	newMaterializeType, err := getMaterializeType(newModel.Materialize)
	if err != nil {
		return err
	}
	// check if sql and materialize type are same and if so, do nothing
	// this includes the cases where materialize is changed from true to inferred or false to unspecified and vice versa
	if oldModel.Sql == newModel.Sql && oldMaterializeType == newMaterializeType {
		return nil
	}
	// if sql is changed and materialize type is the same then just update the sql
	if oldModel.Sql != newModel.Sql && oldMaterializeType == newMaterializeType {
		return m.Create(ctx, olap, repo, newCatalogObj)
	}
	// else drop the old type and create new materialized type using new sql
	err = m.Delete(ctx, olap, oldCatalogObj)
	if err != nil {
		return err
	}
	return m.Create(ctx, olap, repo, newCatalogObj)
}

func getMaterializeType(materialize runtimev1.Model_Materialize) (string, error) {
	materializeType := ""
	switch materialize {
	case runtimev1.Model_MATERIALIZE_TRUE:
		materializeType = "TABLE"
	case runtimev1.Model_MATERIALIZE_INFERRED:
		materializeType = "TABLE"
	case runtimev1.Model_MATERIALIZE_FALSE:
		materializeType = "VIEW"
	case runtimev1.Model_MATERIALIZE_UNSPECIFIED:
		materializeType = "VIEW"
	default:
		return materializeType, errors.New("invalid materialize type")
	}
	return materializeType, nil
}

func (m *modelMigrator) Rename(ctx context.Context, olap drivers.OLAPStore, from string, catalogObj *drivers.CatalogEntry) error {
	materializeType, err := getMaterializeType(catalogObj.GetModel().Materialize)
	if err != nil {
		return err
	}
	if strings.EqualFold(from, catalogObj.Name) {
		tempName := fmt.Sprintf("__rill_temp_%s", from)
		err := olap.Exec(ctx, &drivers.Statement{
			Query:    fmt.Sprintf("ALTER %s %s RENAME TO %s", materializeType, from, tempName),
			Priority: 100,
		})
		if err != nil {
			return err
		}
		from = tempName
	}

	return olap.Exec(ctx, &drivers.Statement{
		Query:    fmt.Sprintf("ALTER %s %s RENAME TO %s", materializeType, from, catalogObj.Name),
		Priority: 100,
	})
}

func (m *modelMigrator) Delete(ctx context.Context, olap drivers.OLAPStore, catalogObj *drivers.CatalogEntry) error {
	materializeType, err := getMaterializeType(catalogObj.GetModel().Materialize)
	if err != nil {
		return err
	}
	return olap.Exec(ctx, &drivers.Statement{
		Query:    fmt.Sprintf("DROP %s IF EXISTS %s", materializeType, catalogObj.Name),
		Priority: 100,
	})
}

func (m *modelMigrator) GetDependencies(ctx context.Context, olap drivers.OLAPStore, catalog *drivers.CatalogEntry) ([]string, []*drivers.CatalogEntry) {
	model := catalog.GetModel()
	dependencies := ExtractTableNames(model.Sql)

	embeddedSourcesMap := make(map[string]*drivers.CatalogEntry)
	for i, dependency := range dependencies {
		source, ok := sources.ParseEmbeddedSource(dependency)
		if !ok {
			continue
		}
		if _, ok := embeddedSourcesMap[source.Name]; ok {
			continue
		}

		embeddedSourcesMap[source.Name] = &drivers.CatalogEntry{
			Name:     source.Name,
			Type:     drivers.ObjectTypeSource,
			Object:   source,
			Path:     source.Properties.AsMap()["path"].(string),
			Embedded: true,
		}

		// replace the dependency
		dependencies[i] = source.Name
		model.Sql = strings.ReplaceAll(model.Sql, dependency, source.Name)
	}

	embeddedSources := make([]*drivers.CatalogEntry, 0)
	for _, embeddedSource := range embeddedSourcesMap {
		embeddedSources = append(embeddedSources, embeddedSource)
	}
	return dependencies, embeddedSources
}

func (m *modelMigrator) Validate(ctx context.Context, olap drivers.OLAPStore, catalog *drivers.CatalogEntry) []*runtimev1.ReconcileError {
	model := catalog.GetModel()
	if model.Materialize == runtimev1.Model_MATERIALIZE_INVALID {
		return migrator.CreateValidationError(catalog.Path, "Invalid materialize type")
	}
	err := olap.Exec(ctx, &drivers.Statement{
		Query:    model.Sql,
		Priority: 100,
		DryRun:   true,
	})
	if err != nil {
		return migrator.CreateValidationError(catalog.Path, err.Error())
	}
	return nil
}

func (m *modelMigrator) IsEqual(ctx context.Context, cat1, cat2 *drivers.CatalogEntry) bool {
	return cat1.GetModel().Dialect == cat2.GetModel().Dialect && strings.EqualFold(cat1.GetModel().Sql, cat2.GetModel().Sql) && cat1.GetModel().Materialize == cat2.GetModel().Materialize
}

func (m *modelMigrator) ExistsInOlap(ctx context.Context, olap drivers.OLAPStore, catalog *drivers.CatalogEntry) (bool, error) {
	_, err := olap.InformationSchema().Lookup(ctx, catalog.Name)
	if errors.Is(err, drivers.ErrNotFound) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}
