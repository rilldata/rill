package queries_test

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/queries"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/clickhouse"
	"github.com/xuri/excelize/v2"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	// Register drivers
	_ "github.com/rilldata/rill/runtime/drivers/duckdb"
)

func TestMetricsViewsComparisonAgainstClickHouse(t *testing.T) {
	if testing.Short() {
		t.Skip("clickhouse: skipping test in short mode")
	}

	ctx := context.Background()
	clickHouseContainer, err := clickhouse.RunContainer(ctx,
		testcontainers.WithImage("clickhouse/clickhouse-server:latest"),
		clickhouse.WithUsername("clickhouse"),
		clickhouse.WithPassword("clickhouse"),
		clickhouse.WithConfigFile("../testruntime/testdata/clickhouse-config.xml"),
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		err := clickHouseContainer.Terminate(ctx)
		require.NoError(t, err)
	})

	host, err := clickHouseContainer.Host(ctx)
	require.NoError(t, err)
	port, err := clickHouseContainer.MappedPort(ctx, "9000/tcp")
	require.NoError(t, err)

	t.Setenv("RILL_RUNTIME_TEST_OLAP_DRIVER", "clickhouse")
	t.Setenv("RILL_RUNTIME_TEST_OLAP_DSN", fmt.Sprintf("clickhouse://clickhouse:clickhouse@%v:%v", host, port.Port()))
	t.Run("TestMetricsViewsComparison_dim_order_comparison_toplist_vs_general_toplist", func(t *testing.T) { TestMetricsViewsComparison_dim_order_comparison_toplist_vs_general_toplist(t) })
	t.Run("TestMetricsViewsComparison_dim_order", func(t *testing.T) { TestMetricsViewsComparison_dim_order(t) })
	t.Run("TestMetricsViewsComparison_measure_order", func(t *testing.T) { TestMetricsViewsComparison_measure_order(t) })
	t.Run("TestMetricsViewsComparison_measure_filters", func(t *testing.T) { TestMetricsViewsComparison_measure_filters(t) })
	t.Run("TestMetricsViewsComparison_measure_filters_with_compare_no_alias", func(t *testing.T) { TestMetricsViewsComparison_measure_filters_with_compare_no_alias(t) })
	t.Run("TestMetricsViewsComparison_measure_filters_with_compare_base_measure", func(t *testing.T) { TestMetricsViewsComparison_measure_filters_with_compare_base_measure(t) })
	t.Run("TestMetricsViewsComparison_measure_filters_with_compare_aliases", func(t *testing.T) { TestMetricsViewsComparison_measure_filters_with_compare_aliases(t) })
	t.Run("TestMetricsViewsComparison_export_xlsx", func(t *testing.T) { TestMetricsViewsComparison_export_xlsx(t) })
	t.Run("TestServer_MetricsViewTimeseries_export_csv", func(t *testing.T) { TestServer_MetricsViewTimeseries_export_csv(t) })
}

func TestMetricsViewsComparison_dim_order_comparison_toplist_vs_general_toplist(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")

	ctr := &queries.ColumnTimeRange{
		TableName:  "ad_bids",
		ColumnName: "timestamp",
	}
	err := ctr.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
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
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				Name:     "dom",
				SortType: runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_BASE_VALUE,
				Desc:     false,
			},
		},
		Limit: 10,
	}

	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	dims := make([]string, 0, 10)
	previous := ""
	for i, r := range q.Result.Rows {
		if i == 10 {
			break
		}

		require.Equal(t, -1, strings.Compare(previous, r.DimensionValue.GetStringValue()))
		previous = r.DimensionValue.GetStringValue()
		dims = append(dims, r.DimensionValue.GetStringValue())
	}

	q = &queries.MetricsViewComparison{
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
				Name:     "dom",
				SortType: runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_BASE_VALUE,
				Desc:     false,
			},
		},
		Limit: 10,
	}
	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)

	comparisonDims := make([]string, 0, 10)
	for i, r := range q.Result.Rows {
		if i == 10 {
			break
		}
		comparisonDims = append(comparisonDims, r.DimensionValue.GetStringValue())
	}
	require.Equal(t, dims, comparisonDims)
}

func TestMetricsViewsComparison_dim_order(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")

	ctr := &queries.ColumnTimeRange{
		TableName:  "ad_bids",
		ColumnName: "timestamp",
	}
	err := ctr.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
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
				Name:     "dom",
				SortType: runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_BASE_VALUE,
				Desc:     true,
			},
		},
		Limit: 250,
	}

	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	require.NotEmpty(t, "sports.yahoo.com", q.Result.Rows[0].DimensionValue)
	require.NotEmpty(t, "news.yahoo.com", q.Result.Rows[1].DimensionValue)
}

func TestMetricsViewsComparison_dim_order_no_sort_order(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")

	ctr := &queries.ColumnTimeRange{
		TableName:  "ad_bids",
		ColumnName: "timestamp",
	}
	err := ctr.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
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
				Name:     "dom",
				SortType: runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_UNSPECIFIED,
				Desc:     true,
			},
		},
		Limit: 250,
	}

	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err) // allow undefined sort type
}

func TestMetricsViewsComparison_measure_order(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")

	ctr := &queries.ColumnTimeRange{
		TableName:  "ad_bids",
		ColumnName: "timestamp",
	}
	err := ctr.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
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

	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	require.NotEmpty(t, "facebook.com", q.Result.Rows[0].DimensionValue)
	require.NotEmpty(t, "msn.com", q.Result.Rows[1].DimensionValue)
}

func TestMetricsViewsComparison_measure_filters(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")

	ctr := &queries.ColumnTimeRange{
		TableName:  "ad_bids",
		ColumnName: "timestamp",
	}
	err := ctr.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
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
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				Name:     "dom",
				SortType: runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_BASE_VALUE,
				Desc:     true,
			},
		},
		Limit: 250,
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
	}

	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	require.Len(t, q.Result.Rows, 3)
	require.Equal(t, "sports.yahoo.com", q.Result.Rows[0].DimensionValue.GetStringValue())
	require.Equal(t, "news.google.com", q.Result.Rows[1].DimensionValue.GetStringValue())
	require.Equal(t, "instagram.com", q.Result.Rows[2].DimensionValue.GetStringValue())
}

func TestMetricsViewsComparison_measure_filters_with_compare_no_alias(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")

	ctr := &queries.ColumnTimeRange{
		TableName:  "ad_bids",
		ColumnName: "timestamp",
	}
	err := ctr.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
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
				Name:     "dom",
				SortType: runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_BASE_VALUE,
				Desc:     true,
			},
		},
		Limit: 250,
		Having: &runtimev1.Expression{
			Expression: &runtimev1.Expression_Cond{
				Cond: &runtimev1.Condition{
					Op: runtimev1.Operation_OPERATION_GT,
					Exprs: []*runtimev1.Expression{
						{
							Expression: &runtimev1.Expression_Ident{
								Ident: "measure_1__delta_rel",
							},
						},
						{
							Expression: &runtimev1.Expression_Val{
								Val: structpb.NewNumberValue(1.0),
							},
						},
					},
				},
			},
		},
	}

	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.ErrorContains(t, err, "unknown column filter: measure_1__delta_rel")
}

func TestMetricsViewsComparison_measure_filters_with_compare_base_measure(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")

	ctr := &queries.ColumnTimeRange{
		TableName:  "ad_bids",
		ColumnName: "timestamp",
	}
	err := ctr.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
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
				Name:     "dom",
				SortType: runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_BASE_VALUE,
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
	}

	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	require.Len(t, q.Result.Rows, 3)
	require.Equal(t, "sports.yahoo.com", q.Result.Rows[0].DimensionValue.GetStringValue())
	require.Equal(t, "news.google.com", q.Result.Rows[1].DimensionValue.GetStringValue())
	require.Equal(t, "instagram.com", q.Result.Rows[2].DimensionValue.GetStringValue())
}

func TestMetricsViewsComparison_measure_filters_with_compare_aliases(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")

	ctr := &queries.ColumnTimeRange{
		TableName:  "ad_bids",
		ColumnName: "timestamp",
	}
	err := ctr.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
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
				Name:     "dom",
				SortType: runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_BASE_VALUE,
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
								Ident: "measure_1_delta",
							},
						},
						{
							Expression: &runtimev1.Expression_Val{
								Val: structpb.NewNumberValue(1),
							},
						},
					},
				},
			},
		},
		Aliases: []*runtimev1.MetricsViewComparisonMeasureAlias{
			{
				Name:  "measure_1",
				Type:  runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_REL_DELTA,
				Alias: "measure_1_delta",
			},
		},
		Limit: 250,
	}

	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	require.Len(t, q.Result.Rows, 3)
	require.Equal(t, "sports.yahoo.com", q.Result.Rows[0].DimensionValue.GetStringValue())
	require.Equal(t, "news.google.com", q.Result.Rows[1].DimensionValue.GetStringValue())
	require.Equal(t, "instagram.com", q.Result.Rows[2].DimensionValue.GetStringValue())
}

func TestMetricsViewsComparison_export_xlsx(t *testing.T) {
	t.Parallel()
	rt, instanceId := testruntime.NewInstanceForProject(t, "ad_bids_2rows")

	ctr := &queries.ColumnTimeRange{
		TableName:  "ad_bids",
		ColumnName: "timestamp",
	}
	err := ctr.Resolve(context.Background(), rt, instanceId, 0)
	require.NoError(t, err)
	diff := ctr.Result.Max.AsTime().Sub(ctr.Result.Min.AsTime())
	maxTime := ctr.Result.Min.AsTime().Add(diff / 2)

	q := &queries.MetricsViewComparison{
		MetricsViewName: "ad_bids_metrics",
		DimensionName:   "domain",
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "measure_1",
			},
		},
		TimeRange: &runtimev1.TimeRange{
			Start: ctr.Result.Min,
			End:   timestamppb.New(maxTime),
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				Name:     "domain",
				SortType: runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_BASE_VALUE,
				Desc:     false,
			},
		},
		Limit: 10,
	}

	var buf bytes.Buffer

	err = q.Export(context.Background(), rt, instanceId, &buf, &runtime.ExportOptions{
		Format: runtimev1.ExportFormat_EXPORT_FORMAT_XLSX,
	})
	require.NoError(t, err)

	file, err := excelize.OpenReader(&buf)
	rows, err := file.GetRows("Sheet1")
	require.NoError(t, err)

	require.Equal(t, 2, len(rows))
	require.Equal(t, "Domain Label", rows[0][0])
	require.Equal(t, "Total volume", rows[0][1])
}

func TestServer_MetricsViewTimeseries_export_csv(t *testing.T) {
	t.Parallel()
	rt, instanceId := testruntime.NewInstanceForProject(t, "ad_bids_2rows")

	ctr := &queries.ColumnTimeRange{
		TableName:  "ad_bids",
		ColumnName: "timestamp",
	}
	err := ctr.Resolve(context.Background(), rt, instanceId, 0)
	require.NoError(t, err)
	diff := ctr.Result.Max.AsTime().Sub(ctr.Result.Min.AsTime())
	maxTime := ctr.Result.Min.AsTime().Add(diff / 2)

	q := &queries.MetricsViewComparison{
		MetricsViewName: "ad_bids_metrics",
		DimensionName:   "domain",
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "measure_1",
			},
		},
		TimeRange: &runtimev1.TimeRange{
			Start: ctr.Result.Min,
			End:   timestamppb.New(maxTime),
		},
		// exports does not support sorting on dimension, so this is irrelevant for now
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				Name:     "domain",
				SortType: runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_BASE_VALUE,
				Desc:     false,
			},
		},
		Limit: 10,
	}

	var buf bytes.Buffer

	err = q.Export(context.Background(), rt, instanceId, &buf, &runtime.ExportOptions{
		Format: runtimev1.ExportFormat_EXPORT_FORMAT_CSV,
	})
	require.NoError(t, err)

	str := string(buf.Bytes())
	require.Equal(t, 2, strings.Count(str, "\n"))
	rowStrings := strings.Split(str, "\n")
	require.Equal(t, "Domain Label,Total volume", rowStrings[0])
}

func TestMetricsViewsComparison_measure_order_druid_exactify(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")

	ctr := &queries.ColumnTimeRange{
		TableName:  "ad_bids",
		ColumnName: "timestamp",
	}
	err := ctr.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
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
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				Name:     "measure_1",
				SortType: runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_COMPARISON_VALUE,
				Desc:     true,
			},
		},
		Limit: 250,
	}

	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)

	exactify := make([]*runtimev1.MetricsViewComparisonRow, 0, len(q.Result.Rows))
	for _, r := range q.Result.Rows {
		exactify = append(exactify, r)
	}

	q.NoDruidExactify = true
	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)

	require.Equal(t, exactify, q.Result.Rows)
}

func TestMetricsViewsComparison_measure_order_druid_exactify_with_comparison(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")

	ctr := &queries.ColumnTimeRange{
		TableName:  "ad_bids",
		ColumnName: "timestamp",
	}
	err := ctr.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
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

	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)

	exactify := make([]*runtimev1.MetricsViewComparisonRow, 0, len(q.Result.Rows))
	for _, r := range q.Result.Rows {
		exactify = append(exactify, r)
	}

	q.NoDruidExactify = true
	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)

	require.Equal(t, exactify, q.Result.Rows)
}

func TestMetricsViewsComparison_measure_order_druid_exactify_with_comparison_by_base(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")

	ctr := &queries.ColumnTimeRange{
		TableName:  "ad_bids",
		ColumnName: "timestamp",
	}
	err := ctr.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
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
				SortType: runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_BASE_VALUE,
				Desc:     true,
			},
		},
		Limit: 250,
	}

	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)

	exactify := make([]*runtimev1.MetricsViewComparisonRow, 0, len(q.Result.Rows))
	for _, r := range q.Result.Rows {
		exactify = append(exactify, r)
	}

	q.NoDruidExactify = true
	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)

	require.Equal(t, exactify, q.Result.Rows)
}

func TestMetricsViewsComparison_measure_order_druid_exactify_with_comparison_by_delta(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")

	ctr := &queries.ColumnTimeRange{
		TableName:  "ad_bids",
		ColumnName: "timestamp",
	}
	err := ctr.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
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
				SortType: runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_ABS_DELTA,
				Desc:     true,
			},
		},
		Limit: 250,
	}

	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)

	exactify := make([]*runtimev1.MetricsViewComparisonRow, 0, len(q.Result.Rows))
	for _, r := range q.Result.Rows {
		exactify = append(exactify, r)
	}

	q.NoDruidExactify = true
	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)

	require.Equal(t, exactify, q.Result.Rows)
}
