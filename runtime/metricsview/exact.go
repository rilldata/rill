package metricsview

import (
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

// ExactMeasureSet returns the names of the metrics view measures that the query must compute with their exact_expression:
// every measure flagged Exact anywhere in the query that defines an exact_expression.
// A flag counts wherever it appears: on a plain measure, on a comparison or percent-of-total compute (it applies to the base measure), or on a measure of a subquery in the where or having expressions.
// Flags on other measures are no-ops. Only simple measures define an exact_expression; a derived measure is calculated from the measures as the query computes them.
//
// A measure name has exactly one value per query: if a measure is requested both exactly and approximately (directly, through a compute, or in a filter subquery), it is computed exactly.
// The executor resolves the set once per query and stores it in Query.ExactMeasures; queries derived from it copy the set instead of resolving their own (see NewAST).
//
// It returns nil when no measure is flagged, which is the common case.
func ExactMeasureSet(mv *runtimev1.MetricsViewSpec, qry *Query) map[string]bool {
	if !hasExactMeasure(qry.Measures, qry.Where, qry.Having) {
		return nil
	}

	byName := make(map[string]*runtimev1.MetricsViewSpec_Measure, len(mv.Measures))
	for _, m := range mv.Measures {
		byName[m.Name] = m
	}

	res := make(map[string]bool)
	for _, qm := range appendQueryMeasures(nil, qry.Measures, qry.Where, qry.Having) {
		if !qm.Exact {
			continue
		}
		name, ok := measureBase(qm)
		if !ok {
			continue
		}
		if m, ok := byName[name]; ok && m.ExactExpression != "" {
			res[name] = true
		}
	}
	if len(res) == 0 {
		return nil
	}
	return res
}

// measureBase returns the metrics view measure a query measure computes: the measure itself for a plain measure, and the base measure of a comparison or percent-of-total compute.
// Count, count distinct, URI, comparison time and expression computes compute no metrics view measure.
func measureBase(qm Measure) (string, bool) {
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
