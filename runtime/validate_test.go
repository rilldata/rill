package runtime_test

import (
	"context"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func TestValidateMetricsView(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")
	res, err := rt.ValidateMetricsView(context.Background(), instanceID, &runtimev1.MetricsViewSpec{
		Connector:     "duckdb",
		Table:         "ad_bids",
		Title:         "Ad Bids",
		TimeDimension: "timestamp",
		Dimensions: []*runtimev1.MetricsViewSpec_DimensionV2{
			{Column: "publisher"},
		},
		Measures: []*runtimev1.MetricsViewSpec_MeasureV2{
			{Name: "records", Expression: "count(*)"},
			{Name: "invalid_nested_aggregation", Expression: "MAX(COUNT(DISTINCT publisher))"},
			{Name: "invalid_partition", Expression: "AVG(bid_price) OVER (PARTITION BY publisher)"},
		},
	})
	require.NoError(t, err)
	require.Empty(t, res.TimeDimensionErr)
	require.Empty(t, res.DimensionErrs)
	require.Empty(t, res.OtherErrs)

	require.Len(t, res.MeasureErrs, 2)
	require.Equal(t, 1, res.MeasureErrs[0].Idx)
	require.Equal(t, 2, res.MeasureErrs[1].Idx)
}
