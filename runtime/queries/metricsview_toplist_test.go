package queries_test

import (
	"context"
	"fmt"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/queries"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/rilldata/rill/runtime/testruntime/testmode"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/clickhouse"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestMetricsViewsToplistAgainstClickHouse(t *testing.T) {
	testmode.Expensive(t)

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
	t.Run("TestMetricsViewsToplist_measure_filters", func(t *testing.T) { TestMetricsViewsToplist_measure_filters(t) })
}

func TestMetricsViewsToplist_measure_filters(t *testing.T) {
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
	q := &queries.MetricsViewToplist{
		MetricsViewName: "ad_bids_metrics",
		DimensionName:   "dom",
		MeasureNames:    []string{"measure_1"},
		TimeStart:       ctr.Result.Min,
		TimeEnd:         timestamppb.New(maxTime),
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
		Sort: []*runtimev1.MetricsViewSort{
			{
				Name:      "dom",
				Ascending: false,
			},
		},
		Limit:          &lmt,
		SecurityClaims: testClaims(),
	}

	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	require.Len(t, q.Result.Data, 3)
	require.Equal(t, "sports.yahoo.com", q.Result.Data[0].AsMap()["dom"])
	require.Equal(t, "news.google.com", q.Result.Data[1].AsMap()["dom"])
	require.Equal(t, "instagram.com", q.Result.Data[2].AsMap()["dom"])
}

func TestMetricsViewsToplist_measure_filters_same_name(t *testing.T) {
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
	q := &queries.MetricsViewToplist{
		MetricsViewName: "ad_bids_metrics",
		DimensionName:   "dom",
		MeasureNames:    []string{"bid_price"},
		TimeStart:       ctr.Result.Min,
		TimeEnd:         timestamppb.New(maxTime),
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
								Val: structpb.NewNumberValue(3.25),
							},
						},
					},
				},
			},
		},
		Sort: []*runtimev1.MetricsViewSort{
			{
				Name:      "bid_price",
				Ascending: false,
			},
		},
		Limit:          &lmt,
		SecurityClaims: testClaims(),
	}

	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	require.Len(t, q.Result.Data, 3)
	require.Equal(t, "sports.yahoo.com", q.Result.Data[0].AsMap()["dom"])
	require.Equal(t, "news.google.com", q.Result.Data[1].AsMap()["dom"])
	require.Equal(t, "instagram.com", q.Result.Data[2].AsMap()["dom"])
}
