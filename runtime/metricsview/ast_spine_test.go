package metricsview

import (
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/drivers/clickhouse"
	"github.com/rilldata/rill/runtime/drivers/druid"
	"github.com/rilldata/rill/runtime/drivers/pinot"
	"github.com/rilldata/rill/runtime/drivers/snowflake"
	"github.com/rilldata/rill/runtime/drivers/starrocks"
	"github.com/stretchr/testify/require"
)

// A time spine over an empty time range has no bins.
// Dialects that inline the bins (ClickHouse, Druid, Pinot, Snowflake, StarRocks) cannot express a spine with zero rows,
// so the spine must instead be derived from the underlying table, where the empty time range guarantees no rows.
func TestTimeSpineEmptyRangeSQL(t *testing.T) {
	mv := &runtimev1.MetricsViewSpec{
		Table:         "test_table",
		TimeDimension: "ts",
		Dimensions: []*runtimev1.MetricsViewSpec_Dimension{
			{Name: "ts", Column: "ts"},
		},
		Measures: []*runtimev1.MetricsViewSpec_Measure{
			{Name: "count", Expression: "count(*)", Type: runtimev1.MetricsViewSpec_MEASURE_TYPE_SIMPLE},
		},
	}
	sec := skipMetricsViewSecurity{}

	day := time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		dialect drivers.Dialect
		wantSQL string
	}{
		{
			dialect: clickhouse.DialectClickhouse,
			wantSQL: `SELECT ("t1"."ts") AS "ts", ("t2"."count") AS "count" FROM (SELECT (date_trunc('day', "ts")::DateTime64) AS "ts", (count(*)) AS "count" FROM "test_table" WHERE ("ts" >= ? AND "ts" < ?) GROUP BY 1) t2 RIGHT JOIN (SELECT (date_trunc('day', "ts")::DateTime64) AS "ts" FROM "test_table" WHERE ("ts" >= ? AND "ts" < ?) GROUP BY 1) t1 ON (isNotDistinctFrom("t2"."ts", "t1"."ts"))`,
		},
		{
			dialect: druid.DialectDruid,
			wantSQL: `SELECT ("t1"."ts") AS "ts", ("t2"."count") AS "count" FROM (SELECT (date_trunc('DAY', "ts")) AS "ts", (count(*)) AS "count" FROM "test_table" WHERE ("ts" >= ? AND "ts" < ?) GROUP BY 1) t2 RIGHT JOIN (SELECT (date_trunc('DAY', "ts")) AS "ts" FROM "test_table" WHERE ("ts" >= ? AND "ts" < ?) GROUP BY 1) t1 ON ("t2"."ts" IS NOT DISTINCT FROM "t1"."ts")`,
		},
		{
			dialect: pinot.DialectPinot,
			wantSQL: `SELECT ("t1"."ts") AS "ts", ("t2"."count") AS "count" FROM (SELECT (CAST(date_trunc('DAY', "ts", 'MILLISECONDS') AS TIMESTAMP)) AS "ts", (count(*)) AS "count" FROM "test_table" WHERE ("ts" >= CAST(? AS TIMESTAMP) AND "ts" < CAST(? AS TIMESTAMP)) GROUP BY 1) t2 RIGHT JOIN (SELECT (CAST(date_trunc('DAY', "ts", 'MILLISECONDS') AS TIMESTAMP)) AS "ts" FROM "test_table" WHERE ("ts" >= CAST(? AS TIMESTAMP) AND "ts" < CAST(? AS TIMESTAMP)) GROUP BY 1) t1 ON ("t2"."ts" IS NOT DISTINCT FROM "t1"."ts")`,
		},
		{
			dialect: snowflake.DialectSnowflake,
			wantSQL: `SELECT (t1."ts") AS "ts", (t2."count") AS "count" FROM (SELECT (date_trunc('DAY', ts::TIMESTAMP)) AS "ts", (count(*)) AS "count" FROM test_table WHERE (ts >= ? AND ts < ?) GROUP BY 1) t2 RIGHT JOIN (SELECT (date_trunc('DAY', ts::TIMESTAMP)) AS "ts" FROM test_table WHERE (ts >= ? AND ts < ?) GROUP BY 1) t1 ON (t2."ts" IS NOT DISTINCT FROM t1."ts")`,
		},
		{
			dialect: starrocks.DialectStarRocks,
			wantSQL: "SELECT (`t1`.`ts`) AS `ts`, (`t2`.`count`) AS `count` FROM (SELECT (date_trunc('DAY', `ts`)) AS `ts`, (count(*)) AS `count` FROM `test_table` WHERE (`ts` >= ? AND `ts` < ?) GROUP BY 1) t2 RIGHT JOIN (SELECT (date_trunc('DAY', `ts`)) AS `ts` FROM `test_table` WHERE (`ts` >= ? AND `ts` < ?) GROUP BY 1) t1 ON (`t2`.`ts` <=> `t1`.`ts`)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.dialect.String(), func(t *testing.T) {
			qry := &Query{
				MetricsView: "test",
				Dimensions: []Dimension{{
					Name:    "ts",
					Compute: &DimensionCompute{TimeFloor: &DimensionComputeTimeFloor{Dimension: "ts", Grain: TimeGrainDay}},
				}},
				Measures:  []Measure{{Name: "count"}},
				TimeRange: &TimeRange{Start: day, End: day},
				Spine:     &Spine{TimeRange: &TimeSpine{Start: day, End: day, Grain: TimeGrainDay}},
			}

			ast, err := NewAST(mv, sec, qry, tt.dialect)
			require.NoError(t, err)

			sql, args, err := ast.SQL()
			require.NoError(t, err)
			require.Equal(t, tt.wantSQL, sql)
			require.Equal(t, []any{day, day, day, day}, args)
		})
	}
}
