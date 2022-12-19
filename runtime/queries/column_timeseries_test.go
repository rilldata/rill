package queries

import (
	"context"
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/testruntime"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func instanceWith2RowsModel(t *testing.T) (*runtime.Runtime, string) {
	rt, instanceID := testruntime.NewInstanceWithModel(t, "test", `
		SELECT 1.0 AS clicks, TIMESTAMP '2019-01-01 00:00:00' AS time, DATE '2019-01-01' as day, 'android' AS device, 'Google' AS publisher, 'google.com' AS domain
		UNION ALL
		SELECT 1.0 AS clicks, TIMESTAMP '2019-01-02 00:00:00' AS time, DATE '2019-01-02' as day, 'iphone' AS device, null AS publisher, 'msn.com' AS domain
	`)
	return rt, instanceID
}

func instanceWithSparkModel(t *testing.T) (*runtime.Runtime, string) {
	rt, instanceID := testruntime.NewInstanceWithModel(t, "test", `
		SELECT 2.0 AS clicks, TIMESTAMP '2019-01-01T00:00:00Z' AS time, 'android' AS device
		UNION ALL
		SELECT 3.0 AS clicks, TIMESTAMP '2019-01-02T00:00:00Z' AS time, 'iphone' AS device
		UNION ALL
		SELECT 1.0 AS clicks, TIMESTAMP '2019-01-03T00:00:00Z' AS time, 'iphone' AS device
		UNION ALL
		SELECT 2.0 AS clicks, TIMESTAMP '2019-01-04T00:00:00Z' AS time, 'android' AS device
		UNION ALL
		SELECT 2.0 AS clicks, TIMESTAMP '2019-01-05T00:00:00Z' AS time, 'iphone' AS device
		UNION ALL
		SELECT 1.0 AS clicks, TIMESTAMP '2019-01-06T00:00:00Z' AS time, 'android' AS device
		UNION ALL
		SELECT 4.0 AS clicks, TIMESTAMP '2019-01-07T00:00:00Z' AS time, 'android' AS device
		UNION ALL
		SELECT 3 AS clicks, TIMESTAMP '2019-01-08T00:00:00Z' AS time, 'iphone' AS device
		UNION ALL
		SELECT 1.0 AS clicks, TIMESTAMP '2019-01-09T00:00:00Z' AS time, 'iphone' AS device
	`)
	return rt, instanceID
}

func TestTimeseries_normaliseTimeRange(t *testing.T) {
	rt, instanceID := instanceWith2RowsModel(t)

	q := &ColumnTimeseries{
		TableName:           "test",
		TimestampColumnName: "time",
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Interval: runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED,
		},
	}
	tr, err := q.normaliseTimeRange(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.Equal(t, parseTime(t, "2019-01-01T00:00:00.000Z"), tr.Start)
	require.Equal(t, parseTime(t, "2019-01-02T00:00:00.000Z"), tr.End)
	require.Equal(t, runtimev1.TimeGrain_TIME_GRAIN_HOUR, tr.Interval)
}

func TestTimeseries_normaliseTimeRange_NoEnd(t *testing.T) {
	rt, instanceID := instanceWith2RowsModel(t)

	q := &ColumnTimeseries{
		TableName:           "test",
		TimestampColumnName: "time",
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Interval: runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED,
			Start:    parseTime(t, "2018-01-01T00:00:00Z"),
		},
	}

	r, err := q.normaliseTimeRange(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.Equal(t, parseTime(t, "2018-01-01T00:00:00Z"), r.Start)
	require.Equal(t, parseTime(t, "2019-01-02T00:00:00.000Z"), r.End)
	require.Equal(t, runtimev1.TimeGrain_TIME_GRAIN_HOUR, r.Interval)
}

func TestTimeseries_normaliseTimeRange_Specified(t *testing.T) {
	rt, instanceID := instanceWith2RowsModel(t)

	q := &ColumnTimeseries{
		TableName:           "test",
		TimestampColumnName: "time",
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Interval: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
			Start:    parseTime(t, "2018-01-01T00:00:00Z"),
		},
	}

	r, err := q.normaliseTimeRange(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.Equal(t, parseTime(t, "2018-01-01T00:00:00Z"), r.Start)
	require.Equal(t, parseTime(t, "2019-01-02T00:00:00.000Z"), r.End)
	require.Equal(t, runtimev1.TimeGrain_TIME_GRAIN_YEAR, r.Interval)
}

func TestTimeseries_SparkOnly(t *testing.T) {
	time.Local = time.UTC

	rt, instanceID := instanceWithSparkModel(t)

	q := &ColumnTimeseries{
		TableName:           "test",
		TimestampColumnName: "time",
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Interval: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
			Start:    parseTime(t, "2018-01-01T00:00:00Z"),
		},
		Pixels: 2.0,
	}
	ctx := context.Background()
	olap, err := rt.OLAP(ctx, instanceID)
	require.NoError(t, err)
	values, err := q.createTimestampRollupReduction(context.Background(), rt, olap, instanceID, 0, "test", "time", "clicks")
	require.NoError(t, err)

	require.Equal(t, 12, len(values))
	require.Equal(t, "2019-01-01T00:00:00.000Z", values[0].Ts)
	require.Equal(t, "2019-01-02T00:00:00.000Z", values[1].Ts)
	require.Equal(t, "2019-01-03T00:00:00.000Z", values[2].Ts)
	require.Equal(t, "2019-01-04T00:00:00.000Z", values[3].Ts)
	require.Equal(t, "2019-01-05T00:00:00.000Z", values[4].Ts)
	require.Equal(t, "2019-01-06T00:00:00.000Z", values[5].Ts)
	require.Equal(t, "2019-01-07T00:00:00.000Z", values[6].Ts)
	require.Equal(t, "2019-01-08T00:00:00.000Z", values[7].Ts)
	require.Equal(t, "2019-01-09T00:00:00.000Z", values[8].Ts)
	require.Equal(t, "2019-01-09T00:00:00.000Z", values[9].Ts)
	require.Equal(t, "2019-01-09T00:00:00.000Z", values[10].Ts)
	require.Equal(t, "2019-01-09T00:00:00.000Z", values[11].Ts)

	require.Equal(t, 0.0, *values[0].Bin)
	require.Equal(t, 0.0, *values[1].Bin)
	require.Equal(t, 0.0, *values[2].Bin)
	require.Equal(t, 0.0, *values[3].Bin)
	require.Equal(t, 1.0, *values[4].Bin)
	require.Equal(t, 1.0, *values[5].Bin)
	require.Equal(t, 1.0, *values[6].Bin)
	require.Equal(t, 1.0, *values[7].Bin)
	require.Equal(t, 2.0, *values[8].Bin)
	require.Equal(t, 2.0, *values[9].Bin)
	require.Equal(t, 2.0, *values[10].Bin)
	require.Equal(t, 2.0, *values[11].Bin)

	require.Equal(t, 2.0, values[0].Records["count"])
	require.Equal(t, 3.0, values[1].Records["count"])
	require.Equal(t, 1.0, values[2].Records["count"])
	require.Equal(t, 2.0, values[3].Records["count"])
	require.Equal(t, 2.0, values[4].Records["count"])
	require.Equal(t, 1.0, values[5].Records["count"])
	require.Equal(t, 4.0, values[6].Records["count"])
	require.Equal(t, 3.0, values[7].Records["count"])
	require.Equal(t, 1.0, values[8].Records["count"])
	require.Equal(t, 1.0, values[9].Records["count"])
	require.Equal(t, 1.0, values[10].Records["count"])
	require.Equal(t, 1.0, values[11].Records["count"])
}

func TestTimeseries_Key(t *testing.T) {
	q := &ColumnTimeseries{
		TableName:           "test",
		TimestampColumnName: "time",
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Interval: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
			Start:    parseTime(t, "2018-01-01T00:00:00Z"),
		},
		Pixels: 2.0,
	}

	k1 := q.Key()
	assert.NotEmpty(t, k1)

	q = &ColumnTimeseries{
		TableName:           "test",
		TimestampColumnName: "time",
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Interval: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
			Start:    parseTime(t, "2018-01-02T00:00:00Z"),
		},
		Pixels: 2.0,
	}

	k2 := q.Key()
	assert.NotEmpty(t, k2)

	assert.NotEqual(t, k1, k2)
}

func parseTime(tst *testing.T, t string) *timestamppb.Timestamp {
	ts, err := time.Parse(time.RFC3339, t)
	require.NoError(tst, err)
	return timestamppb.New(ts)
}
