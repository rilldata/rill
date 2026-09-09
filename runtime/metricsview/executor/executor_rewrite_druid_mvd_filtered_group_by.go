package executor

import (
	"fmt"
	"slices"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/metricsview"
)

// rewriteDruidMVDFilteredGroupBy narrows unnested (multi-value) dimensions in the query's GROUP BY to the values allowed by the query's filter.
//
// Druid unnests multi-value dimensions implicitly.
// A filter such as `tags IN ('a')` keeps every row whose array contains 'a', and a GROUP BY on the dimension then emits every value in those rows' arrays, not just 'a'.
// So values that merely co-occur with the filtered ones leak into the result.
//
// Following Druid's own guidance, the grouped expression is wrapped in MV_FILTER_ONLY with the filter's values, e.g. `MV_FILTER_ONLY("tags", ARRAY['a'])`,
// or in MV_FILTER_REGEX with the filter's regex for an ILIKE filter, e.g. `MV_FILTER_REGEX("tags", '^(?i).*foo.*$')`.
// Druid plans these as a filtered virtual column on the dimension, so the group-by emits only the allowed values while the row filter on the raw column keeps using its index.
// Re-applying the filter to the grouped output instead does not work: Druid's planner pushes a filter on a group key back below the aggregation, where it regains the "array contains" semantics.
//
// The narrowing is sound only if every row that passes the WHERE clause is guaranteed to keep at least one value; otherwise the row's measures would be dropped or attributed to NULL.
// A single top-level conjunct of the WHERE clause gives that guarantee: every matching row contains a value satisfying it.
// Combining conjuncts does not, since `tags IN ('a','b') AND tags IN ('b','c')` is satisfied by a row containing ['a','c'] through different elements.
// So conjuncts are never intersected: a regex stands alone, allow lists are merged by subset or union (see druidMVDRestrictions), and a filter under an OR is never used.
// Only `IN`, `=` and `ILIKE` filters with string values are narrowed.
// For a dimension backed directly by a column, Druid plans the MV_FILTER call as its optimized filtered virtual column; for an expression-backed dimension it falls back to a generic expression virtual column, which is correct but evaluated per row.
// Exclusion filters need no counterpart: a row containing an excluded value is excluded as a whole, so nothing leaks from them.
func (e *Executor) rewriteDruidMVDFilteredGroupBy(ast *metricsview.AST) {
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
	// The spine select has its own filter (Query.Spine.Where), independent of the query's WHERE clause, and exists to keep values the WHERE clause may exclude.
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

// druidMVDRestrictions returns, for each unnested dimension in dims, the restriction implied by the given filter, if any.
// The restriction comes from the top-level conjuncts of the filter of the form `dim IN (...)`, `dim = ...` or, if includeRegex is set, `dim ILIKE ...`.
// If several conjuncts qualify, the first regex is used if there is one (a search must return values matching the search text).
// Otherwise the allow lists are merged: if one list is a subset of the other it refines it and the subset is used (e.g. the values added by rewriteQueryDruidExactify), else the union is used so that every value the filter names is kept.
// See rewriteDruidMVDFilteredGroupBy for why the lists are never intersected.
// Dimensions without such a filter are omitted.
func druidMVDRestrictions(mv *runtimev1.MetricsViewSpec, dims []metricsview.Dimension, where *metricsview.Expression, includeRegex bool) map[string]druidMVDRestriction {
	if where == nil {
		return nil
	}

	// Find the unnested dimensions that the query groups by.
	// Computed dimensions (e.g. time floors) are skipped since their name does not refer to a metrics view dimension.
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
	if len(eligible) == 0 {
		return nil
	}

	res := make(map[string]druidMVDRestriction)
	for _, conj := range topLevelConjuncts(where) {
		dim, vals, regex, ok := mvdRestrictionFromConjunct(conj)
		if !ok || !eligible[dim] {
			continue
		}
		prev, hasPrev := res[dim]
		switch {
		case regex != "":
			// A regex comes from a dimension search, whose results must match the search text, so it takes precedence over an allow list.
			// Only the first regex is used.
			if includeRegex && prev.regex == "" {
				res[dim] = druidMVDRestriction{regex: regex}
			}
		case prev.regex != "":
			// Keep the regex.
		case !hasPrev:
			res[dim] = druidMVDRestriction{values: vals}
		default:
			res[dim] = druidMVDRestriction{values: mergeAllowLists(prev.values, vals)}
		}
	}
	return res
}

// applyDruidMVDRestrictions narrows the dimensions in every select node that reads directly from the underlying table.
// Spine selects are filtered by the spine's own filter rather than the query's, so they get spineRestrictions instead.
func applyDruidMVDRestrictions(n *metricsview.SelectNode, restrictions, spineRestrictions map[string]druidMVDRestriction) {
	if n == nil {
		return
	}

	wrapDimFieldsInMVDFilter(n, restrictions)
	// A spine select reads directly from the table and has no sub-selects of its own.
	wrapDimFieldsInMVDFilter(n.SpineSelect, spineRestrictions)

	applyDruidMVDRestrictions(n.FromSelect, restrictions, spineRestrictions)
	applyDruidMVDRestrictions(n.JoinComparisonSelect, restrictions, spineRestrictions)
	for _, s := range n.LeftJoinSelects {
		applyDruidMVDRestrictions(s, restrictions, spineRestrictions)
	}
	for _, s := range n.CrossJoinSelects {
		applyDruidMVDRestrictions(s, restrictions, spineRestrictions)
	}
}

// wrapDimFieldsInMVDFilter rewrites the restricted dimensions of a select node that reads directly from the underlying table to MV_FILTER_ONLY or MV_FILTER_REGEX calls.
// Only the projected expression changes; the GROUP BY refers to it by ordinal and the WHERE clause has already been compiled against the raw column.
func wrapDimFieldsInMVDFilter(n *metricsview.SelectNode, restrictions map[string]druidMVDRestriction) {
	if n == nil || n.FromTable == nil || len(restrictions) == 0 {
		return
	}

	// The AST shares one DimFields slice between the base select and the spine select, so clone it before mutating.
	// Otherwise the second node to be visited would wrap the expression a second time, which is redundant and makes Druid fall back to a slower expression virtual column instead of its optimized filtered one.
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

// mergeAllowLists combines the allow lists of two conjuncts on the same dimension.
// If one list is a subset of the other, it refines it and is returned as is.
// Otherwise the union is returned, so that every value named by either conjunct is kept.
// Either result is sound: every matching row contains a value from each list, so it keeps at least one value.
// The intersection would not be, since a row may satisfy the two conjuncts through different values.
func mergeAllowLists(a, b []string) []string {
	inA := make(map[string]bool, len(a))
	for _, v := range a {
		inA[v] = true
	}
	inB := make(map[string]bool, len(b))
	for _, v := range b {
		inB[v] = true
	}

	if isSubset(b, inA) {
		return b
	}
	if isSubset(a, inB) {
		return a
	}
	res := slices.Clone(a)
	for _, v := range b {
		if !inA[v] {
			res = append(res, v)
		}
	}
	return res
}

// isSubset reports whether every value in vals is in set.
func isSubset(vals []string, set map[string]bool) bool {
	for _, v := range vals {
		if !set[v] {
			return false
		}
	}
	return true
}

// topLevelConjuncts flattens the top-level AND conditions of expr into the individual conjuncts.
// Any other expression is returned as the only conjunct.
func topLevelConjuncts(expr *metricsview.Expression) []*metricsview.Expression {
	if expr == nil {
		return nil
	}
	if expr.Condition != nil && expr.Condition.Operator == metricsview.OperatorAnd {
		var res []*metricsview.Expression
		for _, sub := range expr.Condition.Expressions {
			res = append(res, topLevelConjuncts(sub)...)
		}
		return res
	}
	return []*metricsview.Expression{expr}
}

// mvdRestrictionFromConjunct returns the dimension and either the values of a `dim IN (...)` or `dim = ...` condition, or the regex of a `dim ILIKE ...` condition.
// Like the AST, it accepts the values of an IN condition either as a single list or as one expression per value (the form the frontend sends).
// It returns false for any other shape, or if any value is not a string, since multi-value dimensions hold strings only and NULL is not an array value.
func mvdRestrictionFromConjunct(expr *metricsview.Expression) (dim string, vals []string, regex string, ok bool) {
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

	vals = make([]string, len(raw))
	for i, v := range raw {
		s, isString := v.(string)
		if !isString {
			return "", nil, "", false
		}
		vals[i] = s
	}
	return left.Name, vals, "", true
}
