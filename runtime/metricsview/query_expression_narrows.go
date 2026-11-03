package metricsview

import (
	"fmt"
	"reflect"
)

// ValidateExpressionNarrows validates that expr narrows base: that it applies every dimension filter of base, so that on those dimensions it selects at most the rows that base selects.
// A filter that selects values (eq or in) may be restated or narrowed to a subset of its values, but not widened or changed.
// Filters that don't select values, such as exclusions, ranges and patterns, are matched by dimension only, since they can be restated in many equivalent ways.
// Only the dimension filters that base requires are checked: the terms of its (nested) ANDs that compare a dimension to literal values. ORs and subqueries are skipped.
// It returns an error describing the first filter of base that expr drops or widens.
func ValidateExpressionNarrows(expr, base *Expression) error {
	filters := dimensionFilters(expr)
	for _, bf := range dimensionFilters(base) {
		var found bool
		for _, f := range filters {
			if f.dimension != bf.dimension {
				continue
			}
			found = true
			if bf.selects() && f.selects() && !f.within(bf) {
				return fmt.Errorf("%s is not within %s", f, bf)
			}
		}
		if !found {
			return fmt.Errorf("no filter on %q", bf.dimension)
		}
	}
	return nil
}

// dimensionFilter is a condition that compares a dimension to literal values, such as country IN ('DK', 'US').
type dimensionFilter struct {
	dimension string
	op        Operator
	values    []any
	expr      *Expression
}

// selects reports whether the filter selects a set of values (eq or in), as opposed to excluding, ranging or pattern matching.
func (f dimensionFilter) selects() bool {
	return f.op == OperatorEq || f.op == OperatorIn
}

// within reports whether every value selected by f is also selected by other.
func (f dimensionFilter) within(other dimensionFilter) bool {
	allowed := make(map[string]bool, len(other.values))
	for _, v := range other.values {
		allowed[fmt.Sprint(v)] = true
	}
	for _, v := range f.values {
		if !allowed[fmt.Sprint(v)] {
			return false
		}
	}
	return true
}

// String renders the filter as SQL for error messages.
func (f dimensionFilter) String() string {
	sql, err := ExpressionToSQL(f.expr)
	if err != nil {
		return fmt.Sprintf("%s %s %v", f.dimension, f.op, f.values)
	}
	return sql
}

// dimensionFilters returns the dimension filters that the expression requires: the expression itself, or the terms of its (nested) ANDs.
// Terms that can't be matched by dimension, such as ORs and subqueries, are skipped.
// Like the SQL builders, it accepts the values of IN and NIN either as one list or as a flattened list of value expressions.
func dimensionFilters(e *Expression) []dimensionFilter {
	if e == nil || e.Condition == nil {
		return nil
	}
	if e.Condition.Operator == OperatorAnd {
		var res []dimensionFilter
		for _, sub := range e.Condition.Expressions {
			res = append(res, dimensionFilters(sub)...)
		}
		return res
	}

	exprs := e.Condition.Expressions
	if len(exprs) < 2 {
		return nil
	}
	for _, x := range exprs {
		if x == nil {
			return nil
		}
	}
	// The dimension is usually the first expression, but a comparison may also be written value-first.
	dim, vals := exprs[0], exprs[1:]
	if dim.Name == "" && len(exprs) == 2 {
		dim, vals = exprs[1], exprs[:1]
	}
	if dim.Name == "" {
		return nil
	}
	if len(vals) > 1 && e.Condition.Operator != OperatorIn && e.Condition.Operator != OperatorNin {
		return nil
	}
	var values []any
	for _, v := range vals {
		if v.Name != "" || v.Condition != nil || v.Subquery != nil {
			return nil
		}
		if rv := reflect.ValueOf(v.Value); rv.Kind() == reflect.Slice {
			for i := range rv.Len() {
				values = append(values, rv.Index(i).Interface())
			}
		} else {
			values = append(values, v.Value)
		}
	}
	return []dimensionFilter{{dimension: dim.Name, op: e.Condition.Operator, values: values, expr: e}}
}
