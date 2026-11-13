package server_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/ratelimit"
	"github.com/rilldata/rill/runtime/server"
	"github.com/rilldata/rill/runtime/testruntime"
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

// A metrics view without data has no min, max or watermark time.
// Ranges anchored at latest or watermark are computed from the zero time and land in year 0;
// they must be clamped to 0001-01-01 instead of failing to marshal.
func TestServer_MetricsViewTimeRanges_NoData(t *testing.T) {
	t.Parallel()
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "",
			"models/empty.sql": `
SELECT TIMESTAMP '2025-01-01 00:00:00' AS event_time, 'a' AS country WHERE 1 = 0
`,
			"metrics/empty_metrics.yaml": `
type: metrics_view
model: empty
timeseries: event_time
dimensions:
- column: country
measures:
- name: count
  expression: COUNT(*)
`,
		},
	})
	server, err := server.NewServer(context.Background(), &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	res, err := server.MetricsViewTimeRanges(testCtx(), &runtimev1.MetricsViewTimeRangesRequest{
		InstanceId:      instanceID,
		MetricsViewName: "empty_metrics",
		Expressions:     []string{"P30D", "7D as of latest/D+1D", "rill-PQC", "inf"},
	})
	require.NoError(t, err)

	// The response must be marshalable, which is where out-of-range timestamps used to fail.
	_, err = protojson.Marshal(res)
	require.NoError(t, err)

	require.Nil(t, res.FullTimeRange.Min)
	require.Nil(t, res.FullTimeRange.Max)
	require.Nil(t, res.FullTimeRange.Watermark)

	require.Len(t, res.ResolvedTimeRanges, 4)
	for _, tr := range res.ResolvedTimeRanges {
		require.NoError(t, tr.Start.CheckValid(), tr.Expression)
		require.NoError(t, tr.End.CheckValid(), tr.Expression)
		require.Equal(t, time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC), tr.Start.AsTime(), tr.Expression)
	}
}

// Placeholder data sits exactly at 0001-01-01T00:00:00Z, which is also Go's zero time.
// The time range resolver treats a zero min time as no data, so such a metrics view currently resolves like an empty one;
// either way, ranges that reach back from it land in year 0 and must be clamped instead of failing to marshal.
func TestServer_MetricsViewTimeRanges_PlaceholderDates(t *testing.T) {
	t.Parallel()
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "",
			"models/placeholder_dates.sql": `
SELECT TIMESTAMP '0001-01-01 00:00:00' AS event_time, 'a' AS country
UNION ALL
SELECT TIMESTAMP '0001-01-05 00:00:00' AS event_time, 'a' AS country
`,
			"metrics/placeholder_metrics.yaml": `
type: metrics_view
model: placeholder_dates
timeseries: event_time
dimensions:
- column: country
measures:
- name: count
  expression: COUNT(*)
`,
		},
	})
	server, err := server.NewServer(context.Background(), &server.Options{}, rt, zap.NewNop(), ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	res, err := server.MetricsViewTimeRanges(testCtx(), &runtimev1.MetricsViewTimeRangesRequest{
		InstanceId:      instanceID,
		MetricsViewName: "placeholder_metrics",
		Expressions:     []string{"P30D", "P12M", "7D as of latest/D+1D"},
		TimeZone:        "Asia/Kolkata",
	})
	require.NoError(t, err)

	// The response must be marshalable, which is where out-of-range timestamps used to fail.
	_, err = protojson.Marshal(res)
	require.NoError(t, err)

	require.Len(t, res.ResolvedTimeRanges, 3)
	for _, tr := range res.ResolvedTimeRanges {
		require.NoError(t, tr.Start.CheckValid(), tr.Expression)
		require.NoError(t, tr.End.CheckValid(), tr.Expression)
		require.Equal(t, time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC), tr.Start.AsTime(), tr.Expression)
		require.True(t, tr.End.AsTime().After(tr.Start.AsTime()), tr.Expression)
	}
}

// Expressions that resolve outside the range of a protobuf timestamp are clamped instead of failing the response.
func TestServer_MetricsViewTimeRanges_ClampsToProtoTimestampRange(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	res, err := server.MetricsViewTimeRanges(testCtx(), &runtimev1.MetricsViewTimeRangesRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		Expressions:     []string{"-1Y to ref as of 0001-06-01", "ref to +1Y as of 9999-06-01"},
	})
	require.NoError(t, err)

	_, err = protojson.Marshal(res)
	require.NoError(t, err)

	require.Len(t, res.ResolvedTimeRanges, 2)
	require.Equal(t, time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC), res.ResolvedTimeRanges[0].Start.AsTime())
	require.Equal(t, time.Date(1, time.June, 1, 0, 0, 0, 0, time.UTC), res.ResolvedTimeRanges[0].End.AsTime())
	require.Equal(t, time.Date(9999, time.June, 1, 0, 0, 0, 0, time.UTC), res.ResolvedTimeRanges[1].Start.AsTime())
	require.Equal(t, time.Date(9999, time.December, 31, 23, 59, 59, 999999999, time.UTC), res.ResolvedTimeRanges[1].End.AsTime())
}
