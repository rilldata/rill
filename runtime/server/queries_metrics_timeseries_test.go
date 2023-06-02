package server

import (
	"fmt"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
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
		Filter: &runtimev1.MetricsViewFilter{
			Include: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "dom",
					In: []*structpb.Value{
						structpb.NewStringValue("msn.com"),
					},
					Like: []string{"%yahoo%"},
				},
			},
		},
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
		Filter: &runtimev1.MetricsViewFilter{
			Include: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "device",
					In:   []*structpb.Value{structpb.NewStringValue("android"), structpb.NewStringValue("iphone")},
				},
			},
		},
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
		Filter: &runtimev1.MetricsViewFilter{
			Exclude: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "latitude",
					In:   []*structpb.Value{structpb.NewNumberValue(25)},
				},
			},
		},
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
		Filter: &runtimev1.MetricsViewFilter{
			Exclude: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "latitude",
					In:   []*structpb.Value{structpb.NewNumberValue(25), structpb.NewNullValue()},
				},
			},
		},
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
		Filter: &runtimev1.MetricsViewFilter{
			Exclude: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "country",
					In:   []*structpb.Value{structpb.NewStringValue("Canada")},
				},
			},
		},
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
		Filter: &runtimev1.MetricsViewFilter{
			Exclude: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "country",
					In:   []*structpb.Value{structpb.NewStringValue("Canada"), structpb.NewNullValue()},
				},
			},
		},
	})

	require.NoError(t, err)
	results := response.Data
	require.Equal(t, 1, len(results))
	require.Equal(t, 0.0, results[0].Records.Fields["Total impressions"].GetNumberValue())
}

func TestServer_Timeseries_exclude_notnull_like(t *testing.T) {
	t.Parallel()
	server, instanceID := getMetricsTestServer(t, "timeseries")

	response, err := server.MetricsViewTimeSeries(testCtx(), &runtimev1.MetricsViewTimeSeriesRequest{
		InstanceId:      instanceID,
		MetricsViewName: "timeseries",
		MeasureNames:    []string{"count"},
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
		Filter: &runtimev1.MetricsViewFilter{
			Exclude: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "device",
					Like: []string{"iphone"},
				},
			},
		},
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
		Filter: &runtimev1.MetricsViewFilter{
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
	results := response.Data
	require.Equal(t, 1, len(results))
	require.Equal(t, 0.0, results[0].Records.Fields["sum_imps"].GetNumberValue())
}

func TestServer_Timeseries_numeric_dim(t *testing.T) {
	t.Parallel()
	server, instanceID := getMetricsTestServer(t, "timeseries")

	response, err := server.MetricsViewTimeSeries(testCtx(), &runtimev1.MetricsViewTimeSeriesRequest{
		InstanceId:      instanceID,
		MetricsViewName: "timeseries",
		MeasureNames:    []string{"count"},
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
		Filter: &runtimev1.MetricsViewFilter{
			Include: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "latitude",
					In:   []*structpb.Value{structpb.NewNumberValue(25)},
				},
			},
		},
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
		Filter: &runtimev1.MetricsViewFilter{
			Include: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "latitude",
					In:   []*structpb.Value{structpb.NewNumberValue(25), structpb.NewNumberValue(35)},
				},
			},
		},
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
		Filter: &runtimev1.MetricsViewFilter{
			Include: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "latitude",
					In:   []*structpb.Value{structpb.NewNumberValue(25), structpb.NewNullValue()},
				},
			},
		},
	})

	require.NoError(t, err)
	results := response.Data
	require.Equal(t, 1, len(results))
	require.Equal(t, 2.0, results[0].Records.Fields["count"].GetNumberValue())
}

func TestServer_Timeseries_Empty_TimeRange(t *testing.T) {
	t.Parallel()
	server, instanceID := getMetricsTestServer(t, "timeseries")

	response, err := server.MetricsViewTimeSeries(testCtx(), &runtimev1.MetricsViewTimeSeriesRequest{
		InstanceId:      instanceID,
		MetricsViewName: "timeseries",
		MeasureNames:    []string{"max_clicks"},
		Filter: &runtimev1.MetricsViewFilter{
			Include: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "device",
					In:   []*structpb.Value{structpb.NewStringValue("android"), structpb.NewStringValue("iphone")},
				},
			},
		},
	})

	require.NoError(t, err)
	results := response.Data
	for i, v := range response.Data {
		fmt.Printf("i: %d, ts: %v\n", i, v.Ts.AsTime())
	}
	require.Equal(t, 25, len(results))
	require.Equal(t, 1.0, results[0].Records.Fields["max_clicks"].GetNumberValue())
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
		Filter: &runtimev1.MetricsViewFilter{
			Include: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "device",
					In:   []*structpb.Value{structpb.NewStringValue("android"), structpb.NewStringValue("iphone")},
				},
			},
		},
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
		Filter: &runtimev1.MetricsViewFilter{
			Include: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "device",
					In:   []*structpb.Value{structpb.NewStringValue("android")},
				},
			},
		},
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
		Filter: &runtimev1.MetricsViewFilter{
			Include: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "publisher",
					In:   []*structpb.Value{structpb.NewNullValue()},
				},
			},
		},
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
		Filter: &runtimev1.MetricsViewFilter{
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
		Filter: &runtimev1.MetricsViewFilter{
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
		Filter: &runtimev1.MetricsViewFilter{
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
		Filter: &runtimev1.MetricsViewFilter{
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
	results := response.Data
	require.Equal(t, 1, len(results))
	require.Equal(t, 0.0, results[0].Records.Fields["sum_clicks"].GetNumberValue())
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
		Filter: &runtimev1.MetricsViewFilter{
			Include: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "device",
					In:   []*structpb.Value{structpb.NewStringValue("android"), structpb.NewStringValue("iphone")},
				},
			},
		},
	})

	require.NoError(t, err)
	results := response.Data
	require.Equal(t, 2, len(results))
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
		Filter: &runtimev1.MetricsViewFilter{
			Include: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "device",
					In:   []*structpb.Value{structpb.NewStringValue("android"), structpb.NewStringValue("iphone")},
				},
			},
		},
	})

	require.NoError(t, err)
	results := response.Data
	require.Equal(t, 2, len(results))
	require.Equal(t, 1.0, results[0].Records.Fields["count"].GetNumberValue())
}
