package executor

import (
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers/clickhouse"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/stretchr/testify/require"
)

func TestRewriteQueryDedupeDimensions(t *testing.T) {
	dims := func(d ...metricsview.Dimension) []metricsview.Dimension { return d }
	plain := metricsview.Dimension{Name: "ts"}
	month := metricsview.Dimension{Name: "ts", Compute: &metricsview.DimensionCompute{TimeFloor: &metricsview.DimensionComputeTimeFloor{Dimension: "ts", Grain: metricsview.TimeGrainMonth}}}
	day := metricsview.Dimension{Name: "ts", Compute: &metricsview.DimensionCompute{TimeFloor: &metricsview.DimensionComputeTimeFloor{Dimension: "ts", Grain: metricsview.TimeGrainDay}}}
	publisher := metricsview.Dimension{Name: "publisher"}

	tests := []struct {
		name    string
		dims    []metricsview.Dimension
		want    []metricsview.Dimension
		wantErr string
	}{
		{name: "distinct names are kept in order", dims: dims(publisher, plain), want: dims(publisher, plain)},
		{name: "a time floor replaces the plain dimension before it", dims: dims(plain, month, publisher), want: dims(month, publisher)},
		{name: "a time floor is kept over the plain dimension after it", dims: dims(month, publisher, plain), want: dims(month, publisher)},
		{name: "the same plain dimension twice keeps one", dims: dims(plain, plain), want: dims(plain)},
		{name: "the same time floor twice keeps one", dims: dims(month, month), want: dims(month)},
		{name: "different time floors under one name are rejected", dims: dims(day, month), wantErr: `dimension "ts" is requested multiple times with different computations`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qry := &metricsview.Query{Dimensions: tt.dims}
			err := (&Executor{}).rewriteQueryDedupeDimensions(qry)
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, qry.Dimensions)
		})
	}
}

// The time dimension detail view compared by the metrics view's own time dimension requests that dimension both plain and at the selected grain.
// Without the dedupe, the ClickHouse computed time dimension wrapper gave both inner fields the same unique alias, which ClickHouse rejects.
func TestRewriteQueryDedupeDimensionsClickHouseSQL(t *testing.T) {
	mv := &runtimev1.MetricsViewSpec{
		Table:         "events",
		TimeDimension: "ts",
		Dimensions:    []*runtimev1.MetricsViewSpec_Dimension{{Name: "ts", Column: "ts"}},
		Measures:      []*runtimev1.MetricsViewSpec_Measure{{Name: "events", Expression: "count(*)", Type: runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE}},
	}
	qry := &metricsview.Query{
		Dimensions: []metricsview.Dimension{
			{Name: "ts"},
			{Name: "ts", Compute: &metricsview.DimensionCompute{TimeFloor: &metricsview.DimensionComputeTimeFloor{Dimension: "ts", Grain: metricsview.TimeGrainMonth}}},
		},
		Measures: []metricsview.Measure{{Name: "events"}},
		Sort:     []metricsview.Sort{{Name: "events", Desc: true}, {Name: "ts"}},
	}

	e := &Executor{}
	require.NoError(t, e.rewriteQueryDedupeDimensions(qry))
	ast, err := metricsview.NewAST(mv, runtime.ResolvedSecurityOpen, qry, clickhouse.DialectClickhouse)
	require.NoError(t, err)
	require.NoError(t, e.rewriteLimitsIntoSubqueries(ast))
	require.NoError(t, e.wrapClickhouseComputedTimeDim(ast))
	sql, _, err := ast.SQL()
	require.NoError(t, err)
	require.Equal(t, `SELECT ("t1"."ts_uniq_c3868e8b") AS "ts", ("t1"."events") AS "events" FROM (SELECT (date_trunc('month', "ts")::DateTime64) AS "ts_uniq_c3868e8b", (count(*)) AS "events" FROM "events" GROUP BY 1 ORDER BY "events" DESC NULLS LAST, "ts_uniq_c3868e8b" NULLS LAST) t1 ORDER BY "events" DESC NULLS LAST, "ts" NULLS LAST`, sql)
}
