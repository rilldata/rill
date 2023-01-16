package queries

import (
	"context"
	"fmt"
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
	tr, err := q.resolveNormaliseTimeRange(context.Background(), rt, instanceID, 0)
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

	r, err := q.resolveNormaliseTimeRange(context.Background(), rt, instanceID, 0)
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

	r, err := q.resolveNormaliseTimeRange(context.Background(), rt, instanceID, 0)
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
	for i := 0; i < 12; i++ {
		fmt.Println(fmt.Sprintf("%s %.0f %.0f", values[i].Fields["ts"].GetStringValue(), values[i].Fields["bin"].GetNumberValue(), values[i].Fields["count"].GetNumberValue()))
	}
	require.Equal(t, "2019-01-01T00:00:00Z", values[0].Fields["ts"].GetStringValue())
	require.Equal(t, "2019-01-02T00:00:00Z", values[1].Fields["ts"].GetStringValue())
	require.Equal(t, "2019-01-03T00:00:00Z", values[2].Fields["ts"].GetStringValue())
	require.Equal(t, "2019-01-04T00:00:00Z", values[3].Fields["ts"].GetStringValue())
	require.Equal(t, "2019-01-05T00:00:00Z", values[4].Fields["ts"].GetStringValue())
	require.Equal(t, "2019-01-06T00:00:00Z", values[5].Fields["ts"].GetStringValue())
	require.Equal(t, "2019-01-07T00:00:00Z", values[6].Fields["ts"].GetStringValue())
	require.Equal(t, "2019-01-08T00:00:00Z", values[7].Fields["ts"].GetStringValue())
	require.Equal(t, "2019-01-09T00:00:00Z", values[8].Fields["ts"].GetStringValue())
	require.Equal(t, "2019-01-09T00:00:00Z", values[9].Fields["ts"].GetStringValue())
	require.Equal(t, "2019-01-09T00:00:00Z", values[10].Fields["ts"].GetStringValue())
	require.Equal(t, "2019-01-09T00:00:00Z", values[11].Fields["ts"].GetStringValue())

	require.Equal(t, 0.0, values[0].Fields["bin"].GetNumberValue())
	require.Equal(t, 0.0, values[1].Fields["bin"].GetNumberValue())
	require.Equal(t, 0.0, values[2].Fields["bin"].GetNumberValue())
	require.Equal(t, 0.0, values[3].Fields["bin"].GetNumberValue())
	require.Equal(t, 1.0, values[4].Fields["bin"].GetNumberValue())
	require.Equal(t, 1.0, values[5].Fields["bin"].GetNumberValue())
	require.Equal(t, 1.0, values[6].Fields["bin"].GetNumberValue())
	require.Equal(t, 1.0, values[7].Fields["bin"].GetNumberValue())
	require.Equal(t, 2.0, values[8].Fields["bin"].GetNumberValue())
	require.Equal(t, 2.0, values[9].Fields["bin"].GetNumberValue())
	require.Equal(t, 2.0, values[10].Fields["bin"].GetNumberValue())
	require.Equal(t, 2.0, values[11].Fields["bin"].GetNumberValue())

	require.Equal(t, 2.0, values[0].Fields["count"].GetNumberValue())
	require.Equal(t, 3.0, values[1].Fields["count"].GetNumberValue())
	require.Equal(t, 1.0, values[2].Fields["count"].GetNumberValue())
	require.Equal(t, 2.0, values[3].Fields["count"].GetNumberValue())
	require.Equal(t, 2.0, values[4].Fields["count"].GetNumberValue())
	require.Equal(t, 1.0, values[5].Fields["count"].GetNumberValue())
	require.Equal(t, 4.0, values[6].Fields["count"].GetNumberValue())
	require.Equal(t, 3.0, values[7].Fields["count"].GetNumberValue())
	require.Equal(t, 1.0, values[8].Fields["count"].GetNumberValue())
	require.Equal(t, 1.0, values[9].Fields["count"].GetNumberValue())
	require.Equal(t, 1.0, values[10].Fields["count"].GetNumberValue())
	require.Equal(t, 1.0, values[11].Fields["count"].GetNumberValue())
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
