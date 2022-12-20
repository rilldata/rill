package server

import (
	"context"
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
	server, instanceID := getTimeseriesTestServerWithEmptyModel(t)

	response, err := server.GenerateTimeSeries(context.Background(), &runtimev1.GenerateTimeSeriesRequest{
		InstanceId: instanceID,
		TableName:  "timeseries",
		Measures: []*runtimev1.GenerateTimeSeriesRequest_BasicMeasure{
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

func TestServer_Timeseries(t *testing.T) {
	server, instanceID := getTimeseriesTestServer(t)

	response, err := server.GenerateTimeSeries(context.Background(), &runtimev1.GenerateTimeSeriesRequest{
		InstanceId: instanceID,
		TableName:  "timeseries",
		Measures: []*runtimev1.GenerateTimeSeriesRequest_BasicMeasure{
			{
				Expression: "max(clicks)",
				SqlName:    "max",
			},
		},
		TimestampColumnName: "time",
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Start:    parseTime(t, "2019-01-01T00:00:00Z"),
			End:      parseTime(t, "2019-12-01T00:00:00Z"),
			Interval: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
		},
		Filters: &runtimev1.MetricsViewRequestFilter{
			Include: []*runtimev1.MetricsViewDimensionValue{
				{
					Name: "device",
					In:   []*structpb.Value{structpb.NewStringValue("android"), structpb.NewStringValue("iphone")},
				},
			},
		},
	})

	require.NoError(t, err)
	results := response.GetRollup().Results
	require.Equal(t, 1, len(results))
	require.Equal(t, 1.0, results[0].Records["max"])
}

func TestServer_Timeseries_numeric_dim(t *testing.T) {
	server, instanceID := getTimeseriesTestServer(t)

	response, err := server.GenerateTimeSeries(context.Background(), &runtimev1.GenerateTimeSeriesRequest{
		InstanceId: instanceID,
		TableName:  "timeseries",
		Measures: []*runtimev1.GenerateTimeSeriesRequest_BasicMeasure{
			{
				Expression: "count(*)",
				SqlName:    "count",
			},
		},
		TimestampColumnName: "time",
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Interval: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
		},
		Filters: &runtimev1.MetricsViewRequestFilter{
			Include: []*runtimev1.MetricsViewDimensionValue{
				{
					Name: "latitude",
					In:   []*structpb.Value{structpb.NewNumberValue(25)},
				},
			},
		},
	})

	require.NoError(t, err)
	results := response.GetRollup().Results
	require.Equal(t, 1, len(results))
	require.Equal(t, 1.0, results[0].Records["count"])
}

func TestServer_Timeseries_numeric_dim_2values(t *testing.T) {
	server, instanceID := getTimeseriesTestServer(t)

	response, err := server.GenerateTimeSeries(context.Background(), &runtimev1.GenerateTimeSeriesRequest{
		InstanceId: instanceID,
		TableName:  "timeseries",
		Measures: []*runtimev1.GenerateTimeSeriesRequest_BasicMeasure{
			{
				Expression: "count(*)",
				SqlName:    "count",
			},
		},
		TimestampColumnName: "time",
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Interval: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
		},
		Filters: &runtimev1.MetricsViewRequestFilter{
			Include: []*runtimev1.MetricsViewDimensionValue{
				{
					Name: "latitude",
					In:   []*structpb.Value{structpb.NewNumberValue(25), structpb.NewNumberValue(35)},
				},
			},
		},
	})

	require.NoError(t, err)
	results := response.GetRollup().Results
	require.Equal(t, 1, len(results))
	require.Equal(t, 1.0, results[0].Records["count"])
}

func TestServer_Timeseries_numeric_dim_and_null(t *testing.T) {
	server, instanceID := getTimeseriesTestServer(t)

	response, err := server.GenerateTimeSeries(context.Background(), &runtimev1.GenerateTimeSeriesRequest{
		InstanceId: instanceID,
		TableName:  "timeseries",
		Measures: []*runtimev1.GenerateTimeSeriesRequest_BasicMeasure{
			{
				Expression: "count(*)",
				SqlName:    "count",
			},
		},
		TimestampColumnName: "time",
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Interval: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
		},
		Filters: &runtimev1.MetricsViewRequestFilter{
			Include: []*runtimev1.MetricsViewDimensionValue{
				{
					Name: "latitude",
					In:   []*structpb.Value{structpb.NewNumberValue(25), structpb.NewNullValue()},
				},
			},
		},
	})

	require.NoError(t, err)
	results := response.GetRollup().Results
	require.Equal(t, 1, len(results))
	require.Equal(t, 2.0, results[0].Records["count"])
}

func TestServer_Timeseries_Empty_TimeRange(t *testing.T) {
	server, instanceID := getTimeseriesTestServer(t)

	mx := "max"
	response, err := server.GenerateTimeSeries(context.Background(), &runtimev1.GenerateTimeSeriesRequest{
		InstanceId: instanceID,
		TableName:  "timeseries",
		Measures: []*runtimev1.GenerateTimeSeriesRequest_BasicMeasure{
			{
				Expression: "max(clicks)",
				SqlName:    mx,
			},
		},
		TimestampColumnName: "time",
		TimeRange:           new(runtimev1.TimeSeriesTimeRange),
		Filters: &runtimev1.MetricsViewRequestFilter{
			Include: []*runtimev1.MetricsViewDimensionValue{
				{
					Name: "device",
					In:   []*structpb.Value{structpb.NewStringValue("android"), structpb.NewStringValue("iphone")},
				},
			},
		},
	})

	require.NoError(t, err)
	results := response.GetRollup().Results
	require.Equal(t, 25, len(results))
	require.Equal(t, 1.0, results[0].Records["max"])
}

func TestServer_Timeseries_Empty_Filter(t *testing.T) {
	server, instanceID := getTimeseriesTestServer(t)

	mx := "max"
	response, err := server.GenerateTimeSeries(context.Background(), &runtimev1.GenerateTimeSeriesRequest{
		InstanceId: instanceID,
		TableName:  "timeseries",
		Measures: []*runtimev1.GenerateTimeSeriesRequest_BasicMeasure{
			{
				Expression: "max(clicks)",
				SqlName:    mx,
			},
		},
		TimestampColumnName: "time",
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Start:    parseTime(t, "2019-01-01T00:00:00Z"),
			End:      parseTime(t, "2019-12-01T00:00:00Z"),
			Interval: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
		},
		Filters: new(runtimev1.MetricsViewRequestFilter),
	})

	require.NoError(t, err)
	results := response.GetRollup().Results
	require.Equal(t, 1, len(results))
	require.Equal(t, 1.0, results[0].Records["max"])
}

func TestServer_Timeseries_No_Measures(t *testing.T) {
	server, instanceID := getTimeseriesTestServer(t)

	response, err := server.GenerateTimeSeries(context.Background(), &runtimev1.GenerateTimeSeriesRequest{
		InstanceId:          instanceID,
		TableName:           "timeseries",
		Measures:            []*runtimev1.GenerateTimeSeriesRequest_BasicMeasure{},
		TimestampColumnName: "time",
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Start:    parseTime(t, "2019-01-01T00:00:00Z"),
			End:      parseTime(t, "2019-12-01T00:00:00Z"),
			Interval: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
		},
		Filters: new(runtimev1.MetricsViewRequestFilter),
	})

	require.NoError(t, err)
	results := response.GetRollup().Results
	require.Equal(t, 1, len(results))
	require.Equal(t, 2.0, results[0].Records["count"])
}

func TestServer_Timeseries_Nil_Measures(t *testing.T) {
	server, instanceID := getTimeseriesTestServer(t)

	response, err := server.GenerateTimeSeries(context.Background(), &runtimev1.GenerateTimeSeriesRequest{
		InstanceId:          instanceID,
		TableName:           "timeseries",
		TimestampColumnName: "time",
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Start:    parseTime(t, "2019-01-01T00:00:00Z"),
			End:      parseTime(t, "2019-12-01T00:00:00Z"),
			Interval: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
		},
		Filters: new(runtimev1.MetricsViewRequestFilter),
	})

	require.NoError(t, err)
	results := response.GetRollup().Results
	require.Equal(t, 1, len(results))
	require.Equal(t, 2.0, results[0].Records["count"])
}

func TestServer_Timeseries_2measures(t *testing.T) {
	server, instanceID := getTimeseriesTestServer(t)

	mx := "max"
	sm := "sum"
	response, err := server.GenerateTimeSeries(context.Background(), &runtimev1.GenerateTimeSeriesRequest{
		InstanceId: instanceID,
		TableName:  "timeseries",
		Measures: []*runtimev1.GenerateTimeSeriesRequest_BasicMeasure{
			{
				Expression: "max(clicks)",
				SqlName:    mx,
			},
			{
				Expression: "sum(clicks)",
				SqlName:    sm,
			},
		},
		TimestampColumnName: "time",
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Start:    parseTime(t, "2019-01-01T00:00:00Z"),
			End:      parseTime(t, "2019-12-01T00:00:00Z"),
			Interval: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
		},
		Filters: &runtimev1.MetricsViewRequestFilter{
			Include: []*runtimev1.MetricsViewDimensionValue{
				{
					Name: "device",
					In:   []*structpb.Value{structpb.NewStringValue("android"), structpb.NewStringValue("iphone")},
				},
			},
		},
	})

	require.NoError(t, err)
	results := response.GetRollup().Results
	require.Equal(t, 1, len(results))
	require.Equal(t, 1.0, results[0].Records["max"])
	require.Equal(t, 2.0, results[0].Records["sum"])
}

func TestServer_Timeseries_1dim(t *testing.T) {
	server, instanceID := getTimeseriesTestServer(t)

	sm := "sum"
	response, err := server.GenerateTimeSeries(context.Background(), &runtimev1.GenerateTimeSeriesRequest{
		InstanceId: instanceID,
		TableName:  "timeseries",
		Measures: []*runtimev1.GenerateTimeSeriesRequest_BasicMeasure{
			{
				Expression: "sum(clicks)",
				SqlName:    sm,
			},
		},
		TimestampColumnName: "time",
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Start:    parseTime(t, "2019-01-01T00:00:00Z"),
			End:      parseTime(t, "2019-12-01T00:00:00Z"),
			Interval: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
		},
		Filters: &runtimev1.MetricsViewRequestFilter{
			Include: []*runtimev1.MetricsViewDimensionValue{
				{
					Name: "device",
					In:   []*structpb.Value{structpb.NewStringValue("android")},
				},
			},
		},
	})

	require.NoError(t, err)
	results := response.GetRollup().Results
	require.Equal(t, 1, len(results))
	require.Equal(t, 1.0, results[0].Records["sum"])
}

func TestServer_Timeseries_1dim_null(t *testing.T) {
	server, instanceID := getTimeseriesTestServer(t)

	response, err := server.GenerateTimeSeries(context.Background(), &runtimev1.GenerateTimeSeriesRequest{
		InstanceId: instanceID,
		TableName:  "timeseries",
		Measures: []*runtimev1.GenerateTimeSeriesRequest_BasicMeasure{
			{
				Expression: "sum(clicks)",
				SqlName:    "sum",
			},
		},
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Interval: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
		},
		TimestampColumnName: "time",
		Filters: &runtimev1.MetricsViewRequestFilter{
			Include: []*runtimev1.MetricsViewDimensionValue{
				{
					Name: "publisher",
					In:   []*structpb.Value{structpb.NewNullValue()},
				},
			},
		},
	})

	require.NoError(t, err)
	results := response.GetRollup().Results
	require.Equal(t, 1, len(results))
	require.Equal(t, 1.0, results[0].Records["sum"])
}

func TestServer_Timeseries_1dim_null_and_in(t *testing.T) {
	server, instanceID := getTimeseriesTestServer(t)

	response, err := server.GenerateTimeSeries(context.Background(), &runtimev1.GenerateTimeSeriesRequest{
		InstanceId: instanceID,
		TableName:  "timeseries",
		Measures: []*runtimev1.GenerateTimeSeriesRequest_BasicMeasure{
			{
				Expression: "sum(clicks)",
				SqlName:    "sum",
			},
		},
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Interval: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
		},
		TimestampColumnName: "time",
		Filters: &runtimev1.MetricsViewRequestFilter{
			Include: []*runtimev1.MetricsViewDimensionValue{
				{
					Name: "publisher",
					In: []*structpb.Value{
						structpb.NewNullValue(),
						structpb.NewStringValue("Google"),
					},
				},
			},
		},
	})

	require.NoError(t, err)
	results := response.GetRollup().Results
	require.Equal(t, 1, len(results))
	require.Equal(t, 2.0, results[0].Records["sum"])
}

func TestServer_Timeseries_1dim_null_and_in_and_like(t *testing.T) {
	server, instanceID := getTimeseriesTestServer(t)

	response, err := server.GenerateTimeSeries(context.Background(), &runtimev1.GenerateTimeSeriesRequest{
		InstanceId: instanceID,
		TableName:  "timeseries",
		Measures: []*runtimev1.GenerateTimeSeriesRequest_BasicMeasure{
			{
				Expression: "sum(clicks)",
				SqlName:    "sum",
			},
		},
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Interval: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
		},
		TimestampColumnName: "time",
		Filters: &runtimev1.MetricsViewRequestFilter{
			Include: []*runtimev1.MetricsViewDimensionValue{
				{
					Name: "publisher",
					In: []*structpb.Value{
						structpb.NewNullValue(),
						structpb.NewStringValue("Google"),
					},
					Like: []*structpb.Value{
						structpb.NewStringValue("Goo%"),
					},
				},
			},
		},
	})

	require.NoError(t, err)
	results := response.GetRollup().Results
	require.Equal(t, 1, len(results))
	require.Equal(t, 2.0, results[0].Records["sum"])
}

func TestServer_Timeseries_1dim_2like(t *testing.T) {
	server, instanceID := getTimeseriesTestServer(t)

	response, err := server.GenerateTimeSeries(context.Background(), &runtimev1.GenerateTimeSeriesRequest{
		InstanceId: instanceID,
		TableName:  "timeseries",
		Measures: []*runtimev1.GenerateTimeSeriesRequest_BasicMeasure{
			{
				Expression: "sum(clicks)",
				SqlName:    "sum",
			},
		},
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Interval: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
		},
		TimestampColumnName: "time",
		Filters: &runtimev1.MetricsViewRequestFilter{
			Include: []*runtimev1.MetricsViewDimensionValue{
				{
					Name: "domain",
					Like: []*structpb.Value{
						structpb.NewStringValue("g%"),
						structpb.NewStringValue("msn%"),
					},
				},
			},
		},
	})

	require.NoError(t, err)
	results := response.GetRollup().Results
	require.Equal(t, 1, len(results))
	require.Equal(t, 2.0, results[0].Records["sum"])
}

func TestServer_Timeseries_2dim_include_and_exclude(t *testing.T) {
	server, instanceID := getTimeseriesTestServer(t)

	response, err := server.GenerateTimeSeries(context.Background(), &runtimev1.GenerateTimeSeriesRequest{
		InstanceId: instanceID,
		TableName:  "timeseries",
		Measures: []*runtimev1.GenerateTimeSeriesRequest_BasicMeasure{
			{
				Expression: "sum(clicks)",
				SqlName:    "sum",
			},
		},
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Interval: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
		},
		TimestampColumnName: "time",
		Filters: &runtimev1.MetricsViewRequestFilter{
			Include: []*runtimev1.MetricsViewDimensionValue{
				{
					Name: "publisher",
					In: []*structpb.Value{
						structpb.NewStringValue("Google"),
					},
				},
				{
					Name: "domain",
					In: []*structpb.Value{
						structpb.NewStringValue("msn.com"),
					},
				},
			},
		},
	})

	require.NoError(t, err)
	results := response.GetRollup().Results
	require.Equal(t, 1, len(results))
	require.Equal(t, 0.0, results[0].Records["sum"])
}

func TestServer_Timeseries_no_measures(t *testing.T) {
	server, instanceID := getTimeseriesTestServer(t)

	response, err := server.GenerateTimeSeries(context.Background(), &runtimev1.GenerateTimeSeriesRequest{
		InstanceId:          instanceID,
		TableName:           "timeseries",
		TimestampColumnName: "time",
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Start:    parseTime(t, "2019-01-01T00:00:00Z"),
			End:      parseTime(t, "2019-01-02T00:00:00Z"),
			Interval: runtimev1.TimeGrain_TIME_GRAIN_DAY,
		},
	})

	require.NoError(t, err)
	results := response.GetRollup().Results
	require.Equal(t, 2, len(results))
	require.Equal(t, 1.0, results[0].Records["count"])
}

func TestServer_Timeseries_1day(t *testing.T) {
	server, instanceID := getTimeseriesTestServer(t)

	mx := "max"
	response, err := server.GenerateTimeSeries(context.Background(), &runtimev1.GenerateTimeSeriesRequest{
		InstanceId: instanceID,
		TableName:  "timeseries",
		Measures: []*runtimev1.GenerateTimeSeriesRequest_BasicMeasure{
			{
				Expression: "max(clicks)",
				SqlName:    mx,
			},
		},
		TimestampColumnName: "time",
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Start:    parseTime(t, "2019-01-01T00:00:00Z"),
			End:      parseTime(t, "2019-01-02T00:00:00Z"),
			Interval: runtimev1.TimeGrain_TIME_GRAIN_DAY,
		},
		Filters: &runtimev1.MetricsViewRequestFilter{
			Include: []*runtimev1.MetricsViewDimensionValue{
				{
					Name: "device",
					In:   []*structpb.Value{structpb.NewStringValue("android"), structpb.NewStringValue("iphone")},
				},
			},
		},
	})

	require.NoError(t, err)
	results := response.GetRollup().Results
	require.Equal(t, 2, len(results))
}

func TestServer_Timeseries_1day_Count(t *testing.T) {
	server, instanceID := getTimeseriesTestServer(t)

	cnt := "count"
	response, err := server.GenerateTimeSeries(context.Background(), &runtimev1.GenerateTimeSeriesRequest{
		InstanceId: instanceID,
		TableName:  "timeseries",
		Measures: []*runtimev1.GenerateTimeSeriesRequest_BasicMeasure{
			{
				Expression: "count(*)",
				SqlName:    cnt,
			},
		},
		TimestampColumnName: "time",
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Start:    parseTime(t, "2019-01-01T00:00:00Z"),
			End:      parseTime(t, "2019-01-02T00:00:00Z"),
			Interval: runtimev1.TimeGrain_TIME_GRAIN_DAY,
		},
		Filters: &runtimev1.MetricsViewRequestFilter{
			Include: []*runtimev1.MetricsViewDimensionValue{
				{
					Name: "device",
					In:   []*structpb.Value{structpb.NewStringValue("android"), structpb.NewStringValue("iphone")},
				},
			},
		},
	})

	require.NoError(t, err)
	results := response.GetRollup().Results
	require.Equal(t, 2, len(results))
	require.Equal(t, 1.0, results[0].Records["count"])
}

func TestServer_RangeSanity(t *testing.T) {
	server, instanceID := getTimeseriesTestServer(t)

	result, err := server.query(context.Background(), instanceID, &drivers.Statement{
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
	server, instanceID := getSparkTimeseriesTestServer(t)

	cnt := "count"
	pxls := int32(2)
	response, err := server.GenerateTimeSeries(context.Background(), &runtimev1.GenerateTimeSeriesRequest{
		InstanceId: instanceID,
		TableName:  "timeseries",
		Measures: []*runtimev1.GenerateTimeSeriesRequest_BasicMeasure{
			{
				Expression: "count(*)",
				SqlName:    cnt,
			},
		},
		TimestampColumnName: "time",
		Pixels:              pxls,
	})

	require.NoError(t, err)
	results := response.GetRollup().Results
	require.Equal(t, 9, len(results))
	require.Equal(t, 12, len(response.Rollup.Spark))
}

func TestServer_Timeseries_Spark_no_count(t *testing.T) {
	server, instanceID := getSparkTimeseriesTestServer(t)

	response, err := server.GenerateTimeSeries(context.Background(), &runtimev1.GenerateTimeSeriesRequest{
		InstanceId: instanceID,
		TableName:  "timeseries",
		Measures: []*runtimev1.GenerateTimeSeriesRequest_BasicMeasure{
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
		SELECT 1.0 AS clicks, TIMESTAMP '2019-01-01 00:00:00' AS time, DATE '2019-01-01' as day, 'android' AS device, 'Google' AS publisher, 'google.com' AS domain, 25 as latitude
		UNION ALL
		SELECT 1.0 AS clicks, TIMESTAMP '2019-01-02 00:00:00' AS time, DATE '2019-01-02' as day, 'iphone' AS device, null AS publisher, 'msn.com' AS domain, NULL as latitude
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
	server, instanceID := getTimeseriesTestServer(t)

	r, err := server.EstimateRollupInterval(context.Background(), &runtimev1.EstimateRollupIntervalRequest{
		InstanceId: instanceID,
		TableName:  "timeseries",
		ColumnName: "time",
		Priority:   1,
	})
	require.NoError(t, err)
	require.Equal(t, parseTime(t, "2019-01-01T00:00:00.000Z"), r.Start)
	require.Equal(t, parseTime(t, "2019-01-02T00:00:00.000Z"), r.End)
	require.Equal(t, runtimev1.TimeGrain_TIME_GRAIN_HOUR, r.Interval)
}

func TestServer_EstimateRollupInterval_date(t *testing.T) {
	server, instanceID := getTimeseriesTestServer(t)

	r, err := server.EstimateRollupInterval(context.Background(), &runtimev1.EstimateRollupIntervalRequest{
		InstanceId: instanceID,
		TableName:  "timeseries",
		ColumnName: "day",
		Priority:   1,
	})
	require.NoError(t, err)
	require.Equal(t, parseTime(t, "2019-01-01T00:00:00.000Z"), r.Start)
	require.Equal(t, parseTime(t, "2019-01-02T00:00:00.000Z"), r.End)
	require.Equal(t, runtimev1.TimeGrain_TIME_GRAIN_HOUR, r.Interval)
}
