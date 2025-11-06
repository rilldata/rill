package queries_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/expressionpb"
	"github.com/rilldata/rill/runtime/queries"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/rilldata/rill/runtime/testruntime/testmode"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	_ "github.com/rilldata/rill/runtime/drivers/duckdb"
)

func TestMetricViewAggregationAgainstClickHouse(t *testing.T) {
	// t.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
	testmode.Expensive(t)

	rt, instanceID := testruntime.NewInstanceWithClickhouseProject(t, true)
	t.Run("testMetricsViewsAggregation", func(t *testing.T) { testMetricsViewsAggregation(t, rt, instanceID) })
	t.Run("testMetricsViewsAggregationURI", func(t *testing.T) { testMetricsViewsAggregationURI(t, rt, instanceID) })
	t.Run("testMetricsViewsAggregation_export_day", func(t *testing.T) { testMetricsViewsAggregation_export_day(t, rt, instanceID) })
	t.Run("testMetricsViewsAggregation_export_hour", func(t *testing.T) { testMetricsViewsAggregation_export_hour(t, rt, instanceID) })
	t.Run("testMetricsViewsAggregation_no_limit", func(t *testing.T) { testMetricsViewsAggregation_no_limit(t, rt, instanceID) })
	t.Run("testMetricsViewAggregation_measure_filters", func(t *testing.T) { testMetricsViewAggregation_measure_filters(t, rt, instanceID) })
	t.Run("testMetricsViewsAggregation_timezone", func(t *testing.T) { testMetricsViewsAggregation_timezone(t, rt, instanceID) })
	t.Run("testMetricsViewsAggregation_filter", func(t *testing.T) { testMetricsViewsAggregation_filter(t, rt, instanceID) })
	t.Run("testMetricsViewsAggregation_filter_with_timestamp", func(t *testing.T) { testMetricsViewsAggregation_filter_with_timestamp(t, rt, instanceID) })
	t.Run("testMetricsViewsAggregation_filter_2dims", func(t *testing.T) { testMetricsViewsAggregation_filter_2dims(t, rt, instanceID) })
	t.Run("testMetricsViewsAggregation_having_gt", func(t *testing.T) { testMetricsViewsAggregation_having_gt(t, rt, instanceID) })
	t.Run("testMetricsViewsAggregation_having_same_name", func(t *testing.T) { testMetricsViewsAggregation_having_same_name(t, rt, instanceID) })
	t.Run("testMetricsViewsAggregation_having", func(t *testing.T) { testMetricsViewsAggregation_having(t, rt, instanceID) })
	t.Run("testMetricsViewsAggregation_where", func(t *testing.T) { testMetricsViewsAggregation_where(t, rt, instanceID) })
	t.Run("testMetricsViewsAggregation_whereAndSQLBoth", func(t *testing.T) { testMetricsViewsAggregation_whereAndSQLBoth(t, rt, instanceID) })
	t.Run("testMetricsViewsAggregation_filter_having_measure", func(t *testing.T) { testMetricsViewsAggregation_filter_having_measure(t, rt, instanceID) })
	t.Run("testMetricsViewsAggregation_filter_with_where_and_having_measure", func(t *testing.T) {
		testMetricsViewsAggregation_filter_with_where_and_having_measure(t, rt, instanceID)
	})
	t.Run("testMetricsViewsAggregation_2time_aggregations", func(t *testing.T) { testMetricsViewsAggregation_2time_aggregations(t, rt, instanceID) })
	t.Run("testMetricsViewsAggregation_comparison_no_time_dim", func(t *testing.T) { testMetricsViewsAggregation_comparison_no_time_dim(t, rt, instanceID) })
	t.Run("testMetricsViewsAggregation_comparison_no_dims", func(t *testing.T) { testMetricsViewsAggregation_comparison_no_dims(t, rt, instanceID) })
	t.Run("TestMetricsViewsAggregation_comparison_measure_filter_with_a_single_derivative_measure", func(t *testing.T) {
		testMetricsViewsAggregation_comparison_measure_filter_with_a_single_derivative_measure(t, rt, instanceID)
	})
	t.Run("testMetricsViewsAggregation_comparison_measure_filter_no_duplicates", func(t *testing.T) {
		testMetricsViewsAggregation_comparison_measure_filter_no_duplicates(t, rt, instanceID)
	})
	t.Run("testMetricsViewsAggregation_comparison_measure_filter_with_totals", func(t *testing.T) {
		testMetricsViewsAggregation_comparison_measure_filter_with_totals(t, rt, instanceID)
	})
	t.Run("testMetricsViewsAggregation_comparison_with_offset", func(t *testing.T) { testMetricsViewsAggregation_comparison_with_offset(t, rt, instanceID) })
	t.Run("testMetricsViewAggregation_percent_of_totals", func(t *testing.T) { testMetricsViewAggregation_percent_of_totals(t, rt, instanceID) })
	t.Run("testMetricsViewAggregation_percent_of_totals_with_limit", func(t *testing.T) { testMetricsViewAggregation_percent_of_totals_with_limit(t, rt, instanceID) })
	t.Run("testMetricsViewsAggregation_comparison_with_offset_and_limit_and_delta", func(t *testing.T) {
		testMetricsViewsAggregation_comparison_with_offset_and_limit_and_delta(t, rt, instanceID)
	})
}

func TestMetricViewAggregationAgainstDuckDB(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")
	t.Run("testMetricsViewsAggregation", func(t *testing.T) { testMetricsViewsAggregation(t, rt, instanceID) })
	t.Run("testMetricsViewsAggregationURI", func(t *testing.T) { testMetricsViewsAggregationURI(t, rt, instanceID) })
	t.Run("testMetricsViewsAggregation_export_day", func(t *testing.T) { testMetricsViewsAggregation_export_day(t, rt, instanceID) })
	t.Run("testMetricsViewsAggregation_export_hour", func(t *testing.T) { testMetricsViewsAggregation_export_hour(t, rt, instanceID) })
	t.Run("testMetricsViewsAggregation_no_limit", func(t *testing.T) { testMetricsViewsAggregation_no_limit(t, rt, instanceID) })
	t.Run("testMetricsViewAggregation_measure_filters", func(t *testing.T) { testMetricsViewAggregation_measure_filters(t, rt, instanceID) })
	t.Run("testMetricsViewsAggregation_timezone", func(t *testing.T) { testMetricsViewsAggregation_timezone(t, rt, instanceID) })
	t.Run("testMetricsViewsAggregation_filter", func(t *testing.T) { testMetricsViewsAggregation_filter(t, rt, instanceID) })
	t.Run("testMetricsViewsAggregation_filter_with_timestamp", func(t *testing.T) { testMetricsViewsAggregation_filter_with_timestamp(t, rt, instanceID) })
	t.Run("testMetricsViewsAggregation_filter_2dims", func(t *testing.T) { testMetricsViewsAggregation_filter_2dims(t, rt, instanceID) })
	t.Run("testMetricsViewsAggregation_having_gt", func(t *testing.T) { testMetricsViewsAggregation_having_gt(t, rt, instanceID) })
	t.Run("testMetricsViewsAggregation_having_same_name", func(t *testing.T) { testMetricsViewsAggregation_having_same_name(t, rt, instanceID) })
	t.Run("testMetricsViewsAggregation_having", func(t *testing.T) { testMetricsViewsAggregation_having(t, rt, instanceID) })
	t.Run("testMetricsViewsAggregation_where", func(t *testing.T) { testMetricsViewsAggregation_where(t, rt, instanceID) })
	t.Run("testMetricsViewsAggregation_whereAndSQLBoth", func(t *testing.T) { testMetricsViewsAggregation_whereAndSQLBoth(t, rt, instanceID) })
	t.Run("testMetricsViewsAggregation_filter_having_measure", func(t *testing.T) { testMetricsViewsAggregation_filter_having_measure(t, rt, instanceID) })
	t.Run("testMetricsViewsAggregation_filter_with_where_and_having_measure", func(t *testing.T) {
		testMetricsViewsAggregation_filter_with_where_and_having_measure(t, rt, instanceID)
	})
	t.Run("testMetricsViewsAggregation_2time_aggregations", func(t *testing.T) { testMetricsViewsAggregation_2time_aggregations(t, rt, instanceID) })
	t.Run("testMetricsViewsAggregation_comparison_no_time_dim", func(t *testing.T) { testMetricsViewsAggregation_comparison_no_time_dim(t, rt, instanceID) })
	t.Run("testMetricsViewsAggregation_comparison_no_dims", func(t *testing.T) { testMetricsViewsAggregation_comparison_no_dims(t, rt, instanceID) })
	t.Run("TestMetricsViewsAggregation_comparison_measure_filter_with_a_single_derivative_measure", func(t *testing.T) {
		testMetricsViewsAggregation_comparison_measure_filter_with_a_single_derivative_measure(t, rt, instanceID)
	})
	t.Run("testMetricsViewsAggregation_comparison_measure_filter_no_duplicates", func(t *testing.T) {
		testMetricsViewsAggregation_comparison_measure_filter_no_duplicates(t, rt, instanceID)
	})
	t.Run("testMetricsViewsAggregation_comparison_measure_filter_with_totals", func(t *testing.T) {
		testMetricsViewsAggregation_comparison_measure_filter_with_totals(t, rt, instanceID)
	})
	t.Run("testMetricsViewsAggregation_comparison_with_offset", func(t *testing.T) { testMetricsViewsAggregation_comparison_with_offset(t, rt, instanceID) })
	t.Run("testMetricsViewAggregation_percent_of_totals", func(t *testing.T) { testMetricsViewAggregation_percent_of_totals(t, rt, instanceID) })
	t.Run("testMetricsViewAggregation_percent_of_totals_with_limit", func(t *testing.T) { testMetricsViewAggregation_percent_of_totals_with_limit(t, rt, instanceID) })
	t.Run("testMetricsViewsAggregation_comparison_with_offset_and_limit_and_delta", func(t *testing.T) {
		testMetricsViewsAggregation_comparison_with_offset_and_limit_and_delta(t, rt, instanceID)
	})
}

func testClaims() *runtime.SecurityClaims {
	return &runtime.SecurityClaims{SkipChecks: true}
}

func testMetricsViewsAggregation(t *testing.T, rt *runtime.Runtime, instanceID string) {
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
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "pub",
			},
			{
				Name: "timestamp",
			},
		},
		Limit:          &limit,
		SecurityClaims: testClaims(),
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	for i, row := range q.Result.Data {
		for _, sf := range q.Result.Schema.Fields {
			fmt.Printf("%v ", row.Fields[sf.Name].AsInterface())
		}
		fmt.Printf(" %d \n", i)

	}
	rows := q.Result.Data

	i := 0
	require.Equal(t, "Facebook,2022-01-01T00:00:00Z", fieldsToString(rows[i], "pub", "timestamp"))
	i++
	require.Equal(t, "Facebook,2022-02-01T00:00:00Z", fieldsToString(rows[i], "pub", "timestamp"))
	i++
	require.Equal(t, "Facebook,2022-03-01T00:00:00Z", fieldsToString(rows[i], "pub", "timestamp"))
	i++
	require.Equal(t, "Google,2022-01-01T00:00:00Z", fieldsToString(rows[i], "pub", "timestamp"))
	i++
	require.Equal(t, "Google,2022-02-01T00:00:00Z", fieldsToString(rows[i], "pub", "timestamp"))
	i++
	require.Equal(t, "Google,2022-03-01T00:00:00Z", fieldsToString(rows[i], "pub", "timestamp"))
	i++
	require.Equal(t, "Microsoft,2022-01-01T00:00:00Z", fieldsToString(rows[i], "pub", "timestamp"))
	i++
	require.Equal(t, "Microsoft,2022-02-01T00:00:00Z", fieldsToString(rows[i], "pub", "timestamp"))
	i++
	require.Equal(t, "Microsoft,2022-03-01T00:00:00Z", fieldsToString(rows[i], "pub", "timestamp"))
	i++
	require.Equal(t, "Yahoo,2022-01-01T00:00:00Z", fieldsToString(rows[i], "pub", "timestamp"))
}

func testMetricsViewsAggregationURI(t *testing.T, rt *runtime.Runtime, instanceID string) {
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
				Name: "pub_uri",
				Compute: &runtimev1.MetricsViewAggregationMeasure_Uri{
					Uri: &runtimev1.MetricsViewAggregationMeasureComputeURI{
						Dimension: "pub",
					},
				},
			},
		},
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "pub",
			},
			{
				Name: "timestamp",
			},
		},
		Limit:          &limit,
		SecurityClaims: testClaims(),
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	rows := q.Result.Data

	i := 0
	require.Equal(t, "http://localhost/Facebook,2022-01-01T00:00:00Z", fieldsToString(rows[i], "pub_uri", "timestamp"))
	i++
	require.Equal(t, "http://localhost/Facebook,2022-02-01T00:00:00Z", fieldsToString(rows[i], "pub_uri", "timestamp"))
	i++
	require.Equal(t, "http://localhost/Facebook,2022-03-01T00:00:00Z", fieldsToString(rows[i], "pub_uri", "timestamp"))
	i++
	require.Equal(t, "http://localhost/Google,2022-01-01T00:00:00Z", fieldsToString(rows[i], "pub_uri", "timestamp"))
	i++
	require.Equal(t, "http://localhost/Google,2022-02-01T00:00:00Z", fieldsToString(rows[i], "pub_uri", "timestamp"))
	i++
	require.Equal(t, "http://localhost/Google,2022-03-01T00:00:00Z", fieldsToString(rows[i], "pub_uri", "timestamp"))
	i++
	require.Equal(t, "http://localhost/Microsoft,2022-01-01T00:00:00Z", fieldsToString(rows[i], "pub_uri", "timestamp"))
	i++
	require.Equal(t, "http://localhost/Microsoft,2022-02-01T00:00:00Z", fieldsToString(rows[i], "pub_uri", "timestamp"))
	i++
	require.Equal(t, "http://localhost/Microsoft,2022-03-01T00:00:00Z", fieldsToString(rows[i], "pub_uri", "timestamp"))
	i++
	require.Equal(t, "http://localhost/Yahoo,2022-01-01T00:00:00Z", fieldsToString(rows[i], "pub_uri", "timestamp"))
}

func testMetricsViewsAggregation_export_day(t *testing.T, rt *runtime.Runtime, instanceID string) {
	limit := int64(10)
	q := &queries.MetricsViewAggregation{
		MetricsViewName: "ad_bids_metrics",
		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
			{
				Name: "pub",
			},

			{
				Name:      "timestamp",
				TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_DAY,
			},
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "measure_1",
			},
		},
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "pub",
			},
			{
				Name: "timestamp",
			},
		},

		Limit:          &limit,
		Exporting:      true,
		SecurityClaims: testClaims(),
	}

	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	rows := q.Result.Data

	i := 0
	require.Equal(t, "Facebook,2022-01-01T00:00:00Z", fieldsToString(rows[i], "Publisher", "Timestamp (day)"))
}

func testMetricsViewsAggregation_export_hour(t *testing.T, rt *runtime.Runtime, instanceID string) {
	limit := int64(10)
	q := &queries.MetricsViewAggregation{
		MetricsViewName: "ad_bids_metrics",
		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
			{
				Name: "pub",
			},

			{
				Name:      "timestamp",
				TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_HOUR,
			},
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "measure_1",
			},
		},
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "pub",
			},
			{
				Name: "timestamp",
			},
		},

		Limit:          &limit,
		Exporting:      true,
		SecurityClaims: testClaims(),
	}

	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	rows := q.Result.Data

	i := 0
	require.Equal(t, "Facebook,2022-01-01T00:00:00Z", fieldsToString(rows[i], "Publisher", "Timestamp (hour)"))
}

func testMetricsViewsAggregation_no_limit(t *testing.T, rt *runtime.Runtime, instanceID string) {
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
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "pub",
			},
			{
				Name: "timestamp",
			},
		},
		SecurityClaims: testClaims(),
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	require.Equal(t, 3, len(q.Result.Schema.Fields))
	require.Equal(t, 15, len(q.Result.Data))
}

func TestMetricsViewsAggregation_no_limit_pivot(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")

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
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "pub",
			},
		},
		PivotOn:        []string{"timestamp"},
		SecurityClaims: testClaims(),
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	require.Equal(t, 4, len(q.Result.Schema.Fields))
	require.Equal(t, 5, len(q.Result.Data))
}

func TestMetricsViewsAggregation_pivot_having_same_name(t *testing.T) {
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
				Name: "bid_price",
			},
		},
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "pub",
			},
		},
		PivotOn: []string{
			"timestamp",
		},
		Limit:          &limit,
		SecurityClaims: testClaims(),
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	rows := q.Result.Data

	require.Equal(t, 4, len(q.Result.Schema.Fields))
	require.Equal(t, "pub", q.Result.Schema.Fields[0].Name)
	require.Equal(t, "2022-01-01 00:00:00_bid_price", q.Result.Schema.Fields[1].Name)
	require.Equal(t, "2022-02-01 00:00:00_bid_price", q.Result.Schema.Fields[2].Name)
	require.Equal(t, "2022-03-01 00:00:00_bid_price", q.Result.Schema.Fields[3].Name)

	i := 0
	require.Equal(t, "Facebook", fieldsToString(rows[i], "pub"))
	i++
	require.Equal(t, "Google", fieldsToString(rows[i], "pub"))
	i++
	require.Equal(t, "Microsoft", fieldsToString(rows[i], "pub"))
	i++
	require.Equal(t, "Yahoo", fieldsToString(rows[i], "pub"))
}

func TestMetricsViewsAggregation_pivot(t *testing.T) {
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
		},
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "pub",
			},
		},
		PivotOn: []string{
			"timestamp",
		},
		Limit:          &limit,
		SecurityClaims: testClaims(),
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	rows := q.Result.Data

	require.Equal(t, 4, len(q.Result.Schema.Fields))
	require.Equal(t, "pub", q.Result.Schema.Fields[0].Name)
	require.Equal(t, "2022-01-01 00:00:00_measure_1", q.Result.Schema.Fields[1].Name)
	require.Equal(t, "2022-02-01 00:00:00_measure_1", q.Result.Schema.Fields[2].Name)
	require.Equal(t, "2022-03-01 00:00:00_measure_1", q.Result.Schema.Fields[3].Name)

	i := 0
	require.Equal(t, "Facebook", fieldsToString(rows[i], "pub"))
	i++
	require.Equal(t, "Google", fieldsToString(rows[i], "pub"))
	i++
	require.Equal(t, "Microsoft", fieldsToString(rows[i], "pub"))
	i++
	require.Equal(t, "Yahoo", fieldsToString(rows[i], "pub"))
}

func TestMetricsViewsAggregation_pivot_export_labels_2_time_columns_limit_exceeded_error(t *testing.T) {
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
			{
				Name:      "timestamp",
				TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_DAY,
				Alias:     "day",
			},
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "measure_1",
			},
		},
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "pub",
			},
		},
		PivotOn: []string{
			"timestamp",
		},
		Limit:          &limit,
		Exporting:      true,
		SecurityClaims: testClaims(),
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.Len(t, q.Result.Data, 10)
}

func TestMetricsViewsAggregation_pivot_export_labels_2_time_columns(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")

	limit := int64(1000)
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
			{
				Name:      "timestamp",
				TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_DAY,
				Alias:     "day",
			},
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "measure_1",
			},
		},
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "pub",
			},
		},
		PivotOn: []string{
			"timestamp",
		},
		Limit:          &limit,
		Exporting:      true,
		SecurityClaims: testClaims(),
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	rows := q.Result.Data

	require.Equal(t, 5, len(q.Result.Schema.Fields))
	require.Equal(t, "Publisher", q.Result.Schema.Fields[0].Name)
	require.Equal(t, "Timestamp (day)", q.Result.Schema.Fields[1].Name)
	require.Equal(t, "2022-01-01 00:00:00_Average bid price", q.Result.Schema.Fields[2].Name)
	require.Equal(t, "2022-02-01 00:00:00_Average bid price", q.Result.Schema.Fields[3].Name)
	require.Equal(t, "2022-03-01 00:00:00_Average bid price", q.Result.Schema.Fields[4].Name)

	i := 0
	require.Equal(t, "Facebook", fieldsToString(rows[i], "Publisher"))
}

func TestMetricsViewsAggregation_pivot_export_labels(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")

	limit := int64(10)
	q := &queries.MetricsViewAggregation{
		MetricsViewName: "ad_bids_metrics",
		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
			{
				Name: "space_label",
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
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "space_label",
			},
		},
		PivotOn: []string{
			"timestamp",
		},
		Limit:          &limit,
		Exporting:      true,
		SecurityClaims: testClaims(),
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	rows := q.Result.Data

	require.Equal(t, 4, len(q.Result.Schema.Fields))
	require.Equal(t, "Space Label", q.Result.Schema.Fields[0].Name)
	require.Equal(t, "2022-01-01 00:00:00_Average bid price", q.Result.Schema.Fields[1].Name)
	require.Equal(t, "2022-02-01 00:00:00_Average bid price", q.Result.Schema.Fields[2].Name)
	require.Equal(t, "2022-03-01 00:00:00_Average bid price", q.Result.Schema.Fields[3].Name)

	i := 0
	require.Equal(t, "Facebook", fieldsToString(rows[i], "Space Label"))
	i++
	require.Equal(t, "Google", fieldsToString(rows[i], "Space Label"))
	i++
	require.Equal(t, "Microsoft", fieldsToString(rows[i], "Space Label"))
	i++
	require.Equal(t, "Yahoo", fieldsToString(rows[i], "Space Label"))
}

func TestMetricsViewsAggregation_pivot_export_nolabel(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")

	limit := int64(10)
	q := &queries.MetricsViewAggregation{
		MetricsViewName: "ad_bids_metrics",
		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
			{
				Name: "nolabel_pub",
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
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "nolabel_pub",
			},
		},
		PivotOn: []string{
			"timestamp",
		},
		Limit:          &limit,
		Exporting:      true,
		SecurityClaims: testClaims(),
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	rows := q.Result.Data

	require.Equal(t, 4, len(q.Result.Schema.Fields))
	require.Equal(t, "nolabel_pub", q.Result.Schema.Fields[0].Name)
	require.Equal(t, "2022-01-01 00:00:00_Average bid price", q.Result.Schema.Fields[1].Name)
	require.Equal(t, "2022-02-01 00:00:00_Average bid price", q.Result.Schema.Fields[2].Name)
	require.Equal(t, "2022-03-01 00:00:00_Average bid price", q.Result.Schema.Fields[3].Name)

	i := 0
	require.Equal(t, "Facebook", fieldsToString(rows[i], "nolabel_pub"))
	i++
	require.Equal(t, "Google", fieldsToString(rows[i], "nolabel_pub"))
	i++
	require.Equal(t, "Microsoft", fieldsToString(rows[i], "nolabel_pub"))
	i++
	require.Equal(t, "Yahoo", fieldsToString(rows[i], "nolabel_pub"))
}

func TestMetricsViewsAggregation_pivot_export_nolabel_measure(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")

	limit := int64(10)
	q := &queries.MetricsViewAggregation{
		MetricsViewName: "ad_bids_metrics",
		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
			{
				Name: "nolabel_pub",
			},
			{
				Name:      "timestamp",
				TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_MONTH,
			},
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "m1",
			},
		},
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "nolabel_pub",
			},
		},
		PivotOn: []string{
			"timestamp",
		},
		Limit:          &limit,
		Exporting:      true,
		SecurityClaims: testClaims(),
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	rows := q.Result.Data

	require.Equal(t, 4, len(q.Result.Schema.Fields))
	require.Equal(t, "nolabel_pub", q.Result.Schema.Fields[0].Name)
	require.Equal(t, "2022-01-01 00:00:00_M1", q.Result.Schema.Fields[1].Name)
	require.Equal(t, "2022-02-01 00:00:00_M1", q.Result.Schema.Fields[2].Name)
	require.Equal(t, "2022-03-01 00:00:00_M1", q.Result.Schema.Fields[3].Name)

	i := 0
	require.Equal(t, "Facebook", fieldsToString(rows[i], "nolabel_pub"))
	i++
	require.Equal(t, "Google", fieldsToString(rows[i], "nolabel_pub"))
	i++
	require.Equal(t, "Microsoft", fieldsToString(rows[i], "nolabel_pub"))
	i++
	require.Equal(t, "Yahoo", fieldsToString(rows[i], "nolabel_pub"))
}

func TestMetricsViewsAggregation_pivot_2_measures(t *testing.T) {
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
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "pub",
			},
		},
		PivotOn: []string{
			"timestamp",
		},
		Limit:          &limit,
		SecurityClaims: testClaims(),
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	for i, row := range q.Result.Data {
		for _, f := range row.Fields {
			fmt.Printf("%v ", f.AsInterface())
		}
		fmt.Printf(" %d \n", i)

	}
	rows := q.Result.Data

	require.Equal(t, q.Result.Schema.Fields[0].Name, "pub")
	require.Equal(t, q.Result.Schema.Fields[1].Name, "2022-01-01 00:00:00_measure_1")
	require.Equal(t, q.Result.Schema.Fields[2].Name, "2022-01-01 00:00:00_measure_0")

	require.Equal(t, q.Result.Schema.Fields[3].Name, "2022-02-01 00:00:00_measure_1")
	require.Equal(t, q.Result.Schema.Fields[4].Name, "2022-02-01 00:00:00_measure_0")

	require.Equal(t, q.Result.Schema.Fields[5].Name, "2022-03-01 00:00:00_measure_1")
	require.Equal(t, q.Result.Schema.Fields[6].Name, "2022-03-01 00:00:00_measure_0")

	i := 0
	require.Equal(t, "Facebook", fieldsToString(rows[i], "pub"))
	i++
	require.Equal(t, "Google", fieldsToString(rows[i], "pub"))
	i++
	require.Equal(t, "Microsoft", fieldsToString(rows[i], "pub"))
	i++
	require.Equal(t, "Yahoo", fieldsToString(rows[i], "pub"))
}

func TestMetricsViewsAggregation_pivot_2_measures_with_labels(t *testing.T) {
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
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "pub",
			},
		},
		PivotOn: []string{
			"timestamp",
		},
		Limit:          &limit,
		Exporting:      true,
		SecurityClaims: testClaims(),
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	rows := q.Result.Data

	require.Equal(t, q.Result.Schema.Fields[0].Name, "Publisher")
	require.Equal(t, q.Result.Schema.Fields[1].Name, "2022-01-01 00:00:00_Average bid price")
	require.Equal(t, q.Result.Schema.Fields[2].Name, "2022-01-01 00:00:00_Number of bids")

	require.Equal(t, q.Result.Schema.Fields[3].Name, "2022-02-01 00:00:00_Average bid price")
	require.Equal(t, q.Result.Schema.Fields[4].Name, "2022-02-01 00:00:00_Number of bids")

	require.Equal(t, q.Result.Schema.Fields[5].Name, "2022-03-01 00:00:00_Average bid price")
	require.Equal(t, q.Result.Schema.Fields[6].Name, "2022-03-01 00:00:00_Number of bids")

	i := 0
	require.Equal(t, "Facebook", fieldsToString(rows[i], "Publisher"))
	i++
	require.Equal(t, "Google", fieldsToString(rows[i], "Publisher"))
	i++
	require.Equal(t, "Microsoft", fieldsToString(rows[i], "Publisher"))
	i++
	require.Equal(t, "Yahoo", fieldsToString(rows[i], "Publisher"))
}

func TestMetricsViewsAggregation_pivot_2_measures_and_filter(t *testing.T) {
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
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "pub",
			},
		},
		PivotOn: []string{
			"timestamp",
		},
		Filter: &runtimev1.MetricsViewFilter{
			Include: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "pub",
					In:   []*structpb.Value{structpb.NewStringValue("Google")},
				},
			},
		},
		Limit:          &limit,
		SecurityClaims: testClaims(),
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	for i, row := range q.Result.Data {
		for _, f := range row.Fields {
			fmt.Printf("%v ", f.AsInterface())
		}
		fmt.Printf(" %d \n", i)

	}
	rows := q.Result.Data

	require.Equal(t, q.Result.Schema.Fields[0].Name, "pub")
	require.Equal(t, q.Result.Schema.Fields[1].Name, "2022-01-01 00:00:00_measure_1")
	require.Equal(t, q.Result.Schema.Fields[2].Name, "2022-01-01 00:00:00_measure_0")

	require.Equal(t, q.Result.Schema.Fields[3].Name, "2022-02-01 00:00:00_measure_1")
	require.Equal(t, q.Result.Schema.Fields[4].Name, "2022-02-01 00:00:00_measure_0")

	require.Equal(t, q.Result.Schema.Fields[5].Name, "2022-03-01 00:00:00_measure_1")
	require.Equal(t, q.Result.Schema.Fields[6].Name, "2022-03-01 00:00:00_measure_0")

	require.Equal(t, 1, len(rows))
	i := 0
	require.Equal(t, "Google", fieldsToString(rows[i], "pub"))
}

func TestMetricsViewsAggregation_pivot_dim_and_measure_labels(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")

	limit := int64(10)
	q := &queries.MetricsViewAggregation{
		MetricsViewName: "ad_bids_metrics",
		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
			{
				Name: "space_label",
			},
			{
				Name: "dom",
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
		Filter: &runtimev1.MetricsViewFilter{
			Include: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "space_label",
					In:   []*structpb.Value{structpb.NewStringValue("Google")},
				},
			},
		},
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "dom",
			},
		},
		PivotOn: []string{
			"timestamp",
			"space_label",
		},
		Limit:          &limit,
		Exporting:      true,
		SecurityClaims: testClaims(),
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	rows := q.Result.Data

	require.Equal(t, q.Result.Schema.Fields[0].Name, "Domain")
	require.Equal(t, q.Result.Schema.Fields[1].Name, "2022-01-01 00:00:00_Google_Average bid price")
	require.Equal(t, q.Result.Schema.Fields[2].Name, "2022-02-01 00:00:00_Google_Average bid price")
	require.Equal(t, q.Result.Schema.Fields[3].Name, "2022-03-01 00:00:00_Google_Average bid price")

	i := 0
	require.Equal(t, "google.com", fieldsToString(rows[i], "Domain"))
}

func TestMetricsViewsAggregation_pivot_dim_and_measure(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")

	limit := int64(10)
	q := &queries.MetricsViewAggregation{
		MetricsViewName: "ad_bids_metrics",
		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
			{
				Name: "pub",
			},
			{
				Name: "dom",
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
		Filter: &runtimev1.MetricsViewFilter{
			Include: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "pub",
					In:   []*structpb.Value{structpb.NewStringValue("Google")},
				},
			},
		},
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "dom",
			},
		},
		PivotOn: []string{
			"timestamp",
			"pub",
		},
		Limit:          &limit,
		SecurityClaims: testClaims(),
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	for _, s := range q.Result.Schema.Fields {
		fmt.Printf("%v ", s.Name)
	}
	for i, row := range q.Result.Data {
		for _, f := range row.Fields {
			fmt.Printf("%v ", f.AsInterface())
		}
		fmt.Printf(" %d \n", i)

	}
	rows := q.Result.Data

	require.Equal(t, q.Result.Schema.Fields[0].Name, "dom")
	require.Equal(t, q.Result.Schema.Fields[1].Name, "2022-01-01 00:00:00_Google_measure_1")
	require.Equal(t, q.Result.Schema.Fields[2].Name, "2022-02-01 00:00:00_Google_measure_1")
	require.Equal(t, q.Result.Schema.Fields[3].Name, "2022-03-01 00:00:00_Google_measure_1")

	i := 0
	require.Equal(t, "google.com", fieldsToString(rows[i], "dom"))
}

// Steps to run this test:
// 1. Unpack Druid distribution.
// 2. Run ./bin/start-micro-quickstart
// 3. Go to localhost:8888 -> Load data and index AdBids.csv as `test_data“ datasource.
// 4. Create Rill project named `rill-untitled` with `test_data`.
// 5. Run this config in VSCode:
//
//	{
//		"name": "Launch main with druid",
//		"type": "go",
//		"request": "launch",
//		"mode": "debug",
//		"program": "cli/main.go",
//		"args": [
//			"start",
//			"--no-ui",
//			"--db-driver",
//			"druid",
//			"--db",
//			"http://localhost:8082/druid/v2/sql/avatica-protobuf?authentication=BASIC&avaticaUser=1&avaticaPassword=2",
//			"rill-untitled"
//		],
//	}
//
// 4. Remove 'Ignore_' and run test.
//
// Later these tests will be integrated in CI
func Ignore_TestMetricsViewsAggregation_Druid(t *testing.T) {
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
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "publisher",
			},
			{
				Name: "__time",
			},
		},
	}

	resp, err := client.MetricsViewAggregation(context.Background(), req)
	if err != nil {
		require.NoError(t, err)
	}
	rows := resp.Data

	i := 0
	require.Equal(t, ",2022-01-01T00:00:00Z", fieldsToString(rows[i], "publisher", "__time"))
	i++
	require.Equal(t, ",2022-02-01T00:00:00Z", fieldsToString(rows[i], "publisher", "__time"))
	i++
	require.Equal(t, ",2022-03-01T00:00:00Z", fieldsToString(rows[i], "publisher", "__time"))
	i++
	require.Equal(t, "Facebook,2022-01-01T00:00:00Z", fieldsToString(rows[i], "publisher", "__time"))
	i++
	require.Equal(t, "Facebook,2022-02-01T00:00:00Z", fieldsToString(rows[i], "publisher", "__time"))
	i++
	require.Equal(t, "Facebook,2022-03-01T00:00:00Z", fieldsToString(rows[i], "publisher", "__time"))
	i++
	require.Equal(t, "Google,2022-01-01T00:00:00Z", fieldsToString(rows[i], "publisher", "__time"))
	i++
	require.Equal(t, "Google,2022-02-01T00:00:00Z", fieldsToString(rows[i], "publisher", "__time"))
	i++
	require.Equal(t, "Google,2022-03-01T00:00:00Z", fieldsToString(rows[i], "publisher", "__time"))
	i++
	require.Equal(t, "Microsoft,2022-01-01T00:00:00Z", fieldsToString(rows[i], "publisher", "__time"))
	i++
	require.Equal(t, "Microsoft,2022-02-01T00:00:00Z", fieldsToString(rows[i], "publisher", "__time"))
	i++
	require.Equal(t, "Microsoft,2022-03-01T00:00:00Z", fieldsToString(rows[i], "publisher", "__time"))
	i++
	require.Equal(t, "Yahoo,2022-01-01T00:00:00Z", fieldsToString(rows[i], "publisher", "__time"))
}

func Ignore_TestMetricsViewsAggregation_Druid_pivot(t *testing.T) {
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
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "publisher",
			},
		},
		PivotOn: []string{
			"__time",
		},
	}

	resp, err := client.MetricsViewAggregation(context.Background(), req)
	if err != nil {
		require.NoError(t, err)
	}
	rows := resp.Data

	for _, s := range resp.Schema.Fields {
		fmt.Printf("%v ", s.Name)
	}
	fmt.Println()
	for i, row := range resp.Data {
		for _, s := range resp.Schema.Fields {
			fmt.Printf("%v ", row.Fields[s.Name].AsInterface())
		}
		fmt.Printf(" %d \n", i)

	}
	require.Equal(t, 4, len(resp.Schema.Fields))
	require.Equal(t, "publisher", resp.Schema.Fields[0].Name)
	require.Equal(t, "2022-01-01 00:00:00_bp", resp.Schema.Fields[1].Name)
	require.Equal(t, "2022-02-01 00:00:00_bp", resp.Schema.Fields[2].Name)
	require.Equal(t, "2022-03-01 00:00:00_bp", resp.Schema.Fields[3].Name)

	i := 0
	require.Equal(t, "Facebook", fieldsToString(rows[i], "publisher"))
	i++
	require.Equal(t, "Google", fieldsToString(rows[i], "publisher"))
	i++
	require.Equal(t, "Microsoft", fieldsToString(rows[i], "publisher"))
	i++
	require.Equal(t, "Yahoo", fieldsToString(rows[i], "publisher"))
	i++
	require.Equal(t, "", fieldsToString(rows[i], "publisher"))
}

func Ignore_TestMetricsViewsAggregation_Druid_measure_filter(t *testing.T) {
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
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name:           "bp",
				BuiltinMeasure: runtimev1.BuiltinMeasure_BUILTIN_MEASURE_COUNT,
				Filter: &runtimev1.Expression{
					Expression: &runtimev1.Expression_Cond{
						Cond: &runtimev1.Condition{
							Op: runtimev1.Operation_OPERATION_EQ,
							Exprs: []*runtimev1.Expression{
								{
									Expression: &runtimev1.Expression_Ident{
										Ident: "domain",
									},
								},
								{
									Expression: &runtimev1.Expression_Val{
										Val: structpb.NewStringValue("news.google.com"),
									},
								},
							},
						},
					},
				},
			},
		},
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "publisher",
			},
		},
	}

	resp, err := client.MetricsViewAggregation(context.Background(), req)
	if err != nil {
		require.NoError(t, err)
	}

	rows := resp.Data
	i := 0
	require.Equal(t, "null,4239", fieldsToString(rows[i], "publisher", "bp"))
	i++
	require.Equal(t, "Facebook,null", fieldsToString(rows[i], "publisher", "bp"))
	i++
	require.Equal(t, "Google,8644", fieldsToString(rows[i], "publisher", "bp"))
	i++
	require.Equal(t, "Microsoft,null", fieldsToString(rows[i], "publisher", "bp"))
	i++
	require.Equal(t, "Yahoo,null", fieldsToString(rows[i], "publisher", "bp"))

	// check where
	req.Where = expressionpb.In(expressionpb.Identifier("publisher"), []*runtimev1.Expression{
		expressionpb.Value(structpb.NewStringValue("Google")),
		expressionpb.Value(structpb.NewStringValue("Microsoft")),
	})

	resp, err = client.MetricsViewAggregation(context.Background(), req)
	if err != nil {
		require.NoError(t, err)
	}

	rows = resp.Data
	i = 0
	require.Equal(t, "Google,8644", fieldsToString(rows[i], "publisher", "bp"))
	i++
	require.Equal(t, "Microsoft,null", fieldsToString(rows[i], "publisher", "bp"))

	// check having
	req.Having = &runtimev1.Expression{
		Expression: &runtimev1.Expression_Cond{
			Cond: &runtimev1.Condition{
				Op: runtimev1.Operation_OPERATION_GT,
				Exprs: []*runtimev1.Expression{
					{
						Expression: &runtimev1.Expression_Ident{
							Ident: "bp",
						},
					},
					{
						Expression: &runtimev1.Expression_Val{
							Val: structpb.NewNumberValue(10),
						},
					},
				},
			},
		},
	}

	resp, err = client.MetricsViewAggregation(context.Background(), req)
	if err != nil {
		require.NoError(t, err)
	}

	rows = resp.Data
	i = 0
	require.Equal(t, "Google,8644", fieldsToString(rows[i], "publisher", "bp"))

}

func testMetricsViewAggregation_measure_filters(t *testing.T, rt *runtime.Runtime, instanceID string) {
	ctr := &queries.ColumnTimeRange{
		TableName:  "ad_bids",
		ColumnName: "timestamp",
	}
	err := ctr.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	diff := ctr.Result.Max.AsTime().Sub(ctr.Result.Min.AsTime())
	maxTime := ctr.Result.Min.AsTime().Add(diff / 2)

	lmt := int64(250)
	q := &queries.MetricsViewAggregation{
		MetricsViewName: "ad_bids_metrics",
		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
			{
				Name: "dom",
			},
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "measure_1",
			},
		},
		TimeRange: &runtimev1.TimeRange{
			Start: ctr.Result.Min,
			End:   timestamppb.New(maxTime),
		},
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "dom",
				Desc: true,
			},
		},
		Limit: &lmt,
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
		SecurityClaims: testClaims(),
	}

	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	require.Len(t, q.Result.Data, 3)
	require.NotEmpty(t, "sports.yahoo.com", q.Result.Data[0].AsMap()["dom"])
	require.NotEmpty(t, "news.google.com", q.Result.Data[1].AsMap()["dom"])
	require.NotEmpty(t, "instagram.com", q.Result.Data[2].AsMap()["dom"])
}

func testMetricsViewsAggregation_timezone(t *testing.T, rt *runtime.Runtime, instanceID string) {
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
				TimeZone:  "America/New_York",
			},
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "measure_1",
			},
		},
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "pub",
			},
			{
				Name: "timestamp",
			},
		},
		Limit:          &limit,
		SecurityClaims: testClaims(),
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	rows := q.Result.Data

	i := 0
	require.Equal(t, "Facebook,2021-12-01T05:00:00Z", fieldsToString(rows[i], "pub", "timestamp"))
	i++
	require.Equal(t, "Facebook,2022-01-01T05:00:00Z", fieldsToString(rows[i], "pub", "timestamp"))
	i++
	require.Equal(t, "Facebook,2022-02-01T05:00:00Z", fieldsToString(rows[i], "pub", "timestamp"))
	i++
	require.Equal(t, "Facebook,2022-03-01T05:00:00Z", fieldsToString(rows[i], "pub", "timestamp"))
	i++
	require.Equal(t, "Google,2021-12-01T05:00:00Z", fieldsToString(rows[i], "pub", "timestamp"))
	i++
	require.Equal(t, "Google,2022-01-01T05:00:00Z", fieldsToString(rows[i], "pub", "timestamp"))
	i++
	require.Equal(t, "Google,2022-02-01T05:00:00Z", fieldsToString(rows[i], "pub", "timestamp"))
	i++
	require.Equal(t, "Google,2022-03-01T05:00:00Z", fieldsToString(rows[i], "pub", "timestamp"))
	i++
	require.Equal(t, "Microsoft,2021-12-01T05:00:00Z", fieldsToString(rows[i], "pub", "timestamp"))
	i++
	require.Equal(t, "Microsoft,2022-01-01T05:00:00Z", fieldsToString(rows[i], "pub", "timestamp"))
}

func testMetricsViewsAggregation_filter(t *testing.T, rt *runtime.Runtime, instanceID string) {
	q := &queries.MetricsViewAggregation{
		MetricsViewName: "ad_bids_metrics",
		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
			{
				Name: "pub",
			},
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name:           "inline_1",
				BuiltinMeasure: runtimev1.BuiltinMeasure_BUILTIN_MEASURE_COUNT,
			},
		},
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "pub",
			},
		},
		SecurityClaims: testClaims(),
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)

	rows := q.Result.Data
	i := 0
	require.Equal(t, "Facebook,19341", fieldsToString(rows[i], "pub", "inline_1"))
	i++
	require.Equal(t, "Google,18763", fieldsToString(rows[i], "pub", "inline_1"))
	i++
	require.Equal(t, "Microsoft,10406", fieldsToString(rows[i], "pub", "inline_1"))

	q.Measures = []*runtimev1.MetricsViewAggregationMeasure{
		{
			Name:           "inline_1",
			BuiltinMeasure: runtimev1.BuiltinMeasure_BUILTIN_MEASURE_COUNT,
			Filter: &runtimev1.Expression{
				Expression: &runtimev1.Expression_Cond{
					Cond: &runtimev1.Condition{
						Op: runtimev1.Operation_OPERATION_EQ,
						Exprs: []*runtimev1.Expression{
							{
								Expression: &runtimev1.Expression_Ident{
									Ident: "dom",
								},
							},
							{
								Expression: &runtimev1.Expression_Val{
									Val: structpb.NewStringValue("instagram.com"),
								},
							},
						},
					},
				},
			},
		},
	}

	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)

	rows = q.Result.Data
	i = 0
	require.Equal(t, "Facebook,8808", fieldsToString(rows[i], "pub", "inline_1"))
	i++
	require.Equal(t, "Google,null", fieldsToString(rows[i], "pub", "inline_1"))
	i++
	require.Equal(t, "Microsoft,null", fieldsToString(rows[i], "pub", "inline_1"))
}

func testMetricsViewsAggregation_filter_with_timestamp(t *testing.T, rt *runtime.Runtime, instanceID string) {
	q := &queries.MetricsViewAggregation{
		MetricsViewName: "ad_bids_metrics",
		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
			{
				Name: "pub",
			},
			{
				Name:      "timestamp",
				TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_DAY,
			},
			{
				Name:      "timestamp",
				TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
				Alias:     "time_year",
			},
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name:           "inline_1",
				BuiltinMeasure: runtimev1.BuiltinMeasure_BUILTIN_MEASURE_COUNT,
			},
		},
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "pub",
				Desc: true,
			},
			{
				Name: "timestamp",
			},
			{
				Name: "time_year",
			},
		},
		SecurityClaims: testClaims(),
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)

	rows := q.Result.Data
	i := 0
	require.Equal(t, "Yahoo,2022-01-01T00:00:00Z,2022-01-01T00:00:00Z,232", fieldsToString(rows[i], "pub", "timestamp", "time_year", "inline_1"))
	i++
	require.Equal(t, "Yahoo,2022-01-02T00:00:00Z,2022-01-01T00:00:00Z,208", fieldsToString(rows[i], "pub", "timestamp", "time_year", "inline_1"))

	q.Measures = []*runtimev1.MetricsViewAggregationMeasure{
		{
			Name:           "inline_1",
			BuiltinMeasure: runtimev1.BuiltinMeasure_BUILTIN_MEASURE_COUNT,
			Filter: &runtimev1.Expression{
				Expression: &runtimev1.Expression_Cond{
					Cond: &runtimev1.Condition{
						Op: runtimev1.Operation_OPERATION_EQ,
						Exprs: []*runtimev1.Expression{
							{
								Expression: &runtimev1.Expression_Ident{
									Ident: "dom",
								},
							},
							{
								Expression: &runtimev1.Expression_Val{
									Val: structpb.NewStringValue("news.yahoo.com"),
								},
							},
						},
					},
				},
			},
		},
	}

	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)

	rows = q.Result.Data
	i = 0
	require.Equal(t, "Yahoo,2022-01-01T00:00:00Z,2022-01-01T00:00:00Z,52", fieldsToString(rows[i], "pub", "timestamp", "time_year", "inline_1"))
	i++
	require.Equal(t, "Yahoo,2022-01-02T00:00:00Z,2022-01-01T00:00:00Z,54", fieldsToString(rows[i], "pub", "timestamp", "time_year", "inline_1"))
}

func testMetricsViewsAggregation_filter_2dims(t *testing.T, rt *runtime.Runtime, instanceID string) {
	q := &queries.MetricsViewAggregation{
		MetricsViewName: "ad_bids_metrics",
		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
			{
				Name: "pub",
			},
			{
				Name: "dom",
			},
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name:           "inline_1",
				BuiltinMeasure: runtimev1.BuiltinMeasure_BUILTIN_MEASURE_COUNT,
			},
		},
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "pub",
			},
			{
				Name: "dom",
			},
		},
		SecurityClaims: testClaims(),
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)

	rows := q.Result.Data
	i := 0
	require.Equal(t, "Facebook,facebook.com,10533", fieldsToString(rows[i], "pub", "dom", "inline_1"))
	i++
	require.Equal(t, "Facebook,instagram.com,8808", fieldsToString(rows[i], "pub", "dom", "inline_1"))
	i++
	require.Equal(t, "Google,google.com,10119", fieldsToString(rows[i], "pub", "dom", "inline_1"))

	q.Measures = []*runtimev1.MetricsViewAggregationMeasure{
		{
			Name:           "inline_1",
			BuiltinMeasure: runtimev1.BuiltinMeasure_BUILTIN_MEASURE_COUNT,
			Filter: &runtimev1.Expression{
				Expression: &runtimev1.Expression_Cond{
					Cond: &runtimev1.Condition{
						Op: runtimev1.Operation_OPERATION_EQ,
						Exprs: []*runtimev1.Expression{
							{
								Expression: &runtimev1.Expression_Ident{
									Ident: "dom",
								},
							},
							{
								Expression: &runtimev1.Expression_Val{
									Val: structpb.NewStringValue("instagram.com"),
								},
							},
						},
					},
				},
			},
		},
	}

	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)

	rows = q.Result.Data
	i = 0
	require.Equal(t, "Facebook,facebook.com,null", fieldsToString(rows[i], "pub", "dom", "inline_1"))
	i++
	require.Equal(t, "Facebook,instagram.com,8808", fieldsToString(rows[i], "pub", "dom", "inline_1"))
	i++
	require.Equal(t, "Google,google.com,null", fieldsToString(rows[i], "pub", "dom", "inline_1"))
}

func testMetricsViewsAggregation_having_gt(t *testing.T, rt *runtime.Runtime, instanceID string) {
	q := &queries.MetricsViewAggregation{
		MetricsViewName: "ad_bids_metrics",
		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
			{
				Name: "pub",
			},
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name:           "inline_1",
				BuiltinMeasure: runtimev1.BuiltinMeasure_BUILTIN_MEASURE_COUNT,
			},
		},
		Having: &runtimev1.Expression{
			Expression: &runtimev1.Expression_Cond{
				Cond: &runtimev1.Condition{
					Op: runtimev1.Operation_OPERATION_GT,
					Exprs: []*runtimev1.Expression{
						{
							Expression: &runtimev1.Expression_Ident{
								Ident: "inline_1",
							},
						},
						{
							Expression: &runtimev1.Expression_Val{
								Val: structpb.NewNumberValue(19000),
							},
						},
					},
				},
			},
		},
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "pub",
			},
		},
		SecurityClaims: testClaims(),
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)

	rows := q.Result.Data
	require.Equal(t, 2, len(rows))
	i := 0
	require.Equal(t, "Facebook,19341", fieldsToString(rows[i], "pub", "inline_1"))
	i++
	require.Equal(t, "null,32897", fieldsToString(rows[i], "pub", "inline_1"))
}

func testMetricsViewsAggregation_having_same_name(t *testing.T, rt *runtime.Runtime, instanceID string) {
	q := &queries.MetricsViewAggregation{
		MetricsViewName: "ad_bids_metrics",
		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
			{
				Name: "dom",
			},
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "bid_price",
			},
		},
		Having: &runtimev1.Expression{
			Expression: &runtimev1.Expression_Cond{
				Cond: &runtimev1.Condition{
					Op: runtimev1.Operation_OPERATION_GT,
					Exprs: []*runtimev1.Expression{
						{
							Expression: &runtimev1.Expression_Ident{
								Ident: "bid_price",
							},
						},
						{
							Expression: &runtimev1.Expression_Val{
								Val: structpb.NewNumberValue(3),
							},
						},
					},
				},
			},
		},
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "dom",
				Desc: true,
			},
		},
		SecurityClaims: testClaims(),
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)

	rows := q.Result.Data
	require.Equal(t, 4, len(rows))
	i := 0
	require.Equal(t, "news.yahoo.com,3", fieldsToString(rows[i], "dom", "bid_price"))
	i++
	require.Equal(t, "msn.com,3", fieldsToString(rows[i], "dom", "bid_price"))
}

func testMetricsViewsAggregation_having(t *testing.T, rt *runtime.Runtime, instanceID string) {
	q := &queries.MetricsViewAggregation{
		MetricsViewName: "ad_bids_metrics",
		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
			{
				Name: "pub",
			},
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "measure_0",
			},
		},
		Having: &runtimev1.Expression{
			Expression: &runtimev1.Expression_Cond{
				Cond: &runtimev1.Condition{
					Op: runtimev1.Operation_OPERATION_EQ,
					Exprs: []*runtimev1.Expression{
						{
							Expression: &runtimev1.Expression_Ident{
								Ident: "measure_0",
							},
						},
						{
							Expression: &runtimev1.Expression_Val{
								Val: structpb.NewNumberValue(10406),
							},
						},
					},
				},
			},
		},
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "pub",
			},
		},
		SecurityClaims: testClaims(),
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)

	rows := q.Result.Data
	require.Equal(t, 1, len(rows))
	i := 0
	require.Equal(t, "Microsoft,10406", fieldsToString(rows[i], "pub", "measure_0"))
}

func testMetricsViewsAggregation_where(t *testing.T, rt *runtime.Runtime, instanceID string) {
	q := &queries.MetricsViewAggregation{
		MetricsViewName: "ad_bids_metrics",
		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
			{
				Name: "pub",
			},
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name:           "inline_1",
				BuiltinMeasure: runtimev1.BuiltinMeasure_BUILTIN_MEASURE_COUNT,
			},
		},
		Where: &runtimev1.Expression{
			Expression: &runtimev1.Expression_Cond{
				Cond: &runtimev1.Condition{
					Op: runtimev1.Operation_OPERATION_LIKE,
					Exprs: []*runtimev1.Expression{
						{
							Expression: &runtimev1.Expression_Ident{
								Ident: "pub",
							},
						},
						{
							Expression: &runtimev1.Expression_Val{
								Val: structpb.NewStringValue("%c%"),
							},
						},
					},
				},
			},
		},
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "pub",
			},
		},
		SecurityClaims: testClaims(),
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)

	rows := q.Result.Data
	i := 0
	require.Equal(t, "Facebook,19341", fieldsToString(rows[i], "pub", "inline_1"))
	i++
	require.Equal(t, "Microsoft,10406", fieldsToString(rows[i], "pub", "inline_1"))
}

func testMetricsViewsAggregation_whereAndSQLBoth(t *testing.T, rt *runtime.Runtime, instanceID string) {
	q := &queries.MetricsViewAggregation{
		MetricsViewName: "ad_bids_metrics",
		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
			{
				Name: "pub",
			},
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name:           "inline_1",
				BuiltinMeasure: runtimev1.BuiltinMeasure_BUILTIN_MEASURE_COUNT,
			},
		},
		WhereSQL: "pub like '%cro%'",
		Where: &runtimev1.Expression{
			Expression: &runtimev1.Expression_Cond{
				Cond: &runtimev1.Condition{
					Op: runtimev1.Operation_OPERATION_LIKE,
					Exprs: []*runtimev1.Expression{
						{
							Expression: &runtimev1.Expression_Ident{
								Ident: "pub",
							},
						},
						{
							Expression: &runtimev1.Expression_Val{
								Val: structpb.NewStringValue("%c%"),
							},
						},
					},
				},
			},
		},
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "pub",
			},
		},
		SecurityClaims: testClaims(),
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)

	rows := q.Result.Data
	require.Equal(t, "Microsoft,10406", fieldsToString(rows[0], "pub", "inline_1"))
	require.Equal(t, len(q.Result.Data), 1)
}

func testMetricsViewsAggregation_filter_having_measure(t *testing.T, rt *runtime.Runtime, instanceID string) {
	q := &queries.MetricsViewAggregation{
		MetricsViewName: "ad_bids_metrics",
		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
			{
				Name: "pub",
			},
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "measure_0",
				Filter: &runtimev1.Expression{
					Expression: &runtimev1.Expression_Cond{
						Cond: &runtimev1.Condition{
							Op: runtimev1.Operation_OPERATION_EQ,
							Exprs: []*runtimev1.Expression{
								{
									Expression: &runtimev1.Expression_Ident{
										Ident: "dom",
									},
								},
								{
									Expression: &runtimev1.Expression_Val{
										Val: structpb.NewStringValue("instagram.com"),
									},
								},
							},
						},
					},
				},
			},
		},
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "pub",
			},
		},
		SecurityClaims: testClaims(),
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)

	rows := q.Result.Data
	i := 0
	require.Equal(t, "Facebook,8808", fieldsToString(rows[i], "pub", "measure_0"))
	i++
	require.Equal(t, "Google,null", fieldsToString(rows[i], "pub", "measure_0"))
	i++
	require.Equal(t, "Microsoft,null", fieldsToString(rows[i], "pub", "measure_0"))
	i++
	require.Equal(t, "Yahoo,null", fieldsToString(rows[i], "pub", "measure_0"))
	i++
	require.Equal(t, "null,4296", fieldsToString(rows[i], "pub", "measure_0"))

	// ================= check m1 > 5000

	q.Having = &runtimev1.Expression{
		Expression: &runtimev1.Expression_Cond{
			Cond: &runtimev1.Condition{
				Op: runtimev1.Operation_OPERATION_GT,
				Exprs: []*runtimev1.Expression{
					{
						Expression: &runtimev1.Expression_Ident{
							Ident: "measure_0",
						},
					},
					{
						Expression: &runtimev1.Expression_Val{
							Val: structpb.NewNumberValue(5000),
						},
					},
				},
			},
		},
	}

	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)

	rows = q.Result.Data
	require.Equal(t, 1, len(rows))
	i = 0
	require.Equal(t, "Facebook,8808", fieldsToString(rows[i], "pub", "measure_0"))

	// ================= check m1 < 5000

	q.Having = &runtimev1.Expression{
		Expression: &runtimev1.Expression_Cond{
			Cond: &runtimev1.Condition{
				Op: runtimev1.Operation_OPERATION_LT,
				Exprs: []*runtimev1.Expression{
					{
						Expression: &runtimev1.Expression_Ident{
							Ident: "measure_0",
						},
					},
					{
						Expression: &runtimev1.Expression_Val{
							Val: structpb.NewNumberValue(5000),
						},
					},
				},
			},
		},
	}

	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)

	rows = q.Result.Data
	require.Equal(t, 1, len(rows))
	i = 0
	require.Equal(t, "null,4296", fieldsToString(rows[i], "pub", "measure_0"))
}

func testMetricsViewsAggregation_filter_with_where_and_having_measure(t *testing.T, rt *runtime.Runtime, instanceID string) {
	q := &queries.MetricsViewAggregation{
		MetricsViewName: "ad_bids_metrics",
		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
			{
				Name: "pub",
			},
		},
		Where: expressionpb.In(expressionpb.Identifier("dom"), []*runtimev1.Expression{
			expressionpb.Value(structpb.NewStringValue("news.google.com")),
			expressionpb.Value(structpb.NewStringValue("msn.com")),
		}),
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "measure_0",
			},
		},
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "pub",
			},
		},
		SecurityClaims: testClaims(),
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)

	rows := q.Result.Data
	i := 0
	require.Equal(t, "Google,8644", fieldsToString(rows[i], "pub", "measure_0"))
	i++
	require.Equal(t, "Microsoft,10406", fieldsToString(rows[i], "pub", "measure_0"))
	i++
	require.Equal(t, "null,9359", fieldsToString(rows[i], "pub", "measure_0"))

	// ================= check measure filter

	q.Measures[0].Filter = &runtimev1.Expression{
		Expression: &runtimev1.Expression_Cond{
			Cond: &runtimev1.Condition{
				Op: runtimev1.Operation_OPERATION_EQ,
				Exprs: []*runtimev1.Expression{
					{
						Expression: &runtimev1.Expression_Ident{
							Ident: "dom",
						},
					},
					{
						Expression: &runtimev1.Expression_Val{
							Val: structpb.NewStringValue("news.google.com"),
						},
					},
				},
			},
		},
	}

	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)

	rows = q.Result.Data
	require.Equal(t, 3, len(rows))
	i = 0
	require.Equal(t, "Google,8644", fieldsToString(rows[i], "pub", "measure_0"))
	i++
	require.Equal(t, "Microsoft,null", fieldsToString(rows[i], "pub", "measure_0"))
	i++
	require.Equal(t, "null,4239", fieldsToString(rows[i], "pub", "measure_0"))

	// ================= check having m1 > 5000

	q.Having = &runtimev1.Expression{
		Expression: &runtimev1.Expression_Cond{
			Cond: &runtimev1.Condition{
				Op: runtimev1.Operation_OPERATION_GT,
				Exprs: []*runtimev1.Expression{
					{
						Expression: &runtimev1.Expression_Ident{
							Ident: "measure_0",
						},
					},
					{
						Expression: &runtimev1.Expression_Val{
							Val: structpb.NewNumberValue(5000),
						},
					},
				},
			},
		},
	}

	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)

	rows = q.Result.Data
	require.Equal(t, 1, len(rows))
	i = 0
	require.Equal(t, "Google,8644", fieldsToString(rows[i], "pub", "measure_0"))
}

func testMetricsViewsAggregation_2time_aggregations(t *testing.T, rt *runtime.Runtime, instanceID string) {
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
			{
				Name:      "timestamp",
				TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
				Alias:     "timestamp_year",
			},
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "measure_1",
			},
		},
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "pub",
			},
			{
				Name: "timestamp",
			},
		},
		Limit:          &limit,
		SecurityClaims: testClaims(),
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	rows := q.Result.Data

	i := 0
	require.Equal(t, "Facebook,2022-01-01T00:00:00Z,2022-01-01T00:00:00Z", fieldsToString(rows[i], "pub", "timestamp", "timestamp_year"))
	i++
	require.Equal(t, "Facebook,2022-02-01T00:00:00Z,2022-01-01T00:00:00Z", fieldsToString(rows[i], "pub", "timestamp", "timestamp_year"))
	i++
	require.Equal(t, "Facebook,2022-03-01T00:00:00Z,2022-01-01T00:00:00Z", fieldsToString(rows[i], "pub", "timestamp", "timestamp_year"))
	i++
	require.Equal(t, "Google,2022-01-01T00:00:00Z,2022-01-01T00:00:00Z", fieldsToString(rows[i], "pub", "timestamp", "timestamp_year"))
	i++
	require.Equal(t, "Google,2022-02-01T00:00:00Z,2022-01-01T00:00:00Z", fieldsToString(rows[i], "pub", "timestamp", "timestamp_year"))
}

func testMetricsViewAggregationClickhouseEnum(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "",
			"models/foo.sql": `
				SELECT
				-- Enum
				CAST('a', 'Enum(\'a\' = 1, \'b\' = 2)') as a,
				-- Nullable enum
				CAST(null, 'Nullable(Enum(\'a\' = 1, \'b\' = 2))') as b
			`,
			"metrics/bar.yaml": `
version: 1
type: metrics_view
model: foo
dimensions:
- column: a
- column: b
measures:
- name: count
  expression: count(*)
`}})

	testruntime.RequireReconcileState(t, rt, instanceID, 3, 0, 0)

	q := &queries.MetricsViewAggregation{
		MetricsViewName: "bar",
		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
			{Name: "a"},
			{Name: "b"},
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{Name: "count"},
		},
		SecurityClaims: testClaims(),
	}

	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result.Data)
	require.Equal(t, "a,null,1", fieldsToString(q.Result.Data[0], "a", "b", "count"))
}

// Uncommenting: We don't support comparison queries with time as a dimension
// func TestMetricsViewsAggregation_comparison_having_of_comparison(t *testing.T) {
// 	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")

// 	limit := int64(10)
// 	q := &queries.MetricsViewAggregation{
// 		MetricsViewName: "ad_bids_metrics",
// 		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
// 			{
// 				Name: "pub",
// 			},
// 			{
// 				Name: "dom",
// 			},

// 			{
// 				Name:      "timestamp",
// 				TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_DAY,
// 			},
// 			{
// 				Name:      "timestamp",
// 				TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
// 				Alias:     "timestamp_year",
// 			},
// 		},
// 		Measures: []*runtimev1.MetricsViewAggregationMeasure{
// 			{
// 				Name: "measure_0",
// 			},
// 			{
// 				Name: "measure_1",
// 			},
// 			{
// 				Name: "m1",
// 			},
// 			{
// 				Name: "measure_0__p",
// 				Compute: &runtimev1.MetricsViewAggregationMeasure_ComparisonValue{
// 					ComparisonValue: &runtimev1.MetricsViewAggregationMeasureComputeComparisonValue{
// 						Measure: "measure_0",
// 					},
// 				},
// 			},
// 		},
// 		Where: expressionpb.OrAll(
// 			expressionpb.Eq("pub", "Yahoo"),
// 			expressionpb.Eq("pub", "Google"),
// 		),
// 		Having: expressionpb.Gt("measure_0__p", 0.0),
// 		Sort: []*runtimev1.MetricsViewAggregationSort{
// 			{
// 				Name: "pub",
// 			},
// 			{
// 				Name: "dom",
// 			},
// 			{
// 				Name: "timestamp",
// 			},
// 			{
// 				Name: "timestamp_year",
// 			},
// 			{
// 				Name: "measure_1",
// 			},
// 		},

// 		TimeRange: &runtimev1.TimeRange{
// 			Start: timestamppb.New(time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)),
// 			End:   timestamppb.New(time.Date(2022, 1, 3, 0, 0, 0, 0, time.UTC)),
// 		},
// 		ComparisonTimeRange: &runtimev1.TimeRange{
// 			Start: timestamppb.New(time.Date(2022, 1, 3, 0, 0, 0, 0, time.UTC)),
// 			End:   timestamppb.New(time.Date(2022, 1, 5, 0, 0, 0, 0, time.UTC)),
// 		},
// 		Limit:          &limit,
// 		SecurityClaims: testClaims(),
// 	}
// 	err := q.Resolve(context.Background(), rt, instanceID, 0)
// 	require.NoError(t, err)
// 	require.NotEmpty(t, q.Result)
// 	fields := q.Result.Schema.Fields
// 	require.Equal(t, "pub,dom,timestamp,timestamp_year,measure_0,measure_1,m1,measure_0__p", columnNames(fields))
// 	i := 0

// 	for _, sf := range q.Result.Schema.Fields {
// 		fmt.Printf("%v ", sf.Name)
// 	}
// 	fmt.Printf("\n")

// 	for i, row := range q.Result.Data {
// 		for _, sf := range q.Result.Schema.Fields {
// 			fmt.Printf("%v ", row.Fields[sf.Name].AsInterface())
// 		}
// 		fmt.Printf(" %d \n", i)

// 	}
// 	rows := q.Result.Data
// 	require.Equal(t, 8, len(rows))

// 	i = 0
// 	require.Equal(t, "Google,google.com,2022-01-01T00:00:00Z,2022-01-01T00:00:00Z,44.00,50.00,1.53,1.53,2022-01-03T00:00:00Z,2022-01-01T00:00:00Z", fieldsToString2digits(rows[i], "pub", "dom", "timestamp", "timestamp_year", "measure_0", "measure_0__p", "measure_1", "m1", "timestamp__previous", "timestamp_year__previous"))
// 	i++
// 	require.Equal(t, "Google,google.com,2022-01-02T00:00:00Z,2022-01-01T00:00:00Z,62.00,51.00,1.45,1.45,2022-01-04T00:00:00Z,2022-01-01T00:00:00Z", fieldsToString2digits(rows[i], "pub", "dom", "timestamp", "timestamp_year", "measure_0", "measure_0__p", "measure_1", "m1", "timestamp__previous", "timestamp_year__previous"))
// 	i++
// 	require.Equal(t, "Google,news.google.com,2022-01-01T00:00:00Z,2022-01-01T00:00:00Z,187.00,183.00,3.55,3.55,2022-01-03T00:00:00Z,2022-01-01T00:00:00Z", fieldsToString2digits(rows[i], "pub", "dom", "timestamp", "timestamp_year", "measure_0", "measure_0__p", "measure_1", "m1", "timestamp__previous", "timestamp_year__previous"))
// }

func TestMetricsViewsAggregation_comparison_single_time_dim(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")

	limit := int64(10)
	q := &queries.MetricsViewAggregation{
		MetricsViewName: "ad_bids_metrics",
		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
			{
				Name: "pub",
			},
			{
				Name: "dom",
			},
			{
				Name:      "timestamp",
				TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_DAY,
			},
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "measure_0",
			},
			{
				Name: "measure_0__p",
				Compute: &runtimev1.MetricsViewAggregationMeasure_ComparisonValue{
					ComparisonValue: &runtimev1.MetricsViewAggregationMeasureComputeComparisonValue{
						Measure: "measure_0",
					},
				},
			},
		},
		Where: expressionpb.AndAll(
			expressionpb.Eq("pub", "Google"),
			expressionpb.Eq("dom", "news.google.com"),
		),
		Having: expressionpb.Gt("measure_0__p", 0.0),
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "pub",
			},
			{
				Name: "dom",
			},
			{
				Name: "timestamp",
			},
			{
				Name: "measure_0",
			},
		},

		TimeRange: &runtimev1.TimeRange{
			Start: timestamppb.New(time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)),
			End:   timestamppb.New(time.Date(2022, 1, 3, 0, 0, 0, 0, time.UTC)),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: timestamppb.New(time.Date(2022, 1, 3, 0, 0, 0, 0, time.UTC)),
			End:   timestamppb.New(time.Date(2022, 1, 5, 0, 0, 0, 0, time.UTC)),
		},
		Limit:          &limit,
		SecurityClaims: testClaims(),
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	fields := q.Result.Schema.Fields
	require.Equal(t, "pub,dom,timestamp,measure_0,measure_0__p", columnNames(fields))
	i := 0

	for _, sf := range q.Result.Schema.Fields {
		fmt.Printf("%v ", sf.Name)
	}
	fmt.Printf("\n")

	for i, row := range q.Result.Data {
		for _, sf := range q.Result.Schema.Fields {
			fmt.Printf("%v ", row.Fields[sf.Name].AsInterface())
		}
		fmt.Printf(" %d \n", i)

	}
	rows := q.Result.Data
	require.Equal(t, 2, len(rows))

	i = 0
	require.Equal(t, "Google,news.google.com,2022-01-01T00:00:00Z,187.00,183.00", fieldsToString2digits(rows[i], "pub", "dom", "timestamp", "timestamp_year", "measure_0", "measure_0__p", "measure_1", "m1", "timestamp__previous", "timestamp_year__previous"))
}

func TestMetricsViewsAggregation_comparison_multiple_time_dim(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")

	limit := int64(10)
	q := &queries.MetricsViewAggregation{
		MetricsViewName: "ad_bids_metrics",
		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
			{
				Name: "pub",
			},
			{
				Name: "dom",
			},

			{
				Name:      "timestamp",
				TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_DAY,
			},
			{
				Name:      "timestamp",
				TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
				Alias:     "timestamp_year",
			},
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "measure_0",
			},
			{
				Name: "measure_1",
			},
			{
				Name: "m1",
			},
			{
				Name: "measure_0__p",
				Compute: &runtimev1.MetricsViewAggregationMeasure_ComparisonValue{
					ComparisonValue: &runtimev1.MetricsViewAggregationMeasureComputeComparisonValue{
						Measure: "measure_0",
					},
				},
			},
		},
		Where: expressionpb.AndAll(
			expressionpb.Eq("pub", "Google"),
			expressionpb.Eq("dom", "news.google.com"),
		),
		Having: expressionpb.Gt("measure_0__p", 0.0),
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "pub",
			},
			{
				Name: "dom",
			},
			{
				Name: "timestamp",
			},
			{
				Name: "timestamp_year",
			},
			{
				Name: "measure_1",
			},
		},

		TimeRange: &runtimev1.TimeRange{
			Start: timestamppb.New(time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)),
			End:   timestamppb.New(time.Date(2022, 1, 3, 0, 0, 0, 0, time.UTC)),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: timestamppb.New(time.Date(2022, 1, 3, 0, 0, 0, 0, time.UTC)),
			End:   timestamppb.New(time.Date(2022, 1, 5, 0, 0, 0, 0, time.UTC)),
		},
		Limit:          &limit,
		SecurityClaims: testClaims(),
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	fields := q.Result.Schema.Fields
	require.Equal(t, "pub,dom,timestamp,timestamp_year,measure_0,measure_1,m1,measure_0__p", columnNames(fields))
	i := 0

	for _, sf := range q.Result.Schema.Fields {
		fmt.Printf("%v ", sf.Name)
	}
	fmt.Printf("\n")

	for i, row := range q.Result.Data {
		for _, sf := range q.Result.Schema.Fields {
			fmt.Printf("%v ", row.Fields[sf.Name].AsInterface())
		}
		fmt.Printf(" %d \n", i)

	}
	rows := q.Result.Data
	require.Equal(t, 2, len(rows))

	i = 0
	require.Equal(t, "Google,news.google.com,2022-01-01T00:00:00Z,2022-01-01T00:00:00Z,187.00,183.00,3.55,3.55", fieldsToString2digits(rows[i], "pub", "dom", "timestamp", "timestamp_year", "measure_0", "measure_0__p", "measure_1", "m1", "timestamp__previous", "timestamp_year__previous"))
}

func TestMetricsViewsAggregation_comparison_multiple_time_dim_cross_max_grain_boundaries(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")

	limit := int64(10)
	q := &queries.MetricsViewAggregation{
		MetricsViewName: "ad_bids_metrics",
		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
			{
				Name: "pub",
			},
			{
				Name: "dom",
			},

			{
				Name:      "timestamp",
				TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_HOUR,
			},
			{
				Name:      "timestamp",
				TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_DAY,
				Alias:     "timestamp_day",
			},
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "measure_0",
			},
			{
				Name: "measure_0__p",
				Compute: &runtimev1.MetricsViewAggregationMeasure_ComparisonValue{
					ComparisonValue: &runtimev1.MetricsViewAggregationMeasureComputeComparisonValue{
						Measure: "measure_0",
					},
				},
			},
		},
		Where: expressionpb.AndAll(
			expressionpb.Eq("pub", "Google"),
			expressionpb.Eq("dom", "news.google.com"),
		),
		Having: expressionpb.Gt("measure_0__p", 0.0),
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "pub",
			},
			{
				Name: "dom",
			},
			{
				Name: "timestamp",
			},
			{
				Name: "timestamp_day",
			},
			{
				Name: "measure_0",
			},
		},

		TimeRange: &runtimev1.TimeRange{
			Start: timestamppb.New(time.Date(2022, 1, 1, 23, 0, 0, 0, time.UTC)),
			End:   timestamppb.New(time.Date(2022, 1, 2, 1, 0, 0, 0, time.UTC)),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: timestamppb.New(time.Date(2022, 1, 2, 1, 0, 0, 0, time.UTC)),
			End:   timestamppb.New(time.Date(2022, 1, 2, 3, 0, 0, 0, time.UTC)),
		},
		Limit:          &limit,
		SecurityClaims: testClaims(),
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	fields := q.Result.Schema.Fields
	require.Equal(t, "pub,dom,timestamp,timestamp_day,measure_0,measure_0__p", columnNames(fields))
	i := 0

	for _, sf := range q.Result.Schema.Fields {
		fmt.Printf("%v ", sf.Name)
	}
	fmt.Printf("\n")

	for i, row := range q.Result.Data {
		for _, sf := range q.Result.Schema.Fields {
			fmt.Printf("%v ", row.Fields[sf.Name].AsInterface())
		}
		fmt.Printf(" %d \n", i)

	}
	rows := q.Result.Data
	require.Equal(t, 2, len(rows))

	i = 0
	require.Equal(t, "Google,news.google.com,2022-01-01T23:00:00Z,2022-01-01T00:00:00Z,9.00,3.00", fieldsToString2digits(rows[i], "pub", "dom", "timestamp", "timestamp_day", "measure_0", "measure_0__p"))
	i++
	require.Equal(t, "Google,news.google.com,2022-01-02T00:00:00Z,2022-01-02T00:00:00Z,4.00,16.00", fieldsToString2digits(rows[i], "pub", "dom", "timestamp", "timestamp_day", "measure_0", "measure_0__p"))

}

func TestMetricsViewsAggregation_Druid_comparison_no_time_dim(t *testing.T) {
	rt, instanceID, err := testruntime.NewInstanceForDruidProject(t)
	require.NoError(t, err)

	limit := int64(10)
	q := &queries.MetricsViewAggregation{
		MetricsViewName: "ad_bids_metrics",
		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
			{
				Name: "pub",
			},
			{
				Name: "dom",
			},
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "measure_0",
			},
			{
				Name: "measure_1",
			},
			{
				Name: "m1",
			},
			{
				Name: "measure_0__p",
				Compute: &runtimev1.MetricsViewAggregationMeasure_ComparisonValue{
					ComparisonValue: &runtimev1.MetricsViewAggregationMeasureComputeComparisonValue{
						Measure: "measure_0",
					},
				},
			},
		},
		Where: expressionpb.OrAll(
			expressionpb.Eq("pub", "Yahoo"),
			expressionpb.Eq("pub", "Google"),
		),
		Having: expressionpb.Gt("measure_1", 0.0),
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "pub",
			},
			{
				Name: "dom",
			},
			{
				Name: "measure_1",
			},
		},
		TimeRange: &runtimev1.TimeRange{
			Start: timestamppb.New(time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)),
			End:   timestamppb.New(time.Date(2022, 1, 3, 0, 0, 0, 0, time.UTC)),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: timestamppb.New(time.Date(2022, 1, 3, 0, 0, 0, 0, time.UTC)),
			End:   timestamppb.New(time.Date(2022, 1, 5, 0, 0, 0, 0, time.UTC)),
		},
		Limit:          &limit,
		SecurityClaims: testClaims(),
	}
	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	fields := q.Result.Schema.Fields
	require.Equal(t, "pub,dom,measure_0,measure_1,m1,measure_0__p", columnNames(fields))
	i := 0

	for _, sf := range q.Result.Schema.Fields {
		fmt.Printf("%v ", sf.Name)
	}
	fmt.Printf("\n")

	for i, row := range q.Result.Data {
		for _, sf := range q.Result.Schema.Fields {
			fmt.Printf("%v ", row.Fields[sf.Name].AsInterface())
		}
		fmt.Printf(" %d \n", i)

	}
	rows := q.Result.Data
	require.Equal(t, 4, len(rows))

	i = 0
	require.Equal(t, "Google,google.com,106.00,101.00,1.48,1.48", fieldsToString2digits(rows[i], "pub", "dom", "measure_0", "measure_0__p", "measure_1", "m1"))
	i++
	require.Equal(t, "Google,news.google.com,381.00,372.00,3.65,3.65", fieldsToString2digits(rows[i], "pub", "dom", "measure_0", "measure_0__p", "measure_1", "m1"))
	i++
	require.Equal(t, "Yahoo,news.yahoo.com,106.00,106.00,1.50,1.50", fieldsToString2digits(rows[i], "pub", "dom", "measure_0", "measure_0__p", "measure_1", "m1"))
}

func testMetricsViewsAggregation_comparison_no_time_dim(t *testing.T, rt *runtime.Runtime, instanceID string) {
	limit := int64(10)
	q := &queries.MetricsViewAggregation{
		MetricsViewName: "ad_bids_metrics",
		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
			{
				Name: "pub",
			},
			{
				Name: "dom",
			},
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "measure_0",
			},
			{
				Name: "measure_1",
			},
			{
				Name: "m1",
			},
			{
				Name: "measure_0__p",
				Compute: &runtimev1.MetricsViewAggregationMeasure_ComparisonValue{
					ComparisonValue: &runtimev1.MetricsViewAggregationMeasureComputeComparisonValue{
						Measure: "measure_0",
					},
				},
			},
		},
		Where: expressionpb.OrAll(
			expressionpb.Eq("pub", "Yahoo"),
			expressionpb.Eq("pub", "Google"),
		),
		Having: expressionpb.Gt("measure_1", 0.0),
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "pub",
			},
			{
				Name: "dom",
			},
			{
				Name: "measure_1",
			},
		},

		TimeRange: &runtimev1.TimeRange{
			Start: timestamppb.New(time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)),
			End:   timestamppb.New(time.Date(2022, 1, 3, 0, 0, 0, 0, time.UTC)),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: timestamppb.New(time.Date(2022, 1, 3, 0, 0, 0, 0, time.UTC)),
			End:   timestamppb.New(time.Date(2022, 1, 5, 0, 0, 0, 0, time.UTC)),
		},
		Limit:          &limit,
		SecurityClaims: testClaims(),
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	fields := q.Result.Schema.Fields
	require.Equal(t, "pub,dom,measure_0,measure_1,m1,measure_0__p", columnNames(fields))
	i := 0

	for _, sf := range q.Result.Schema.Fields {
		fmt.Printf("%v ", sf.Name)
	}
	fmt.Printf("\n")

	for i, row := range q.Result.Data {
		for _, sf := range q.Result.Schema.Fields {
			fmt.Printf("%v ", row.Fields[sf.Name].AsInterface())
		}
		fmt.Printf(" %d \n", i)

	}
	rows := q.Result.Data
	require.Equal(t, 4, len(rows))

	i = 0
	require.Equal(t, "Google,google.com,106.00,101.00,1.48,1.48", fieldsToString2digits(rows[i], "pub", "dom", "measure_0", "measure_0__p", "measure_1", "m1"))
	i++
	require.Equal(t, "Google,news.google.com,381.00,372.00,3.65,3.65", fieldsToString2digits(rows[i], "pub", "dom", "measure_0", "measure_0__p", "measure_1", "m1"))
	i++
	require.Equal(t, "Yahoo,news.yahoo.com,106.00,106.00,1.50,1.50", fieldsToString2digits(rows[i], "pub", "dom", "measure_0", "measure_0__p", "measure_1", "m1"))
}

func TestMetricsViewsAggregation_comparison_Druid_no_dims(t *testing.T) {
	rt, instanceID, err := testruntime.NewInstanceForDruidProject(t)
	require.NoError(t, err)

	limit := int64(10)
	q := &queries.MetricsViewAggregation{
		MetricsViewName: "ad_bids_metrics",
		Dimensions:      []*runtimev1.MetricsViewAggregationDimension{},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "measure_0",
			},
			{
				Name: "measure_1",
			},
			{
				Name: "m1",
			},
			{
				Name: "measure_0__p",
				Compute: &runtimev1.MetricsViewAggregationMeasure_ComparisonValue{
					ComparisonValue: &runtimev1.MetricsViewAggregationMeasureComputeComparisonValue{
						Measure: "measure_0",
					},
				},
			},
		},
		Where: expressionpb.OrAll(
			expressionpb.Eq("pub", "Yahoo"),
			expressionpb.Eq("pub", "Google"),
		),
		Having: expressionpb.Gt("measure_1", 0.0),
		TimeRange: &runtimev1.TimeRange{
			Start: timestamppb.New(time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)),
			End:   timestamppb.New(time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC)),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: timestamppb.New(time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC)),
			End:   timestamppb.New(time.Date(2022, 1, 3, 0, 0, 0, 0, time.UTC)),
		},
		Limit:          &limit,
		SecurityClaims: testClaims(),
	}
	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	fields := q.Result.Schema.Fields
	require.Equal(t, "measure_0,measure_1,m1,measure_0__p", columnNames(fields))
	i := 0

	for _, sf := range q.Result.Schema.Fields {
		fmt.Printf("%v ", sf.Name)
	}
	fmt.Printf("\n")

	for i, row := range q.Result.Data {
		for _, sf := range q.Result.Schema.Fields {
			fmt.Printf("%v ", row.Fields[sf.Name].AsInterface())
		}
		fmt.Printf(" %d \n", i)

	}
	rows := q.Result.Data
	require.Equal(t, 1, len(rows))

	i = 0
	require.Equal(t, "463.00,464.00,3.20,3.20", fieldsToString2digits(rows[i], "measure_0", "measure_0__p", "measure_1", "m1"))
}

func testMetricsViewsAggregation_comparison_no_dims(t *testing.T, rt *runtime.Runtime, instanceID string) {
	limit := int64(10)
	q := &queries.MetricsViewAggregation{
		MetricsViewName: "ad_bids_metrics",
		Dimensions:      []*runtimev1.MetricsViewAggregationDimension{},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "measure_0",
			},
			{
				Name: "measure_1",
			},
			{
				Name: "m1",
			},
			{
				Name: "measure_0__p",
				Compute: &runtimev1.MetricsViewAggregationMeasure_ComparisonValue{
					ComparisonValue: &runtimev1.MetricsViewAggregationMeasureComputeComparisonValue{
						Measure: "measure_0",
					},
				},
			},
		},
		Where: expressionpb.OrAll(
			expressionpb.Eq("pub", "Yahoo"),
			expressionpb.Eq("pub", "Google"),
		),
		Having: expressionpb.Gt("measure_1", 0.0),
		TimeRange: &runtimev1.TimeRange{
			Start: timestamppb.New(time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)),
			End:   timestamppb.New(time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC)),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: timestamppb.New(time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC)),
			End:   timestamppb.New(time.Date(2022, 1, 3, 0, 0, 0, 0, time.UTC)),
		},
		Limit:          &limit,
		SecurityClaims: testClaims(),
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	fields := q.Result.Schema.Fields
	require.Equal(t, "measure_0,measure_1,m1,measure_0__p", columnNames(fields))
	i := 0

	for _, sf := range q.Result.Schema.Fields {
		fmt.Printf("%v ", sf.Name)
	}
	fmt.Printf("\n")

	for i, row := range q.Result.Data {
		for _, sf := range q.Result.Schema.Fields {
			fmt.Printf("%v ", row.Fields[sf.Name].AsInterface())
		}
		fmt.Printf(" %d \n", i)

	}
	rows := q.Result.Data
	require.Equal(t, 1, len(rows))

	i = 0
	require.Equal(t, "463.00,464.00,3.20,3.20", fieldsToString2digits(rows[i], "measure_0", "measure_0__p", "measure_1", "m1"))
}

func TestMetricsViewsAggregation_Druid_comparison_measure_filter_with_totals(t *testing.T) {
	rt, instanceID, err := testruntime.NewInstanceForDruidProject(t)
	require.NoError(t, err)

	limit := int64(10)
	q := &queries.MetricsViewAggregation{
		MetricsViewName: "ad_bids_metrics",
		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
			{
				Name: "pub",
			},
			{
				Name: "dom",
			},
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "m1_p",
				Compute: &runtimev1.MetricsViewAggregationMeasure_ComparisonValue{
					ComparisonValue: &runtimev1.MetricsViewAggregationMeasureComputeComparisonValue{
						Measure: "m1",
					},
				},
				Filter: &runtimev1.Expression{
					Expression: &runtimev1.Expression_Cond{
						Cond: &runtimev1.Condition{
							Op: runtimev1.Operation_OPERATION_EQ,
							Exprs: []*runtimev1.Expression{
								{
									Expression: &runtimev1.Expression_Ident{
										Ident: "dom",
									},
								},
								{
									Expression: &runtimev1.Expression_Val{
										Val: structpb.NewStringValue("news.google.com"),
									},
								},
							},
						},
					},
				},
			},
		},
		Where: expressionpb.OrAll(
			expressionpb.Eq("pub", "Yahoo"),
			expressionpb.Eq("pub", "Google"),
		),
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "pub",
			},
			{
				Name: "dom",
			},
			{
				Name: "m1_p",
			},
		},
		TimeRange: &runtimev1.TimeRange{
			Start: timestamppb.New(time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)),
			End:   timestamppb.New(time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC)),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: timestamppb.New(time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC)),
			End:   timestamppb.New(time.Date(2022, 1, 3, 0, 0, 0, 0, time.UTC)),
		},
		Limit:          &limit,
		SecurityClaims: testClaims(),
	}
	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	fields := q.Result.Schema.Fields
	require.Equal(t, "pub,dom,m1_p", columnNames(fields))
	i := 0

	for _, sf := range q.Result.Schema.Fields {
		fmt.Printf("%v ", sf.Name)
	}
	fmt.Printf("\n")

	for i, row := range q.Result.Data {
		for _, sf := range q.Result.Schema.Fields {
			fmt.Printf("%v ", row.Fields[sf.Name].AsInterface())
		}
		fmt.Printf(" %d \n", i)

	}
	rows := q.Result.Data
	require.Equal(t, 4, len(rows))

	i = 0
	require.Equal(t, "Google,google.com,null", fieldsToString2digits(rows[i], "pub", "dom", "m1_p"))
	i++
	require.Equal(t, "Google,news.google.com,3.74", fieldsToString2digits(rows[i], "pub", "dom", "m1_p"))
	i++
	require.Equal(t, "Yahoo,news.yahoo.com,null", fieldsToString2digits(rows[i], "pub", "dom", "m1_p"))
	i++
	require.Equal(t, "Yahoo,sports.yahoo.com,null", fieldsToString2digits(rows[i], "pub", "dom", "m1_p"))
}

func testMetricsViewsAggregation_comparison_measure_filter_with_a_single_derivative_measure(t *testing.T, rt *runtime.Runtime, instanceID string) {
	limit := int64(10)
	q := &queries.MetricsViewAggregation{
		MetricsViewName: "ad_bids_metrics",
		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
			{
				Name: "pub",
			},
			{
				Name: "dom",
			},
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "m1_p",
				Compute: &runtimev1.MetricsViewAggregationMeasure_ComparisonValue{
					ComparisonValue: &runtimev1.MetricsViewAggregationMeasureComputeComparisonValue{
						Measure: "m1",
					},
				},
				Filter: &runtimev1.Expression{
					Expression: &runtimev1.Expression_Cond{
						Cond: &runtimev1.Condition{
							Op: runtimev1.Operation_OPERATION_EQ,
							Exprs: []*runtimev1.Expression{
								{
									Expression: &runtimev1.Expression_Ident{
										Ident: "dom",
									},
								},
								{
									Expression: &runtimev1.Expression_Val{
										Val: structpb.NewStringValue("news.google.com"),
									},
								},
							},
						},
					},
				},
			},
		},
		Where: expressionpb.OrAll(
			expressionpb.Eq("pub", "Yahoo"),
			expressionpb.Eq("pub", "Google"),
		),
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "pub",
			},
			{
				Name: "dom",
			},
			{
				Name: "m1_p",
			},
		},
		TimeRange: &runtimev1.TimeRange{
			Start: timestamppb.New(time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)),
			End:   timestamppb.New(time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC)),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: timestamppb.New(time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC)),
			End:   timestamppb.New(time.Date(2022, 1, 3, 0, 0, 0, 0, time.UTC)),
		},
		Limit:          &limit,
		SecurityClaims: testClaims(),
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	fields := q.Result.Schema.Fields
	require.Equal(t, "pub,dom,m1_p", columnNames(fields))
	i := 0

	for _, sf := range q.Result.Schema.Fields {
		fmt.Printf("%v ", sf.Name)
	}
	fmt.Printf("\n")

	for i, row := range q.Result.Data {
		for _, sf := range q.Result.Schema.Fields {
			fmt.Printf("%v ", row.Fields[sf.Name].AsInterface())
		}
		fmt.Printf(" %d \n", i)

	}
	rows := q.Result.Data
	require.Equal(t, 4, len(rows))

	i = 0
	require.Equal(t, "Google,google.com,null", fieldsToString2digits(rows[i], "pub", "dom", "m1_p"))
	i++
	require.Equal(t, "Google,news.google.com,3.74", fieldsToString2digits(rows[i], "pub", "dom", "m1_p"))
	i++
	require.Equal(t, "Yahoo,news.yahoo.com,null", fieldsToString2digits(rows[i], "pub", "dom", "m1_p"))
	i++
	require.Equal(t, "Yahoo,sports.yahoo.com,null", fieldsToString2digits(rows[i], "pub", "dom", "m1_p"))
}

func TestMetricsViewsAggregation_Druid_comparison_measure_filter_no_duplicates(t *testing.T) {
	rt, instanceID, err := testruntime.NewInstanceForDruidProject(t)
	require.NoError(t, err)

	limit := int64(10)
	q := &queries.MetricsViewAggregation{
		MetricsViewName: "ad_bids_metrics",
		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
			{
				Name: "pub",
			},
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			// {
			// 	Name: "m1",
			// },
			{
				Name: "m1_p",
				Compute: &runtimev1.MetricsViewAggregationMeasure_ComparisonValue{
					ComparisonValue: &runtimev1.MetricsViewAggregationMeasureComputeComparisonValue{
						Measure: "m1",
					},
				},
				Filter: &runtimev1.Expression{
					Expression: &runtimev1.Expression_Cond{
						Cond: &runtimev1.Condition{
							Op: runtimev1.Operation_OPERATION_EQ,
							Exprs: []*runtimev1.Expression{
								{
									Expression: &runtimev1.Expression_Ident{
										Ident: "dom",
									},
								},
								{
									Expression: &runtimev1.Expression_Val{
										Val: structpb.NewStringValue("news.google.com"),
									},
								},
							},
						},
					},
				},
			},
		},
		Where: expressionpb.OrAll(
			expressionpb.Eq("pub", "Yahoo"),
			expressionpb.Eq("pub", "Google"),
		),
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "pub",
			},
			{
				Name: "m1_p",
			},
		},
		TimeRange: &runtimev1.TimeRange{
			Start: timestamppb.New(time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)),
			End:   timestamppb.New(time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC)),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: timestamppb.New(time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC)),
			End:   timestamppb.New(time.Date(2022, 1, 3, 0, 0, 0, 0, time.UTC)),
		},
		Limit:          &limit,
		SecurityClaims: testClaims(),
	}
	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	fields := q.Result.Schema.Fields
	require.Equal(t, "pub,m1_p", columnNames(fields))
	i := 0

	for _, sf := range q.Result.Schema.Fields {
		fmt.Printf("%v ", sf.Name)
	}
	fmt.Printf("\n")

	for i, row := range q.Result.Data {
		for _, sf := range q.Result.Schema.Fields {
			fmt.Printf("%v ", row.Fields[sf.Name].AsInterface())
		}
		fmt.Printf(" %d \n", i)

	}
	rows := q.Result.Data
	require.Equal(t, 2, len(rows))

	i = 0
	require.Equal(t, "Google,3.74", fieldsToString2digits(rows[i], "pub", "m1_p"))
	i++
	require.Equal(t, "Yahoo,null", fieldsToString2digits(rows[i], "pub", "m1_p"))
}

func testMetricsViewsAggregation_comparison_measure_filter_no_duplicates(t *testing.T, rt *runtime.Runtime, instanceID string) {
	limit := int64(10)
	q := &queries.MetricsViewAggregation{
		MetricsViewName: "ad_bids_metrics",
		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
			{
				Name: "pub",
			},
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "m1_p",
				Compute: &runtimev1.MetricsViewAggregationMeasure_ComparisonValue{
					ComparisonValue: &runtimev1.MetricsViewAggregationMeasureComputeComparisonValue{
						Measure: "m1",
					},
				},
				Filter: &runtimev1.Expression{
					Expression: &runtimev1.Expression_Cond{
						Cond: &runtimev1.Condition{
							Op: runtimev1.Operation_OPERATION_EQ,
							Exprs: []*runtimev1.Expression{
								{
									Expression: &runtimev1.Expression_Ident{
										Ident: "dom",
									},
								},
								{
									Expression: &runtimev1.Expression_Val{
										Val: structpb.NewStringValue("news.google.com"),
									},
								},
							},
						},
					},
				},
			},
		},
		Where: expressionpb.OrAll(
			expressionpb.Eq("pub", "Yahoo"),
			expressionpb.Eq("pub", "Google"),
		),
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "pub",
			},
			{
				Name: "m1_p",
			},
		},
		TimeRange: &runtimev1.TimeRange{
			Start: timestamppb.New(time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)),
			End:   timestamppb.New(time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC)),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: timestamppb.New(time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC)),
			End:   timestamppb.New(time.Date(2022, 1, 3, 0, 0, 0, 0, time.UTC)),
		},
		Limit:          &limit,
		SecurityClaims: testClaims(),
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	fields := q.Result.Schema.Fields
	require.Equal(t, "pub,m1_p", columnNames(fields))
	i := 0

	for _, sf := range q.Result.Schema.Fields {
		fmt.Printf("%v ", sf.Name)
	}
	fmt.Printf("\n")

	for i, row := range q.Result.Data {
		for _, sf := range q.Result.Schema.Fields {
			fmt.Printf("%v ", row.Fields[sf.Name].AsInterface())
		}
		fmt.Printf(" %d \n", i)

	}
	rows := q.Result.Data
	require.Equal(t, 2, len(rows))

	i = 0
	require.Equal(t, "Google,3.74", fieldsToString2digits(rows[i], "pub", "m1_p"))
	i++
	require.Equal(t, "Yahoo,null", fieldsToString2digits(rows[i], "pub", "m1_p"))
}

func testMetricsViewsAggregation_comparison_measure_filter_with_totals(t *testing.T, rt *runtime.Runtime, instanceID string) {
	limit := int64(10)
	q := &queries.MetricsViewAggregation{
		MetricsViewName: "ad_bids_metrics",
		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
			{
				Name: "pub",
			},
			{
				Name: "dom",
			},
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "m1_p",
				Compute: &runtimev1.MetricsViewAggregationMeasure_ComparisonValue{
					ComparisonValue: &runtimev1.MetricsViewAggregationMeasureComputeComparisonValue{
						Measure: "m1",
					},
				},
				Filter: &runtimev1.Expression{
					Expression: &runtimev1.Expression_Cond{
						Cond: &runtimev1.Condition{
							Op: runtimev1.Operation_OPERATION_EQ,
							Exprs: []*runtimev1.Expression{
								{
									Expression: &runtimev1.Expression_Ident{
										Ident: "dom",
									},
								},
								{
									Expression: &runtimev1.Expression_Val{
										Val: structpb.NewStringValue("news.google.com"),
									},
								},
							},
						},
					},
				},
			},
		},
		Where: expressionpb.OrAll(
			expressionpb.Eq("pub", "Yahoo"),
			expressionpb.Eq("pub", "Google"),
		),
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "pub",
			},
			{
				Name: "dom",
			},
			{
				Name: "m1_p",
			},
		},
		TimeRange: &runtimev1.TimeRange{
			Start: timestamppb.New(time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)),
			End:   timestamppb.New(time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC)),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: timestamppb.New(time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC)),
			End:   timestamppb.New(time.Date(2022, 1, 3, 0, 0, 0, 0, time.UTC)),
		},
		Limit:          &limit,
		SecurityClaims: testClaims(),
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	fields := q.Result.Schema.Fields
	require.Equal(t, "pub,dom,m1_p", columnNames(fields))
	i := 0

	for _, sf := range q.Result.Schema.Fields {
		fmt.Printf("%v ", sf.Name)
	}
	fmt.Printf("\n")

	for i, row := range q.Result.Data {
		for _, sf := range q.Result.Schema.Fields {
			fmt.Printf("%v ", row.Fields[sf.Name].AsInterface())
		}
		fmt.Printf(" %d \n", i)

	}
	rows := q.Result.Data
	require.Equal(t, 4, len(rows))

	i = 0
	require.Equal(t, "Google,google.com,null", fieldsToString2digits(rows[i], "pub", "dom", "m1_p"))
	i++
	require.Equal(t, "Google,news.google.com,3.74", fieldsToString2digits(rows[i], "pub", "dom", "m1_p"))
	i++
	require.Equal(t, "Yahoo,news.yahoo.com,null", fieldsToString2digits(rows[i], "pub", "dom", "m1_p"))
	i++
	require.Equal(t, "Yahoo,sports.yahoo.com,null", fieldsToString2digits(rows[i], "pub", "dom", "m1_p"))
}

// Uncommenting: We don't support comparison queries with time as a dimension
// func TestMetricsViewsAggregation_Druid_comparison_measure_filter_with_limit(t *testing.T) {
// 	if os.Getenv("METRICS_CREDS") == "" {
// 		t.Skip("skipping the test without the test instance")
// 	}
// 	rt, instanceID, err := testruntime.NewInstanceForDruidProject(t)
// 	require.NoError(t, err)

// 	limit := int64(3)
// 	q := &queries.MetricsViewAggregation{
// 		MetricsViewName: "ad_bids_metrics",
// 		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
// 			{
// 				Name: "pub",
// 			},
// 			{
// 				Name: "dom",
// 			},
// 		},
// 		Measures: []*runtimev1.MetricsViewAggregationMeasure{
// 			{
// 				Name: "m1",
// 			},
// 			{
// 				Name: "m1_p",
// 				Compute: &runtimev1.MetricsViewAggregationMeasure_ComparisonValue{
// 					ComparisonValue: &runtimev1.MetricsViewAggregationMeasureComputeComparisonValue{
// 						Measure: "m1",
// 					},
// 				},
// 				Filter: &runtimev1.Expression{
// 					Expression: &runtimev1.Expression_Cond{
// 						Cond: &runtimev1.Condition{
// 							Op: runtimev1.Operation_OPERATION_EQ,
// 							Exprs: []*runtimev1.Expression{
// 								{
// 									Expression: &runtimev1.Expression_Ident{
// 										Ident: "dom",
// 									},
// 								},
// 								{
// 									Expression: &runtimev1.Expression_Val{
// 										Val: structpb.NewStringValue("news.google.com"),
// 									},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 		},
// 		// Having: expressionpb.Gt("m1", 0.0),
// 		Sort: []*runtimev1.MetricsViewAggregationSort{
// 			{
// 				Name: "m1",
// 				Desc: true,
// 			},
// 			{
// 				Name: "pub",
// 				Desc: true,
// 			},
// 			{
// 				Name: "dom",
// 			},
// 		},

// 		TimeRange: &runtimev1.TimeRange{
// 			Start: timestamppb.New(time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)),
// 			End:   timestamppb.New(time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC)),
// 		},
// 		ComparisonTimeRange: &runtimev1.TimeRange{
// 			Start: timestamppb.New(time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC)),
// 			End:   timestamppb.New(time.Date(2022, 1, 3, 0, 0, 0, 0, time.UTC)),
// 		},
// 		Limit:  &limit,
// 		Offset: 1,
// 	}
// 	err = q.Resolve(context.Background(), rt, instanceID, 0)
// 	require.NoError(t, err)
// 	require.NotEmpty(t, q.Result)
// 	fields := q.Result.Schema.Fields
// 	require.Equal(t, "pub,dom,m1,m1_p", columnNames(fields))
// 	i := 0

// 	for _, sf := range q.Result.Schema.Fields {
// 		fmt.Printf("%v ", sf.Name)
// 	}
// 	fmt.Printf("\n")

// 	for i, row := range q.Result.Data {
// 		for _, sf := range q.Result.Schema.Fields {
// 			fmt.Printf("%v ", row.Fields[sf.Name].AsInterface())
// 		}
// 		fmt.Printf(" %d \n", i)

// 	}
// 	rows := q.Result.Data
// 	// require.Equal(t, 2, len(rows))

// 	i = 0
// 	require.Equal(t, "Google,news.google.com,3.55", fieldsToString2digits(rows[i], "pub", "dom", "m1"))
// 	i++
// 	require.Equal(t, "Yahoo,news.yahoo.com,null", fieldsToString2digits(rows[i], "pub", "dom", "m1"))
// 	i++
// 	require.Equal(t, "Yahoo,sports.yahoo.com,null", fieldsToString2digits(rows[i], "pub", "dom", "m1"))
// }

// Uncommenting: We don't support comparison queries with time as a dimension
// func TestMetricsViewsAggregation_comparison_measure_filter(t *testing.T) {
// 	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")

// 	limit := int64(10)
// 	q := &queries.MetricsViewAggregation{
// 		MetricsViewName: "ad_bids_metrics",
// 		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
// 			{
// 				Name: "pub",
// 			},
// 			{
// 				Name: "dom",
// 			},

// 			{
// 				Name:      "timestamp",
// 				TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_DAY,
// 			},
// 			{
// 				Name:      "timestamp",
// 				TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
// 				Alias:     "timestamp_year",
// 			},
// 		},
// 		Measures: []*runtimev1.MetricsViewAggregationMeasure{
// 			{
// 				Name: "m1",
// 			},
// 			{
// 				Name: "m1_p",
// 				Compute: &runtimev1.MetricsViewAggregationMeasure_ComparisonValue{
// 					ComparisonValue: &runtimev1.MetricsViewAggregationMeasureComputeComparisonValue{
// 						Measure: "m1",
// 					},
// 				},
// 				Filter: &runtimev1.Expression{
// 					Expression: &runtimev1.Expression_Cond{
// 						Cond: &runtimev1.Condition{
// 							Op: runtimev1.Operation_OPERATION_EQ,
// 							Exprs: []*runtimev1.Expression{
// 								{
// 									Expression: &runtimev1.Expression_Ident{
// 										Ident: "dom",
// 									},
// 								},
// 								{
// 									Expression: &runtimev1.Expression_Val{
// 										Val: structpb.NewStringValue("news.google.com"),
// 									},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 		},
// 		Where: expressionpb.OrAll(
// 			expressionpb.Eq("pub", "Yahoo"),
// 			expressionpb.Eq("pub", "Google"),
// 		),
// 		// Having: expressionpb.Gt("m1", 0.0),
// 		Sort: []*runtimev1.MetricsViewAggregationSort{
// 			{
// 				Name: "timestamp",
// 			},
// 			{
// 				Name: "pub",
// 			},
// 			{
// 				Name: "dom",
// 			},
// 			{
// 				Name: "timestamp_year",
// 			},
// 			{
// 				Name: "m1_p",
// 			},
// 		},

// 		TimeRange: &runtimev1.TimeRange{
// 			Start: timestamppb.New(time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)),
// 			End:   timestamppb.New(time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC)),
// 		},
// 		ComparisonTimeRange: &runtimev1.TimeRange{
// 			Start: timestamppb.New(time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC)),
// 			End:   timestamppb.New(time.Date(2022, 1, 3, 0, 0, 0, 0, time.UTC)),
// 		},
// 		Limit: &limit,
// 	}
// 	err := q.Resolve(context.Background(), rt, instanceID, 0)
// 	require.NoError(t, err)
// 	require.NotEmpty(t, q.Result)
// 	fields := q.Result.Schema.Fields
// 	require.Equal(t, "pub,dom,timestamp,timestamp_year,m1,m1_p,timestamp__previous,timestamp_year__previous", columnNames(fields))
// 	i := 0

// 	for _, sf := range q.Result.Schema.Fields {
// 		fmt.Printf("%v ", sf.Name)
// 	}
// 	fmt.Printf("\n")

// 	for i, row := range q.Result.Data {
// 		for _, sf := range q.Result.Schema.Fields {
// 			fmt.Printf("%v ", row.Fields[sf.Name].AsInterface())
// 		}
// 		fmt.Printf(" %d \n", i)

// 	}
// 	rows := q.Result.Data
// 	require.Equal(t, 4, len(rows))

// 	i = 0
// 	require.Equal(t, "Google,google.com,2022-01-01T00:00:00Z,2022-01-01T00:00:00Z,null,null,2022-01-02T00:00:00Z,2022-01-01T00:00:00Z", fieldsToString2digits(rows[i], "pub", "dom", "timestamp", "timestamp_year", "m1", "m1_p", "timestamp__previous", "timestamp_year__previous"))
// 	i++
// 	require.Equal(t, "Google,news.google.com,2022-01-01T00:00:00Z,2022-01-01T00:00:00Z,3.55,3.74,2022-01-02T00:00:00Z,2022-01-01T00:00:00Z", fieldsToString2digits(rows[i], "pub", "dom", "timestamp", "timestamp_year", "m1", "m1_p", "timestamp__previous", "timestamp_year__previous"))
// 	i++
// 	require.Equal(t, "Yahoo,news.yahoo.com,2022-01-01T00:00:00Z,2022-01-01T00:00:00Z,null,null,2022-01-02T00:00:00Z,2022-01-01T00:00:00Z", fieldsToString2digits(rows[i], "pub", "dom", "timestamp", "timestamp_year", "m1", "m1_p", "timestamp__previous", "timestamp_year__previous"))
// 	i++
// 	require.Equal(t, "Yahoo,sports.yahoo.com,2022-01-01T00:00:00Z,2022-01-01T00:00:00Z,null,null,2022-01-02T00:00:00Z,2022-01-01T00:00:00Z", fieldsToString2digits(rows[i], "pub", "dom", "timestamp", "timestamp_year", "m1", "m1_p", "timestamp__previous", "timestamp_year__previous"))
// }

// Uncommenting: We don't support comparison queries with time as a dimension
// func TestMetricsViewsAggregation_comparison_measure_filter_with_having(t *testing.T) {
// 	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")
//
// 	limit := int64(10)
// 	q := &queries.MetricsViewAggregation{
// 		MetricsViewName: "ad_bids_metrics",
// 		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
// 			{
// 				Name: "pub",
// 			},
// 			{
// 				Name: "dom",
// 			},
//
// 			{
// 				Name:      "timestamp",
// 				TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_DAY,
// 			},
// 			{
// 				Name:      "timestamp",
// 				TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
// 				Alias:     "timestamp_year",
// 			},
// 		},
// 		Measures: []*runtimev1.MetricsViewAggregationMeasure{
// 			{
// 				Name: "m1",
// 			},
// 			{
// 				Name: "m1_p",
// 				Compute: &runtimev1.MetricsViewAggregationMeasure_ComparisonValue{
// 					ComparisonValue: &runtimev1.MetricsViewAggregationMeasureComputeComparisonValue{
// 						Measure: "m1",
// 					},
// 				},
// 				Filter: &runtimev1.Expression{
// 					Expression: &runtimev1.Expression_Cond{
// 						Cond: &runtimev1.Condition{
// 							Op: runtimev1.Operation_OPERATION_EQ,
// 							Exprs: []*runtimev1.Expression{
// 								{
// 									Expression: &runtimev1.Expression_Ident{
// 										Ident: "dom",
// 									},
// 								},
// 								{
// 									Expression: &runtimev1.Expression_Val{
// 										Val: structpb.NewStringValue("news.google.com"),
// 									},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 		},
// 		Where: expressionpb.OrAll(
// 			expressionpb.Eq("pub", "Yahoo"),
// 			expressionpb.Eq("pub", "Google"),
// 		),
// 		Having: expressionpb.Gt("m1", 0.0),
// 		Sort: []*runtimev1.MetricsViewAggregationSort{
// 			{
// 				Name: "timestamp",
// 			},
// 			{
// 				Name: "pub",
// 			},
// 			{
// 				Name: "dom",
// 			},
// 			{
// 				Name: "timestamp_year",
// 			},
// 			{
// 				Name: "m1_p",
// 			},
// 		},
//
// 		TimeRange: &runtimev1.TimeRange{
// 			Start: timestamppb.New(time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)),
// 			End:   timestamppb.New(time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC)),
// 		},
// 		ComparisonTimeRange: &runtimev1.TimeRange{
// 			Start: timestamppb.New(time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC)),
// 			End:   timestamppb.New(time.Date(2022, 1, 3, 0, 0, 0, 0, time.UTC)),
// 		},
// 		Limit: &limit,
// 	}
// 	err := q.Resolve(context.Background(), rt, instanceID, 0)
// 	require.NoError(t, err)
// 	require.NotEmpty(t, q.Result)
// 	fields := q.Result.Schema.Fields
// 	require.Equal(t, "pub,dom,timestamp,timestamp_year,m1,m1_p,timestamp__previous,timestamp_year__previous", columnNames(fields))
// 	i := 0
//
// 	for _, sf := range q.Result.Schema.Fields {
// 		fmt.Printf("%v ", sf.Name)
// 	}
// 	fmt.Printf("\n")
//
// 	for i, row := range q.Result.Data {
// 		for _, sf := range q.Result.Schema.Fields {
// 			fmt.Printf("%v ", row.Fields[sf.Name].AsInterface())
// 		}
// 		fmt.Printf(" %d \n", i)
//
// 	}
// 	rows := q.Result.Data
// 	require.Equal(t, 1, len(rows))
//
// 	i = 0
// 	require.Equal(t, "Google,news.google.com,2022-01-01T00:00:00Z,2022-01-01T00:00:00Z,3.55,3.74,2022-01-02T00:00:00Z,2022-01-01T00:00:00Z", fieldsToString2digits(rows[i], "pub", "dom", "timestamp", "timestamp_year", "m1", "m1_p", "timestamp__previous", "timestamp_year__previous"))
// }

// Uncommenting: We don't support comparison queries with time as a dimension
// func TestMetricsViewsAggregation_comparison(t *testing.T) {
// 	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")

// 	limit := int64(10)
// 	q := &queries.MetricsViewAggregation{
// 		MetricsViewName: "ad_bids_metrics",
// 		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
// 			{
// 				Name: "pub",
// 			},
// 			{
// 				Name: "dom",
// 			},

// 			{
// 				Name:      "timestamp",
// 				TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_DAY,
// 			},
// 			{
// 				Name:      "timestamp",
// 				TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
// 				Alias:     "timestamp_year",
// 			},
// 		},
// 		Measures: []*runtimev1.MetricsViewAggregationMeasure{
// 			{
// 				Name: "measure_0",
// 			},
// 			{
// 				Name: "measure_1",
// 			},
// 			{
// 				Name: "m1",
// 			},
// 			{
// 				Name: "measure_0__p",
// 				Compute: &runtimev1.MetricsViewAggregationMeasure_ComparisonValue{
// 					ComparisonValue: &runtimev1.MetricsViewAggregationMeasureComputeComparisonValue{
// 						Measure: "measure_0",
// 					},
// 				},
// 			},
// 		},
// 		Where: expressionpb.OrAll(
// 			expressionpb.Eq("pub", "Yahoo"),
// 			expressionpb.Eq("pub", "Google"),
// 		),
// 		Having: expressionpb.Gt("measure_1", 0.0),
// 		Sort: []*runtimev1.MetricsViewAggregationSort{
// 			{
// 				Name: "pub",
// 			},
// 			{
// 				Name: "dom",
// 			},
// 			{
// 				Name: "timestamp",
// 			},
// 			{
// 				Name: "timestamp_year",
// 			},
// 			{
// 				Name: "measure_1",
// 			},
// 		},

// 		TimeRange: &runtimev1.TimeRange{
// 			Start: timestamppb.New(time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)),
// 			End:   timestamppb.New(time.Date(2022, 1, 3, 0, 0, 0, 0, time.UTC)),
// 		},
// 		ComparisonTimeRange: &runtimev1.TimeRange{
// 			Start: timestamppb.New(time.Date(2022, 1, 3, 0, 0, 0, 0, time.UTC)),
// 			End:   timestamppb.New(time.Date(2022, 1, 5, 0, 0, 0, 0, time.UTC)),
// 		},
// 		Limit: &limit,
// 	}
// 	err := q.Resolve(context.Background(), rt, instanceID, 0)
// 	require.NoError(t, err)
// 	require.NotEmpty(t, q.Result)
// 	fields := q.Result.Schema.Fields
// 	require.Equal(t, "pub,dom,timestamp,timestamp_year,measure_0,measure_1,m1,measure_0__p,timestamp__previous,timestamp_year__previous", columnNames(fields))
// 	i := 0

// 	for _, sf := range q.Result.Schema.Fields {
// 		fmt.Printf("%v ", sf.Name)
// 	}
// 	fmt.Printf("\n")

// 	for i, row := range q.Result.Data {
// 		for _, sf := range q.Result.Schema.Fields {
// 			fmt.Printf("%v ", row.Fields[sf.Name].AsInterface())
// 		}
// 		fmt.Printf(" %d \n", i)

// 	}
// 	rows := q.Result.Data
// 	require.Equal(t, 8, len(rows))

// 	i = 0
// 	require.Equal(t, "Google,google.com,2022-01-01T00:00:00Z,2022-01-01T00:00:00Z,44.00,50.00,1.53,1.53,2022-01-03T00:00:00Z,2022-01-01T00:00:00Z", fieldsToString2digits(rows[i], "pub", "dom", "timestamp", "timestamp_year", "measure_0", "measure_0__p", "measure_1", "m1", "timestamp__previous", "timestamp_year__previous"))
// 	i++
// 	require.Equal(t, "Google,google.com,2022-01-02T00:00:00Z,2022-01-01T00:00:00Z,62.00,51.00,1.45,1.45,2022-01-04T00:00:00Z,2022-01-01T00:00:00Z", fieldsToString2digits(rows[i], "pub", "dom", "timestamp", "timestamp_year", "measure_0", "measure_0__p", "measure_1", "m1", "timestamp__previous", "timestamp_year__previous"))
// 	i++
// 	require.Equal(t, "Google,news.google.com,2022-01-01T00:00:00Z,2022-01-01T00:00:00Z,187.00,183.00,3.55,3.55,2022-01-03T00:00:00Z,2022-01-01T00:00:00Z", fieldsToString2digits(rows[i], "pub", "dom", "timestamp", "timestamp_year", "measure_0", "measure_0__p", "measure_1", "m1", "timestamp__previous", "timestamp_year__previous"))
// }

func TestMetricsViewsAggregation_comparison_pivot(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")

	limit := int64(10)
	q := &queries.MetricsViewAggregation{
		MetricsViewName: "ad_bids_metrics",
		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
			{
				Name: "pub",
			},
			{
				Name: "dom",
			},
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "measure_0",
			},
			{
				Name: "measure_0__p",
				Compute: &runtimev1.MetricsViewAggregationMeasure_ComparisonValue{
					ComparisonValue: &runtimev1.MetricsViewAggregationMeasureComputeComparisonValue{
						Measure: "measure_0",
					},
				},
			},
		},
		Where: expressionpb.OrAll(
			expressionpb.Eq("pub", "Yahoo"),
			expressionpb.Eq("pub", "Google"),
		),
		Having: expressionpb.Gt("measure_0", 0.0),
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "pub",
			},
		},
		TimeRange: &runtimev1.TimeRange{
			Start: timestamppb.New(time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)),
			End:   timestamppb.New(time.Date(2022, 1, 3, 0, 0, 0, 0, time.UTC)),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: timestamppb.New(time.Date(2022, 1, 3, 0, 0, 0, 0, time.UTC)),
			End:   timestamppb.New(time.Date(2022, 1, 5, 0, 0, 0, 0, time.UTC)),
		},
		PivotOn:        []string{"dom"},
		Limit:          &limit,
		SecurityClaims: testClaims(),
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	for _, sf := range q.Result.Schema.Fields {
		fmt.Printf("%v ", sf.Name)
	}
	fmt.Printf("\n")
	fields := q.Result.Schema.Fields
	require.Equal(t, "pub,google.com_measure_0,google.com_measure_0__p,news.google.com_measure_0,news.google.com_measure_0__p,news.yahoo.com_measure_0,news.yahoo.com_measure_0__p,sports.yahoo.com_measure_0,sports.yahoo.com_measure_0__p", columnNames(fields))
}

// Uncommenting: We don't support comparison queries with time as a dimension
// func TestMetricsViewsAggregation_comparison_Druid_one_dim_base_order(t *testing.T) {
// 	if os.Getenv("METRICS_CREDS") == "" {
// 		t.Skip("skipping the test without the test instance")
// 	}
// 	rt, instanceID, err := testruntime.NewInstanceForDruidProject(t)
// 	require.NoError(t, err)

// 	limit := int64(10)
// 	q := &queries.MetricsViewAggregation{
// 		MetricsViewName: "ad_bids_metrics",
// 		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
// 			{
// 				Name: "pub",
// 			},
// 			{
// 				Name:      "__time",
// 				TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_DAY,
// 				Alias:     "timestamp_day",
// 			},
// 		},
// 		Measures: []*runtimev1.MetricsViewAggregationMeasure{
// 			{
// 				Name: "m1",
// 			},
// 			{
// 				Name: "m1__previous",
// 				Compute: &runtimev1.MetricsViewAggregationMeasure_ComparisonValue{
// 					ComparisonValue: &runtimev1.MetricsViewAggregationMeasureComputeComparisonValue{
// 						Measure: "m1",
// 					},
// 				},
// 			},
// 			{
// 				Name: "m1__delta_abs",
// 				Compute: &runtimev1.MetricsViewAggregationMeasure_ComparisonDelta{
// 					ComparisonDelta: &runtimev1.MetricsViewAggregationMeasureComputeComparisonDelta{
// 						Measure: "m1",
// 					},
// 				},
// 			},
// 			{
// 				Name: "m1__delta_rel",
// 				Compute: &runtimev1.MetricsViewAggregationMeasure_ComparisonRatio{
// 					ComparisonRatio: &runtimev1.MetricsViewAggregationMeasureComputeComparisonRatio{
// 						Measure: "m1",
// 					},
// 				},
// 			},
// 		},
// 		Where: expressionpb.OrAll(
// 			expressionpb.Eq("pub", "Yahoo"),
// 			expressionpb.Eq("pub", "Google"),
// 		),
// 		Sort: []*runtimev1.MetricsViewAggregationSort{
// 			{
// 				Name: "m1",
// 			},
// 			{
// 				Name: "pub",
// 			},
// 		},
// 		TimeRange: &runtimev1.TimeRange{
// 			Start: timestamppb.New(time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)),
// 			End:   timestamppb.New(time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC)),
// 		},
// 		ComparisonTimeRange: &runtimev1.TimeRange{
// 			Start: timestamppb.New(time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC)),
// 			End:   timestamppb.New(time.Date(2022, 1, 3, 0, 0, 0, 0, time.UTC)),
// 		},
// 		Limit: &limit,
// 	}
// 	err = q.Resolve(context.Background(), rt, instanceID, 0)
// 	require.NoError(t, err)
// 	require.NotEmpty(t, q.Result)
// 	fields := q.Result.Schema.Fields
// 	for _, sf := range q.Result.Schema.Fields {
// 		fmt.Printf("%v ", sf.Name)
// 	}
// 	fmt.Printf("\n")

// 	require.Equal(t, "pub,timestamp_day,m1,m1__previous,m1__delta_abs,m1__delta_rel", columnNames(fields))
// 	i := 0
// 	for i, row := range q.Result.Data {
// 		for _, sf := range q.Result.Schema.Fields {
// 			fmt.Printf("%v ", row.Fields[sf.Name].AsInterface())
// 		}
// 		fmt.Printf(" %d \n", i)

// 	}
// 	rows := q.Result.Data
// 	require.Equal(t, 2, len(rows))

// 	i = 0
// 	require.Equal(t, "Google,2022-01-01T00:00:00Z,3.17,3.18,-0.02,-0.00,2022-01-02T00:00:00Z", fieldsToString2digits(rows[i], "pub", "timestamp_day", "m1", "m1__previous", "m1__delta_abs", "m1__delta_rel", "timestamp_day__previous"))
// 	i++
// 	require.Equal(t, "Yahoo,2022-01-01T00:00:00Z,3.23,3.13,0.11,0.03,2022-01-02T00:00:00Z", fieldsToString2digits(rows[i], "pub", "timestamp_day", "m1", "m1__previous", "m1__delta_abs", "m1__delta_rel", "timestamp_day__previous"))
// }

// Uncommenting: We don't support comparison queries with time as a dimension
// func TestMetricsViewsAggregation_comparison_Druid_one_dim_comparison_order(t *testing.T) {
// 	if os.Getenv("METRICS_CREDS") == "" {
// 		t.Skip("skipping the test without the test instance")
// 	}
// 	rt, instanceID, err := testruntime.NewInstanceForDruidProject(t)
// 	require.NoError(t, err)

// 	limit := int64(10)
// 	q := &queries.MetricsViewAggregation{
// 		MetricsViewName: "ad_bids_metrics",
// 		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
// 			{
// 				Name: "pub",
// 			},
// 			{
// 				Name:      "__time",
// 				TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_DAY,
// 				Alias:     "timestamp_day",
// 			},
// 		},
// 		Measures: []*runtimev1.MetricsViewAggregationMeasure{
// 			{
// 				Name: "m1",
// 			},
// 			{
// 				Name: "m1__previous",
// 				Compute: &runtimev1.MetricsViewAggregationMeasure_ComparisonValue{
// 					ComparisonValue: &runtimev1.MetricsViewAggregationMeasureComputeComparisonValue{
// 						Measure: "m1",
// 					},
// 				},
// 			},
// 			{
// 				Name: "m1__delta_abs",
// 				Compute: &runtimev1.MetricsViewAggregationMeasure_ComparisonDelta{
// 					ComparisonDelta: &runtimev1.MetricsViewAggregationMeasureComputeComparisonDelta{
// 						Measure: "m1",
// 					},
// 				},
// 			},
// 			{
// 				Name: "m1__delta_rel",
// 				Compute: &runtimev1.MetricsViewAggregationMeasure_ComparisonRatio{
// 					ComparisonRatio: &runtimev1.MetricsViewAggregationMeasureComputeComparisonRatio{
// 						Measure: "m1",
// 					},
// 				},
// 			},
// 		},
// 		Where: expressionpb.OrAll(
// 			expressionpb.Eq("pub", "Yahoo"),
// 			expressionpb.Eq("pub", "Google"),
// 		),
// 		Sort: []*runtimev1.MetricsViewAggregationSort{
// 			{
// 				Name: "m1__previous",
// 			},
// 			{
// 				Name: "pub",
// 			},
// 		},
// 		TimeRange: &runtimev1.TimeRange{
// 			Start: timestamppb.New(time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)),
// 			End:   timestamppb.New(time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC)),
// 		},
// 		ComparisonTimeRange: &runtimev1.TimeRange{
// 			Start: timestamppb.New(time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC)),
// 			End:   timestamppb.New(time.Date(2022, 1, 3, 0, 0, 0, 0, time.UTC)),
// 		},
// 		Limit: &limit,
// 	}
// 	err = q.Resolve(context.Background(), rt, instanceID, 0)
// 	require.NoError(t, err)
// 	require.NotEmpty(t, q.Result)
// 	fields := q.Result.Schema.Fields
// 	for _, sf := range q.Result.Schema.Fields {
// 		fmt.Printf("%v ", sf.Name)
// 	}
// 	fmt.Printf("\n")

// 	require.Equal(t, "pub,timestamp_day,m1,m1__previous,m1__delta_abs,m1__delta_rel,timestamp_day__previous", columnNames(fields))
// 	i := 0

// 	for i, row := range q.Result.Data {
// 		for _, sf := range q.Result.Schema.Fields {
// 			fmt.Printf("%v ", row.Fields[sf.Name].AsInterface())
// 		}
// 		fmt.Printf(" %d \n", i)

// 	}
// 	rows := q.Result.Data
// 	require.Equal(t, 2, len(rows))

// 	i = 0
// 	require.Equal(t, "Yahoo,2022-01-01T00:00:00Z,3.23,3.13,0.11,0.03,2022-01-02T00:00:00Z", fieldsToString2digits(rows[i], "pub", "timestamp_day", "m1", "m1__previous", "m1__delta_abs", "m1__delta_rel", "timestamp_day__previous"))
// 	i++
// 	require.Equal(t, "Google,2022-01-01T00:00:00Z,3.17,3.18,-0.02,-0.00,2022-01-02T00:00:00Z", fieldsToString2digits(rows[i], "pub", "timestamp_day", "m1", "m1__previous", "m1__delta_abs", "m1__delta_rel", "timestamp_day__previous"))
// }

func TestMetricsViewsAggregation_Druid_comparison_empty_set_previous_sorted(t *testing.T) {
	rt, instanceID, err := testruntime.NewInstanceForDruidProject(t)
	require.NoError(t, err)

	limit := int64(10)
	q := &queries.MetricsViewAggregation{
		MetricsViewName: "ad_bids_metrics",
		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
			{
				Name: "pub",
			},
			{
				Name:      "__time",
				TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_DAY,
			},
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "m1",
			},
			{
				Name: "m1_p",
				Compute: &runtimev1.MetricsViewAggregationMeasure_ComparisonValue{
					ComparisonValue: &runtimev1.MetricsViewAggregationMeasureComputeComparisonValue{
						Measure: "m1",
					},
				},
			},
		},
		Where: expressionpb.OrAll(
			expressionpb.Eq("pub", "nothing"),
		),
		Having: expressionpb.Gt("m1", 0.0),
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "pub",
			},
			{
				Name: "__time",
			},
			{
				Name: "m1_p",
			},
		},
		TimeRange: &runtimev1.TimeRange{
			Start: timestamppb.New(time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)),
			End:   timestamppb.New(time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC)),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: timestamppb.New(time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC)),
			End:   timestamppb.New(time.Date(2022, 1, 3, 0, 0, 0, 0, time.UTC)),
		},
		Limit:          &limit,
		SecurityClaims: testClaims(),
	}
	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	fields := q.Result.Schema.Fields
	require.Equal(t, "pub,__time,m1,m1_p", columnNames(fields))

	for _, sf := range q.Result.Schema.Fields {
		fmt.Printf("%v ", sf.Name)
	}
	fmt.Printf("\n")

	for i, row := range q.Result.Data {
		for _, sf := range q.Result.Schema.Fields {
			fmt.Printf("%v ", row.Fields[sf.Name].AsInterface())
		}
		fmt.Printf(" %d \n", i)

	}
	rows := q.Result.Data
	require.Equal(t, 0, len(rows))
}

func TestMetricsViewsAggregation_Druid_comparison_empty_set(t *testing.T) {
	rt, instanceID, err := testruntime.NewInstanceForDruidProject(t)
	require.NoError(t, err)

	limit := int64(10)
	q := &queries.MetricsViewAggregation{
		MetricsViewName: "ad_bids_metrics",
		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
			{
				Name: "pub",
			},
			{
				Name:      "__time",
				TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_DAY,
			},
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "m1",
			},
			{
				Name: "m1_p",
				Compute: &runtimev1.MetricsViewAggregationMeasure_ComparisonValue{
					ComparisonValue: &runtimev1.MetricsViewAggregationMeasureComputeComparisonValue{
						Measure: "m1",
					},
				},
			},
		},
		Where: expressionpb.OrAll(
			expressionpb.Eq("pub", "nothing"),
		),
		Having: expressionpb.Gt("m1", 0.0),
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "pub",
			},
			{
				Name: "__time",
			},
			{
				Name: "m1",
			},
		},
		TimeRange: &runtimev1.TimeRange{
			Start: timestamppb.New(time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)),
			End:   timestamppb.New(time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC)),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: timestamppb.New(time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC)),
			End:   timestamppb.New(time.Date(2022, 1, 3, 0, 0, 0, 0, time.UTC)),
		},
		Limit:          &limit,
		SecurityClaims: testClaims(),
	}
	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	fields := q.Result.Schema.Fields
	require.Equal(t, "pub,__time,m1,m1_p", columnNames(fields))

	for _, sf := range q.Result.Schema.Fields {
		fmt.Printf("%v ", sf.Name)
	}
	fmt.Printf("\n")

	for i, row := range q.Result.Data {
		for _, sf := range q.Result.Schema.Fields {
			fmt.Printf("%v ", row.Fields[sf.Name].AsInterface())
		}
		fmt.Printf(" %d \n", i)

	}
	rows := q.Result.Data
	require.Equal(t, 0, len(rows))
}

// Uncommenting: We don't support comparison queries with time as a dimension
// func TestMetricsViewsAggregation_Druid_comparison_measure_filter_with_having(t *testing.T) {
// 	if os.Getenv("METRICS_CREDS") == "" {
// 		t.Skip("skipping the test without the test instance")
// 	}
// 	rt, instanceID, err := testruntime.NewInstanceForDruidProject(t)
// 	require.NoError(t, err)

// 	limit := int64(10)
// 	q := &queries.MetricsViewAggregation{
// 		MetricsViewName: "ad_bids_metrics",
// 		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
// 			{
// 				Name: "pub",
// 			},
// 			{
// 				Name: "dom",
// 			},

// 			{
// 				Name:      "__time",
// 				TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_DAY,
// 			},
// 			{
// 				Name:      "__time",
// 				TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
// 				Alias:     "timestamp_year",
// 			},
// 		},
// 		Measures: []*runtimev1.MetricsViewAggregationMeasure{
// 			{
// 				Name: "m1",
// 			},
// 			{
// 				Name: "m1_p",
// 				Compute: &runtimev1.MetricsViewAggregationMeasure_ComparisonValue{
// 					ComparisonValue: &runtimev1.MetricsViewAggregationMeasureComputeComparisonValue{
// 						Measure: "m1",
// 					},
// 				},
// 				Filter: &runtimev1.Expression{
// 					Expression: &runtimev1.Expression_Cond{
// 						Cond: &runtimev1.Condition{
// 							Op: runtimev1.Operation_OPERATION_EQ,
// 							Exprs: []*runtimev1.Expression{
// 								{
// 									Expression: &runtimev1.Expression_Ident{
// 										Ident: "dom",
// 									},
// 								},
// 								{
// 									Expression: &runtimev1.Expression_Val{
// 										Val: structpb.NewStringValue("news.google.com"),
// 									},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 		},
// 		Where: expressionpb.OrAll(
// 			expressionpb.Eq("pub", "Yahoo"),
// 			expressionpb.Eq("pub", "Google"),
// 		),
// 		Having: expressionpb.Gt("m1", 0.0),
// 		Sort: []*runtimev1.MetricsViewAggregationSort{
// 			{
// 				Name: "__time",
// 			},
// 			{
// 				Name: "pub",
// 			},
// 			{
// 				Name: "dom",
// 			},
// 			{
// 				Name: "timestamp_year",
// 			},
// 			{
// 				Name: "m1_p",
// 			},
// 		},

// 		TimeRange: &runtimev1.TimeRange{
// 			Start: timestamppb.New(time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)),
// 			End:   timestamppb.New(time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC)),
// 		},
// 		ComparisonTimeRange: &runtimev1.TimeRange{
// 			Start: timestamppb.New(time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC)),
// 			End:   timestamppb.New(time.Date(2022, 1, 3, 0, 0, 0, 0, time.UTC)),
// 		},
// 		Limit: &limit,
// 	}
// 	err = q.Resolve(context.Background(), rt, instanceID, 0)
// 	require.NoError(t, err)
// 	require.NotEmpty(t, q.Result)
// 	fields := q.Result.Schema.Fields
// 	require.Equal(t, "pub,dom,__time,timestamp_year,m1,m1_p,__time__previous,timestamp_year__previous", columnNames(fields))
// 	i := 0

// 	for _, sf := range q.Result.Schema.Fields {
// 		fmt.Printf("%v ", sf.Name)
// 	}
// 	fmt.Printf("\n")

// 	for i, row := range q.Result.Data {
// 		for _, sf := range q.Result.Schema.Fields {
// 			fmt.Printf("%v ", row.Fields[sf.Name].AsInterface())
// 		}
// 		fmt.Printf(" %d \n", i)

// 	}
// 	rows := q.Result.Data
// 	require.Equal(t, 1, len(rows))

// 	i = 0
// 	require.Equal(t, "Google,news.google.com,2022-01-01T00:00:00Z,2022-01-01T00:00:00Z,3.55,3.74,2022-01-02T00:00:00Z,2022-01-01T00:00:00Z", fieldsToString2digits(rows[i], "pub", "dom", "__time", "timestamp_year", "m1", "m1_p", "__time__previous", "timestamp_year__previous"))
// }

// Uncommenting: We don't support comparison queries with time as a dimension
// func TestMetricsViewsAggregation_Druid_comparison_measure_filter(t *testing.T) {
// 	if os.Getenv("METRICS_CREDS") == "" {
// 		t.Skip("skipping the test without the test instance")
// 	}
// 	rt, instanceID, err := testruntime.NewInstanceForDruidProject(t)
// 	require.NoError(t, err)

// 	limit := int64(10)
// 	q := &queries.MetricsViewAggregation{
// 		MetricsViewName: "ad_bids_metrics",
// 		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
// 			{
// 				Name: "pub",
// 			},
// 			{
// 				Name: "dom",
// 			},

// 			{
// 				Name:      "__time",
// 				TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_DAY,
// 			},
// 			{
// 				Name:      "__time",
// 				TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_YEAR,
// 				Alias:     "timestamp_year",
// 			},
// 		},
// 		Measures: []*runtimev1.MetricsViewAggregationMeasure{
// 			{
// 				Name: "m1_p",
// 				Compute: &runtimev1.MetricsViewAggregationMeasure_ComparisonValue{
// 					ComparisonValue: &runtimev1.MetricsViewAggregationMeasureComputeComparisonValue{
// 						Measure: "m1",
// 					},
// 				},
// 				Filter: &runtimev1.Expression{
// 					Expression: &runtimev1.Expression_Cond{
// 						Cond: &runtimev1.Condition{
// 							Op: runtimev1.Operation_OPERATION_EQ,
// 							Exprs: []*runtimev1.Expression{
// 								{
// 									Expression: &runtimev1.Expression_Ident{
// 										Ident: "dom",
// 									},
// 								},
// 								{
// 									Expression: &runtimev1.Expression_Val{
// 										Val: structpb.NewStringValue("news.google.com"),
// 									},
// 								},
// 							},
// 						},
// 					},
// 				},
// 			},
// 		},
// 		Where: expressionpb.OrAll(
// 			expressionpb.Eq("pub", "Yahoo"),
// 			expressionpb.Eq("pub", "Google"),
// 		),
// 		Sort: []*runtimev1.MetricsViewAggregationSort{
// 			{
// 				Name: "__time",
// 			},
// 			{
// 				Name: "pub",
// 			},
// 			{
// 				Name: "dom",
// 			},
// 			{
// 				Name: "timestamp_year",
// 			},
// 			{
// 				Name: "m1_p",
// 			},
// 		},

// 		TimeRange: &runtimev1.TimeRange{
// 			Start: timestamppb.New(time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)),
// 			End:   timestamppb.New(time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC)),
// 		},
// 		ComparisonTimeRange: &runtimev1.TimeRange{
// 			Start: timestamppb.New(time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC)),
// 			End:   timestamppb.New(time.Date(2022, 1, 3, 0, 0, 0, 0, time.UTC)),
// 		},
// 		Limit: &limit,
// 	}
// 	err = q.Resolve(context.Background(), rt, instanceID, 0)
// 	require.NoError(t, err)
// 	require.NotEmpty(t, q.Result)
// 	fields := q.Result.Schema.Fields
// 	require.Equal(t, "pub,dom,__time,timestamp_year,m1,m1_p,__time__previous,timestamp_year__previous", columnNames(fields))
// 	i := 0

// 	for _, sf := range q.Result.Schema.Fields {
// 		fmt.Printf("%v ", sf.Name)
// 	}
// 	fmt.Printf("\n")

// 	for i, row := range q.Result.Data {
// 		for _, sf := range q.Result.Schema.Fields {
// 			fmt.Printf("%v ", row.Fields[sf.Name].AsInterface())
// 		}
// 		fmt.Printf(" %d \n", i)

// 	}
// 	rows := q.Result.Data
// 	require.Equal(t, 4, len(rows))

// 	i = 0
// 	require.Equal(t, "Google,google.com,2022-01-01T00:00:00Z,2022-01-01T00:00:00Z,null,null,2022-01-02T00:00:00Z,2022-01-01T00:00:00Z", fieldsToString2digits(rows[i], "pub", "dom", "__time", "timestamp_year", "m1", "m1_p", "__time__previous", "timestamp_year__previous"))
// 	i++
// 	require.Equal(t, "Google,news.google.com,2022-01-01T00:00:00Z,2022-01-01T00:00:00Z,3.55,3.74,2022-01-02T00:00:00Z,2022-01-01T00:00:00Z", fieldsToString2digits(rows[i], "pub", "dom", "__time", "timestamp_year", "m1", "m1_p", "__time__previous", "timestamp_year__previous"))
// 	i++
// 	require.Equal(t, "Yahoo,news.yahoo.com,2022-01-01T00:00:00Z,2022-01-01T00:00:00Z,null,null,2022-01-02T00:00:00Z,2022-01-01T00:00:00Z", fieldsToString2digits(rows[i], "pub", "dom", "__time", "timestamp_year", "m1", "m1_p", "__time__previous", "timestamp_year__previous"))
// 	i++
// 	require.Equal(t, "Yahoo,sports.yahoo.com,2022-01-01T00:00:00Z,2022-01-01T00:00:00Z,null,null,2022-01-02T00:00:00Z,2022-01-01T00:00:00Z", fieldsToString2digits(rows[i], "pub", "dom", "__time", "timestamp_year", "m1", "m1_p", "__time__previous", "timestamp_year__previous"))
// }

func TestMetricsViewsAggregation_Druid_comparison_with_offset(t *testing.T) {
	rt, instanceID, err := testruntime.NewInstanceForDruidProject(t)
	require.NoError(t, err)

	limit := int64(2)
	q := &queries.MetricsViewAggregation{
		MetricsViewName: "ad_bids_metrics",
		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
			{
				Name: "dom",
			},
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "m1",
			},
			{
				Name: "m1__p",
				Compute: &runtimev1.MetricsViewAggregationMeasure_ComparisonValue{
					ComparisonValue: &runtimev1.MetricsViewAggregationMeasureComputeComparisonValue{
						Measure: "m1",
					},
				},
			},
		},
		Where: expressionpb.OrAll(
			expressionpb.Eq("pub", "Yahoo"),
			expressionpb.Eq("pub", "Google"),
		),
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "dom",
				Desc: true,
			},
		},

		TimeRange: &runtimev1.TimeRange{
			Start: timestamppb.New(time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)),
			End:   timestamppb.New(time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC)),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: timestamppb.New(time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC)),
			End:   timestamppb.New(time.Date(2022, 1, 3, 0, 0, 0, 0, time.UTC)),
		},
		Limit:          &limit,
		Offset:         1,
		SecurityClaims: testClaims(),
	}
	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	fields := q.Result.Schema.Fields
	require.Equal(t, "dom,m1,m1__p", columnNames(fields))
	i := 0

	for _, sf := range q.Result.Schema.Fields {
		fmt.Printf("%v ", sf.Name)
	}
	fmt.Printf("\n")

	for i, row := range q.Result.Data {
		for _, sf := range q.Result.Schema.Fields {
			fmt.Printf("%v ", row.Fields[sf.Name].AsInterface())
		}
		fmt.Printf(" %d \n", i)

	}
	rows := q.Result.Data
	require.Equal(t, 2, len(rows))

	i = 0
	require.Equal(t, "news.yahoo.com,1.49,1.51", fieldsToString2digits(rows[i], "dom", "m1", "m1__p"))
	i++
	require.Equal(t, "news.google.com,3.55,3.74", fieldsToString2digits(rows[i], "dom", "m1", "m1__p"))
}

func testMetricsViewsAggregation_comparison_with_offset_and_limit_and_delta(t *testing.T, rt *runtime.Runtime, instanceID string) {
	limit := int64(2)
	q := &queries.MetricsViewAggregation{
		MetricsViewName: "ad_bids_metrics",
		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
			{
				Name: "dom",
			},
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "m1",
			},
			{
				Name: "m1__d",
				Compute: &runtimev1.MetricsViewAggregationMeasure_ComparisonDelta{
					ComparisonDelta: &runtimev1.MetricsViewAggregationMeasureComputeComparisonDelta{
						Measure: "m1",
					},
				},
			},
		},
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "m1__d",
				Desc: true,
			},
		},

		TimeRange: &runtimev1.TimeRange{
			Start: timestamppb.New(time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)),
			End:   timestamppb.New(time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC)),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: timestamppb.New(time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC)),
			End:   timestamppb.New(time.Date(2022, 1, 3, 0, 0, 0, 0, time.UTC)),
		},
		Limit:          &limit,
		Offset:         3,
		SecurityClaims: testClaims(),
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	fields := q.Result.Schema.Fields
	require.Equal(t, "dom,m1,m1__d", columnNames(fields))
	i := 0

	for _, sf := range q.Result.Schema.Fields {
		fmt.Printf("%v ", sf.Name)
	}
	fmt.Printf("\n")

	for i, row := range q.Result.Data {
		for _, sf := range q.Result.Schema.Fields {
			fmt.Printf("%v ", row.Fields[sf.Name].AsInterface())
		}
		fmt.Printf(" %d \n", i)

	}
	rows := q.Result.Data
	require.Equal(t, 2, len(rows))

	i = 0
	require.Equal(t, "google.com,1.48,0.03", fieldsToString2digits(rows[i], "dom", "m1", "m1__d"))
	i++
	require.Equal(t, "sports.yahoo.com,3.74,-0.02", fieldsToString2digits(rows[i], "dom", "m1", "m1__d"))
}

func testMetricsViewsAggregation_comparison_with_offset(t *testing.T, rt *runtime.Runtime, instanceID string) {
	limit := int64(2)
	q := &queries.MetricsViewAggregation{
		MetricsViewName: "ad_bids_metrics",
		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
			{
				Name: "dom",
			},
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "m1",
			},
			{
				Name: "m1__p",
				Compute: &runtimev1.MetricsViewAggregationMeasure_ComparisonValue{
					ComparisonValue: &runtimev1.MetricsViewAggregationMeasureComputeComparisonValue{
						Measure: "m1",
					},
				},
			},
		},
		// Having: expressionpb.Gt("measure_1", 0.0),
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "dom",
				Desc: true,
			},
		},

		TimeRange: &runtimev1.TimeRange{
			Start: timestamppb.New(time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)),
			End:   timestamppb.New(time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC)),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: timestamppb.New(time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC)),
			End:   timestamppb.New(time.Date(2022, 1, 3, 0, 0, 0, 0, time.UTC)),
		},
		Limit:          &limit,
		Offset:         1,
		SecurityClaims: testClaims(),
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	fields := q.Result.Schema.Fields
	require.Equal(t, "dom,m1,m1__p", columnNames(fields))
	i := 0

	for _, sf := range q.Result.Schema.Fields {
		fmt.Printf("%v ", sf.Name)
	}
	fmt.Printf("\n")

	for i, row := range q.Result.Data {
		for _, sf := range q.Result.Schema.Fields {
			fmt.Printf("%v ", row.Fields[sf.Name].AsInterface())
		}
		fmt.Printf(" %d \n", i)

	}
	rows := q.Result.Data
	require.Equal(t, 2, len(rows))

	i = 0
	require.Equal(t, "news.yahoo.com,1.50,1.53", fieldsToString2digits(rows[i], "dom", "m1", "m1__p"))
	i++
	require.Equal(t, "news.google.com,3.59,3.69", fieldsToString2digits(rows[i], "dom", "m1", "m1__p"))
}

func testMetricsViewAggregation_percent_of_totals(t *testing.T, rt *runtime.Runtime, instanceID string) {
	q := &queries.MetricsViewAggregation{
		MetricsViewName: "ad_bids_metrics_view",
		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
			{
				Name: "domain",
			},
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "total_records",
			},
			{
				Name: "total_records__pt",
				Compute: &runtimev1.MetricsViewAggregationMeasure_PercentOfTotal{
					PercentOfTotal: &runtimev1.MetricsViewAggregationMeasureComputePercentOfTotal{
						Measure: "total_records",
					},
				},
			},
		},
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "domain",
				Desc: true,
			},
		},

		TimeRange: &runtimev1.TimeRange{
			Start: timestamppb.New(time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)),
			End:   timestamppb.New(time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC)),
		},
		Offset:         1,
		SecurityClaims: testClaims(),
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	fields := q.Result.Schema.Fields
	require.Equal(t, "domain,total_records,total_records__pt", columnNames(fields))
	i := 0

	rows := q.Result.Data
	require.Equal(t, 6, len(rows))

	i = 0
	require.Equal(t, "news.yahoo.com,77.00,0.07", fieldsToString2digits(rows[i], "domain", "total_records", "total_records__pt"))
	i++
	require.Equal(t, "news.google.com,256.00,0.23", fieldsToString2digits(rows[i], "domain", "total_records", "total_records__pt"))
	i++
	require.Equal(t, "msn.com,86.00,0.08", fieldsToString2digits(rows[i], "domain", "total_records", "total_records__pt"))
	i++
	require.Equal(t, "instagram.com,276.00,0.25", fieldsToString2digits(rows[i], "domain", "total_records", "total_records__pt"))
	i++
	require.Equal(t, "google.com,64.00,0.06", fieldsToString2digits(rows[i], "domain", "total_records", "total_records__pt"))
	i++
	require.Equal(t, "facebook.com,77.00,0.07", fieldsToString2digits(rows[i], "domain", "total_records", "total_records__pt"))
}

func testMetricsViewAggregation_percent_of_totals_with_limit(t *testing.T, rt *runtime.Runtime, instanceID string) {
	limit := int64(2)
	q := &queries.MetricsViewAggregation{
		MetricsViewName: "ad_bids_metrics_view",
		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
			{
				Name: "domain",
			},
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "total_records",
			},
			{
				Name: "total_records__pt",
				Compute: &runtimev1.MetricsViewAggregationMeasure_PercentOfTotal{
					PercentOfTotal: &runtimev1.MetricsViewAggregationMeasureComputePercentOfTotal{
						Measure: "total_records",
					},
				},
			},
		},
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "domain",
				Desc: true,
			},
		},

		TimeRange: &runtimev1.TimeRange{
			Start: timestamppb.New(time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)),
			End:   timestamppb.New(time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC)),
		},
		Limit:          &limit,
		Offset:         1,
		SecurityClaims: testClaims(),
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	fields := q.Result.Schema.Fields
	require.Equal(t, "domain,total_records,total_records__pt", columnNames(fields))
	i := 0

	rows := q.Result.Data
	require.Equal(t, 2, len(rows))

	i = 0
	require.Equal(t, "news.yahoo.com,77.00,0.07", fieldsToString2digits(rows[i], "domain", "total_records", "total_records__pt"))
	i++
	require.Equal(t, "news.google.com,256.00,0.23", fieldsToString2digits(rows[i], "domain", "total_records", "total_records__pt"))
}

func fieldsToString2digits(row *structpb.Struct, args ...string) string {
	s := make([]string, 0, len(args))
	for _, arg := range args {
		v := row.Fields[arg]
		switch vv := v.GetKind().(type) {
		case *structpb.Value_StringValue:
			s = append(s, vv.StringValue)
		case *structpb.Value_NumberValue:
			s = append(s, fmt.Sprintf("%.2f", vv.NumberValue))
		case *structpb.Value_NullValue:
			s = append(s, fmt.Sprintf("null"))
		}
	}
	return strings.Join(s, ",")
}

func columnNames(fields []*runtimev1.StructType_Field) string {
	var cols []string
	for _, f := range fields {
		cols = append(cols, f.Name)
	}
	return strings.Join(cols, ",")
}

func fieldsToString(row *structpb.Struct, args ...string) string {
	s := make([]string, 0, len(args))
	for _, arg := range args {
		v := row.Fields[arg]
		switch vv := v.GetKind().(type) {
		case *structpb.Value_StringValue:
			s = append(s, vv.StringValue)
		case *structpb.Value_NumberValue:
			s = append(s, fmt.Sprintf("%.0f", vv.NumberValue))
		case *structpb.Value_NullValue:
			s = append(s, fmt.Sprintf("null"))
		}
	}
	return strings.Join(s, ",")
}
