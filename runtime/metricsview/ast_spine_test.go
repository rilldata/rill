package metricsview

import (
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/drivers/clickhouse"
	"github.com/rilldata/rill/runtime/drivers/druid"
	"github.com/rilldata/rill/runtime/drivers/duckdb"
	"github.com/rilldata/rill/runtime/drivers/pinot"
	"github.com/rilldata/rill/runtime/drivers/snowflake"
	"github.com/rilldata/rill/runtime/drivers/starrocks"
	"github.com/stretchr/testify/require"
)

// A time spine over an empty time range has no bins and nothing to fill, so it must be skipped.
// Dialects that inline the bins (ClickHouse, Druid, Pinot, Snowflake, StarRocks) cannot express a spine with zero rows.
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

	dialects := []drivers.Dialect{
		clickhouse.DialectClickhouse,
		druid.DialectDruid,
		duckdb.DialectDuckDB,
		pinot.DialectPinot,
		snowflake.DialectSnowflake,
		starrocks.DialectStarRocks,
	}

	for _, dialect := range dialects {
		t.Run(dialect.String(), func(t *testing.T) {
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

			ast, err := NewAST(mv, sec, qry, dialect)
			require.NoError(t, err)

			sql, args, err := ast.SQL()
			require.NoError(t, err)
			require.NotContains(t, sql, "JOIN", "the empty spine must be skipped, not joined")
			require.Equal(t, []any{day, day}, args, "only the base select's time range should be bound")
		})
	}
}
