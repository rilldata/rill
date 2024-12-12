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

// ValidateMetricsView validates the dimensions and measures in the executor's metrics view.
func (e *Executor) ValidateMetricsView(ctx context.Context) (*ValidateMetricsViewResult, error) {
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

	// Check time dimension exists
	if mv.TimeDimension != "" {
		f, ok := cols[strings.ToLower(mv.TimeDimension)]
		if !ok {
			res.TimeDimensionErr = fmt.Errorf("timeseries %q is not a column in table %q", mv.TimeDimension, mv.Table)
		} else if f.Type.Code != runtimev1.Type_CODE_TIMESTAMP && f.Type.Code != runtimev1.Type_CODE_DATE {
			res.TimeDimensionErr = fmt.Errorf("timeseries %q is not a TIMESTAMP column", mv.TimeDimension)
		}
	}

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

	// Pinot does have any native support for time shift using time grain specifiers
	if e.olap.Dialect() == drivers.DialectPinot && (mv.FirstDayOfWeek > 1 || mv.FirstMonthOfYear > 1) {
		res.OtherErrs = append(res.OtherErrs, fmt.Errorf("time shift not supported for Pinot dialect, so FirstDayOfWeek and FirstMonthOfYear should be 1"))
	}

	// Check the default theme exists
	if mv.DefaultTheme != "" {
		ctrl, err := e.rt.Controller(ctx, e.instanceID)
		if err != nil {
			return nil, fmt.Errorf("could not get controller: %w", err)
		}

		_, err = ctrl.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindTheme, Name: mv.DefaultTheme}, false)
		if err != nil {
			if errors.Is(err, drivers.ErrNotFound) {
				res.OtherErrs = append(res.OtherErrs, fmt.Errorf("theme %q does not exist", mv.DefaultTheme))
			} else {
				return nil, fmt.Errorf("could not find theme %q: %w", mv.DefaultTheme, err)
			}
		}
	}

	// Validate the metrics view schema.
	if res.IsZero() { // All dimensions and measures need to be valid to compute the schema.
		err = e.validateSchema(ctx, res)
		if err != nil {
			res.OtherErrs = append(res.OtherErrs, fmt.Errorf("failed to validate metrics view schema: %w", err))
		}
	}

	return res, nil
}

// validateAllDimensionsAndMeasures validates all dimensions and measures with one query. It returns an error if any of the expressions are invalid.
func (e *Executor) validateAllDimensionsAndMeasures(ctx context.Context, t *drivers.Table, mv *runtimev1.MetricsViewSpec) error {
	dialect := e.olap.Dialect()
	var dimExprs []string
	var unnestClauses []string
	var groupIndexes []string
	for idx, d := range mv.Dimensions {
		dimExpr, unnestClause := dialect.DimensionSelect(t.Database, t.DatabaseSchema, t.Name, d)
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
func (e *Executor) validateIndividualDimensionsAndMeasures(ctx context.Context, t *drivers.Table, mv *runtimev1.MetricsViewSpec, cols map[string]*runtimev1.StructType_Field, res *ValidateMetricsViewResult) {
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

// validateDimension validates a metrics view dimension.
func (e *Executor) validateDimension(ctx context.Context, t *drivers.Table, d *runtimev1.MetricsViewSpec_DimensionV2, fields map[string]*runtimev1.StructType_Field) error {
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
	expr, unnestClause := dialect.DimensionSelect(t.Database, t.DatabaseSchema, t.Name, d)

	// Validate with a query if it's an expression
	err := e.olap.Exec(ctx, &drivers.Statement{
		Query:  fmt.Sprintf("SELECT %s FROM %s %s GROUP BY 1", expr, dialect.EscapeTable(t.Database, t.DatabaseSchema, t.Name), unnestClause),
		DryRun: true,
	})
	if err != nil {
		return fmt.Errorf("failed to validate expression for dimension %q: %w", d.Name, err)
	}
	return nil
}

// validateMeasure validates a metrics view measure.
func (e *Executor) validateMeasure(ctx context.Context, t *drivers.Table, m *runtimev1.MetricsViewSpec_MeasureV2) error {
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
	}

	return nil
}
