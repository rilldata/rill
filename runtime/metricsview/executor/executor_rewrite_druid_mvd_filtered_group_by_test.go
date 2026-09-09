package executor

import (
	"strings"
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers/druid"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/stretchr/testify/require"
)

func TestDruidMVDRestrictions(t *testing.T) {
	mv := &runtimev1.MetricsViewSpec{
		Table:         "events",
		TimeDimension: "__time",
		Dimensions: []*runtimev1.MetricsViewSpec_Dimension{
			{Name: "tags", Column: "tags", Unnest: true},
			{Name: "cats", Column: "cats", Unnest: true},
			{Name: "tags_expr", Expression: "UPPER(tags)", Unnest: true},
			{Name: "tags_lookup", Column: "tags", Unnest: true, LookupTable: "lk", LookupKeyColumn: "k", LookupValueColumn: "v"},
			{Name: "city", Column: "city"},
		},
	}

	cond := func(op metricsview.Operator, exprs ...*metricsview.Expression) *metricsview.Expression {
		return &metricsview.Expression{Condition: &metricsview.Condition{Operator: op, Expressions: exprs}}
	}
	name := func(n string) *metricsview.Expression { return &metricsview.Expression{Name: n} }
	val := func(v any) *metricsview.Expression { return &metricsview.Expression{Value: v} }
	in := func(dim string, vals ...any) *metricsview.Expression {
		return cond(metricsview.OperatorIn, name(dim), val(vals))
	}
	eq := func(dim string, v any) *metricsview.Expression {
		return cond(metricsview.OperatorEq, name(dim), val(v))
	}
	ilike := func(dim, pattern string) *metricsview.Expression {
		return cond(metricsview.OperatorIlike, name(dim), val(pattern))
	}
	and := func(exprs ...*metricsview.Expression) *metricsview.Expression {
		return cond(metricsview.OperatorAnd, exprs...)
	}
	dims := func(names ...string) []metricsview.Dimension {
		res := make([]metricsview.Dimension, len(names))
		for i, n := range names {
			res[i] = metricsview.Dimension{Name: n}
		}
		return res
	}
	values := func(vals ...string) druidMVDRestriction { return druidMVDRestriction{values: vals} }
	regex := func(rx string) druidMVDRestriction { return druidMVDRestriction{regex: rx} }

	tests := []struct {
		name       string
		dimensions []metricsview.Dimension
		where      *metricsview.Expression
		rows       bool
		// noRegex disables the ILIKE narrowing (the default is to test with it enabled).
		noRegex bool
		want    map[string]druidMVDRestriction
	}{
		{
			name:       "IN filter",
			dimensions: dims("tags"),
			where:      in("tags", "a", "b"),
			want:       map[string]druidMVDRestriction{"tags": values("a", "b")},
		},
		{
			name:       "IN filter with one expression per value, as sent by the frontend",
			dimensions: dims("tags"),
			where:      cond(metricsview.OperatorIn, name("tags"), val("a"), val("b")),
			want:       map[string]druidMVDRestriction{"tags": values("a", "b")},
		},
		{
			name:       "IN filter with a single scalar value",
			dimensions: dims("tags"),
			where:      cond(metricsview.OperatorIn, name("tags"), val("a")),
			want:       map[string]druidMVDRestriction{"tags": values("a")},
		},
		{
			name:       "IN filter with a non-scalar operand is not narrowed",
			dimensions: dims("tags"),
			where:      cond(metricsview.OperatorIn, name("tags"), val("a"), name("city")),
			want:       nil,
		},
		{
			name:       "EQ filter",
			dimensions: dims("tags"),
			where:      eq("tags", "a"),
			want:       map[string]druidMVDRestriction{"tags": values("a")},
		},
		{
			name:       "ILIKE filter uses the regex the WHERE clause is compiled with",
			dimensions: dims("tags"),
			where:      ilike("tags", "%foo%"),
			want:       map[string]druidMVDRestriction{"tags": regex("^(?i).*foo.*$")},
		},
		{
			name:       "ILIKE filter is ignored when regex narrowing is disabled",
			dimensions: dims("tags"),
			where:      ilike("tags", "%foo%"),
			noRegex:    true,
			want:       nil,
		},
		{
			name:       "only the conjunct on the unnested dim is used",
			dimensions: dims("tags", "city"),
			where:      and(in("tags", "a"), eq("city", "NYC")),
			want:       map[string]druidMVDRestriction{"tags": values("a")},
		},
		{
			name:       "several IN filters on one dim are intersected, nested ANDs are flattened",
			dimensions: dims("tags"),
			where:      and(and(in("tags", "a", "b", "c"), eq("city", "NYC")), in("tags", "c", "a", "z")),
			want:       map[string]druidMVDRestriction{"tags": values("a", "c")},
		},
		{
			name:       "IN and ILIKE on one dim: the allow list is filtered by the regex",
			dimensions: dims("tags"),
			where:      and(in("tags", "apple", "banana", "Pineapple"), ilike("tags", "%APPLE%")),
			want:       map[string]druidMVDRestriction{"tags": values("apple", "Pineapple")},
		},
		{
			name:       "IN and ILIKE on one dim with nothing in common is dropped",
			dimensions: dims("tags"),
			where:      and(in("tags", "a"), ilike("tags", "%foo%")),
			want:       nil,
		},
		{
			name:       "two ILIKE filters on one dim: only the first is applied",
			dimensions: dims("tags"),
			where:      and(ilike("tags", "%foo%"), ilike("tags", "%bar%")),
			want:       map[string]druidMVDRestriction{"tags": regex("^(?i).*foo.*$")},
		},
		{
			name:       "empty intersection is dropped",
			dimensions: dims("tags"),
			where:      and(in("tags", "a"), in("tags", "b")),
			want:       nil,
		},
		{
			name:       "exclusion filters are ignored",
			dimensions: dims("tags"),
			where:      and(in("tags", "a", "b"), cond(metricsview.OperatorNin, name("tags"), val([]any{"b"})), cond(metricsview.OperatorNilike, name("tags"), val("%b%"))),
			want:       map[string]druidMVDRestriction{"tags": values("a", "b")},
		},
		{
			name:       "two unnested dims are narrowed independently",
			dimensions: dims("tags", "cats"),
			where:      and(in("tags", "a"), ilike("cats", "x%")),
			want:       map[string]druidMVDRestriction{"tags": values("a"), "cats": regex("^(?i)x.*$")},
		},
		{
			name:       "filter under an OR is not a top-level conjunct",
			dimensions: dims("tags", "city"),
			where:      cond(metricsview.OperatorOr, ilike("tags", "%a%"), eq("city", "NYC")),
			want:       nil,
		},
		{
			name:       "non-string values are not narrowed",
			dimensions: dims("tags"),
			where:      in("tags", 1, 2),
			want:       nil,
		},
		{
			name:       "null in the list is not narrowed",
			dimensions: dims("tags"),
			where:      in("tags", "a", nil),
			want:       nil,
		},
		{
			name:       "empty ILIKE pattern yields the same regex as the filter",
			dimensions: dims("tags"),
			where:      ilike("tags", ""),
			want:       map[string]druidMVDRestriction{"tags": regex("^(?i)$")},
		},
		{
			name:       "subquery filter is not narrowed",
			dimensions: dims("tags"),
			where:      cond(metricsview.OperatorIn, name("tags"), &metricsview.Expression{Subquery: &metricsview.Subquery{Dimension: metricsview.Dimension{Name: "tags"}}}),
			want:       nil,
		},
		{
			name:       "expression-backed dim is not narrowed",
			dimensions: dims("tags_expr"),
			where:      in("tags_expr", "A"),
			want:       nil,
		},
		{
			name:       "lookup dim is not narrowed",
			dimensions: dims("tags_lookup"),
			where:      in("tags_lookup", "a"),
			want:       nil,
		},
		{
			name:       "filtered but not grouped dim is not narrowed",
			dimensions: dims("city"),
			where:      in("tags", "a"),
			want:       nil,
		},
		{
			name:       "grouped but not unnested dim is not narrowed",
			dimensions: dims("city"),
			where:      eq("city", "NYC"),
			want:       nil,
		},
		{
			name:       "computed dim is ignored",
			dimensions: []metricsview.Dimension{{Name: "tags", Compute: &metricsview.DimensionCompute{TimeFloor: &metricsview.DimensionComputeTimeFloor{Dimension: "__time", Grain: metricsview.TimeGrainDay}}}},
			where:      in("tags", "a"),
			want:       nil,
		},
		{
			name:       "no filter",
			dimensions: dims("tags"),
			where:      nil,
			want:       nil,
		},
		{
			name:       "raw rows",
			dimensions: nil,
			where:      in("tags", "a"),
			rows:       true,
			want:       nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qry := &metricsview.Query{Dimensions: tt.dimensions, Where: tt.where, Rows: tt.rows}
			got := druidMVDRestrictions(mv, qry, !tt.noRegex)
			if tt.want == nil {
				require.Empty(t, got)
				return
			}
			require.Equal(t, tt.want, got)
		})
	}
}

// TestDruidMVDFilteredGroupBySQL checks the SQL produced for a narrowed dimension:
// the projected expression is wrapped in MV_FILTER_ONLY while the WHERE clause still filters the raw column, and the query keeps its single-select shape.
func TestDruidMVDFilteredGroupBySQL(t *testing.T) {
	mv := &runtimev1.MetricsViewSpec{
		Table: "events",
		Dimensions: []*runtimev1.MetricsViewSpec_Dimension{
			{Name: "tags", Column: "tags", Unnest: true},
			{Name: "city", Column: "city"},
		},
		Measures: []*runtimev1.MetricsViewSpec_Measure{
			{Name: "count", Expression: "count(*)", Type: runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE},
		},
	}
	limit := int64(8)
	qry := &metricsview.Query{
		MetricsView: "mv",
		Dimensions:  []metricsview.Dimension{{Name: "tags"}},
		Measures:    []metricsview.Measure{{Name: "count"}},
		Where: &metricsview.Expression{Condition: &metricsview.Condition{Operator: metricsview.OperatorAnd, Expressions: []*metricsview.Expression{
			// One expression per value, as the frontend sends it.
			{Condition: &metricsview.Condition{Operator: metricsview.OperatorIn, Expressions: []*metricsview.Expression{{Name: "tags"}, {Value: "MRAID-1"}, {Value: "MRAID-2"}}}},
			{Condition: &metricsview.Condition{Operator: metricsview.OperatorEq, Expressions: []*metricsview.Expression{{Name: "city"}, {Value: "NYC"}}}},
		}}},
		Sort:  []metricsview.Sort{{Name: "count", Desc: true}},
		Limit: &limit,
	}

	ast, err := metricsview.NewAST(mv, runtime.ResolvedSecurityOpen, qry, druid.DialectDruid)
	require.NoError(t, err)
	applyDruidMVDRestrictions(ast.Root, druidMVDRestrictions(mv, qry, true))

	sql, args, err := ast.SQL()
	require.NoError(t, err)
	require.Equal(t, `SELECT (MV_FILTER_ONLY("tags", ARRAY['MRAID-1', 'MRAID-2'])) AS "tags", (count(*)) AS "count" FROM "events" WHERE ((("tags") IN (?,?)) AND (("city") = ?)) GROUP BY 1 ORDER BY "count" DESC LIMIT 8`, sql)
	require.Equal(t, []any{"MRAID-1", "MRAID-2", "NYC"}, args)
}

// TestDruidMVDFilteredSearchSQL checks the SQL produced for a dimension search (an ILIKE filter, no measures):
// the projected expression is wrapped in MV_FILTER_REGEX with a literal regex identical to the one bound for the REGEXP_LIKE row filter.
func TestDruidMVDFilteredSearchSQL(t *testing.T) {
	mv := &runtimev1.MetricsViewSpec{
		Table: "events",
		Dimensions: []*runtimev1.MetricsViewSpec_Dimension{
			{Name: "tags", Column: "tags", Unnest: true},
			{Name: "city", Column: "city"},
		},
	}
	limit := int64(250)
	qry := &metricsview.Query{
		MetricsView: "mv",
		Dimensions:  []metricsview.Dimension{{Name: "tags"}},
		Where: &metricsview.Expression{Condition: &metricsview.Condition{Operator: metricsview.OperatorAnd, Expressions: []*metricsview.Expression{
			{Condition: &metricsview.Condition{Operator: metricsview.OperatorIlike, Expressions: []*metricsview.Expression{{Name: "tags"}, {Value: "%mraid-1%"}}}},
			{Condition: &metricsview.Condition{Operator: metricsview.OperatorEq, Expressions: []*metricsview.Expression{{Name: "city"}, {Value: "NYC"}}}},
		}}},
		Sort:  []metricsview.Sort{{Name: "tags"}},
		Limit: &limit,
	}

	ast, err := metricsview.NewAST(mv, runtime.ResolvedSecurityOpen, qry, druid.DialectDruid)
	require.NoError(t, err)
	applyDruidMVDRestrictions(ast.Root, druidMVDRestrictions(mv, qry, true))

	sql, args, err := ast.SQL()
	require.NoError(t, err)
	require.Equal(t, `SELECT (MV_FILTER_REGEX("tags", '^(?i).*mraid-1.*$')) AS "tags" FROM "events" WHERE ((REGEXP_LIKE(("tags"), ?)) AND (("city") = ?)) GROUP BY 1 ORDER BY "tags" LIMIT 250`, sql)
	require.Equal(t, []any{"^(?i).*mraid-1.*$", "NYC"}, args, "the bound regex must equal the literal in MV_FILTER_REGEX")
}

// TestDruidMVDFilteredGroupBySQLComparison checks that both the base and the comparison select are narrowed, and the outer join select is left alone.
func TestDruidMVDFilteredGroupBySQLComparison(t *testing.T) {
	mv := &runtimev1.MetricsViewSpec{
		Table:         "events",
		TimeDimension: "__time",
		Dimensions: []*runtimev1.MetricsViewSpec_Dimension{
			{Name: "tags", Column: "tags", Unnest: true},
		},
		Measures: []*runtimev1.MetricsViewSpec_Measure{
			{Name: "count", Expression: "count(*)", Type: runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE},
		},
	}
	qry := &metricsview.Query{
		MetricsView: "mv",
		Dimensions:  []metricsview.Dimension{{Name: "tags"}},
		Measures: []metricsview.Measure{
			{Name: "count"},
			{Name: "count_prev", Compute: &metricsview.MeasureCompute{ComparisonValue: &metricsview.MeasureComputeComparisonValue{Measure: "count"}}},
		},
		Where: &metricsview.Expression{Condition: &metricsview.Condition{Operator: metricsview.OperatorIn, Expressions: []*metricsview.Expression{{Name: "tags"}, {Value: []any{"a"}}}}},
		TimeRange: &metricsview.TimeRange{
			Start: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		ComparisonTimeRange: &metricsview.TimeRange{
			Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	ast, err := metricsview.NewAST(mv, runtime.ResolvedSecurityOpen, qry, druid.DialectDruid)
	require.NoError(t, err)
	applyDruidMVDRestrictions(ast.Root, druidMVDRestrictions(mv, qry, true))

	sql, _, err := ast.SQL()
	require.NoError(t, err)
	require.Equal(t, 2, strings.Count(sql, `MV_FILTER_ONLY("tags", ARRAY['a'])`), "generated SQL: %s", sql)
	require.Equal(t, 2, strings.Count(sql, `("tags") IN (?)`), "generated SQL: %s", sql)
}
