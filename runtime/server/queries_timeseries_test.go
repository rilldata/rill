package server_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/ratelimit"
	"github.com/rilldata/rill/runtime/server"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestServer_Timeseries_EmptyModel(t *testing.T) {
	t.Parallel()
	server, instanceID := getTimeseriesTestServerWithEmptyModel(t)

	response, err := server.ColumnTimeSeries(testCtx(), &runtimev1.ColumnTimeSeriesRequest{
		InstanceId: instanceID,
		TableName:  "timeseries",
		Measures: []*runtimev1.ColumnTimeSeriesRequest_BasicMeasure{
			{
				Expression: "max(clicks)",
				SqlName:    "max",
			},
		},
		TimestampColumnName: "time",
	})

	require.NoError(t, err)
	results := response.GetRollup().Results
	require.Nil(t, results)
}

func TestServer_Timeseries_Spark_NoParams(t *testing.T) {
	t.Parallel()
	server, instanceID := getTimeseriesTestServer(t)

	response, err := server.ColumnTimeSeries(testCtx(), &runtimev1.ColumnTimeSeriesRequest{
		InstanceId:          instanceID,
		TableName:           "timeseries",
		TimestampColumnName: "time",
		Pixels:              2,
	})

	require.NoError(t, err)
	require.True(t, len(response.GetRollup().Results) > 0)
	require.True(t, len(response.Rollup.Spark) > 0)
}

func TestServer_Timeseries_nulls_for_empty_intervals(t *testing.T) {
	t.Parallel()
	server, instanceID := getTimeseriesTestServer(t)

	response, err := server.ColumnTimeSeries(testCtx(), &runtimev1.ColumnTimeSeriesRequest{
		InstanceId: instanceID,
		TableName:  "timeseries",
		Measures: []*runtimev1.ColumnTimeSeriesRequest_BasicMeasure{
			{
				Expression: "max(clicks)",
				SqlName:    "max",
			},
			{
				Expression: "count(*)",
				SqlName:    "count",
			},
		},
		TimestampColumnName: "time",
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Start:    parseTimeToProtoTimeStamps(t, "2019-01-01T00:00:00Z"),
			End:      parseTimeToProtoTimeStamps(t, "2019-01-01T02:00:00Z"),
			Interval: runtimev1.TimeGrain_TIME_GRAIN_HOUR,
		},
	})

	require.NoError(t, err)
	results := response.GetRollup().Results
	require.Equal(t, 2, len(results))

	require.Equal(t, 1.0, results[0].Records.Fields["count"].GetNumberValue())
	require.Equal(t, 1.0, results[0].Records.Fields["max"].GetNumberValue())

	require.True(t, isNull(results[1].Records.Fields["count"]))
	require.True(t, isNull(results[1].Records.Fields["max"]))
}

func isNull(v *structpb.Value) bool {
	_, ok := v.Kind.(*structpb.Value_NullValue)
	return ok
}

func TestServer_Timeseries_Empty_Filter(t *testing.T) {
	t.Parallel()
	server, instanceID := getTimeseriesTestServer(t)

	mx := "max"
	response, err := server.ColumnTimeSeries(testCtx(), &runtimev1.ColumnTimeSeriesRequest{
		InstanceId: instanceID,
		TableName:  "timeseries",
		Measures: []*runtimev1.ColumnTimeSeriesRequest_BasicMeasure{
			{
				Expression: "max(clicks)",
				SqlName:    mx,
			},
		},
		TimestampColumnName: "time",
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Start:    parseTimeToProtoTimeStamps(t, "2019-01-01T00:00:00Z"),
			End:      parseTimeToProtoTimeStamps(t, "2019-12-01T00:00:00Z"),
			Interval: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
		},
	})

	require.NoError(t, err)
	results := response.GetRollup().Results
	require.Equal(t, 1, len(results))
	require.Equal(t, 1.0, results[0].Records.Fields["max"].GetNumberValue())
}

func TestServer_Timeseries_TimeEnd_exclusive(t *testing.T) {
	t.Parallel()
	server, instanceID := getTimeseriesTestServer(t)

	response, err := server.ColumnTimeSeries(testCtx(), &runtimev1.ColumnTimeSeriesRequest{
		InstanceId: instanceID,
		TableName:  "timeseries",
		Measures: []*runtimev1.ColumnTimeSeriesRequest_BasicMeasure{
			{
				Expression: "max(clicks)",
				SqlName:    "max",
			},
		},
		TimestampColumnName: "time",
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Start:    parseTimeToProtoTimeStamps(t, "2019-01-01T00:00:00Z"),
			End:      parseTimeToProtoTimeStamps(t, "2019-01-02T00:00:00Z"),
			Interval: runtimev1.TimeGrain_TIME_GRAIN_DAY,
		},
	})

	require.NoError(t, err)
	results := response.GetRollup().Results
	require.Equal(t, 1, len(results))
	require.Equal(t, 1.0, results[0].Records.Fields["max"].GetNumberValue())
}

func TestServer_Timeseries_No_Measures(t *testing.T) {
	t.Parallel()
	server, instanceID := getTimeseriesTestServer(t)

	response, err := server.ColumnTimeSeries(testCtx(), &runtimev1.ColumnTimeSeriesRequest{
		InstanceId:          instanceID,
		TableName:           "timeseries",
		Measures:            []*runtimev1.ColumnTimeSeriesRequest_BasicMeasure{},
		TimestampColumnName: "time",
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Start:    parseTimeToProtoTimeStamps(t, "2019-01-01T00:00:00Z"),
			End:      parseTimeToProtoTimeStamps(t, "2019-12-01T00:00:00Z"),
			Interval: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
		},
	})

	require.NoError(t, err)
	results := response.GetRollup().Results
	require.Equal(t, 1, len(results))
	require.Equal(t, 2.0, results[0].Records.Fields["count"].GetNumberValue())
}

func TestServer_Timeseries_Nil_Measures(t *testing.T) {
	t.Parallel()
	server, instanceID := getTimeseriesTestServer(t)

	response, err := server.ColumnTimeSeries(testCtx(), &runtimev1.ColumnTimeSeriesRequest{
		InstanceId:          instanceID,
		TableName:           "timeseries",
		TimestampColumnName: "time",
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Start:    parseTimeToProtoTimeStamps(t, "2019-01-01T00:00:00Z"),
			End:      parseTimeToProtoTimeStamps(t, "2019-12-01T00:00:00Z"),
			Interval: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
		},
	})

	require.NoError(t, err)
	results := response.GetRollup().Results
	require.Equal(t, 1, len(results))
	require.Equal(t, 2.0, results[0].Records.Fields["count"].GetNumberValue())
}

func TestServer_Timeseries_no_measures(t *testing.T) {
	t.Parallel()
	server, instanceID := getTimeseriesTestServer(t)

	response, err := server.ColumnTimeSeries(testCtx(), &runtimev1.ColumnTimeSeriesRequest{
		InstanceId:          instanceID,
		TableName:           "timeseries",
		TimestampColumnName: "time",
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Start:    parseTimeToProtoTimeStamps(t, "2019-01-01T00:00:00Z"),
			End:      parseTimeToProtoTimeStamps(t, "2019-01-03T00:00:00Z"),
			Interval: runtimev1.TimeGrain_TIME_GRAIN_DAY,
		},
	})

	require.NoError(t, err)
	results := response.GetRollup().Results
	require.Equal(t, 2, len(results))
	require.Equal(t, 1.0, results[0].Records.Fields["count"].GetNumberValue())
}

func TestServer_Timeseries_Spark(t *testing.T) {
	t.Parallel()
	server, instanceID := getSparkTimeseriesTestServer(t)

	response, err := server.ColumnTimeSeries(testCtx(), &runtimev1.ColumnTimeSeriesRequest{
		InstanceId: instanceID,
		TableName:  "timeseries",
		Measures: []*runtimev1.ColumnTimeSeriesRequest_BasicMeasure{
			{
				Expression: "count(*)",
				SqlName:    "count",
			},
		},
		TimestampColumnName: "time",
		Pixels:              2,
	})

	require.NoError(t, err)
	for i, v := range response.GetRollup().Results {
		fmt.Printf("i: %d, ts: %v\n", i, v.Ts.AsTime())
	}
	results := response.GetRollup().Results
	require.Equal(t, 9, len(results))
	require.Equal(t, 12, len(response.Rollup.Spark))
}

func TestServer_Timeseries_Spark_no_count(t *testing.T) {
	t.Parallel()
	server, instanceID := getSparkTimeseriesTestServer(t)

	response, err := server.ColumnTimeSeries(testCtx(), &runtimev1.ColumnTimeSeriesRequest{
		InstanceId: instanceID,
		TableName:  "timeseries",
		Measures: []*runtimev1.ColumnTimeSeriesRequest_BasicMeasure{
			{
				Expression: "sum(clicks)",
				SqlName:    "clicks_sum",
			},
		},
		TimestampColumnName: "time",
		Pixels:              2,
	})

	require.NoError(t, err)
	results := response.GetRollup().Results
	require.Equal(t, 9, len(results))
	require.Equal(t, 12, len(response.Rollup.Spark))
}

func getTimeseriesTestServerWithDSTForward(t *testing.T) (*server.Server, string) {
	rt, instanceID := testruntime.NewInstanceWithModel(t, "timeseries", `
		SELECT 1.0 AS clicks, TIMESTAMP '2023-03-25 23:00:00' AS time, 'android' AS device
		UNION ALL
		SELECT 1.0 AS clicks, TIMESTAMP '2023-03-26 00:00:00' AS time, 'iphone' AS device
		UNION ALL
		SELECT 1.0 AS clicks, TIMESTAMP '2023-03-26 01:00:00' AS time, 'iphone' AS device
		UNION ALL
		SELECT 1.0 AS clicks, TIMESTAMP '2023-03-26 02:00:00' AS time, 'iphone' AS device
		UNION ALL
		SELECT 1.0 AS clicks, TIMESTAMP '2023-03-26 03:00:00' AS time, 'iphone' AS device
		UNION ALL
		SELECT 1.0 AS clicks, TIMESTAMP '2023-03-26 04:00:00' AS time, 'iphone' AS device

	`)

	server, err := server.NewServer(context.Background(), &server.Options{}, rt, nil, ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	return server, instanceID
}

func getTimeseriesTestServerWithDSTBackward(t *testing.T) (*server.Server, string) {
	rt, instanceID := testruntime.NewInstanceWithModel(t, "timeseries", `
		SELECT 1.0 AS clicks, TIMESTAMP '2023-10-28 23:00:00' AS time, 'iphone' AS device
		UNION ALL
		SELECT 1.0 AS clicks, TIMESTAMP '2023-10-29 00:00:00' AS time, 'iphone' AS device
		UNION ALL
		SELECT 1.0 AS clicks, TIMESTAMP '2023-10-29 01:00:00' AS time, 'iphone' AS device
		UNION ALL
		SELECT 1.0 AS clicks, TIMESTAMP '2023-10-29 02:00:00' AS time, 'iphone' AS device
		UNION ALL
		SELECT 1.0 AS clicks, TIMESTAMP '2023-10-29 03:00:00' AS time, 'iphone' AS device
		UNION ALL
		SELECT 1.0 AS clicks, TIMESTAMP '2023-10-29 04:00:00' AS time, 'iphone' AS device
	`)

	server, err := server.NewServer(context.Background(), &server.Options{}, rt, nil, ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	return server, instanceID
}

func getTimeseriesTestServerWithKathmandu(t *testing.T) (*server.Server, string) {
	rt, instanceID := testruntime.NewInstanceWithModel(t, "timeseries", `
		SELECT 1.0 AS clicks, TIMESTAMP '2023-10-29 00:15:00' AS time, 'iphone' AS device
		UNION ALL
		SELECT 1.0 AS clicks, TIMESTAMP '2023-10-29 01:15:00' AS time, 'iphone' AS device
		UNION ALL
		SELECT 1.0 AS clicks, TIMESTAMP '2023-10-29 02:15:00' AS time, 'iphone' AS device
	`)

	server, err := server.NewServer(context.Background(), &server.Options{}, rt, nil, ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	return server, instanceID
}

func getTimeseriesTestServer(t *testing.T) (*server.Server, string) {
	rt, instanceID := testruntime.NewInstanceWithModel(t, "timeseries", `
		SELECT 1.0 AS clicks, 3 as imps, TIMESTAMP '2019-01-01 00:00:00' AS time, DATE '2019-01-01' as day, 'android' AS device, 'Google' AS publisher, 'google.com' AS domain, 25 as latitude, 'Canada' as country
		UNION ALL
		SELECT 1.0 AS clicks, 5 as imps, TIMESTAMP '2019-01-02 00:00:00' AS time, DATE '2019-01-02' as day, 'iphone' AS device, null AS publisher, 'msn.com' AS domain, NULL as latitude, NULL as country
	`)

	server, err := server.NewServer(context.Background(), &server.Options{}, rt, nil, ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	return server, instanceID
}

func getTimeseriesTestServerWithEmptyModel(t *testing.T) (*server.Server, string) {
	rt, instanceID := testruntime.NewInstanceWithModel(t, "timeseries", `
		SELECT 1.0 AS clicks, TIMESTAMP '2019-01-01 00:00:00' AS time, 'android' AS device, 'Google' AS publisher, 'google.com' AS domain where 1<>1
	`)

	server, err := server.NewServer(context.Background(), &server.Options{}, rt, nil, ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	return server, instanceID
}

func getSparkTimeseriesTestServer(t *testing.T) (*server.Server, string) {
	rt, instanceID := testruntime.NewInstanceWithModel(t, "timeseries", `
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

	server, err := server.NewServer(context.Background(), &server.Options{}, rt, nil, ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	return server, instanceID
}

func TestServer_EstimateRollupInterval_timestamp(t *testing.T) {
	t.Parallel()
	server, instanceID := getTimeseriesTestServer(t)

	r, err := server.ColumnRollupInterval(testCtx(), &runtimev1.ColumnRollupIntervalRequest{
		InstanceId: instanceID,
		TableName:  "timeseries",
		ColumnName: "time",
		Priority:   1,
	})
	require.NoError(t, err)
	require.Equal(t, parseTime(t, "2019-01-01T00:00:00.000Z"), r.Start.AsTime())
	require.Equal(t, parseTime(t, "2019-01-02T00:00:00.000Z"), r.End.AsTime())
	require.Equal(t, runtimev1.TimeGrain_TIME_GRAIN_HOUR, r.Interval)
}

func TestServer_EstimateRollupInterval_date(t *testing.T) {
	t.Parallel()
	server, instanceID := getTimeseriesTestServer(t)

	r, err := server.ColumnRollupInterval(testCtx(), &runtimev1.ColumnRollupIntervalRequest{
		InstanceId: instanceID,
		TableName:  "timeseries",
		ColumnName: "day",
		Priority:   1,
	})
	require.NoError(t, err)
	require.Equal(t, parseTime(t, "2019-01-01T00:00:00.000Z"), r.Start.AsTime())
	require.Equal(t, parseTime(t, "2019-01-02T00:00:00.000Z"), r.End.AsTime())
	require.Equal(t, runtimev1.TimeGrain_TIME_GRAIN_HOUR, r.Interval)
}

/*
select

	time_bucket(interval '1 hour', range, 'Europe/Copenhagen') bucket_in_utc,
	bucket_in_utc at time zone 'Europe/Copenhagen' copenhagen_bucket,
	range event_time

from

	range(timestamptz '2023-03-26 00:00:00', timestamptz '2023-03-26 05:00:00', interval '1 hour');

┌──────────────────────────┬─────────────────────┬──────────────────────────┐
│      bucket_in_utc       │  copenhagen_bucket  │        event_time        │
│ timestamp with time zone │      timestamp      │ timestamp with time zone │
├──────────────────────────┼─────────────────────┼──────────────────────────┤
│ 2023-03-26 00:00:00+00   │ 2023-03-26 01:00:00 │ 2023-03-26 00:00:00+00   │
│ 2023-03-26 01:00:00+00   │ 2023-03-26 03:00:00 │ 2023-03-26 01:00:00+00   │
│ 2023-03-26 02:00:00+00   │ 2023-03-26 04:00:00 │ 2023-03-26 02:00:00+00   │
│ 2023-03-26 03:00:00+00   │ 2023-03-26 05:00:00 │ 2023-03-26 03:00:00+00   │
│ 2023-03-26 04:00:00+00   │ 2023-03-26 06:00:00 │ 2023-03-26 04:00:00+00   │
└──────────────────────────┴─────────────────────┴──────────────────────────┘
*/
func TestServer_Timeseries_timezone_dst_forward(t *testing.T) {
	t.Parallel()
	server, instanceID := getTimeseriesTestServerWithDSTForward(t)

	resp, err := server.Query(testCtx(), &runtimev1.QueryRequest{
		InstanceId: instanceID,
		Sql:        "select current_setting('TimeZone') as value",
	})
	require.NoError(t, err)
	require.Equal(t, "Etc/UTC", resp.Data[0].Fields["value"].GetStringValue())

	response, err := server.ColumnTimeSeries(testCtx(), &runtimev1.ColumnTimeSeriesRequest{
		InstanceId:          instanceID,
		TableName:           "timeseries",
		TimestampColumnName: "time",
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Start:    parseTimeToProtoTimeStamps(t, "2023-03-26T00:00:00Z"),
			End:      parseTimeToProtoTimeStamps(t, "2023-03-26T04:00:00Z"),
			Interval: runtimev1.TimeGrain_TIME_GRAIN_HOUR,
		},
		TimeZone: "Europe/Copenhagen",
	})

	require.NoError(t, err)
	results := response.GetRollup().Results
	require.Equal(t, 4, len(results))
	require.Equal(t, "2023-03-26 00:00:00", results[0].Ts.AsTime().Format(time.DateTime))
	require.Equal(t, 1.0, results[0].Records.Fields["count"].GetNumberValue())
	require.Equal(t, "2023-03-26 01:00:00", results[1].Ts.AsTime().Format(time.DateTime))
	require.Equal(t, 1.0, results[1].Records.Fields["count"].GetNumberValue())
	require.Equal(t, "2023-03-26 02:00:00", results[2].Ts.AsTime().Format(time.DateTime))
	require.Equal(t, 1.0, results[2].Records.Fields["count"].GetNumberValue())
	require.Equal(t, "2023-03-26 03:00:00", results[3].Ts.AsTime().Format(time.DateTime))
	require.Equal(t, 1.0, results[3].Records.Fields["count"].GetNumberValue())

}

/*
	select
		time_bucket(interval '1 hour', range, 'Europe/Copenhagen') bucket_in_utc,
		bucket_in_utc at time zone 'Europe/Copenhagen' copenhagen_bucket,
		range event_time
	from
		range(timestamptz '2023-10-29 00:00:00', timestamptz '2023-10-29 03:00:00', interval '1 hour');

	┌──────────────────────────┬─────────────────────┬──────────────────────────┐
	│      bucket_in_utc       │  copenhagen_bucket  │        event_time        │
	│ timestamp with time zone │      timestamp      │ timestamp with time zone │
	├──────────────────────────┼─────────────────────┼──────────────────────────┤
	│ 2023-10-29 00:00:00+00   │ 2023-10-29 02:00:00 │ 2023-10-29 00:00:00+00   │
	│ 2023-10-29 01:00:00+00   │ 2023-10-29 02:00:00 │ 2023-10-29 01:00:00+00   │
	│ 2023-10-29 02:00:00+00   │ 2023-10-29 03:00:00 │ 2023-10-29 02:00:00+00   │
	└──────────────────────────┴─────────────────────┴──────────────────────────┘
*/

func TestServer_Timeseries_timezone_dst_backward(t *testing.T) {
	t.Parallel()
	server, instanceID := getTimeseriesTestServerWithDSTBackward(t)

	response, err := server.ColumnTimeSeries(testCtx(), &runtimev1.ColumnTimeSeriesRequest{
		InstanceId:          instanceID,
		TableName:           "timeseries",
		TimestampColumnName: "time",
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Start:    parseTimeToProtoTimeStamps(t, "2023-10-28T23:00:00Z"),
			End:      parseTimeToProtoTimeStamps(t, "2023-10-29T04:00:00Z"),
			Interval: runtimev1.TimeGrain_TIME_GRAIN_HOUR,
		},
		TimeZone: "Europe/Copenhagen",
	})

	require.NoError(t, err)
	results := response.GetRollup().Results
	require.Equal(t, 4, len(results))

	require.Equal(t, "2023-10-28 23:00:00", results[0].Ts.AsTime().Format(time.DateTime))
	require.Equal(t, 1.0, results[0].Records.Fields["count"].GetNumberValue())
	require.Equal(t, "2023-10-29 01:00:00", results[1].Ts.AsTime().Format(time.DateTime))
	require.Equal(t, 2.0, results[1].Records.Fields["count"].GetNumberValue())
	require.Equal(t, "2023-10-29 02:00:00", results[2].Ts.AsTime().Format(time.DateTime))
	require.Equal(t, 1.0, results[2].Records.Fields["count"].GetNumberValue())
	require.Equal(t, "2023-10-29 03:00:00", results[3].Ts.AsTime().Format(time.DateTime))
	require.Equal(t, 1.0, results[3].Records.Fields["count"].GetNumberValue())

}

func TestServer_Timeseries_timezone_Kathmandu_with_hour(t *testing.T) {
	t.Parallel()
	server, instanceID := getTimeseriesTestServerWithKathmandu(t)

	response, err := server.ColumnTimeSeries(testCtx(), &runtimev1.ColumnTimeSeriesRequest{
		InstanceId:          instanceID,
		TableName:           "timeseries",
		TimestampColumnName: "time",
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Start:    parseTimeToProtoTimeStamps(t, "2023-10-29T00:15:00Z"),
			End:      parseTimeToProtoTimeStamps(t, "2023-10-29T03:15:00Z"),
			Interval: runtimev1.TimeGrain_TIME_GRAIN_HOUR,
		},
		TimeZone: "Asia/Kathmandu",
	})

	require.NoError(t, err)
	results := response.GetRollup().Results
	require.Equal(t, 3, len(results))

	require.Equal(t, "2023-10-29 00:15:00", results[0].Ts.AsTime().Format(time.DateTime))
	require.Equal(t, 1.0, results[0].Records.Fields["count"].GetNumberValue())
	require.Equal(t, "2023-10-29 01:15:00", results[1].Ts.AsTime().Format(time.DateTime))
	require.Equal(t, 1.0, results[1].Records.Fields["count"].GetNumberValue())
	require.Equal(t, "2023-10-29 02:15:00", results[2].Ts.AsTime().Format(time.DateTime))
	require.Equal(t, 1.0, results[2].Records.Fields["count"].GetNumberValue())
}

func getTimeseriesTestServerWithWeekGrain(t *testing.T) (*server.Server, string) {
	selects := make([]string, 0, 12)
	tm, err := time.Parse(time.DateOnly, "2022-09-01")
	require.NoError(t, err)
	for i := 0; i < 52; i++ {
		selects = append(selects, `SELECT 1.0 AS clicks, TIMESTAMP '`+tm.Format(time.DateTime)+`' AS time, 'iphone' AS device`)
		tm = tm.AddDate(0, 0, 7)
	}

	sql := strings.Join(selects, " UNION ALL ")
	rt, instanceID := testruntime.NewInstanceWithModel(t, "timeseries", sql)

	server, err := server.NewServer(context.Background(), &server.Options{}, rt, nil, ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	return server, instanceID
}

func getTimeseriesTestServerWithMonthGrain(t *testing.T) (*server.Server, string) {
	selects := make([]string, 0, 120)
	tm, err := time.Parse(time.DateOnly, "2013-09-01")
	require.NoError(t, err)
	for i := 0; i < 120; i++ {
		selects = append(selects, `SELECT 1.0 AS clicks, TIMESTAMP '`+tm.Format(time.DateTime)+`' AS time, 'iphone' AS device`)
		tm = tm.AddDate(0, 1, 0)
	}

	sql := strings.Join(selects, " UNION ALL ")
	rt, instanceID := testruntime.NewInstanceWithModel(t, "timeseries", sql)
	server, err := server.NewServer(context.Background(), &server.Options{}, rt, nil, ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	return server, instanceID
}

func TestServer_Timeseries_Kathmandu_with_week(t *testing.T) {
	t.Parallel()
	server, instanceID := getTimeseriesTestServerWithWeekGrain(t)

	response, err := server.ColumnTimeSeries(testCtx(), &runtimev1.ColumnTimeSeriesRequest{
		InstanceId:          instanceID,
		TableName:           "timeseries",
		TimestampColumnName: "time",
		Measures: []*runtimev1.ColumnTimeSeriesRequest_BasicMeasure{
			{
				Expression: "max(clicks)",
				SqlName:    "max_clicks",
			},
		},
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Start:    parseTimeToProtoTimeStamps(t, "2022-09-01T00:00:00Z"),
			End:      parseTimeToProtoTimeStamps(t, "2023-09-01T00:00:00Z"),
			Interval: runtimev1.TimeGrain_TIME_GRAIN_WEEK,
		},
		TimeZone: "Asia/Kathmandu",
	})

	require.NoError(t, err)
	results := response.GetRollup().Results
	require.Equal(t, 52, len(results))

	var previousTime *time.Time
	for i := 0; i < 52; i++ {
		value := results[i].Records.Fields["max_clicks"]
		n, ok := value.GetKind().(*structpb.Value_NumberValue)
		tm := results[i].Ts.AsTime()
		require.True(t, ok, "Element %d %s", i, tm)
		if previousTime != nil {
			require.Less(t, 0, tm.Compare(*previousTime))
		}
		previousTime = &tm
		require.Equal(t, 1.0, n.NumberValue, "Element %d %s", i, results[i].Ts.AsTime())
	}
}

func TestServer_Timeseries_Copenhagen_with_week(t *testing.T) {
	t.Parallel()
	server, instanceID := getTimeseriesTestServerWithWeekGrain(t)

	response, err := server.ColumnTimeSeries(testCtx(), &runtimev1.ColumnTimeSeriesRequest{
		InstanceId:          instanceID,
		TableName:           "timeseries",
		TimestampColumnName: "time",
		Measures: []*runtimev1.ColumnTimeSeriesRequest_BasicMeasure{
			{
				Expression: "max(clicks)",
				SqlName:    "max_clicks",
			},
		},
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Start:    parseTimeToProtoTimeStamps(t, "2022-09-01T00:00:00Z"),
			End:      parseTimeToProtoTimeStamps(t, "2023-09-01T00:00:00Z"),
			Interval: runtimev1.TimeGrain_TIME_GRAIN_WEEK,
		},
		TimeZone: "Europe/Copenhagen",
	})

	require.NoError(t, err)
	results := response.GetRollup().Results
	require.Equal(t, 52, len(results))

	var previousTime *time.Time

	for i := 0; i < len(results); i++ {
		value := results[i].Records.Fields["max_clicks"]
		n, ok := value.GetKind().(*structpb.Value_NumberValue)
		tm := results[i].Ts.AsTime()
		require.True(t, ok, "Element %d %s", i, tm)
		if previousTime != nil {
			require.Less(t, 0, tm.Compare(*previousTime))
		}
		previousTime = &tm
		require.Equal(t, 1.0, n.NumberValue, "Element %d %s", i, results[i].Ts.AsTime())
	}
}

func TestServer_Timeseries_Copenhagen_with_month(t *testing.T) {
	t.Parallel()
	server, instanceID := getTimeseriesTestServerWithMonthGrain(t)

	response, err := server.ColumnTimeSeries(testCtx(), &runtimev1.ColumnTimeSeriesRequest{
		InstanceId:          instanceID,
		TableName:           "timeseries",
		TimestampColumnName: "time",
		Measures: []*runtimev1.ColumnTimeSeriesRequest_BasicMeasure{
			{
				Expression: "max(clicks)",
				SqlName:    "max_clicks",
			},
		},
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Start:    parseTimeToProtoTimeStamps(t, "2013-09-01T00:00:00Z"),
			End:      parseTimeToProtoTimeStamps(t, "2023-09-01T00:00:00Z"),
			Interval: runtimev1.TimeGrain_TIME_GRAIN_MONTH,
		},
		TimeZone: "Europe/Copenhagen",
	})

	require.NoError(t, err)
	results := response.GetRollup().Results
	require.Equal(t, 120, len(results))

	var previousTime *time.Time

	for i := 0; i < len(results); i++ {
		value := results[i].Records.Fields["max_clicks"]
		n, ok := value.GetKind().(*structpb.Value_NumberValue)
		tm := results[i].Ts.AsTime()
		require.True(t, ok, "Element %d %s", i, tm)
		fmt.Printf("%s %.1f\n", tm.Format(time.DateOnly), value.GetNumberValue())
		if previousTime != nil {
			require.Less(t, 0, tm.Compare(*previousTime))
		}
		previousTime = &tm
		require.Equal(t, 1.0, n.NumberValue, "Element %d %s", i, results[i].Ts.AsTime())
	}
}
