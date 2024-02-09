package queries_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/queries"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	_ "github.com/rilldata/rill/runtime/drivers/duckdb"
)

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
		for _, f := range row.Fields {
			fmt.Printf("%v ", f.AsInterface())
		}
		fmt.Printf(" %d \n", i)

	}
	rows := q.Result.Data

	i := 0
	require.Equal(t, "Facebook,2022-01-01", fieldsToString(rows[i], "pub", "timestamp"))
	i++
	require.Equal(t, "Facebook,2022-02-01", fieldsToString(rows[i], "pub", "timestamp"))
	i++
	require.Equal(t, "Facebook,2022-03-01", fieldsToString(rows[i], "pub", "timestamp"))
	i++
	require.Equal(t, "Google,2022-01-01", fieldsToString(rows[i], "pub", "timestamp"))
	i++
	require.Equal(t, "Google,2022-02-01", fieldsToString(rows[i], "pub", "timestamp"))
	i++
	require.Equal(t, "Google,2022-03-01", fieldsToString(rows[i], "pub", "timestamp"))
	i++
	require.Equal(t, "Microsoft,2022-01-01", fieldsToString(rows[i], "pub", "timestamp"))
	i++
	require.Equal(t, "Microsoft,2022-02-01", fieldsToString(rows[i], "pub", "timestamp"))
	i++
	require.Equal(t, "Microsoft,2022-03-01", fieldsToString(rows[i], "pub", "timestamp"))
	i++
	require.Equal(t, "Yahoo,2022-01-01", fieldsToString(rows[i], "pub", "timestamp"))
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
	for i, row := range q.Result.Data {
		for _, f := range row.Fields {
			fmt.Printf("%v ", f.AsInterface())
		}
		fmt.Printf(" %d \n", i)

	}
	rows := q.Result.Data

	require.Equal(t, 4, len(q.Result.Schema.Fields))
	require.Equal(t, "pub", q.Result.Schema.Fields[0].Name)
	require.Equal(t, "2022-01-01_measure_1", q.Result.Schema.Fields[1].Name)
	require.Equal(t, "2022-02-01_measure_1", q.Result.Schema.Fields[2].Name)
	require.Equal(t, "2022-03-01_measure_1", q.Result.Schema.Fields[3].Name)

	i := 0
	require.Equal(t, "Facebook", fieldsToString(rows[i], "pub"))
	i++
	require.Equal(t, "Google", fieldsToString(rows[i], "pub"))
	i++
	require.Equal(t, "Microsoft", fieldsToString(rows[i], "pub"))
	i++
	require.Equal(t, "Yahoo", fieldsToString(rows[i], "pub"))
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
	require.Equal(t, q.Result.Schema.Fields[1].Name, "2022-01-01_measure_1")
	require.Equal(t, q.Result.Schema.Fields[2].Name, "2022-01-01_measure_0")

	require.Equal(t, q.Result.Schema.Fields[3].Name, "2022-02-01_measure_1")
	require.Equal(t, q.Result.Schema.Fields[4].Name, "2022-02-01_measure_0")

	require.Equal(t, q.Result.Schema.Fields[5].Name, "2022-03-01_measure_1")
	require.Equal(t, q.Result.Schema.Fields[6].Name, "2022-03-01_measure_0")

	i := 0
	require.Equal(t, "Facebook", fieldsToString(rows[i], "pub"))
	i++
	require.Equal(t, "Google", fieldsToString(rows[i], "pub"))
	i++
	require.Equal(t, "Microsoft", fieldsToString(rows[i], "pub"))
	i++
	require.Equal(t, "Yahoo", fieldsToString(rows[i], "pub"))
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
	require.Equal(t, q.Result.Schema.Fields[1].Name, "2022-01-01_measure_1")
	require.Equal(t, q.Result.Schema.Fields[2].Name, "2022-01-01_measure_0")

	require.Equal(t, q.Result.Schema.Fields[3].Name, "2022-02-01_measure_1")
	require.Equal(t, q.Result.Schema.Fields[4].Name, "2022-02-01_measure_0")

	require.Equal(t, q.Result.Schema.Fields[5].Name, "2022-03-01_measure_1")
	require.Equal(t, q.Result.Schema.Fields[6].Name, "2022-03-01_measure_0")

	require.Equal(t, 1, len(rows))
	i := 0
	require.Equal(t, "Google", fieldsToString(rows[i], "pub"))
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
	require.Equal(t, q.Result.Schema.Fields[1].Name, "2022-01-01_Google_measure_1")
	require.Equal(t, q.Result.Schema.Fields[2].Name, "2022-02-01_Google_measure_1")
	require.Equal(t, q.Result.Schema.Fields[3].Name, "2022-03-01_Google_measure_1")

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
	require.Equal(t, "Google,0", fieldsToString(rows[i], "pub", "measure_1"))
	i++
	require.Equal(t, "Microsoft,0", fieldsToString(rows[i], "pub", "measure_1"))
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
