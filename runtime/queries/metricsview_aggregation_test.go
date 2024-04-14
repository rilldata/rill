package queries_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/expressionpb"
	"github.com/rilldata/rill/runtime/queries"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/clickhouse"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	_ "github.com/rilldata/rill/runtime/drivers/duckdb"
)

func TestMetricViewAggregationAgainstClickHouse(t *testing.T) {
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

	t.Run("TestMetricsViewsAggregation", func(t *testing.T) { TestMetricsViewsAggregation(t) })
	t.Run("TestMetricsViewsAggregation_no_limit", func(t *testing.T) { TestMetricsViewsAggregation_no_limit(t) })
	t.Run("TestMetricsViewsAggregation_no_limit_pivot", func(t *testing.T) { TestMetricsViewsAggregation_no_limit_pivot(t) })
	t.Run("TestMetricsViewsAggregation_pivot", func(t *testing.T) { TestMetricsViewsAggregation_pivot(t) })
	t.Run("TestMetricsViewsAggregation_pivot_2_measures", func(t *testing.T) { TestMetricsViewsAggregation_pivot_2_measures(t) })
	t.Run("TestMetricsViewsAggregation_pivot_2_measures_and_filter", func(t *testing.T) { TestMetricsViewsAggregation_pivot_2_measures_and_filter(t) })
	t.Run("TestMetricsViewsAggregation_pivot_dim_and_measure", func(t *testing.T) { TestMetricsViewsAggregation_pivot_dim_and_measure(t) })
	t.Run("TestMetricsViewAggregation_measure_filters", func(t *testing.T) { TestMetricsViewAggregation_measure_filters(t) })
	t.Run("TestMetricsViewsAggregation_timezone", func(t *testing.T) { TestMetricsViewsAggregation_timezone(t) })
	t.Run("TestMetricsViewAggregationClickhouseEnum", func(t *testing.T) { testMetricsViewAggregationClickhouseEnum(t) })
}

func TestMetricsViewsAggregation(t *testing.T) {
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
			{
				Name: "timestamp",
			},
		},

		Limit: &limit,
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

func TestMetricsViewsAggregation_export_day(t *testing.T) {
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

		Limit:     &limit,
		Exporting: true,
	}

	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	rows := q.Result.Data

	i := 0
	require.Equal(t, "Facebook,2022-01-01T00:00:00Z", fieldsToString(rows[i], "pub", "timestamp"))
}

func TestMetricsViewsAggregation_export_hour(t *testing.T) {
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

		Limit:     &limit,
		Exporting: true,
	}

	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	rows := q.Result.Data

	i := 0
	require.Equal(t, "Facebook,2022-01-01T00:00:00Z", fieldsToString(rows[i], "pub", "timestamp"))
}

func TestMetricsViewsAggregation_no_limit(t *testing.T) {
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
			{
				Name: "timestamp",
			},
		},
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
		PivotOn: []string{"timestamp"},
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	require.Equal(t, 4, len(q.Result.Schema.Fields))
	require.Equal(t, 5, len(q.Result.Data))
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
		Limit: &limit,
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

func TestMetricsViewsAggregation_pivot_export_labels_2_time_columns(t *testing.T) {
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
		Limit:     &limit,
		Exporting: true,
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	rows := q.Result.Data

	require.Equal(t, 5, len(q.Result.Schema.Fields))
	require.Equal(t, "Publisher", q.Result.Schema.Fields[0].Name)
	require.Equal(t, "day", q.Result.Schema.Fields[1].Name)
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
		Limit:     &limit,
		Exporting: true,
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
		Limit:     &limit,
		Exporting: true,
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
		Limit:     &limit,
		Exporting: true,
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	rows := q.Result.Data

	require.Equal(t, 4, len(q.Result.Schema.Fields))
	require.Equal(t, "nolabel_pub", q.Result.Schema.Fields[0].Name)
	require.Equal(t, "2022-01-01 00:00:00_m1", q.Result.Schema.Fields[1].Name)
	require.Equal(t, "2022-02-01 00:00:00_m1", q.Result.Schema.Fields[2].Name)
	require.Equal(t, "2022-03-01 00:00:00_m1", q.Result.Schema.Fields[3].Name)

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
		Limit: &limit,
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
		Limit:     &limit,
		Exporting: true,
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
		Limit: &limit,
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
		Limit:     &limit,
		Exporting: true,
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	rows := q.Result.Data

	require.Equal(t, q.Result.Schema.Fields[0].Name, "Domain")
	require.Equal(t, q.Result.Schema.Fields[1].Name, "2022-01-01 00:00:00_google_Average bid price")
	require.Equal(t, q.Result.Schema.Fields[2].Name, "2022-02-01 00:00:00_google_Average bid price")
	require.Equal(t, q.Result.Schema.Fields[3].Name, "2022-03-01 00:00:00_google_Average bid price")

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
		Limit: &limit,
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

func TestMetricsViewAggregation_measure_filters(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")

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
	}

	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	require.Len(t, q.Result.Data, 3)
	require.NotEmpty(t, "sports.yahoo.com", q.Result.Data[0].AsMap()["dom"])
	require.NotEmpty(t, "news.google.com", q.Result.Data[1].AsMap()["dom"])
	require.NotEmpty(t, "instagram.com", q.Result.Data[2].AsMap()["dom"])
}

func TestMetricsViewsAggregation_timezone(t *testing.T) {
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

		Limit: &limit,
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

func TestMetricsViewsAggregation_filter(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")

	q := &queries.MetricsViewAggregation{
		MetricsViewName: "ad_bids_metrics",
		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
			{
				Name: "pub",
			},
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name:           "measure_1",
				BuiltinMeasure: runtimev1.BuiltinMeasure_BUILTIN_MEASURE_COUNT,
			},
		},
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "pub",
			},
		},
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)

	rows := q.Result.Data
	i := 0
	require.Equal(t, "Facebook,19341", fieldsToString(rows[i], "pub", "measure_1"))
	i++
	require.Equal(t, "Google,18763", fieldsToString(rows[i], "pub", "measure_1"))
	i++
	require.Equal(t, "Microsoft,10406", fieldsToString(rows[i], "pub", "measure_1"))

	q.Measures = []*runtimev1.MetricsViewAggregationMeasure{
		{
			Name:           "measure_1",
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
	require.Equal(t, "Facebook,8808", fieldsToString(rows[i], "pub", "measure_1"))
	i++
	require.Equal(t, "Google,null", fieldsToString(rows[i], "pub", "measure_1"))
	i++
	require.Equal(t, "Microsoft,null", fieldsToString(rows[i], "pub", "measure_1"))
}

func TestMetricsViewsAggregation_filter_2dims(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")

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
				Name:           "measure_1",
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
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)

	rows := q.Result.Data
	i := 0
	require.Equal(t, "Facebook,facebook.com,10533", fieldsToString(rows[i], "pub", "dom", "measure_1"))
	i++
	require.Equal(t, "Facebook,instagram.com,8808", fieldsToString(rows[i], "pub", "dom", "measure_1"))
	i++
	require.Equal(t, "Google,google.com,10119", fieldsToString(rows[i], "pub", "dom", "measure_1"))

	q.Measures = []*runtimev1.MetricsViewAggregationMeasure{
		{
			Name:           "measure_1",
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
	require.Equal(t, "Facebook,facebook.com,null", fieldsToString(rows[i], "pub", "dom", "measure_1"))
	i++
	require.Equal(t, "Facebook,instagram.com,8808", fieldsToString(rows[i], "pub", "dom", "measure_1"))
	i++
	require.Equal(t, "Google,google.com,null", fieldsToString(rows[i], "pub", "dom", "measure_1"))
}

func TestMetricsViewsAggregation_having_gt(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")

	q := &queries.MetricsViewAggregation{
		MetricsViewName: "ad_bids_metrics",
		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
			{
				Name: "pub",
			},
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name:           "measure_1",
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
								Ident: "measure_1",
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
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)

	rows := q.Result.Data
	require.Equal(t, 2, len(rows))
	i := 0
	require.Equal(t, "Facebook,19341", fieldsToString(rows[i], "pub", "measure_1"))
	i++
	require.Equal(t, "null,32897", fieldsToString(rows[i], "pub", "measure_1"))
}

func TestMetricsViewsAggregation_having(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")

	q := &queries.MetricsViewAggregation{
		MetricsViewName: "ad_bids_metrics",
		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
			{
				Name: "pub",
			},
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name:           "measure_1",
				BuiltinMeasure: runtimev1.BuiltinMeasure_BUILTIN_MEASURE_COUNT,
			},
		},
		Having: &runtimev1.Expression{
			Expression: &runtimev1.Expression_Cond{
				Cond: &runtimev1.Condition{
					Op: runtimev1.Operation_OPERATION_EQ,
					Exprs: []*runtimev1.Expression{
						{
							Expression: &runtimev1.Expression_Ident{
								Ident: "measure_1",
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
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)

	rows := q.Result.Data
	require.Equal(t, 1, len(rows))
	i := 0
	require.Equal(t, "Microsoft,10406", fieldsToString(rows[i], "pub", "measure_1"))
}

func TestMetricsViewsAggregation_where(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")

	q := &queries.MetricsViewAggregation{
		MetricsViewName: "ad_bids_metrics",
		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
			{
				Name: "pub",
			},
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name:           "measure_1",
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
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)

	rows := q.Result.Data
	i := 0
	require.Equal(t, "Facebook,19341", fieldsToString(rows[i], "pub", "measure_1"))
	i++
	require.Equal(t, "Microsoft,10406", fieldsToString(rows[i], "pub", "measure_1"))
}

func TestMetricsViewsAggregation_filter_having_measure(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")

	q := &queries.MetricsViewAggregation{
		MetricsViewName: "ad_bids_metrics",
		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
			{
				Name: "pub",
			},
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name:           "measure_1",
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
		},
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "pub",
			},
		},
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)

	rows := q.Result.Data
	i := 0
	require.Equal(t, "Facebook,8808", fieldsToString(rows[i], "pub", "measure_1"))
	i++
	require.Equal(t, "Google,null", fieldsToString(rows[i], "pub", "measure_1"))
	i++
	require.Equal(t, "Microsoft,null", fieldsToString(rows[i], "pub", "measure_1"))
	i++
	require.Equal(t, "Yahoo,null", fieldsToString(rows[i], "pub", "measure_1"))
	i++
	require.Equal(t, "null,4296", fieldsToString(rows[i], "pub", "measure_1"))

	// ================= check m1 > 5000

	q.Having = &runtimev1.Expression{
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
	require.Equal(t, "Facebook,8808", fieldsToString(rows[i], "pub", "measure_1"))

	// ================= check m1 < 5000

	q.Having = &runtimev1.Expression{
		Expression: &runtimev1.Expression_Cond{
			Cond: &runtimev1.Condition{
				Op: runtimev1.Operation_OPERATION_LT,
				Exprs: []*runtimev1.Expression{
					{
						Expression: &runtimev1.Expression_Ident{
							Ident: "measure_1",
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
	require.Equal(t, 0, len(rows))
}

func TestMetricsViewsAggregation_filter_with_where_and_having_measure(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")

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
				Name:           "measure_1",
				BuiltinMeasure: runtimev1.BuiltinMeasure_BUILTIN_MEASURE_COUNT,
			},
		},
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{
				Name: "pub",
			},
		},
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)

	rows := q.Result.Data
	i := 0
	require.Equal(t, "Google,8644", fieldsToString(rows[i], "pub", "measure_1"))
	i++
	require.Equal(t, "Microsoft,10406", fieldsToString(rows[i], "pub", "measure_1"))
	i++
	require.Equal(t, "null,9359", fieldsToString(rows[i], "pub", "measure_1"))

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
	require.Equal(t, "Google,8644", fieldsToString(rows[i], "pub", "measure_1"))
	i++
	require.Equal(t, "Microsoft,null", fieldsToString(rows[i], "pub", "measure_1"))
	i++
	require.Equal(t, "null,4239", fieldsToString(rows[i], "pub", "measure_1"))

	// ================= check having m1 > 5000

	q.Having = &runtimev1.Expression{
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
	require.Equal(t, "Google,8644", fieldsToString(rows[i], "pub", "measure_1"))
}

func TestMetricsViewsAggregation_2time_aggregations(t *testing.T) {
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

		Limit: &limit,
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
			"dashboards/bar.yaml": `
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
	}

	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result.Data)
	require.Equal(t, "a,null,1", fieldsToString(q.Result.Data[0], "a", "b", "count"))
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
