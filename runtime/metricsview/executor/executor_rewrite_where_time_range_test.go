package executor

import (
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/metricsview/metricssql"
	"github.com/stretchr/testify/require"
)

func TestRewriteQueryWhereTimeRange(t *testing.T) {
	jan1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	jan15 := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	feb1 := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)
	ist := time.FixedZone("IST", 19800)
	name := func(n string) *metricsview.Expression { return &metricsview.Expression{Name: n} }
	val := func(v any) *metricsview.Expression { return &metricsview.Expression{Value: v} }
	cond := func(op metricsview.Operator, exprs ...*metricsview.Expression) *metricsview.Expression {
		return &metricsview.Expression{Condition: &metricsview.Condition{Operator: op, Expressions: exprs}}
	}
	subquery := &metricsview.Expression{Subquery: &metricsview.Subquery{Dimension: metricsview.Dimension{Name: "publisher"}}}

	cases := []struct {
		name string
		// filter is metrics SQL; where takes precedence when set
		filter    string
		where     *metricsview.Expression
		mutate    func(q *metricsview.Query)
		wantStart time.Time
		wantEnd   time.Time
		// wantMoved is false if the query should be left unchanged
		wantMoved bool
		// wantWhere is the metrics SQL of the remaining WHERE after moving; empty means nil
		wantWhere string
	}{
		{name: "both_bounds", filter: "timestamp >= '2024-01-01T00:00:00Z' AND timestamp < '2024-02-01T00:00:00Z'", wantStart: jan1, wantEnd: feb1, wantMoved: true},
		{name: "keeps_other_conditions", filter: "timestamp >= '2024-01-01T00:00:00Z' AND publisher = 'Google' AND domain = 'news.com' AND timestamp < '2024-02-01T00:00:00Z'", wantStart: jan1, wantEnd: feb1, wantWhere: "publisher = 'Google' AND domain = 'news.com'", wantMoved: true},
		{name: "one_other_condition", filter: "publisher = 'Google' AND timestamp >= '2024-01-01T00:00:00Z'", wantStart: jan1, wantWhere: "publisher = 'Google'", wantMoved: true},
		{name: "tightest_bound", filter: "timestamp >= '2024-01-01T00:00:00Z' AND timestamp >= '2024-01-15T00:00:00Z' AND timestamp < '2024-02-01T00:00:00Z'", wantStart: jan15, wantEnd: feb1, wantMoved: true},
		{name: "mirrored", filter: "'2024-01-01T00:00:00Z' <= timestamp AND '2024-02-01T00:00:00Z' > timestamp", wantStart: jan1, wantEnd: feb1, wantMoved: true},
		{name: "offset_converted_to_utc", filter: "timestamp >= '2024-01-01T05:30:00+05:30'", wantStart: jan1, wantMoved: true},
		{name: "time_value", where: cond(metricsview.OperatorGte, name("timestamp"), val(jan1.In(ist))), wantStart: jan1, wantMoved: true},
		{name: "no_time_filter", filter: "publisher = 'Google' AND domain = 'news.com'"},
		{name: "other_operators_stay", filter: "timestamp > '2024-01-01T00:00:00Z' AND timestamp <= '2024-02-01T00:00:00Z'"},
		{name: "or_stays", filter: "timestamp >= '2024-01-01T00:00:00Z' OR publisher = 'Google'"},
		{name: "value_without_zone_stays", filter: "timestamp >= '2024-01-01' AND publisher = 'Google'"},
		{name: "moves_only_simple_bounds", filter: "timestamp >= '2024-01-01T00:00:00Z' AND timestamp <= '2024-02-01T00:00:00Z'", wantStart: jan1, wantWhere: "timestamp <= '2024-02-01T00:00:00Z'", wantMoved: true},
		{
			name:   "existing_time_range",
			filter: "timestamp >= '2024-01-01T00:00:00Z'",
			mutate: func(q *metricsview.Query) { q.TimeRange = &metricsview.TimeRange{Start: jan15} },
		},
		{
			name:   "comparison_time_range",
			filter: "timestamp >= '2024-01-01T00:00:00Z'",
			mutate: func(q *metricsview.Query) { q.ComparisonTimeRange = &metricsview.TimeRange{Start: jan15} },
		},
		{
			name:  "where_subquery",
			where: cond(metricsview.OperatorAnd, cond(metricsview.OperatorGte, name("timestamp"), val(jan1)), cond(metricsview.OperatorIn, name("publisher"), subquery)),
		},
		{
			name:   "having_subquery",
			filter: "timestamp >= '2024-01-01T00:00:00Z'",
			mutate: func(q *metricsview.Query) { q.Having = cond(metricsview.OperatorIn, name("publisher"), subquery) },
		},
		{
			name:   "spine_where",
			filter: "timestamp >= '2024-01-01T00:00:00Z'",
			mutate: func(q *metricsview.Query) {
				q.Spine = &metricsview.Spine{Where: &metricsview.WhereSpine{Expression: cond(metricsview.OperatorEq, name("publisher"), val("Google"))}}
			},
		},
		{
			name:   "computed_dimension_shadows_time",
			filter: "timestamp >= '2024-01-01T00:00:00Z'",
			mutate: func(q *metricsview.Query) {
				q.Dimensions = []metricsview.Dimension{{Name: "timestamp", Compute: &metricsview.DimensionCompute{TimeFloor: &metricsview.DimensionComputeTimeFloor{Dimension: "timestamp", Grain: metricsview.TimeGrainMonth}}}}
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			where := tc.where
			if where == nil {
				var err error
				where, err = metricssql.ParseFilter(tc.filter)
				require.NoError(t, err)
			}
			qry := &metricsview.Query{Where: where}
			if tc.mutate != nil {
				tc.mutate(qry)
			}
			original := qry.TimeRange

			e := &Executor{metricsView: &runtimev1.MetricsViewSpec{TimeDimension: "timestamp"}}
			e.rewriteQueryWhereTimeRange(qry)

			if !tc.wantMoved {
				require.Same(t, original, qry.TimeRange)
				require.Same(t, where, qry.Where)
				return
			}
			require.NotNil(t, qry.TimeRange)
			require.True(t, tc.wantStart.Equal(qry.TimeRange.Start), "start: got %v, want %v", qry.TimeRange.Start, tc.wantStart)
			require.True(t, tc.wantEnd.Equal(qry.TimeRange.End), "end: got %v, want %v", qry.TimeRange.End, tc.wantEnd)
			require.Equal(t, time.UTC, qry.TimeRange.Start.Location())
			require.Equal(t, time.UTC, qry.TimeRange.End.Location())
			if tc.wantWhere == "" {
				require.Nil(t, qry.Where)
				return
			}
			wantWhere, err := metricssql.ParseFilter(tc.wantWhere)
			require.NoError(t, err)
			require.Equal(t, flattenAnd(wantWhere), flattenAnd(qry.Where))
		})
	}
}

// flattenAnd returns the conjuncts of nested AND expressions, so expressions can be compared regardless of how the ANDs are nested.
func flattenAnd(expr *metricsview.Expression) []*metricsview.Expression {
	if expr != nil && expr.Condition != nil && expr.Condition.Operator == metricsview.OperatorAnd {
		var res []*metricsview.Expression
		for _, sub := range expr.Condition.Expressions {
			res = append(res, flattenAnd(sub)...)
		}
		return res
	}
	return []*metricsview.Expression{expr}
}
