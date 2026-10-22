package metricsview

import (
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

// ExactMeasureSet returns the names of the metrics view measures that the query must compute with their exact_expression.
//
// It starts from every measure flagged Exact anywhere in the query, including the measures of subqueries in its where and having expressions,
// follows the measures they are calculated from (a derived measure's requires, the base measure of comparison and percent-of-total computes, and the refs of ephemeral expression measures),
// and keeps the measures that define an exact_expression.
// It then closes the set under the rule that a measure calculated from an exact measure is exact as well:
// every measure of the query that is calculated from a measure of the set adds the measures it is calculated from,
// so a ratio of two unique counts never divides an exact count by an approximate one.
// The frontend resolves the precision chips of a dashboard with the same rule.
//
// A measure name has exactly one value per query: if a measure is requested both exactly and approximately (directly, through a derived measure, or in a filter subquery), it is computed exactly.
// FlagExactMeasures writes the set back into the query's flags, so that a query built from a subset of its measures resolves the same set.
//
// It returns nil when no measure is flagged, which is the common case.
func ExactMeasureSet(mv *runtimev1.MetricsViewSpec, qry *Query) map[string]bool {
	if !hasExactMeasure(qry.Measures, qry.Where, qry.Having) {
		return nil
	}
	measures := appendQueryMeasures(nil, qry.Measures, qry.Where, qry.Having)
	return newMeasureRefs(mv, measures).exactSet(measures)
}

// FlagExactMeasures sets the Exact flag on every measure of the query that is calculated from a measure of its exact set (see ExactMeasureSet), including the measures of its filter subqueries,
// and reports whether the set is non-empty.
// After it, the flags carry the resolved set: a query built from a subset of the measures (the totals query of a percent-of-total measure, the first phase of a two-phase comparison, a filter subquery)
// resolves the same precision for every measure it shares with this query, so the set never has to be passed along with it.
// Flags on measures without an exact variant are left in place; they have no effect.
func FlagExactMeasures(mv *runtimev1.MetricsViewSpec, qry *Query) bool {
	if !hasExactMeasure(qry.Measures, qry.Where, qry.Having) {
		return false
	}
	measures := appendQueryMeasures(nil, qry.Measures, qry.Where, qry.Having)
	refs := newMeasureRefs(mv, measures)
	set := refs.exactSet(measures)
	if len(set) == 0 {
		return false
	}
	refs.flag(qry.Measures, set)
	refs.flagSubqueries(set, qry.Where, qry.Having)
	return true
}

// measureRefs resolves which metrics view measures the measures of a query are calculated from.
type measureRefs struct {
	byName map[string]*runtimev1.MetricsViewSpec_Measure
	// The refs of the expression measures defined in the query itself (ephemeral measures), which can reference metrics view measures.
	exprRefs map[string][]string
}

func newMeasureRefs(mv *runtimev1.MetricsViewSpec, measures []Measure) measureRefs {
	r := measureRefs{
		byName:   make(map[string]*runtimev1.MetricsViewSpec_Measure, len(mv.Measures)),
		exprRefs: make(map[string][]string),
	}
	for _, m := range mv.Measures {
		r.byName[m.Name] = m
	}
	for _, qm := range measures {
		if qm.Compute == nil || qm.Compute.Expression == nil {
			continue
		}
		parsed, err := ParseMeasureExpression(qm.Compute.Expression.Expression)
		if err != nil {
			// Invalid expressions are reported when the AST resolves the measure.
			continue
		}
		r.exprRefs[qm.Name] = parsed.Refs()
	}
	return r
}

// exactSet implements ExactMeasureSet for the measures of a query and of its filter subqueries (see appendQueryMeasures).
func (r measureRefs) exactSet(measures []Measure) map[string]bool {
	set := make(map[string]bool)
	for _, qm := range measures {
		if qm.Exact {
			r.addExact(qm, set)
		}
	}
	if len(set) == 0 {
		return nil
	}

	// A measure calculated from an exact measure is exact as well, so the measures it is calculated from join the set.
	// They can make further measures exact in turn, so repeat until the set stops growing.
	for grown := true; grown; {
		grown = false
		for _, qm := range measures {
			if qm.Exact || !r.meets(qm, set) {
				continue
			}
			n := len(set)
			r.addExact(qm, set)
			grown = grown || len(set) > n
		}
	}
	return set
}

// flag sets Exact on the measures that are calculated from a measure of the set.
func (r measureRefs) flag(measures []Measure, set map[string]bool) {
	for i := range measures {
		if !measures[i].Exact && r.meets(measures[i], set) {
			measures[i].Exact = true
		}
	}
}

// flagSubqueries calls flag for the measures of every subquery in the expressions, recursively.
func (r measureRefs) flagSubqueries(set map[string]bool, exprs ...*Expression) {
	for _, e := range exprs {
		if e == nil {
			continue
		}
		if e.Subquery != nil {
			r.flag(e.Subquery.Measures, set)
			r.flagSubqueries(set, e.Subquery.Where, e.Subquery.Having)
		}
		if e.Condition != nil {
			r.flagSubqueries(set, e.Condition.Expressions...)
		}
	}
}

// addExact adds the measures with an exact_expression that qm is calculated from, including qm itself, to the set.
func (r measureRefs) addExact(qm Measure, set map[string]bool) {
	r.walk(qm, func(m *runtimev1.MetricsViewSpec_Measure) bool {
		if m.ExactExpression != "" {
			set[m.Name] = true
		}
		return true
	})
}

// meets reports whether qm is calculated from a measure of the set, or is one itself.
func (r measureRefs) meets(qm Measure, set map[string]bool) bool {
	found := false
	r.walk(qm, func(m *runtimev1.MetricsViewSpec_Measure) bool {
		found = set[m.Name]
		return !found
	})
	return found
}

// walk calls fn with every metrics view measure that qm is calculated from, transitively, starting with qm itself if it is a metrics view measure.
// It follows a derived measure's requires and an expression measure's refs, and stops when fn returns false.
func (r measureRefs) walk(qm Measure, fn func(m *runtimev1.MetricsViewSpec_Measure) bool) {
	root, ok := measureRoot(qm)
	if !ok {
		return
	}
	visited := make(map[string]bool)
	var visit func(name string) bool
	visit = func(name string) bool {
		if visited[name] {
			return true
		}
		visited[name] = true
		if m, ok := r.byName[name]; ok {
			if !fn(m) {
				return false
			}
			for _, ref := range m.ReferencedMeasures {
				if !visit(ref) {
					return false
				}
			}
			return true
		}
		for _, ref := range r.exprRefs[name] {
			if !visit(ref) {
				return false
			}
		}
		return true
	}
	visit(root)
}

// measureRoot returns the name a query measure is calculated from directly: the measure itself for a plain or expression measure, and the base measure of a comparison or percent-of-total compute.
// Count, count distinct, URI and comparison time computes do not reference metrics view measures.
func measureRoot(qm Measure) (string, bool) {
	if qm.Compute == nil {
		return qm.Name, true
	}
	c := qm.Compute
	switch {
	case c.ComparisonValue != nil:
		return c.ComparisonValue.Measure, true
	case c.ComparisonDelta != nil:
		return c.ComparisonDelta.Measure, true
	case c.ComparisonRatio != nil:
		return c.ComparisonRatio.Measure, true
	case c.PercentOfTotal != nil:
		return c.PercentOfTotal.Measure, true
	case c.Expression != nil:
		return qm.Name, true
	}
	return "", false
}

// hasExactMeasure reports whether any of the measures, or any measure of a subquery in the expressions, is flagged Exact.
// It allocates nothing so that the common case of a query without exact measures stays cheap.
func hasExactMeasure(measures []Measure, exprs ...*Expression) bool {
	for _, m := range measures {
		if m.Exact {
			return true
		}
	}
	for _, e := range exprs {
		if e == nil {
			continue
		}
		if e.Subquery != nil && hasExactMeasure(e.Subquery.Measures, e.Subquery.Where, e.Subquery.Having) {
			return true
		}
		if e.Condition != nil && hasExactMeasure(nil, e.Condition.Expressions...) {
			return true
		}
	}
	return false
}

// appendQueryMeasures appends the measures and the measures of every subquery in the expressions (recursively) to res.
func appendQueryMeasures(res, measures []Measure, exprs ...*Expression) []Measure {
	res = append(res, measures...)
	for _, e := range exprs {
		if e == nil {
			continue
		}
		if e.Subquery != nil {
			res = appendQueryMeasures(res, e.Subquery.Measures, e.Subquery.Where, e.Subquery.Having)
		}
		if e.Condition != nil {
			res = appendQueryMeasures(res, nil, e.Condition.Expressions...)
		}
	}
	return res
}
