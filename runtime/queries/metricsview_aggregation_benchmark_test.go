package queries_test

import (
	"context"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/queries"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	_ "github.com/rilldata/rill/runtime/drivers/duckdb"
)

func BenchmarkMetricsViewsAggregation(b *testing.B) {
	rt, instanceID := testruntime.NewInstanceForProject(b, "ad_bids")

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
		Sort0: []*runtimev1.MetricsViewAggregationSort{
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

func BenchmarkMetricsViewsAggregation_pivot_2_measures(t *testing.B) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")

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
			{
				Name: "measure_0",
			},
		},
		Sort0: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "pub",
			},
		},
		PivotOn: []string{
			"timestamp",
		},
		Limit: &limit,
	}
	for i := 0; i < t.N; i++ {
		err := q.Resolve(context.Background(), rt, instanceID, 0)
		require.NoError(t, err)
		require.NotEmpty(t, q.Result)
	}
}

func BenchmarkMetricsViewsAggregation_pivot(b *testing.B) {
	rt, instanceID := testruntime.NewInstanceForProject(b, "ad_bids")

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
		Sort0: []*runtimev1.MetricsViewAggregationSort{
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
		Sort0: []*runtimev1.MetricsViewAggregationSort{
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
		Sort0: []*runtimev1.MetricsViewAggregationSort{
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
		Sort0: []*runtimev1.MetricsViewAggregationSort{
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
		Sort0: []*runtimev1.MetricsViewAggregationSort{
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

func BenchmarkMetricsViewsAggregation_Druid(t *testing.B) {
	dialOpts := []grpc.DialOption{grpc.WithInsecure()}

	conn, err := grpc.Dial(":49009", dialOpts...)
	if err != nil {
		require.NoError(t, err)
	}
	defer conn.Close()

	client := runtimev1.NewQueryServiceClient(conn)
	req := &runtimev1.MetricsViewAggregationRequest{
		InstanceId:  "default",
		MetricsView: "test_data_test",
		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
			{
				Name: "publisher",
			},
			{
				Name:      "__time",
				TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_MONTH,
			},
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "bp",
			},
		},
		Sort0: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "publisher",
			},
			{
				Name: "__time",
			},
		},
		Limit: 10,
	}

	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		resp, err := client.MetricsViewAggregation(context.Background(), req)
		if err != nil {
			require.NoError(t, err)
		}
		rows := resp.Data
		require.NotEmpty(t, rows)
	}
}

func BenchmarkMetricsViewsAggregation_Druid_2_measures(t *testing.B) {
	dialOpts := []grpc.DialOption{grpc.WithInsecure()}

	conn, err := grpc.Dial(":49009", dialOpts...)
	if err != nil {
		require.NoError(t, err)
	}
	defer conn.Close()

	client := runtimev1.NewQueryServiceClient(conn)
	req := &runtimev1.MetricsViewAggregationRequest{
		InstanceId:  "default",
		MetricsView: "test_data_test",
		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
			{
				Name: "publisher",
			},
			{
				Name:      "__time",
				TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_MONTH,
			},
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "bp",
			},
			{
				Name: "rate",
			},
		},
		Sort0: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "publisher",
			},
			{
				Name: "__time",
			},
		},
		Limit: 10,
	}

	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		resp, err := client.MetricsViewAggregation(context.Background(), req)
		if err != nil {
			require.NoError(t, err)
		}
		rows := resp.Data
		require.NotEmpty(t, rows)
	}
}

func BenchmarkMetricsViewsAggregation_Druid_pivot(t *testing.B) {
	dialOpts := []grpc.DialOption{grpc.WithInsecure()}

	conn, err := grpc.Dial(":49009", dialOpts...)
	if err != nil {
		require.NoError(t, err)
	}
	defer conn.Close()

	client := runtimev1.NewQueryServiceClient(conn)
	req := &runtimev1.MetricsViewAggregationRequest{
		InstanceId:  "default",
		MetricsView: "test_data_test",
		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
			{
				Name: "publisher",
			},
			{
				Name:      "__time",
				TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_MONTH,
			},
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "bp",
			},
		},
		Sort0: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "publisher",
			},
		},
		PivotOn: []string{
			"__time",
		},
		Limit: 10,
	}

	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		resp, err := client.MetricsViewAggregation(context.Background(), req)
		if err != nil {
			require.NoError(t, err)
		}
		rows := resp.Data
		require.NotEmpty(t, rows)
	}
}

func BenchmarkMetricsViewsAggregation_Druid_pivot_2_measures(t *testing.B) {
	dialOpts := []grpc.DialOption{grpc.WithInsecure()}

	conn, err := grpc.Dial(":49009", dialOpts...)
	if err != nil {
		require.NoError(t, err)
	}
	defer conn.Close()

	client := runtimev1.NewQueryServiceClient(conn)
	req := &runtimev1.MetricsViewAggregationRequest{
		InstanceId:  "default",
		MetricsView: "test_data_test",
		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
			{
				Name: "publisher",
			},

			{
				Name:      "__time",
				TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_MONTH,
			},
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "bp",
			},
			{
				Name: "rate",
			},
		},
		Sort0: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "publisher",
			},
		},
		PivotOn: []string{
			"__time",
		},
		Limit: 10,
	}

	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		resp, err := client.MetricsViewAggregation(context.Background(), req)
		if err != nil {
			require.NoError(t, err)
		}
		rows := resp.Data
		require.NotEmpty(t, rows)
	}
}
