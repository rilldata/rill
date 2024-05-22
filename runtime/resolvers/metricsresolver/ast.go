package metricsresolver

import (
	"context"
	"fmt"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/queries"
)

func (r *Resolver) BuildAST(ctx context.Context) (*AST, error) {
	d := r.olap.Dialect()
	ast := newAST(d)

	// Set the base table and filter
	ast.SetRawSelect(r.metricsView.Database, r.metricsView.DatabaseSchema, r.metricsView.Table, r.security.RowFilter, r.query.Where)

	// Set the base time range (if any)
	col, start, end, ok, err := r.ResolveTimeRange(ctx, r.query.TimeRange)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve time range: %w", err)
	}
	if ok {
		ast.SetBaseTimeRange(col, start, end)
	}

	// Set the comparison time range (if any)
	col, start, end, ok, err = r.ResolveTimeRange(ctx, r.query.ComparisonTimeRange)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve comparison time range: %w", err)
	}
	if ok {
		ast.SetComparisonTimeRange(col, start, end)
	}

	// Add each output dimension
	for _, qd := range r.query.Dimensions {
		// Handle if "compute.time_floor" is configured
		if qd.Compute != nil {
			if qd.Compute.TimeFloor == nil {
				return nil, fmt.Errorf(`unsupported "compute" for dimension %q`, qd.Name)
			}

			if !qd.Compute.TimeFloor.Grain.Valid() {
				return nil, fmt.Errorf(`invalid "grain" for dimension %q`, qd.Name)
			}

			dim, err := r.lookupDimension(qd.Compute.TimeFloor.Dimension)
			if err != nil {
				return nil, err
			}

			var tz string
			if r.query.TimeZone != nil {
				tz = *r.query.TimeZone
			}

			grain := qd.Compute.TimeFloor.Grain.ToProto()
			expr, err := d.DateTruncExpr(dim, grain, tz, int(r.metricsView.FirstDayOfWeek), int(r.metricsView.FirstMonthOfYear))
			if err != nil {
				return nil, fmt.Errorf(`failed to compute time floor for dimension %q: %w`, qd.Name, err)
			}

			ast.AddDimensionField(qd.Name, "", expr, dim.Unnest)
			continue
		}

		// Handle regular dimension
		dim, err := r.lookupDimension(qd.Name)
		if err != nil {
			return nil, err
		}

		expr := dim.Expression
		if expr == "" {
			expr = d.EscapeIdentifier(dim.Column)
		}

		ast.AddDimensionField(qd.Name, dim.Label, expr, dim.Unnest)
	}

	// Add each output measure
	for _, qm := range r.query.Measures {
		_, err := r.resolveMeasure(qm, d)
		if err != nil {
			return nil, fmt.Errorf("invalid measure %q: %w", qm.Name, err)
		}

		// TODO: Add
	}

	return ast, nil
}

func (r *Resolver) lookupDimension(name string) (*runtimev1.MetricsViewSpec_DimensionV2, error) {
	if name == r.metricsView.TimeDimension {
		return &runtimev1.MetricsViewSpec_DimensionV2{
			Name:   name,
			Column: name,
		}, nil
	}

	if !r.security.CanAccessField(name) {
		return nil, queries.ErrForbidden // TODO: Change type
	}

	for _, dim := range r.metricsView.Dimensions {
		if dim.Name == name {
			return dim, nil
		}
	}

	return nil, fmt.Errorf("dimension %q not found", name)
}

func (r *Resolver) lookupMeasure(name string) (*runtimev1.MetricsViewSpec_MeasureV2, error) {
	if !r.security.CanAccessField(name) {
		return nil, queries.ErrForbidden // TODO: Change type
	}

	for _, m := range r.metricsView.Measures {
		if m.Name == name {
			return m, nil
		}
	}

	return nil, fmt.Errorf("measure %q not found", name)
}

func (r *Resolver) resolveMeasure(qm Measure, d drivers.Dialect) (*runtimev1.MetricsViewSpec_MeasureV2, error) {
	if qm.Compute == nil {
		return r.lookupMeasure(qm.Name)
	}

	if err := qm.Compute.Validate(); err != nil {
		return nil, fmt.Errorf(`invalid "compute": %w`, qm.Name, err)
	}

	if qm.Compute.Count {
		return &runtimev1.MetricsViewSpec_MeasureV2{
			Name:       qm.Name,
			Expression: "COUNT(*)",
			Label:      "Count",
		}, nil
	}

	if qm.Compute.CountDistinct != nil {
		dim, err := r.lookupDimension(*qm.Compute.CountDistinct)
		if err != nil {
			return nil, err
		}

		expr := dim.Expression
		if expr == "" {
			expr = d.EscapeIdentifier(dim.Column)
		}

		return &runtimev1.MetricsViewSpec_MeasureV2{
			Name:       qm.Name,
			Expression: fmt.Sprintf("COUNT(DISTINCT %s)", expr),
			Label:      fmt.Sprintf("Unique %s", dim.Label),
		}, nil
	}

	if qm.Compute.ComparisonValue != nil {
		m, err := r.lookupMeasure(*qm.Compute.ComparisonValue)
		if err != nil {
			return nil, err
		}

		return &runtimev1.MetricsViewSpec_MeasureV2{
			Name:               qm.Name,
			Expression:         fmt.Sprintf("comparison.%s", d.EscapeIdentifier(m.Name)),
			Type:               runtimev1.MetricsViewSpec_MEASURE_TYPE_TIME_COMPARISON,
			ReferencedMeasures: []string{*qm.Compute.ComparisonValue},
			Label:              fmt.Sprintf("%s (prev)", m.Label),
		}, nil
	}

	if qm.Compute.ComparisonDelta != nil {
		m, err := r.lookupMeasure(*qm.Compute.ComparisonDelta)
		if err != nil {
			return nil, err
		}

		return &runtimev1.MetricsViewSpec_MeasureV2{
			Name:               qm.Name,
			Expression:         fmt.Sprintf("base.%s - comparison.%s", d.EscapeIdentifier(m.Name), d.EscapeIdentifier(m.Name)),
			Type:               runtimev1.MetricsViewSpec_MEASURE_TYPE_TIME_COMPARISON,
			ReferencedMeasures: []string{*qm.Compute.ComparisonDelta},
			Label:              fmt.Sprintf("%s (Δ)", m.Label),
		}, nil
	}

	if qm.Compute.ComparisonRatio != nil {
		m, err := r.lookupMeasure(*qm.Compute.ComparisonRatio)
		if err != nil {
			return nil, err
		}

		return &runtimev1.MetricsViewSpec_MeasureV2{
			Name:               qm.Name,
			Expression:         d.SafeDivideExpression(fmt.Sprintf("base.%s", d.EscapeIdentifier(m.Name)), fmt.Sprintf("base.%s", d.EscapeIdentifier(m.Name))),
			Type:               runtimev1.MetricsViewSpec_MEASURE_TYPE_TIME_COMPARISON,
			ReferencedMeasures: []string{*qm.Compute.ComparisonDelta},
			Label:              fmt.Sprintf("%s (Δ%)", m.Label),
		}, nil
	}

	return nil, fmt.Errorf(`unhandled compute operation`)
}

type AST struct {
	RawSelect        *SimpleSelect
	BaseSelect       *SimpleSelect
	ComparisonSelect *SimpleSelect
	DimensionFields  []SelectField
	MeasureFields    []SelectField

	dialect   drivers.Dialect
	nextIdent int
}

type SimpleSelect struct {
	Alias     string
	FromRaw   string
	WhereRaw  string
	WhereExpr *Expression
	Args      []any
}

type SelectField struct {
	Alias  string
	Label  string
	Expr   string
	Unnest bool
}

func newAST(dialect drivers.Dialect) *AST {
	return &AST{dialect: dialect}
}

func (a *AST) SetRawSelect(db, schema, table, rowFilter string, where *Expression) {
	a.RawSelect = &SimpleSelect{
		Alias:     a.generateIdentifier(),
		FromRaw:   a.dialect.EscapeTable(db, schema, table),
		WhereRaw:  rowFilter,
		WhereExpr: where,
	}
	a.BaseSelect = a.RawSelect
}

func (a *AST) SetBaseTimeRange(timeCol string, start, end *time.Time) {
	where, args, ok := a.timeRangeWhereClause(timeCol, start, end)
	if !ok {
		return
	}

	a.BaseSelect = &SimpleSelect{
		Alias:    a.generateIdentifier(),
		FromRaw:  a.dialect.EscapeIdentifier(a.RawSelect.Alias),
		WhereRaw: where,
		Args:     args,
	}
}

func (a *AST) SetComparisonTimeRange(timeCol string, start, end *time.Time) {
	where, args, ok := a.timeRangeWhereClause(timeCol, start, end)
	if !ok {
		return
	}

	a.ComparisonSelect = &SimpleSelect{
		Alias:    a.generateIdentifier(),
		FromRaw:  a.dialect.EscapeIdentifier(a.RawSelect.Alias),
		WhereRaw: where,
		Args:     args,
	}
}

// Note on unnest: Adds a lateral unnest. Might be faster to apply after WHERE, unless Druid. Note: WhereExpr uses different operators, doesn't rely on unnested table.
func (a *AST) AddDimensionField(alias, label, expr string, unnest bool) {
	a.DimensionFields = append(a.DimensionFields, SelectField{
		Alias:  alias,
		Label:  label,
		Expr:   expr,
		Unnest: unnest,
	})
}

func (a *AST) generateIdentifier() string {
	tmp := fmt.Sprintf("t%d", a.nextIdent)
	a.nextIdent++
	return tmp
}

func (a *AST) timeRangeWhereClause(timeCol string, start, end *time.Time) (string, []any, bool) {
	var where string
	var args []any
	if start != nil && end != nil {
		col := a.dialect.EscapeIdentifier(timeCol)
		where = fmt.Sprintf("%s >= ? AND %s < ?", col, col)
		args = []any{*start, *end}
	} else if start != nil {
		where = fmt.Sprintf("%s >= ?", a.dialect.EscapeIdentifier(timeCol))
		args = []any{*start}
	} else if end != nil {
		where = fmt.Sprintf("%s < ?", a.dialect.EscapeIdentifier(timeCol))
		args = []any{*end}
	} else {
		return "", nil, false
	}
	return where, args, true
}
