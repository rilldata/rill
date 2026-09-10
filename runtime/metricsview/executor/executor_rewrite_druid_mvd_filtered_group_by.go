package executor

import (
	"fmt"
	"slices"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/metricsview"
)

// rewriteDruidMVDFilteredGroupBy narrows unnested (multi-value) dimensions in the GROUP BY to the values the query's filter allows.
//
// Druid unnests multi-value dimensions implicitly: `tags IN ('a')` keeps every row whose array contains 'a', and a GROUP BY on tags then emits every value in those arrays, not just 'a'.
// Wrapping the grouped expression in MV_FILTER_ONLY (or MV_FILTER_REGEX for an ILIKE filter) makes the group-by emit only the allowed values, while the WHERE clause keeps filtering the raw column.
// Re-filtering the grouped output instead does not work: Druid's planner pushes a filter on a group key back below the aggregation.
//
// Not handled; the group-by then still emits every value co-occurring in the matching rows, as it does without this rewrite:
//   - The dimension is filtered inside an OR, e.g. `tags IN ('a') OR city = 'NYC'`: rows admitted through the other branch need not contain an allowed value, so narrowing would drop their values.
//     Does not occur from the UI, which ANDs its per-dimension filters.
//   - The dimension has two IN/= conditions: applying one is arbitrary and combining them is unsound, since a row may satisfy each through a different value.
//     Does not occur from the UI, which keeps one filter per dimension; the one known producer, rewriteQueryDruidExactify, is handled through the exactified parameter.
//   - The dimension is filtered by ILIKE while MetricsDruidMVDFilteredSearch is off (MV_FILTER_REGEX needs Druid 35), or by a second ILIKE (only the first narrows).
//     The former is every dimension search on an older Druid; dimension search issues a single ILIKE.
//   - The dimension is filtered by a comparison, a subquery (a filter on a measure), or an IN containing a non-string value or only NULL: MV_FILTER has no counterpart.
//     NULL entries alongside strings are simply ignored: a NULL row stays NULL under MV_FILTER, so narrowing does not affect it.
//   - The dimension is filtered only by a security policy or another filter outside the query's WHERE clause: those are not inspected. Rare for a multi-value dimension.
//   - The dimension is aggregated with count_distinct without being grouped: the leak is inside the measure and would need MV_FILTER in the measure expression.
//
// exactified is the `dim IN (...)` expression appended by rewriteQueryDruidExactify, if any.
// Its values are exactly the groups to return, so they replace the restriction derived from the user's filter for that dimension.
func (e *Executor) rewriteDruidMVDFilteredGroupBy(ast *metricsview.AST, exactified *metricsview.Expression) {
	if !e.instanceCfg.MetricsDruidMVDFilteredGroupBy {
		return
	}
	if ast.Dialect.String() != drivers.DialectNameDruid {
		return
	}

	// A raw rows query has no GROUP BY to narrow.
	if ast.Query.Rows {
		return
	}

	restrictions := druidMVDRestrictions(ast.MetricsView, ast.Query.Dimensions, ast.Query.Where, e.instanceCfg.MetricsDruidMVDFilteredSearch)
	if exactified != nil {
		if dim, vals, _, ok := mvdRestrictionFromCondition(exactified); ok && druidMVDEligibleDims(ast.MetricsView, ast.Query.Dimensions)[dim] {
			if restrictions == nil {
				restrictions = make(map[string]druidMVDRestriction)
			}
			restrictions[dim] = druidMVDRestriction{values: vals}
		}
	}
	// A where-spine has its own filter, so its restriction is derived from that instead.
	var spineRestrictions map[string]druidMVDRestriction
	if ast.Query.Spine != nil && ast.Query.Spine.Where != nil {
		spineRestrictions = druidMVDRestrictions(ast.MetricsView, ast.Query.Dimensions, ast.Query.Spine.Where.Expression, e.instanceCfg.MetricsDruidMVDFilteredSearch)
	}
	if len(restrictions) == 0 && len(spineRestrictions) == 0 {
		return
	}
	applyDruidMVDRestrictions(ast.Root, restrictions, spineRestrictions)
}

// druidMVDRestriction describes the values a group-by on an unnested dimension may emit: either an explicit allow list or a regex the values must match.
// Exactly one of the two is set.
type druidMVDRestriction struct {
	values []string
	regex  string
}

// druidMVDRestrictions returns, per unnested dimension, the restriction implied by the top-level filter of the form `dim IN (...)`, `dim = ...` or, if includeRegex is set, `dim ILIKE ...`.
// A regex takes precedence over an allow list; a dimension with two allow lists is left out (see the assumptions on rewriteDruidMVDFilteredGroupBy).
func druidMVDRestrictions(mv *runtimev1.MetricsViewSpec, dims []metricsview.Dimension, where *metricsview.Expression, includeRegex bool) map[string]druidMVDRestriction {
	if where == nil {
		return nil
	}

	eligible := druidMVDEligibleDims(mv, dims)
	if len(eligible) == 0 {
		return nil
	}

	res := make(map[string]druidMVDRestriction)
	ambiguous := make(map[string]bool) // dimensions with more than one allow list
	for _, cond := range topLevelConditions(where) {
		dim, vals, regex, ok := mvdRestrictionFromCondition(cond)
		if !ok || !eligible[dim] {
			continue
		}
		prev, hasPrev := res[dim]
		switch {
		case regex != "":
			if includeRegex && prev.regex == "" {
				res[dim] = druidMVDRestriction{regex: regex}
			}
		case prev.regex != "":
			// The regex wins.
		case hasPrev:
			ambiguous[dim] = true
		default:
			res[dim] = druidMVDRestriction{values: vals}
		}
	}
	for dim := range ambiguous {
		if res[dim].regex == "" {
			delete(res, dim)
		}
	}
	return res
}

// druidMVDEligibleDims returns the names of the unnested dimensions in dims.
// Computed dimensions are skipped since their name is not a metrics view dimension.
func druidMVDEligibleDims(mv *runtimev1.MetricsViewSpec, dims []metricsview.Dimension) map[string]bool {
	eligible := make(map[string]bool)
	for _, qd := range dims {
		if qd.Compute != nil {
			continue
		}
		for _, dim := range mv.Dimensions {
			if dim.Name == qd.Name && dim.Unnest {
				eligible[qd.Name] = true
				break
			}
		}
	}
	return eligible
}

// applyDruidMVDRestrictions narrows the dimensions in every select node that reads from the underlying table.
// A where-spine select is filtered by the spine's own filter, so it gets spineRestrictions instead.
func applyDruidMVDRestrictions(n *metricsview.SelectNode, restrictions, spineRestrictions map[string]druidMVDRestriction) {
	if n == nil {
		return
	}

	wrapDimFieldsInMVDFilter(n, restrictions)

	if sp := n.SpineSelect; sp != nil {
		if sp.FromTable != nil {
			// A where-spine reads from the table with the spine's own filter.
			wrapDimFieldsInMVDFilter(sp, spineRestrictions)
		} else {
			// A time spine wraps a select that reads from the table with the query's filter (see AST.buildSpineSelect).
			applyDruidMVDRestrictions(sp, restrictions, spineRestrictions)
		}
	}

	applyDruidMVDRestrictions(n.FromSelect, restrictions, spineRestrictions)
	applyDruidMVDRestrictions(n.JoinComparisonSelect, restrictions, spineRestrictions)
	for _, s := range n.LeftJoinSelects {
		applyDruidMVDRestrictions(s, restrictions, spineRestrictions)
	}
	for _, s := range n.CrossJoinSelects {
		applyDruidMVDRestrictions(s, restrictions, spineRestrictions)
	}
}

// wrapDimFieldsInMVDFilter rewrites the restricted dimensions of a select node that reads from the underlying table to MV_FILTER_ONLY or MV_FILTER_REGEX calls.
// Only the projection changes; the GROUP BY refers to it by ordinal.
func wrapDimFieldsInMVDFilter(n *metricsview.SelectNode, restrictions map[string]druidMVDRestriction) {
	if n == nil || n.FromTable == nil || len(restrictions) == 0 {
		return
	}

	// The base and spine selects share one DimFields slice; clone it so each node wraps the expression once.
	n.DimFields = slices.Clone(n.DimFields)
	for i := range n.DimFields {
		f := &n.DimFields[i]
		r, ok := restrictions[f.Name]
		if !ok || !f.Unnest {
			continue
		}
		if r.regex != "" {
			f.Expr = fmt.Sprintf("MV_FILTER_REGEX(%s, %s)", f.Expr, drivers.EscapeStringValue(r.regex))
			continue
		}
		quoted := make([]string, len(r.values))
		for j, v := range r.values {
			quoted[j] = drivers.EscapeStringValue(v)
		}
		f.Expr = fmt.Sprintf("MV_FILTER_ONLY(%s, ARRAY[%s])", f.Expr, strings.Join(quoted, ", "))
	}
}

// topLevelConditions flattens the top-level ANDs of expr into the individual conditions.
// Any other expression is returned as the only condition.
func topLevelConditions(expr *metricsview.Expression) []*metricsview.Expression {
	if expr == nil {
		return nil
	}
	if expr.Condition != nil && expr.Condition.Operator == metricsview.OperatorAnd {
		var res []*metricsview.Expression
		for _, sub := range expr.Condition.Expressions {
			res = append(res, topLevelConditions(sub)...)
		}
		return res
	}
	return []*metricsview.Expression{expr}
}

// mvdRestrictionFromCondition returns the dimension and either the values of a `dim IN (...)` or `dim = ...` condition, or the regex of a `dim ILIKE ...` condition.
// Like the AST, it accepts the values of an IN condition either as a single list or as one expression per value (the form the frontend sends).
func mvdRestrictionFromCondition(expr *metricsview.Expression) (dim string, vals []string, regex string, ok bool) {
	if expr.Condition == nil || len(expr.Condition.Expressions) < 2 {
		return "", nil, "", false
	}
	exprs := expr.Condition.Expressions
	left := exprs[0]
	if left == nil || left.Name == "" || left.Condition != nil || left.Subquery != nil {
		return "", nil, "", false
	}
	isScalar := func(e *metricsview.Expression) bool {
		return e != nil && e.Name == "" && e.Condition == nil && e.Subquery == nil
	}

	var raw []any
	switch expr.Condition.Operator {
	case metricsview.OperatorIn:
		if len(exprs) == 2 {
			if list, isList := exprs[1].Value.([]any); isList {
				raw = list
			} else if isScalar(exprs[1]) {
				raw = []any{exprs[1].Value}
			}
		} else {
			for _, e := range exprs[1:] {
				if !isScalar(e) {
					return "", nil, "", false
				}
				raw = append(raw, e.Value)
			}
		}
		if len(raw) == 0 {
			return "", nil, "", false
		}
	case metricsview.OperatorEq:
		if len(exprs) != 2 || !isScalar(exprs[1]) {
			return "", nil, "", false
		}
		raw = []any{exprs[1].Value}
	case metricsview.OperatorIlike:
		if len(exprs) != 2 {
			return "", nil, "", false
		}
		pattern, isString := exprs[1].Value.(string)
		if !isString {
			return "", nil, "", false
		}
		return left.Name, nil, metricsview.LikePatternToRegex(pattern), true
	default:
		return "", nil, "", false
	}

	for _, v := range raw {
		if v == nil {
			// skip null values as they do not affect results when added to MV_FILTER_ONLY
			// it will exist in the WHERE clause as dim IS NULL condition
			continue
		}
		s, isString := v.(string)
		if !isString {
			return "", nil, "", false
		}
		vals = append(vals, s)
	}
	if len(vals) == 0 {
		return "", nil, "", false
	}
	return left.Name, vals, "", true
}
