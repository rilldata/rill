package server

import (
	"fmt"
	"testing"
	"time"

	"github.com/marcboeker/go-duckdb"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	structpb "google.golang.org/protobuf/types/known/structpb"
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

func TestServer_RangeSanity(t *testing.T) {
	t.Parallel()
	server, instanceID := getTimeseriesTestServer(t)

	olap, err := server.runtime.OLAP(testCtx(), instanceID)
	require.NoError(t, err)

	result, err := olap.Execute(testCtx(), &drivers.Statement{
		Query: "select min(time) min, max(time) max, max(time)-min(time) as r from timeseries",
	})
	require.NoError(t, err)

	var min, max time.Time
	var r duckdb.Interval
	result.Next()
	err = result.Scan(&min, &max, &r)
	require.NoError(t, err)
	require.Equal(t, time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC), min)
	require.Equal(t, int32(1), r.Days)
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

func getTimeseriesTestServer(t *testing.T) (*Server, string) {
	rt, instanceID := testruntime.NewInstanceWithModel(t, "timeseries", `
		SELECT 1.0 AS clicks, 3 as imps, TIMESTAMP '2019-01-01 00:00:00' AS time, DATE '2019-01-01' as day, 'android' AS device, 'Google' AS publisher, 'google.com' AS domain, 25 as latitude, 'Canada' as country
		UNION ALL
		SELECT 1.0 AS clicks, 5 as imps, TIMESTAMP '2019-01-02 00:00:00' AS time, DATE '2019-01-02' as day, 'iphone' AS device, null AS publisher, 'msn.com' AS domain, NULL as latitude, NULL as country
	`)

	server, err := NewServer(&Options{}, rt, nil)
	require.NoError(t, err)

	return server, instanceID
}

func getTimeseriesTestServerWithEmptyModel(t *testing.T) (*Server, string) {
	rt, instanceID := testruntime.NewInstanceWithModel(t, "timeseries", `
		SELECT 1.0 AS clicks, TIMESTAMP '2019-01-01 00:00:00' AS time, 'android' AS device, 'Google' AS publisher, 'google.com' AS domain where 1<>1
	`)

	server, err := NewServer(&Options{}, rt, nil)
	require.NoError(t, err)

	return server, instanceID
}

func getSparkTimeseriesTestServer(t *testing.T) (*Server, string) {
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

	server, err := NewServer(&Options{}, rt, nil)
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
