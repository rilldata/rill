package server

import (
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestServer_MetricsViewTotals(t *testing.T) {
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewTotals(testCtx(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"measure_0"},
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Data.Fields))
	require.Equal(t, 2.0, tr.Data.Fields["measure_0"].GetNumberValue())
}

func TestServer_MetricsViewTotals_row_null_exclude(t *testing.T) {
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewTotals(testCtx(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"measure_0"},
		Filter: &runtimev1.MetricsViewFilter{
			Exclude: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "device",
					In: []*structpb.Value{
						structpb.NewStringValue("iphone"),
					},
				},
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Data.Fields))
	require.Equal(t, 1.0, tr.Data.Fields["measure_0"].GetNumberValue())
}

func TestServer_MetricsViewTotals_row_null_exclude_null(t *testing.T) {
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewTotals(testCtx(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"measure_0"},
		Filter: &runtimev1.MetricsViewFilter{
			Exclude: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "device",
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

func TestServer_MetricsViewTotals_row_null_exclude_all(t *testing.T) {
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewTotals(testCtx(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"measure_0"},
		Filter: &runtimev1.MetricsViewFilter{
			Exclude: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "device",
					In: []*structpb.Value{
						structpb.NewNullValue(),
						structpb.NewStringValue("iphone"),
					},
				},
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Data.Fields))
	require.Equal(t, 0.0, tr.Data.Fields["measure_0"].GetNumberValue())
}

func TestServer_MetricsViewTotals_row_null_include(t *testing.T) {
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewTotals(testCtx(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"measure_0"},
		Filter: &runtimev1.MetricsViewFilter{
			Include: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "device",
					In: []*structpb.Value{
						structpb.NewStringValue("iphone"),
					},
				},
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Data.Fields))
	require.Equal(t, 1.0, tr.Data.Fields["measure_0"].GetNumberValue())
}

func TestServer_MetricsViewTotals_row_null_include_null(t *testing.T) {
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewTotals(testCtx(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"measure_0"},
		Filter: &runtimev1.MetricsViewFilter{
			Include: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "device",
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

func TestServer_MetricsViewTotals_row_null_include_all(t *testing.T) {
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewTotals(testCtx(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"measure_0"},
		Filter: &runtimev1.MetricsViewFilter{
			Include: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "device",
					In: []*structpb.Value{
						structpb.NewNullValue(),
						structpb.NewStringValue("iphone"),
					},
				},
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Data.Fields))
	require.Equal(t, 2.0, tr.Data.Fields["measure_0"].GetNumberValue())
}

func TestServer_MetricsViewTotals_row_null_exclude_like(t *testing.T) {
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewTotals(testCtx(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"measure_0"},
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
	require.Equal(t, 1, len(tr.Data.Fields))
	require.Equal(t, 1.0, tr.Data.Fields["measure_0"].GetNumberValue())
}

func TestServer_MetricsViewTotals_row_null_exclude_like_and_null(t *testing.T) {
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewTotals(testCtx(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"measure_0"},
		Filter: &runtimev1.MetricsViewFilter{
			Exclude: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "device",
					In: []*structpb.Value{
						structpb.NewNullValue(),
					},
					Like: []string{"iphone"},
				},
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Data.Fields))
	require.Equal(t, 0.0, tr.Data.Fields["measure_0"].GetNumberValue())
}

func TestServer_MetricsViewTotals_row_null_exclude_like_doesntexist(t *testing.T) {
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewTotals(testCtx(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"measure_0"},
		Filter: &runtimev1.MetricsViewFilter{
			Exclude: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "device",
					Like: []string{"doesntexist"},
				},
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Data.Fields))
	require.Equal(t, 2.0, tr.Data.Fields["measure_0"].GetNumberValue())
}

func TestServer_MetricsViewTotals_timestamp_name_with_spaces(t *testing.T) {
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewTotals(testCtx(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics_garbled",
		MeasureNames:    []string{"measure_0"},
		TimeEnd:         parseTime(t, "2022-01-02T00:00:00Z"),
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Data.Fields))
	require.Equal(t, 1.0, tr.Data.Fields["measure_0"].GetNumberValue())
}

func TestServer_MetricsViewTotals_EmptyModel(t *testing.T) {
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewTotals(testCtx(), &runtimev1.MetricsViewTotalsRequest{
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
	tr, err := server.MetricsViewTotals(testCtx(), &runtimev1.MetricsViewTotalsRequest{
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

	tr, err := server.MetricsViewTotals(testCtx(), &runtimev1.MetricsViewTotalsRequest{
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

	tr, err := server.MetricsViewTotals(testCtx(), &runtimev1.MetricsViewTotalsRequest{
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

	tr, err := server.MetricsViewTotals(testCtx(), &runtimev1.MetricsViewTotalsRequest{
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

	tr, err := server.MetricsViewTotals(testCtx(), &runtimev1.MetricsViewTotalsRequest{
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

func TestServer_MetricsViewTotals_1dim_special_symbol_values(t *testing.T) {
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewTotals(testCtx(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"measure_0"},
		Filter: &runtimev1.MetricsViewFilter{
			Include: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "domain",
					In: []*structpb.Value{
						structpb.NewStringValue("msn.'com"), structpb.NewStringValue("msn.\"com"), structpb.NewStringValue("msn. com"),
					},
				},
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Data.Fields))
	require.Equal(t, 0.0, tr.Data.Fields["measure_0"].GetNumberValue())
}

func TestServer_MetricsViewTotals_1dim_2In(t *testing.T) {
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewTotals(testCtx(), &runtimev1.MetricsViewTotalsRequest{
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

	tr, err := server.MetricsViewTotals(testCtx(), &runtimev1.MetricsViewTotalsRequest{
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

	tr, err := server.MetricsViewTotals(testCtx(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"measure_0"},
		Filter: &runtimev1.MetricsViewFilter{
			Include: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "domain",
					Like: []string{"%com"},
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

	tr, err := server.MetricsViewTotals(testCtx(), &runtimev1.MetricsViewTotalsRequest{
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
					Like: []string{"%com"},
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

	tr, err := server.MetricsViewTotals(testCtx(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"measure_0"},
		Filter: &runtimev1.MetricsViewFilter{
			Include: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "domain",
					Like: []string{"msn%", "y%"},
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

	tr, err := server.MetricsViewTotals(testCtx(), &runtimev1.MetricsViewTotalsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"measure_0"},
		Filter: &runtimev1.MetricsViewFilter{
			Include: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "domain",
					Like: []string{"%com"},
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

	tr, err := server.MetricsViewTotals(testCtx(), &runtimev1.MetricsViewTotalsRequest{
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

	tr, err := server.MetricsViewTotals(testCtx(), &runtimev1.MetricsViewTotalsRequest{
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
					Like: []string{"%yahoo%"},
				},
			},
			Exclude: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "publisher",
					In: []*structpb.Value{
						structpb.NewNullValue(),
					},
					Like: []string{"Y%"},
				},
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Data.Fields))
	require.Equal(t, 0.0, tr.Data.Fields["measure_0"].GetNumberValue())
}
