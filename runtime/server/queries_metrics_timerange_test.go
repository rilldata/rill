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
