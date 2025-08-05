package metricsview

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"sync"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/fieldselectorpb"
	"golang.org/x/sync/errgroup"
)

const validateConcurrencyLimit = 10

// ValidateMetricsViewResult contains the results of validating a metrics view.
type ValidateMetricsViewResult struct {
	TimeDimensionErr error
	DimensionErrs    []IndexErr
	MeasureErrs      []IndexErr
	OtherErrs        []error
}

// IndexErr contains an error and the index of the dimension or measure that caused the error.
type IndexErr struct {
	Idx int
	Err error
}

// IsZero returns true if the result contains no errors.
func (r *ValidateMetricsViewResult) IsZero() bool {
	return r.TimeDimensionErr == nil && len(r.DimensionErrs) == 0 && len(r.MeasureErrs) == 0 && len(r.OtherErrs) == 0
}

// Error returns a single error containing all validation errors.
// If there are no errors, it returns nil.
func (r *ValidateMetricsViewResult) Error() error {
	var errs []error
	errs = append(errs, r.TimeDimensionErr)
	for _, e := range r.DimensionErrs {
		errs = append(errs, e.Err)
	}
	for _, e := range r.MeasureErrs {
		errs = append(errs, e.Err)
	}
	errs = append(errs, r.OtherErrs...)

	// NOTE: errors.Join returns nil if all input errs are nil.
	return errors.Join(errs...)
}

// ValidateAndNormalizeMetricsView validates the dimensions and measures in the executor's metrics view and returns a ValidateMetricsViewResult
// It also populates the schema of the metrics view if all dimensions and measures are valid.
// Note - Beware that it modifies the metrics view spec in place to populate the dimension and measure types.
func (e *Executor) ValidateAndNormalizeMetricsView(ctx context.Context) (*ValidateMetricsViewResult, error) {
	// Create the result
	res := &ValidateMetricsViewResult{}

	// Check underlying table exists
	mv := e.metricsView
	t, err := e.olap.InformationSchema().Lookup(ctx, mv.Database, mv.DatabaseSchema, mv.Table)
	if err != nil {
		if errors.Is(err, drivers.ErrNotFound) {
			res.OtherErrs = append(res.OtherErrs, fmt.Errorf("table %q does not exist", mv.Table))
			return res, nil
		}
		return nil, fmt.Errorf("could not find table %q: %w", mv.Table, err)
	}
	cols := make(map[string]*runtimev1.StructType_Field, len(t.Schema.Fields))
	for _, f := range t.Schema.Fields {
		cols[strings.ToLower(f.Name)] = f
	}

	// First check time dimension is valid type if exists
	e.validateTimeDimension(ctx, t, cols, res)

	// Check security policy rules apply to fields that exist
	fields := make(map[string]bool, len(mv.Dimensions)+len(mv.Measures))
	for _, d := range mv.Dimensions {
		fields[strings.ToLower(d.Name)] = true
	}
	for _, m := range mv.Measures {
		fields[strings.ToLower(m.Name)] = true
	}
	for _, rule := range mv.SecurityRules {
		fa := rule.GetFieldAccess()
		if fa == nil {
			continue
		}
		for _, f := range fa.Fields {
			if _, ok := fields[strings.ToLower(f)]; !ok {
				res.OtherErrs = append(res.OtherErrs, fmt.Errorf("field %q referenced in 'security' is not a dimension or measure", f))
			}
		}
	}

	// ClickHouse specifically does not support using a column name as a dimension or measure name if the dimension or measure has an expression.
	// This is due to ClickHouse's aggressive substitution of aliases: https://github.com/ClickHouse/ClickHouse/issues/9715.
	if e.olap.Dialect() == drivers.DialectClickHouse {
		for _, d := range mv.Dimensions {
			if d.Expression == "" && !d.Unnest {
				continue
			}
			if d.Expression == d.Name {
				// If the expression exactly matches the name, substitution is not a problem.
				continue
			}
			if _, ok := cols[strings.ToLower(d.Name)]; ok {
				res.OtherErrs = append(res.OtherErrs, fmt.Errorf("invalid dimension %q: dimensions that use `expression` or `unnest` cannot have the same name as a column in the underlying table when backed by clickhouse", d.Name))
			}
		}
		for _, m := range mv.Measures {
			if _, ok := cols[strings.ToLower(m.Name)]; ok {
				res.OtherErrs = append(res.OtherErrs, fmt.Errorf("invalid measure %q: measures cannot have the same name as a column in the underlying table when backed by clickhouse", m.Name))
			}
		}
	}

	// For performance, attempt to validate all dimensions and measures at once
	err = e.validateAllDimensionsAndMeasures(ctx, t, mv)
	if err != nil {
		// One or more dimension/measure expressions failed to validate. We need to check each one individually to provide useful errors.
		e.validateIndividualDimensionsAndMeasures(ctx, t, mv, cols, res)
	}

	// Check and rewrite annotations
	err = e.validateAndNormalizeAnnotations(ctx, mv, res)
	if err != nil {
		return res, err
	}

	// Pinot does have any native support for time shift using time grain specifiers
	if e.olap.Dialect() == drivers.DialectPinot && (mv.FirstDayOfWeek > 1 || mv.FirstMonthOfYear > 1) {
		res.OtherErrs = append(res.OtherErrs, fmt.Errorf("time shift not supported for Pinot dialect, so FirstDayOfWeek and FirstMonthOfYear should be 1"))
	}

	// Validate the metrics view schema.
	if res.IsZero() { // All dimensions and measures need to be valid to compute the schema.
		err = e.validateSchema(ctx, res)
		if err != nil {
			res.OtherErrs = append(res.OtherErrs, fmt.Errorf("failed to validate metrics view schema: %w", err))
		}
	}

	// Validate the cache key can be resolved
	_, _, err = e.CacheKey(ctx)
	if err != nil {
		res.OtherErrs = append(res.OtherErrs, fmt.Errorf("failed to get cache key: %w", err))
	}

	return res, nil
}

// validateAllDimensionsAndMeasures validates all dimensions and measures with one query. It returns an error if any of the expressions are invalid.
func (e *Executor) validateAllDimensionsAndMeasures(ctx context.Context, t *drivers.OlapTable, mv *runtimev1.MetricsViewSpec) error {
	dialect := e.olap.Dialect()
	var dimExprs []string
	var unnestClauses []string
	var groupIndexes []string
	for idx, d := range mv.Dimensions {
		dimExpr, unnestClause, err := dialect.DimensionSelect(t.Database, t.DatabaseSchema, t.Name, d)
		if err != nil {
			return fmt.Errorf("failed to validate dimension %q: %w", d.Name, err)
		}
		dimExprs = append(dimExprs, dimExpr)
		if unnestClause != "" {
			unnestClauses = append(unnestClauses, unnestClause)
		}
		groupIndexes = append(groupIndexes, strconv.Itoa(idx+1))
	}
	var metricExprs []string
	for _, m := range mv.Measures {
		if m.Type != runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE || m.Window != nil { // TODO: Validate advanced measures
			continue
		}
		metricExprs = append(metricExprs, "("+m.Expression+")")
	}
	var query string
	if len(dimExprs) == 0 && len(metricExprs) == 0 {
		// No metric and dimension, nothing to check
		return nil
	}
	if len(dimExprs) == 0 {
		// Only metrics
		query = fmt.Sprintf("SELECT 1, %s FROM %s GROUP BY 1", strings.Join(metricExprs, ","), e.olap.Dialect().EscapeTable(t.Database, t.DatabaseSchema, t.Name))
	} else if len(metricExprs) == 0 {
		// No metrics
		query = fmt.Sprintf(
			"SELECT %s FROM %s %s GROUP BY %s",
			strings.Join(dimExprs, ","),
			e.olap.Dialect().EscapeTable(t.Database, t.DatabaseSchema, t.Name),
			strings.Join(unnestClauses, ""),
			strings.Join(groupIndexes, ","),
		)
	} else {
		query = fmt.Sprintf(
			"SELECT %s, %s FROM %s %s GROUP BY %s",
			strings.Join(dimExprs, ","),
			strings.Join(metricExprs, ","),
			e.olap.Dialect().EscapeTable(t.Database, t.DatabaseSchema, t.Name),
			strings.Join(unnestClauses, ""),
			strings.Join(groupIndexes, ","),
		)
	}
	err := e.olap.Exec(ctx, &drivers.Statement{
		Query:  query,
		DryRun: true,
	})
	if err != nil {
		return fmt.Errorf("failed to validate dims and metrics: %w", err)
	}
	return nil
}

// validateIndividualDimensionsAndMeasures validates each dimension and measure individually.
// It adds validation errors to the provided res.
func (e *Executor) validateIndividualDimensionsAndMeasures(ctx context.Context, t *drivers.OlapTable, mv *runtimev1.MetricsViewSpec, cols map[string]*runtimev1.StructType_Field, res *ValidateMetricsViewResult) {
	// Validate dimensions and measures concurrently with a limit of 10 concurrent validations
	var mu sync.Mutex
	var grp errgroup.Group
	grp.SetLimit(validateConcurrencyLimit)

	// Check dimension expressions are valid
	for idx, d := range mv.Dimensions {
		idx := idx
		d := d
		grp.Go(func() error {
			err := e.validateDimension(ctx, t, d, cols)
			if err != nil {
				mu.Lock()
				defer mu.Unlock()

				res.DimensionErrs = append(res.DimensionErrs, IndexErr{
					Idx: idx,
					Err: err,
				})
			}
			return nil
		})
	}

	// Check measure expressions are valid
	for idx, m := range mv.Measures {
		if m.Type != runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE || m.Window != nil { // TODO: Validate advanced measures
			continue
		}

		idx := idx
		m := m
		grp.Go(func() error {
			err := e.validateMeasure(ctx, t, m)
			if err != nil {
				mu.Lock()
				defer mu.Unlock()

				res.MeasureErrs = append(res.MeasureErrs, IndexErr{
					Idx: idx,
					Err: fmt.Errorf("invalid expression for measure %q: %w", m.Name, err),
				})
			}
			return nil
		})
	}

	// Wait for all validations to complete
	_ = grp.Wait()

	// Sort errors by index (for stable output)
	slices.SortFunc(res.DimensionErrs, func(a, b IndexErr) int { return a.Idx - b.Idx })
	slices.SortFunc(res.MeasureErrs, func(a, b IndexErr) int { return a.Idx - b.Idx })
}

// validateAndNormalizeAnnotations validates the annotations by checking the model/table defined with expected columns.
// Rewrites the annotations to use the resolved table name from the defined model.
// Resolves the measure selector and stores the resolved measures in the annotation.
func (e *Executor) validateAndNormalizeAnnotations(ctx context.Context, mv *runtimev1.MetricsViewSpec, res *ValidateMetricsViewResult) error {
	allMeasures := make([]string, 0, len(mv.Measures))
	for _, m := range mv.Measures {
		allMeasures = append(allMeasures, m.Name)
	}

	// Get the controller used for getting the annotation's model
	ct, err := e.rt.Controller(ctx, e.instanceID)
	if err != nil {
		return fmt.Errorf("failed to get controller: %w", err)
	}

	// Different models could be in the same or different connector. Maintain a map to reuse connections.
	olaps := make(map[string]drivers.OLAPStore)
	olapReleases := make([]func(), 0)
	for _, annotation := range mv.Annotations {
		// Resolve the measures selector
		annotation.Measures, err = fieldselectorpb.ResolveFields(annotation.Measures, annotation.MeasuresSelector, allMeasures)
		if err != nil {
			res.OtherErrs = append(res.OtherErrs, fmt.Errorf("invalid measures for annotation %q: %w", annotation.Name, err))
		}
		annotation.MeasuresSelector = nil

		if annotation.Model != "" {
			res, err := ct.Get(ctx, &runtimev1.ResourceName{Name: annotation.Model, Kind: runtime.ResourceKindModel}, false)
			if err == nil && res.GetModel().State.ResultTable != "" {
				annotation.Table = res.GetModel().State.ResultTable
				annotation.Connector = res.GetModel().State.ResultConnector
			} else {
				annotation.Table = annotation.Model
			}
		}

		// Get the connector for the model either from the map or acquire a new one
		olap, ok := olaps[annotation.Connector]
		if !ok {
			var release func()
			olap, release, err = e.rt.OLAP(ctx, e.instanceID, annotation.Connector)
			if err != nil {
				res.OtherErrs = append(res.OtherErrs, fmt.Errorf("failed to acquire connection to table %q for annotation %q: %w", annotation.Table, annotation.Name, err))
				break // other connections might fail as well
			}
			olapReleases = append(olapReleases, release)
		}

		// Get the table schema
		tableSchema, err := olap.InformationSchema().Lookup(ctx, annotation.Database, annotation.DatabaseSchema, annotation.Table)
		if err != nil {
			res.OtherErrs = append(res.OtherErrs, fmt.Errorf("failed to get table details %q for annotation %q: %w", annotation.Table, annotation.Name, err))
			continue
		}

		// Validate the table for required columns and save metadata about optional columns. This metadata will be used during querying the table.
		var hasTime, hasDesc bool
		for _, field := range tableSchema.Schema.Fields {
			switch field.Name {
			case "time":
				hasTime = true

			case "time_end":
				annotation.HasTimeEnd = true

			case "grain":
				annotation.HasGrain = true

			case "description":
				hasDesc = true
			}
		}

		if !hasTime {
			res.OtherErrs = append(res.OtherErrs, fmt.Errorf(`table %q for annotation %q does not have the required "time" column`, annotation.Table, annotation.Name))
		}
		if !hasDesc {
			res.OtherErrs = append(res.OtherErrs, fmt.Errorf(`table %q for annotation %q does not have the required "description" column`, annotation.Table, annotation.Name))
		}
	}

	for _, release := range olapReleases {
		release()
	}

	return nil
}

// validateTimeDimension validates the time dimension in the metrics view.
func (e *Executor) validateTimeDimension(ctx context.Context, t *drivers.OlapTable, tableSchema map[string]*runtimev1.StructType_Field, res *ValidateMetricsViewResult) {
	if e.metricsView.TimeDimension == "" {
		return
	}

	// Time dimension should either be defined in the metrics view or exist in the table schema if referring to a model column directly
	for _, d := range e.metricsView.Dimensions {
		if !strings.EqualFold(d.Name, e.metricsView.TimeDimension) {
			continue
		}

		dialect := e.olap.Dialect()
		expr, err := dialect.MetricsViewDimensionExpression(d)
		if err != nil {
			res.TimeDimensionErr = fmt.Errorf("failed to validate time dimension %q: %w", e.metricsView.TimeDimension, err)
			return
		}

		query := fmt.Sprintf("SELECT %s FROM %s LIMIT 0", expr, dialect.EscapeTable(t.Database, t.DatabaseSchema, t.Name))
		schema, err := e.olap.QuerySchema(ctx, query, nil)
		if err != nil {
			res.TimeDimensionErr = fmt.Errorf("failed to validate time dimension %q: %w", e.metricsView.TimeDimension, err)
			return
		}
		if len(schema.Fields) == 0 {
			res.TimeDimensionErr = fmt.Errorf("time dimension %q is not a column in table %q or defined in metrics view", e.metricsView.TimeDimension, e.metricsView.Table)
			return
		}
		typeCode := schema.Fields[0].Type.Code

		if typeCode != runtimev1.Type_CODE_TIMESTAMP && typeCode != runtimev1.Type_CODE_DATE && !(e.olap.Dialect() == drivers.DialectPinot && typeCode == runtimev1.Type_CODE_INT64) {
			res.TimeDimensionErr = fmt.Errorf("time dimension %q is not a TIMESTAMP column, got %s", e.metricsView.TimeDimension, typeCode)
		}
		return
	}

	// If the time dimension is not defined in the metrics view dimensions, check if it exists in the table schema
	f, ok := tableSchema[strings.ToLower(e.metricsView.TimeDimension)]
	if !ok {
		res.TimeDimensionErr = fmt.Errorf("timeseries %q is not a column in table %q or defined in metrics view", e.metricsView.TimeDimension, e.metricsView.Table)
		return
	} else if f.Type.Code != runtimev1.Type_CODE_TIMESTAMP && f.Type.Code != runtimev1.Type_CODE_DATE && !(e.olap.Dialect() == drivers.DialectPinot && f.Type.Code == runtimev1.Type_CODE_INT64) {
		res.TimeDimensionErr = fmt.Errorf("time dimension %q is not a TIMESTAMP column, got %s", e.metricsView.TimeDimension, f.Type.Code)
		return
	}
}

// validateDimension validates a metrics view dimension.
func (e *Executor) validateDimension(ctx context.Context, t *drivers.OlapTable, d *runtimev1.MetricsViewSpec_Dimension, fields map[string]*runtimev1.StructType_Field) error {
	// Validate with a simple check if it's a column
	if d.Column != "" {
		if _, isColumn := fields[strings.ToLower(d.Column)]; !isColumn {
			return fmt.Errorf("failed to validate dimension %q: column %q not found in table", d.Name, d.Column)
		}
		if !d.Unnest {
			// for dimensions that have column and no unnest skip the expr validation since the above validation is enough
			return nil
		}
	}

	dialect := e.olap.Dialect()
	expr, unnestClause, err := dialect.DimensionSelect(t.Database, t.DatabaseSchema, t.Name, d)
	if err != nil {
		return fmt.Errorf("failed to validate dimension %q: %w", d.Name, err)
	}

	// Validate with a query if it's an expression
	err = e.olap.Exec(ctx, &drivers.Statement{
		Query:  fmt.Sprintf("SELECT %s FROM %s %s GROUP BY 1", expr, dialect.EscapeTable(t.Database, t.DatabaseSchema, t.Name), unnestClause),
		DryRun: true,
	})
	if err != nil {
		return fmt.Errorf("failed to validate expression for dimension %q: %w", d.Name, err)
	}
	return nil
}

// validateMeasure validates a metrics view measure.
func (e *Executor) validateMeasure(ctx context.Context, t *drivers.OlapTable, m *runtimev1.MetricsViewSpec_Measure) error {
	err := e.olap.Exec(ctx, &drivers.Statement{
		Query:  fmt.Sprintf("SELECT 1, (%s) FROM %s GROUP BY 1", m.Expression, e.olap.Dialect().EscapeTable(t.Database, t.DatabaseSchema, t.Name)),
		DryRun: true,
	})
	return err
}

// validateSchema validates that the metrics view's measures are numeric.
func (e *Executor) validateSchema(ctx context.Context, res *ValidateMetricsViewResult) error {
	// Resolve the schema of the metrics view's dimensions and measures
	schema, err := e.Schema(ctx)
	if err != nil {
		return err
	}
	types := make(map[string]*runtimev1.Type, len(schema.Fields))
	for _, f := range schema.Fields {
		types[f.Name] = f.Type
	}

	// Check that the measures are not strings
	for i, m := range e.metricsView.Measures {
		typ, ok := types[m.Name]
		if !ok {
			// Don't error: schemas are not always reliable
			continue
		}

		switch typ.Code {
		case runtimev1.Type_CODE_TIMESTAMP, runtimev1.Type_CODE_DATE, runtimev1.Type_CODE_TIME, runtimev1.Type_CODE_STRING, runtimev1.Type_CODE_BYTES, runtimev1.Type_CODE_ARRAY, runtimev1.Type_CODE_STRUCT, runtimev1.Type_CODE_MAP, runtimev1.Type_CODE_JSON, runtimev1.Type_CODE_UUID:
			res.MeasureErrs = append(res.MeasureErrs, IndexErr{
				Idx: i,
				Err: fmt.Errorf("measure %q is of type %s, but must be a numeric type", m.Name, typ.Code),
			})
		}
		m.DataType = typ
	}

	for _, d := range e.metricsView.Dimensions {
		if typ, ok := types[d.Name]; ok {
			d.DataType = typ
		} // ignore dimensions that don't have a type in the schema
	}

	return nil
}
