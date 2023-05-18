package server

import (
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestServer_MetricsViewTimeSeries(t *testing.T) {
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

	require.Equal(t, parseTime(t, "2022-01-01T00:00:00Z"), tr.Data[0].Ts)
	require.Equal(t, 1.0, tr.Data[0].Records.Fields["measure_0"].GetNumberValue())
	require.Equal(t, 2.0, tr.Data[0].Records.Fields["measure_2"].GetNumberValue())

	require.Equal(t, parseTime(t, "2022-01-02T00:00:00Z"), tr.Data[1].Ts)
	require.Equal(t, 1.0, tr.Data[1].Records.Fields["measure_0"].GetNumberValue())
	require.Equal(t, 1.0, tr.Data[1].Records.Fields["measure_2"].GetNumberValue())
}

func TestServer_MetricsViewTimeSeries_TimeEnd_exclusive(t *testing.T) {
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewTimeSeries(testCtx(), &runtimev1.MetricsViewTimeSeriesRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_DAY,
		TimeStart:       parseTime(t, "2022-01-01T00:00:00Z"),
		TimeEnd:         parseTime(t, "2022-01-02T00:00:00Z"),
		MeasureNames:    []string{"measure_0", "measure_2"},
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Data))
	require.Equal(t, 2, len(tr.Data[0].Records.Fields))

	require.Equal(t, parseTime(t, "2022-01-01T00:00:00Z"), tr.Data[0].Ts)
	require.Equal(t, 1.0, tr.Data[0].Records.Fields["measure_0"].GetNumberValue())
	require.Equal(t, 2.0, tr.Data[0].Records.Fields["measure_2"].GetNumberValue())
}

func TestServer_MetricsViewTimeSeries_complete_source_sanity_test(t *testing.T) {
	server, instanceId := getMetricsTestServer(t, "ad_bids")

	tr, err := server.MetricsViewTimeSeries(testCtx(), &runtimev1.MetricsViewTimeSeriesRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_DAY,
		MeasureNames:    []string{"measure_0", "measure_1"},
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
		},
	})
	require.NoError(t, err)
	require.True(t, len(tr.Data) > 0)
	require.Equal(t, 2, len(tr.Data[0].Records.Fields))
	require.NotEmpty(t, tr.Data[0].Ts)
	require.True(t, tr.Data[0].Records.Fields["measure_0"].GetNumberValue() > 0)
	require.True(t, tr.Data[0].Records.Fields["measure_1"].GetNumberValue() > 0)
}
