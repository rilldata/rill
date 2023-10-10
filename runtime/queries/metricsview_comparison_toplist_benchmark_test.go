package queries

import (
	"context"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	_ "github.com/rilldata/rill/runtime/drivers/duckdb"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func BenchmarkMetricsViewsComparison_compare(b *testing.B) {
	rt, instanceID := testruntime.NewInstanceForProject(b, "ad_bids")

	ctr := ColumnTimeRange{
		TableName:  "ad_bids",
		ColumnName: "timestamp",
	}
	err := ctr.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(b, err)
	diff := ctr.Result.Max.AsTime().Sub(ctr.Result.Min.AsTime())
	maxTime := ctr.Result.Min.AsTime().Add(diff / 2)

	ctrl, err := rt.Controller(instanceID)
	require.NoError(b, err)
	res, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: "ad_bids_metrics"}, false)
	require.NoError(b, err)
	mv := res.GetMetricsView().Spec

	q := &MetricsViewComparisonToplist{
		MetricsViewName: "ad_bids_metrics",
		DimensionName:   "dom",
		MeasureNames:    []string{"measure_1"},
		MetricsView:     mv,
		BaseTimeRange: &runtimev1.TimeRange{
			Start: ctr.Result.Min,
			End:   timestamppb.New(maxTime),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: timestamppb.New(maxTime),
			End:   ctr.Result.Max,
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				MeasureName: "measure_1",
				Type:        runtimev1.MetricsViewComparisonSortType_METRICS_VIEW_COMPARISON_SORT_TYPE_COMPARISON_VALUE,
				Ascending:   false,
			},
		},
		Limit: 250,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {

		err := q.Resolve(context.Background(), rt, instanceID, 0)
		require.NoError(b, err)
		require.NotEmpty(b, q.Result)
	}
}

func BenchmarkMetricsViewsComparison_nocompare_all(b *testing.B) {
	rt, instanceID := testruntime.NewInstanceForProject(b, "ad_bids")

	ctr := ColumnTimeRange{
		TableName:  "ad_bids",
		ColumnName: "timestamp",
	}
	err := ctr.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(b, err)

	ctrl, err := rt.Controller(instanceID)
	require.NoError(b, err)
	res, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: "ad_bids_metrics"}, false)
	require.NoError(b, err)
	mv := res.GetMetricsView().Spec

	q := &MetricsViewComparisonToplist{
		MetricsViewName: "ad_bids_metrics",
		DimensionName:   "dom",
		MeasureNames:    []string{"measure_1"},
		MetricsView:     mv,
		BaseTimeRange: &runtimev1.TimeRange{
			Start: ctr.Result.Min,
			End:   ctr.Result.Max,
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				MeasureName: "measure_1",
				Type:        runtimev1.MetricsViewComparisonSortType_METRICS_VIEW_COMPARISON_SORT_TYPE_COMPARISON_VALUE,
				Ascending:   false,
			},
		},
		Limit: 250,
		Exact: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {

		err := q.Resolve(context.Background(), rt, instanceID, 0)
		require.NoError(b, err)
		require.NotEmpty(b, q.Result)
	}
}

func BenchmarkMetricsViewsComparison_compare_spending(b *testing.B) {
	rt, instanceID := testruntime.NewInstanceForProject(b, "spending")

	ctrl, err := rt.Controller(instanceID)
	require.NoError(b, err)
	res, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: "ad_bids_metrics"}, false)
	require.NoError(b, err)
	mv := res.GetMetricsView().Spec

	q := &MetricsViewComparisonToplist{
		MetricsViewName: "spending_dashboard",
		DimensionName:   "recipient_state_name",
		MeasureNames:    []string{"total_agencies"},
		MetricsView:     mv,
		BaseTimeRange: &runtimev1.TimeRange{
			Start: parseTimeB(b, "2020-01-01T00:00:00Z"),
			End:   parseTimeB(b, "2021-11-01T00:00:00Z"),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: parseTimeB(b, "2021-11-01T00:00:00Z"),
			End:   parseTimeB(b, "2024-10-01T00:00:00Z"),
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				MeasureName: "total_agencies",
				Type:        runtimev1.MetricsViewComparisonSortType_METRICS_VIEW_COMPARISON_SORT_TYPE_COMPARISON_VALUE,
				Ascending:   false,
			},
		},
		Limit: 250,
		Exact: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {

		err := q.Resolve(context.Background(), rt, instanceID, 0)
		require.NoError(b, err)
		require.NotEmpty(b, q.Result)
	}
}

func BenchmarkMetricsViewsComparison_nocompare_all_spending(b *testing.B) {
	rt, instanceID := testruntime.NewInstanceForProject(b, "spending")

	ctr := ColumnTimeRange{
		TableName:  "spending",
		ColumnName: "action_date",
	}
	err := ctr.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(b, err)
	ctrl, err := rt.Controller(instanceID)
	require.NoError(b, err)
	res, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: "ad_bids_metrics"}, false)
	require.NoError(b, err)
	mv := res.GetMetricsView().Spec

	q := &MetricsViewComparisonToplist{
		MetricsViewName: "spending_dashboard",
		DimensionName:   "recipient_state_name",
		MeasureNames:    []string{"total_agencies"},
		MetricsView:     mv,
		BaseTimeRange: &runtimev1.TimeRange{
			Start: ctr.Result.Min,
			End:   ctr.Result.Max,
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				MeasureName: "total_agencies",
				Type:        runtimev1.MetricsViewComparisonSortType_METRICS_VIEW_COMPARISON_SORT_TYPE_COMPARISON_VALUE,
				Ascending:   false,
			},
		},
		Limit: 250,
		Exact: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {

		err := q.Resolve(context.Background(), rt, instanceID, 0)
		require.NoError(b, err)
		require.NotEmpty(b, q.Result)
	}
}

func BenchmarkMetricsViewsComparison_delta_compare(b *testing.B) {
	rt, instanceID := testruntime.NewInstanceForProject(b, "ad_bids")

	ctr := ColumnTimeRange{
		TableName:  "ad_bids",
		ColumnName: "timestamp",
	}
	err := ctr.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(b, err)
	diff := ctr.Result.Max.AsTime().Sub(ctr.Result.Min.AsTime())
	maxTime := ctr.Result.Min.AsTime().Add(diff / 2)
	ctrl, err := rt.Controller(instanceID)
	require.NoError(b, err)
	res, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: "ad_bids_metrics"}, false)
	require.NoError(b, err)
	mv := res.GetMetricsView().Spec

	q := &MetricsViewComparisonToplist{
		MetricsViewName: "ad_bids_metrics",
		DimensionName:   "dom",
		MeasureNames:    []string{"measure_1"},
		MetricsView:     mv,
		BaseTimeRange: &runtimev1.TimeRange{
			Start: ctr.Result.Min,
			End:   timestamppb.New(maxTime),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: timestamppb.New(maxTime),
			End:   ctr.Result.Max,
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				MeasureName: "measure_1",
				Type:        runtimev1.MetricsViewComparisonSortType_METRICS_VIEW_COMPARISON_SORT_TYPE_COMPARISON_VALUE,
				Ascending:   false,
			},
		},
		Limit: 250,
		Exact: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {

		err := q.Resolve(context.Background(), rt, instanceID, 0)
		require.NoError(b, err)
		require.NotEmpty(b, q.Result)
	}
}

func BenchmarkMetricsViewsComparison_delta_nocompare_all(b *testing.B) {
	rt, instanceID := testruntime.NewInstanceForProject(b, "ad_bids")

	ctr := ColumnTimeRange{
		TableName:  "ad_bids",
		ColumnName: "timestamp",
	}
	err := ctr.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(b, err)
	ctrl, err := rt.Controller(instanceID)
	require.NoError(b, err)
	res, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: "ad_bids_metrics"}, false)
	require.NoError(b, err)
	mv := res.GetMetricsView().Spec

	q := &MetricsViewComparisonToplist{
		MetricsViewName: "ad_bids_metrics",
		DimensionName:   "dom",
		MeasureNames:    []string{"measure_1"},
		MetricsView:     mv,
		BaseTimeRange: &runtimev1.TimeRange{
			Start: ctr.Result.Min,
			End:   ctr.Result.Max,
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				MeasureName: "measure_1",
				Type:        runtimev1.MetricsViewComparisonSortType_METRICS_VIEW_COMPARISON_SORT_TYPE_ABS_DELTA,
				Ascending:   false,
			},
		},
		Limit: 250,
		Exact: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {

		err := q.Resolve(context.Background(), rt, instanceID, 0)
		require.NoError(b, err)
		require.NotEmpty(b, q.Result)
	}
}

func BenchmarkMetricsViewsComparison_delta_compare_spending(b *testing.B) {
	rt, instanceID := testruntime.NewInstanceForProject(b, "spending")
	ctrl, err := rt.Controller(instanceID)
	require.NoError(b, err)
	res, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: "ad_bids_metrics"}, false)
	require.NoError(b, err)
	mv := res.GetMetricsView().Spec

	q := &MetricsViewComparisonToplist{
		MetricsViewName: "spending_dashboard",
		DimensionName:   "recipient_state_name",
		MeasureNames:    []string{"total_agencies"},
		MetricsView:     mv,
		BaseTimeRange: &runtimev1.TimeRange{
			Start: parseTimeB(b, "2020-01-01T00:00:00Z"),
			End:   parseTimeB(b, "2021-11-01T00:00:00Z"),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: parseTimeB(b, "2021-11-01T00:00:00Z"),
			End:   parseTimeB(b, "2024-10-01T00:00:00Z"),
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				MeasureName: "total_agencies",
				Type:        runtimev1.MetricsViewComparisonSortType_METRICS_VIEW_COMPARISON_SORT_TYPE_ABS_DELTA,
				Ascending:   false,
			},
		},
		Limit: 250,
		Exact: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {

		err := q.Resolve(context.Background(), rt, instanceID, 0)
		require.NoError(b, err)
		require.NotEmpty(b, q.Result)
	}
}

func BenchmarkMetricsViewsComparison_delta_nocompare_all_spending(b *testing.B) {
	rt, instanceID := testruntime.NewInstanceForProject(b, "spending")

	ctr := ColumnTimeRange{
		TableName:  "spending",
		ColumnName: "action_date",
	}
	err := ctr.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(b, err)
	ctrl, err := rt.Controller(instanceID)
	require.NoError(b, err)
	res, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: "ad_bids_metrics"}, false)
	require.NoError(b, err)
	mv := res.GetMetricsView().Spec

	q := &MetricsViewComparisonToplist{
		MetricsViewName: "spending_dashboard",
		DimensionName:   "recipient_state_name",
		MeasureNames:    []string{"total_agencies"},
		MetricsView:     mv,
		BaseTimeRange: &runtimev1.TimeRange{
			Start: ctr.Result.Min,
			End:   ctr.Result.Max,
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				MeasureName: "total_agencies",
				Type:        runtimev1.MetricsViewComparisonSortType_METRICS_VIEW_COMPARISON_SORT_TYPE_ABS_DELTA,
				Ascending:   false,
			},
		},
		Limit: 250,
		Exact: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {

		err := q.Resolve(context.Background(), rt, instanceID, 0)
		require.NoError(b, err)
		require.NotEmpty(b, q.Result)
	}
}

func BenchmarkMetricsViewsComparison_delta_high_cardinality_compare_spending(b *testing.B) {
	rt, instanceID := testruntime.NewInstanceForProject(b, "spending")
	ctrl, err := rt.Controller(instanceID)
	require.NoError(b, err)
	res, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: "ad_bids_metrics"}, false)
	require.NoError(b, err)
	mv := res.GetMetricsView().Spec

	q := &MetricsViewComparisonToplist{
		MetricsViewName: "spending_dashboard",
		DimensionName:   "recipient_parent_name",
		MeasureNames:    []string{"total_agencies"},
		MetricsView:     mv,
		BaseTimeRange: &runtimev1.TimeRange{
			Start: parseTimeB(b, "2020-01-01T00:00:00Z"),
			End:   parseTimeB(b, "2021-11-01T00:00:00Z"),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: parseTimeB(b, "2021-11-01T00:00:00Z"),
			End:   parseTimeB(b, "2024-10-01T00:00:00Z"),
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				MeasureName: "total_agencies",
				Type:        runtimev1.MetricsViewComparisonSortType_METRICS_VIEW_COMPARISON_SORT_TYPE_ABS_DELTA,
				Ascending:   false,
			},
		},
		Limit: 250,
		Exact: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {

		err := q.Resolve(context.Background(), rt, instanceID, 0)
		require.NoError(b, err)
		require.NotEmpty(b, q.Result)
	}
}

func BenchmarkMetricsViewsComparison_delta_high_cardinality_compare_spending_approximate(b *testing.B) {
	rt, instanceID := testruntime.NewInstanceForProject(b, "spending")
	ctrl, err := rt.Controller(instanceID)
	require.NoError(b, err)
	res, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: "ad_bids_metrics"}, false)
	require.NoError(b, err)
	mv := res.GetMetricsView().Spec

	q := &MetricsViewComparisonToplist{
		MetricsViewName: "spending_dashboard",
		DimensionName:   "recipient_parent_name",
		MeasureNames:    []string{"total_agencies"},
		MetricsView:     mv,
		BaseTimeRange: &runtimev1.TimeRange{
			Start: parseTimeB(b, "2020-01-01T00:00:00Z"),
			End:   parseTimeB(b, "2021-11-01T00:00:00Z"),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: parseTimeB(b, "2021-11-01T00:00:00Z"),
			End:   parseTimeB(b, "2024-10-01T00:00:00Z"),
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				MeasureName: "total_agencies",
				Type:        runtimev1.MetricsViewComparisonSortType_METRICS_VIEW_COMPARISON_SORT_TYPE_ABS_DELTA,
				Ascending:   false,
			},
		},
		Limit: 250,
		Exact: false,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {

		err := q.Resolve(context.Background(), rt, instanceID, 0)
		require.NoError(b, err)
		require.NotEmpty(b, q.Result)
	}
}

func BenchmarkMetricsViewsComparison_delta_high_cardinality_nocompare_all_spending(b *testing.B) {
	rt, instanceID := testruntime.NewInstanceForProject(b, "spending")

	ctr := ColumnTimeRange{
		TableName:  "spending",
		ColumnName: "action_date",
	}
	err := ctr.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(b, err)
	ctrl, err := rt.Controller(instanceID)
	require.NoError(b, err)
	res, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: "ad_bids_metrics"}, false)
	require.NoError(b, err)
	mv := res.GetMetricsView().Spec

	q := &MetricsViewComparisonToplist{
		MetricsViewName: "spending_dashboard",
		DimensionName:   "recipient_parent_name",
		MeasureNames:    []string{"total_agencies"},
		MetricsView:     mv,
		BaseTimeRange: &runtimev1.TimeRange{
			Start: ctr.Result.Min,
			End:   ctr.Result.Max,
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				MeasureName: "total_agencies",
				Type:        runtimev1.MetricsViewComparisonSortType_METRICS_VIEW_COMPARISON_SORT_TYPE_ABS_DELTA,
				Ascending:   false,
			},
		},
		Limit: 250,
		Exact: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {

		err := q.Resolve(context.Background(), rt, instanceID, 0)
		require.NoError(b, err)
		require.NotEmpty(b, q.Result)
	}
}

func BenchmarkMetricsViewsComparison_high_cardinality_compare_spending(b *testing.B) {
	rt, instanceID := testruntime.NewInstanceForProject(b, "spending")
	ctrl, err := rt.Controller(instanceID)
	require.NoError(b, err)
	res, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: "ad_bids_metrics"}, false)
	require.NoError(b, err)
	mv := res.GetMetricsView().Spec

	q := &MetricsViewComparisonToplist{
		MetricsViewName: "spending_dashboard",
		DimensionName:   "recipient_parent_name",
		MeasureNames:    []string{"total_agencies"},
		MetricsView:     mv,
		BaseTimeRange: &runtimev1.TimeRange{
			Start: parseTimeB(b, "2020-01-01T00:00:00Z"),
			End:   parseTimeB(b, "2021-11-01T00:00:00Z"),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: parseTimeB(b, "2021-11-01T00:00:00Z"),
			End:   parseTimeB(b, "2024-10-01T00:00:00Z"),
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				MeasureName: "total_agencies",
				Type:        runtimev1.MetricsViewComparisonSortType_METRICS_VIEW_COMPARISON_SORT_TYPE_COMPARISON_VALUE,
				Ascending:   false,
			},
		},
		Limit: 250,
		Exact: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {

		err := q.Resolve(context.Background(), rt, instanceID, 0)
		require.NoError(b, err)
		require.NotEmpty(b, q.Result)
	}
}

func BenchmarkMetricsViewsComparison_high_cardinality_compare_spending_approximate_limit_250(b *testing.B) {
	rt, instanceID := testruntime.NewInstanceForProject(b, "spending")
	ctrl, err := rt.Controller(instanceID)
	require.NoError(b, err)
	res, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: "ad_bids_metrics"}, false)
	require.NoError(b, err)
	mv := res.GetMetricsView().Spec

	q := &MetricsViewComparisonToplist{
		MetricsViewName: "spending_dashboard",
		DimensionName:   "recipient_parent_name",
		MeasureNames:    []string{"total_agencies"},
		MetricsView:     mv,
		BaseTimeRange: &runtimev1.TimeRange{
			Start: parseTimeB(b, "2020-01-01T00:00:00Z"),
			End:   parseTimeB(b, "2021-11-01T00:00:00Z"),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: parseTimeB(b, "2021-11-01T00:00:00Z"),
			End:   parseTimeB(b, "2024-10-01T00:00:00Z"),
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				MeasureName: "total_agencies",
				Type:        runtimev1.MetricsViewComparisonSortType_METRICS_VIEW_COMPARISON_SORT_TYPE_COMPARISON_VALUE,
				Ascending:   false,
			},
		},
		Limit: 250,
		Exact: false,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {

		err := q.Resolve(context.Background(), rt, instanceID, 0)
		require.NoError(b, err)
		require.NotEmpty(b, q.Result)
	}
}

func BenchmarkMetricsViewsComparison_delta_high_cardinality_compare_spending_approximate_limit_500(b *testing.B) {
	rt, instanceID := testruntime.NewInstanceForProject(b, "spending")
	ctrl, err := rt.Controller(instanceID)
	require.NoError(b, err)
	res, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: "ad_bids_metrics"}, false)
	require.NoError(b, err)
	mv := res.GetMetricsView().Spec

	q := &MetricsViewComparisonToplist{
		MetricsViewName: "spending_dashboard",
		DimensionName:   "recipient_parent_name",
		MeasureNames:    []string{"total_agencies"},
		MetricsView:     mv,
		BaseTimeRange: &runtimev1.TimeRange{
			Start: parseTimeB(b, "2020-01-01T00:00:00Z"),
			End:   parseTimeB(b, "2021-11-01T00:00:00Z"),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: parseTimeB(b, "2021-11-01T00:00:00Z"),
			End:   parseTimeB(b, "2024-10-01T00:00:00Z"),
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				MeasureName: "total_agencies",
				Type:        runtimev1.MetricsViewComparisonSortType_METRICS_VIEW_COMPARISON_SORT_TYPE_ABS_DELTA,
				Ascending:   false,
			},
		},
		Limit: 250 * 2,
		Exact: false,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {

		err := q.Resolve(context.Background(), rt, instanceID, 0)
		require.NoError(b, err)
		require.NotEmpty(b, q.Result)
	}
}
func BenchmarkMetricsViewsComparison_delta_high_cardinality_compare_spending_approximate_limit_1250(b *testing.B) {
	rt, instanceID := testruntime.NewInstanceForProject(b, "spending")
	ctrl, err := rt.Controller(instanceID)
	require.NoError(b, err)
	res, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: "ad_bids_metrics"}, false)
	require.NoError(b, err)
	mv := res.GetMetricsView().Spec

	q := &MetricsViewComparisonToplist{
		MetricsViewName: "spending_dashboard",
		DimensionName:   "recipient_parent_name",
		MeasureNames:    []string{"total_agencies"},
		MetricsView:     mv,
		BaseTimeRange: &runtimev1.TimeRange{
			Start: parseTimeB(b, "2020-01-01T00:00:00Z"),
			End:   parseTimeB(b, "2021-11-01T00:00:00Z"),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: parseTimeB(b, "2021-11-01T00:00:00Z"),
			End:   parseTimeB(b, "2024-10-01T00:00:00Z"),
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				MeasureName: "total_agencies",
				Type:        runtimev1.MetricsViewComparisonSortType_METRICS_VIEW_COMPARISON_SORT_TYPE_ABS_DELTA,
				Ascending:   false,
			},
		},
		Limit: 250 * 5,
		Exact: false,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {

		err := q.Resolve(context.Background(), rt, instanceID, 0)
		require.NoError(b, err)
		require.NotEmpty(b, q.Result)
	}
}

func BenchmarkMetricsViewsComparison_high_cardinality_nocompare_all_spending(b *testing.B) {
	rt, instanceID := testruntime.NewInstanceForProject(b, "spending")

	ctr := ColumnTimeRange{
		TableName:  "spending",
		ColumnName: "action_date",
	}
	err := ctr.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(b, err)
	ctrl, err := rt.Controller(instanceID)
	require.NoError(b, err)
	res, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: "ad_bids_metrics"}, false)
	require.NoError(b, err)
	mv := res.GetMetricsView().Spec

	q := &MetricsViewComparisonToplist{
		MetricsViewName: "spending_dashboard",
		DimensionName:   "recipient_parent_name",
		MeasureNames:    []string{"total_agencies"},
		MetricsView:     mv,
		BaseTimeRange: &runtimev1.TimeRange{
			Start: ctr.Result.Min,
			End:   ctr.Result.Max,
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				MeasureName: "total_agencies",
				Type:        runtimev1.MetricsViewComparisonSortType_METRICS_VIEW_COMPARISON_SORT_TYPE_COMPARISON_VALUE,
				Ascending:   false,
			},
		},
		Limit: 250,
		Exact: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {

		err := q.Resolve(context.Background(), rt, instanceID, 0)
		require.NoError(b, err)
		require.NotEmpty(b, q.Result)
	}
}
