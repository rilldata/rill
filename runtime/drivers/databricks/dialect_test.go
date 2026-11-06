package databricks_test

import (
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers/databricks"
	"github.com/stretchr/testify/require"
)

func TestSelectTimeRangeBinsMultiples(t *testing.T) {
	ny, err := time.LoadLocation("America/New_York")
	require.NoError(t, err)
	start := time.Date(2024, 11, 3, 4, 0, 0, 0, time.UTC)
	end := time.Date(2024, 11, 4, 5, 0, 0, 0, time.UTC)

	sql, args, err := databricks.DialectDatabricks.SelectTimeRangeBins(start, end, runtimev1.TimeGrain_TIME_GRAIN_10_MINUTE, "ts", ny, 1, 1)
	require.NoError(t, err)
	require.Nil(t, args)
	require.Equal(t, "SELECT ts AS `ts` FROM (SELECT EXPLODE(SEQUENCE(CAST('2024-11-03T04:00:00Z' AS TIMESTAMP), CAST('2024-11-04T05:00:00Z' AS TIMESTAMP), INTERVAL 10 MINUTE)) AS ts) WHERE ts < CAST('2024-11-04T05:00:00Z' AS TIMESTAMP)", sql)

	// Hour multiples step on the local wall clock, so the fall-back day still yields bins at local 00:00, 06:00, 12:00 and 18:00.
	sql, args, err = databricks.DialectDatabricks.SelectTimeRangeBins(start, end, runtimev1.TimeGrain_TIME_GRAIN_6_HOUR, "ts", ny, 1, 1)
	require.NoError(t, err)
	require.Nil(t, args)
	require.Equal(t, "SELECT TO_UTC_TIMESTAMP(ts, 'America/New_York') AS `ts` FROM (SELECT EXPLODE(SEQUENCE(FROM_UTC_TIMESTAMP(CAST('2024-11-03T04:00:00Z' AS TIMESTAMP), 'America/New_York'), FROM_UTC_TIMESTAMP(CAST('2024-11-04T05:00:00Z' AS TIMESTAMP), 'America/New_York'), INTERVAL 6 HOUR)) AS ts) WHERE TO_UTC_TIMESTAMP(ts, 'America/New_York') < CAST('2024-11-04T05:00:00Z' AS TIMESTAMP)", sql)
}
