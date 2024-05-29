package metricsresolver

import (
	"errors"
	"fmt"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/queries"
)

// AST is the abstract syntax tree for a metrics SQL query.
type AST struct {
	Root *MetricsSelect

	baseSelect *PlainSelect // Cached here because it's needed when adding new JOIN nodes.
	dimFields  []SelectField

	metricsView    *runtimev1.MetricsViewSpec
	security       *runtime.ResolvedMetricsViewSecurity
	query          *Query
	dialect        drivers.Dialect
	nextIdentifier int
}

// PlainSelect represents a "SELECT * FROM ... WHERE ..." query.
type PlainSelect struct {
	From  string
	Where *WhereExpr
}

// MetricsSelect represents a query that computes measures by dimensions.
// The from/join clauses are not all compatible. The allowed combinations are:
//   - FromPlain
//   - FromSelect and optionally LeftJoinSelects
//   - FromSelect and optionally JoinComparisonSelect
type MetricsSelect struct {
	Alias                string
	DimFields            []SelectField
	MeasureFields        []SelectField
	Group                bool
	FromPlain            *PlainSelect
	FromSelect           *MetricsSelect
	LeftJoinSelects      []*MetricsSelect
	JoinComparisonSelect *MetricsSelect
	JoinComparisonType   string
	Where                *WhereExpr
	Having               *WhereExpr
	OrderBy              []OrderByField
	Limit                *int
	Offset               *int
}

// SelectField represents a field in a SELECT clause.
// The Name must always match a the name of a dimension/measure in the metrics view or a computed field specified in the request.
// This means that if two fields in different places in the AST have the same Name, they're guaranteed to resolve to the same value.
type SelectField struct {
	Name        string
	Label       string
	Expr        string
	Unnest      bool
	UnnestAlias string
}

// WhereExpr represents an expression for a WHERE clause.
type WhereExpr struct {
	Expr string
	Args []any
}

// OrderByField represents a field in an ORDER BY clause.
type OrderByField struct {
	Name string
	Desc bool
}

func BuildAST(mv *runtimev1.MetricsViewSpec, sec *runtime.ResolvedMetricsViewSecurity, qry *Query, dialect drivers.Dialect) (*AST, error) {
	// Validate there's at least one dim or measure
	if len(qry.Dimensions) == 0 && len(qry.Measures) == 0 {
		return nil, fmt.Errorf("must specify at least one dimension or measure")
	}

	// Init
	ast := &AST{
		metricsView: mv,
		security:    sec,
		query:       qry,
		dialect:     dialect,
	}

	// Build dimensions to apply against the underlying SELECT.
	// We cache these in the AST type because when resolving expressions and adding new JOINs, we need the ability to reference these.
	dimFields := make([]SelectField, 0, len(ast.query.Dimensions))
	for _, qd := range ast.query.Dimensions {
		dim, err := ast.resolveDimension(qd, true)
		if err != nil {
			return nil, fmt.Errorf("invalid dimension %q: %w", qd.Name, err)
		}

		var unnestAlias string
		if dim.Unnest {
			unnestAlias = ast.generateIdentifier()
		}

		dimFields = append(dimFields, SelectField{
			Name:        dim.Name,
			Label:       dim.Label,
			Expr:        ast.dialect.MetricsViewDimensionExpression(dim),
			Unnest:      dim.Unnest,
			UnnestAlias: unnestAlias,
		})
	}
	ast.dimFields = dimFields

	// Build underlying SELECT
	where, err := ast.buildUnderlyingWhere()
	if err != nil {
		return nil, err
	}
	ast.baseSelect = &PlainSelect{
		From:  ast.dialect.EscapeTable(mv.Database, mv.DatabaseSchema, mv.Table),
		Where: where,
	}

	// Build initial root node (empty query against the base select)
	ast.Root = &MetricsSelect{
		Alias:     ast.generateIdentifier(),
		DimFields: ast.dimFields,
		Group:     true,
		FromPlain: ast.baseSelect,
	}

	// Add time range to the root node
	ast.addTimeRange(ast.Root, ast.query.TimeRange)

	// Incrementally add each output measure.
	// As each measure is added, the AST is transformed to accommodate it based on its type.
	for _, qm := range ast.query.Measures {
		m, err := ast.resolveMeasure(qm, true)
		if err != nil {
			return nil, fmt.Errorf("invalid measure %q: %w", qm.Name, err)
		}

		err = ast.addMeasureField(ast.Root, m)
		if err != nil {
			return nil, fmt.Errorf("can't query measure %q: %w", qm.Name, err)
		}
	}

	// Handle Having. If the root node is grouped, we add it as a HAVING clause, otherwise wrap it in a SELECT and add it as a WHERE clause.
	if ast.query.Having != nil {
		expr, args, err := ast.buildExpression(ast.query.Having, ast.Root.Group, ast.Root)
		if err != nil {
			return nil, fmt.Errorf("failed to compile 'having': %w", err)
		}

		res := &WhereExpr{
			Expr: expr,
			Args: args,
		}

		if ast.Root.Group {
			ast.Root.Having = res
		} else {
			// We need to wrap in a new SELECT because a WHERE clause cannot apply directly to a field with a window function.
			// If this turns out to have performance implications, we could consider only wrapping if one of ast.Root.MeasureFields contains a window function.
			ast.wrapSelect(ast.Root, ast.generateIdentifier())
			ast.addWhere(ast.Root, res)
		}
	}

	// Incrementally add each sort criterion.
	for _, s := range ast.query.Sort {
		err = ast.addSortField(ast.Root, s.Name, s.Desc)
		if err != nil {
			return nil, fmt.Errorf("can't sort by %q: %w", s.Name, err)
		}
	}

	// Add limit and offset
	ast.Root.Limit = ast.query.Limit
	ast.Root.Offset = ast.query.Offset

	return ast, nil
}

// resolveDimension returns a dimension spec for the given dimension query.
// If the dimension query specifies a computed dimension, it constructs a dimension spec to match it.
func (a *AST) resolveDimension(qd Dimension, visible bool) (*runtimev1.MetricsViewSpec_DimensionV2, error) {
	// Handle regular dimension
	if qd.Compute == nil {
		return a.lookupDimension(qd.Name, visible)
	}

	// Handle computed dimension. This means "compute.time_floor" must be configured.

	if qd.Compute.TimeFloor == nil {
		return nil, errors.New(`unsupported "compute"`)
	}

	if !qd.Compute.TimeFloor.Grain.Valid() {
		return nil, errors.New(`invalid "grain"`)
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

// resolveMeasure returns a measure spec for the given measure query.
// If the measure query specifies a computed measure, it constructs a measure spec to match it.
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
			Type:       runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE,
			Label:      "Count",
		}, nil
	}

	if qm.Compute.CountDistinct != nil {
		dim, err := a.lookupDimension(qm.Compute.CountDistinct.Dimension, visible)
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
			Type:       runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE,
			Label:      fmt.Sprintf("Unique %s", dim.Label),
		}, nil
	}

	if qm.Compute.ComparisonValue != nil {
		m, err := a.lookupMeasure(qm.Compute.ComparisonValue.Measure, visible)
		if err != nil {
			return nil, err
		}

		return &runtimev1.MetricsViewSpec_MeasureV2{
			Name:               qm.Name,
			Expression:         fmt.Sprintf("comparison.%s", a.dialect.EscapeIdentifier(m.Name)),
			Type:               runtimev1.MetricsViewSpec_MEASURE_TYPE_TIME_COMPARISON,
			ReferencedMeasures: []string{qm.Compute.ComparisonValue.Measure},
			Label:              fmt.Sprintf("%s (prev)", m.Label),
		}, nil
	}

	if qm.Compute.ComparisonDelta != nil {
		m, err := a.lookupMeasure(qm.Compute.ComparisonDelta.Measure, visible)
		if err != nil {
			return nil, err
		}

		return &runtimev1.MetricsViewSpec_MeasureV2{
			Name:               qm.Name,
			Expression:         fmt.Sprintf("base.%s - comparison.%s", a.dialect.EscapeIdentifier(m.Name), a.dialect.EscapeIdentifier(m.Name)),
			Type:               runtimev1.MetricsViewSpec_MEASURE_TYPE_TIME_COMPARISON,
			ReferencedMeasures: []string{qm.Compute.ComparisonDelta.Measure},
			Label:              fmt.Sprintf("%s (Δ)", m.Label),
		}, nil
	}

	if qm.Compute.ComparisonRatio != nil {
		m, err := a.lookupMeasure(qm.Compute.ComparisonRatio.Measure, visible)
		if err != nil {
			return nil, err
		}

		return &runtimev1.MetricsViewSpec_MeasureV2{
			Name:               qm.Name,
			Expression:         a.dialect.SafeDivideExpression(fmt.Sprintf("base.%s", a.dialect.EscapeIdentifier(m.Name)), fmt.Sprintf("comparison.%s", a.dialect.EscapeIdentifier(m.Name))),
			Type:               runtimev1.MetricsViewSpec_MEASURE_TYPE_TIME_COMPARISON,
			ReferencedMeasures: []string{qm.Compute.ComparisonRatio.Measure},
			Label:              fmt.Sprintf("%s (Δ%%)", m.Label),
		}, nil
	}

	return nil, errors.New("unhandled compute operation")
}

// lookupDimension finds a dimension spec in the metrics view.
// If visible is true, it returns an error if the security policy does not grant access to the dimension.
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

// lookupMeasure finds a measure spec in the metrics view.
// If visible is true, it returns an error if the security policy does not grant access to the measure.
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

// checkNameForComputedField checks that the name for a computed field does not collide with an existing dimension or measure name.
// (This is necessary because even if the other name is not used in the query, it might be referenced by a derived measure.)
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

// checkRequiredDimensionsPresentInQuery checks that all the dimensions required by the measure are present in the query.
// It checks against the query's dimensions because:
//  1. This enables correctly checking against the underlying time dimension name for dimensions with time floor applied.
//  2. The query's dimensions are projected into every sub-query, so it's not necessary to check against the current sub-query's DimFields.
func (a *AST) checkRequiredDimensionsPresentInQuery(m *runtimev1.MetricsViewSpec_MeasureV2) error {
	for _, rd := range m.RequiredDimensions {
		var found bool
		for _, qd := range a.query.Dimensions {
			// Handle computed dimension
			if qd.Compute != nil {
				if qd.Compute.TimeFloor == nil {
					continue
				}

				if rd.Name != qd.Compute.TimeFloor.Dimension {
					continue
				}

				isGrainSpecified := rd.TimeGrain != runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED
				isGrainDifferent := TimeGrainFromProto(rd.TimeGrain) != qd.Compute.TimeFloor.Grain
				if isGrainSpecified && isGrainDifferent {
					continue
				}

				found = true
				break
			}

			// Check for time dimension with time floor applied
			if rd.TimeGrain != runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED {
				if qd.Compute == nil || qd.Compute.TimeFloor == nil {
					continue
				}

				if TimeGrainFromProto(rd.TimeGrain) != qd.Compute.TimeFloor.Grain {
					continue
				}

				if rd.Name == qd.Compute.TimeFloor.Dimension {
					found = true
					break
				}
			}

			// Checking for regular dimension
			if qd.Compute != nil {
				continue
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

// buildUnderlyingWhere constructs the base WHERE clause for the query.
func (a *AST) buildUnderlyingWhere() (*WhereExpr, error) {
	expr, args, err := a.buildExpression(a.query.Where, false, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to compile 'where': %w", err)
	}

	if a.security != nil && a.security.RowFilter != "" {
		if expr == "" {
			expr = a.security.RowFilter
		} else {
			expr = fmt.Sprintf("(%s) AND (%s)", a.security.RowFilter, expr)
		}
	}

	return &WhereExpr{
		Expr: expr,
		Args: args,
	}, nil
}

// addTimeRange adds a time range to the given MetricsSelect's WHERE clause.
func (a *AST) addTimeRange(n *MetricsSelect, tr *TimeRange) {
	if tr == nil || tr.IsZero() {
		return
	}

	// Since resolving time ranges may require contextual info (like watermarks), the upstream caller is responsible for resolving them.
	if tr.Start.IsZero() && tr.End.IsZero() {
		panic("ast received a non-empty, unresolved time range")
	}

	expr, args := a.expressionForTimeRange(a.metricsView.TimeDimension, tr.Start, tr.End)
	a.addWhere(n, &WhereExpr{
		Expr: expr,
		Args: args,
	})
}

// addWhere adds or merges the given WHERE expression to the MetricsSelect's WHERE clause.
func (a *AST) addWhere(n *MetricsSelect, w *WhereExpr) {
	if n.Where == nil {
		n.Where = w
	} else {
		n.Where.Expr = fmt.Sprintf("(%s) AND (%s)", n.Where.Expr, w.Expr)
		n.Where.Args = append(n.Where.Args, w.Args...)
	}
}

// addMeasureField adds a measure field to the given MetricsSelect.
// Depending on the measure type, it may rewrite the MetricsSelect to accommodate the measure.
func (a *AST) addMeasureField(n *MetricsSelect, m *runtimev1.MetricsViewSpec_MeasureV2) error {
	// Skip if the measure has already been added.
	// This can happen if the measure was already added as a referenced measure of a derived measure.
	if a.hasMeasure(n, m.Name) {
		return nil
	}

	// Check that the measure's required dimensions are satisfied
	err := a.checkRequiredDimensionsPresentInQuery(m)
	if err != nil {
		return err
	}

	switch m.Type {
	case runtimev1.MetricsViewSpec_MEASURE_TYPE_UNSPECIFIED, runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE:
		return a.addSimpleMeasure(n, m)
	case runtimev1.MetricsViewSpec_MEASURE_TYPE_DERIVED:
		return a.addDerivedMeasure(n, m)
	case runtimev1.MetricsViewSpec_MEASURE_TYPE_TIME_COMPARISON:
		return a.addTimeComparisonMeasure(n, m)
	default:
		panic("unhandled measure type")
	}
}

// addSimpleMeasure adds a measure of type simple to the given MetricsSelect.
// When called, we know the measure is not present in the MetricsSelect, but it might be present in a sub-select.
func (a *AST) addSimpleMeasure(n *MetricsSelect, m *runtimev1.MetricsViewSpec_MeasureV2) error {
	// Base case: it targets a plain SELECT.
	// Add the measure directly to the SELECT list.
	if n.FromPlain != nil {
		n.MeasureFields = append(n.MeasureFields, SelectField{
			Name:  m.Name,
			Label: m.Label,
			Expr:  a.expressionForMeasure(m, n),
		})

		return nil
	}

	// Recursive case: it targets another MetricsSelect.
	// We recurse on the sub-select and add a pass-through field to the current node.

	if !a.hasMeasure(n.FromSelect, m.Name) { // Don't recurse if already in scope in sub-query
		err := a.addSimpleMeasure(n.FromSelect, m)
		if err != nil {
			return err
		}
	}

	expr := a.expressionForMember(n.FromSelect.Alias, m.Name)
	if n.Group {
		expr = a.expressionForAnyInGroup(expr)
	}

	n.MeasureFields = append(n.MeasureFields, SelectField{
		Name:  m.Name,
		Label: m.Label,
		Expr:  expr,
	})

	return nil
}

// addDerivedMeasure adds a measure of type derived to the given MetricsSelect.
// When called, we know the measure is not present in the MetricsSelect, but it might be present in a sub-select.
func (a *AST) addDerivedMeasure(n *MetricsSelect, m *runtimev1.MetricsViewSpec_MeasureV2) error {
	// Handle derived measures with "per" dimensions separately.
	if len(m.PerDimensions) > 0 {
		return a.addDerivedMeasureWithPer(n, m)
	}

	// If the current node has a comparison join, push calculation of the derived measure into its FromSelect and add a pass-through field in the current node.
	// This avoids a potential ambiguity issue because the derived measure expression does not use "base.name" and "comparison.name" to identify referenced measures,
	// so we need to ensure the referenced names exist only in ONE sub-query.
	if n.JoinComparisonSelect != nil {
		if !a.hasMeasure(n.FromSelect, m.Name) {
			err := a.addDerivedMeasure(n.FromSelect, m)
			if err != nil {
				return err
			}
		}

		expr := a.expressionForMember(n.FromSelect.Alias, m.Name)
		if n.Group {
			expr = a.expressionForAnyInGroup(expr)
		}

		n.MeasureFields = append(n.MeasureFields, SelectField{
			Name:  m.Name,
			Label: m.Label,
			Expr:  expr,
		})

		return nil
	}

	// Now we know it's NOT a node with a comparison join.

	// If the referenced measures are not already in scope in the sub-selects, add them.
	// Since the node doesn't have a comparison join, addReferencedMeasuresToScope guarantees the referenced measures are ONLY in scope in ONE sub-query, which prevents ambiguous references in the measure expression.
	err := a.addReferencedMeasuresToScope(n, m.ReferencedMeasures)
	if err != nil {
		return err
	}

	// Add the derived measure expression to the current node.
	expr := a.expressionForMeasure(m, n)
	if n.Group {
		// TODO: There's a risk of expr containing a window, which can't be wrapped by ANY_VALUE. Need to fix it by wrapping with a non-grouped SELECT. Doesn't matter until we implement addDerivedMeasureWithPer.
		expr = a.expressionForAnyInGroup(expr)
	}

	n.MeasureFields = append(n.MeasureFields, SelectField{
		Name:  m.Name,
		Label: m.Label,
		Expr:  expr,
	})

	return nil
}

// addDerivedMeasureWithPer adds a measure of type derived with "per" dimensions to the given MetricsSelect.
// When called, we know the measure is not present in the MetricsSelect, but it might be present in a sub-select.
func (a *AST) addDerivedMeasureWithPer(_ *MetricsSelect, _ *runtimev1.MetricsViewSpec_MeasureV2) error {
	return errors.New(`support for "per" not implemented`)
}

// addTimeComparisonMeasure adds a measure of type time comparison to the given MetricsSelect.
// When called, we know the measure is not present in the MetricsSelect, but it might be present in a sub-select.
func (a *AST) addTimeComparisonMeasure(n *MetricsSelect, m *runtimev1.MetricsViewSpec_MeasureV2) error {
	// If the node doesn't have a comparison join, we wrap it in a new SELECT that we add the comparison join to.
	// We use the hardcoded aliases "base" and "comparison" for the two SELECTs (which must be used in the comparison measure expression).
	if n.JoinComparisonSelect == nil {
		if a.query.ComparisonTimeRange == nil {
			return errors.New("comparison time range not provided")
		}

		a.wrapSelect(n, "base")

		n.JoinComparisonSelect = &MetricsSelect{
			Alias:     "comparison",
			DimFields: a.dimFields,
			Group:     true,
			FromPlain: a.baseSelect,
		}

		a.addTimeRange(n.JoinComparisonSelect, a.query.ComparisonTimeRange)

		n.JoinComparisonType = "FULL OUTER"

		for i, f := range n.DimFields {
			f.Expr = fmt.Sprintf("COALESCE(%s, %s)", f.Expr, a.expressionForMember("comparison", f.Name))
			n.DimFields[i] = f // Because it's not a value, not a pointer
		}
	}

	// Add the referenced measures to the base and comparison SELECTs.
	err := a.addReferencedMeasuresToScope(n, m.ReferencedMeasures)
	if err != nil {
		return err
	}

	// Add the comparison measure expression to the current node.
	expr := a.expressionForMeasure(m, n)
	if n.Group {
		// TODO: There's a risk of expr containing a window, which can't be wrapped by ANY_VALUE. Need to fix it by wrapping with a non-grouped SELECT. Doesn't matter until we implement addDerivedMeasureWithPer.
		// TODO: Can a node with a comparison ever have Group==true?
		expr = a.expressionForAnyInGroup(expr)
	}

	n.MeasureFields = append(n.MeasureFields, SelectField{
		Name:  m.Name,
		Label: m.Label,
		Expr:  expr,
	})

	return nil
}

// addSortField adds a sort field to the given MetricsSelect.
func (a *AST) addSortField(n *MetricsSelect, name string, desc bool) error {
	// We currently only allow sorting by selected dimensions and measures.
	if !a.hasName(n, name) {
		return errors.New("name not present in context")
	}

	n.OrderBy = append(n.OrderBy, OrderByField{
		Name: name,
		Desc: desc,
	})

	return nil
}

// addReferencedMeasuresToScope adds the referenced measures to the node's scope (skipping any that are already in scope).
// If the node has a comparison join, it adds the measures to both the base (FromSelect) and comparison (JoinComparisonSelect) nodes.
// Otherwise, it guarantees that the referenced measures are only in scope in one of the node's sub-select (so no need to worry about ambiguous references).
//
// Note that it does not add the measures to the current node's SELECT list, only to one of its sub-selects such that it's in scope for querying.
func (a *AST) addReferencedMeasuresToScope(n *MetricsSelect, referencedMeasures []string) error {
	if len(referencedMeasures) == 0 {
		return nil
	}

	// If the node targets a plain SELECT, we need to add a new level of nesting.
	// This ensures n.FromSelect is set, so we can bring the referenced measures into scope.
	if n.FromPlain != nil {
		a.wrapSelect(n, a.generateIdentifier())
	}

	for _, rm := range referencedMeasures {
		// Note we pass visible==false because the measure won't be projected into the current node's SELECT list, only brought into scope for derived measures.
		m, err := a.lookupMeasure(rm, false)
		if err != nil {
			return err
		}

		// Add to the base SELECT. addMeasureField skips it if it's already present.
		err = a.addMeasureField(n.FromSelect, m)
		if err != nil {
			return err
		}

		// Add to the comparison SELECT if it exists.
		if n.JoinComparisonSelect != nil {
			err = a.addMeasureField(n.JoinComparisonSelect, m)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// wrapSelect rewrites the given node with a wrapping SELECT that includes the same dimensions and measures as the original node.
// The innerAlias is used as the alias of the inner SELECT in the new outer SELECT.
// Example: wrapSelect("SELECT a, count(*) as b FROM c", "t") -> "SELECT t.a, t.b FROM (SELECT a, count(*) as b FROM c) t".
func (a *AST) wrapSelect(s *MetricsSelect, innerAlias string) {
	cpy := *s
	cpy.Alias = innerAlias

	s.DimFields = make([]SelectField, 0, len(cpy.DimFields))
	for _, f := range cpy.DimFields {
		s.DimFields = append(s.DimFields, SelectField{
			Name:  f.Name,
			Label: f.Label,
			Expr:  a.expressionForMember(cpy.Alias, f.Name),
			// Not copying Unnest because we should only unnest once (at the innermost level).
		})
	}

	s.MeasureFields = make([]SelectField, 0, len(cpy.MeasureFields))
	for _, f := range cpy.MeasureFields {
		s.MeasureFields = append(s.MeasureFields, SelectField{
			Name:  f.Name,
			Label: f.Label,
			Expr:  a.expressionForMember(cpy.Alias, f.Name),
		})
	}

	s.Group = false
	s.FromPlain = nil
	s.FromSelect = &cpy
	s.LeftJoinSelects = nil
	s.JoinComparisonSelect = nil
	s.JoinComparisonType = ""
	s.Where = nil
	s.Having = nil

	s.OrderBy = make([]OrderByField, 0, len(cpy.OrderBy))
	s.OrderBy = append(s.OrderBy, cpy.OrderBy...) // Fresh copy

	s.Limit = nil
	s.Offset = nil
}

// hasName checks if the given name is present as either a dimension or measure field in the node.
// It relies on field names always resolving to the same value regardless of where in the AST they're referenced.
// I.e. a name always corresponds to a dimension/measure name in the metrics view or as a computed field in the query.
// NOTE: Even if it returns false, the measure may still be present in a sub-select (which can happen if it was added as a referenced measure of a derived measure).
func (a *AST) hasName(n *MetricsSelect, name string) bool {
	for _, f := range n.DimFields {
		if f.Name == name {
			return true
		}
	}
	for _, f := range n.MeasureFields {
		if f.Name == name {
			return true
		}
	}
	return false
}

// hasMeasure checks if the given measure name is already present in the given MetricsSelect.
// See hasName for details about name checks.
func (a *AST) hasMeasure(n *MetricsSelect, name string) bool {
	for _, f := range n.MeasureFields {
		if f.Name == name {
			return true
		}
	}
	return false
}

// expressionForTimeRange builds a SQL expression and query args for filtering by a time range.
func (a *AST) expressionForTimeRange(timeCol string, start, end time.Time) (string, []any) {
	var where string
	var args []any
	if !start.IsZero() && !end.IsZero() {
		col := a.dialect.EscapeIdentifier(timeCol)
		where = fmt.Sprintf("%s >= ? AND %s < ?", col, col)
		args = []any{start, end}
	} else if !start.IsZero() {
		where = fmt.Sprintf("%s >= ?", a.dialect.EscapeIdentifier(timeCol))
		args = []any{start}
	} else if end.IsZero() {
		where = fmt.Sprintf("%s < ?", a.dialect.EscapeIdentifier(timeCol))
		args = []any{end}
	} else {
		return "", nil
	}
	return where, args
}

// expressionForMeasure builds a SQL expression for a measure, including its window if present.
// It uses the provided n to resolve dimensions expressions for window partitions.
func (a *AST) expressionForMeasure(m *runtimev1.MetricsViewSpec_MeasureV2, n *MetricsSelect) string {
	// If not applying a window, just return the measure expression.
	if m.Window == nil {
		return m.Expression
	}

	// For windows, we currently have a very hard-coded logic:
	// 1. If partitioning is configured, we partition by all dimensions in the query except the time dimension.
	// 2. If the time dimension is present in the query, we always order by time.

	var partitionExprs []string
	var orderExprs []string
	for _, f := range n.DimFields {
		expr := f.Expr
		if f.Unnest {
			expr = a.expressionForMember(f.UnnestAlias, f.Name)
		}

		isTimeDimension := f.Name == a.metricsView.TimeDimension
		if !isTimeDimension {
			// The field name may not be the time dim itself, but an alias for the floored time dimension.
			// Search the query dimensions to check if that's the case.
			for _, qd := range a.query.Dimensions {
				if f.Name == qd.Name {
					if qd.Compute != nil && qd.Compute.TimeFloor != nil {
						isTimeDimension = true
					}
					break
				}
			}
		}

		if isTimeDimension {
			orderExprs = append(orderExprs, expr)
			continue
		}

		if m.Window.Partition {
			partitionExprs = append(partitionExprs, expr)
		}
	}

	var partitionClause string
	if len(partitionExprs) > 0 && len(orderExprs) > 0 {
		partitionClause = fmt.Sprintf("PARTITION BY %s ORDER BY %s", strings.Join(partitionExprs, ", "), strings.Join(orderExprs, ", "))
	} else if len(partitionExprs) > 0 {
		partitionClause = fmt.Sprintf("PARTITION BY %s", strings.Join(partitionExprs, ", "))
	} else if len(orderExprs) > 0 {
		partitionClause = fmt.Sprintf("ORDER BY %s", strings.Join(orderExprs, ", "))
	}

	return fmt.Sprintf("%s OVER (%s %s)", m.Expression, partitionClause, m.Window.FrameExpression)
}

// expressionForMember builds a SQL expression for a field in a table.
// It does not escape the tbl identifier because we currently only use it for internally generated aliases.
func (a *AST) expressionForMember(tbl, name string) string {
	if tbl == "" {
		return a.dialect.EscapeIdentifier(name)
	}
	return fmt.Sprintf("%s.%s", tbl, a.dialect.EscapeIdentifier(name))
}

// expressionForAnyInGroup returns a SQL expression for passing through a field in a GROUP BY.
func (a *AST) expressionForAnyInGroup(expr string) string {
	return fmt.Sprintf("ANY_VALUE(%s)", expr)
}

// generateIdentifier generates a unique table identifier for use in the AST.
func (a *AST) generateIdentifier() string {
	tmp := fmt.Sprintf("t%d", a.nextIdentifier)
	a.nextIdentifier++
	return tmp
}
