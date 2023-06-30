package server

import (
	"bytes"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/queries"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
)

func TestServer_MetricsViewRows(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewRows(testCtx(), &runtimev1.MetricsViewRowsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
	})
	require.NoError(t, err)
	require.Equal(t, 2, len(tr.Data))
	require.Equal(t, 11, len(tr.Meta))

}

func TestServer_MetricsViewRows_Granularity(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewRows(testCtx(), &runtimev1.MetricsViewRowsRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_DAY,
	})
	require.NoError(t, err)
	require.Equal(t, 2, len(tr.Data))
	require.Equal(t, 12, len(tr.Meta))
	require.Equal(t, "timestamp__day", tr.Meta[0].Name)
}

func TestServer_MetricsViewRows_export(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	ctx := testCtx()
	q := &queries.MetricsViewRows{
		MetricsViewName: "ad_bids_metrics",
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_DAY,
	}

	var buf bytes.Buffer

	err := q.Export(ctx, server.runtime, instanceId, 0, runtimev1.ExportFormat_EXPORT_FORMAT_XLSX, &buf)
	require.NoError(t, err)

	file, err := excelize.OpenReader(&buf)
	rows, err := file.GetRows("Sheet1")
	require.NoError(t, err)

	require.Equal(t, 3, len(rows))
	require.Equal(t, []string{"timestamp", "publisher", "domain", "bid_price", "volume", "impressions", "ad words", "clicks", "numeric_dim", "device"}, rows[0][2:])
	require.Equal(t, []string{"2022-01-01T14:49:50.459Z", "", "msn.com", "2", "4", "2", "cars", "", "1", "iphone"}, rows[1][2:])
	require.Equal(t, []string{"2022-01-02T11:58:12.475Z", "Yahoo", "yahoo.com", "2", "4", "1", "cars", "1", "1"}, rows[2][2:])
}
