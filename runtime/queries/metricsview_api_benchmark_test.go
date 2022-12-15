package queries

import (
	"context"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	_ "github.com/rilldata/rill/runtime/drivers/duckdb"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func BenchmarkMetricsViewsTotals(b *testing.B) {
	rt, instanceID := testruntime.NewInstanceForProject(b, "ad_bids")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q := &MetricsViewTotals{
			MetricsViewName: "ad_bids_metrics",
			MeasureNames:    []string{"measure_0"},
		}
		err := q.Resolve(context.Background(), rt, instanceID, 0)
		require.NoError(b, err)
		require.NotEmpty(b, q.Result)
	}
}

func BenchmarkMetricsViewsToplist(b *testing.B) {
	rt, instanceID := testruntime.NewInstanceForProject(b, "ad_bids")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q := &MetricsViewToplist{
			MetricsViewName: "ad_bids_metrics",
			DimensionName:   "domain",
			MeasureNames:    []string{"measure_1"},
			Sort: []*runtimev1.MetricsViewSort{
				{
					Name: "measure_1",
				},
			},
		}
		err := q.Resolve(context.Background(), rt, instanceID, 0)
		require.NoError(b, err)
		require.NotEmpty(b, q.Result)
	}
}

func BenchmarkMetricsViewsTimeSeries(b *testing.B) {
	rt, instanceID := testruntime.NewInstanceForProject(b, "ad_bids")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q := &MetricsViewTimeSeries{
			MetricsViewName: "ad_bids_metrics",
			TimeGranularity: "DAY",
			MeasureNames:    []string{"measure_0"},
		}
		err := q.Resolve(context.Background(), rt, instanceID, 0)
		require.NoError(b, err)
		require.NotEmpty(b, q.Result)
	}
}
