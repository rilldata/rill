package metrics_views

import (
	"context"
	"errors"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/services/catalog/migrator"
)

func init() {
	migrator.Register(drivers.ObjectTypeMetricsView, &metricsViewMigrator{})
}

var SourceNotSelected = errors.New("metrics view source not selected")
var SourceNotFound = errors.New("metrics view source not found")
var TimestampNotSelected = errors.New("metrics view timestamp not selected")
var TimestampNotFound = errors.New("metrics view selected timestamp not found")

type metricsViewMigrator struct{}

func (m *metricsViewMigrator) Create(ctx context.Context, olap drivers.OLAPStore, repo drivers.RepoStore, catalogObj *drivers.CatalogEntry) error {
	return nil
}

func (m *metricsViewMigrator) Update(ctx context.Context, olap drivers.OLAPStore, repo drivers.RepoStore, catalogObj *drivers.CatalogEntry) error {
	return nil
}

func (m *metricsViewMigrator) Rename(ctx context.Context, olap drivers.OLAPStore, from string, catalogObj *drivers.CatalogEntry) error {
	return nil
}

func (m *metricsViewMigrator) Delete(ctx context.Context, olap drivers.OLAPStore, catalogObj *drivers.CatalogEntry) error {
	return nil
}

func (m *metricsViewMigrator) GetDependencies(ctx context.Context, olap drivers.OLAPStore, catalog *drivers.CatalogEntry) []string {
	return []string{catalog.GetMetricsView().From}
}

func (m *metricsViewMigrator) Validate(ctx context.Context, olap drivers.OLAPStore, catalog *drivers.CatalogEntry) error {
	mv := catalog.GetMetricsView()
	if mv.From == "" {
		return SourceNotSelected
	}
	if mv.TimeDimension == "" {
		return TimestampNotSelected
	}
	model, err := olap.InformationSchema().Lookup(ctx, mv.From)
	if err != nil {
		if err == drivers.ErrNotFound {
			return SourceNotFound
		}
		return err
	}

	fieldsMap := make(map[string]*runtimev1.StructType_Field)
	for _, field := range model.Schema.Fields {
		fieldsMap[field.Name] = field
	}

	if _, ok := fieldsMap[mv.TimeDimension]; !ok {
		return TimestampNotFound
	}

	for _, dimension := range mv.Dimensions {
		err := validateDimension(ctx, model, dimension)
		if err != nil {
			return err
		}
	}

	for _, measure := range mv.Measures {
		err := validateMeasure(ctx, olap, model, measure)
		if err != nil {
			return err
		}
	}

	// dimension and measure errors are not marked as error
	return nil
}

func (m *metricsViewMigrator) IsEqual(ctx context.Context, cat1 *drivers.CatalogEntry, cat2 *drivers.CatalogEntry) bool {
	// TODO: do we need a deep check here?
	return false
}

func (m *metricsViewMigrator) ExistsInOlap(ctx context.Context, olap drivers.OLAPStore, catalog *drivers.CatalogEntry) (bool, error) {
	return true, nil
}

func validateDimension(ctx context.Context, model *drivers.Table, dimension *runtimev1.MetricsView_Dimension) error {
	for _, field := range model.Schema.Fields {
		// TODO: check type
		if field.Name == dimension.Name {
			return nil
		}
	}

	return fmt.Errorf("dimension not found: %s", dimension.Name)
}

func validateMeasure(ctx context.Context, olap drivers.OLAPStore, model *drivers.Table, measure *runtimev1.MetricsView_Measure) error {
	_, err := olap.Execute(ctx, &drivers.Statement{
		Query:    fmt.Sprintf("SELECT %s from %s", measure.Expression, model.Name),
		DryRun:   true,
		Priority: 0,
	})
	return err
}
