package server

import (
	"context"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

func getMetricsTestServer(t *testing.T, projectName string) (*Server, string) {
	rt, instanceID := testruntime.NewInstanceForProject(t, projectName)

	server, err := NewServer(&Options{}, rt, nil)
	require.NoError(t, err)

	return server, instanceID
}

func TestServer_LookupMetricsView(t *testing.T) {
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	mv, err := server.lookupMetricsView(context.Background(), instanceId, "ad_bids_metrics")
	require.NoError(t, err)
	require.Equal(t, 4, len(mv.Measures))
	require.Equal(t, 3, len(mv.Dimensions))
}

func TestServer_MetricsViewTotals(t *testing.T) {
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewTotals(context.Background(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"measure_0"},
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Data.Fields))
	require.Equal(t, 2.0, tr.Data.Fields["measure_0"].GetNumberValue())
}

func TestServer_MetricsViewTotals_EmptyModel(t *testing.T) {
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewTotals(context.Background(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "no_rows_metrics",
		MeasureNames:    []string{"measure_0", "measure_1"},
	})
	require.NoError(t, err)
	require.Equal(t, 2, len(tr.Data.Fields))
	require.Equal(t, 0.0, tr.Data.Fields["measure_0"].GetNumberValue())
	require.Equal(t, 0.0, tr.Data.Fields["measure_2"].GetNumberValue())
}

func TestServer_MetricsViewTotals_2measures(t *testing.T) {
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")
	tr, err := server.MetricsViewTotals(context.Background(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"measure_0", "measure_1"},
	})
	require.NoError(t, err)
	require.Equal(t, 2, len(tr.Data.Fields))
	require.Equal(t, 2.0, tr.Data.Fields["measure_0"].GetNumberValue())
	require.Equal(t, 8.0, tr.Data.Fields["measure_1"].GetNumberValue())
}

func TestServer_MetricsViewTotals_TimeStart(t *testing.T) {
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewTotals(context.Background(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"measure_0"},
		TimeStart:       parseTime(t, "2022-01-02T00:00:00Z"),
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Data.Fields))
	require.Equal(t, 1.0, tr.Data.Fields["measure_0"].GetNumberValue())
}

func TestServer_MetricsViewTotals_TimeEnd(t *testing.T) {
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewTotals(context.Background(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"measure_0"},
		TimeEnd:         parseTime(t, "2022-01-02T00:00:00Z"),
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Data.Fields))
	require.Equal(t, 1.0, tr.Data.Fields["measure_0"].GetNumberValue())
}

func TestServer_MetricsViewTotals_TimeStart_TimeEnd(t *testing.T) {
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewTotals(context.Background(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"measure_0"},
		TimeStart:       parseTime(t, "2022-01-01T00:00:00Z"),
		TimeEnd:         parseTime(t, "2022-01-02T00:00:00Z"),
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Data.Fields))
	require.Equal(t, 1.0, tr.Data.Fields["measure_0"].GetNumberValue())
}

func TestServer_MetricsViewTotals_1dim(t *testing.T) {
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewTotals(context.Background(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"measure_0"},
		Filter: &runtimev1.MetricsViewFilter{
			Include: []*runtimev1.MetricsViewFilter_Cond{
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
	require.Equal(t, 1, len(tr.Data.Fields))
	require.Equal(t, 1.0, tr.Data.Fields["measure_0"].GetNumberValue())
}

func TestServer_MetricsViewTotals_1dim_2In(t *testing.T) {
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewTotals(context.Background(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"measure_0"},
		Filter: &runtimev1.MetricsViewFilter{
			Include: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "domain",
					In: []*structpb.Value{
						structpb.NewStringValue("msn.com"),
						structpb.NewStringValue("yahoo.com"),
					},
				},
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Data.Fields))
	require.Equal(t, 2.0, tr.Data.Fields["measure_0"].GetNumberValue())
}

func TestServer_MetricsViewTotals_2dim(t *testing.T) {
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewTotals(context.Background(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"measure_0"},
		Filter: &runtimev1.MetricsViewFilter{
			Include: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "domain",
					In: []*structpb.Value{
						structpb.NewStringValue("yahoo.com"),
					},
				},
				{
					Name: "publisher",
					In: []*structpb.Value{
						structpb.NewStringValue("Yahoo"),
					},
				},
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Data.Fields))
	require.Equal(t, 1.0, tr.Data.Fields["measure_0"].GetNumberValue())
}

func TestServer_MetricsViewTotals_1dim_like(t *testing.T) {
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewTotals(context.Background(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"measure_0"},
		Filter: &runtimev1.MetricsViewFilter{
			Include: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "domain",
					Like: []*structpb.Value{
						structpb.NewStringValue("%com"),
					},
				},
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Data.Fields))
	require.Equal(t, 2.0, tr.Data.Fields["measure_0"].GetNumberValue())
}

func TestServer_MetricsViewTotals_1dim_in_and_like(t *testing.T) {
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewTotals(context.Background(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"measure_0"},
		Filter: &runtimev1.MetricsViewFilter{
			Include: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "domain",
					In: []*structpb.Value{
						structpb.NewStringValue("yahoo"),
					},
					Like: []*structpb.Value{
						structpb.NewStringValue("%com"),
					},
				},
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Data.Fields))
	require.Equal(t, 2.0, tr.Data.Fields["measure_0"].GetNumberValue())
}

func TestServer_MetricsViewTotals_1dim_2like(t *testing.T) {
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewTotals(context.Background(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"measure_0"},
		Filter: &runtimev1.MetricsViewFilter{
			Include: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "domain",
					Like: []*structpb.Value{
						structpb.NewStringValue("msn%"),
						structpb.NewStringValue("y%"),
					},
				},
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Data.Fields))
	require.Equal(t, 2.0, tr.Data.Fields["measure_0"].GetNumberValue())
}

func TestServer_MetricsViewTotals_1dim_include_and_exclude(t *testing.T) {
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewTotals(context.Background(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"measure_0"},
		Filter: &runtimev1.MetricsViewFilter{
			Include: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "domain",
					Like: []*structpb.Value{
						structpb.NewStringValue("%com"),
					},
				},
			},
			Exclude: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "domain",
					In: []*structpb.Value{
						structpb.NewStringValue("yahoo.com"),
					},
				},
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Data.Fields))
	require.Equal(t, 1.0, tr.Data.Fields["measure_0"].GetNumberValue())
}

func TestServer_MetricsViewTotals_1dim_null(t *testing.T) {
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewTotals(context.Background(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"measure_0"},
		Filter: &runtimev1.MetricsViewFilter{
			Include: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "publisher",
					In: []*structpb.Value{
						structpb.NewNullValue(),
					},
				},
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Data.Fields))
	require.Equal(t, 1.0, tr.Data.Fields["measure_0"].GetNumberValue())
}

func TestServer_MetricsViewTotals_1dim_include_and_exclude_in_and_like(t *testing.T) {
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewTotals(context.Background(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"measure_0"},
		Filter: &runtimev1.MetricsViewFilter{
			Include: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "domain",
					In: []*structpb.Value{
						structpb.NewStringValue("msn.com"),
					},
					Like: []*structpb.Value{
						structpb.NewStringValue("%yahoo%"),
					},
				},
			},
			Exclude: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "publisher",
					In: []*structpb.Value{
						structpb.NewNullValue(),
					},
					Like: []*structpb.Value{
						structpb.NewStringValue("Y%"),
					},
				},
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Data.Fields))
	require.Equal(t, 0.0, tr.Data.Fields["measure_0"].GetNumberValue())
}

func TestServer_MetricsViewToplist(t *testing.T) {
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewToplist(context.Background(), &runtimev1.MetricsViewToplistRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		DimensionName:   "domain",
		MeasureNames:    []string{"measure_2"},
		Sort: []*runtimev1.MetricsViewSort{
			{
				Name: "measure_2",
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, 2, len(tr.Data))

	require.Equal(t, 2, len(tr.Data[0].Fields))
	require.Equal(t, 2, len(tr.Data[1].Fields))

	require.Equal(t, "msn.com", tr.Data[0].Fields["domain"].GetStringValue())
	require.Equal(t, 2.0, tr.Data[0].Fields["measure_2"].GetNumberValue())

	require.Equal(t, "yahoo.com", tr.Data[1].Fields["domain"].GetStringValue())
	require.Equal(t, 1.0, tr.Data[1].Fields["measure_2"].GetNumberValue())
}

func TestServer_MetricsViewToplist_quotes(t *testing.T) {
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewToplist(context.Background(), &runtimev1.MetricsViewToplistRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		DimensionName:   "ad words",
		MeasureNames:    []string{"measure_0"},
		Sort: []*runtimev1.MetricsViewSort{
			{
				Name: "measure_0",
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Data))

	require.Equal(t, 2, len(tr.Data[0].Fields))

	require.Equal(t, "cars", tr.Data[0].Fields["ad words"].GetStringValue())
	require.Equal(t, 2.0, tr.Data[0].Fields["measure_0"].GetNumberValue())
}

func Ignore_TestServer_MetricsViewToplist_HugeInt(t *testing.T) {
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewToplist(context.Background(), &runtimev1.MetricsViewToplistRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		DimensionName:   "id",
		MeasureNames:    []string{"measure_2"},
		Sort: []*runtimev1.MetricsViewSort{
			{
				Name: "measure_2",
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, 2, len(tr.Data))

	require.Equal(t, 2, len(tr.Data[0].Fields))
	require.Equal(t, 2, len(tr.Data[1].Fields))

	require.Equal(t, "170141183460469231731687303715884105727", tr.Data[0].Fields["Id"].GetStringValue())
	require.Equal(t, 1.0, tr.Data[0].Fields["measure_2"].GetNumberValue())

	require.Equal(t, "170141183460469231731687303715884105726", tr.Data[1].Fields["Id"].GetStringValue())
	require.Equal(t, 2.0, tr.Data[1].Fields["measure_2"].GetNumberValue())
}

func TestServer_MetricsViewToplist_asc(t *testing.T) {
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewToplist(context.Background(), &runtimev1.MetricsViewToplistRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		DimensionName:   "domain",
		MeasureNames:    []string{"measure_2"},
		Sort: []*runtimev1.MetricsViewSort{
			{
				Name:      "measure_2",
				Ascending: true,
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, 2, len(tr.Data))

	require.Equal(t, 2, len(tr.Data[0].Fields))
	require.Equal(t, 2, len(tr.Data[1].Fields))

	require.Equal(t, "yahoo.com", tr.Data[0].Fields["domain"].GetStringValue())
	require.Equal(t, 1.0, tr.Data[0].Fields["measure_2"].GetNumberValue())

	require.Equal(t, "msn.com", tr.Data[1].Fields["domain"].GetStringValue())
	require.Equal(t, 2.0, tr.Data[1].Fields["measure_2"].GetNumberValue())
}

func TestServer_MetricsViewToplist_nulls_last(t *testing.T) {
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewToplist(context.Background(), &runtimev1.MetricsViewToplistRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		DimensionName:   "domain",
		MeasureNames:    []string{"measure_3"},
		Sort: []*runtimev1.MetricsViewSort{
			{
				Name:      "measure_3",
				Ascending: true,
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, 2, len(tr.Data))

	require.Equal(t, 2, len(tr.Data[0].Fields))
	require.Equal(t, 2, len(tr.Data[1].Fields))

	require.Equal(t, "yahoo.com", tr.Data[0].Fields["domain"].GetStringValue())
	require.Equal(t, 1.0, tr.Data[0].Fields["measure_3"].GetNumberValue())

	require.Equal(t, "msn.com", tr.Data[1].Fields["domain"].GetStringValue())
	require.Equal(t, structpb.NullValue(0), tr.Data[1].Fields["measure_3"].GetNullValue())

	tr, err = server.MetricsViewToplist(context.Background(), &runtimev1.MetricsViewToplistRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		DimensionName:   "domain",
		MeasureNames:    []string{"measure_3"},
		Sort: []*runtimev1.MetricsViewSort{
			{
				Name:      "measure_3",
				Ascending: false,
			},
		},
	})

	require.NoError(t, err)
	require.Equal(t, 2, len(tr.Data))

	require.Equal(t, 2, len(tr.Data[0].Fields))
	require.Equal(t, 2, len(tr.Data[1].Fields))

	require.Equal(t, "yahoo.com", tr.Data[0].Fields["domain"].GetStringValue())
	require.Equal(t, 1.0, tr.Data[0].Fields["measure_3"].GetNumberValue())

	require.Equal(t, "msn.com", tr.Data[1].Fields["domain"].GetStringValue())
	require.Equal(t, structpb.NullValue(0), tr.Data[1].Fields["measure_3"].GetNullValue())
}

func TestServer_MetricsViewToplist_asc_limit(t *testing.T) {
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewToplist(context.Background(), &runtimev1.MetricsViewToplistRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		DimensionName:   "domain",
		MeasureNames:    []string{"measure_2"},
		Sort: []*runtimev1.MetricsViewSort{
			{
				Name:      "measure_2",
				Ascending: true,
			},
		},
		Limit: 1,
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Data))
	require.Equal(t, 2, len(tr.Data[0].Fields))

	require.Equal(t, "yahoo.com", tr.Data[0].Fields["domain"].GetStringValue())
	require.Equal(t, 1.0, tr.Data[0].Fields["measure_2"].GetNumberValue())
}

func TestServer_MetricsViewToplist_2measures(t *testing.T) {
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewToplist(context.Background(), &runtimev1.MetricsViewToplistRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		DimensionName:   "domain",
		MeasureNames:    []string{"measure_0", "measure_2"},
		Sort: []*runtimev1.MetricsViewSort{
			{
				Name:      "measure_0",
				Ascending: true,
			},
			{
				Name:      "measure_2",
				Ascending: true,
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, 2, len(tr.Data))
	require.Equal(t, 3, len(tr.Data[0].Fields))

	require.Equal(t, "yahoo.com", tr.Data[0].Fields["domain"].GetStringValue())
	require.Equal(t, 1.0, tr.Data[0].Fields["measure_0"].GetNumberValue())
	require.Equal(t, 1.0, tr.Data[0].Fields["measure_2"].GetNumberValue())

	require.Equal(t, "msn.com", tr.Data[1].Fields["domain"].GetStringValue())
	require.Equal(t, 1.0, tr.Data[1].Fields["measure_0"].GetNumberValue())
	require.Equal(t, 2.0, tr.Data[1].Fields["measure_2"].GetNumberValue())
}

func TestServer_MetricsViewToplist_complete_source_sanity_test(t *testing.T) {
	server, instanceId := getMetricsTestServer(t, "ad_bids")

	tr, err := server.MetricsViewToplist(context.Background(), &runtimev1.MetricsViewToplistRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		DimensionName:   "domain",
		MeasureNames:    []string{"measure_0"},
		Sort: []*runtimev1.MetricsViewSort{
			{
				Name:      "measure_0",
				Ascending: true,
			},
		},
		Filter: &runtimev1.MetricsViewFilter{
			Exclude: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "publisher",
					In: []*structpb.Value{
						structpb.NewStringValue("Yahoo"),
					},
				},
			},
		},
	})
	require.NoError(t, err)
	require.True(t, len(tr.Data) > 1)
	require.Equal(t, 2, len(tr.Data[0].Fields))
}

func TestServer_MetricsViewTimeSeries(t *testing.T) {
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewTimeSeries(context.Background(), &runtimev1.MetricsViewTimeSeriesRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		TimeGranularity: "DAY",
		MeasureNames:    []string{"measure_0", "measure_2"},
	})
	require.NoError(t, err)
	require.Equal(t, 2, len(tr.Data))
	require.Equal(t, 3, len(tr.Data[0].Fields))

	require.Equal(t, "2022-01-01T00:00:00Z", tr.Data[0].Fields["timestamp"].GetStringValue())
	require.Equal(t, 1.0, tr.Data[0].Fields["measure_0"].GetNumberValue())
	require.Equal(t, 2.0, tr.Data[0].Fields["measure_2"].GetNumberValue())

	require.Equal(t, "2022-01-02T00:00:00Z", tr.Data[1].Fields["timestamp"].GetStringValue())
	require.Equal(t, 1.0, tr.Data[1].Fields["measure_0"].GetNumberValue())
	require.Equal(t, 1.0, tr.Data[1].Fields["measure_2"].GetNumberValue())
}

func TestServer_MetricsViewTimeSeries_complete_source_sanity_test(t *testing.T) {
	server, instanceId := getMetricsTestServer(t, "ad_bids")

	tr, err := server.MetricsViewTimeSeries(context.Background(), &runtimev1.MetricsViewTimeSeriesRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		TimeGranularity: "DAY",
		MeasureNames:    []string{"measure_0", "measure_1"},
		Filter: &runtimev1.MetricsViewFilter{
			Include: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "domain",
					In: []*structpb.Value{
						structpb.NewStringValue("msn.com"),
					},
					Like: []*structpb.Value{
						structpb.NewStringValue("%yahoo%"),
					},
				},
			},
		},
	})
	require.NoError(t, err)
	require.True(t, len(tr.Data) > 0)
	require.Equal(t, 3, len(tr.Data[0].Fields))
	require.True(t, tr.Data[0].Fields["measure_0"].GetNumberValue() > 0)
	require.True(t, tr.Data[0].Fields["measure_1"].GetNumberValue() > 0)
}
