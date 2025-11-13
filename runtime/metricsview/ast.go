package metricsview

import (
	"errors"
	"fmt"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/timeutil"
)

// ErrForbidden is returned when a query violates the constraints of MetricsViewSecurity.
var ErrForbidden = errors.New("action not allowed")

// MetricsViewSecurity defines access restrictions to a metrics view.
// The interface is currently a subset of the *runtime.ResolvedSecurity concrete type.
type MetricsViewSecurity interface {
	CanAccessField(field string) bool
	RowFilter() string
	QueryFilter() *runtimev1.Expression
}

// AST is the abstract syntax tree for a metrics SQL query.
type AST struct {
	// Root of the AST
	Root *SelectNode
	// List of CTEs to add to the query
	CTEs []*SelectNode

	// Contextual info for building the AST
	MetricsView *runtimev1.MetricsViewSpec
	Security    MetricsViewSecurity
	Query       *Query
	Dialect     drivers.Dialect

	// Cached internal state for building the AST
	underlyingTable     *string
	underlyingWhere     *ExprNode
	dimFields           []FieldNode
	comparisonDimFields []FieldNode
	unnests             []string
	nextIdentifier      int
}

// SelectNode represents a query that computes measures by dimensions.
// The from/join clauses are not all compatible. The allowed combinations are:
//   - FromTable
//   - FromSelect and optionally SpineSelect and/or LeftJoinSelects
//   - FromSelect and optionally JoinComparisonSelect (for comparison CTE based optimization, this combination is used, both should be set and one of them will be used as CTE)
type SelectNode struct {
	RawSelect            *ExprNode        // Raw SQL SELECT statement to use
	Alias                string           // Alias for the node used by outer SELECTs to reference it.
	IsCTE                bool             // Whether this node is a Common Table Expression
	DimFields            []FieldNode      // Dimensions fields to select
	MeasureFields        []FieldNode      // Measure fields to select
	FromTable            *string          // Underlying table expression to select from (if set, FromSelect must not be set)
	FromSelect           *SelectNode      // Sub-select to select from (if set, FromTable must not be set)
	SpineSelect          *SelectNode      // Sub-select that returns a spine of dimensions. Currently it will be right-joined onto FromSelect.
	LeftJoinSelects      []*SelectNode    // Sub-selects to left join onto FromSelect, to enable "per-dimension" measures
	CrossJoinSelects     []*SelectNode    // sub-selects to cross join onto FromSelect
	JoinComparisonSelect *SelectNode      // Sub-select to join onto FromSelect for comparison measures
	JoinComparisonType   JoinType         // Type of join to use for JoinComparisonSelect
	Unnests              []string         // Unnest expressions to add in the FROM clause
	Group                bool             // Whether the SELECT is grouped. If yes, it will group by all DimFields.
	Where                *ExprNode        // Expression for the WHERE clause
	TimeWhere            *ExprNode        // Expression for the time range to add to the WHERE clause
	Having               *ExprNode        // Expression for the HAVING clause. If HAVING is not allowed in the current context, it will added as a WHERE in a wrapping SELECT.
	OrderBy              []OrderFieldNode // Fields to order by
	Limit                *int64           // Limit for the query
	Offset               *int64           // Offset for the query
}

// HasName checks if the given name is present as either a dimension or measure field in the node.
// It relies on field names always resolving to the same value regardless of where in the AST they're referenced.
// I.e. a name always corresponds to a dimension/measure name in the metrics view or as a computed field in the query.
// NOTE: Even if it returns false, the measure may still be present in a sub-select (which can happen if it was added as a referenced measure of a derived measure).
func (n *SelectNode) HasName(name string) bool {
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

// HasMeasure checks if the given measure name is already present in the given SelectNode.
// See hasName for details about name checks.
func (n *SelectNode) HasMeasure(name string) bool {
	for _, f := range n.MeasureFields {
		if f.Name == name {
			return true
		}
	}
	return false
}

// FieldNode represents a column in a SELECT clause. It also carries metadata related to the dimension/measure it was derived from.
// The Name must always match a the name of a dimension/measure in the metrics view or a computed field specified in the request.
// This means that if two columns in different places in the AST have the same Name, they're guaranteed to resolve to the same value.
type FieldNode struct {
	Name        string
	DisplayName string
	Expr        string
	Unnest      bool
	TreatNullAs string // only used for measures
}

// ExprNode represents an expression for a WHERE clause.
type ExprNode struct {
	Expr string
	Args []any
}

// And returns a new node that is the AND of the current node and the given expression.
func (n *ExprNode) And(expr string, args []any) *ExprNode {
	if expr == "" {
		return n
	}

	if n == nil || n.Expr == "" {
		return &ExprNode{
			Expr: expr,
			Args: args,
		}
	}

	return &ExprNode{
		Expr: fmt.Sprintf("(%s) AND (%s)", n.Expr, expr),
		Args: append(n.Args, args...),
	}
}

// OrderFieldNode represents a field in an ORDER BY clause.
type OrderFieldNode struct {
	Name string
	Desc bool
}

// JoinType represents types of SQL joins.
type JoinType string

const (
	JoinTypeUnspecified JoinType = ""
	JoinTypeFull        JoinType = "FULL OUTER"
	JoinTypeLeft        JoinType = "LEFT OUTER"
	JoinTypeRight       JoinType = "RIGHT OUTER"
	JoinTypeCross       JoinType = "CROSS"
)

// NewAST builds a new SQL AST based on a metrics query.
//
// Dynamic time ranges in the qry must be resolved to static start/end timestamps before calling this function.
// This is due to NewAST not being able (or intended) to resolve external time anchors such as watermarks.
//
// The qry's PivotOn must be empty. Pivot queries must be rewritten/handled upstream of NewAST.
func NewAST(mv *runtimev1.MetricsViewSpec, sec MetricsViewSecurity, qry *Query, dialect drivers.Dialect) (*AST, error) {
	// Validation
	if len(qry.PivotOn) > 0 {
		return nil, errors.New("cannot build AST for pivot queries")
	}
	if len(qry.Dimensions) == 0 && len(qry.Measures) == 0 && !qry.Rows {
		return nil, fmt.Errorf("must specify at least one dimension or measure")
	}

	// Use provided time column if available, otherwise fall back to TimeDimension
	timeDim := mv.TimeDimension
	if qry.TimeRange != nil && qry.TimeRange.TimeDimension != "" {
		timeDim = qry.TimeRange.TimeDimension
	}

	// Init
	ast := &AST{
		MetricsView: mv,
		Security:    sec,
		Query:       qry,
		Dialect:     dialect,
	}

	// Determine the minimum time grain for the time dimension in the query.
	minGrain := TimeGrainUnspecified
	for _, qd := range ast.Query.Dimensions {
		if qd.Compute == nil || qd.Compute.TimeFloor == nil {
			continue
		}
		if !strings.EqualFold(qd.Compute.TimeFloor.Dimension, timeDim) {
			continue
		}
		tg := qd.Compute.TimeFloor.Grain
		if tg == TimeGrainUnspecified {
			continue
		}
		if minGrain == TimeGrainUnspecified {
			minGrain = tg
			continue
		}
		if tg.ToTimeutil() < minGrain.ToTimeutil() {
			minGrain = tg
		}
	}

	// Build dimensions to apply against the underlying SELECT.
	// We cache these in the AST type because when resolving expressions and adding new JOINs, we need the ability to reference these.
	ast.dimFields = make([]FieldNode, 0, len(ast.Query.Dimensions))
	ast.comparisonDimFields = make([]FieldNode, 0, len(ast.Query.Dimensions))
	for _, qd := range ast.Query.Dimensions {
		dim, err := ast.ResolveDimension(qd, true)
		if err != nil {
			return nil, fmt.Errorf("invalid dimension %q: %w", qd.Name, err)
		}

		expr, err := ast.Dialect.MetricsViewDimensionExpression(dim)
		if err != nil {
			return nil, fmt.Errorf("failed to compile dimension %q expression: %w", dim.Name, err)
		}

		f := FieldNode{
			Name:        dim.Name,
			DisplayName: dim.DisplayName,
			Expr:        expr,
			Unnest:      dim.Unnest,
		}

		if dim.Unnest {
			unnestAlias := ast.GenerateIdentifier()

			tblWithAlias, tupleStyle, auto, err := ast.Dialect.LateralUnnest(f.Expr, unnestAlias, f.Name)
			if err != nil {
				return nil, fmt.Errorf("failed to unnest field %q: %w", f.Name, err)
			}

			if !auto {
				ast.unnests = append(ast.unnests, tblWithAlias)
				if tupleStyle {
					f.Expr = ast.Dialect.EscapeMember(unnestAlias, f.Name)
				} else {
					f.Expr = ast.Dialect.EscapeMember("", f.Name)
				}
			}
		}

		// If a comparison time range is provided and the time dimension is in DimFields,
		// we need to add the time interval between the base and comparison time ranges to the time dimension expression in the comparison sub-query.
		// This makes the time dimension values comparable across the base and comparison select, so they can be joined on.
		// Note that estimating the time interval between the two time ranges is best effort and may not always be possible.
		// Also note that the comparison time range currently always targets a.metricsView.TimeDimension, so we only apply the correction for that dimension.
		cf := f // Clone
		if ast.Query.ComparisonTimeRange != nil && qd.Compute != nil && qd.Compute.TimeFloor != nil {
			if strings.EqualFold(qd.Compute.TimeFloor.Dimension, timeDim) {
				cf.Expr, err = ast.sqlForExpressionAdjustedByComparisonTimeRangeOffset(f.Expr, timeDim, qd.Compute.TimeFloor.Grain, minGrain)
				if err != nil {
					return nil, err
				}
			}
		}

		ast.dimFields = append(ast.dimFields, f)
		ast.comparisonDimFields = append(ast.comparisonDimFields, cf)
	}

	if qry.Rows {
		// when Rows is set we want underlying rows from the model, adding only * as the dim field, query validation is done earlier which disallows any dimensions
		ast.dimFields = append(ast.dimFields, FieldNode{
			Name: "*",
			Expr: "*",
		})
	}

	// Build underlying SELECT
	tbl := ast.Dialect.EscapeTable(mv.Database, mv.DatabaseSchema, mv.Table)
	where, err := ast.buildWhereForUnderlyingTable(ast.Query.Where)
	if err != nil {
		return nil, err
	}
	ast.underlyingTable = &tbl
	ast.underlyingWhere = where

	// Build initial root node (empty query against the base select)
	n, err := ast.buildBaseSelect(ast.GenerateIdentifier(), false)
	if err != nil {
		return nil, err
	}
	ast.Root = n

	// Incrementally add each output measure.
	// As each measure is added, the AST is transformed to accommodate it based on its type.
	for _, qm := range ast.Query.Measures {
		m, err := ast.ResolveMeasure(qm, true)
		if err != nil {
			return nil, fmt.Errorf("invalid measure %q: %w", qm.Name, err)
		}

		err = ast.AddMeasureField(ast.Root, m)
		if err != nil {
			return nil, fmt.Errorf("can't query measure %q: %w", qm.Name, err)
		}
	}

	// Handle Having. If the root node is grouped, we add it as a HAVING clause, otherwise wrap it in a SELECT and add it as a WHERE clause.
	if ast.Query.Having != nil {
		// We need to wrap in a new SELECT because a WHERE/HAVING clause cannot apply directly to a field with a window function.
		// This also enables us to template the field name instead of the field expression into the expression.
		ast.WrapSelect(ast.Root, ast.GenerateIdentifier())

		expr, args, err := ast.SQLForExpression(ast.Query.Having, ast.Root, true, true)
		if err != nil {
			return nil, fmt.Errorf("failed to compile 'having': %w", err)
		}
		res := &ExprNode{
			Expr: expr,
			Args: args,
		}

		ast.Root.Where = res
	}

	// Incrementally add each sort criterion.
	for _, s := range ast.Query.Sort {
		err = ast.addOrderField(ast.Root, s.Name, s.Desc)
		if err != nil {
			return nil, fmt.Errorf("can't sort by %q: %w", s.Name, err)
		}
	}

	// Add limit and offset
	ast.Root.Limit = ast.Query.Limit
	ast.Root.Offset = ast.Query.Offset

	return ast, nil
}

// ResolveDimension returns a dimension spec for the given dimension query.
// If the dimension query specifies a computed dimension, it constructs a dimension spec to match it.
func (a *AST) ResolveDimension(qd Dimension, visible bool) (*runtimev1.MetricsViewSpec_Dimension, error) {
	// Handle regular dimension
	if qd.Compute == nil {
		return a.LookupDimension(qd.Name, visible)
	}

	// Handle computed dimension. This means "compute.time_floor" must be configured.

	if qd.Compute.TimeFloor == nil {
		return nil, errors.New(`unsupported "compute"`)
	}

	if !qd.Compute.TimeFloor.Grain.Valid() {
		return nil, errors.New(`invalid "grain"`)
	}
	if qd.Compute.TimeFloor.Grain == TimeGrainUnspecified {
		return nil, errors.New(`"grain" must be specified for time floor`)
	}

	dim, err := a.LookupDimension(qd.Compute.TimeFloor.Dimension, visible)
	if err != nil {
		return nil, err
	}

	if qd.Name != qd.Compute.TimeFloor.Dimension {
		err := a.checkNameForComputedField(qd.Name)
		if err != nil {
			return nil, err
		}
	}

	grain := qd.Compute.TimeFloor.Grain.ToProto()
	expr, err := a.Dialect.DateTruncExpr(dim, grain, a.Query.TimeZone, int(a.MetricsView.FirstDayOfWeek), int(a.MetricsView.FirstMonthOfYear))
	if err != nil {
		return nil, fmt.Errorf(`failed to compute time floor: %w`, err)
	}

	displayName := dim.DisplayName
	if displayName == "" {
		displayName = qd.Name
	}
	displayName = fmt.Sprintf("%s (%s)", displayName, qd.Compute.TimeFloor.Grain)

	return &runtimev1.MetricsViewSpec_Dimension{
		Name:        qd.Name,
		Expression:  expr,
		DisplayName: displayName,
		Unnest:      dim.Unnest,
	}, nil
}

// ResolveMeasure returns a measure spec for the given measure query.
// If the measure query specifies a computed measure, it constructs a measure spec to match it.
func (a *AST) ResolveMeasure(qm Measure, visible bool) (*runtimev1.MetricsViewSpec_Measure, error) {
	if qm.Compute == nil {
		return a.LookupMeasure(qm.Name, visible)
	}

	if err := qm.Compute.Validate(); err != nil {
		return nil, fmt.Errorf(`invalid "compute": %w`, err)
	}

	err := a.checkNameForComputedField(qm.Name)
	if err != nil {
		return nil, err
	}

	if qm.Compute.Count {
		return &runtimev1.MetricsViewSpec_Measure{
			Name:        qm.Name,
			Expression:  "COUNT(*)",
			Type:        runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE,
			DisplayName: "Count",
		}, nil
	}

	if qm.Compute.CountDistinct != nil {
		dim, err := a.LookupDimension(qm.Compute.CountDistinct.Dimension, visible)
		if err != nil {
			return nil, err
		}

		expr := dim.Expression
		if expr == "" {
			expr = a.Dialect.EscapeIdentifier(dim.Column)
		}

		return &runtimev1.MetricsViewSpec_Measure{
			Name:        qm.Name,
			Expression:  fmt.Sprintf("COUNT(DISTINCT %s)", expr),
			Type:        runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE,
			DisplayName: fmt.Sprintf("Unique %s", dim.DisplayName),
		}, nil
	}

	if qm.Compute.ComparisonValue != nil {
		m, err := a.LookupMeasure(qm.Compute.ComparisonValue.Measure, visible)
		if err != nil {
			return nil, err
		}

		expr := fmt.Sprintf("comparison.%s", a.Dialect.EscapeIdentifier(m.Name))
		if m.TreatNullsAs != "" {
			expr = fmt.Sprintf("COALESCE(%s, %s)", expr, m.TreatNullsAs)
		}

		return &runtimev1.MetricsViewSpec_Measure{
			Name:               qm.Name,
			Expression:         expr,
			Type:               runtimev1.MetricsViewSpec_MEASURE_TYPE_TIME_COMPARISON,
			ReferencedMeasures: []string{qm.Compute.ComparisonValue.Measure},
			DisplayName:        fmt.Sprintf("%s (prev)", m.DisplayName),
		}, nil
	}

	if qm.Compute.ComparisonDelta != nil {
		m, err := a.LookupMeasure(qm.Compute.ComparisonDelta.Measure, visible)
		if err != nil {
			return nil, err
		}

		compareExpr := fmt.Sprintf("comparison.%s", a.Dialect.EscapeIdentifier(m.Name))
		if m.TreatNullsAs != "" {
			compareExpr = fmt.Sprintf("COALESCE(%s, %s)", compareExpr, m.TreatNullsAs)
		}

		return &runtimev1.MetricsViewSpec_Measure{
			Name:               qm.Name,
			Expression:         fmt.Sprintf("base.%s - %s", a.Dialect.EscapeIdentifier(m.Name), compareExpr),
			Type:               runtimev1.MetricsViewSpec_MEASURE_TYPE_TIME_COMPARISON,
			ReferencedMeasures: []string{qm.Compute.ComparisonDelta.Measure},
			DisplayName:        fmt.Sprintf("%s (Δ)", m.DisplayName),
		}, nil
	}

	if qm.Compute.ComparisonRatio != nil {
		m, err := a.LookupMeasure(qm.Compute.ComparisonRatio.Measure, visible)
		if err != nil {
			return nil, err
		}

		base := fmt.Sprintf("base.%s", a.Dialect.EscapeIdentifier(m.Name))
		compareExpr := fmt.Sprintf("comparison.%s", a.Dialect.EscapeIdentifier(m.Name))
		if m.TreatNullsAs != "" {
			compareExpr = fmt.Sprintf("COALESCE(%s, %s)", compareExpr, m.TreatNullsAs)
		}
		expr := a.Dialect.SafeDivideExpression(fmt.Sprintf("%s - %s", base, compareExpr), compareExpr)

		return &runtimev1.MetricsViewSpec_Measure{
			Name:               qm.Name,
			Expression:         expr,
			Type:               runtimev1.MetricsViewSpec_MEASURE_TYPE_TIME_COMPARISON,
			ReferencedMeasures: []string{qm.Compute.ComparisonRatio.Measure},
			DisplayName:        fmt.Sprintf("%s (Δ%%)", m.DisplayName),
		}, nil
	}

	if qm.Compute.PercentOfTotal != nil {
		if qm.Compute.PercentOfTotal.Total == nil {
			return nil, fmt.Errorf("totals not computed for %s", qm.Name)
		}

		m, err := a.LookupMeasure(qm.Compute.PercentOfTotal.Measure, visible)
		if err != nil {
			return nil, err
		}

		return &runtimev1.MetricsViewSpec_Measure{
			Name:               qm.Name,
			Expression:         fmt.Sprintf("%s/%#f", a.Dialect.EscapeIdentifier(m.Name), *qm.Compute.PercentOfTotal.Total),
			Type:               runtimev1.MetricsViewSpec_MEASURE_TYPE_DERIVED,
			ReferencedMeasures: []string{qm.Compute.PercentOfTotal.Measure},
			DisplayName:        fmt.Sprintf("%s (Σ%%)", m.DisplayName),
		}, nil
	}

	if qm.Compute.URI != nil {
		dim, err := a.LookupDimension(qm.Compute.URI.Dimension, visible)
		if err != nil {
			return nil, err
		}

		uri := dim.Uri
		if uri == "" {
			return nil, fmt.Errorf("`uri` not set for the dimension %v", qm.Compute.URI.Dimension)
		}

		return &runtimev1.MetricsViewSpec_Measure{
			Name:        qm.Name,
			Expression:  a.Dialect.AnyValueExpression(uri),
			Type:        runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE,
			DisplayName: fmt.Sprintf("URI for %s", dim.DisplayName),
		}, nil
	}

	if qm.Compute.ComparisonTime != nil {
		if a.Query.ComparisonTimeRange == nil || (a.Query.ComparisonTimeRange.TimeDimension != "" && a.Query.ComparisonTimeRange.TimeDimension != qm.Compute.ComparisonTime.Dimension) || a.MetricsView.TimeDimension != qm.Compute.ComparisonTime.Dimension {
			return nil, fmt.Errorf("comparison time measure %q must be based on the metrics view's time dimension %q or the query's comparison time dimension %q", qm.Name, a.MetricsView.TimeDimension, a.Query.ComparisonTimeRange.TimeDimension)
		}

		// find the base computed time dimension from the query
		var qd *Dimension
		for _, q := range a.Query.Dimensions {
			if q.Compute != nil && q.Compute.TimeFloor != nil && strings.EqualFold(q.Compute.TimeFloor.Dimension, qm.Compute.ComparisonTime.Dimension) {
				qd = &q
				break
			}
		}
		if qd == nil {
			return nil, fmt.Errorf("comparison time measure %q must be based on a computed time dimension in the query", qm.Name)
		}

		baseStart := a.Query.TimeRange.Start
		compStart := a.Query.ComparisonTimeRange.Start

		dateDiff, err := a.Dialect.DateDiff(qd.Compute.TimeFloor.Grain.ToProto(), compStart, baseStart)
		if err != nil {
			return nil, fmt.Errorf("failed to compute date difference for comparison time measure %q: %w", qm.Name, err)
		}

		baseExpr := fmt.Sprintf("COALESCE(base.%s, comparison.%s)", a.Dialect.EscapeIdentifier(qd.Name), a.Dialect.EscapeIdentifier(qd.Name))

		expr, err := a.Dialect.IntervalSubtract(baseExpr, dateDiff, qd.Compute.TimeFloor.Grain.ToProto())
		if err != nil {
			return nil, fmt.Errorf("failed to compute comparison time measure %q expression: %w", qm.Name, err)
		}

		return &runtimev1.MetricsViewSpec_Measure{
			Name:        qm.Name,
			Expression:  expr,
			Type:        runtimev1.MetricsViewSpec_MEASURE_TYPE_TIME_COMPARISON,
			DisplayName: fmt.Sprintf("Comparison time for %s", qd.Name),
		}, nil
	}

	return nil, errors.New("unhandled compute operation")
}

// lookupDimension finds a dimension spec in the metrics view.
// If visible is true, it returns an error if the security policy does not grant access to the dimension.
func (a *AST) LookupDimension(name string, visible bool) (*runtimev1.MetricsViewSpec_Dimension, error) {
	if name == "" {
		return nil, errors.New("received empty dimension name")
	}

	// not checking access if its primary time dimension
	if name == a.MetricsView.TimeDimension {
		// check if its defined in the dimensions or time dimension list otherwise return a default dimension spec
		for _, dim := range a.MetricsView.Dimensions {
			if dim.Name == name {
				return dim, nil
			}
		}
		return &runtimev1.MetricsViewSpec_Dimension{
			Name:   name,
			Column: name,
		}, nil
	}

	if visible {
		if !a.Security.CanAccessField(name) {
			return nil, ErrForbidden
		}
	}

	for _, dim := range a.MetricsView.Dimensions {
		if dim.Name == name {
			return dim, nil
		}
	}

	return nil, fmt.Errorf("dimension %q not found", name)
}

// lookupMeasure finds a measure spec in the metrics view.
// If visible is true, it returns an error if the security policy does not grant access to the measure.
func (a *AST) LookupMeasure(name string, visible bool) (*runtimev1.MetricsViewSpec_Measure, error) {
	if visible {
		if !a.Security.CanAccessField(name) {
			return nil, ErrForbidden
		}
	}

	for _, m := range a.MetricsView.Measures {
		if m.Name == name {
			return m, nil
		}
	}

	return nil, fmt.Errorf("measure %q not found", name)
}

// GenerateIdentifier generates a unique table identifier for use in the AST.
func (a *AST) GenerateIdentifier() string {
	tmp := fmt.Sprintf("t%d", a.nextIdentifier)
	a.nextIdentifier++
	return tmp
}

// WrapSelect rewrites the given node with a wrapping SELECT that includes the same dimensions and measures as the original node.
// The innerAlias is used as the alias of the inner SELECT in the new outer SELECT.
// Example: wrapSelect("SELECT a, count(*) as b FROM c", "t") -> "SELECT t.a, t.b FROM (SELECT a, count(*) as b FROM c) t".
func (a *AST) WrapSelect(s *SelectNode, innerAlias string) {
	cpy := *s
	cpy.Alias = innerAlias

	s.DimFields = make([]FieldNode, 0, len(cpy.DimFields))
	for _, f := range cpy.DimFields {
		s.DimFields = append(s.DimFields, FieldNode{
			Name:        f.Name,
			DisplayName: f.DisplayName,
			Expr:        a.Dialect.EscapeMember(cpy.Alias, f.Name),
		})
	}

	for _, cjs := range cpy.CrossJoinSelects {
		for _, f := range cjs.DimFields {
			s.DimFields = append(s.DimFields, FieldNode{
				Name:        f.Name,
				DisplayName: f.DisplayName,
				Expr:        a.Dialect.EscapeMember(cpy.Alias, f.Name),
			})
		}
	}

	s.MeasureFields = make([]FieldNode, 0, len(cpy.MeasureFields))
	for _, f := range cpy.MeasureFields {
		s.MeasureFields = append(s.MeasureFields, FieldNode{
			Name:        f.Name,
			DisplayName: f.DisplayName,
			Expr:        a.Dialect.EscapeMember(cpy.Alias, f.Name),
			TreatNullAs: "",
		})
	}

	s.FromTable = nil
	s.FromSelect = &cpy
	s.SpineSelect = nil
	s.LeftJoinSelects = nil
	s.JoinComparisonSelect = nil
	s.JoinComparisonType = JoinTypeUnspecified
	s.Unnests = nil
	s.Group = false
	s.Where = nil
	s.TimeWhere = nil
	s.Having = nil

	s.OrderBy = make([]OrderFieldNode, 0, len(cpy.OrderBy))
	s.OrderBy = append(s.OrderBy, cpy.OrderBy...) // Fresh copy

	s.Limit = nil
	s.Offset = nil
	s.CrossJoinSelects = nil
}

// ConvertToCTE util func that sets IsCTE and only adds to a.CTEs if IsCTE was false
func (a *AST) ConvertToCTE(n *SelectNode) {
	if n.IsCTE {
		return
	}

	n.IsCTE = true
	a.CTEs = append(a.CTEs, n)
}

// AddTimeRange adds a time range to the given SelectNode's WHERE clause.
func (a *AST) AddTimeRange(n *SelectNode, tr *TimeRange) error {
	if tr == nil || tr.IsZero() || (a.MetricsView.TimeDimension == "" && tr.TimeDimension == "") {
		return nil
	}

	// Since resolving time ranges may require contextual info (like watermarks), the upstream caller is responsible for resolving them.
	if tr.Start.IsZero() && tr.End.IsZero() {
		panic("ast received a non-empty, unresolved time range")
	}

	timeDimExpr, err := a.getTimeDimensionExpression(tr)
	if err != nil {
		return err
	}

	expr, args := a.sqlForTimeRange(timeDimExpr, tr.Start, tr.End)
	n.TimeWhere = &ExprNode{
		Expr: expr,
		Args: args,
	}

	return nil
}

// AddMeasureField adds a measure field to the given SelectNode.
// Depending on the measure type, it may rewrite the SelectNode to accommodate the measure.
func (a *AST) AddMeasureField(n *SelectNode, m *runtimev1.MetricsViewSpec_Measure) error {
	// Skip if the measure has already been added.
	// This can happen if the measure was already added as a referenced measure of a derived measure.
	if n.HasMeasure(m.Name) {
		return nil
	}

	// Check that the measure's required dimensions are satisfied
	err := a.checkRequiredDimensionsPresentInQuery(m)
	if err != nil {
		return err
	}

	switch m.Type {
	case runtimev1.MetricsViewSpec_MEASURE_TYPE_UNSPECIFIED, runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE:
		err = a.addSimpleMeasure(n, m)
	case runtimev1.MetricsViewSpec_MEASURE_TYPE_DERIVED:
		err = a.addDerivedMeasure(n, m)
	case runtimev1.MetricsViewSpec_MEASURE_TYPE_TIME_COMPARISON:
		err = a.addTimeComparisonMeasure(n, m)
	default:
		panic("unhandled measure type")
	}

	if m.TreatNullsAs != "" {
		n.MeasureFields[len(n.MeasureFields)-1].TreatNullAs = m.TreatNullsAs
	}

	return err
}

// addSimpleMeasure adds a measure of type simple to the given SelectNode.
// When called, we know the measure is not present in the SelectNode, but it might be present in a sub-select.
func (a *AST) addSimpleMeasure(n *SelectNode, m *runtimev1.MetricsViewSpec_Measure) error {
	// Base case: it targets the underlying table.
	// Add the measure directly to the SELECT list.
	if n.FromTable != nil {
		expr, err := a.sqlForMeasure(m, n)
		if err != nil {
			return err
		}

		n.MeasureFields = append(n.MeasureFields, FieldNode{
			Name:        m.Name,
			DisplayName: m.DisplayName,
			Expr:        expr,
		})

		return nil
	}

	// Recursive case: it targets another SelectNode.
	// We recurse on the sub-select and add a pass-through field to the current node.

	if !n.FromSelect.HasMeasure(m.Name) { // Don't recurse if already in scope in sub-query
		err := a.addSimpleMeasure(n.FromSelect, m)
		if err != nil {
			return err
		}
	}

	expr := a.Dialect.EscapeMember(n.FromSelect.Alias, m.Name)
	if n.Group {
		expr = a.Dialect.AnyValueExpression(expr)
	}

	n.MeasureFields = append(n.MeasureFields, FieldNode{
		Name:        m.Name,
		DisplayName: m.DisplayName,
		Expr:        expr,
	})

	return nil
}

// addDerivedMeasure adds a measure of type derived to the given SelectNode.
// When called, we know the measure is not present in the SelectNode, but it might be present in a sub-select.
func (a *AST) addDerivedMeasure(n *SelectNode, m *runtimev1.MetricsViewSpec_Measure) error {
	// Handle derived measures with "per" dimensions separately.
	if len(m.PerDimensions) > 0 {
		return a.addDerivedMeasureWithPer(n, m)
	}

	// If the current node has a comparison join, push calculation of the derived measure into its FromSelect and add a pass-through field in the current node.
	// This avoids a potential ambiguity issue because the derived measure expression does not use "base.name" and "comparison.name" to identify referenced measures,
	// so we need to ensure the referenced names exist only in ONE sub-query.
	if n.JoinComparisonSelect != nil {
		if !n.FromSelect.HasMeasure(m.Name) {
			err := a.addDerivedMeasure(n.FromSelect, m)
			if err != nil {
				return err
			}
		}

		expr := a.Dialect.EscapeMember(n.FromSelect.Alias, m.Name)
		if n.Group {
			expr = a.Dialect.AnyValueExpression(expr)
		}

		n.MeasureFields = append(n.MeasureFields, FieldNode{
			Name:        m.Name,
			DisplayName: m.DisplayName,
			Expr:        expr,
		})

		return nil
	}
	// Now we know it's NOT a node with a comparison join.

	// If the current node has a spine join, we wrap it in a new SELECT that we add the derived measure to.
	// Even though the spine join won't add any ambiguous measure names, it will make dimension names ambiguous.
	// So in case the derived measure references a dimension by name (unlikely, but possible), we need to ensure the dimension is only in scope in one sub-query.
	if n.SpineSelect != nil {
		a.WrapSelect(n, a.GenerateIdentifier())
	}
	// Now we know it's ALSO NOT a node with a spine join.

	// If the referenced measures are not already in scope in the sub-selects, add them.
	// Since the node doesn't have a comparison join, addReferencedMeasuresToScope guarantees the referenced measures are ONLY in scope in ONE sub-query, which prevents ambiguous references in the measure expression.
	err := a.addReferencedMeasuresToScope(n, m.ReferencedMeasures)
	if err != nil {
		return err
	}

	// Add the derived measure expression to the current node.
	expr, err := a.sqlForMeasure(m, n)
	if err != nil {
		return err
	}
	if n.Group {
		// TODO: There's a risk of expr containing a window, which can't be wrapped by ANY_VALUE. Need to fix it by wrapping with a non-grouped SELECT. Doesn't matter until we implement addDerivedMeasureWithPer.
		expr = a.Dialect.AnyValueExpression(expr)
	}

	n.MeasureFields = append(n.MeasureFields, FieldNode{
		Name:        m.Name,
		DisplayName: m.DisplayName,
		Expr:        expr,
	})

	return nil
}

// addDerivedMeasureWithPer adds a measure of type derived with "per" dimensions to the given SelectNode.
// When called, we know the measure is not present in the SelectNode, but it might be present in a sub-select.
func (a *AST) addDerivedMeasureWithPer(_ *SelectNode, _ *runtimev1.MetricsViewSpec_Measure) error {
	return errors.New(`support for "per" not implemented`)
}

// addTimeComparisonMeasure adds a measure of type time comparison to the given SelectNode.
// When called, we know the measure is not present in the SelectNode, but it might be present in a sub-select.
func (a *AST) addTimeComparisonMeasure(n *SelectNode, m *runtimev1.MetricsViewSpec_Measure) error {
	// If the node doesn't have a comparison join, we wrap it in a new SELECT that we add the comparison join to.
	// We use the hardcoded aliases "base" and "comparison" for the two SELECTs (which must be used in the comparison measure expression).
	if n.JoinComparisonSelect == nil {
		if a.Query.ComparisonTimeRange == nil {
			return errors.New("comparison time range not provided")
		}

		a.WrapSelect(n, "base")

		csn, err := a.buildBaseSelect("comparison", true)
		if err != nil {
			return err
		}
		n.JoinComparisonSelect = csn

		n.JoinComparisonType = JoinTypeFull

		for i, f := range n.DimFields {
			f.Expr = fmt.Sprintf("COALESCE(%s, %s)", f.Expr, a.Dialect.EscapeMember("comparison", f.Name))
			n.DimFields[i] = f // Because it's not a value, not a pointer
		}
	}

	// Add the referenced measures to the base and comparison SELECTs.
	err := a.addReferencedMeasuresToScope(n, m.ReferencedMeasures)
	if err != nil {
		return err
	}

	// Add the comparison measure expression to the current node.
	expr, err := a.sqlForMeasure(m, n)
	if err != nil {
		return err
	}
	if n.Group {
		// TODO: There's a risk of expr containing a window, which can't be wrapped by ANY_VALUE. Need to fix it by wrapping with a non-grouped SELECT. Doesn't matter until we implement addDerivedMeasureWithPer.
		// TODO: Can a node with a comparison ever have Group==true?
		expr = a.Dialect.AnyValueExpression(expr)
	}

	n.MeasureFields = append(n.MeasureFields, FieldNode{
		Name:        m.Name,
		DisplayName: m.DisplayName,
		Expr:        expr,
	})

	return nil
}

// addOrderField adds a sort field to the given SelectNode.
func (a *AST) addOrderField(n *SelectNode, name string, desc bool) error {
	// We currently only allow sorting by selected dimensions and measures.
	if !n.HasName(name) {
		return errors.New("name not present in context")
	}

	n.OrderBy = append(n.OrderBy, OrderFieldNode{
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
func (a *AST) addReferencedMeasuresToScope(n *SelectNode, referencedMeasures []string) error {
	if len(referencedMeasures) == 0 {
		return nil
	}

	// If the node targets the underlying table, we need to add a new level of nesting.
	// This ensures n.FromSelect is set, so we can bring the referenced measures into scope.
	if n.FromTable != nil {
		a.WrapSelect(n, a.GenerateIdentifier())
	}

	for _, rm := range referencedMeasures {
		// Note we pass visible==false because the measure won't be projected into the current node's SELECT list, only brought into scope for derived measures.
		m, err := a.LookupMeasure(rm, false)
		if err != nil {
			return err
		}

		// Add to the base SELECT. addMeasureField skips it if it's already present.
		err = a.AddMeasureField(n.FromSelect, m)
		if err != nil {
			return err
		}

		// Add to the comparison SELECT if it exists.
		if n.JoinComparisonSelect != nil {
			err = a.AddMeasureField(n.JoinComparisonSelect, m)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// buildWhereForUnderlyingTable constructs an expression for a WHERE clause for the underlying table.
// It combines the provided where expression with any security policy filters.
// It allows the input `where` to be nil, and returns nil if there are no conditions to apply.
func (a *AST) buildWhereForUnderlyingTable(where *Expression) (*ExprNode, error) {
	var res *ExprNode

	expr, args, err := a.SQLForExpression(where, nil, false, true)
	if err != nil {
		return nil, fmt.Errorf("failed to compile 'where': %w", err)
	}
	res = res.And(expr, args)

	if qf := a.Security.QueryFilter(); qf != nil {
		e := NewExpressionFromProto(qf)
		expr, args, err = a.SQLForExpression(e, nil, false, false)
		if err != nil {
			return nil, fmt.Errorf("failed to compile the security policy's query filter: %w", err)
		}
		res = res.And(expr, args)
	}

	if rf := a.Security.RowFilter(); rf != "" {
		res = res.And(rf, nil)
	}

	return res, nil
}

// buildBaseSelect constructs a base SELECT node against the underlying table.
func (a *AST) buildBaseSelect(alias string, comparison bool) (*SelectNode, error) {
	n := &SelectNode{
		Alias:     alias,
		DimFields: a.dimFields,
		Unnests:   a.unnests,
		Group:     true,
		FromTable: a.underlyingTable,
		Where:     a.underlyingWhere,
	}

	if a.Query.Rows {
		n.Group = false
	}

	tr := a.Query.TimeRange
	if comparison {
		n.DimFields = a.comparisonDimFields
		tr = a.Query.ComparisonTimeRange
	}

	err := a.AddTimeRange(n, tr)
	if err != nil {
		return nil, fmt.Errorf("failed to add time range: %w", err)
	}

	// If there is a spine, we wrap the base SELECT in a new SELECT that we add the spine to.
	// We do not join the spine directly to the FromTable because the join would be evaluated before the GROUP BY,
	// which would impact the measure aggregations (e.g. counts per group would be wrong).
	if a.Query.Spine != nil && !(a.Query.Spine.TimeRange != nil && comparison) { // Skip time range spines in the comparison select
		sn, err := a.buildSpineSelect(a.GenerateIdentifier(), a.Query.Spine, tr)
		if err != nil {
			return nil, err
		}

		a.WrapSelect(n, a.GenerateIdentifier())
		n.SpineSelect = sn

		// Update the dimension fields to derive from the SpineSelect instead of the FromSelect
		// (since by definition, some dimension values in the spine might not be present in FromSelect).
		for i, f := range n.DimFields {
			f.Expr = a.Dialect.EscapeMember(sn.Alias, f.Name)
			n.DimFields[i] = f
		}
	}

	return n, nil
}

// buildSpineSelect constructs a SELECT node for the given spine of dimension values.
func (a *AST) buildSpineSelect(alias string, spine *Spine, tr *TimeRange) (*SelectNode, error) {
	if spine == nil {
		return nil, nil
	}

	if spine.Where != nil {
		// Using buildWhereForUnderlyingTable to include security filters.
		// Note that buildWhereForUnderlyingTable handles nil expressions gracefully.
		where, err := a.buildWhereForUnderlyingTable(spine.Where.Expression)
		if err != nil {
			return nil, fmt.Errorf("failed to compile 'spine.where': %w", err)
		}

		n := &SelectNode{
			Alias:     alias,
			DimFields: a.dimFields,
			FromTable: a.underlyingTable,
			Unnests:   a.unnests,
			Group:     true,
			Where:     where,
		}
		err = a.AddTimeRange(n, tr)
		if err != nil {
			return nil, fmt.Errorf("failed to add time range: %w", err)
		}

		return n, nil
	}

	if spine.TimeRange != nil {
		// if spine generates more than 1000 values then return an error
		bins := timeutil.ApproximateBins(spine.TimeRange.Start, spine.TimeRange.End, spine.TimeRange.Grain.ToTimeutil())
		if bins > 1000 {
			return nil, errors.New("failed to apply time spine: time range has more than 1000 bins")
		}

		timeDim := a.MetricsView.TimeDimension
		if spine.TimeRange.TimeDimension != "" {
			timeDim = spine.TimeRange.TimeDimension
		}

		tf, ok := a.findFieldForComputedTimeDimension(a.dimFields, timeDim)
		if !ok {
			return nil, fmt.Errorf("failed to find computed time dimension %q", timeDim)
		}
		timeAlias := tf.Name

		var newDims []FieldNode
		for _, f := range a.dimFields {
			if f.Name == tf.Name {
				continue
			}
			newDims = append(newDims, f)
		}

		start := spine.TimeRange.Start
		end := spine.TimeRange.End
		grain := spine.TimeRange.Grain
		tz := time.UTC
		if a.Query.TimeZone != "" {
			var err error
			tz, err = time.LoadLocation(a.Query.TimeZone)
			if err != nil {
				return nil, fmt.Errorf("invalid time zone %q: %w", a.Query.TimeZone, err)
			}
		}
		sel, args, err := a.Dialect.SelectTimeRangeBins(start, end, grain.ToProto(), timeAlias, tz)
		if err != nil {
			return nil, fmt.Errorf("failed to generate time spine: %w", err)
		}

		rangeSelect := &SelectNode{
			Alias: alias,
			RawSelect: &ExprNode{
				Expr: sel,
				Args: args,
			},
		}

		// if there is only one dimension in the query, then we can directly join the spine time range with the dimension
		if len(a.dimFields) == 1 {
			return rangeSelect, nil
		}

		// give alias to the outer select as range select will be moved to cross join
		rangeSelect.Alias = a.GenerateIdentifier()

		dimSelect := &SelectNode{
			Alias:     alias,
			DimFields: newDims,
			Unnests:   a.unnests,
			FromTable: a.underlyingTable,
			Where:     a.underlyingWhere,
			Group:     true,
		}

		err = a.AddTimeRange(dimSelect, tr)
		if err != nil {
			return nil, fmt.Errorf("failed to add time range: %w", err)
		}

		a.WrapSelect(dimSelect, a.GenerateIdentifier())

		dimSelect.CrossJoinSelects = []*SelectNode{rangeSelect}

		// now add cross join field to the dimension list
		dimSelect.DimFields = append(dimSelect.DimFields, FieldNode{
			Name:        timeAlias,
			DisplayName: timeAlias,
			Expr:        a.Dialect.EscapeMember(rangeSelect.Alias, timeAlias),
		})

		return dimSelect, nil
	}

	return nil, errors.New("unhandled spine type")
}

// findFieldForDimension finds the field in the SelectNode that corresponds to the dimension selector.
// It takes computed dimensions into account, comparing against the underlying dimension name instead of the query alias.
func (a *AST) findFieldForDimension(n *SelectNode, dim *runtimev1.MetricsViewSpec_DimensionSelector) (FieldNode, bool) {
	for _, f := range n.DimFields {
		// If name matches, we're done
		if dim.Name == f.Name {
			return f, true
		}

		// Find original query dimension for the field
		var fqd Dimension
		for _, qd := range a.Query.Dimensions {
			if f.Name == qd.Name {
				fqd = qd
				break
			}
		}

		// If it's a computed dimension, check against the underlying dimension name (and time grain if specified)
		if fqd.Compute != nil && fqd.Compute.TimeFloor != nil {
			if dim.Name != fqd.Compute.TimeFloor.Dimension {
				continue
			}

			if dim.TimeGrain != runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED && dim.TimeGrain != fqd.Compute.TimeFloor.Grain.ToProto() {
				continue
			}

			return f, true
		}
	}

	return FieldNode{}, false
}

// findFieldForComputedTimeDimension finds the first field in the provided dims that represents a TimeFloor computation over baseTimeDim.
// If no field is found, it returns false.
func (a *AST) findFieldForComputedTimeDimension(dims []FieldNode, baseTimeDim string) (FieldNode, bool) {
	for _, f := range dims {
		// Find original query dimension for the field
		var fqd Dimension
		for _, qd := range a.Query.Dimensions {
			if f.Name == qd.Name {
				fqd = qd
				break
			}
		}

		// If it's a computed dimension, check against the underlying dimension name (and time grain if specified)
		if fqd.Compute != nil && fqd.Compute.TimeFloor != nil {
			if baseTimeDim != fqd.Compute.TimeFloor.Dimension {
				continue
			}

			return f, true
		}
	}

	return FieldNode{}, false
}

// getTimeDimensionExpression returns the SQL expression for the time dimension specified in the TimeRange or the metrics view's time dimension.
// It looks up the time dimension definition in the metrics view or returns the escaped column name from the model.
func (a *AST) getTimeDimensionExpression(tr *TimeRange) (string, error) {
	timeDim := a.MetricsView.TimeDimension
	if tr.TimeDimension != "" {
		timeDim = tr.TimeDimension
	}

	t, err := a.LookupDimension(timeDim, true)
	if err != nil {
		return "", fmt.Errorf("time dimension %q not found: %w", timeDim, err)
	}

	expr, err := a.Dialect.MetricsViewDimensionExpression(t)
	if err != nil {
		return "", fmt.Errorf("failed to compile time dimension %q expression: %w", t.Name, err)
	}

	return expr, nil
}

// checkNameForComputedField checks that the name for a computed field does not collide with an existing dimension or measure name.
// (This is necessary because even if the other name is not used in the query, it might be referenced by a derived measure.)
func (a *AST) checkNameForComputedField(name string) error {
	if name == "" {
		return errors.New("name for computed field is empty")
	}

	if name == a.MetricsView.TimeDimension {
		return errors.New("name for computed field collides with the time dimension name")
	}

	for _, d := range a.MetricsView.Dimensions {
		if d.Name == name {
			return errors.New("name for computed field collides with an existing dimension name")
		}
	}

	for _, m := range a.MetricsView.Measures {
		if m.Name == name {
			return errors.New("name for computed field collides with an existing measure name")
		}
	}

	return nil
}

// checkRequiredDimensionsPresentInQuery checks that all the dimensions required by the measure are present in the query.
// It checks against the query's dimensions because:
//  1. This enables correctly checking against the underlying time dimension name for dimensions with time floor applied.
//  2. The query's dimensions are projected into every sub-query, so it's not necessary to check against the current sub-query's Dimensions.
func (a *AST) checkRequiredDimensionsPresentInQuery(m *runtimev1.MetricsViewSpec_Measure) error {
	for _, rd := range m.RequiredDimensions {
		var found bool
		for _, qd := range a.Query.Dimensions {
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

// sqlForTimeRange builds a SQL expression and query args for filtering by a time range.
func (a *AST) sqlForTimeRange(timeDimExpr string, start, end time.Time) (string, []any) {
	var where string
	var args []any
	if !start.IsZero() && !end.IsZero() {
		where = fmt.Sprintf("%s >= %s AND %s < %s", timeDimExpr, a.Dialect.GetTimeDimensionParameter(), timeDimExpr, a.Dialect.GetTimeDimensionParameter())
		args = []any{start, end}
	} else if !start.IsZero() {
		where = fmt.Sprintf("%s >= %s", timeDimExpr, a.Dialect.GetTimeDimensionParameter())
		args = []any{start}
	} else if !end.IsZero() {
		where = fmt.Sprintf("%s < %s", timeDimExpr, a.Dialect.GetTimeDimensionParameter())
		args = []any{end}
	} else {
		return "", nil
	}
	return where, args
}

// sqlForMeasure builds a SQL expression for a measure, including its window if present.
// It uses the provided n to resolve dimensions expressions for window partitions.
func (a *AST) sqlForMeasure(m *runtimev1.MetricsViewSpec_Measure, n *SelectNode) (string, error) {
	// If not applying a window, just return the measure expression.
	if m.Window == nil {
		return m.Expression, nil
	}

	// If partitioning is not enabled, ordering and framing doesn't matter
	if !m.Window.Partition {
		return fmt.Sprintf("%s OVER ()", m.Expression), nil
	}

	// If partitioning is enabled, we partition by all dimensions that we don't order by.

	// Gather order by fields
	orderFields := make([]FieldNode, 0, len(m.Window.OrderBy))
	orderDesc := make([]bool, 0, len(m.Window.OrderBy))
	for _, d := range m.Window.OrderBy {
		f, ok := a.findFieldForDimension(n, d)
		if !ok {
			// In practice, this should never happen because the OrderBy dimensions are in RequiredDimensions and have been checked before this point.
			return "", fmt.Errorf("dimension %q required by window measure %q not found in query", d.Name, m.Name)
		}
		orderFields = append(orderFields, f)
		orderDesc = append(orderDesc, d.Desc)
	}

	// Gather partition fields
	var partitionFields []FieldNode
	for _, f := range n.DimFields {
		found := false
		for _, of := range orderFields {
			if f.Name == of.Name {
				found = true
				break
			}
		}
		if !found {
			partitionFields = append(partitionFields, f)
		}
	}

	// Build the window expression
	b := &strings.Builder{}
	b.WriteString(m.Expression)
	b.WriteString(" OVER (")
	if len(partitionFields) > 0 {
		b.WriteString("PARTITION BY ")
		for i, f := range partitionFields {
			if i > 0 {
				b.WriteString(", ")
			}

			b.WriteString(f.Expr)
		}
	}
	if len(orderFields) > 0 {
		if len(partitionFields) > 0 {
			b.WriteString(" ")
		}
		b.WriteString("ORDER BY ")
		for i, f := range orderFields {
			if i > 0 {
				b.WriteString(", ")
			}

			b.WriteString(f.Expr)
			if orderDesc[i] {
				b.WriteString(" DESC")
			}
		}
	}
	if m.Window.FrameExpression != "" {
		b.WriteString(" ")
		b.WriteString(m.Window.FrameExpression)
	}
	b.WriteString(")")

	return b.String(), nil
}

// sqlForExpression returns the provided time expression adjusted by the fixed time offset between the current query's base and comparison time ranges.
// The timestamp column (ie a.metricsView.TimeDimension) is expected to be the base timestamp for `expr` (in case of multiple metrics view time dimensions defined).
func (a *AST) sqlForExpressionAdjustedByComparisonTimeRangeOffset(expr, timeDim string, g, mg TimeGrain) (string, error) {
	if a.Query.TimeRange == nil || a.Query.TimeRange.Start.IsZero() || a.Query.ComparisonTimeRange == nil || a.Query.ComparisonTimeRange.Start.IsZero() {
		return "", errors.New("must specify an explicit start time for both the base and comparison time range when comparing by a time dimension")
	}

	start1 := a.Query.TimeRange.Start
	start2 := a.Query.ComparisonTimeRange.Start

	var dateDiff string
	if g == TimeGrainUnspecified {
		g = TimeGrainMillisecond // todo millis won't work for druid
		res, err := a.Dialect.DateDiff(g.ToProto(), start1, start2)
		if err != nil {
			return "", err
		}
		dateDiff = res
	} else if g == mg {
		res, err := a.Dialect.DateDiff(g.ToProto(), start1, start2)
		if err != nil {
			return "", err
		}
		dateDiff = res
	} else {
		// larger time grain values can change as well
		res, err := a.Dialect.DateDiff(mg.ToProto(), start1, start2)
		if err != nil {
			return "", err
		}
		dateDiff = res

		// DATE_TRUNC('year', t - INTERVAL (DATE_DIFF(start, end)) day)
		tc := a.Dialect.EscapeIdentifier(timeDim)
		expr, err := a.Dialect.IntervalSubtract(tc, dateDiff, mg.ToProto())
		if err != nil {
			return "", err
		}
		dim := &runtimev1.MetricsViewSpec_Dimension{
			Expression: expr,
		}
		expr, err = a.Dialect.DateTruncExpr(dim, g.ToProto(), a.Query.TimeZone, int(a.MetricsView.FirstDayOfWeek), int(a.MetricsView.FirstMonthOfYear))
		if err != nil {
			return "", fmt.Errorf(`failed to compute time floor: %w`, err)
		}
		return expr, nil
	}

	return a.Dialect.IntervalSubtract(expr, dateDiff, g.ToProto())
}
