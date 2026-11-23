package executor

import (
	"strings"
	"time"

	"github.com/rilldata/rill/runtime/metricsview"
)

// rewriteQueryWhereTimeRange moves time bounds on the primary time dimension from the WHERE clause into the query's time range.
// Metrics SQL (and therefore custom APIs) bounds time in the WHERE clause, so without this rewrite those queries never use features that rely on the time range, such as rollups.
//
// It only moves top-level AND conjuncts of the form `time >= <timestamp>` and `time < <timestamp>` (or their mirrored forms),
// since those map exactly onto the inclusive start and exclusive end of a time range.
// It only rewrites the query when doing so doesn't change the results:
//   - The query must not already have a time range or comparison time range.
//   - The WHERE and HAVING must not contain subqueries, and the spine must not have a WHERE: they inherit the time range but not the WHERE.
//   - No computed query dimension may have the time dimension's name: the WHERE would then filter on the computed value, but the time range filters on the time dimension.
//
// It must run after query limits are enforced: it should not apply time range limits to queries that only bound time in the WHERE.
func (e *Executor) rewriteQueryWhereTimeRange(qry *metricsview.Query) {
	timeDim := e.metricsView.TimeDimension
	if timeDim == "" || qry.Where == nil {
		return
	}
	if qry.TimeRange != nil && (!qry.TimeRange.IsZero() || (qry.TimeRange.TimeDimension != "" && qry.TimeRange.TimeDimension != timeDim)) {
		return
	}
	if qry.ComparisonTimeRange != nil && !qry.ComparisonTimeRange.IsZero() {
		return
	}
	if hasSubquery(qry.Where) || hasSubquery(qry.Having) || (qry.Spine != nil && qry.Spine.Where != nil) {
		return
	}
	for _, d := range qry.Dimensions {
		if d.Compute != nil && strings.EqualFold(d.Name, timeDim) {
			return
		}
	}

	// Split the top-level AND conjuncts into time bounds and the remaining conditions
	var start, end time.Time
	var rest []*metricsview.Expression
	var walk func(expr *metricsview.Expression)
	walk = func(expr *metricsview.Expression) {
		if expr == nil {
			return
		}
		if expr.Condition != nil && expr.Condition.Operator == metricsview.OperatorAnd {
			for _, sub := range expr.Condition.Expressions {
				walk(sub)
			}
			return
		}
		t, op, ok := timeBound(expr, timeDim)
		switch {
		case ok && op == metricsview.OperatorGte:
			if t.After(start) {
				start = t
			}
		case ok && op == metricsview.OperatorLt:
			if end.IsZero() || t.Before(end) {
				end = t
			}
		default:
			rest = append(rest, expr)
		}
	}
	walk(qry.Where)
	if start.IsZero() && end.IsZero() {
		return
	}

	qry.TimeRange = &metricsview.TimeRange{Start: start, End: end}
	switch len(rest) {
	case 0:
		qry.Where = nil
	case 1:
		qry.Where = rest[0]
	default:
		qry.Where = &metricsview.Expression{Condition: &metricsview.Condition{Operator: metricsview.OperatorAnd, Expressions: rest}}
	}
}

// timeBound parses a `time >= <timestamp>` or `time < <timestamp>` condition (or its mirrored form).
// It returns the timestamp in UTC and the normalized operator, or ok=false if the expression is not such a condition.
func timeBound(expr *metricsview.Expression, timeDim string) (time.Time, metricsview.Operator, bool) {
	if expr.Condition == nil || len(expr.Condition.Expressions) != 2 {
		return time.Time{}, "", false
	}
	isTimeDim := func(e *metricsview.Expression) bool {
		return e != nil && strings.EqualFold(e.Name, timeDim)
	}
	ident, val := expr.Condition.Expressions[0], expr.Condition.Expressions[1]
	op := expr.Condition.Operator
	if !isTimeDim(ident) {
		// Normalize `<timestamp> <= time` to `time >= <timestamp>` and `<timestamp> > time` to `time < <timestamp>`
		ident, val = val, ident
		switch op {
		case metricsview.OperatorLte:
			op = metricsview.OperatorGte
		case metricsview.OperatorGt:
			op = metricsview.OperatorLt
		default:
			return time.Time{}, "", false
		}
	}
	if !isTimeDim(ident) || val == nil || val.Name != "" {
		return time.Time{}, "", false
	}
	if op != metricsview.OperatorGte && op != metricsview.OperatorLt {
		return time.Time{}, "", false
	}

	// Only move values that identify an exact instant: the metrics SQL compiler produces either time.Time or RFC3339 strings.
	// Strings without a zone are read in the OLAP engine's or column's time zone, so they stay in the WHERE.
	// The time range passes the bound as a typed timestamp, so the instant is preserved (unlike some engines' handling of offsets in strings).
	var t time.Time
	switch v := val.Value.(type) {
	case time.Time:
		t = v
	case string:
		var err error
		t, err = time.Parse(time.RFC3339Nano, v)
		if err != nil {
			return time.Time{}, "", false
		}
	default:
		return time.Time{}, "", false
	}
	// A zero time would be read as an unbounded side of the time range
	if t.IsZero() {
		return time.Time{}, "", false
	}
	return t.UTC(), op, true
}

// hasSubquery reports whether an expression contains a subquery anywhere.
func hasSubquery(expr *metricsview.Expression) bool {
	if expr == nil {
		return false
	}
	if expr.Subquery != nil {
		return true
	}
	if expr.Condition != nil {
		for _, sub := range expr.Condition.Expressions {
			if hasSubquery(sub) {
				return true
			}
		}
	}
	return false
}
