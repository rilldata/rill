package druid_test

import (
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers/druid"
	"github.com/stretchr/testify/require"
)

// TestDateTruncExprMultiples pins the SQL for multiples.
// The metrics_time_grains_druid resolver tests run it against a Druid cluster across a DST change.
func TestDateTruncExprMultiples(t *testing.T) {
	dim := &runtimev1.MetricsViewSpec_Dimension{Column: "__time"}

	expr, err := druid.DialectDruid.DateTruncExpr(dim, runtimev1.TimeGrain_TIME_GRAIN_10_MINUTE, "America/New_York", 1, 1)
	require.NoError(t, err)
	require.Equal(t, `time_floor("__time", 'PT10M', null, 'America/New_York')`, expr)

	expr, err = druid.DialectDruid.DateTruncExpr(dim, runtimev1.TimeGrain_TIME_GRAIN_6_HOUR, "", 1, 1)
	require.NoError(t, err)
	require.Equal(t, `time_floor("__time", 'PT6H')`, expr)

	// Hour multiples with a time zone are built from local midnight, not from time_floor's absolute-hour origin.
	expr, err = druid.DialectDruid.DateTruncExpr(dim, runtimev1.TimeGrain_TIME_GRAIN_6_HOUR, "America/New_York", 1, 1)
	require.NoError(t, err)
	hour := `time_extract("__time", 'HOUR', 'America/New_York')`
	bucketHour := "(" + hour + " - mod(" + hour + ", 6))"
	fromMidnight := `time_shift(time_floor("__time", 'P1D', null, 'America/New_York'), 'PT1H', ` + bucketHour + `, 'America/New_York')`
	fromHour := `time_shift(time_floor("__time", 'PT1H', null, 'America/New_York'), 'PT1H', -mod(` + hour + `, 6), 'America/New_York')`
	require.Equal(t, "CASE WHEN time_extract("+fromMidnight+", 'HOUR', 'America/New_York') = "+bucketHour+" THEN "+fromMidnight+" ELSE "+fromHour+" END", expr)
}
