package server_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/expressionpb"
	"github.com/rilldata/rill/runtime/queries"
	"github.com/rilldata/rill/runtime/server/auth"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestServer_MetricsViewTimeSeries(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewTimeSeries(testCtx(), &runtimev1.MetricsViewTimeSeriesRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_DAY,
		MeasureNames:    []string{"measure_0", "measure_2"},
	})
	require.NoError(t, err)
	require.Equal(t, 2, len(tr.Data))
	require.Equal(t, 2, len(tr.Data[0].Records.Fields))

	require.Equal(t, parseTime(t, "2022-01-01T00:00:00Z"), tr.Data[0].Ts.AsTime())
	require.Equal(t, 1.0, tr.Data[0].Records.Fields["measure_0"].GetNumberValue())
	require.Equal(t, 2.0, tr.Data[0].Records.Fields["measure_2"].GetNumberValue())

	require.Equal(t, parseTime(t, "2022-01-02T00:00:00Z"), tr.Data[1].Ts.AsTime())
	require.Equal(t, 1.0, tr.Data[1].Records.Fields["measure_0"].GetNumberValue())
	require.Equal(t, 1.0, tr.Data[1].Records.Fields["measure_2"].GetNumberValue())
}

func TestServer_MetricsViewTimeSeries_TimeEnd_exclusive(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewTimeSeries(testCtx(), &runtimev1.MetricsViewTimeSeriesRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_DAY,
		TimeStart:       parseTimeToProtoTimeStamps(t, "2022-01-01T00:00:00Z"),
		TimeEnd:         parseTimeToProtoTimeStamps(t, "2022-01-02T00:00:00Z"),
		MeasureNames:    []string{"measure_0", "measure_2"},
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Data))
	require.Equal(t, 2, len(tr.Data[0].Records.Fields))

	require.Equal(t, parseTime(t, "2022-01-01T00:00:00Z"), tr.Data[0].Ts.AsTime())
	require.Equal(t, 1.0, tr.Data[0].Records.Fields["measure_0"].GetNumberValue())
	require.Equal(t, 2.0, tr.Data[0].Records.Fields["measure_2"].GetNumberValue())
}

func TestServer_MetricsViewTimeSeries_complete_source_sanity_test(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids")

	tr, err := server.MetricsViewTimeSeries(testCtx(), &runtimev1.MetricsViewTimeSeriesRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_DAY,
		MeasureNames:    []string{"measure_0", "measure_1"},
		Where: expressionpb.Or([]*runtimev1.Expression{
			expressionpb.In(expressionpb.Identifier("dom"), []*runtimev1.Expression{expressionpb.Value(structpb.NewStringValue("msn.com"))}),
			expressionpb.Like(expressionpb.Identifier("dom"), expressionpb.Value(structpb.NewStringValue("%yahoo%"))),
		}),
	})
	require.NoError(t, err)
	require.True(t, len(tr.Data) > 0)
	require.Equal(t, 2, len(tr.Data[0].Records.Fields))
	require.NotEmpty(t, tr.Data[0].Ts.AsTime())
	require.True(t, tr.Data[0].Records.Fields["measure_0"].GetNumberValue() > 0)
	require.True(t, tr.Data[0].Records.Fields["measure_1"].GetNumberValue() > 0)
}

func TestServer_Timeseries(t *testing.T) {
	t.Parallel()
	server, instanceID := getMetricsTestServer(t, "timeseries")

	response, err := server.MetricsViewTimeSeries(testCtx(), &runtimev1.MetricsViewTimeSeriesRequest{
		InstanceId:      instanceID,
		MetricsViewName: "timeseries",
		MeasureNames:    []string{"max_clicks"},
		TimeStart:       parseTimeToProtoTimeStamps(t, "2019-01-01T00:00:00Z"),
		TimeEnd:         parseTimeToProtoTimeStamps(t, "2019-12-02T00:00:00Z"),
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
		Where: expressionpb.In(
			expressionpb.Identifier("device"),
			[]*runtimev1.Expression{expressionpb.Value(structpb.NewStringValue("android")), expressionpb.Value(structpb.NewStringValue("iphone"))},
		),
	})

	require.NoError(t, err)
	results := response.Data
	require.Equal(t, 1, len(results))
	require.Equal(t, 1.0, results[0].Records.Fields["max_clicks"].GetNumberValue())
}

func Ignore_TestServer_Timeseries_exclude_notnull(t *testing.T) {
	t.Parallel()
	server, instanceID := getMetricsTestServer(t, "timeseries")

	response, err := server.MetricsViewTimeSeries(testCtx(), &runtimev1.MetricsViewTimeSeriesRequest{
		InstanceId:      instanceID,
		MetricsViewName: "timeseries",
		MeasureNames:    []string{"count"},
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
		Where: expressionpb.NotIn(
			expressionpb.Identifier("latitude"),
			[]*runtimev1.Expression{expressionpb.Value(structpb.NewNumberValue(25))},
		),
	})

	require.NoError(t, err)
	results := response.Data
	require.Equal(t, 1, len(results))
	require.Equal(t, 1.0, results[0].Records.Fields["count"])
}

func Ignore_TestServer_Timeseries_exclude_all(t *testing.T) {
	t.Parallel()
	server, instanceID := getMetricsTestServer(t, "timeseries")

	response, err := server.MetricsViewTimeSeries(testCtx(), &runtimev1.MetricsViewTimeSeriesRequest{
		InstanceId:      instanceID,
		MetricsViewName: "timeseries",
		MeasureNames:    []string{"count"},
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
		Where: expressionpb.NotIn(
			expressionpb.Identifier("latitude"),
			[]*runtimev1.Expression{expressionpb.Value(structpb.NewNumberValue(25)), expressionpb.Value(structpb.NewNullValue())},
		),
	})

	require.NoError(t, err)
	results := response.Data
	require.Equal(t, 1, len(results))
	require.Equal(t, 0.0, results[0].Records.Fields["count"])
}

func TestServer_Timeseries_exclude_notnull_string(t *testing.T) {
	t.Parallel()
	server, instanceID := getMetricsTestServer(t, "timeseries")

	response, err := server.MetricsViewTimeSeries(testCtx(), &runtimev1.MetricsViewTimeSeriesRequest{
		InstanceId:      instanceID,
		MetricsViewName: "timeseries",
		MeasureNames:    []string{"count"},
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
		Where: expressionpb.NotIn(
			expressionpb.Identifier("country"),
			[]*runtimev1.Expression{expressionpb.Value(structpb.NewStringValue("Canada"))},
		),
	})

	require.NoError(t, err)
	results := response.Data
	require.Equal(t, 1, len(results))
	require.Equal(t, 1.0, results[0].Records.Fields["count"].GetNumberValue())
}

func TestServer_Timeseries_exclude_all_string(t *testing.T) {
	t.Parallel()
	server, instanceID := getMetricsTestServer(t, "timeseries")

	response, err := server.MetricsViewTimeSeries(testCtx(), &runtimev1.MetricsViewTimeSeriesRequest{
		InstanceId:      instanceID,
		MetricsViewName: "timeseries",
		MeasureNames:    []string{"sum_imps"},
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
		Where: expressionpb.NotIn(
			expressionpb.Identifier("country"),
			[]*runtimev1.Expression{expressionpb.Value(structpb.NewStringValue("Canada")), expressionpb.Value(structpb.NewNullValue())},
		),
	})

	require.NoError(t, err)
	results := response.Data
	require.Equal(t, 0, len(results))
}

func TestServer_Timeseries_exclude_notnull_like(t *testing.T) {
	t.Parallel()
	server, instanceID := getMetricsTestServer(t, "timeseries")

	response, err := server.MetricsViewTimeSeries(testCtx(), &runtimev1.MetricsViewTimeSeriesRequest{
		InstanceId:      instanceID,
		MetricsViewName: "timeseries",
		MeasureNames:    []string{"count"},
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
		Where: expressionpb.NotLike(
			expressionpb.Identifier("device"),
			expressionpb.Value(structpb.NewStringValue("iphone")),
		),
	})

	require.NoError(t, err)
	results := response.Data
	require.Equal(t, 1, len(results))
	require.Equal(t, 1.0, results[0].Records.Fields["count"].GetNumberValue())
}

func TestServer_Timeseries_exclude_like_all(t *testing.T) {
	t.Parallel()
	server, instanceID := getMetricsTestServer(t, "timeseries")

	response, err := server.MetricsViewTimeSeries(testCtx(), &runtimev1.MetricsViewTimeSeriesRequest{
		InstanceId:      instanceID,
		MetricsViewName: "timeseries",
		MeasureNames:    []string{"sum_imps"},
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
		Where: expressionpb.And([]*runtimev1.Expression{
			expressionpb.NotIn(
				expressionpb.Identifier("country"),
				[]*runtimev1.Expression{expressionpb.Value(structpb.NewNullValue())},
			),
			expressionpb.NotLike(
				expressionpb.Identifier("country"),
				expressionpb.Value(structpb.NewStringValue("Canada")),
			),
		}),
	})

	require.NoError(t, err)
	results := response.Data
	require.Equal(t, 0, len(results))
}

func TestServer_Timeseries_numeric_dim(t *testing.T) {
	t.Parallel()
	server, instanceID := getMetricsTestServer(t, "timeseries")

	response, err := server.MetricsViewTimeSeries(testCtx(), &runtimev1.MetricsViewTimeSeriesRequest{
		InstanceId:      instanceID,
		MetricsViewName: "timeseries",
		MeasureNames:    []string{"count"},
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
		Where: expressionpb.In(
			expressionpb.Identifier("latitude"),
			[]*runtimev1.Expression{expressionpb.Value(structpb.NewNumberValue(25))},
		),
	})

	require.NoError(t, err)
	results := response.Data
	require.Equal(t, 1, len(results))
	require.Equal(t, 1.0, results[0].Records.Fields["count"].GetNumberValue())
}

func TestServer_Timeseries_numeric_dim_2values(t *testing.T) {
	t.Parallel()
	server, instanceID := getMetricsTestServer(t, "timeseries")

	response, err := server.MetricsViewTimeSeries(testCtx(), &runtimev1.MetricsViewTimeSeriesRequest{
		InstanceId:      instanceID,
		MetricsViewName: "timeseries",
		MeasureNames:    []string{"count"},
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
		Where: expressionpb.In(
			expressionpb.Identifier("latitude"),
			[]*runtimev1.Expression{expressionpb.Value(structpb.NewNumberValue(25)), expressionpb.Value(structpb.NewNumberValue(35))},
		),
	})

	require.NoError(t, err)
	results := response.Data
	require.Equal(t, 1, len(results))
	require.Equal(t, 1.0, results[0].Records.Fields["count"].GetNumberValue())
}

func TestServer_Timeseries_numeric_dim_and_null(t *testing.T) {
	t.Parallel()
	server, instanceID := getMetricsTestServer(t, "timeseries")

	response, err := server.MetricsViewTimeSeries(testCtx(), &runtimev1.MetricsViewTimeSeriesRequest{
		InstanceId:      instanceID,
		MetricsViewName: "timeseries",
		MeasureNames:    []string{"count"},
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
		Where: expressionpb.In(
			expressionpb.Identifier("latitude"),
			[]*runtimev1.Expression{expressionpb.Value(structpb.NewNumberValue(25)), expressionpb.Value(structpb.NewNullValue())},
		),
	})

	require.NoError(t, err)
	results := response.Data
	require.Equal(t, 1, len(results))
	require.Equal(t, 2.0, results[0].Records.Fields["count"].GetNumberValue())
}

func TestServer_Timeseries_TimeRange_Day(t *testing.T) {
	t.Parallel()
	server, instanceID := getMetricsTestServer(t, "timeseries")

	response, err := server.MetricsViewTimeSeries(testCtx(), &runtimev1.MetricsViewTimeSeriesRequest{
		InstanceId:      instanceID,
		MetricsViewName: "timeseries",
		MeasureNames:    []string{"max_clicks"},
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_DAY,
	})

	require.NoError(t, err)
	results := response.Data
	for i, v := range response.Data {
		fmt.Printf("i: %d, ts: %v\n", i, v.Ts.AsTime())
	}
	require.Equal(t, 2, len(results))
	require.Equal(t, 1.0, results[0].Records.Fields["max_clicks"].GetNumberValue())
	require.Equal(t, parseTime(t, "2019-01-01T00:00:00Z"), results[0].Ts.AsTime())
	require.Equal(t, parseTime(t, "2019-01-02T00:00:00Z"), results[1].Ts.AsTime())
}

func TestServer_Timeseries_TimeRange_Start(t *testing.T) {
	t.Parallel()
	server, instanceID := getMetricsTestServer(t, "timeseries")

	response, err := server.MetricsViewTimeSeries(testCtx(), &runtimev1.MetricsViewTimeSeriesRequest{
		InstanceId:      instanceID,
		MetricsViewName: "timeseries",
		MeasureNames:    []string{"max_clicks"},
		TimeStart:       timestamppb.New(parseTime(t, "2018-12-31T00:00:00Z")),
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_DAY,
	})

	require.NoError(t, err)
	results := response.Data
	for i, v := range response.Data {
		fmt.Printf("i: %d, ts: %v\n", i, v.Ts.AsTime())
	}
	require.Equal(t, 3, len(results))
	require.Equal(t, parseTime(t, "2018-12-31T00:00:00Z"), results[0].Ts.AsTime())
	require.Equal(t, structpb.NullValue_NULL_VALUE, results[0].Records.Fields["max_clicks"].GetNullValue())
	require.Equal(t, parseTime(t, "2019-01-01T00:00:00Z"), results[1].Ts.AsTime())
	require.Equal(t, 1.0, results[1].Records.Fields["max_clicks"].GetNumberValue())
	require.Equal(t, parseTime(t, "2019-01-02T00:00:00Z"), results[2].Ts.AsTime())
	require.Equal(t, 1.0, results[2].Records.Fields["max_clicks"].GetNumberValue())
}

func TestServer_Timeseries_TimeRange_Start_2day_before(t *testing.T) {
	t.Parallel()
	server, instanceID := getMetricsTestServer(t, "timeseries")

	response, err := server.MetricsViewTimeSeries(testCtx(), &runtimev1.MetricsViewTimeSeriesRequest{
		InstanceId:      instanceID,
		MetricsViewName: "timeseries",
		MeasureNames:    []string{"max_clicks"},
		TimeStart:       timestamppb.New(parseTime(t, "2018-12-30T00:00:00Z")),
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_DAY,
	})

	require.NoError(t, err)
	results := response.Data
	for i, v := range response.Data {
		fmt.Printf("i: %d, ts: %v\n", i, v.Ts.AsTime())
	}
	require.Equal(t, 4, len(results))
	i := 0
	require.Equal(t, parseTime(t, "2018-12-30T00:00:00Z"), results[i].Ts.AsTime())
	require.Equal(t, structpb.NullValue_NULL_VALUE, results[i].Records.Fields["max_clicks"].GetNullValue())
	i += 1
	require.Equal(t, parseTime(t, "2018-12-31T00:00:00Z"), results[i].Ts.AsTime())
	require.Equal(t, structpb.NullValue_NULL_VALUE, results[i].Records.Fields["max_clicks"].GetNullValue())
	i += 1
	require.Equal(t, parseTime(t, "2019-01-01T00:00:00Z"), results[i].Ts.AsTime())
	require.Equal(t, 1.0, results[i].Records.Fields["max_clicks"].GetNumberValue())
	i += 1
	require.Equal(t, parseTime(t, "2019-01-02T00:00:00Z"), results[i].Ts.AsTime())
	require.Equal(t, 1.0, results[i].Records.Fields["max_clicks"].GetNumberValue())
}

func TestServer_Timeseries_TimeRange_End(t *testing.T) {
	t.Parallel()
	server, instanceID := getMetricsTestServer(t, "timeseries")

	response, err := server.MetricsViewTimeSeries(testCtx(), &runtimev1.MetricsViewTimeSeriesRequest{
		InstanceId:      instanceID,
		MetricsViewName: "timeseries",
		MeasureNames:    []string{"max_clicks"},
		TimeEnd:         timestamppb.New(parseTime(t, "2019-01-04T00:00:00Z")),
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_DAY,
	})

	require.NoError(t, err)
	results := response.Data
	for i, v := range response.Data {
		fmt.Printf("i: %d, ts: %v\n", i, v.Ts.AsTime())
	}
	require.Equal(t, 3, len(results))
	i := 0
	require.Equal(t, parseTime(t, "2019-01-01T00:00:00Z"), results[i].Ts.AsTime())
	require.Equal(t, 1.0, results[i].Records.Fields["max_clicks"].GetNumberValue())
	i += 1
	require.Equal(t, parseTime(t, "2019-01-02T00:00:00Z"), results[i].Ts.AsTime())
	require.Equal(t, 1.0, results[i].Records.Fields["max_clicks"].GetNumberValue())
	i += 1
	require.Equal(t, parseTime(t, "2019-01-03T00:00:00Z"), results[i].Ts.AsTime())
	require.Equal(t, structpb.NullValue_NULL_VALUE, results[i].Records.Fields["max_clicks"].GetNullValue())
}

func TestServer_Timeseries_TimeRange_End_2day_after(t *testing.T) {
	t.Parallel()
	server, instanceID := getMetricsTestServer(t, "timeseries")

	response, err := server.MetricsViewTimeSeries(testCtx(), &runtimev1.MetricsViewTimeSeriesRequest{
		InstanceId:      instanceID,
		MetricsViewName: "timeseries",
		MeasureNames:    []string{"max_clicks"},
		TimeEnd:         timestamppb.New(parseTime(t, "2019-01-05T00:00:00Z")),
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_DAY,
	})

	require.NoError(t, err)
	results := response.Data
	for i, v := range response.Data {
		fmt.Printf("i: %d, ts: %v\n", i, v.Ts.AsTime())
	}
	require.Equal(t, 4, len(results))
	i := 0
	require.Equal(t, parseTime(t, "2019-01-01T00:00:00Z"), results[i].Ts.AsTime())
	require.Equal(t, 1.0, results[i].Records.Fields["max_clicks"].GetNumberValue())
	i += 1
	require.Equal(t, parseTime(t, "2019-01-02T00:00:00Z"), results[i].Ts.AsTime())
	require.Equal(t, 1.0, results[i].Records.Fields["max_clicks"].GetNumberValue())
	i += 1
	require.Equal(t, parseTime(t, "2019-01-03T00:00:00Z"), results[i].Ts.AsTime())
	require.Equal(t, structpb.NullValue_NULL_VALUE, results[i].Records.Fields["max_clicks"].GetNullValue())
	i += 1
	require.Equal(t, parseTime(t, "2019-01-04T00:00:00Z"), results[i].Ts.AsTime())
	require.Equal(t, structpb.NullValue_NULL_VALUE, results[i].Records.Fields["max_clicks"].GetNullValue())

}

func TestServer_Timeseries_TimeRange_middle_nulls(t *testing.T) {
	t.Parallel()
	server, instanceID := getMetricsTestServer(t, "timeseries")

	response, err := server.MetricsViewTimeSeries(testCtx(), &runtimev1.MetricsViewTimeSeriesRequest{
		InstanceId:      instanceID,
		MetricsViewName: "timeseries_gaps",
		MeasureNames:    []string{"max_clicks"},
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_DAY,
	})

	require.NoError(t, err)
	results := response.Data
	for i, v := range response.Data {
		fmt.Printf("i: %d, ts: %v\n", i, v.Ts.AsTime())
	}
	require.Equal(t, 6, len(results))
	i := 0
	require.Equal(t, parseTime(t, "2019-01-01T00:00:00Z"), results[i].Ts.AsTime())
	require.Equal(t, 1.0, results[i].Records.Fields["max_clicks"].GetNumberValue())
	i += 1
	require.Equal(t, parseTime(t, "2019-01-02T00:00:00Z"), results[i].Ts.AsTime())
	require.Equal(t, structpb.NullValue_NULL_VALUE, results[i].Records.Fields["max_clicks"].GetNullValue())
	i += 1
	require.Equal(t, parseTime(t, "2019-01-03T00:00:00Z"), results[i].Ts.AsTime())
	require.Equal(t, 1.0, results[i].Records.Fields["max_clicks"].GetNumberValue())
	i += 1
	require.Equal(t, parseTime(t, "2019-01-04T00:00:00Z"), results[i].Ts.AsTime())
	require.Equal(t, structpb.NullValue_NULL_VALUE, results[i].Records.Fields["max_clicks"].GetNullValue())
	i += 1
	require.Equal(t, parseTime(t, "2019-01-05T00:00:00Z"), results[i].Ts.AsTime())
	require.Equal(t, structpb.NullValue_NULL_VALUE, results[i].Records.Fields["max_clicks"].GetNullValue())
	i += 1
	require.Equal(t, parseTime(t, "2019-01-06T00:00:00Z"), results[i].Ts.AsTime())
	require.Equal(t, 1.0, results[i].Records.Fields["max_clicks"].GetNumberValue())
}

func TestServer_Timeseries_2measures(t *testing.T) {
	t.Parallel()
	server, instanceID := getMetricsTestServer(t, "timeseries")

	response, err := server.MetricsViewTimeSeries(testCtx(), &runtimev1.MetricsViewTimeSeriesRequest{
		InstanceId:      instanceID,
		MetricsViewName: "timeseries",
		MeasureNames:    []string{"max_clicks", "sum_clicks"},
		TimeStart:       parseTimeToProtoTimeStamps(t, "2019-01-01T00:00:00Z"),
		TimeEnd:         parseTimeToProtoTimeStamps(t, "2019-12-01T00:00:00Z"),
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
		Where: expressionpb.In(
			expressionpb.Identifier("device"),
			[]*runtimev1.Expression{expressionpb.Value(structpb.NewStringValue("android")), expressionpb.Value(structpb.NewStringValue("iphone"))},
		),
	})

	require.NoError(t, err)
	results := response.Data
	require.Equal(t, 1, len(results))
	require.Equal(t, 1.0, results[0].Records.Fields["max_clicks"].GetNumberValue())
	require.Equal(t, 2.0, results[0].Records.Fields["sum_clicks"].GetNumberValue())
}

func TestServer_Timeseries_1dim(t *testing.T) {
	t.Parallel()
	server, instanceID := getMetricsTestServer(t, "timeseries")

	response, err := server.MetricsViewTimeSeries(testCtx(), &runtimev1.MetricsViewTimeSeriesRequest{
		InstanceId:      instanceID,
		MetricsViewName: "timeseries",
		MeasureNames:    []string{"sum_clicks"},
		TimeStart:       parseTimeToProtoTimeStamps(t, "2019-01-01T00:00:00Z"),
		TimeEnd:         parseTimeToProtoTimeStamps(t, "2019-12-01T00:00:00Z"),
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
		Where: expressionpb.In(
			expressionpb.Identifier("device"),
			[]*runtimev1.Expression{expressionpb.Value(structpb.NewStringValue("android"))},
		),
	})

	require.NoError(t, err)
	results := response.Data
	require.Equal(t, 1, len(results))
	require.Equal(t, 1.0, results[0].Records.Fields["sum_clicks"].GetNumberValue())
}

func TestServer_Timeseries_1dim_null(t *testing.T) {
	t.Parallel()
	server, instanceID := getMetricsTestServer(t, "timeseries")

	response, err := server.MetricsViewTimeSeries(testCtx(), &runtimev1.MetricsViewTimeSeriesRequest{
		InstanceId:      instanceID,
		MetricsViewName: "timeseries",
		MeasureNames:    []string{"sum_clicks"},
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
		Where: expressionpb.In(
			expressionpb.Identifier("publisher"),
			[]*runtimev1.Expression{expressionpb.Value(structpb.NewNullValue())},
		),
	})

	require.NoError(t, err)
	results := response.Data
	require.Equal(t, 1, len(results))
	require.Equal(t, 1.0, results[0].Records.Fields["sum_clicks"].GetNumberValue())
}

func TestServer_Timeseries_1dim_null_and_in(t *testing.T) {
	t.Parallel()
	server, instanceID := getMetricsTestServer(t, "timeseries")

	response, err := server.MetricsViewTimeSeries(testCtx(), &runtimev1.MetricsViewTimeSeriesRequest{
		InstanceId:      instanceID,
		MetricsViewName: "timeseries",
		MeasureNames:    []string{"sum_clicks"},
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
		Where: expressionpb.In(
			expressionpb.Identifier("publisher"),
			[]*runtimev1.Expression{expressionpb.Value(structpb.NewNullValue()), expressionpb.Value(structpb.NewStringValue("Google"))},
		),
	})

	require.NoError(t, err)
	results := response.Data
	require.Equal(t, 1, len(results))
	require.Equal(t, 2.0, results[0].Records.Fields["sum_clicks"].GetNumberValue())
}

func TestServer_Timeseries_1dim_null_and_in_and_like(t *testing.T) {
	t.Parallel()
	server, instanceID := getMetricsTestServer(t, "timeseries")

	response, err := server.MetricsViewTimeSeries(testCtx(), &runtimev1.MetricsViewTimeSeriesRequest{
		InstanceId:      instanceID,
		MetricsViewName: "timeseries",
		MeasureNames:    []string{"sum_clicks"},
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
		Where: expressionpb.Or([]*runtimev1.Expression{
			expressionpb.In(
				expressionpb.Identifier("publisher"),
				[]*runtimev1.Expression{expressionpb.Value(structpb.NewNullValue()), expressionpb.Value(structpb.NewStringValue("Google"))},
			),
			expressionpb.Like(
				expressionpb.Identifier("publisher"),
				expressionpb.Value(structpb.NewStringValue("Goo%")),
			),
		}),
	})

	require.NoError(t, err)
	results := response.Data
	require.Equal(t, 1, len(results))
	require.Equal(t, 2.0, results[0].Records.Fields["sum_clicks"].GetNumberValue())
}

func TestServer_Timeseries_1dim_2like(t *testing.T) {
	t.Parallel()
	server, instanceID := getMetricsTestServer(t, "timeseries")

	response, err := server.MetricsViewTimeSeries(testCtx(), &runtimev1.MetricsViewTimeSeriesRequest{
		InstanceId:      instanceID,
		MetricsViewName: "timeseries",
		MeasureNames:    []string{"sum_clicks"},
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
		Where: expressionpb.Or([]*runtimev1.Expression{
			expressionpb.Like(
				expressionpb.Identifier("domain"),
				expressionpb.Value(structpb.NewStringValue("g%")),
			),
			expressionpb.Like(
				expressionpb.Identifier("domain"),
				expressionpb.Value(structpb.NewStringValue("msn%")),
			),
		}),
	})

	require.NoError(t, err)
	results := response.Data
	require.Equal(t, 1, len(results))
	require.Equal(t, 2.0, results[0].Records.Fields["sum_clicks"].GetNumberValue())
}

func TestServer_Timeseries_2dim_include_and_exclude(t *testing.T) {
	t.Parallel()
	server, instanceID := getMetricsTestServer(t, "timeseries")

	response, err := server.MetricsViewTimeSeries(testCtx(), &runtimev1.MetricsViewTimeSeriesRequest{
		InstanceId:      instanceID,
		MetricsViewName: "timeseries",
		MeasureNames:    []string{"sum_clicks"},
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
		Where: expressionpb.And([]*runtimev1.Expression{
			expressionpb.In(
				expressionpb.Identifier("publisher"),
				[]*runtimev1.Expression{expressionpb.Value(structpb.NewStringValue("Google"))},
			),
			expressionpb.In(
				expressionpb.Identifier("domain"),
				[]*runtimev1.Expression{expressionpb.Value(structpb.NewStringValue("msn.com"))},
			),
		}),
	})

	require.NoError(t, err)
	results := response.Data
	require.Equal(t, 0, len(results))
}

func TestServer_Timeseries_1day(t *testing.T) {
	t.Parallel()
	server, instanceID := getMetricsTestServer(t, "timeseries")

	response, err := server.MetricsViewTimeSeries(testCtx(), &runtimev1.MetricsViewTimeSeriesRequest{
		InstanceId:      instanceID,
		MetricsViewName: "timeseries",
		MeasureNames:    []string{"max_clicks"},
		TimeStart:       parseTimeToProtoTimeStamps(t, "2019-01-01T00:00:00Z"),
		TimeEnd:         parseTimeToProtoTimeStamps(t, "2019-01-03T00:00:00Z"),
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_DAY,
		Where: expressionpb.In(
			expressionpb.Identifier("device"),
			[]*runtimev1.Expression{expressionpb.Value(structpb.NewStringValue("android")), expressionpb.Value(structpb.NewStringValue("iphone"))},
		),
	})

	require.NoError(t, err)
	results := response.Data
	require.Equal(t, 2, len(results))
}

func TestServer_Timeseries_1day_no_data(t *testing.T) {
	t.Parallel()
	server, instanceID := getMetricsTestServer(t, "timeseries")

	response, err := server.MetricsViewTimeSeries(testCtx(), &runtimev1.MetricsViewTimeSeriesRequest{
		InstanceId:      instanceID,
		MetricsViewName: "timeseries",
		MeasureNames:    []string{"max_clicks"},
		TimeStart:       parseTimeToProtoTimeStamps(t, "2018-01-01T00:00:00Z"),
		TimeEnd:         parseTimeToProtoTimeStamps(t, "2018-01-03T00:00:00Z"),
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_DAY,
	})

	require.NoError(t, err)
	results := response.Data
	require.Equal(t, 2, len(results))
	require.Equal(t, parseTime(t, "2018-01-01T00:00:00Z"), results[0].Ts.AsTime())
	require.Equal(t, parseTime(t, "2018-01-02T00:00:00Z"), results[1].Ts.AsTime())
}

func TestServer_Timeseries_1day_no_data_no_range(t *testing.T) {
	t.Parallel()
	server, instanceID := getMetricsTestServer(t, "timeseries")

	response, err := server.MetricsViewTimeSeries(testCtx(), &runtimev1.MetricsViewTimeSeriesRequest{
		InstanceId:      instanceID,
		MetricsViewName: "timeseries",
		MeasureNames:    []string{"max_clicks"},
		TimeEnd:         parseTimeToProtoTimeStamps(t, "2018-01-03T00:00:00Z"),
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_DAY,
	})

	require.NoError(t, err)
	results := response.Data
	require.Equal(t, 0, len(results))

	response, err = server.MetricsViewTimeSeries(testCtx(), &runtimev1.MetricsViewTimeSeriesRequest{
		InstanceId:      instanceID,
		MetricsViewName: "timeseries",
		MeasureNames:    []string{"max_clicks"},
		TimeStart:       parseTimeToProtoTimeStamps(t, "2022-01-01T00:00:00Z"),
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_DAY,
	})

	require.NoError(t, err)
	results = response.Data
	require.Equal(t, 0, len(results))
}

func TestServer_Timeseries_1day_Count(t *testing.T) {
	t.Parallel()
	server, instanceID := getMetricsTestServer(t, "timeseries")

	response, err := server.MetricsViewTimeSeries(testCtx(), &runtimev1.MetricsViewTimeSeriesRequest{
		InstanceId:      instanceID,
		MetricsViewName: "timeseries",
		MeasureNames:    []string{"count"},
		TimeStart:       parseTimeToProtoTimeStamps(t, "2019-01-01T00:00:00Z"),
		TimeEnd:         parseTimeToProtoTimeStamps(t, "2019-01-03T00:00:00Z"),
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_DAY,
		Where: expressionpb.In(
			expressionpb.Identifier("device"),
			[]*runtimev1.Expression{expressionpb.Value(structpb.NewStringValue("android")), expressionpb.Value(structpb.NewStringValue("iphone"))},
		),
	})

	require.NoError(t, err)
	results := response.Data
	require.Equal(t, 2, len(results))
	require.Equal(t, 1.0, results[0].Records.Fields["count"].GetNumberValue())
}

func TestServer_MetricsViewTimeseries_export_xlsx(t *testing.T) {
	t.Parallel()
	rt, instanceId := testruntime.NewInstanceForProject(t, "ad_bids_2rows")

	ctx := testCtx()

	q := &queries.MetricsViewTimeSeries{
		MetricsViewName: "ad_bids_metrics",
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_DAY,
		MeasureNames:    []string{"measure_0"},
		SecurityClaims:  testClaims(),
	}

	var buf bytes.Buffer

	err := q.Export(ctx, rt, instanceId, &buf, &runtime.ExportOptions{
		Format: runtimev1.ExportFormat_EXPORT_FORMAT_XLSX,
	})
	require.NoError(t, err)

	file, err := excelize.OpenReader(&buf)
	rows, err := file.GetRows("Sheet1")
	require.NoError(t, err)

	require.Equal(t, 3, len(rows))
	require.Equal(t, 2, len(rows[0]))
	require.Equal(t, 2, len(rows[1]))
	require.Equal(t, 2, len(rows[2]))
}

func TestServer_MetricsViewTimeseries_export_csv(t *testing.T) {
	t.Parallel()
	rt, instanceId := testruntime.NewInstanceForProject(t, "ad_bids_2rows")

	ctx := testCtx()

	q := &queries.MetricsViewTimeSeries{
		MetricsViewName: "ad_bids_metrics",
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_DAY,
		MeasureNames:    []string{"measure_0"},
		SecurityClaims:  testClaims(),
	}

	var buf bytes.Buffer

	err := q.Export(ctx, rt, instanceId, &buf, &runtime.ExportOptions{
		Format: runtimev1.ExportFormat_EXPORT_FORMAT_CSV,
	})
	require.NoError(t, err)

	require.Equal(t, 3, strings.Count(buf.String(), "\n"))
}

func resolveMVAndSecurity(t *testing.T, rt *runtime.Runtime, instanceID, metricsViewName string) (*runtimev1.MetricsViewSpec, *runtime.ResolvedSecurity) {
	ctx := testCtx()

	ctrl, err := rt.Controller(ctx, instanceID)
	require.NoError(t, err)

	res, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: metricsViewName}, false)
	require.NoError(t, err)

	mvRes := res.GetMetricsView()
	mv := mvRes.State.ValidSpec
	require.NoError(t, err)

	resolvedSecurity, err := rt.ResolveSecurity(ctx, instanceID, auth.GetClaims(ctx, instanceID), res)
	require.NoError(t, err)

	return mv, resolvedSecurity
}

func testClaims() *runtime.SecurityClaims {
	return &runtime.SecurityClaims{SkipChecks: true}
}
