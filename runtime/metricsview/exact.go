package metricsview

import (
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

// ExactMeasureSet returns the names of the metrics view measures that the query must compute with their exact_expression.
// It is the transitive closure of every measure flagged Exact anywhere in the query, including the measures of subqueries in its where and having expressions,
// following referenced measures (a derived measure's requires, the base measure of comparison and percent-of-total computes, and the refs of ephemeral expression measures),
// filtered to measures that define an exact_expression.
//
// The set is resolved once for the whole query because a measure name has exactly one value per query:
// if a measure is requested both exactly and approximately (directly, through a derived measure, or in a filter subquery), it is computed exactly.
// Queries derived from it (totals for percent-of-total, the first phase of a two-phase comparison, filter subqueries) must reuse the set rather than re-resolve it:
// a derived query may no longer contain the flagged measure that made another measure exact, so re-resolving could change that measure's value.
//
// It returns nil when no measure is flagged, which is the common case.
func ExactMeasureSet(mv *runtimev1.MetricsViewSpec, qry *Query) map[string]bool {
	if !hasExactMeasure(qry.Measures, qry.Where, qry.Having) {
		return nil
	}

	var measures []Measure
	measures = appendQueryMeasures(measures, qry.Measures, qry.Where, qry.Having)

	byName := make(map[string]*runtimev1.MetricsViewSpec_Measure, len(mv.Measures))
	for _, m := range mv.Measures {
		byName[m.Name] = m
	}

	// Expression measures defined in the query itself (ephemeral measures) can reference metrics view measures.
	exprMeasures := make(map[string]*MeasureComputeExpression)
	for _, qm := range measures {
		if qm.Compute != nil && qm.Compute.Expression != nil {
			exprMeasures[qm.Name] = qm.Compute.Expression
		}
	}

	res := make(map[string]bool)
	visited := make(map[string]bool)
	var visit func(name string)
	visit = func(name string) {
		if visited[name] {
			return
		}
		visited[name] = true

		if m, ok := byName[name]; ok {
			if m.ExactExpression != "" {
				res[name] = true
			}
			for _, ref := range m.ReferencedMeasures {
				visit(ref)
			}
			return
		}

		if c, ok := exprMeasures[name]; ok {
			parsed, err := ParseMeasureExpression(c.Expression)
			if err != nil {
				// Invalid expressions are reported when the AST resolves the measure.
				return
			}
			for _, ref := range parsed.Refs() {
				visit(ref)
			}
		}
	}

	for _, qm := range measures {
		if !qm.Exact {
			continue
		}
		if qm.Compute == nil {
			visit(qm.Name)
			continue
		}
		c := qm.Compute
		switch {
		case c.ComparisonValue != nil:
			visit(c.ComparisonValue.Measure)
		case c.ComparisonDelta != nil:
			visit(c.ComparisonDelta.Measure)
		case c.ComparisonRatio != nil:
			visit(c.ComparisonRatio.Measure)
		case c.PercentOfTotal != nil:
			visit(c.PercentOfTotal.Measure)
		case c.Expression != nil:
			visit(qm.Name)
		}
		// Count, count distinct, URI and comparison time computes do not reference metrics view measures.
	}

	if len(res) == 0 {
		return nil
	}
	return res
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
