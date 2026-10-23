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
			{Name: "sessions", Expression: "approx_count_distinct(session_id)", ExactExpression: "count(DISTINCT session_id)", Type: runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE},
			{Name: "revenue", Expression: "sum(revenue)", Type: runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE},
			{Name: "revenue_per_user", Expression: "revenue / users", Type: runtimev1.MetricsViewSpec_MEASURE_TYPE_DERIVED, ReferencedMeasures: []string{"revenue", "users"}},
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
			name:     "every flagged measure with an exact expression",
			measures: []Measure{{Name: "users", Exact: true}, {Name: "sessions", Exact: true}},
			want:     map[string]bool{"users": true, "sessions": true},
		},
		{
			name:     "flag on a derived measure is a no-op: derived measures have no exact variant",
			measures: []Measure{{Name: "revenue_per_user", Exact: true}},
			want:     nil,
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
			name:     "flag on an expression compute is a no-op: ad hoc measures have no exact variant",
			measures: []Measure{{Name: "double_users", Compute: &MeasureCompute{Expression: &MeasureComputeExpression{Expression: "users * 2"}}, Exact: true}},
			want:     nil,
		},
		{
			name:     "flagged count compute has no base measure",
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
			name:            "flag on a derived measure leaves the measures it is calculated from approximate",
			qry:             &Query{MetricsView: "test", Dimensions: []Dimension{{Name: "publisher"}}, Measures: []Measure{{Name: "revenue_per_user", Exact: true}}},
			wantExactCount:  0,
			wantApproxCount: 1,
		},
		{
			name:            "a derived measure is calculated from the exact measure when the query computes it exactly",
			qry:             &Query{MetricsView: "test", Dimensions: []Dimension{{Name: "publisher"}}, Measures: []Measure{{Name: "users", Exact: true}, {Name: "revenue_per_user"}}},
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
			name:            "flag on an expression compute leaves the measures it references approximate",
			qry:             &Query{MetricsView: "test", Dimensions: []Dimension{{Name: "publisher"}}, Measures: []Measure{{Name: "double_users", Compute: &MeasureCompute{Expression: &MeasureComputeExpression{Expression: "users * 2"}}, Exact: true}}},
			wantExactCount:  0,
			wantApproxCount: 1,
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

func TestNewASTUsesResolvedExactMeasures(t *testing.T) {
	// The executor resolves the exact measures once per query and derived queries copy the set (see Query.ExactMeasures),
	// so a derived query that no longer contains the flagged measure still computes the measures it kept with the same precision.
	mv := exactTestMetricsView()
	sec := skipMetricsViewSecurity{}

	qry := &Query{MetricsView: "test", Measures: []Measure{{Name: "users"}}, ExactMeasures: map[string]bool{"users": true}}
	ast, err := NewAST(mv, sec, qry, duckdb.DialectDuckDB)
	require.NoError(t, err)
	sql, _, err := ast.SQL()
	require.NoError(t, err)
	require.Equal(t, 1, strings.Count(sql, exactUsersSQL), sql)
	require.Equal(t, 0, strings.Count(sql, approxUsersSQL), sql)

	// The set applies to filter subqueries as well, although none of their measures is flagged.
	qry = &Query{MetricsView: "test", Dimensions: []Dimension{{Name: "publisher"}}, Measures: []Measure{{Name: "revenue"}}, Where: usersAbove10Filter(false), ExactMeasures: map[string]bool{"users": true}}
	ast, err = NewAST(mv, sec, qry, duckdb.DialectDuckDB)
	require.NoError(t, err)
	sql, _, err = ast.SQL()
	require.NoError(t, err)
	require.Equal(t, 1, strings.Count(sql, exactUsersSQL), sql)
	require.Equal(t, 0, strings.Count(sql, approxUsersSQL), sql)

	// A query that did not pass through the executor resolves the set from its flags.
	qry = &Query{MetricsView: "test", Measures: []Measure{{Name: "users", Exact: true}}}
	ast, err = NewAST(mv, sec, qry, duckdb.DialectDuckDB)
	require.NoError(t, err)
	sql, _, err = ast.SQL()
	require.NoError(t, err)
	require.Equal(t, 1, strings.Count(sql, exactUsersSQL), sql)
	require.Equal(t, 0, strings.Count(sql, approxUsersSQL), sql)
}
