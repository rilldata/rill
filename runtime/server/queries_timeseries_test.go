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

func TestServer_Timeseries(t *testing.T) {
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
			End:      parseTimeToProtoTimeStamps(t, "2019-12-02T00:00:00Z"),
			Interval: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
		},
		Filters: &runtimev1.MetricsViewFilter{
			Include: []*runtimev1.MetricsViewFilter_Cond{
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
	require.Equal(t, 1.0, results[0].Records.Fields["max"].GetNumberValue())
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

func Ignore_TestServer_Timeseries_exclude_notnull(t *testing.T) {
	t.Parallel()
	server, instanceID := getTimeseriesTestServer(t)

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
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Interval: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
		},
		Filters: &runtimev1.MetricsViewFilter{
			Exclude: []*runtimev1.MetricsViewFilter_Cond{
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
	require.Equal(t, 1.0, results[0].Records.Fields["count"])
}

func Ignore_TestServer_Timeseries_exclude_all(t *testing.T) {
	t.Parallel()
	server, instanceID := getTimeseriesTestServer(t)

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
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Interval: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
		},
		Filters: &runtimev1.MetricsViewFilter{
			Exclude: []*runtimev1.MetricsViewFilter_Cond{
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
	require.Equal(t, 0.0, results[0].Records.Fields["count"])
}

func TestServer_Timeseries_exclude_notnull_string(t *testing.T) {
	t.Parallel()
	server, instanceID := getTimeseriesTestServer(t)

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
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Interval: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
		},
		Filters: &runtimev1.MetricsViewFilter{
			Exclude: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "country",
					In:   []*structpb.Value{structpb.NewStringValue("Canada")},
				},
			},
		},
	})

	require.NoError(t, err)
	results := response.GetRollup().Results
	require.Equal(t, 1, len(results))
	require.Equal(t, 1.0, results[0].Records.Fields["count"].GetNumberValue())
}

func TestServer_Timeseries_exclude_all_string(t *testing.T) {
	t.Parallel()
	server, instanceID := getTimeseriesTestServer(t)

	response, err := server.ColumnTimeSeries(testCtx(), &runtimev1.ColumnTimeSeriesRequest{
		InstanceId: instanceID,
		TableName:  "timeseries",
		Measures: []*runtimev1.ColumnTimeSeriesRequest_BasicMeasure{
			{
				Expression: "sum(imps)",
				SqlName:    "Total impressions",
			},
		},
		TimestampColumnName: "time",
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Interval: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
		},
		Filters: &runtimev1.MetricsViewFilter{
			Exclude: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "country",
					In:   []*structpb.Value{structpb.NewStringValue("Canada"), structpb.NewNullValue()},
				},
			},
		},
	})

	require.NoError(t, err)
	results := response.GetRollup().Results
	require.Equal(t, 1, len(results))
	require.Equal(t, 0.0, results[0].Records.Fields["Total impressions"].GetNumberValue())
}

func TestServer_Timeseries_exclude_notnull_like(t *testing.T) {
	t.Parallel()
	server, instanceID := getTimeseriesTestServer(t)

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
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Interval: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
		},
		Filters: &runtimev1.MetricsViewFilter{
			Exclude: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "device",
					Like: []string{"iphone"},
				},
			},
		},
	})

	require.NoError(t, err)
	results := response.GetRollup().Results
	require.Equal(t, 1, len(results))
	require.Equal(t, 1.0, results[0].Records.Fields["count"].GetNumberValue())
}

func TestServer_Timeseries_exclude_like_all(t *testing.T) {
	t.Parallel()
	server, instanceID := getTimeseriesTestServer(t)

	response, err := server.ColumnTimeSeries(testCtx(), &runtimev1.ColumnTimeSeriesRequest{
		InstanceId: instanceID,
		TableName:  "timeseries",
		Measures: []*runtimev1.ColumnTimeSeriesRequest_BasicMeasure{
			{
				Expression: "sum(imps)",
				SqlName:    "Total impressions",
			},
		},
		TimestampColumnName: "time",
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Interval: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
		},
		Filters: &runtimev1.MetricsViewFilter{
			Exclude: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "country",
					In:   []*structpb.Value{structpb.NewNullValue()},
					Like: []string{"Canada"},
				},
			},
		},
	})

	require.NoError(t, err)
	results := response.GetRollup().Results
	require.Equal(t, 1, len(results))
	require.Equal(t, 0.0, results[0].Records.Fields["Total impressions"].GetNumberValue())
}

func TestServer_Timeseries_numeric_dim(t *testing.T) {
	t.Parallel()
	server, instanceID := getTimeseriesTestServer(t)

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
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Interval: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
		},
		Filters: &runtimev1.MetricsViewFilter{
			Include: []*runtimev1.MetricsViewFilter_Cond{
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
	require.Equal(t, 1.0, results[0].Records.Fields["count"].GetNumberValue())
}

func TestServer_Timeseries_numeric_dim_2values(t *testing.T) {
	t.Parallel()
	server, instanceID := getTimeseriesTestServer(t)

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
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Interval: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
		},
		Filters: &runtimev1.MetricsViewFilter{
			Include: []*runtimev1.MetricsViewFilter_Cond{
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
	require.Equal(t, 1.0, results[0].Records.Fields["count"].GetNumberValue())
}

func TestServer_Timeseries_numeric_dim_and_null(t *testing.T) {
	t.Parallel()
	server, instanceID := getTimeseriesTestServer(t)

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
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Interval: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
		},
		Filters: &runtimev1.MetricsViewFilter{
			Include: []*runtimev1.MetricsViewFilter_Cond{
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
	require.Equal(t, 2.0, results[0].Records.Fields["count"].GetNumberValue())
}

func TestServer_Timeseries_Empty_TimeRange(t *testing.T) {
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
		TimeRange:           new(runtimev1.TimeSeriesTimeRange),
		Filters: &runtimev1.MetricsViewFilter{
			Include: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "device",
					In:   []*structpb.Value{structpb.NewStringValue("android"), structpb.NewStringValue("iphone")},
				},
			},
		},
	})

	require.NoError(t, err)
	results := response.GetRollup().Results
	for i, v := range response.GetRollup().Results {
		fmt.Printf("i: %d, ts: %v\n", i, v.Ts.AsTime())
	}
	require.Equal(t, 25, len(results))
	require.Equal(t, 1.0, results[0].Records.Fields["max"].GetNumberValue())
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
		Filters: new(runtimev1.MetricsViewFilter),
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
		Filters: new(runtimev1.MetricsViewFilter),
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
		Filters: new(runtimev1.MetricsViewFilter),
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
		Filters: new(runtimev1.MetricsViewFilter),
	})

	require.NoError(t, err)
	results := response.GetRollup().Results
	require.Equal(t, 1, len(results))
	require.Equal(t, 2.0, results[0].Records.Fields["count"].GetNumberValue())
}

func TestServer_Timeseries_2measures(t *testing.T) {
	t.Parallel()
	server, instanceID := getTimeseriesTestServer(t)

	mx := "max"
	sm := "sum"
	response, err := server.ColumnTimeSeries(testCtx(), &runtimev1.ColumnTimeSeriesRequest{
		InstanceId: instanceID,
		TableName:  "timeseries",
		Measures: []*runtimev1.ColumnTimeSeriesRequest_BasicMeasure{
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
			Start:    parseTimeToProtoTimeStamps(t, "2019-01-01T00:00:00Z"),
			End:      parseTimeToProtoTimeStamps(t, "2019-12-01T00:00:00Z"),
			Interval: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
		},
		Filters: &runtimev1.MetricsViewFilter{
			Include: []*runtimev1.MetricsViewFilter_Cond{
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
	require.Equal(t, 1.0, results[0].Records.Fields["max"].GetNumberValue())
	require.Equal(t, 2.0, results[0].Records.Fields["sum"].GetNumberValue())
}

func TestServer_Timeseries_1dim(t *testing.T) {
	t.Parallel()
	server, instanceID := getTimeseriesTestServer(t)

	sm := "sum"
	response, err := server.ColumnTimeSeries(testCtx(), &runtimev1.ColumnTimeSeriesRequest{
		InstanceId: instanceID,
		TableName:  "timeseries",
		Measures: []*runtimev1.ColumnTimeSeriesRequest_BasicMeasure{
			{
				Expression: "sum(clicks)",
				SqlName:    sm,
			},
		},
		TimestampColumnName: "time",
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Start:    parseTimeToProtoTimeStamps(t, "2019-01-01T00:00:00Z"),
			End:      parseTimeToProtoTimeStamps(t, "2019-12-01T00:00:00Z"),
			Interval: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
		},
		Filters: &runtimev1.MetricsViewFilter{
			Include: []*runtimev1.MetricsViewFilter_Cond{
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
	require.Equal(t, 1.0, results[0].Records.Fields["sum"].GetNumberValue())
}

func TestServer_Timeseries_1dim_null(t *testing.T) {
	t.Parallel()
	server, instanceID := getTimeseriesTestServer(t)

	response, err := server.ColumnTimeSeries(testCtx(), &runtimev1.ColumnTimeSeriesRequest{
		InstanceId: instanceID,
		TableName:  "timeseries",
		Measures: []*runtimev1.ColumnTimeSeriesRequest_BasicMeasure{
			{
				Expression: "sum(clicks)",
				SqlName:    "sum",
			},
		},
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Interval: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
		},
		TimestampColumnName: "time",
		Filters: &runtimev1.MetricsViewFilter{
			Include: []*runtimev1.MetricsViewFilter_Cond{
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
	require.Equal(t, 1.0, results[0].Records.Fields["sum"].GetNumberValue())
}

func TestServer_Timeseries_1dim_null_and_in(t *testing.T) {
	t.Parallel()
	server, instanceID := getTimeseriesTestServer(t)

	response, err := server.ColumnTimeSeries(testCtx(), &runtimev1.ColumnTimeSeriesRequest{
		InstanceId: instanceID,
		TableName:  "timeseries",
		Measures: []*runtimev1.ColumnTimeSeriesRequest_BasicMeasure{
			{
				Expression: "sum(clicks)",
				SqlName:    "sum",
			},
		},
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Interval: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
		},
		TimestampColumnName: "time",
		Filters: &runtimev1.MetricsViewFilter{
			Include: []*runtimev1.MetricsViewFilter_Cond{
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
	require.Equal(t, 2.0, results[0].Records.Fields["sum"].GetNumberValue())
}

func TestServer_Timeseries_1dim_null_and_in_and_like(t *testing.T) {
	t.Parallel()
	server, instanceID := getTimeseriesTestServer(t)

	response, err := server.ColumnTimeSeries(testCtx(), &runtimev1.ColumnTimeSeriesRequest{
		InstanceId: instanceID,
		TableName:  "timeseries",
		Measures: []*runtimev1.ColumnTimeSeriesRequest_BasicMeasure{
			{
				Expression: "sum(clicks)",
				SqlName:    "sum",
			},
		},
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Interval: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
		},
		TimestampColumnName: "time",
		Filters: &runtimev1.MetricsViewFilter{
			Include: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "publisher",
					In: []*structpb.Value{
						structpb.NewNullValue(),
						structpb.NewStringValue("Google"),
					},
					Like: []string{
						"Goo%",
					},
				},
			},
		},
	})

	require.NoError(t, err)
	results := response.GetRollup().Results
	require.Equal(t, 1, len(results))
	require.Equal(t, 2.0, results[0].Records.Fields["sum"].GetNumberValue())
}

func TestServer_Timeseries_1dim_2like(t *testing.T) {
	t.Parallel()
	server, instanceID := getTimeseriesTestServer(t)

	response, err := server.ColumnTimeSeries(testCtx(), &runtimev1.ColumnTimeSeriesRequest{
		InstanceId: instanceID,
		TableName:  "timeseries",
		Measures: []*runtimev1.ColumnTimeSeriesRequest_BasicMeasure{
			{
				Expression: "sum(clicks)",
				SqlName:    "sum",
			},
		},
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Interval: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
		},
		TimestampColumnName: "time",
		Filters: &runtimev1.MetricsViewFilter{
			Include: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "domain",
					Like: []string{
						"g%",
						"msn%",
					},
				},
			},
		},
	})

	require.NoError(t, err)
	results := response.GetRollup().Results
	require.Equal(t, 1, len(results))
	require.Equal(t, 2.0, results[0].Records.Fields["sum"].GetNumberValue())
}

func TestServer_Timeseries_2dim_include_and_exclude(t *testing.T) {
	t.Parallel()
	server, instanceID := getTimeseriesTestServer(t)

	response, err := server.ColumnTimeSeries(testCtx(), &runtimev1.ColumnTimeSeriesRequest{
		InstanceId: instanceID,
		TableName:  "timeseries",
		Measures: []*runtimev1.ColumnTimeSeriesRequest_BasicMeasure{
			{
				Expression: "sum(clicks)",
				SqlName:    "sum",
			},
		},
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Interval: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
		},
		TimestampColumnName: "time",
		Filters: &runtimev1.MetricsViewFilter{
			Include: []*runtimev1.MetricsViewFilter_Cond{
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
	require.Equal(t, 0.0, results[0].Records.Fields["sum"].GetNumberValue())
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

func TestServer_Timeseries_1day(t *testing.T) {
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
			End:      parseTimeToProtoTimeStamps(t, "2019-01-03T00:00:00Z"),
			Interval: runtimev1.TimeGrain_TIME_GRAIN_DAY,
		},
		Filters: &runtimev1.MetricsViewFilter{
			Include: []*runtimev1.MetricsViewFilter_Cond{
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
	t.Parallel()
	server, instanceID := getTimeseriesTestServer(t)

	cnt := "count"
	response, err := server.ColumnTimeSeries(testCtx(), &runtimev1.ColumnTimeSeriesRequest{
		InstanceId: instanceID,
		TableName:  "timeseries",
		Measures: []*runtimev1.ColumnTimeSeriesRequest_BasicMeasure{
			{
				Expression: "count(*)",
				SqlName:    cnt,
			},
		},
		TimestampColumnName: "time",
		TimeRange: &runtimev1.TimeSeriesTimeRange{
			Start:    parseTimeToProtoTimeStamps(t, "2019-01-01T00:00:00Z"),
			End:      parseTimeToProtoTimeStamps(t, "2019-01-03T00:00:00Z"),
			Interval: runtimev1.TimeGrain_TIME_GRAIN_DAY,
		},
		Filters: &runtimev1.MetricsViewFilter{
			Include: []*runtimev1.MetricsViewFilter_Cond{
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
