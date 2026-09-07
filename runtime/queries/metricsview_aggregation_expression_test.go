package queries_test

import (
	"context"
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/expressionpb"
	"github.com/rilldata/rill/runtime/queries"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func expressionMeasure(name, expr string) *runtimev1.MetricsViewAggregationMeasure {
	return &runtimev1.MetricsViewAggregationMeasure{
		Name: name,
		Compute: &runtimev1.MetricsViewAggregationMeasure_Expression{
			Expression: &runtimev1.MetricsViewAggregationMeasureComputeExpression{
				Expression: expr,
			},
		},
	}
}

// testMetricsViewsAggregation_expression_measures runs against all OLAP drivers.
// It asserts identities that hold regardless of the underlying data:
// measure_1 and m1 are both avg(bid_price), so their difference is 0 and their ratio is 1.
func testMetricsViewsAggregation_expression_measures(t *testing.T, rt *runtime.Runtime, instanceID string) {
	limit := int64(10)
	q := &queries.MetricsViewAggregation{
		MetricsViewName: "ad_bids_metrics",
		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
			{Name: "pub"},
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{Name: "measure_0"},
			expressionMeasure("zero", "measure_1 - m1"),
			expressionMeasure("ratio", "m1 / measure_1"),
			expressionMeasure("double_count", "measure_0 * 2 - measure_0"),
			expressionMeasure("rounded", "round(coalesce(m1, 0) + 1.5, 2)"),
		},
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{Name: "pub"},
		},
		Limit:          &limit,
		SecurityClaims: testClaims(),
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result.Data)
	for _, row := range q.Result.Data {
		f := row.Fields
		require.InDelta(t, 0.0, f["zero"].GetNumberValue(), 1e-9)
		require.InDelta(t, 1.0, f["ratio"].GetNumberValue(), 1e-9)
		require.InDelta(t, f["measure_0"].GetNumberValue(), f["double_count"].GetNumberValue(), 1e-9)
	}
}

// The ad_bids_2rows project has exactly two rows:
// domain=msn.com: volume=4, impressions=2; domain=yahoo.com: volume=4, impressions=1.
// Its ad_bids_metrics measures are measure_0=count(*), measure_1=sum(volume), measure_2=sum(impressions), measure_3=sum(clicks).
func TestMetricsViewsAggregation_expression_values(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids_2rows")

	q := &queries.MetricsViewAggregation{
		MetricsViewName: "ad_bids_metrics",
		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
			{Name: "domain"},
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{Name: "measure_2"},
			expressionMeasure("profit", "measure_1 - measure_2"),
		},
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{Name: "profit", Desc: true},
		},
		SecurityClaims: testClaims(),
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.Len(t, q.Result.Data, 2)

	require.Equal(t, "yahoo.com", q.Result.Data[0].Fields["domain"].GetStringValue())
	require.Equal(t, 3.0, q.Result.Data[0].Fields["profit"].GetNumberValue())
	require.Equal(t, "msn.com", q.Result.Data[1].Fields["domain"].GetStringValue())
	require.Equal(t, 2.0, q.Result.Data[1].Fields["profit"].GetNumberValue())
}

func TestMetricsViewsAggregation_expression_totals(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids_2rows")

	q := &queries.MetricsViewAggregation{
		MetricsViewName: "ad_bids_metrics",
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			expressionMeasure("profit", "measure_1 - measure_2"),
			expressionMeasure("avg_volume", "measure_1 / measure_0"),
		},
		SecurityClaims: testClaims(),
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.Len(t, q.Result.Data, 1)
	require.Equal(t, 5.0, q.Result.Data[0].Fields["profit"].GetNumberValue())
	require.Equal(t, 4.0, q.Result.Data[0].Fields["avg_volume"].GetNumberValue())
}

func TestMetricsViewsAggregation_expression_having(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids_2rows")

	q := &queries.MetricsViewAggregation{
		MetricsViewName: "ad_bids_metrics",
		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
			{Name: "domain"},
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			expressionMeasure("profit", "measure_1 - measure_2"),
		},
		Having: expressionpb.Gt("profit", 2.5),
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{Name: "profit", Desc: true},
		},
		SecurityClaims: testClaims(),
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.Len(t, q.Result.Data, 1)
	require.Equal(t, "yahoo.com", q.Result.Data[0].Fields["domain"].GetStringValue())
}

func TestMetricsViewsAggregation_expression_pivot(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")

	limit := int64(10)
	q := &queries.MetricsViewAggregation{
		MetricsViewName: "ad_bids_metrics",
		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
			{Name: "pub"},
			{Name: "timestamp", TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_MONTH},
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			expressionMeasure("profit", "m1 * 2"),
		},
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{Name: "pub"},
		},
		PivotOn:        []string{"timestamp"},
		Limit:          &limit,
		SecurityClaims: testClaims(),
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result.Data)

	require.Equal(t, 4, len(q.Result.Schema.Fields))
	require.Equal(t, "pub", q.Result.Schema.Fields[0].Name)
	require.Equal(t, "2022-01-01 00:00:00_profit", q.Result.Schema.Fields[1].Name)
	require.Equal(t, "2022-02-01 00:00:00_profit", q.Result.Schema.Fields[2].Name)
	require.Equal(t, "2022-03-01 00:00:00_profit", q.Result.Schema.Fields[3].Name)
}

func TestMetricsViewsAggregation_expression_alongside_comparison(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")

	limit := int64(10)
	q := &queries.MetricsViewAggregation{
		MetricsViewName: "ad_bids_metrics",
		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
			{Name: "pub"},
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{Name: "measure_1"},
			{
				Name: "measure_1__delta",
				Compute: &runtimev1.MetricsViewAggregationMeasure_ComparisonDelta{
					ComparisonDelta: &runtimev1.MetricsViewAggregationMeasureComputeComparisonDelta{
						Measure: "measure_1",
					},
				},
			},
			expressionMeasure("zero", "measure_1 - m1"),
		},
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{Name: "pub"},
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
	require.NotEmpty(t, q.Result.Data)
	for _, row := range q.Result.Data {
		require.InDelta(t, 0.0, row.Fields["zero"].GetNumberValue(), 1e-9)
	}
}

func TestMetricsViewsAggregation_expression_window_measure_ref(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")

	limit := int64(10)
	q := &queries.MetricsViewAggregation{
		MetricsViewName: "ad_bids_metrics_advanced",
		Dimensions: []*runtimev1.MetricsViewAggregationDimension{
			{Name: "timestamp", TimeGrain: runtimev1.TimeGrain_TIME_GRAIN_DAY},
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{Name: "bids_1day_rolling_avg"},
			expressionMeasure("rolling_avg_doubled", "bids_1day_rolling_avg * 2"),
		},
		Sort: []*runtimev1.MetricsViewAggregationSort{
			{Name: "timestamp"},
		},
		Limit:          &limit,
		SecurityClaims: testClaims(),
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.NotEmpty(t, q.Result.Data)
	for _, row := range q.Result.Data {
		require.InDelta(t, row.Fields["bids_1day_rolling_avg"].GetNumberValue()*2, row.Fields["rolling_avg_doubled"].GetNumberValue(), 1e-9)
	}
}

func TestMetricsViewsAggregation_expression_errors(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids_2rows")

	cases := []struct {
		name        string
		measure     *runtimev1.MetricsViewAggregationMeasure
		errContains string
	}{
		{"unknown ref", expressionMeasure("x", "unknown_measure * 2"), "not found"},
		{"aggregate", expressionMeasure("x", "sum(measure_1)"), "aggregate function"},
		{"injection", expressionMeasure("x", "measure_1); DROP TABLE ad_bids;--"), "invalid expression"},
		{"name collision", expressionMeasure("measure_1", "measure_0 + 1"), "collides with an existing measure name"},
		{"ephemeral ref", expressionMeasure("x", "y + measure_0"), "not found"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			q := &queries.MetricsViewAggregation{
				MetricsViewName: "ad_bids_metrics",
				Measures:        []*runtimev1.MetricsViewAggregationMeasure{c.measure},
				SecurityClaims:  testClaims(),
			}
			err := q.Resolve(context.Background(), rt, instanceID, 0)
			require.Error(t, err)
			require.Contains(t, err.Error(), c.errContains)
		})
	}
}

// TestMetricsViewsAggregation_expression_security checks that an expression may only reference
// measures the user can access under the metrics view's security policy.
// The policy on ad_bids_mini_metrics_with_policy excludes "total volume" unless the user's domain is msn.com.
func TestMetricsViewsAggregation_expression_security(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids")

	newQuery := func() *queries.MetricsViewAggregation {
		return &queries.MetricsViewAggregation{
			MetricsViewName: "ad_bids_mini_metrics_with_policy",
			Measures: []*runtimev1.MetricsViewAggregationMeasure{
				expressionMeasure("net", `"total impressions" - "total volume"`),
			},
		}
	}

	q := newQuery()
	q.SecurityClaims = &runtime.SecurityClaims{UserAttributes: map[string]any{"domain": "yahoo.com", "email": "user@yahoo.com"}}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.Error(t, err)
	require.ErrorContains(t, err, "total volume")

	q = newQuery()
	q.SecurityClaims = &runtime.SecurityClaims{UserAttributes: map[string]any{"domain": "msn.com", "email": "user@msn.com"}}
	err = q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.Len(t, q.Result.Data, 1)
}

// The time series API accepts expression measures via the `ephemeral_measures` field.
func TestMetricsViewTimeSeries_expression_measure(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids_2rows")

	q := &queries.MetricsViewTimeSeries{
		MetricsViewName: "ad_bids_metrics",
		MeasureNames:    []string{"measure_1"},
		EphemeralMeasures: []*runtimev1.MetricsViewAggregationMeasure{
			expressionMeasure("profit", "measure_1 - measure_2"),
		},
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_DAY,
		SecurityClaims:  testClaims(),
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.Len(t, q.Result.Data, 2)

	// 2022-01-01 (msn.com): volume=4, impressions=2 => profit=2
	// 2022-01-02 (yahoo.com): volume=4, impressions=1 => profit=3
	require.Equal(t, 2.0, q.Result.Data[0].Records.Fields["profit"].GetNumberValue())
	require.Equal(t, 3.0, q.Result.Data[1].Records.Fields["profit"].GetNumberValue())
}

func TestMetricsViewTimeSeries_expression_measure_only(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids_2rows")

	q := &queries.MetricsViewTimeSeries{
		MetricsViewName: "ad_bids_metrics",
		EphemeralMeasures: []*runtimev1.MetricsViewAggregationMeasure{
			expressionMeasure("profit", "measure_1 - measure_2"),
		},
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_DAY,
		SecurityClaims:  testClaims(),
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.NoError(t, err)
	require.Len(t, q.Result.Data, 2)
}

func TestMetricsViewTimeSeries_expression_rejects_other_computes(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceForProject(t, "ad_bids_2rows")

	q := &queries.MetricsViewTimeSeries{
		MetricsViewName: "ad_bids_metrics",
		EphemeralMeasures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "prev",
				Compute: &runtimev1.MetricsViewAggregationMeasure_ComparisonValue{
					ComparisonValue: &runtimev1.MetricsViewAggregationMeasureComputeComparisonValue{
						Measure: "measure_1",
					},
				},
			},
		},
		TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_DAY,
		SecurityClaims:  testClaims(),
	}
	err := q.Resolve(context.Background(), rt, instanceID, 0)
	require.Error(t, err)
	require.Contains(t, err.Error(), "only the `expression` compute is supported")
}
