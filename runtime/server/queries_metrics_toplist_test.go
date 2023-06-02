package server

import (
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestServer_MetricsViewToplist(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewToplist(testCtx(), &runtimev1.MetricsViewToplistRequest{
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

func TestServer_MetricsViewToplist_InlineMeasures(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewToplist(testCtx(), &runtimev1.MetricsViewToplistRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		DimensionName:   "domain",
		MeasureNames:    []string{"measure_0", "tmp_measure", "measure_2"},
		InlineMeasures: []*runtimev1.InlineMeasure{
			{
				Name:       "tmp_measure",
				Expression: "count(*)",
			},
		},
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
	require.Equal(t, 4, len(tr.Data[0].Fields))

	require.Equal(t, "yahoo.com", tr.Data[0].Fields["domain"].GetStringValue())
	require.Equal(t, 1.0, tr.Data[0].Fields["measure_0"].GetNumberValue())
	require.Equal(t, 1.0, tr.Data[0].Fields["measure_2"].GetNumberValue())
	require.Equal(t, 1.0, tr.Data[0].Fields["tmp_measure"].GetNumberValue())

	require.Equal(t, "msn.com", tr.Data[1].Fields["domain"].GetStringValue())
	require.Equal(t, 1.0, tr.Data[1].Fields["measure_0"].GetNumberValue())
	require.Equal(t, 2.0, tr.Data[1].Fields["measure_2"].GetNumberValue())
	require.Equal(t, 1.0, tr.Data[1].Fields["tmp_measure"].GetNumberValue())
}

func TestServer_MetricsViewToplist_quotes(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewToplist(testCtx(), &runtimev1.MetricsViewToplistRequest{
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

func TestServer_MetricsViewToplist_numeric_dim(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewToplist(testCtx(), &runtimev1.MetricsViewToplistRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		DimensionName:   "numeric_dim",
		MeasureNames:    []string{"measure_0"},
	})

	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Data))
	require.Equal(t, 2, len(tr.Data[0].Fields))

	require.Equal(t, float64(1), tr.Data[0].Fields["numeric_dim"].GetNumberValue())
	require.Equal(t, 2.0, tr.Data[0].Fields["measure_0"].GetNumberValue())
}

func Ignore_TestServer_MetricsViewToplist_HugeInt(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewToplist(testCtx(), &runtimev1.MetricsViewToplistRequest{
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
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewToplist(testCtx(), &runtimev1.MetricsViewToplistRequest{
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
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewToplist(testCtx(), &runtimev1.MetricsViewToplistRequest{
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

	tr, err = server.MetricsViewToplist(testCtx(), &runtimev1.MetricsViewToplistRequest{
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
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewToplist(testCtx(), &runtimev1.MetricsViewToplistRequest{
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
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewToplist(testCtx(), &runtimev1.MetricsViewToplistRequest{
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
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids")

	tr, err := server.MetricsViewToplist(testCtx(), &runtimev1.MetricsViewToplistRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		DimensionName:   "dom",
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
					Name: "pub",
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

func TestServer_MetricsViewToplist_DimensionsByName(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids")

	tr, err := server.MetricsViewToplist(testCtx(), &runtimev1.MetricsViewToplistRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		DimensionName:   "dom",
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
					Name: "pub",
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
