package queries_test

import (
	"context"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/queries"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"

	_ "github.com/rilldata/rill/runtime/drivers/duckdb"
)

func BenchmarkMetricsViewsAggregation(b *testing.B) {
	rt, instanceID := testruntime.NewInstanceForProject(b, "ad_bids")

	ctrl, err := rt.Controller(context.Background(), instanceID)
	require.NoError(b, err)
	r, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: "ad_bids_metrics"}, false)
	require.NoError(b, err)
	mv := r.GetMetricsView().Spec

	limit := int64(10)
	q := &queries.MetricsViewAggregation{
		MetricsViewName: "ad_bids_metrics",
		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
			{
				Name: "pub",
			},

			{
				Name:      "timestamp",
				TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_MONTH,
			},
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "measure_1",
			},
		},
		MetricsView: mv,
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "pub",
			},
			{
				Name: "timestamp",
			},
		},

		Limit: &limit,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {

		err := q.Resolve(context.Background(), rt, instanceID, 0)
		require.NoError(b, err)
		require.NotEmpty(b, q.Result)
	}
}

func BenchmarkMetricsViewsAggregation_pivot(b *testing.B) {
	rt, instanceID := testruntime.NewInstanceForProject(b, "ad_bids")

	ctrl, err := rt.Controller(context.Background(), instanceID)
	require.NoError(b, err)
	r, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: "ad_bids_metrics"}, false)
	require.NoError(b, err)
	mv := r.GetMetricsView().Spec

	limit := int64(10)
	q := &queries.MetricsViewAggregation{
		MetricsViewName: "ad_bids_metrics",
		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
			{
				Name: "pub",
			},

			{
				Name:      "timestamp",
				TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_MONTH,
			},
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "measure_1",
			},
		},
		MetricsView: mv,
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "pub",
			},
		},
		PivotOn: []string{
			"timestamp",
		},
		Limit: &limit,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := q.Resolve(context.Background(), rt, instanceID, 0)
		require.NoError(b, err)
		require.NotEmpty(b, q.Result)
	}
}

func BenchmarkMetricsViewsAggregation_spending(b *testing.B) {
	rt, instanceID := testruntime.NewInstanceForProject(b, "spending")

	ctrl, err := rt.Controller(context.Background(), instanceID)
	require.NoError(b, err)
	r, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: "spending_dashboard"}, false)
	require.NoError(b, err)
	mv := r.GetMetricsView().Spec

	limit := int64(10)
	q := &queries.MetricsViewAggregation{
		MetricsViewName: "spending_dashboard",
		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
			{
				Name: "recipient_parent_name",
			},

			{
				Name:      "action_date",
				TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
			},
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "total_records",
			},
		},
		MetricsView: mv,
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "recipient_parent_name",
			},
			{
				Name: "action_date",
			},
		},

		Limit: &limit,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {

		err := q.Resolve(context.Background(), rt, instanceID, 0)
		require.NoError(b, err)
		require.NotEmpty(b, q.Result)
	}
}

func BenchmarkMetricsViewsAggregation_spending_100(b *testing.B) {
	rt, instanceID := testruntime.NewInstanceForProject(b, "spending")

	ctrl, err := rt.Controller(context.Background(), instanceID)
	require.NoError(b, err)
	r, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: "spending_dashboard"}, false)
	require.NoError(b, err)
	mv := r.GetMetricsView().Spec

	limit := int64(100)
	q := &queries.MetricsViewAggregation{
		MetricsViewName: "spending_dashboard",
		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
			{
				Name: "recipient_parent_name",
			},

			{
				Name:      "action_date",
				TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
			},
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "total_records",
			},
		},
		MetricsView: mv,
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "recipient_parent_name",
			},
			{
				Name: "action_date",
			},
		},

		Limit: &limit,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {

		err := q.Resolve(context.Background(), rt, instanceID, 0)
		require.NoError(b, err)
		require.NotEmpty(b, q.Result)
	}
}

func BenchmarkMetricsViewsAggregation_spending_pivot(b *testing.B) {
	rt, instanceID := testruntime.NewInstanceForProject(b, "spending")

	ctrl, err := rt.Controller(context.Background(), instanceID)
	require.NoError(b, err)
	r, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: "spending_dashboard"}, false)
	require.NoError(b, err)
	mv := r.GetMetricsView().Spec

	limit := int64(10)
	q := &queries.MetricsViewAggregation{
		MetricsViewName: "spending_dashboard",
		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
			{
				Name: "recipient_parent_name",
			},

			{
				Name:      "action_date",
				TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
			},
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "total_records",
			},
		},
		MetricsView: mv,
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "recipient_parent_name",
			},
		},
		PivotOn: []string{
			"action_date",
		},
		Limit: &limit,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {

		err := q.Resolve(context.Background(), rt, instanceID, 0)
		require.NoError(b, err)
		require.NotEmpty(b, q.Result)
	}
}

func BenchmarkMetricsViewsAggregation_spending_pivot_100(b *testing.B) {
	rt, instanceID := testruntime.NewInstanceForProject(b, "spending")

	ctrl, err := rt.Controller(context.Background(), instanceID)
	require.NoError(b, err)
	r, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: "spending_dashboard"}, false)
	require.NoError(b, err)
	mv := r.GetMetricsView().Spec

	limit := int64(100)
	q := &queries.MetricsViewAggregation{
		MetricsViewName: "spending_dashboard",
		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
			{
				Name: "recipient_parent_name",
			},

			{
				Name:      "action_date",
				TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
			},
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "total_records",
			},
		},
		MetricsView: mv,
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "recipient_parent_name",
			},
		},
		PivotOn: []string{
			"action_date",
		},
		Limit: &limit,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {

		err := q.Resolve(context.Background(), rt, instanceID, 0)
		require.NoError(b, err)
		require.NotEmpty(b, q.Result)
	}
}
