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

	ctrl, err := rt.Controller(context.Background(), instanceID)
	require.NoError(t, err)
	r, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: "ad_bids_metrics"}, false)
	require.NoError(t, err)
	mv := r.GetMetricsView()

	lmt := int64(250)
	q := &queries.MetricsViewToplist{
		MetricsViewName: "ad_bids_metrics",
		DimensionName:   "dom",
		MeasureNames:    []string{"measure_1"},
		MetricsView:     mv.Spec,
		TimeStart:       ctr.Result.Min,
		TimeEnd:         timestamppb.New(maxTime),
		Sort: []*runtimev1.MetricsViewSort{
			{
				Name:      "domain",
				Ascending: false,
			},
		},
		Limit: &lmt,
		MeasureFilter: &runtimev1.MeasureFilter{
			Expression: &runtimev1.MeasureFilterExpression{
				Entries: []*runtimev1.MeasureFilterNode{
					{
						Entry: &runtimev1.MeasureFilterNode_MeasureFilterMeasure{
							MeasureFilterMeasure: &runtimev1.MeasureFilterMeasure{
								Measure: &runtimev1.MetricsViewAggregationMeasure{Name: "measure_1"},
							},
						},
					},
					{
						Entry: &runtimev1.MeasureFilterNode_Value{
							Value: structpb.NewNumberValue(3.25),
						},
					},
				},
				OperationType: runtimev1.MeasureFilterExpression_OPERATION_TYPE_GREATER,
			},
		},
	}

	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result)
	require.Len(t, q.Result.Data, 3)
	require.Equal(t, "sports.yahoo.com", q.Result.Data[0].AsMap()["domain"])
	require.Equal(t, "news.google.com", q.Result.Data[1].AsMap()["domain"])
	require.Equal(t, "instagram.com", q.Result.Data[2].AsMap()["domain"])
}
