package queries_test

import (
	"context"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	_ "github.com/rilldata/rill/runtime/drivers/duckdb"
	"github.com/rilldata/rill/runtime/queries"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func BenchmarkMetricsViewsTotals(b *testing.B) {
	rt, instanceID, mv := prepareEnvironment(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q := &queries.MetricsViewTotals{
			MetricsViewName: "ad_bids_metrics",
			MeasureNames:    []string{"measure_1"},
			MetricsView:     mv,
		}
		err := q.Resolve(context.Background(), rt, instanceID, 0)
		require.NoError(b, err)
		require.NotEmpty(b, q.Result)
	}
}

func BenchmarkMetricsViewsToplist(b *testing.B) {
	rt, instanceID, mv := prepareEnvironment(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q := &queries.MetricsViewToplist{
			MetricsViewName: "ad_bids_metrics",
			DimensionName:   "dom",
			MeasureNames:    []string{"measure_1"},
			Sort: []*runtimev1.MetricsViewSort{
				{
					Name: "measure_1",
				},
			},
			MetricsView: mv,
		}
		err := q.Resolve(context.Background(), rt, instanceID, 0)
		require.NoError(b, err)
		require.NotEmpty(b, q.Result)
	}
}

func BenchmarkMetricsViewsTimeSeries(b *testing.B) {
	rt, instanceID, mv := prepareEnvironment(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q := &queries.MetricsViewTimeSeries{
			MetricsViewName: "ad_bids_metrics",
			TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_DAY,
			MeasureNames:    []string{"measure_1"},
			MetricsView:     mv,
		}
		err := q.Resolve(context.Background(), rt, instanceID, 0)
		require.NoError(b, err)
		require.NotEmpty(b, q.Result)
	}
}

func BenchmarkMetricsViewsTimeSeries_TimeZone(b *testing.B) {
	rt, instanceID, mv := prepareEnvironment(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q := &queries.MetricsViewTimeSeries{
			MetricsViewName: "ad_bids_metrics",
			TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_DAY,
			MeasureNames:    []string{"measure_1"},
			TimeZone:        "Asia/Kathmandu",
			MetricsView:     mv,
		}
		err := q.Resolve(context.Background(), rt, instanceID, 0)
		require.NoError(b, err)
		require.NotEmpty(b, q.Result)
	}
}

func BenchmarkMetricsViewsTimeSeries_TimeZone_Hour(b *testing.B) {
	rt, instanceID, mv := prepareEnvironment(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q := &queries.MetricsViewTimeSeries{
			MetricsViewName: "ad_bids_metrics",
			TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_HOUR,
			MeasureNames:    []string{"measure_1"},
			TimeZone:        "Asia/Kathmandu",
			MetricsView:     mv,
		}
		err := q.Resolve(context.Background(), rt, instanceID, 0)
		require.NoError(b, err)
		require.NotEmpty(b, q.Result)
	}
}

func prepareEnvironment(b *testing.B) (*runtime.Runtime, string, *runtimev1.MetricsViewSpec) {
	rt, instanceID := testruntime.NewInstanceForProject(b, "ad_bids")

	ctrl, err := rt.Controller(instanceID)
	require.NoError(b, err)

	obj, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: "ad_bids_metrics"}, false)
	require.NoError(b, err)

	mv := obj.GetMetricsView().Spec
	return rt, instanceID, mv
}
