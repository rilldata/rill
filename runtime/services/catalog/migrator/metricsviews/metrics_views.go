package metricsviews

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/services/catalog/migrator"
)

func init() {
	migrator.Register(drivers.ObjectTypeMetricsView, &metricsViewMigrator{})
}

const (
	SourceNotSelected    = "metrics view source not selected"
	SourceNotFound       = "metrics view source not found"
	TimestampNotSelected = "metrics view timestamp not selected"
	TimestampNotFound    = "metrics view selected timestamp not found"
	MissingDimension     = "at least one dimension should be present"
	MissingMeasure       = "at least one measure should be present"
)

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
	return []string{catalog.GetMetricsView().Model}
}

func (m *metricsViewMigrator) Validate(ctx context.Context, olap drivers.OLAPStore, catalog *drivers.CatalogEntry) []*runtimev1.ReconcileError {
	mv := catalog.GetMetricsView()
	if mv.Model == "" {
		return migrator.CreateValidationError(catalog.Path, SourceNotSelected)
	}
	if mv.TimeDimension == "" {
		return migrator.CreateValidationError(catalog.Path, TimestampNotSelected)
	}
	model, err := olap.InformationSchema().Lookup(ctx, mv.Model)
	if err != nil {
		if errors.Is(err, drivers.ErrNotFound) {
			return migrator.CreateValidationError(catalog.Path, SourceNotFound)
		}
		return migrator.CreateValidationError(catalog.Path, err.Error())
	}

	fieldsMap := make(map[string]*runtimev1.StructType_Field)
	for _, field := range model.Schema.Fields {
		fieldsMap[field.Name] = field
	}

	if _, ok := fieldsMap[mv.TimeDimension]; !ok {
		return migrator.CreateValidationError(catalog.Path, TimestampNotFound)
	}

	var validationErrors []*runtimev1.ReconcileError

	for i, dimension := range mv.Dimensions {
		err := validateDimension(ctx, model, dimension)
		if err != nil {
			validationErrors = append(validationErrors, &runtimev1.ReconcileError{
				Code:         runtimev1.ReconcileError_CODE_VALIDATION,
				FilePath:     catalog.Path,
				Message:      err.Error(),
				PropertyPath: []string{"Dimensions", strconv.Itoa(i)},
			})
		}
	}
	if len(mv.Dimensions) == 0 {
		validationErrors = append(validationErrors, &runtimev1.ReconcileError{
			Code:     runtimev1.ReconcileError_CODE_VALIDATION,
			FilePath: catalog.Path,
			Message:  MissingDimension,
		})
	}

	for i, measure := range mv.Measures {
		err := validateMeasure(ctx, olap, model, measure)
		if err != nil {
			validationErrors = append(validationErrors, &runtimev1.ReconcileError{
				Code:         runtimev1.ReconcileError_CODE_VALIDATION,
				FilePath:     catalog.Path,
				Message:      err.Error(),
				PropertyPath: []string{"Measures", strconv.Itoa(i)},
			})
		}
	}
	if len(mv.Measures) == 0 {
		validationErrors = append(validationErrors, &runtimev1.ReconcileError{
			Code:     runtimev1.ReconcileError_CODE_VALIDATION,
			FilePath: catalog.Path,
			Message:  MissingMeasure,
		})
	}

	return validationErrors
}

func (m *metricsViewMigrator) IsEqual(ctx context.Context, cat1, cat2 *drivers.CatalogEntry) bool {
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
