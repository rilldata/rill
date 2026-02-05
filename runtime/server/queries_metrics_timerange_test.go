package server_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

func TestServer_MetricsViewTimeRange(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	res, err := server.MetricsViewTimeRange(testCtx(), &runtimev1.MetricsViewTimeRangeRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, parseTime(t, "2022-01-01T14:49:50.459Z"), res.TimeRangeSummary.Min.AsTime())
	require.Equal(t, parseTime(t, "2022-01-02T11:58:12.475Z"), res.TimeRangeSummary.Max.AsTime())
}

func TestServer_MetricsViewTimeRangeS(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	res, err := server.MetricsViewTimeRanges(testCtx(), &runtimev1.MetricsViewTimeRangesRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		Expressions:     []string{"5m as of watermark"},
	})
	require.NoError(t, err)
	require.NotNil(t, res)

	require.Equal(t, parseTime(t, "2022-01-01T14:49:50.459Z"), res.FullTimeRange.Min.AsTime())
	require.Equal(t, parseTime(t, "2022-01-02T11:58:12.475Z"), res.FullTimeRange.Max.AsTime())
	require.Equal(t, parseTime(t, "2022-01-02T11:58:12.475Z"), res.FullTimeRange.Watermark.AsTime())

	require.Len(t, res.ResolvedTimeRanges, 1)
	require.Equal(t, parseTime(t, "2022-01-02T11:53:12.475Z"), res.ResolvedTimeRanges[0].Start.AsTime())
	require.Equal(t, parseTime(t, "2022-01-02T11:58:12.475Z"), res.ResolvedTimeRanges[0].End.AsTime())
	require.Equal(t, runtimev1.TimeGrain_TIME_GRAIN_MINUTE, res.ResolvedTimeRanges[0].Grain)
	require.Equal(t, "timestamp", res.ResolvedTimeRanges[0].TimeDimension)
	require.Equal(t, "5m as of watermark", res.ResolvedTimeRanges[0].Expression)
}
