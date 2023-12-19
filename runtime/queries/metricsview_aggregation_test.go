package queries_test

import (
	"context"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/queries"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

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

	ctrl, err := rt.Controller(context.Background(), instanceID)
	require.NoError(t, err)
	r, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: "ad_bids_metrics"}, false)
	require.NoError(t, err)
	mv := r.GetMetricsView()

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
		MetricsView: mv.Spec,
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
