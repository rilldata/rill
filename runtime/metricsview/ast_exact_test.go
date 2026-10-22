package metricsview

import (
	"strings"
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers/duckdb"
	"github.com/stretchr/testify/require"
)

const (
	approxUsersSQL = "approx_count_distinct(user_id)"
	exactUsersSQL  = "count(DISTINCT user_id)"
)

func exactTestMetricsView() *runtimev1.MetricsViewSpec {
	return &runtimev1.MetricsViewSpec{
		Table:         "events",
		TimeDimension: "ts",
		Dimensions: []*runtimev1.MetricsViewSpec_Dimension{
			{Name: "ts", Column: "ts", Type: runtimev1.MetricsViewSpec_DIMENSION_TYPE_TIME, DataType: &runtimev1.Type{Code: runtimev1.Type_CODE_TIMESTAMP}},
			{Name: "publisher", Column: "publisher"},
		},
		Measures: []*runtimev1.MetricsViewSpec_Measure{
			{Name: "users", Expression: approxUsersSQL, ExactExpression: exactUsersSQL, Type: runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE},
			{Name: "revenue", Expression: "sum(revenue)", Type: runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE},
			{Name: "revenue_per_user", Expression: "revenue / users", Type: runtimev1.MetricsViewSpec_MEASURE_TYPE_DERIVED, ReferencedMeasures: []string{"revenue", "users"}},
			{Name: "sessions", Expression: "approx_count_distinct(session_id)", ExactExpression: "count(DISTINCT session_id)", Type: runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE},
			{Name: "views", Expression: "approx_count_distinct(view_id)", ExactExpression: "count(DISTINCT view_id)", Type: runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE},
			{Name: "sessions_per_user", Expression: "sessions / users", Type: runtimev1.MetricsViewSpec_MEASURE_TYPE_DERIVED, ReferencedMeasures: []string{"sessions", "users"}},
			{Name: "views_per_session", Expression: "views / sessions", Type: runtimev1.MetricsViewSpec_MEASURE_TYPE_DERIVED, ReferencedMeasures: []string{"views", "sessions"}},
		},
	}
}

// usersAbove10Filter returns a `publisher IN (SELECT publisher ... HAVING users > 10)` filter whose subquery optionally requests users exactly.
func usersAbove10Filter(exact bool) *Expression {
	return &Expression{Condition: &Condition{
		Operator: OperatorIn,
		Expressions: []*Expression{
			{Name: "publisher"},
			{Subquery: &Subquery{
				Dimension: Dimension{Name: "publisher"},
				Measures:  []Measure{{Name: "users", Exact: exact}},
				Having: &Expression{Condition: &Condition{
					Operator:    OperatorGt,
					Expressions: []*Expression{{Name: "users"}, {Value: 10}},
				}},
			}},
		},
	}}
}

func TestExactMeasureSet(t *testing.T) {
	mv := exactTestMetricsView()

	tests := []struct {
		name     string
		measures []Measure
		where    *Expression
		want     map[string]bool
	}{
		{
			name:     "no flags",
			measures: []Measure{{Name: "users"}, {Name: "revenue_per_user"}},
			want:     nil,
		},
		{
			name:     "flag on a measure without an exact expression is a no-op",
			measures: []Measure{{Name: "revenue", Exact: true}},
			want:     nil,
		},
		{
			name:     "flagged simple measure",
			measures: []Measure{{Name: "users", Exact: true}, {Name: "revenue"}},
			want:     map[string]bool{"users": true},
		},
		{
			name:     "flagged derived measure resolves its referenced measures",
			measures: []Measure{{Name: "revenue_per_user", Exact: true}},
			want:     map[string]bool{"users": true},
		},
		{
			name:     "flagged comparison compute resolves its base measure",
			measures: []Measure{{Name: "users"}, {Name: "users_prev", Compute: &MeasureCompute{ComparisonValue: &MeasureComputeComparisonValue{Measure: "users"}}, Exact: true}},
			want:     map[string]bool{"users": true},
		},
		{
			name:     "flagged percent of total compute resolves its base measure",
			measures: []Measure{{Name: "users_pct", Compute: &MeasureCompute{PercentOfTotal: &MeasureComputePercentOfTotal{Measure: "users"}}, Exact: true}},
			want:     map[string]bool{"users": true},
		},
		{
			name:     "flagged expression compute resolves its referenced measures",
			measures: []Measure{{Name: "double_users", Compute: &MeasureCompute{Expression: &MeasureComputeExpression{Expression: "users * 2"}}, Exact: true}},
			want:     map[string]bool{"users": true},
		},
		{
			name:     "flagged count compute has no referenced measures",
			measures: []Measure{{Name: "cnt", Compute: &MeasureCompute{Count: true}, Exact: true}},
			want:     nil,
		},
		{
			name:     "flagged measure in a filter subquery",
			measures: []Measure{{Name: "users"}},
			where:    usersAbove10Filter(true),
			want:     map[string]bool{"users": true},
		},
		{
			name:     "flagged measure in a filter subquery nested in a condition",
			measures: []Measure{{Name: "revenue"}},
			where: &Expression{Condition: &Condition{
				Operator: OperatorAnd,
				Expressions: []*Expression{
					{Condition: &Condition{Operator: OperatorEq, Expressions: []*Expression{{Name: "publisher"}, {Value: "x"}}}},
					usersAbove10Filter(true),
				},
			}},
			want: map[string]bool{"users": true},
		},
		{
			name:     "unflagged measure in a filter subquery",
			measures: []Measure{{Name: "users"}},
			where:    usersAbove10Filter(false),
			want:     nil,
		},
		{
			name:     "a derived measure calculated from an exact measure makes its other inputs exact",
			measures: []Measure{{Name: "users", Exact: true}, {Name: "sessions"}, {Name: "sessions_per_user"}},
			want:     map[string]bool{"users": true, "sessions": true},
		},
		{
			name:     "without that derived measure in the query, its other inputs stay approximate",
			measures: []Measure{{Name: "users", Exact: true}, {Name: "sessions"}},
			want:     map[string]bool{"users": true},
		},
		{
			name:     "a derived measure whose other inputs have no exact expression adds nothing",
			measures: []Measure{{Name: "users", Exact: true}, {Name: "revenue_per_user"}},
			want:     map[string]bool{"users": true},
		},
		{
			// views_per_session comes first so that it only meets the set once sessions_per_user has added sessions, in a second round.
			name:     "measures made exact by a derived measure make further derived measures exact",
			measures: []Measure{{Name: "users", Exact: true}, {Name: "views_per_session"}, {Name: "sessions_per_user"}},
			want:     map[string]bool{"users": true, "sessions": true, "views": true},
		},
		{
			name:     "a derived measure in a filter subquery follows the same rule",
			measures: []Measure{{Name: "users", Exact: true}},
			where: &Expression{Condition: &Condition{
				Operator: OperatorIn,
				Expressions: []*Expression{
					{Name: "publisher"},
					{Subquery: &Subquery{Dimension: Dimension{Name: "publisher"}, Measures: []Measure{{Name: "sessions_per_user"}}}},
				},
			}},
			want: map[string]bool{"users": true, "sessions": true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExactMeasureSet(mv, &Query{MetricsView: "test", Measures: tt.measures, Where: tt.where})
			require.Equal(t, tt.want, got)
		})
	}
}

func TestExactMeasureSQL(t *testing.T) {
	mv := exactTestMetricsView()
	sec := skipMetricsViewSecurity{}

	timeRange := &TimeRange{
		Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
	}
	comparisonTimeRange := &TimeRange{
		Start: time.Date(2023, 12, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	tests := []struct {
		name            string
		qry             *Query
		wantExactCount  int // expected occurrences of the exact expression in the SQL
		wantApproxCount int // expected occurrences of the approximate expression in the SQL
	}{
		{
			name:            "unflagged measure uses the approximate expression",
			qry:             &Query{MetricsView: "test", Dimensions: []Dimension{{Name: "publisher"}}, Measures: []Measure{{Name: "users"}}},
			wantExactCount:  0,
			wantApproxCount: 1,
		},
		{
			name:            "flagged measure uses the exact expression",
			qry:             &Query{MetricsView: "test", Dimensions: []Dimension{{Name: "publisher"}}, Measures: []Measure{{Name: "users", Exact: true}}},
			wantExactCount:  1,
			wantApproxCount: 0,
		},
		{
			name:            "flag on a measure without an exact expression changes nothing",
			qry:             &Query{MetricsView: "test", Dimensions: []Dimension{{Name: "publisher"}}, Measures: []Measure{{Name: "users"}, {Name: "revenue", Exact: true}}},
			wantExactCount:  0,
			wantApproxCount: 1,
		},
		{
			name:            "flagged derived measure computes its referenced measure exactly",
			qry:             &Query{MetricsView: "test", Dimensions: []Dimension{{Name: "publisher"}}, Measures: []Measure{{Name: "revenue_per_user", Exact: true}}},
			wantExactCount:  1,
			wantApproxCount: 0,
		},
		{
			name:            "flagged and unflagged requests of one measure promote it to exact (unflagged first)",
			qry:             &Query{MetricsView: "test", Dimensions: []Dimension{{Name: "publisher"}}, Measures: []Measure{{Name: "users"}, {Name: "revenue_per_user", Exact: true}}},
			wantExactCount:  1,
			wantApproxCount: 0,
		},
		{
			name:            "flagged and unflagged requests of one measure promote it to exact (flagged first)",
			qry:             &Query{MetricsView: "test", Dimensions: []Dimension{{Name: "publisher"}}, Measures: []Measure{{Name: "revenue_per_user", Exact: true}, {Name: "users"}}},
			wantExactCount:  1,
			wantApproxCount: 0,
		},
		{
			name: "flagged comparison compute computes both the base and the comparison side exactly",
			qry: &Query{
				MetricsView:         "test",
				Dimensions:          []Dimension{{Name: "publisher"}},
				Measures:            []Measure{{Name: "users"}, {Name: "users_prev", Compute: &MeasureCompute{ComparisonValue: &MeasureComputeComparisonValue{Measure: "users"}}, Exact: true}},
				TimeRange:           timeRange,
				ComparisonTimeRange: comparisonTimeRange,
			},
			wantExactCount:  2,
			wantApproxCount: 0,
		},
		{
			name:            "flagged expression compute computes its referenced measure exactly",
			qry:             &Query{MetricsView: "test", Dimensions: []Dimension{{Name: "publisher"}}, Measures: []Measure{{Name: "double_users", Compute: &MeasureCompute{Expression: &MeasureComputeExpression{Expression: "users * 2"}}, Exact: true}}},
			wantExactCount:  1,
			wantApproxCount: 0,
		},
		{
			name:            "where subquery inherits the outer query's exact measures",
			qry:             &Query{MetricsView: "test", Dimensions: []Dimension{{Name: "publisher"}}, Measures: []Measure{{Name: "users", Exact: true}}, Where: usersAbove10Filter(false)},
			wantExactCount:  2,
			wantApproxCount: 0,
		},
		{
			name:            "where subquery stays approximate when the outer query is approximate",
			qry:             &Query{MetricsView: "test", Dimensions: []Dimension{{Name: "publisher"}}, Measures: []Measure{{Name: "users"}}, Where: usersAbove10Filter(false)},
			wantExactCount:  0,
			wantApproxCount: 2,
		},
		{
			name:            "a measure flagged only in a where subquery is exact in the outer query too",
			qry:             &Query{MetricsView: "test", Dimensions: []Dimension{{Name: "publisher"}}, Measures: []Measure{{Name: "users"}}, Where: usersAbove10Filter(true)},
			wantExactCount:  2,
			wantApproxCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// The executor flags the measures the exact set makes exact before it builds any AST (see FlagExactMeasures);
			// filter subqueries resolve their precision from those flags.
			FlagExactMeasures(mv, tt.qry)
			ast, err := NewAST(mv, sec, tt.qry, duckdb.DialectDuckDB)
			require.NoError(t, err)

			sql, _, err := ast.SQL()
			require.NoError(t, err)
			require.Equal(t, tt.wantExactCount, strings.Count(sql, exactUsersSQL), sql)
			require.Equal(t, tt.wantApproxCount, strings.Count(sql, approxUsersSQL), sql)
		})
	}

	// The spec measures must never be mutated by building an AST with exact measures.
	require.Equal(t, approxUsersSQL, mv.Measures[0].Expression)
	require.Equal(t, exactUsersSQL, mv.Measures[0].ExactExpression)
}

func TestFlagExactMeasures(t *testing.T) {
	mv := exactTestMetricsView()

	t.Run("nothing to flag", func(t *testing.T) {
		qry := &Query{MetricsView: "test", Measures: []Measure{{Name: "users"}, {Name: "revenue", Exact: true}}}
		require.False(t, FlagExactMeasures(mv, qry))
		// A flag on a measure without an exact expression is left in place; it has no effect.
		require.Equal(t, []Measure{{Name: "users"}, {Name: "revenue", Exact: true}}, qry.Measures)
	})

	t.Run("flags every measure the set makes exact, including in filter subqueries", func(t *testing.T) {
		qry := &Query{
			MetricsView: "test",
			Measures: []Measure{
				{Name: "users", Exact: true},
				{Name: "sessions"},
				{Name: "sessions_per_user"},
				{Name: "revenue"},
				{Name: "revenue_per_user"},
				{Name: "users_prev", Compute: &MeasureCompute{ComparisonValue: &MeasureComputeComparisonValue{Measure: "users"}}},
				{Name: "cnt", Compute: &MeasureCompute{Count: true}},
			},
			Where: usersAbove10Filter(false),
		}
		require.True(t, FlagExactMeasures(mv, qry))

		var flagged []string
		for _, m := range qry.Measures {
			if m.Exact {
				flagged = append(flagged, m.Name)
			}
		}
		// revenue has no exact variant, and the count compute is not calculated from any measure.
		require.Equal(t, []string{"users", "sessions", "sessions_per_user", "revenue_per_user", "users_prev"}, flagged)
		require.True(t, qry.Where.Condition.Expressions[1].Subquery.Measures[0].Exact)

		// The flags carry the set: a query built from any subset of the measures resolves the same precision for them.
		set := ExactMeasureSet(mv, qry)
		require.Equal(t, map[string]bool{"users": true, "sessions": true}, set)
		require.Equal(t, map[string]bool{"sessions": true}, ExactMeasureSet(mv, &Query{MetricsView: "test", Measures: []Measure{qry.Measures[1]}}))
		require.Equal(t, map[string]bool{"users": true}, ExactMeasureSet(mv, &Query{MetricsView: "test", Measures: []Measure{qry.Measures[4]}}))

		// Flagging again changes nothing.
		measures := append([]Measure(nil), qry.Measures...)
		require.True(t, FlagExactMeasures(mv, qry))
		require.Equal(t, measures, qry.Measures)
		require.Equal(t, set, ExactMeasureSet(mv, qry))
	})
}
