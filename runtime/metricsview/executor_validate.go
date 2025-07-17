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

// ReconcileWithParentMetricsView resolves the parent metrics view and inherits all its dimensions and measures unless they are overridden in the current metrics view.
func (e *Executor) ReconcileWithParentMetricsView(ctx context.Context) (*runtimev1.MetricsViewSpec, error) {
	if e.metricsView.Parent == "" {
		// No parent metrics view to normalize
		return nil, nil
	}
	// Resolve the parent metrics view
	ctrl, err := e.rt.Controller(ctx, e.instanceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get controller: %w", err)
	}
	// deep copy of parent metrics view that will be modified
	parent, err := ctrl.Get(ctx, &runtimev1.ResourceName{
		Name: e.metricsView.Parent,
		Kind: runtime.ResourceKindMetricsView,
	}, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get parent metrics view %q: %w", e.metricsView.Parent, err)
	}
	if parent.GetMetricsView() == nil {
		return nil, fmt.Errorf("parent resource %q is not a metrics view", e.metricsView.Parent)
	}
	newSpec := parent.GetMetricsView().State.ValidSpec
	newSpec.Parent = e.metricsView.Parent

	// Override the dimensions and measures in the normalized metrics view if defined in the current metrics view.
	allDims := make(map[string]*runtimev1.MetricsViewSpec_Dimension, len(newSpec.Dimensions))
	all := make([]string, 0, len(newSpec.Dimensions))
	for _, d := range newSpec.Dimensions {
		allDims[d.Name] = d
		all = append(all, d.Name)
	}

	var dimNames []string
	dimSelector := e.metricsView.DimensionsSelector
	if dimSelector != nil && !dimSelector.Invert && dimSelector.GetFields() != nil && len(dimSelector.GetFields().Values) > 0 {
		dimNames = dimSelector.GetFields().Values
	} else {
		dimNames, err = e.resolveFields(dimSelector, all)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve dimensions selector: %w", err)
		}
	}

	// resolve dim names to actual dimension spec
	for _, dimName := range dimNames {
		dim, ok := allDims[dimName]
		if !ok {
			return nil, fmt.Errorf("dimension %q not found in parent metrics view %q", dimName, e.metricsView.Parent)
		}
		e.metricsView.Dimensions = append(e.metricsView.Dimensions, dim)
	}
	newSpec.Dimensions = e.metricsView.Dimensions

	allMeasures := make(map[string]*runtimev1.MetricsViewSpec_Measure, len(newSpec.Measures))
	all = make([]string, 0, len(newSpec.Measures))
	for _, m := range newSpec.Measures {
		allMeasures[m.Name] = m
		all = append(all, m.Name)
	}
	var measureNames []string
	measureSelector := e.metricsView.MeasuresSelector
	if measureSelector != nil && !measureSelector.Invert && measureSelector.GetFields() != nil && len(measureSelector.GetFields().Values) > 0 {
		measureNames = measureSelector.GetFields().Values
	} else {
		measureNames, err = e.resolveFields(measureSelector, all)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve measures selector: %w", err)
		}
	}

	// resolve measure names to actual measure spec
	for _, measureName := range measureNames {
		measure, ok := allMeasures[measureName]
		if !ok {
			return nil, fmt.Errorf("measure %q not found in parent metrics view %q", measureName, e.metricsView.Parent)
		}
		e.metricsView.Measures = append(e.metricsView.Measures, measure)
	}
	newSpec.Measures = e.metricsView.Measures

	securityRules := make([]*runtimev1.SecurityRule, 0)
	var access, fieldAccess, rowFilter, parentAccess, parentRowFilter []*runtimev1.SecurityRule
	for _, rule := range newSpec.SecurityRules {
		if rule.GetAccess() != nil {
			parentAccess = append(parentAccess, rule)
		} else if rule.GetRowFilter() != nil {
			parentRowFilter = append(parentRowFilter, rule)
		}
	}

	for _, rule := range e.metricsView.SecurityRules {
		if rule.GetAccess() != nil {
			access = append(access, rule)
		} else if rule.GetFieldAccess() != nil {
			fieldAccess = append(fieldAccess, rule)
		} else if rule.GetRowFilter() != nil {
			if len(parentRowFilter) > 1 || len(rowFilter) > 1 {
				return nil, fmt.Errorf("unable to merge multiple row filters into one")
			}
			rowFilter = append(rowFilter, rule)
		}
	}

	if len(access) > 0 {
		securityRules = append(securityRules, access...)
	} else if len(parentAccess) > 0 {
		securityRules = append(securityRules, parentAccess...)
	}

	if len(fieldAccess) > 0 {
		securityRules = append(securityRules, fieldAccess...)
	} // field access cannot be inherited from parent metrics view, so we ignore parentFieldAccess

	if len(rowFilter) > 0 {
		// If the metrics view has a row filter, we need to AND the row filter with parent row filter
		if len(parentRowFilter) > 0 {
			rowFilter[0].GetRowFilter().Sql = fmt.Sprintf("(%s) AND (%s)", parentRowFilter[0].GetRowFilter().Sql, rowFilter[0].GetRowFilter().Sql)
		}
		securityRules = append(securityRules, rowFilter[0])
	} else if len(parentRowFilter) > 0 {
		// If the metrics view does not have a row filter, we need to inherit the parent row filter
		securityRules = append(securityRules, parentRowFilter...)
	}

	newSpec.SecurityRules = securityRules
	newSpec.DisplayName = e.metricsView.DisplayName
	newSpec.Description = e.metricsView.Description

	// If the metrics view has a time dimension, override the parent metrics view time dimension
	if e.metricsView.TimeDimension != "" {
		newSpec.TimeDimension = e.metricsView.TimeDimension
	}
	// If the metrics view has a first day of week, override the parent metrics view first day of week
	if e.metricsView.FirstDayOfWeek > 0 {
		if e.metricsView.FirstDayOfWeek < 1 || e.metricsView.FirstDayOfWeek > 7 {
			return nil, fmt.Errorf("invalid first day of week %d in metrics view %q, must be between 1 and 7", e.metricsView.FirstDayOfWeek, e.metricsView.Parent)
		}
		newSpec.FirstDayOfWeek = e.metricsView.FirstDayOfWeek
	}
	// If the metrics view has a first month of year, override the parent metrics view first month of year
	if e.metricsView.FirstMonthOfYear > 0 {
		if e.metricsView.FirstMonthOfYear < 1 || e.metricsView.FirstMonthOfYear > 12 {
			return nil, fmt.Errorf("invalid first month of year %d in metrics view %q, must be between 1 and 12", e.metricsView.FirstMonthOfYear, e.metricsView.Parent)
		}
		newSpec.FirstMonthOfYear = e.metricsView.FirstMonthOfYear
	}
	// If the metrics view has a smallest time grain, override the parent metrics view smallest time grain
	if e.metricsView.SmallestTimeGrain != runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED {
		if e.metricsView.SmallestTimeGrain < newSpec.SmallestTimeGrain {
			return nil, fmt.Errorf("invalid smallest time grain %s in metrics view %q, must be greater than or equal to parent metrics view smallest time grain %s", e.metricsView.SmallestTimeGrain, e.metricsView.Parent, newSpec.SmallestTimeGrain)
		}
		newSpec.SmallestTimeGrain = e.metricsView.SmallestTimeGrain
	}
	// If the metrics view has ai instructions, override the parent metrics view ai instructions
	if e.metricsView.AiInstructions != "" {
		newSpec.AiInstructions = e.metricsView.AiInstructions
	}

	e.metricsView = newSpec

	return newSpec, nil
}

func (e *Executor) resolveFields(selector *runtimev1.FieldSelector, all []string) ([]string, error) {
	// Resolve the selector (it includes validation of the resulting fields against `all` if needed).
	res, err := fieldselectorpb.Resolve(selector, all)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve dimension or measure name selector: %w", err)
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
		// Validate time dimension type with a query
		rows, err := e.olap.Query(ctx, &drivers.Statement{
			Query: fmt.Sprintf("SELECT %s FROM %s LIMIT 0", expr, dialect.EscapeTable(t.Database, t.DatabaseSchema, t.Name)),
		})
		if err != nil {
			res.TimeDimensionErr = fmt.Errorf("failed to validate time dimension %q: %w", e.metricsView.TimeDimension, err)
			return
		}
		rows.Close() // Close rows immediately

		typeCode := rows.Schema.Fields[0].Type.Code
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
