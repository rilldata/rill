package executor

import (
	"fmt"
	"regexp"
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
// Only top-level conjuncts of the WHERE clause are used, since every matching row is guaranteed to satisfy them; a filter under an OR could leave a row with no allowed value at all.
// Only `IN`, `=` and `ILIKE` filters with string values are narrowed, and only for dimensions backed by a plain column, since the MV_FILTER functions require a direct column reference.
func (e *Executor) rewriteDruidMVDFilteredGroupBy(ast *metricsview.AST) {
	if !e.instanceCfg.MetricsDruidMVDFilteredGroupBy {
		return
	}
	if ast.Dialect.String() != drivers.DialectNameDruid {
		return
	}

	restrictions := druidMVDRestrictions(ast.MetricsView, ast.Query, e.instanceCfg.MetricsDruidMVDFilteredSearch)
	if len(restrictions) == 0 {
		return
	}
	applyDruidMVDRestrictions(ast.Root, restrictions)
}

// druidMVDRestriction describes the values a group-by on an unnested dimension may emit: either an explicit allow list or a regex the values must match.
// Exactly one of the two is set.
type druidMVDRestriction struct {
	values []string
	regex  string
}

// druidMVDRestrictions returns, for each unnested dimension in the query's GROUP BY, the restriction implied by the query's filter, if any.
// The restrictions are taken from the top-level conjuncts of the WHERE clause of the form `dim IN (...)`, `dim = ...` or, if includeRegex is set, `dim ILIKE ...`.
// Several conjuncts on one dimension are combined as far as a single MV_FILTER call can express: allow lists are intersected, and an allow list is filtered by any regexes.
// Dimensions without such a filter, or not backed by a plain column, are omitted.
func druidMVDRestrictions(mv *runtimev1.MetricsViewSpec, qry *metricsview.Query, includeRegex bool) map[string]druidMVDRestriction {
	if qry.Rows || qry.Where == nil {
		return nil
	}

	// Find the unnested, plain-column dimensions that the query groups by.
	// Computed dimensions (e.g. time floors) are skipped since their name does not refer to a metrics view dimension.
	eligible := make(map[string]bool)
	for _, qd := range qry.Dimensions {
		if qd.Compute != nil {
			continue
		}
		for _, dim := range mv.Dimensions {
			if dim.Name == qd.Name && dim.Unnest && dim.Column != "" && dim.Expression == "" && dim.LookupTable == "" {
				eligible[qd.Name] = true
				break
			}
		}
	}
	if len(eligible) == 0 {
		return nil
	}

	// Collect the allow lists and regexes per dimension.
	allowLists := make(map[string][]string)
	regexes := make(map[string][]string)
	for _, conj := range topLevelConjuncts(qry.Where) {
		dim, vals, regex, ok := mvdRestrictionFromConjunct(conj)
		if !ok || !eligible[dim] {
			continue
		}
		if regex != "" {
			if includeRegex {
				regexes[dim] = append(regexes[dim], regex)
			}
			continue
		}
		if prev, ok := allowLists[dim]; ok {
			vals = intersectStrings(prev, vals)
		}
		allowLists[dim] = vals
	}

	res := make(map[string]druidMVDRestriction)
	for dim := range eligible {
		vals, hasVals := allowLists[dim]
		rxs := regexes[dim]
		switch {
		case hasVals:
			// Narrow the allow list by the regexes here rather than in Druid, since MV_FILTER calls cannot be nested (they require a direct column reference).
			// The regexes are of the simple form produced by LikePatternToRegex, which Go and Druid (Java) interpret alike.
			// A regex Go cannot compile is skipped; the allow list alone is still a safe superset.
			for _, rx := range rxs {
				re, err := regexp.Compile(rx)
				if err != nil {
					continue
				}
				var kept []string
				for _, v := range vals {
					if re.MatchString(v) {
						kept = append(kept, v)
					}
				}
				vals = kept
			}
			// An empty allow list means the filter cannot match any row; leave the dimension alone rather than emit an empty ARRAY.
			if len(vals) > 0 {
				res[dim] = druidMVDRestriction{values: vals}
			}
		case len(rxs) > 0:
			// Only one regex can be applied. Any single top-level conjunct is a safe restriction on its own, so the first one is used and the others are left to leak.
			res[dim] = druidMVDRestriction{regex: rxs[0]}
		}
	}
	return res
}

// applyDruidMVDRestrictions wraps the given dimensions in MV_FILTER_ONLY or MV_FILTER_REGEX in every select node that reads directly from the underlying table.
// Only the projected expression changes; the GROUP BY refers to it by ordinal and the WHERE clause has already been compiled against the raw column.
func applyDruidMVDRestrictions(n *metricsview.SelectNode, restrictions map[string]druidMVDRestriction) {
	if n == nil {
		return
	}

	if n.FromTable != nil {
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

	applyDruidMVDRestrictions(n.FromSelect, restrictions)
	applyDruidMVDRestrictions(n.SpineSelect, restrictions)
	applyDruidMVDRestrictions(n.JoinComparisonSelect, restrictions)
	for _, s := range n.LeftJoinSelects {
		applyDruidMVDRestrictions(s, restrictions)
	}
	for _, s := range n.CrossJoinSelects {
		applyDruidMVDRestrictions(s, restrictions)
	}
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

// intersectStrings returns the values of a that are also in b, in the order of a.
func intersectStrings(a, b []string) []string {
	inB := make(map[string]bool, len(b))
	for _, v := range b {
		inB[v] = true
	}
	var res []string
	for _, v := range a {
		if inB[v] {
			res = append(res, v)
		}
	}
	return res
}
