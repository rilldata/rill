package server

import (
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

/*
Source:

|id |timestamp               |publisher|domain   |bid_price|volume|impressions|ad words|clicks|device|
|---|------------------------|---------|---------|---------|------|-----------|--------|------|------|
|0  |2022-01-01T14:49:50.459Z|         |msn.com  |2        |4     |2          |cars    |      |iphone|
|1  |2022-01-02T11:58:12.475Z|Yahoo    |yahoo.com|2        |4     |1          |cars    |1     |      |

Measures:

  - label: "Number of bids"
    expression: count(*)
    description: ""
    format_preset: ""
  - label: "Total volume"
    expression: sum(volume)
    description: ""
    format_preset: ""
  - label: "Total impressions"
    expression: sum(impressions)
  - label: "Total clicks"
    expression: sum(clicks)
*/
func TestServer_MetricsViewToplist_mig(t *testing.T) {
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewComparisonToplist(testCtx(), &runtimev1.MetricsViewComparisonToplistRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		DimensionName:   "domain",
		MeasureNames:    []string{"measure_2"},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				MeasureName: "measure_2",
				Ascending:   false,
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, 2, len(tr.Rows))

	require.Equal(t, 1, len(tr.Rows[0].MeasureValues))
	require.Equal(t, 1, len(tr.Rows[1].MeasureValues))

	require.Equal(t, "msn.com", tr.Rows[0].DimensionValue.GetStringValue())
	require.Equal(t, 2.0, tr.Rows[0].MeasureValues[0].BaseValue.GetNumberValue())

	require.Equal(t, "yahoo.com", tr.Rows[1].DimensionValue.GetStringValue())
	require.Equal(t, 1.0, tr.Rows[1].MeasureValues[0].BaseValue.GetNumberValue())
}

func TestServer_MetricsViewToplist_mig_quotes(t *testing.T) {
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

func TestServer_MetricsViewToplist_mig_numeric_dim(t *testing.T) {
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

func Ignore_TestServer_MetricsViewToplist_mig_HugeInt(t *testing.T) {
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

func TestServer_MetricsViewToplist_mig_asc(t *testing.T) {
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

func TestServer_MetricsViewToplist_mig_nulls_last(t *testing.T) {
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

func TestServer_MetricsViewToplist_mig_asc_limit(t *testing.T) {
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

func TestServer_MetricsViewToplist_mig_2measures(t *testing.T) {
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

func TestServer_MetricsViewToplist_mig_complete_source_sanity_test(t *testing.T) {
	server, instanceId := getMetricsTestServer(t, "ad_bids")

	tr, err := server.MetricsViewToplist(testCtx(), &runtimev1.MetricsViewToplistRequest{
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
