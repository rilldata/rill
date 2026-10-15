package executor

import (
	"context"
	"fmt"
	"hash/crc32"
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers/clickhouse"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func lookupGroupByKeyTestMetricsView() *runtimev1.MetricsViewSpec {
	return &runtimev1.MetricsViewSpec{
		Table:         "facts",
		TimeDimension: "ts",
		Dimensions: []*runtimev1.MetricsViewSpec_Dimension{
			{Name: "ts", Column: "ts"},
			{Name: "advertiser", Column: "advertiser_key", LookupTable: "dict_adv", LookupKeyColumn: "key", LookupValueColumn: "name", LookupGroupByKey: proto.Bool(true)},
			{Name: "advertiser_category", Column: "advertiser_key", LookupTable: "dict_adv", LookupKeyColumn: "key", LookupValueColumn: "category", LookupGroupByKey: proto.Bool(true)},
			{Name: "advertiser_key", Column: "advertiser_key"},
			{Name: "campaign", Column: "campaign_key", LookupTable: "dict_camp", LookupKeyColumn: "key", LookupValueColumn: "name"},
			{Name: "region", Column: "country_code", LookupTable: "dict_country", LookupKeyColumn: "code", LookupValueColumn: "region", LookupDefaultExpression: "'Unknown'", LookupGroupByKey: proto.Bool(true)},
			{Name: "country_code", Expression: "upper(country_code)"},
			{Name: "device", Column: "device"},
			{Name: "ctr__1", Column: "device"},
		},
		Measures: []*runtimev1.MetricsViewSpec_Measure{
			{Name: "impressions", Expression: "sum(impressions)", Type: runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE},
			{Name: "ctr", Expression: "sum(clicks)/sum(impressions)", Type: runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE},
			{Name: "count", Expression: "count(*)", Type: runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE},
			{Name: "users", Expression: "count(DISTINCT user_id)", Type: runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE, TreatNullsAs: "0"},
			{Name: "running", Expression: "sum(impressions)", Type: runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE, Window: &runtimev1.MetricsViewSpec_MeasureWindow{}},
			{Name: "imp_x2", Expression: "impressions * 2", Type: runtimev1.MetricsViewSpec_MEASURE_TYPE_DERIVED, ReferencedMeasures: []string{"impressions"}},
		},
	}
}

// lookupGroupByKeyTestSQL builds the AST for the query and runs the rewrites that precede rewriteLookupGroupByKey in the executor, then the rewrite itself.
func lookupGroupByKeyTestSQL(t *testing.T, qry *metricsview.Query) (string, []any) {
	t.Helper()
	ast, err := metricsview.NewAST(lookupGroupByKeyTestMetricsView(), runtime.ResolvedSecurityOpen, qry, clickhouse.DialectClickhouse)
	require.NoError(t, err)
	e := &Executor{}
	require.NoError(t, e.rewriteLimitsIntoSubqueries(ast))
	require.NoError(t, e.wrapClickhouseComputedTimeDim(ast))
	e.rewriteLookupGroupByKey(context.Background(), ast)
	sql, args, err := ast.SQL()
	require.NoError(t, err)
	return sql, args
}

func TestRewriteLookupGroupByKey(t *testing.T) {
	dims := func(names ...string) []metricsview.Dimension {
		res := make([]metricsview.Dimension, len(names))
		for i, n := range names {
			res[i] = metricsview.Dimension{Name: n}
		}
		return res
	}
	measures := func(names ...string) []metricsview.Measure {
		res := make([]metricsview.Measure, len(names))
		for i, n := range names {
			res[i] = metricsview.Measure{Name: n}
		}
		return res
	}
	limit := int64(100)
	tsDay := fmt.Sprintf("ts_uniq_%x", crc32.ChecksumIEEE([]byte("ts")))

	tests := []struct {
		name     string
		qry      *metricsview.Query
		wantSQL  string
		wantArgs []any
	}{
		{
			name: "measures are split, the lookup is resolved once per key and the sort and limit stay on the outer select",
			qry: &metricsview.Query{
				Dimensions: dims("advertiser"),
				Measures:   measures("impressions", "ctr"),
				Sort:       []metricsview.Sort{{Name: "impressions", Desc: true}},
				Limit:      &limit,
			},
			wantSQL: `SELECT (dictGet('dict_adv', 'name', "t1"."advertiser_key")) AS "advertiser", (sumMerge("t1"."impressions")) AS "impressions", (sumMerge("t1"."ctr")/sumMerge("t1"."ctr__1")) AS "ctr" FROM (SELECT ("advertiser_key") AS "advertiser_key", (sumState(impressions)) AS "impressions", (sumState(clicks)) AS "ctr", (sumState(impressions)) AS "ctr__1" FROM "facts" GROUP BY 1) t1 GROUP BY 1 ORDER BY "impressions" DESC NULLS LAST LIMIT 100`,
		},
		{
			name: "a dimension on the raw key shares its inner field with the lookup",
			qry: &metricsview.Query{
				Dimensions: dims("advertiser_key", "advertiser"),
				Measures:   measures("impressions"),
			},
			wantSQL: `SELECT ("t1"."advertiser_key") AS "advertiser_key", (dictGet('dict_adv', 'name', "t1"."advertiser_key")) AS "advertiser", (sumMerge("t1"."impressions")) AS "impressions" FROM (SELECT ("advertiser_key") AS "advertiser_key", (sumState(impressions)) AS "impressions" FROM "facts" GROUP BY 1) t1 GROUP BY 1, 2`,
		},
		{
			name: "two lookups on the same key share one inner field",
			qry: &metricsview.Query{
				Dimensions: dims("advertiser", "advertiser_category"),
				Measures:   measures("count"),
			},
			wantSQL: `SELECT (dictGet('dict_adv', 'name', "t1"."advertiser_key")) AS "advertiser", (dictGet('dict_adv', 'category', "t1"."advertiser_key")) AS "advertiser_category", (countMerge("t1"."count")) AS "count" FROM (SELECT ("advertiser_key") AS "advertiser_key", (countState(*)) AS "count" FROM "facts" GROUP BY 1) t1 GROUP BY 1, 2`,
		},
		{
			name: "a lookup dimension without the flag and a plain dimension pass through the inner select",
			qry: &metricsview.Query{
				Dimensions: dims("advertiser", "campaign", "device"),
				Measures:   measures("impressions"),
			},
			wantSQL: `SELECT (dictGet('dict_adv', 'name', "t1"."advertiser_key")) AS "advertiser", ("t1"."campaign") AS "campaign", ("t1"."device") AS "device", (sumMerge("t1"."impressions")) AS "impressions" FROM (SELECT ("advertiser_key") AS "advertiser_key", (dictGet('dict_camp', 'name', "campaign_key")) AS "campaign", ("device") AS "device", (sumState(impressions)) AS "impressions" FROM "facts" GROUP BY 1, 2, 3) t1 GROUP BY 1, 2, 3`,
		},
		{
			name: "filters stay in the inner select, including the key subquery for a filter on the lookup dimension",
			qry: &metricsview.Query{
				Dimensions: dims("advertiser"),
				Measures:   measures("impressions"),
				Where: &metricsview.Expression{Condition: &metricsview.Condition{
					Operator:    metricsview.OperatorIn,
					Expressions: []*metricsview.Expression{{Name: "advertiser"}, {Value: []any{"A", "B"}}},
				}},
				TimeRange: &metricsview.TimeRange{Start: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), End: time.Date(2026, 1, 8, 0, 0, 0, 0, time.UTC)},
			},
			wantSQL:  `SELECT (dictGet('dict_adv', 'name', "t1"."advertiser_key")) AS "advertiser", (sumMerge("t1"."impressions")) AS "impressions" FROM (SELECT ("advertiser_key") AS "advertiser_key", (sumState(impressions)) AS "impressions" FROM "facts" WHERE ("ts" >= ? AND "ts" < ?) AND ("advertiser_key" IN (SELECT "key" FROM "dict_adv" WHERE ((name) IN (?,?)))) GROUP BY 1) t1 GROUP BY 1`,
			wantArgs: []any{time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 1, 8, 0, 0, 0, 0, time.UTC), "A", "B"},
		},
		{
			name: "a dimension-only query groups the keys first",
			qry: &metricsview.Query{
				Dimensions: dims("advertiser"),
				Sort:       []metricsview.Sort{{Name: "advertiser"}},
			},
			wantSQL: `SELECT (dictGet('dict_adv', 'name', "t1"."advertiser_key")) AS "advertiser" FROM (SELECT ("advertiser_key") AS "advertiser_key" FROM "facts" GROUP BY 1) t1 GROUP BY 1 ORDER BY "advertiser" NULLS LAST`,
		},
		{
			name: "treat_nulls_as applies to the merged value only",
			qry: &metricsview.Query{
				Dimensions: dims("advertiser"),
				Measures:   measures("users"),
			},
			wantSQL: `SELECT (dictGet('dict_adv', 'name', "t1"."advertiser_key")) AS "advertiser", COALESCE((uniqExactMerge("t1"."users")), 0) AS "users" FROM (SELECT ("advertiser_key") AS "advertiser_key", (uniqExactState(user_id)) AS "users" FROM "facts" GROUP BY 1) t1 GROUP BY 1`,
		},
		{
			name: "a default expression resolves against the key exposed under its column name",
			qry: &metricsview.Query{
				Dimensions: dims("region"),
				Measures:   measures("impressions"),
			},
			wantSQL: `SELECT (dictGetOrDefault('dict_country', 'region', "t1"."country_code", 'Unknown')) AS "region", (sumMerge("t1"."impressions")) AS "impressions" FROM (SELECT ("country_code") AS "country_code", (sumState(impressions)) AS "impressions" FROM "facts" GROUP BY 1) t1 GROUP BY 1`,
		},
		{
			name: "a window measure leaves the query untouched",
			qry: &metricsview.Query{
				Dimensions: dims("advertiser"),
				Measures:   measures("impressions", "running"),
			},
			wantSQL: `SELECT (dictGet('dict_adv', 'name', "advertiser_key")) AS "advertiser", (sum(impressions)) AS "impressions", (sum(impressions) OVER ()) AS "running" FROM "facts" GROUP BY 1`,
		},
		{
			name: "a limit pushed into the base select by the limit pushdown moves to the merging select, never into the key-grouped select",
			qry: &metricsview.Query{
				Dimensions: dims("advertiser"),
				Measures:   measures("impressions", "imp_x2"),
				Sort:       []metricsview.Sort{{Name: "impressions", Desc: true}},
				Limit:      &limit,
			},
			wantSQL: `SELECT ("t1"."advertiser") AS "advertiser", ("t1"."impressions") AS "impressions", (impressions * 2) AS "imp_x2" FROM (SELECT (dictGet('dict_adv', 'name', "t2"."advertiser_key")) AS "advertiser", (sumMerge("t2"."impressions")) AS "impressions" FROM (SELECT ("advertiser_key") AS "advertiser_key", (sumState(impressions)) AS "impressions" FROM "facts" GROUP BY 1) t2 GROUP BY 1 ORDER BY "impressions" DESC NULLS LAST LIMIT 100) t1 ORDER BY "impressions" DESC NULLS LAST`,
		},
		{
			name: "composes with the computed time dimension wrapper",
			qry: &metricsview.Query{
				Dimensions: []metricsview.Dimension{
					{Name: "ts", Compute: &metricsview.DimensionCompute{TimeFloor: &metricsview.DimensionComputeTimeFloor{Dimension: "ts", Grain: metricsview.TimeGrainDay}}},
					{Name: "advertiser"},
				},
				Measures: measures("impressions"),
			},
			wantSQL: `SELECT ("t1"."` + tsDay + `") AS "ts", ("t1"."advertiser") AS "advertiser", ("t1"."impressions") AS "impressions" FROM (SELECT ("t2"."` + tsDay + `") AS "` + tsDay + `", (dictGet('dict_adv', 'name', "t2"."advertiser_key")) AS "advertiser", (sumMerge("t2"."impressions")) AS "impressions" FROM (SELECT (date_trunc('day', "ts")::DateTime64) AS "` + tsDay + `", ("advertiser_key") AS "advertiser_key", (sumState(impressions)) AS "impressions" FROM "facts" GROUP BY 1, 2) t2 GROUP BY 1, 2) t1`,
		},
		{
			name: "the base and comparison selects are both rewritten under the comparison join",
			qry: &metricsview.Query{
				Dimensions: dims("advertiser"),
				Measures: []metricsview.Measure{
					{Name: "impressions"},
					{Name: "impressions_delta", Compute: &metricsview.MeasureCompute{ComparisonDelta: &metricsview.MeasureComputeComparisonDelta{Measure: "impressions"}}},
				},
				TimeRange:           &metricsview.TimeRange{Start: time.Date(2026, 1, 8, 0, 0, 0, 0, time.UTC), End: time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)},
				ComparisonTimeRange: &metricsview.TimeRange{Start: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), End: time.Date(2026, 1, 8, 0, 0, 0, 0, time.UTC)},
			},
			wantSQL: `SELECT (COALESCE("base"."advertiser", "comparison"."advertiser")) AS "advertiser", ("base"."impressions") AS "impressions", (base."impressions" - comparison."impressions") AS "impressions_delta" FROM (SELECT (dictGet('dict_adv', 'name', "t1"."advertiser_key")) AS "advertiser", (sumMerge("t1"."impressions")) AS "impressions" FROM (SELECT ("advertiser_key") AS "advertiser_key", (sumState(impressions)) AS "impressions" FROM "facts" WHERE ("ts" >= ? AND "ts" < ?) GROUP BY 1) t1 GROUP BY 1) base FULL OUTER JOIN (SELECT (dictGet('dict_adv', 'name', "t2"."advertiser_key")) AS "advertiser", (sumMerge("t2"."impressions")) AS "impressions" FROM (SELECT ("advertiser_key") AS "advertiser_key", (sumState(impressions)) AS "impressions" FROM "facts" WHERE ("ts" >= ? AND "ts" < ?) GROUP BY 1) t2 GROUP BY 1) comparison ON (isNotDistinctFrom("base"."advertiser", "comparison"."advertiser"))`,
			wantArgs: []any{
				time.Date(2026, 1, 8, 0, 0, 0, 0, time.UTC), time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
				time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 1, 8, 0, 0, 0, 0, time.UTC),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, args := lookupGroupByKeyTestSQL(t, tt.qry)
			require.Equal(t, tt.wantSQL, sql)
			if tt.wantArgs != nil {
				require.Equal(t, tt.wantArgs, args)
			}
		})
	}
}

func TestRewriteLookupGroupByKeySkipsWhenTheKeyAliasIsTakenAndADefaultExpressionNeedsIt(t *testing.T) {
	// The "country_code" dimension occupies the name the "region" lookup needs for its key, with a different expression,
	// so the default expression could not reference the key column; the select is left untouched.
	qry := &metricsview.Query{
		Dimensions: []metricsview.Dimension{{Name: "region"}, {Name: "country_code"}},
		Measures:   []metricsview.Measure{{Name: "impressions"}},
	}
	ast, err := metricsview.NewAST(lookupGroupByKeyTestMetricsView(), runtime.ResolvedSecurityOpen, qry, clickhouse.DialectClickhouse)
	require.NoError(t, err)
	before, _, err := ast.SQL()
	require.NoError(t, err)

	ok, reason := rewriteLookupGroupByKeyNode(ast, ast.Root)
	require.False(t, ok)
	require.Equal(t, "key_alias_conflict:region", reason)
	after, _, err := ast.SQL()
	require.NoError(t, err)
	require.Equal(t, before, after)
}

func TestRewriteLookupGroupByKeySkipsWhenADimensionNameCollidesWithAPartialAlias(t *testing.T) {
	// The "ctr" measure splits into the partials "ctr" and "ctr__1", and the "ctr__1" dimension passes through the inner select under the same name,
	// so the select is left untouched.
	qry := &metricsview.Query{
		Dimensions: []metricsview.Dimension{{Name: "advertiser"}, {Name: "ctr__1"}},
		Measures:   []metricsview.Measure{{Name: "ctr"}},
	}
	ast, err := metricsview.NewAST(lookupGroupByKeyTestMetricsView(), runtime.ResolvedSecurityOpen, qry, clickhouse.DialectClickhouse)
	require.NoError(t, err)
	before, _, err := ast.SQL()
	require.NoError(t, err)

	ok, reason := rewriteLookupGroupByKeyNode(ast, ast.Root)
	require.False(t, ok)
	require.Equal(t, "alias_conflict:ctr__1", reason)
	after, _, err := ast.SQL()
	require.NoError(t, err)
	require.Equal(t, before, after)
}
