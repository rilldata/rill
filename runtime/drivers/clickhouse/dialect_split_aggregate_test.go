package clickhouse

import (
	"testing"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
)

func TestSplitAggregateExpression(t *testing.T) {
	tests := []struct {
		expr     string
		partials []string // Expected partial aggregate expressions, aliased "m", "m__1", ...; nil means the expression can't be split.
		final    string
	}{
		{expr: "sum(x)", partials: []string{"sumState(x)"}, final: `sumMerge("t0"."m")`},
		{expr: "count(*)", partials: []string{"countState(*)"}, final: `countMerge("t0"."m")`},
		{expr: "count()", partials: []string{"countState()"}, final: `countMerge("t0"."m")`},
		{expr: "COUNT(*)", partials: []string{"COUNTState(*)"}, final: `COUNTMerge("t0"."m")`},
		{expr: "Sum(x)", partials: []string{"SumState(x)"}, final: `SumMerge("t0"."m")`},
		// The DISTINCT keyword can't stay inside -State (ClickHouse would dedupe per group, not per merged state): count(DISTINCT) becomes uniqExact
		// and other functions get the -Distinct combinator.
		{expr: "count(DISTINCT user_id)", partials: []string{"uniqExactState(user_id)"}, final: `uniqExactMerge("t0"."m")`},
		{expr: "COUNT(distinct a, b)", partials: []string{"uniqExactState(a, b)"}, final: `uniqExactMerge("t0"."m")`},
		{expr: "sum(DISTINCT x)", partials: []string{"sumDistinctState(x)"}, final: `sumDistinctMerge("t0"."m")`},
		{expr: "countIf(DISTINCT x, c)"},
		{expr: "sum(clicks) / sum(impressions)", partials: []string{"sumState(clicks)", "sumState(impressions)"}, final: `sumMerge("t0"."m") / sumMerge("t0"."m__1")`},
		{expr: "round(sum(clicks) / nullif(sum(impressions), 0), 4)", partials: []string{"sumState(clicks)", "sumState(impressions)"}, final: `round(sumMerge("t0"."m") / nullif(sumMerge("t0"."m__1"), 0), 4)`},
		{expr: "if(sum(a) > 0, sum(b) / sum(a), 0)", partials: []string{"sumState(a)", "sumState(b)", "sumState(a)"}, final: `if(sumMerge("t0"."m") > 0, sumMerge("t0"."m__1") / sumMerge("t0"."m__2"), 0)`},
		{expr: "coalesce(sum(x), 0)", partials: []string{"sumState(x)"}, final: `coalesce(sumMerge("t0"."m"), 0)`},
		{expr: "toFloat64(sum(x)) * 100.0", partials: []string{"sumState(x)"}, final: `toFloat64(sumMerge("t0"."m")) * 100.0`},
		{expr: "-sum(x)", partials: []string{"sumState(x)"}, final: `-sumMerge("t0"."m")`},
		{expr: "CAST(sum(x) AS Float64)", partials: []string{"sumState(x)"}, final: `CAST(sumMerge("t0"."m") AS Float64)`},
		{expr: "CAST(sum(x) as Nullable(Float64))", partials: []string{"sumState(x)"}, final: `CAST(sumMerge("t0"."m") as Nullable(Float64))`},
		{expr: "sum(x)::Decimal(10, 2)", partials: []string{"sumState(x)"}, final: `sumMerge("t0"."m")::Decimal(10, 2)`},
		{expr: "sum(x) :: Float64", partials: []string{"sumState(x)"}, final: `sumMerge("t0"."m") :: Float64`},
		{expr: "CASE WHEN sum(x) > 0 THEN sum(y) / sum(x) ELSE NULL END", partials: []string{"sumState(x)", "sumState(y)", "sumState(x)"}, final: `CASE WHEN sumMerge("t0"."m") > 0 THEN sumMerge("t0"."m__1") / sumMerge("t0"."m__2") ELSE NULL END`},
		{expr: "max(ts) - INTERVAL 1 DAY", partials: []string{"maxState(ts)"}, final: `maxMerge("t0"."m") - INTERVAL 1 DAY`},
		{expr: "countIf(clicks > 0)", partials: []string{"countIfState(clicks > 0)"}, final: `countIfMerge("t0"."m")`},
		{expr: "sumIf(imp, clicks > 0)", partials: []string{"sumIfState(imp, clicks > 0)"}, final: `sumIfMerge("t0"."m")`},
		{expr: "uniqExactIf(user_id, clicks > 0)", partials: []string{"uniqExactIfState(user_id, clicks > 0)"}, final: `uniqExactIfMerge("t0"."m")`},
		{expr: "sumDistinct(x)", partials: []string{"sumDistinctState(x)"}, final: `sumDistinctMerge("t0"."m")`},
		{expr: "quantile(0.5)(x)", partials: []string{"quantileState(0.5)(x)"}, final: `quantileMerge(0.5)("t0"."m")`},
		{expr: "quantilesTDigest(0.5, 0.9)(x)", partials: []string{"quantilesTDigestState(0.5, 0.9)(x)"}, final: `quantilesTDigestMerge(0.5, 0.9)("t0"."m")`},
		{expr: "median(x)", partials: []string{"medianState(x)"}, final: `medianMerge("t0"."m")`},
		{expr: "argMax(a, b)", partials: []string{"argMaxState(a, b)"}, final: `argMaxMerge("t0"."m")`},
		{expr: "avg(x)", partials: []string{"avgState(x)"}, final: `avgMerge("t0"."m")`},
		// The approximate and exact uniq functions keep mergeable sketches in their states.
		{expr: "uniq(user_id)", partials: []string{"uniqState(user_id)"}, final: `uniqMerge("t0"."m")`},
		{expr: "uniqExact(user_id)", partials: []string{"uniqExactState(user_id)"}, final: `uniqExactMerge("t0"."m")`},
		{expr: "uniqCombined(user_id)", partials: []string{"uniqCombinedState(user_id)"}, final: `uniqCombinedMerge("t0"."m")`},
		{expr: "uniqCombined(17)(user_id)", partials: []string{"uniqCombinedState(17)(user_id)"}, final: `uniqCombinedMerge(17)("t0"."m")`},
		{expr: "uniqCombined64(user_id)", partials: []string{"uniqCombined64State(user_id)"}, final: `uniqCombined64Merge("t0"."m")`},
		{expr: "uniqHLL12(user_id)", partials: []string{"uniqHLL12State(user_id)"}, final: `uniqHLL12Merge("t0"."m")`},
		{expr: "uniqTheta(user_id)", partials: []string{"uniqThetaState(user_id)"}, final: `uniqThetaMerge("t0"."m")`},
		{expr: "uniqUpTo(5)(user_id)", partials: []string{"uniqUpToState(5)(user_id)"}, final: `uniqUpToMerge(5)("t0"."m")`},
		{expr: "uniqIf(user_id, clicks > 0)", partials: []string{"uniqIfState(user_id, clicks > 0)"}, final: `uniqIfMerge("t0"."m")`},
		{expr: "uniqCombinedIf(user_id, clicks > 0)", partials: []string{"uniqCombinedIfState(user_id, clicks > 0)"}, final: `uniqCombinedIfMerge("t0"."m")`},
		{expr: "uniqMerge(users_state)", partials: []string{"uniqMergeState(users_state)"}, final: `uniqMerge("t0"."m")`},
		{expr: "sum(length(concat(s, '(')))", partials: []string{"sumState(length(concat(s, '(')))"}, final: `sumMerge("t0"."m")`},
		{expr: "sum(arrayMax(x -> x + 1, [a, b]))", partials: []string{"sumState(arrayMax(x -> x + 1, [a, b]))"}, final: `sumMerge("t0"."m")`},
		{expr: `sum("quoted col") + sum(` + "`other`" + `)`, partials: []string{`sumState("quoted col")`, "sumState(`other`)"}, final: `sumMerge("t0"."m") + sumMerge("t0"."m__1")`},
		{expr: "sum(x) IS NOT NULL", partials: []string{"sumState(x)"}, final: `sumMerge("t0"."m") IS NOT NULL`},
		// Windows, comments, nested aggregates, states and expressions without an aggregate can't be split.
		{expr: "sum(x) OVER ()"},
		{expr: "sum(x) OVER (PARTITION BY y)"},
		{expr: "sum(x) -- comment"},
		{expr: "sum(x) /* comment */"},
		{expr: "sum(max(x))"},
		{expr: "sumState(x)"},
		{expr: "sumSimpleState(x)"},
		{expr: "uniqMergeIf(s, c)"},
		{expr: "1"},
		{expr: "x"},
		{expr: "sum(x"},
		{expr: "sum(x))"},
		{expr: "'unterminated"},
		// Identifiers outside aggregate calls are column references, including the arguments of an aggregate function that is not in the list.
		{expr: "sum(x) * y"},
		{expr: "fooAgg(x)"},
		{expr: "arraySum(x -> x, [sum(a)])"},
		{expr: "sum(x) FILTER (WHERE c)"},
		{expr: "(SELECT 1)"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			partials, final, ok := DialectClickhouse.SplitAggregateExpression(tt.expr, "m", "t0")
			if tt.partials == nil {
				require.False(t, ok)
				return
			}
			require.True(t, ok)
			want := make([]drivers.PartialAggregate, len(tt.partials))
			for i, p := range tt.partials {
				alias := "m"
				if i > 0 {
					alias = "m__" + string(rune('0'+i))
				}
				want[i] = drivers.PartialAggregate{Alias: alias, Expr: p}
			}
			require.Equal(t, want, partials)
			require.Equal(t, tt.final, final)
		})
	}
}
