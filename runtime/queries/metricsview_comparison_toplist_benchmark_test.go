package queries_test

import (
	"context"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	_ "github.com/rilldata/rill/runtime/drivers/duckdb"
	"github.com/rilldata/rill/runtime/queries"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func BenchmarkMetricsViewsComparison_compare(b *testing.B) {
	rt, instanceID := testruntime.NewInstanceForProject(b, "ad_bids")

	ctr := &queries.ColumnTimeRange{
		TableName:  "ad_bids",
		ColumnName: "timestamp",
	}
	err := ctr.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(b, err)
	diff := ctr.Result.Max.AsTime().Sub(ctr.Result.Min.AsTime())
	maxTime := ctr.Result.Min.AsTime().Add(diff / 2)

	q := &queries.MetricsViewComparison{
		MetricsViewName: "ad_bids_metrics",
		DimensionName:   "dom",
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "measure_1",
			},
		},
		TimeRange: &runtimev1.TimeRange{
			Start: ctr.Result.Min,
			End:   timestamppb.New(maxTime),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: timestamppb.New(maxTime),
			End:   ctr.Result.Max,
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				Name:     "measure_1",
				SortType: runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_COMPARISON_VALUE,
				Desc:     true,
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

	ctr := &queries.ColumnTimeRange{
		TableName:  "ad_bids",
		ColumnName: "timestamp",
	}
	err := ctr.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(b, err)

	q := &queries.MetricsViewComparison{
		MetricsViewName: "ad_bids_metrics",
		DimensionName:   "dom",
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "measure_1",
			},
		},
		TimeRange: &runtimev1.TimeRange{
			Start: ctr.Result.Min,
			End:   ctr.Result.Max,
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				Name:     "measure_1",
				SortType: runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_COMPARISON_VALUE,
				Desc:     true,
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

	q := &queries.MetricsViewComparison{
		MetricsViewName: "spending_dashboard",
		DimensionName:   "recipient_state_name",
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "total_agencies",
			},
		},
		TimeRange: &runtimev1.TimeRange{
			Start: parseTimeB(b, "2020-01-01T00:00:00Z"),
			End:   parseTimeB(b, "2021-11-01T00:00:00Z"),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: parseTimeB(b, "2021-11-01T00:00:00Z"),
			End:   parseTimeB(b, "2024-10-01T00:00:00Z"),
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				Name:     "total_agencies",
				SortType: runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_COMPARISON_VALUE,
				Desc:     true,
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

	ctr := &queries.ColumnTimeRange{
		TableName:  "spending",
		ColumnName: "action_date",
	}
	err := ctr.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(b, err)

	q := &queries.MetricsViewComparison{
		MetricsViewName: "spending_dashboard",
		DimensionName:   "recipient_state_name",
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "total_agencies",
			},
		},
		TimeRange: &runtimev1.TimeRange{
			Start: ctr.Result.Min,
			End:   ctr.Result.Max,
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				Name:     "total_agencies",
				SortType: runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_COMPARISON_VALUE,
				Desc:     true,
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

	ctr := &queries.ColumnTimeRange{
		TableName:  "ad_bids",
		ColumnName: "timestamp",
	}
	err := ctr.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(b, err)
	diff := ctr.Result.Max.AsTime().Sub(ctr.Result.Min.AsTime())
	maxTime := ctr.Result.Min.AsTime().Add(diff / 2)

	q := &queries.MetricsViewComparison{
		MetricsViewName: "ad_bids_metrics",
		DimensionName:   "dom",
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "measure_1",
			},
		},
		TimeRange: &runtimev1.TimeRange{
			Start: ctr.Result.Min,
			End:   timestamppb.New(maxTime),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: timestamppb.New(maxTime),
			End:   ctr.Result.Max,
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				Name:     "measure_1",
				SortType: runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_COMPARISON_VALUE,
				Desc:     true,
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

func BenchmarkMetricsViewsComparison_delta_compare_with_having(b *testing.B) {
	rt, instanceID := testruntime.NewInstanceForProject(b, "ad_bids")

	ctr := &queries.ColumnTimeRange{
		TableName:  "ad_bids",
		ColumnName: "timestamp",
	}
	err := ctr.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(b, err)
	diff := ctr.Result.Max.AsTime().Sub(ctr.Result.Min.AsTime())
	maxTime := ctr.Result.Min.AsTime().Add(diff / 2)

	q := &queries.MetricsViewComparison{
		MetricsViewName: "ad_bids_metrics",
		DimensionName:   "dom",
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "measure_1",
			},
		},
		TimeRange: &runtimev1.TimeRange{
			Start: ctr.Result.Min,
			End:   timestamppb.New(maxTime),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: timestamppb.New(maxTime),
			End:   ctr.Result.Max,
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				Name:     "measure_1",
				SortType: runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_COMPARISON_VALUE,
				Desc:     true,
			},
		},
		Having: &runtimev1.Expression{
			Expression: &runtimev1.Expression_Cond{
				Cond: &runtimev1.Condition{
					Op: runtimev1.Operation_OPERATION_GT,
					Exprs: []*runtimev1.Expression{
						{
							Expression: &runtimev1.Expression_Ident{
								Ident: "measure_1",
							},
						},
						{
							Expression: &runtimev1.Expression_Val{
								Val: structpb.NewNumberValue(3.25),
							},
						},
					},
				},
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

	ctr := &queries.ColumnTimeRange{
		TableName:  "ad_bids",
		ColumnName: "timestamp",
	}
	err := ctr.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(b, err)

	q := &queries.MetricsViewComparison{
		MetricsViewName: "ad_bids_metrics",
		DimensionName:   "dom",
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "measure_1",
			},
		},
		TimeRange: &runtimev1.TimeRange{
			Start: ctr.Result.Min,
			End:   ctr.Result.Max,
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				Name:     "measure_1",
				SortType: runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_ABS_DELTA,
				Desc:     true,
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

	q := &queries.MetricsViewComparison{
		MetricsViewName: "spending_dashboard",
		DimensionName:   "recipient_state_name",
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "total_agencies",
			},
		},
		TimeRange: &runtimev1.TimeRange{
			Start: parseTimeB(b, "2020-01-01T00:00:00Z"),
			End:   parseTimeB(b, "2021-11-01T00:00:00Z"),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: parseTimeB(b, "2021-11-01T00:00:00Z"),
			End:   parseTimeB(b, "2024-10-01T00:00:00Z"),
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				Name:     "total_agencies",
				SortType: runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_ABS_DELTA,
				Desc:     true,
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

	ctr := &queries.ColumnTimeRange{
		TableName:  "spending",
		ColumnName: "action_date",
	}
	err := ctr.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(b, err)

	q := &queries.MetricsViewComparison{
		MetricsViewName: "spending_dashboard",
		DimensionName:   "recipient_state_name",
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "total_agencies",
			},
		},
		TimeRange: &runtimev1.TimeRange{
			Start: ctr.Result.Min,
			End:   ctr.Result.Max,
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				Name:     "total_agencies",
				SortType: runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_ABS_DELTA,
				Desc:     true,
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

	q := &queries.MetricsViewComparison{
		MetricsViewName: "spending_dashboard",
		DimensionName:   "recipient_parent_name",
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "total_agencies",
			},
		},
		TimeRange: &runtimev1.TimeRange{
			Start: parseTimeB(b, "2020-01-01T00:00:00Z"),
			End:   parseTimeB(b, "2021-11-01T00:00:00Z"),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: parseTimeB(b, "2021-11-01T00:00:00Z"),
			End:   parseTimeB(b, "2024-10-01T00:00:00Z"),
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				Name:     "total_agencies",
				SortType: runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_ABS_DELTA,
				Desc:     true,
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

	q := &queries.MetricsViewComparison{
		MetricsViewName: "spending_dashboard",
		DimensionName:   "recipient_parent_name",
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "total_agencies",
			},
		},
		TimeRange: &runtimev1.TimeRange{
			Start: parseTimeB(b, "2020-01-01T00:00:00Z"),
			End:   parseTimeB(b, "2021-11-01T00:00:00Z"),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: parseTimeB(b, "2021-11-01T00:00:00Z"),
			End:   parseTimeB(b, "2024-10-01T00:00:00Z"),
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				Name:     "total_agencies",
				SortType: runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_ABS_DELTA,
				Desc:     true,
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

	ctr := &queries.ColumnTimeRange{
		TableName:  "spending",
		ColumnName: "action_date",
	}
	err := ctr.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(b, err)

	q := &queries.MetricsViewComparison{
		MetricsViewName: "spending_dashboard",
		DimensionName:   "recipient_parent_name",
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "total_agencies",
			},
		},
		TimeRange: &runtimev1.TimeRange{
			Start: ctr.Result.Min,
			End:   ctr.Result.Max,
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				Name:     "total_agencies",
				SortType: runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_ABS_DELTA,
				Desc:     true,
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

	q := &queries.MetricsViewComparison{
		MetricsViewName: "spending_dashboard",
		DimensionName:   "recipient_parent_name",
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "total_agencies",
			},
		},
		TimeRange: &runtimev1.TimeRange{
			Start: parseTimeB(b, "2020-01-01T00:00:00Z"),
			End:   parseTimeB(b, "2021-11-01T00:00:00Z"),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: parseTimeB(b, "2021-11-01T00:00:00Z"),
			End:   parseTimeB(b, "2024-10-01T00:00:00Z"),
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				Name:     "total_agencies",
				SortType: runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_COMPARISON_VALUE,
				Desc:     true,
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

	q := &queries.MetricsViewComparison{
		MetricsViewName: "spending_dashboard",
		DimensionName:   "recipient_parent_name",
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "total_agencies",
			},
		},
		TimeRange: &runtimev1.TimeRange{
			Start: parseTimeB(b, "2020-01-01T00:00:00Z"),
			End:   parseTimeB(b, "2021-11-01T00:00:00Z"),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: parseTimeB(b, "2021-11-01T00:00:00Z"),
			End:   parseTimeB(b, "2024-10-01T00:00:00Z"),
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				Name:     "total_agencies",
				SortType: runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_COMPARISON_VALUE,
				Desc:     true,
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

	q := &queries.MetricsViewComparison{
		MetricsViewName: "spending_dashboard",
		DimensionName:   "recipient_parent_name",
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "total_agencies",
			},
		},
		TimeRange: &runtimev1.TimeRange{
			Start: parseTimeB(b, "2020-01-01T00:00:00Z"),
			End:   parseTimeB(b, "2021-11-01T00:00:00Z"),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: parseTimeB(b, "2021-11-01T00:00:00Z"),
			End:   parseTimeB(b, "2024-10-01T00:00:00Z"),
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				Name:     "total_agencies",
				SortType: runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_ABS_DELTA,
				Desc:     true,
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

	q := &queries.MetricsViewComparison{
		MetricsViewName: "spending_dashboard",
		DimensionName:   "recipient_parent_name",
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "total_agencies",
			},
		},
		TimeRange: &runtimev1.TimeRange{
			Start: parseTimeB(b, "2020-01-01T00:00:00Z"),
			End:   parseTimeB(b, "2021-11-01T00:00:00Z"),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: parseTimeB(b, "2021-11-01T00:00:00Z"),
			End:   parseTimeB(b, "2024-10-01T00:00:00Z"),
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				Name:     "total_agencies",
				SortType: runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_ABS_DELTA,
				Desc:     true,
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

	ctr := &queries.ColumnTimeRange{
		TableName:  "spending",
		ColumnName: "action_date",
	}
	err := ctr.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(b, err)

	q := &queries.MetricsViewComparison{
		MetricsViewName: "spending_dashboard",
		DimensionName:   "recipient_parent_name",
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "total_agencies",
			},
		},
		TimeRange: &runtimev1.TimeRange{
			Start: ctr.Result.Min,
			End:   ctr.Result.Max,
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				Name:     "total_agencies",
				SortType: runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_COMPARISON_VALUE,
				Desc:     true,
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

func BenchmarkMetricsViewsComparison_high_cardinality_nocompare_all_spending_noExploreExactify(b *testing.B) {
	rt, instanceID := testruntime.NewInstanceForProject(b, "spending")

	ctr := &queries.ColumnTimeRange{
		TableName:  "spending",
		ColumnName: "action_date",
	}
	err := ctr.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(b, err)

	q := &queries.MetricsViewComparison{
		MetricsViewName: "spending_dashboard",
		DimensionName:   "recipient_parent_name",
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "total_agencies",
			},
		},
		TimeRange: &runtimev1.TimeRange{
			Start: ctr.Result.Min,
			End:   ctr.Result.Max,
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				Name:     "total_agencies",
				SortType: runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_COMPARISON_VALUE,
				Desc:     true,
			},
		},
		Limit:                 250,
		Exact:                 true,
		NoExploreExactifyMode: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {

		err := q.Resolve(context.Background(), rt, instanceID, 0)
		require.NoError(b, err)
		require.NotEmpty(b, q.Result)
	}
}
