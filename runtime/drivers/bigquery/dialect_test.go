package bigquery_test

import (
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers/bigquery"
	"github.com/stretchr/testify/require"
)

func TestDateTruncExprMultiples(t *testing.T) {
	dim := &runtimev1.MetricsViewSpec_Dimension{Column: "ts"}

	// Minute multiples add whole minutes to the local hour, which DST changes never split.
	expr, err := bigquery.DialectBigQuery.DateTruncExpr(dim, runtimev1.TimeGrain_TIME_GRAIN_10_MINUTE, "America/New_York", 1, 1)
	require.NoError(t, err)
	require.Equal(t, "TIMESTAMP_ADD(TIMESTAMP_TRUNC(`ts`, HOUR, 'America/New_York'), INTERVAL DIV(EXTRACT(MINUTE FROM `ts` AT TIME ZONE 'America/New_York'), 10) * 10 MINUTE)", expr)

	// Hour multiples floor the local DATETIME, so 06:30 EDT on the spring-forward day lands in the local 06:00 bucket.
	expr, err = bigquery.DialectBigQuery.DateTruncExpr(dim, runtimev1.TimeGrain_TIME_GRAIN_6_HOUR, "America/New_York", 1, 1)
	require.NoError(t, err)
	require.Equal(t, "TIMESTAMP(DATETIME_ADD(DATETIME_TRUNC(DATETIME(`ts`, 'America/New_York'), DAY), INTERVAL DIV(EXTRACT(HOUR FROM DATETIME(`ts`, 'America/New_York')), 6) * 6 HOUR), 'America/New_York')", expr)

	expr, err = bigquery.DialectBigQuery.DateTruncExpr(dim, runtimev1.TimeGrain_TIME_GRAIN_6_HOUR, "", 1, 1)
	require.NoError(t, err)
	require.Equal(t, "TIMESTAMP_ADD(TIMESTAMP_TRUNC(`ts`, DAY), INTERVAL DIV(EXTRACT(HOUR FROM `ts`), 6) * 6 HOUR)", expr)
}

func TestSelectTimeRangeBinsMultiples(t *testing.T) {
	ny, err := time.LoadLocation("America/New_York")
	require.NoError(t, err)
	start := time.Date(2024, 11, 3, 4, 0, 0, 0, time.UTC)
	end := time.Date(2024, 11, 4, 5, 0, 0, 0, time.UTC)

	sql, args, err := bigquery.DialectBigQuery.SelectTimeRangeBins(start, end, runtimev1.TimeGrain_TIME_GRAIN_10_MINUTE, "ts", ny, 1, 1)
	require.NoError(t, err)
	require.Nil(t, args)
	require.Equal(t, "SELECT ts AS `ts` FROM UNNEST(GENERATE_TIMESTAMP_ARRAY(TIMESTAMP '2024-11-03T04:00:00Z', TIMESTAMP '2024-11-04T05:00:00Z', INTERVAL 10 MINUTE)) AS ts WHERE ts < TIMESTAMP '2024-11-04T05:00:00Z'", sql)

	// Hour multiples step on the local wall clock; 25 hours on the fall-back day need 4 bins plus the generated margin.
	sql, args, err = bigquery.DialectBigQuery.SelectTimeRangeBins(start, end, runtimev1.TimeGrain_TIME_GRAIN_6_HOUR, "ts", ny, 1, 1)
	require.NoError(t, err)
	require.Nil(t, args)
	require.Equal(t, "SELECT ts AS `ts` FROM (SELECT TIMESTAMP(DATETIME_ADD(DATETIME '2024-11-03 00:00:00', INTERVAL i * 6 HOUR), 'America/New_York') AS ts FROM UNNEST(GENERATE_ARRAY(0, 6)) AS i) WHERE ts < TIMESTAMP '2024-11-04T05:00:00Z'", sql)

	// A spine that starts in the bucket after the spring-forward gap steps from its wall-clock start, 02:00, not from 03:00 EDT.
	sql, _, err = bigquery.DialectBigQuery.SelectTimeRangeBins(time.Date(2024, 3, 10, 7, 30, 0, 0, time.UTC), time.Date(2024, 3, 10, 13, 0, 0, 0, time.UTC), runtimev1.TimeGrain_TIME_GRAIN_2_HOUR, "ts", ny, 1, 1)
	require.NoError(t, err)
	require.Equal(t, "SELECT ts AS `ts` FROM (SELECT TIMESTAMP(DATETIME_ADD(DATETIME '2024-03-10 02:00:00', INTERVAL i * 2 HOUR), 'America/New_York') AS ts FROM UNNEST(GENERATE_ARRAY(0, 5)) AS i) WHERE ts < TIMESTAMP '2024-03-10T13:00:00Z'", sql)
}
