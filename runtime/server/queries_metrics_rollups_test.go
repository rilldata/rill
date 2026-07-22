package server_test

import (
	"context"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/ratelimit"
	"github.com/rilldata/rill/runtime/server"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// getRollupTestServer creates a test server for a project with a metrics view that has a monthly rollup.
// The base table covers Jan-Mar 2024 hourly; the rollup aggregates it by month.
// The `country` dimension is only available in the base table, not in the rollup.
func getRollupTestServer(t *testing.T) (*server.Server, string) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "",
			"models/base_events.sql": `
SELECT
	ts AS timestamp,
	CASE (row_number() OVER ()) % 2 WHEN 0 THEN 'Google' ELSE 'Facebook' END AS publisher,
	'US' AS country,
	10 AS impressions
FROM generate_series(TIMESTAMP '2024-01-01 00:00:00', TIMESTAMP '2024-03-31 23:00:00', INTERVAL '1 HOUR') t(ts)
`,
			"models/rollup_month.sql": `
SELECT date_trunc('month', timestamp) AS timestamp, publisher, SUM(impressions) AS impressions
FROM base_events GROUP BY 1, 2
`,
			"metrics_views/mv.yaml": `
type: metrics_view
version: 1
model: base_events
timeseries: timestamp
dimensions:
  - name: publisher
    column: publisher
  - name: country
    column: country
measures:
  - name: total_impressions
    expression: 'SUM("impressions")'
rollups:
  - model: rollup_month
    time_grain: month
    dimensions:
      - publisher
    measures:
      - total_impressions
explore:
  skip: true
`,
		},
	})
	// 1 project_parser + 2 models + 1 metrics_view = 4 resources; 0 errors
	testruntime.RequireReconcileState(t, rt, instanceID, 4, 0, 0)

	srv, err := server.NewServer(context.Background(), &server.Options{}, rt, nil, ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	return srv, instanceID
}

func TestServer_MetricsViewTimeRange_Rollups(t *testing.T) {
	t.Parallel()
	srv, instanceID := getRollupTestServer(t)

	res, err := srv.MetricsViewTimeRange(testCtx(), &runtimev1.MetricsViewTimeRangeRequest{
		InstanceId:      instanceID,
		MetricsViewName: "mv",
	})
	require.NoError(t, err)
	require.Equal(t, parseTime(t, "2024-01-01T00:00:00Z"), res.TimeRangeSummary.Min.AsTime())
	require.Equal(t, parseTime(t, "2024-03-31T23:00:00Z"), res.TimeRangeSummary.Max.AsTime())

	require.Len(t, res.RollupTimeRanges, 1)
	rollup, ok := res.RollupTimeRanges["rollup_month"]
	require.True(t, ok)
	require.Equal(t, parseTime(t, "2024-01-01T00:00:00Z"), rollup.Min.AsTime())
	require.Equal(t, parseTime(t, "2024-03-01T00:00:00Z"), rollup.Max.AsTime())
}

func TestServer_MetricsViewTimeRanges_Rollups(t *testing.T) {
	t.Parallel()
	srv, instanceID := getRollupTestServer(t)

	res, err := srv.MetricsViewTimeRanges(testCtx(), &runtimev1.MetricsViewTimeRangesRequest{
		InstanceId:      instanceID,
		MetricsViewName: "mv",
		Expressions:     []string{"P1M as of watermark"},
	})
	require.NoError(t, err)
	require.Equal(t, parseTime(t, "2024-01-01T00:00:00Z"), res.FullTimeRange.Min.AsTime())
	require.Equal(t, parseTime(t, "2024-03-31T23:00:00Z"), res.FullTimeRange.Max.AsTime())

	require.Len(t, res.RollupTimeRanges, 1)
	rollup, ok := res.RollupTimeRanges["rollup_month"]
	require.True(t, ok)
	require.Equal(t, parseTime(t, "2024-01-01T00:00:00Z"), rollup.Min.AsTime())
	require.Equal(t, parseTime(t, "2024-03-01T00:00:00Z"), rollup.Max.AsTime())
}

func TestServer_MetricsViewAggregation_ServingTable(t *testing.T) {
	t.Parallel()
	srv, instanceID := getRollupTestServer(t)

	timeRange := &runtimev1.TimeRange{
		Start: timestamppb.New(parseTime(t, "2024-01-01T00:00:00Z")),
		End:   timestamppb.New(parseTime(t, "2024-04-01T00:00:00Z")),
	}

	// Month grain on a rollup dimension: routed to the rollup table
	res, err := srv.MetricsViewAggregation(testCtx(), &runtimev1.MetricsViewAggregationRequest{
		InstanceId:  instanceID,
		MetricsView: "mv",
		Dimensions:  []*runtimev1.MetricsViewAggregationDimension{{Name: "publisher"}},
		Measures:    []*runtimev1.MetricsViewAggregationMeasure{{Name: "total_impressions"}},
		TimeRange:   timeRange,
	})
	require.NoError(t, err)
	require.Equal(t, "rollup_month", res.ServingTable)

	// A dimension not in the rollup: served from the base table
	res, err = srv.MetricsViewAggregation(testCtx(), &runtimev1.MetricsViewAggregationRequest{
		InstanceId:  instanceID,
		MetricsView: "mv",
		Dimensions:  []*runtimev1.MetricsViewAggregationDimension{{Name: "country"}},
		Measures:    []*runtimev1.MetricsViewAggregationMeasure{{Name: "total_impressions"}},
		TimeRange:   timeRange,
	})
	require.NoError(t, err)
	require.Empty(t, res.ServingTable)
}

func TestServer_MetricsViewTimeSeries_ServingTable(t *testing.T) {
	t.Parallel()
	srv, instanceID := getRollupTestServer(t)

	res, err := srv.MetricsViewTimeSeries(testCtx(), &runtimev1.MetricsViewTimeSeriesRequest{
		InstanceId:      instanceID,
		MetricsViewName: "mv",
		MeasureNames:    []string{"total_impressions"},
		TimeStart:       timestamppb.New(parseTime(t, "2024-01-01T00:00:00Z")),
		TimeEnd:         timestamppb.New(parseTime(t, "2024-04-01T00:00:00Z")),
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_MONTH,
	})
	require.NoError(t, err)
	require.Equal(t, "rollup_month", res.ServingTable)

	// Hour grain is not derivable from the monthly rollup: served from the base table
	res, err = srv.MetricsViewTimeSeries(testCtx(), &runtimev1.MetricsViewTimeSeriesRequest{
		InstanceId:      instanceID,
		MetricsViewName: "mv",
		MeasureNames:    []string{"total_impressions"},
		TimeStart:       timestamppb.New(parseTime(t, "2024-01-01T00:00:00Z")),
		TimeEnd:         timestamppb.New(parseTime(t, "2024-01-02T00:00:00Z")),
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_HOUR,
	})
	require.NoError(t, err)
	require.Empty(t, res.ServingTable)
}

func TestServer_MetricsViewComparison_ServingTable(t *testing.T) {
	t.Parallel()
	srv, instanceID := getRollupTestServer(t)

	res, err := srv.MetricsViewComparison(testCtx(), &runtimev1.MetricsViewComparisonRequest{
		InstanceId:      instanceID,
		MetricsViewName: "mv",
		Dimension:       &runtimev1.MetricsViewAggregationDimension{Name: "publisher"},
		Measures:        []*runtimev1.MetricsViewAggregationMeasure{{Name: "total_impressions"}},
		TimeRange: &runtimev1.TimeRange{
			Start: timestamppb.New(parseTime(t, "2024-01-01T00:00:00Z")),
			End:   timestamppb.New(parseTime(t, "2024-04-01T00:00:00Z")),
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{{
			Name:     "total_impressions",
			SortType: runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_BASE_VALUE,
		}},
	})
	require.NoError(t, err)
	require.Equal(t, "rollup_month", res.ServingTable)
}
