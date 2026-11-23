package metricsview

import (
	"fmt"
	"reflect"
)

// ValidateExpressionNarrows validates that expr narrows base: that it applies every dimension filter of base, so that on those dimensions it selects at most the rows that base selects.
// A filter that selects values (eq or in) may be restated or narrowed to a subset of its values; any other kind of filter on the dimension is rejected, since it can't be shown to stay within the selection.
// A filter that excludes values (neq or nin) may be restated, extended to exclude more values, or replaced by a selection that avoids the excluded values.
// Filters that range or pattern match, such as lt or ilike, are matched by dimension only, since they can't be compared without the data.
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
			if !f.narrows(bf) {
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

// narrows reports whether f selects at most the rows that base selects on their shared dimension, as far as the operators and values can tell.
func (f dimensionFilter) narrows(base dimensionFilter) bool {
	switch {
	case base.selects():
		// Only a selection of a subset of the values is known to stay within a selection
		return f.selects() && f.within(base)
	case base.excludes() && f.selects():
		// A selection stays within an exclusion if it avoids the excluded values
		return !f.overlaps(base)
	case base.excludes() && f.excludes():
		// An exclusion stays within an exclusion if it excludes at least the same values
		return base.within(f)
	default:
		// Ranges and patterns can't be compared without the data, so they are matched by dimension only
		return true
	}
}

// selects reports whether the filter selects a set of values (eq or in).
func (f dimensionFilter) selects() bool {
	return f.op == OperatorEq || f.op == OperatorIn
}

// excludes reports whether the filter excludes a set of values (neq or nin).
func (f dimensionFilter) excludes() bool {
	return f.op == OperatorNeq || f.op == OperatorNin
}

// within reports whether every value of f is among the values of other.
func (f dimensionFilter) within(other dimensionFilter) bool {
	values := other.valueSet()
	for _, v := range f.values {
		if !values[fmt.Sprint(v)] {
			return false
		}
	}
	return true
}

// overlaps reports whether any value of f is among the values of other.
func (f dimensionFilter) overlaps(other dimensionFilter) bool {
	values := other.valueSet()
	for _, v := range f.values {
		if values[fmt.Sprint(v)] {
			return true
		}
	}
	return false
}

// valueSet returns the filter's values keyed by their printed form, so that equal numbers of different types match.
func (f dimensionFilter) valueSet() map[string]bool {
	res := make(map[string]bool, len(f.values))
	for _, v := range f.values {
		res[fmt.Sprint(v)] = true
	}
	return res
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
