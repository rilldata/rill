package metricsviews

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

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

func (m *metricsViewMigrator) Update(ctx context.Context, olap drivers.OLAPStore, repo drivers.RepoStore, oldCatalogObj, newCatalogObj *drivers.CatalogEntry) error {
	return nil
}

func (m *metricsViewMigrator) Rename(ctx context.Context, olap drivers.OLAPStore, from string, catalogObj *drivers.CatalogEntry) error {
	return nil
}

func (m *metricsViewMigrator) Delete(ctx context.Context, olap drivers.OLAPStore, catalogObj *drivers.CatalogEntry) error {
	return nil
}

func (m *metricsViewMigrator) GetDependencies(ctx context.Context, olap drivers.OLAPStore, catalog *drivers.CatalogEntry) ([]string, []*drivers.CatalogEntry) {
	return []string{catalog.GetMetricsView().Model}, nil
}

func (m *metricsViewMigrator) Validate(ctx context.Context, olap drivers.OLAPStore, catalog *drivers.CatalogEntry) []*runtimev1.ReconcileError {
	mv := catalog.GetMetricsView()
	if mv.Model == "" {
		return migrator.CreateValidationError(catalog.Path, SourceNotSelected)
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
		fieldsMap[strings.ToLower(field.Name)] = field
	}

	// if a time dimension is selected it should exist
	if mv.TimeDimension != "" {
		if _, ok := fieldsMap[strings.ToLower(mv.TimeDimension)]; !ok {
			return migrator.CreateValidationError(catalog.Path, TimestampNotFound)
		}
	}

	var validationErrors []*runtimev1.ReconcileError

	for i, dimension := range mv.Dimensions {
		if _, ok := fieldsMap[strings.ToLower(dimension.Name)]; !ok {
			validationErrors = append(validationErrors, &runtimev1.ReconcileError{
				Code:         runtimev1.ReconcileError_CODE_VALIDATION,
				FilePath:     catalog.Path,
				Message:      fmt.Sprintf("dimension not found: %s", dimension.Name),
				PropertyPath: []string{"Dimensions", strconv.Itoa(i)},
			})
		}
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
	// at least one measure has to be there in the metrics view
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

func validateMeasure(ctx context.Context, olap drivers.OLAPStore, model *drivers.Table, measure *runtimev1.MetricsView_Measure) error {
	err := olap.Exec(ctx, &drivers.Statement{
		Query:  fmt.Sprintf("SELECT %s from %s", measure.Expression, model.Name),
		DryRun: true,
	})
	return err
}
