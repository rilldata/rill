package metricsresolver

import (
	"errors"
	"fmt"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/queries"
)

type AST struct {
	Root                *MetricsSelect
	UnderlyingSelect    *RawSelect
	BaseTimeWhere       *WhereExpr
	ComparisonTimeWhere *WhereExpr
	// OrderByExprs        []string
	// HavingWhere         *WhereExpr

	metricsView    *runtimev1.MetricsViewSpec
	security       *runtime.ResolvedMetricsViewSecurity
	query          *Query
	dialect        drivers.Dialect
	nextIdentifier int
	dimFields      []SelectField
}

type RawSelect struct {
	Alias    string
	FromExpr string
	Where    WhereExpr
}

type MetricsSelect struct {
	Alias              string
	DimFields          []SelectField
	MeasureFields      []SelectField
	Group              bool
	FromUnderlying     bool
	FromSelect         *MetricsSelect
	LeftJoinSelects    []*MetricsSelect
	ComparisonSelect   *MetricsSelect
	ComparisonJoinType string
}

type SelectField struct {
	Name   string
	Label  string
	Expr   string
	Unnest bool // Note on unnest: Adds a lateral unnest. Might be faster to apply after WHERE, unless Druid. Note: WhereExpr uses different operators, doesn't rely on unnested table.
}

type WhereExpr struct {
	Expr string
	Args []any
}

type OrderByField struct {
	Name string
	Expr string
	Desc bool
}

func buildAST(mv *runtimev1.MetricsViewSpec, sec *runtime.ResolvedMetricsViewSecurity, qry *Query, dialect drivers.Dialect) (*AST, error) {
	// Init
	ast := &AST{
		metricsView: mv,
		security:    sec,
		query:       qry,
		dialect:     dialect,
	}

	// Set the underlying table and filter
	ast.setUnderlyingSelect(mv.Database, mv.DatabaseSchema, mv.Table, sec.RowFilter, "", nil) // TODO: qry.Where

	// Set the time ranges
	if ast.query.TimeRange != nil {
		ast.setBaseTimeRange(ast.metricsView.TimeDimension, ast.query.TimeRange.StartTime, ast.query.TimeRange.EndTime)
	}
	if ast.query.ComparisonTimeRange != nil {
		ast.setComparisonTimeRange(ast.metricsView.TimeDimension, ast.query.ComparisonTimeRange.StartTime, ast.query.ComparisonTimeRange.EndTime)
	}

	// Build dimensions for underlying
	dimFields := make([]SelectField, 0, len(ast.query.Dimensions))
	for _, qd := range ast.query.Dimensions {
		dim, err := ast.resolveDimension(qd, true)
		if err != nil {
			return nil, fmt.Errorf("invalid dimension %q: %w", qd.Name, err)
		}

		dimFields = append(dimFields, SelectField{
			Name:   dim.Name,
			Label:  dim.Label,
			Expr:   ast.dialect.MetricsViewDimensionExpression(dim),
			Unnest: dim.Unnest,
		})
	}

	// Add dimensions to the root select and cache it in the AST (for later use in case we need to add JOINs against the underlying)
	ast.Root.DimFields = dimFields
	ast.dimFields = dimFields

	// Add each output measure
	for _, qm := range ast.query.Measures {
		m, err := ast.resolveMeasure(qm, true)
		if err != nil {
			return nil, fmt.Errorf("invalid measure %q: %w", qm.Name, err)
		}

		// TODO: Move down to addMeasureField
		err = ast.checkRequiredDimensionsPresent(m)
		if err != nil {
			return nil, fmt.Errorf("can't query measure %q: %w", qm.Name, err)
		}

		err = ast.addMeasureField(ast.Root, m)
		if err != nil {
			return nil, fmt.Errorf("can't query measure %q: %w", qm.Name, err)
		}
	}

	// TODO:
	// Sort (implication: comparison join type)
	// Having
	// Limit
	// Offset

	return ast, nil
}

func (a *AST) lookupDimension(name string, visible bool) (*runtimev1.MetricsViewSpec_DimensionV2, error) {
	if name == a.metricsView.TimeDimension {
		return &runtimev1.MetricsViewSpec_DimensionV2{
			Name:   name,
			Column: name,
		}, nil
	}

	if visible {
		if !a.security.CanAccessField(name) {
			return nil, queries.ErrForbidden // TODO: Change type
		}
	}

	for _, dim := range a.metricsView.Dimensions {
		if dim.Name == name {
			return dim, nil
		}
	}

	return nil, fmt.Errorf("dimension %q not found", name)
}

func (a *AST) lookupMeasure(name string, visible bool) (*runtimev1.MetricsViewSpec_MeasureV2, error) {
	if visible {
		if !a.security.CanAccessField(name) {
			return nil, queries.ErrForbidden // TODO: Change type
		}
	}

	for _, m := range a.metricsView.Measures {
		if m.Name == name {
			return m, nil
		}
	}

	return nil, fmt.Errorf("measure %q not found", name)
}

func (a *AST) checkNameForComputedField(name string) error {
	if name == a.metricsView.TimeDimension {
		return errors.New("name for computed field collides with the time dimension name")
	}

	for _, d := range a.metricsView.Dimensions {
		if d.Name == name {
			return errors.New("name for computed field collides with an existing dimension name")
		}
	}

	for _, m := range a.metricsView.Measures {
		if m.Name == name {
			return errors.New("name for computed field collides with an existing measure name")
		}
	}

	return nil
}

func (a *AST) resolveDimension(qd Dimension, visible bool) (*runtimev1.MetricsViewSpec_DimensionV2, error) {
	// Handle regular dimension
	if qd.Compute == nil {
		return a.lookupDimension(qd.Name, visible)
	}

	// Handle computed dimension. This means "compute.time_floor" must be configured.

	if qd.Compute.TimeFloor == nil {
		return nil, fmt.Errorf(`unsupported "compute"`)
	}

	if !qd.Compute.TimeFloor.Grain.Valid() {
		return nil, fmt.Errorf(`invalid "grain"`)
	}

	dim, err := a.lookupDimension(qd.Compute.TimeFloor.Dimension, visible)
	if err != nil {
		return nil, err
	}

	if qd.Name != qd.Compute.TimeFloor.Dimension {
		err := a.checkNameForComputedField(qd.Name)
		if err != nil {
			return nil, err
		}
	}

	var tz string
	if a.query.TimeZone != nil {
		tz = *a.query.TimeZone
	}

	grain := qd.Compute.TimeFloor.Grain.ToProto()
	expr, err := a.dialect.DateTruncExpr(dim, grain, tz, int(a.metricsView.FirstDayOfWeek), int(a.metricsView.FirstMonthOfYear))
	if err != nil {
		return nil, fmt.Errorf(`failed to compute time floor: %w`, err)
	}

	return &runtimev1.MetricsViewSpec_DimensionV2{
		Name:       qd.Name,
		Expression: expr,
		Label:      dim.Label,
		Unnest:     dim.Unnest,
	}, nil
}

func (a *AST) resolveMeasure(qm Measure, visible bool) (*runtimev1.MetricsViewSpec_MeasureV2, error) {
	if qm.Compute == nil {
		return a.lookupMeasure(qm.Name, visible)
	}

	if err := qm.Compute.Validate(); err != nil {
		return nil, fmt.Errorf(`invalid "compute": %w`, err)
	}

	err := a.checkNameForComputedField(qm.Name)
	if err != nil {
		return nil, err
	}

	if qm.Compute.Count {
		return &runtimev1.MetricsViewSpec_MeasureV2{
			Name:       qm.Name,
			Expression: "COUNT(*)",
			Label:      "Count",
		}, nil
	}

	if qm.Compute.CountDistinct != nil {
		dim, err := a.lookupDimension(*qm.Compute.CountDistinct, visible)
		if err != nil {
			return nil, err
		}

		expr := dim.Expression
		if expr == "" {
			expr = a.dialect.EscapeIdentifier(dim.Column)
		}

		return &runtimev1.MetricsViewSpec_MeasureV2{
			Name:       qm.Name,
			Expression: fmt.Sprintf("COUNT(DISTINCT %s)", expr),
			Label:      fmt.Sprintf("Unique %s", dim.Label),
		}, nil
	}

	if qm.Compute.ComparisonValue != nil {
		m, err := a.lookupMeasure(*qm.Compute.ComparisonValue, visible)
		if err != nil {
			return nil, err
		}

		return &runtimev1.MetricsViewSpec_MeasureV2{
			Name:               qm.Name,
			Expression:         fmt.Sprintf("comparison.%s", a.dialect.EscapeIdentifier(m.Name)),
			Type:               runtimev1.MetricsViewSpec_MEASURE_TYPE_TIME_COMPARISON,
			ReferencedMeasures: []string{*qm.Compute.ComparisonValue},
			Label:              fmt.Sprintf("%s (prev)", m.Label),
		}, nil
	}

	if qm.Compute.ComparisonDelta != nil {
		m, err := a.lookupMeasure(*qm.Compute.ComparisonDelta, visible)
		if err != nil {
			return nil, err
		}

		return &runtimev1.MetricsViewSpec_MeasureV2{
			Name:               qm.Name,
			Expression:         fmt.Sprintf("base.%s - comparison.%s", a.dialect.EscapeIdentifier(m.Name), a.dialect.EscapeIdentifier(m.Name)),
			Type:               runtimev1.MetricsViewSpec_MEASURE_TYPE_TIME_COMPARISON,
			ReferencedMeasures: []string{*qm.Compute.ComparisonDelta},
			Label:              fmt.Sprintf("%s (Δ)", m.Label),
		}, nil
	}

	if qm.Compute.ComparisonRatio != nil {
		m, err := a.lookupMeasure(*qm.Compute.ComparisonRatio, visible)
		if err != nil {
			return nil, err
		}

		return &runtimev1.MetricsViewSpec_MeasureV2{
			Name:               qm.Name,
			Expression:         a.dialect.SafeDivideExpression(fmt.Sprintf("base.%s", a.dialect.EscapeIdentifier(m.Name)), fmt.Sprintf("base.%s", a.dialect.EscapeIdentifier(m.Name))),
			Type:               runtimev1.MetricsViewSpec_MEASURE_TYPE_TIME_COMPARISON,
			ReferencedMeasures: []string{*qm.Compute.ComparisonDelta},
			Label:              fmt.Sprintf("%s (Δ%%)", m.Label),
		}, nil
	}

	return nil, fmt.Errorf(`unhandled compute operation`)
}

func (a *AST) checkRequiredDimensionsPresent(m *runtimev1.MetricsViewSpec_MeasureV2) error {
	for _, rd := range m.RequiredDimensions {
		var found bool
		for _, qd := range a.query.Dimensions {
			if rd.TimeGrain == runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED {
				if qd.Compute != nil {
					continue
				}
			} else {
				if qd.Compute == nil || qd.Compute.TimeFloor == nil {
					continue
				}

				if TimeGrainFromProto(rd.TimeGrain) != qd.Compute.TimeFloor.Grain {
					continue
				}
			}

			if rd.Name == qd.Name {
				found = true
				break
			}
		}

		if !found {
			if rd.TimeGrain != runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED {
				return fmt.Errorf("missing required dimension %q at %q granularity", rd.Name, TimeGrainFromProto(rd.TimeGrain))
			}
			return fmt.Errorf("missing required dimension %q", rd.Name)
		}
	}

	return nil
}

func (a *AST) setUnderlyingSelect(db, schema, table, rowFilter, whereExpr string, whereArgs []any) {
	if rowFilter != "" {
		if whereExpr == "" {
			whereExpr = rowFilter
		} else {
			whereExpr = fmt.Sprintf("(%s) AND (%s)", rowFilter, whereExpr)
		}
	}

	a.UnderlyingSelect = &RawSelect{
		Alias:    a.generateIdentifier(),
		FromExpr: a.dialect.EscapeTable(db, schema, table),
		Where: WhereExpr{
			Expr: whereExpr,
			Args: whereArgs,
		},
	}

	a.Root = &MetricsSelect{
		Alias:          a.generateIdentifier(),
		Group:          true,
		FromUnderlying: true,
	}
}

func (a *AST) setBaseTimeRange(timeCol string, start, end *time.Time) {
	expr, args, ok := a.expressionForTimeRange(timeCol, start, end)
	if !ok {
		return
	}

	a.BaseTimeWhere = &WhereExpr{
		Expr: expr,
		Args: args,
	}
}

func (a *AST) setComparisonTimeRange(timeCol string, start, end *time.Time) {
	expr, args, ok := a.expressionForTimeRange(timeCol, start, end)
	if !ok {
		return
	}

	a.ComparisonTimeWhere = &WhereExpr{
		Expr: expr,
		Args: args,
	}
}

func (a *AST) addMeasureField(s *MetricsSelect, m *runtimev1.MetricsViewSpec_MeasureV2) error {
	if a.hasMeasure(s, m.Name) {
		return nil
	}

	switch m.Type {
	case runtimev1.MetricsViewSpec_MEASURE_TYPE_UNSPECIFIED, runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE:
		return a.addSimpleMeasure(s, m)
	case runtimev1.MetricsViewSpec_MEASURE_TYPE_DERIVED:
		return a.addDerivedMeasure(s, m)
	case runtimev1.MetricsViewSpec_MEASURE_TYPE_TIME_COMPARISON:
		return a.addTimeComparisonMeasure(s, m)
	default:
		panic("unhandled measure type")
	}
}

func (a *AST) hasMeasure(s *MetricsSelect, name string) bool {
	for _, f := range s.MeasureFields {
		if f.Name == name {
			return true
		}
	}
	return false
}

// we know the measure is not in s, but might be in a subselect
func (a *AST) addSimpleMeasure(s *MetricsSelect, m *runtimev1.MetricsViewSpec_MeasureV2) error {
	if s.FromUnderlying {
		s.MeasureFields = append(s.MeasureFields, SelectField{
			Name:  m.Name,
			Label: m.Label,
			Expr:  a.expressionForMeasure(m),
		})

		return nil
	}

	if !a.hasMeasure(s.FromSelect, m.Name) { // Check because could have been added as a ref
		err := a.addSimpleMeasure(s.FromSelect, m)
		if err != nil {
			return err
		}
	}

	expr := a.expressionForMember(s.FromSelect.Alias, m.Name)
	if s.Group {
		expr = a.expressionForAnyInGroup(expr)
	}

	s.MeasureFields = append(s.MeasureFields, SelectField{
		Name:  m.Name,
		Label: m.Label,
		Expr:  expr,
	})

	return nil
}

func (a *AST) addDerivedMeasure(s *MetricsSelect, m *runtimev1.MetricsViewSpec_MeasureV2) error {
	// TODO: Recurse if comparison?

	if len(m.PerDimensions) > 0 {
		return a.addDerivedMeasureWithPer(s, m)
	}

	err := a.addReferencedMeasuresToScope(s, m.ReferencedMeasures)
	if err != nil {
		return err
	}

	// NOTE: Invariants ensure there's no ambiguity for the referenced names.
	// Only case of ambiguity should be for comparisons, where we require users to use base.xxx and comparison.xxx.

	expr := a.expressionForMeasure(m)
	if s.Group {
		// TODO: Risk of expr containing a window (can't be wrapped by ANY_VALUE).
		// Fix by wrapping with a non-grouped.
		expr = a.expressionForAnyInGroup(expr)
	}

	s.MeasureFields = append(s.MeasureFields, SelectField{
		Name:  m.Name,
		Label: m.Label,
		Expr:  expr,
	})

	return nil
}

func (a *AST) addDerivedMeasureWithPer(_ *MetricsSelect, _ *runtimev1.MetricsViewSpec_MeasureV2) error {
	return fmt.Errorf("support for \"per\" not implemented")
}

func (a *AST) addTimeComparisonMeasure(s *MetricsSelect, m *runtimev1.MetricsViewSpec_MeasureV2) error {
	if s.ComparisonSelect == nil {
		a.wrapSelect(s, "base")
		s.ComparisonSelect = &MetricsSelect{
			Alias:          "comparison",
			DimFields:      a.dimFields,
			Group:          true,
			FromUnderlying: true,
		}
	}

	err := a.addReferencedMeasuresToScope(s, m.ReferencedMeasures)
	if err != nil {
		return err
	}

	expr := a.expressionForMeasure(m)
	if s.Group {
		// TODO: You know the thing
		expr = a.expressionForAnyInGroup(expr)
	}

	s.MeasureFields = append(s.MeasureFields, SelectField{
		Name:  m.Name,
		Label: m.Label,
		Expr:  expr,
	})

	return nil
}

// We rely on the name being unique to the measure (so can't mean something else)
func (a *AST) addReferencedMeasuresToScope(s *MetricsSelect, referencedMeasures []string) error {
	if len(referencedMeasures) == 0 {
		return nil
	}

	if s.FromUnderlying {
		a.wrapSelect(s, a.generateIdentifier())
	}

	for _, rm := range referencedMeasures {
		m, err := a.lookupMeasure(rm, false)
		if err != nil {
			return err
		}

		switch m.Type {
		case runtimev1.MetricsViewSpec_MEASURE_TYPE_UNSPECIFIED, runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE:
			// Keep going
		default:
			return fmt.Errorf("referenced measure %q is not simple", rm)
		}

		err = a.addMeasureField(s.FromSelect, m)
		if err != nil {
			return err
		}

		if s.ComparisonSelect != nil {
			err = a.addMeasureField(s.ComparisonSelect, m)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (a *AST) wrapSelect(s *MetricsSelect, innerAlias string) {
	cpy := *s
	cpy.Alias = innerAlias

	s.DimFields = make([]SelectField, 0, len(cpy.DimFields))
	for _, f := range cpy.DimFields {
		f.Expr = a.expressionForMember(cpy.Alias, f.Name)
		s.DimFields = append(s.DimFields, f)
	}

	s.MeasureFields = make([]SelectField, 0, len(cpy.MeasureFields))
	for _, f := range cpy.MeasureFields {
		f.Expr = a.expressionForMember(cpy.Alias, f.Name)
		s.MeasureFields = append(s.MeasureFields, f)
	}

	s.Group = false
	s.FromUnderlying = false
	s.FromSelect = &cpy
	s.LeftJoinSelects = nil
	s.ComparisonSelect = nil
	s.ComparisonJoinType = ""
}

func (a *AST) expressionForTimeRange(timeCol string, start, end *time.Time) (string, []any, bool) {
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

func (a *AST) expressionForMeasure(m *runtimev1.MetricsViewSpec_MeasureV2) string {
	if m.Window != nil {
		return m.Expression
	}

	// incorporate required dimensions. maybe dims?

	// TODO:
	return ""
}

// not escaping tbl because only used for generated aliases
func (a *AST) expressionForMember(tbl, name string) string {
	return fmt.Sprintf("%s.%s", tbl, a.dialect.EscapeIdentifier(name))
}

func (a *AST) expressionForAnyInGroup(expr string) string {
	return fmt.Sprintf("ANY_VALUE(%s)", expr)
}

func (a *AST) generateIdentifier() string {
	tmp := fmt.Sprintf("t%d", a.nextIdentifier)
	a.nextIdentifier++
	return tmp
}
