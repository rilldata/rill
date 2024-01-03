package server_test

import (
	"context"
	"fmt"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/expressionpb"
	"github.com/rilldata/rill/runtime/pkg/ratelimit"
	"github.com/rilldata/rill/runtime/server"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

func getMetricsTestServer(t *testing.T, projectName string) (*server.Server, string) {
	rt, instanceID := testruntime.NewInstanceForProject(t, projectName)

	server, err := server.NewServer(context.Background(), &server.Options{}, rt, nil, ratelimit.NewNoop(), activity.NewNoopClient())
	require.NoError(t, err)

	return server, instanceID
}

/*
|id |timestamp               |publisher|domain   |bid_price|volume|impressions|ad words|clicks|device|
|---|------------------------|---------|---------|---------|------|-----------|--------|------|------|
|0  |2022-01-01T14:49:50.459Z|         |msn.com  |2        |4     |2          |cars    |      |iphone|
|1  |2022-01-02T11:58:12.475Z|Yahoo    |yahoo.com|2        |4     |1          |cars    |1     |      |

dimensions:
  - label: Publisher
    property: publisher
    description: ""
  - label: Domain
    property: domain
    description: ""
  - label: Id
    property: id
  - label: Numeric Dim
    property: numeric_dim
  - label: Device
    property: device

measures:
  - label: "Number of bids"
    expression: count(*)
    description: ""
    format_preset: ""
  - label: "Total volume"
    expression: sum(volume)
    description: ""
    format_preset: ""
  - label: "Total impressions"
    expression: sum(impressions)
  - label: "Total clicks"
    expression: sum(clicks)
*/
func TestServer_MetricsViewComparison(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewComparison(testCtx(), &runtimev1.MetricsViewComparisonRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		Dimension: &runtimev1.MetricsViewAggregationDimension{
			Name: "ad words",
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "measure_2",
			},
		},
		TimeRange: &runtimev1.TimeRange{
			Start: parseTimeToProtoTimeStamps(t, "2022-01-02T00:00:00Z"),
			End:   parseTimeToProtoTimeStamps(t, "2022-01-02T23:59:00Z"),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: parseTimeToProtoTimeStamps(t, "2022-01-01T00:00:00Z"),
			End:   parseTimeToProtoTimeStamps(t, "2022-01-01T23:59:00Z"),
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				Name:     "measure_2",
				SortType: runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_BASE_VALUE,
				Desc:     false,
			},
		},
		Exact: true,
	})

	require.NoError(t, err)
	rows := tr.Rows
	require.Equal(t, 1, len(rows))
	require.Equal(t, "cars", rows[0].DimensionValue.GetStringValue())

	require.Equal(t, 1, len(rows[0].MeasureValues))

	require.Equal(t, 1.0, rows[0].MeasureValues[0].BaseValue.GetNumberValue())
	require.Equal(t, 2.0, rows[0].MeasureValues[0].ComparisonValue.GetNumberValue())
	require.Equal(t, -1.0, rows[0].MeasureValues[0].DeltaAbs.GetNumberValue())
	require.Equal(t, -0.5, rows[0].MeasureValues[0].DeltaRel.GetNumberValue())
}

/*
|domain                  |base |comparison|delta|rel    |base |comparison|delta|rel    |
|------------------------|-----|----------|-----|-------|-----|----------|-----|-------|
|msn.com                 |1    |1         |0    |0      |2    |1         | -1  |-0.5   |
|yahoo.com               |1    |1         |0    |0      |1    |2         |1    |1      |
*/
func TestServer_MetricsViewComparison_inline_measures(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewComparison(testCtx(), &runtimev1.MetricsViewComparisonRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		Dimension: &runtimev1.MetricsViewAggregationDimension{
			Name: "ad words",
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name:           "tmp_measure",
				BuiltinMeasure: runtimev1.BuiltinMeasure_BUILTIN_MEASURE_COUNT,
			},
			{
				Name: "measure_2",
			},
		},
		TimeRange: &runtimev1.TimeRange{
			Start: parseTimeToProtoTimeStamps(t, "2022-01-02T00:00:00Z"),
			End:   parseTimeToProtoTimeStamps(t, "2022-01-02T23:59:00Z"),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: parseTimeToProtoTimeStamps(t, "2022-01-01T00:00:00Z"),
			End:   parseTimeToProtoTimeStamps(t, "2022-01-01T23:59:00Z"),
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				Name:     "measure_2",
				SortType: runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_BASE_VALUE,
				Desc:     false,
			},
		},
		Exact: true,
	})

	require.NoError(t, err)
	rows := tr.Rows
	require.Equal(t, 1, len(rows))
	require.Equal(t, "cars", rows[0].DimensionValue.GetStringValue())

	require.Equal(t, 2, len(rows[0].MeasureValues))

	require.Equal(t, 1.0, rows[0].MeasureValues[0].BaseValue.GetNumberValue())
	require.Equal(t, 1.0, rows[0].MeasureValues[0].ComparisonValue.GetNumberValue())
	require.Equal(t, 0.0, rows[0].MeasureValues[0].DeltaAbs.GetNumberValue())
	require.Equal(t, 0.0, rows[0].MeasureValues[0].DeltaRel.GetNumberValue())

	require.Equal(t, 1.0, rows[0].MeasureValues[1].BaseValue.GetNumberValue())
	require.Equal(t, 2.0, rows[0].MeasureValues[1].ComparisonValue.GetNumberValue())
	require.Equal(t, -1.0, rows[0].MeasureValues[1].DeltaAbs.GetNumberValue())
	require.Equal(t, -0.5, rows[0].MeasureValues[1].DeltaRel.GetNumberValue())
}

func TestServer_MetricsViewComparison_nulls(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewComparison(testCtx(), &runtimev1.MetricsViewComparisonRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		Dimension: &runtimev1.MetricsViewAggregationDimension{
			Name: "domain",
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "measure_2",
			},
		},
		TimeRange: &runtimev1.TimeRange{
			Start: parseTimeToProtoTimeStamps(t, "2022-01-02T00:00:00Z"),
			End:   parseTimeToProtoTimeStamps(t, "2022-01-02T23:59:00Z"),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: parseTimeToProtoTimeStamps(t, "2022-01-01T00:00:00Z"),
			End:   parseTimeToProtoTimeStamps(t, "2022-01-01T23:59:00Z"),
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				Name:     "measure_2",
				SortType: runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_BASE_VALUE,
				Desc:     false,
			},
		},
		Exact: true,
	})

	require.NoError(t, err)
	rows := tr.Rows
	require.Equal(t, 2, len(rows))

	require.Equal(t, "yahoo.com", rows[0].DimensionValue.GetStringValue())
	require.Equal(t, 1.0, rows[0].MeasureValues[0].BaseValue.GetNumberValue())
	require.Equal(t, structpb.NullValue(0), rows[0].MeasureValues[0].ComparisonValue.GetNullValue())
	require.Equal(t, structpb.NullValue(0), rows[0].MeasureValues[0].DeltaAbs.GetNullValue())
	require.Equal(t, structpb.NullValue(0), rows[0].MeasureValues[0].DeltaRel.GetNullValue())

	require.Equal(t, "msn.com", rows[1].DimensionValue.GetStringValue())
	require.Equal(t, structpb.NullValue(0), rows[1].MeasureValues[0].BaseValue.GetNullValue())
	require.Equal(t, 2.0, rows[1].MeasureValues[0].ComparisonValue.GetNumberValue())
	require.Equal(t, structpb.NullValue(0), rows[1].MeasureValues[0].DeltaAbs.GetNullValue())
	require.Equal(t, structpb.NullValue(0), rows[1].MeasureValues[0].DeltaRel.GetNullValue())
}

/*
model:

|id |timestamp               |publisher|domain   |bid_price|volume|impressions|ad words|clicks|device|
|---|------------------------|---------|---------|---------|------|-----------|--------|------|------|
|0  |2022-01-01T14:49:50.459Z|         |msn.com  |2        |4     |2          |cars    |      |iphone|
|2  |2022-01-03T14:49:50.459Z|         |msn.com  |2.5      |4.5   |1          |cars    |      |iphone|
|1  |2022-01-02T11:58:12.475Z|Yahoo    |yahoo.com|2        |4     |1          |cars    |1     |      |
|3  |2022-01-04T11:58:12.475Z|Yahoo    |yahoo.com|2.5      |4.5   |2          |cars    |1.5   |      |

the result should be:

|domain                  |base |comparison|delta|rel    |
|------------------------|-----|----------|-----|-------|
|msn.com                 |2    |1         | -1  |-0.5   |
|yahoo.com               |1    |2         |1    |1      |
*/
func TestServer_MetricsViewComparison_sort_by_base(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids")

	tr, err := server.MetricsViewComparison(testCtx(), &runtimev1.MetricsViewComparisonRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_mini_metrics",
		Dimension: &runtimev1.MetricsViewAggregationDimension{
			Name: "domain",
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "measure_2",
			},
		},
		TimeRange: &runtimev1.TimeRange{
			Start: parseTimeToProtoTimeStamps(t, "2022-01-03T00:00:00Z"),
			End:   parseTimeToProtoTimeStamps(t, "2022-01-04T23:59:00Z"),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: parseTimeToProtoTimeStamps(t, "2022-01-01T00:00:00Z"),
			End:   parseTimeToProtoTimeStamps(t, "2022-01-02T23:59:00Z"),
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				Name:     "measure_2",
				SortType: runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_BASE_VALUE,
				Desc:     true,
			},
		},
		Exact: true,
	})

	require.NoError(t, err)
	rows := tr.Rows
	require.Equal(t, 2, len(rows))

	require.Equal(t, "yahoo.com", rows[0].DimensionValue.GetStringValue())
	require.Equal(t, 2.0, rows[0].MeasureValues[0].BaseValue.GetNumberValue())
	require.Equal(t, 1.0, rows[0].MeasureValues[0].ComparisonValue.GetNumberValue())
	require.Equal(t, 1.0, rows[0].MeasureValues[0].DeltaAbs.GetNumberValue())
	require.Equal(t, 1.0, rows[0].MeasureValues[0].DeltaRel.GetNumberValue())

	require.Equal(t, "msn.com", rows[1].DimensionValue.GetStringValue())
	require.Equal(t, 1.0, rows[1].MeasureValues[0].BaseValue.GetNumberValue())
	require.Equal(t, 2.0, rows[1].MeasureValues[0].ComparisonValue.GetNumberValue())
	require.Equal(t, -1.0, rows[1].MeasureValues[0].DeltaAbs.GetNumberValue())
	require.Equal(t, -0.5, rows[1].MeasureValues[0].DeltaRel.GetNumberValue())
}

/*
the result should be:

|domain                  |base |comparison|delta|rel    |
|------------------------|-----|----------|-----|-------|
|yahoo.com               |1    |2         |1    |1      |
|msn.com                 |2    |1         | -1  |-0.5   |
*/
func TestServer_MetricsViewComparison_sort_by_comparison(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids")

	tr, err := server.MetricsViewComparison(testCtx(), &runtimev1.MetricsViewComparisonRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_mini_metrics",
		Dimension: &runtimev1.MetricsViewAggregationDimension{
			Name: "domain",
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "measure_2",
			},
		},
		TimeRange: &runtimev1.TimeRange{
			Start: parseTimeToProtoTimeStamps(t, "2022-01-03T00:00:00Z"),
			End:   parseTimeToProtoTimeStamps(t, "2022-01-04T23:59:00Z"),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: parseTimeToProtoTimeStamps(t, "2022-01-01T00:00:00Z"),
			End:   parseTimeToProtoTimeStamps(t, "2022-01-02T23:59:00Z"),
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				Name:     "measure_2",
				SortType: runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_COMPARISON_VALUE,
				Desc:     true,
			},
		},
		Exact: true,
	})

	require.NoError(t, err)
	rows := tr.Rows
	require.Equal(t, 2, len(rows))

	require.Equal(t, "msn.com", rows[0].DimensionValue.GetStringValue())
	require.Equal(t, 1.0, rows[0].MeasureValues[0].BaseValue.GetNumberValue())
	require.Equal(t, 2.0, rows[0].MeasureValues[0].ComparisonValue.GetNumberValue())
	require.Equal(t, -1.0, rows[0].MeasureValues[0].DeltaAbs.GetNumberValue())
	require.Equal(t, -0.5, rows[0].MeasureValues[0].DeltaRel.GetNumberValue())

	require.Equal(t, "yahoo.com", rows[1].DimensionValue.GetStringValue())
	require.Equal(t, 2.0, rows[1].MeasureValues[0].BaseValue.GetNumberValue())
	require.Equal(t, 1.0, rows[1].MeasureValues[0].ComparisonValue.GetNumberValue())
	require.Equal(t, 1.0, rows[1].MeasureValues[0].DeltaAbs.GetNumberValue())
	require.Equal(t, 1.0, rows[1].MeasureValues[0].DeltaRel.GetNumberValue())
}

/*
the result should be:

|domain                  |base |comparison|delta|rel    |
|------------------------|-----|----------|-----|-------|
|yahoo.com               |100  |200       |100  |1      |
|msn.com                 |1    |10        |9    |9      |
*/

func TestServer_MetricsViewComparison_sort_by_abs_delta(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids")

	tr, err := server.MetricsViewComparison(testCtx(), &runtimev1.MetricsViewComparisonRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_mini_metrics",
		Dimension: &runtimev1.MetricsViewAggregationDimension{
			Name: "domain",
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "measure_1",
			},
		},
		TimeRange: &runtimev1.TimeRange{
			Start: parseTimeToProtoTimeStamps(t, "2022-01-03T00:00:00Z"),
			End:   parseTimeToProtoTimeStamps(t, "2022-01-04T23:59:00Z"),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: parseTimeToProtoTimeStamps(t, "2022-01-01T00:00:00Z"),
			End:   parseTimeToProtoTimeStamps(t, "2022-01-02T23:59:00Z"),
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				Name:     "measure_1",
				SortType: runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_ABS_DELTA,
				Desc:     true,
			},
		},
		Exact: true,
	})

	require.NoError(t, err)
	rows := tr.Rows
	require.Equal(t, 2, len(rows))

	require.Equal(t, "yahoo.com", rows[0].DimensionValue.GetStringValue())
	require.Equal(t, 200.0, rows[0].MeasureValues[0].BaseValue.GetNumberValue())
	require.Equal(t, 100.0, rows[0].MeasureValues[0].ComparisonValue.GetNumberValue())
	require.Equal(t, 100.0, rows[0].MeasureValues[0].DeltaAbs.GetNumberValue())
	require.Equal(t, 1.0, rows[0].MeasureValues[0].DeltaRel.GetNumberValue())

	require.Equal(t, "msn.com", rows[1].DimensionValue.GetStringValue())
	require.Equal(t, 10.0, rows[1].MeasureValues[0].BaseValue.GetNumberValue())
	require.Equal(t, 1.0, rows[1].MeasureValues[0].ComparisonValue.GetNumberValue())
	require.Equal(t, 9.0, rows[1].MeasureValues[0].DeltaAbs.GetNumberValue())
	require.Equal(t, 9.0, rows[1].MeasureValues[0].DeltaRel.GetNumberValue())
}

/*
the result should be:

|domain                  |base |comparison|delta|rel    |
|------------------------|-----|----------|-----|-------|
|msn.com                 |1    |10        |9    |9      |
|yahoo.com               |100  |200       |100  |1      |
*/
func TestServer_MetricsViewComparison_sort_by_rel_delta(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids")

	tr, err := server.MetricsViewComparison(testCtx(), &runtimev1.MetricsViewComparisonRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_mini_metrics",
		Dimension: &runtimev1.MetricsViewAggregationDimension{
			Name: "domain",
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "measure_1",
			},
		},
		TimeRange: &runtimev1.TimeRange{
			Start: parseTimeToProtoTimeStamps(t, "2022-01-03T00:00:00Z"),
			End:   parseTimeToProtoTimeStamps(t, "2022-01-04T23:59:00Z"),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: parseTimeToProtoTimeStamps(t, "2022-01-01T00:00:00Z"),
			End:   parseTimeToProtoTimeStamps(t, "2022-01-02T23:59:00Z"),
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				Name:     "measure_1",
				SortType: runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_REL_DELTA,
				Desc:     true,
			},
		},
		Exact: true,
	})

	require.NoError(t, err)
	rows := tr.Rows
	require.Equal(t, 2, len(rows))

	require.Equal(t, "msn.com", rows[0].DimensionValue.GetStringValue())
	require.Equal(t, 10.0, rows[0].MeasureValues[0].BaseValue.GetNumberValue())
	require.Equal(t, 1.0, rows[0].MeasureValues[0].ComparisonValue.GetNumberValue())
	require.Equal(t, 9.0, rows[0].MeasureValues[0].DeltaAbs.GetNumberValue())
	require.Equal(t, 9.0, rows[0].MeasureValues[0].DeltaRel.GetNumberValue())

	require.Equal(t, "yahoo.com", rows[1].DimensionValue.GetStringValue())
	require.Equal(t, 200.0, rows[1].MeasureValues[0].BaseValue.GetNumberValue())
	require.Equal(t, 100.0, rows[1].MeasureValues[0].ComparisonValue.GetNumberValue())
	require.Equal(t, 100.0, rows[1].MeasureValues[0].DeltaAbs.GetNumberValue())
	require.Equal(t, 1.0, rows[1].MeasureValues[0].DeltaRel.GetNumberValue())
}

func TestServer_MetricsViewComparison_sort_error(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids")

	_, err := server.MetricsViewComparison(testCtx(), &runtimev1.MetricsViewComparisonRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_mini_metrics",
		Dimension: &runtimev1.MetricsViewAggregationDimension{
			Name: "domain",
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "measure_2",
			},
		},
		TimeRange: &runtimev1.TimeRange{
			Start: parseTimeToProtoTimeStamps(t, "2022-01-03T00:00:00Z"),
			End:   parseTimeToProtoTimeStamps(t, "2022-01-04T23:59:00Z"),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: parseTimeToProtoTimeStamps(t, "2022-01-01T00:00:00Z"),
			End:   parseTimeToProtoTimeStamps(t, "2022-01-02T23:59:00Z"),
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				Name:     "measure_1",
				SortType: runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_ABS_DELTA,
				Desc:     true,
			},
		},
		Exact: true,
	})
	require.Error(t, err)
}

/*
the result should be:

|domain                  |base |comparison|delta|rel    |
|------------------------|-----|----------|-----|-------|
|yahoo.com               |1    |2         |1    |1      |
*/

func TestServer_MetricsViewComparison_sort_by_delta_limit_1(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids")

	tr, err := server.MetricsViewComparison(testCtx(), &runtimev1.MetricsViewComparisonRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_mini_metrics",
		Dimension: &runtimev1.MetricsViewAggregationDimension{
			Name: "domain",
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "measure_2",
			},
		},
		TimeRange: &runtimev1.TimeRange{
			Start: parseTimeToProtoTimeStamps(t, "2022-01-03T00:00:00Z"),
			End:   parseTimeToProtoTimeStamps(t, "2022-01-04T23:59:00Z"),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: parseTimeToProtoTimeStamps(t, "2022-01-01T00:00:00Z"),
			End:   parseTimeToProtoTimeStamps(t, "2022-01-02T23:59:00Z"),
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				Name:     "measure_2",
				SortType: runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_ABS_DELTA,
				Desc:     true,
			},
		},
		Limit: 1,
		Exact: true,
	})

	require.NoError(t, err)
	rows := tr.Rows
	require.Equal(t, 1, len(rows))

	require.Equal(t, "yahoo.com", rows[0].DimensionValue.GetStringValue())
	require.Equal(t, 2.0, rows[0].MeasureValues[0].BaseValue.GetNumberValue())
	require.Equal(t, 1.0, rows[0].MeasureValues[0].ComparisonValue.GetNumberValue())
	require.Equal(t, 1.0, rows[0].MeasureValues[0].DeltaAbs.GetNumberValue())
	require.Equal(t, 1.0, rows[0].MeasureValues[0].DeltaRel.GetNumberValue())
}

func TestServer_MetricsViewComparison_sort_by_base_limit_1(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids")

	tr, err := server.MetricsViewComparison(testCtx(), &runtimev1.MetricsViewComparisonRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_mini_metrics",
		Dimension: &runtimev1.MetricsViewAggregationDimension{
			Name: "domain",
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "measure_2",
			},
		},
		TimeRange: &runtimev1.TimeRange{
			Start: parseTimeToProtoTimeStamps(t, "2022-01-03T00:00:00Z"),
			End:   parseTimeToProtoTimeStamps(t, "2022-01-04T23:59:00Z"),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: parseTimeToProtoTimeStamps(t, "2022-01-01T00:00:00Z"),
			End:   parseTimeToProtoTimeStamps(t, "2022-01-02T23:59:00Z"),
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				Name:     "measure_2",
				SortType: runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_BASE_VALUE,
				Desc:     true,
			},
		},
		Limit: 1,
		Exact: true,
	})

	require.NoError(t, err)
	rows := tr.Rows
	require.Equal(t, 1, len(rows))

	require.Equal(t, "yahoo.com", rows[0].DimensionValue.GetStringValue())
	require.Equal(t, 2.0, rows[0].MeasureValues[0].BaseValue.GetNumberValue())
	require.Equal(t, 1.0, rows[0].MeasureValues[0].ComparisonValue.GetNumberValue())
	require.Equal(t, 1.0, rows[0].MeasureValues[0].DeltaAbs.GetNumberValue())
	require.Equal(t, 1.0, rows[0].MeasureValues[0].DeltaRel.GetNumberValue())
}

func TestServer_MetricsViewComparison_sort_by_base_filter(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids")

	tr, err := server.MetricsViewComparison(testCtx(), &runtimev1.MetricsViewComparisonRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_mini_metrics",
		Dimension: &runtimev1.MetricsViewAggregationDimension{
			Name: "domain",
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "measure_2",
			},
		},
		TimeRange: &runtimev1.TimeRange{
			Start: parseTimeToProtoTimeStamps(t, "2022-01-03T00:00:00Z"),
			End:   parseTimeToProtoTimeStamps(t, "2022-01-04T23:59:00Z"),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: parseTimeToProtoTimeStamps(t, "2022-01-01T00:00:00Z"),
			End:   parseTimeToProtoTimeStamps(t, "2022-01-02T23:59:00Z"),
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				Name:     "measure_2",
				SortType: runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_BASE_VALUE,
				Desc:     true,
			},
		},
		Where: expressionpb.NotIn(
			expressionpb.Identifier("domain"),
			[]*runtimev1.Expression{expressionpb.Value(structpb.NewStringValue("yahoo.com"))},
		),
		Exact: true,
	})

	require.NoError(t, err)
	rows := tr.Rows
	require.Equal(t, 1, len(rows))

	require.Equal(t, "msn.com", rows[0].DimensionValue.GetStringValue())
	require.Equal(t, 1.0, rows[0].MeasureValues[0].BaseValue.GetNumberValue())
	require.Equal(t, 2.0, rows[0].MeasureValues[0].ComparisonValue.GetNumberValue())
	require.Equal(t, -1.0, rows[0].MeasureValues[0].DeltaAbs.GetNumberValue())
	require.Equal(t, -0.5, rows[0].MeasureValues[0].DeltaRel.GetNumberValue())
}

/*
Model:

|id |timestamp               |publisher|domain   |bid_price|volume|impressions|ad words|clicks|device|
|---|------------------------|---------|---------|---------|------|-----------|--------|------|------|
|0  |2022-01-01T14:49:50.459Z|         |msn.com  |2        |1     |2          |cars    |      |iphone|
|2  |2022-01-03T14:49:50.459Z|         |msn.com  |2.5      |10    |1          |cars    |      |iphone|
|1  |2022-01-02T11:58:12.475Z|Yahoo    |yahoo.com|2        |100   |1          |cars    |1     |      |
|3  |2022-01-04T11:58:12.475Z|Yahoo    |yahoo.com|2.5      |200   |2          |cars    |1.5   |      |

the result should be:

|domain                  |base|comparison |delta|rel    |base |comparison|delta|rel    |
|------------------------|-----|----------|-----|-------|-----|----------|-----|-------|
|yahoo.com               |100  |200       |100  |1      | 1   |2         |1    |1      |
|msn.com                 |1    |10        |9    |9      | 2   |1         | -1  |-0.5   |
*/
func TestServer_MetricsViewComparison_2_measures(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids")

	tr, err := server.MetricsViewComparison(testCtx(), &runtimev1.MetricsViewComparisonRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_mini_metrics",
		Dimension: &runtimev1.MetricsViewAggregationDimension{
			Name: "domain",
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "measure_1",
			},
			{
				Name: "measure_2",
			},
		},
		TimeRange: &runtimev1.TimeRange{
			Start: parseTimeToProtoTimeStamps(t, "2022-01-03T00:00:00Z"),
			End:   parseTimeToProtoTimeStamps(t, "2022-01-04T23:59:00Z"),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: parseTimeToProtoTimeStamps(t, "2022-01-01T00:00:00Z"),
			End:   parseTimeToProtoTimeStamps(t, "2022-01-02T23:59:00Z"),
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				Name:     "measure_2",
				SortType: runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_ABS_DELTA,
				Desc:     true,
			},
		},
		Exact: true,
	})

	require.NoError(t, err)
	rows := tr.Rows
	require.Equal(t, 2, len(rows))

	require.Equal(t, "yahoo.com", rows[0].DimensionValue.GetStringValue())
	require.Equal(t, 200.0, rows[0].MeasureValues[0].BaseValue.GetNumberValue())
	require.Equal(t, 100.0, rows[0].MeasureValues[0].ComparisonValue.GetNumberValue())
	require.Equal(t, 100.0, rows[0].MeasureValues[0].DeltaAbs.GetNumberValue())
	require.Equal(t, 1.0, rows[0].MeasureValues[0].DeltaRel.GetNumberValue())
	require.Equal(t, 2.0, rows[0].MeasureValues[1].BaseValue.GetNumberValue())
	require.Equal(t, 1.0, rows[0].MeasureValues[1].ComparisonValue.GetNumberValue())
	require.Equal(t, 1.0, rows[0].MeasureValues[1].DeltaAbs.GetNumberValue())
	require.Equal(t, 1.0, rows[0].MeasureValues[1].DeltaRel.GetNumberValue())

	require.Equal(t, "msn.com", rows[1].DimensionValue.GetStringValue())
	require.Equal(t, 10.0, rows[1].MeasureValues[0].BaseValue.GetNumberValue())
	require.Equal(t, 1.0, rows[1].MeasureValues[0].ComparisonValue.GetNumberValue())
	require.Equal(t, 9.0, rows[1].MeasureValues[0].DeltaAbs.GetNumberValue())
	require.Equal(t, 9.0, rows[1].MeasureValues[0].DeltaRel.GetNumberValue())
	require.Equal(t, 1.0, rows[1].MeasureValues[1].BaseValue.GetNumberValue())
	require.Equal(t, 2.0, rows[1].MeasureValues[1].ComparisonValue.GetNumberValue())
	require.Equal(t, -1.0, rows[1].MeasureValues[1].DeltaAbs.GetNumberValue())
	require.Equal(t, -0.5, rows[1].MeasureValues[1].DeltaRel.GetNumberValue())
}

func TestServer_MetricsViewComparison_dimension_expression(t *testing.T) {
	t.Parallel()
	srv, instanceId := getMetricsTestServer(t, "ad_bids")

	tr, err := srv.MetricsViewComparison(testCtx(), &runtimev1.MetricsViewComparisonRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		Dimension: &runtimev1.MetricsViewAggregationDimension{
			Name: "tld",
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "measure_0",
			},
		},
		TimeRange: &runtimev1.TimeRange{
			Start: parseTimeToProtoTimeStamps(t, "2022-01-03T00:00:00Z"),
			End:   parseTimeToProtoTimeStamps(t, "2022-01-04T23:59:00Z"),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: parseTimeToProtoTimeStamps(t, "2022-01-01T00:00:00Z"),
			End:   parseTimeToProtoTimeStamps(t, "2022-01-02T23:59:00Z"),
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				Name:     "measure_0",
				SortType: runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_ABS_DELTA,
				Desc:     false,
			},
		},
		Where: expressionpb.NotLike(
			expressionpb.Identifier("dom"),
			expressionpb.Value(structpb.NewStringValue("%yahoo%")),
		),
		Exact: true,
	})
	require.NoError(t, err)
	require.Len(t, tr.Rows, 4)
	require.Equal(t, 1, len(tr.Rows[0].MeasureValues))
	require.Equal(t, "instagram.com", tr.Rows[0].DimensionValue.GetStringValue())
	require.Equal(t, "facebook.com", tr.Rows[1].DimensionValue.GetStringValue())
	require.Equal(t, "google.com", tr.Rows[2].DimensionValue.GetStringValue())
	require.Equal(t, "msn.com", tr.Rows[3].DimensionValue.GetStringValue())
}

func TestServer_MetricsViewComparison_dimension_expression_in_filter(t *testing.T) {
	t.Parallel()
	srv, instanceId := getMetricsTestServer(t, "ad_bids")

	tr, err := srv.MetricsViewComparison(testCtx(), &runtimev1.MetricsViewComparisonRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		Dimension: &runtimev1.MetricsViewAggregationDimension{
			Name: "tld",
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "measure_0",
			},
		},
		TimeRange: &runtimev1.TimeRange{
			Start: parseTimeToProtoTimeStamps(t, "2022-01-03T00:00:00Z"),
			End:   parseTimeToProtoTimeStamps(t, "2022-01-04T23:59:00Z"),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: parseTimeToProtoTimeStamps(t, "2022-01-01T00:00:00Z"),
			End:   parseTimeToProtoTimeStamps(t, "2022-01-02T23:59:00Z"),
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				Name:     "measure_0",
				SortType: runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_ABS_DELTA,
				Desc:     false,
			},
		},
		Where: expressionpb.NotIn(
			expressionpb.Identifier("tld"),
			[]*runtimev1.Expression{expressionpb.Value(structpb.NewStringValue("yahoo.com")), expressionpb.Value(structpb.NewStringValue("google.com"))},
		),
		Exact: true,
	})
	require.NoError(t, err)
	require.Len(t, tr.Rows, 3)
	require.Equal(t, 1, len(tr.Rows[0].MeasureValues))
	require.Equal(t, "instagram.com", tr.Rows[0].DimensionValue.GetStringValue())
	require.Equal(t, "facebook.com", tr.Rows[1].DimensionValue.GetStringValue())
	require.Equal(t, "msn.com", tr.Rows[2].DimensionValue.GetStringValue())
}

func TestServer_MetricsViewComparison_dimension_expression_like_filter(t *testing.T) {
	t.Parallel()
	srv, instanceId := getMetricsTestServer(t, "ad_bids")

	tr, err := srv.MetricsViewComparison(testCtx(), &runtimev1.MetricsViewComparisonRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		Dimension: &runtimev1.MetricsViewAggregationDimension{
			Name: "tld",
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "measure_0",
			},
		},
		TimeRange: &runtimev1.TimeRange{
			Start: parseTimeToProtoTimeStamps(t, "2022-01-03T00:00:00Z"),
			End:   parseTimeToProtoTimeStamps(t, "2022-01-04T23:59:00Z"),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: parseTimeToProtoTimeStamps(t, "2022-01-01T00:00:00Z"),
			End:   parseTimeToProtoTimeStamps(t, "2022-01-02T23:59:00Z"),
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				Name:     "measure_0",
				SortType: runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_ABS_DELTA,
				Desc:     false,
			},
		},
		Where: expressionpb.NotLike(
			expressionpb.Identifier("tld"),
			expressionpb.Value(structpb.NewStringValue("%yahoo%")),
		),
		Exact: true,
	})
	require.NoError(t, err)
	for _, row := range tr.Rows {
		fmt.Println(row.DimensionValue.GetStringValue())
	}
	require.Len(t, tr.Rows, 4)
	require.Equal(t, 1, len(tr.Rows[0].MeasureValues))
	require.Equal(t, "instagram.com", tr.Rows[0].DimensionValue.GetStringValue())
	require.Equal(t, "facebook.com", tr.Rows[1].DimensionValue.GetStringValue())
	require.Equal(t, "google.com", tr.Rows[2].DimensionValue.GetStringValue())
	require.Equal(t, "msn.com", tr.Rows[3].DimensionValue.GetStringValue())
}

func TestServer_MetricsViewComparison_unnested_dimension_expression_in_filter(t *testing.T) {
	t.Parallel()
	srv, instanceId := getMetricsTestServer(t, "ad_bids")

	tr, err := srv.MetricsViewComparison(testCtx(), &runtimev1.MetricsViewComparisonRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		Dimension: &runtimev1.MetricsViewAggregationDimension{
			Name: "domain_parts",
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "measure_0",
			},
		},
		TimeRange: &runtimev1.TimeRange{
			Start: parseTimeToProtoTimeStamps(t, "2022-01-03T00:00:00Z"),
			End:   parseTimeToProtoTimeStamps(t, "2022-01-04T23:59:00Z"),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: parseTimeToProtoTimeStamps(t, "2022-01-01T00:00:00Z"),
			End:   parseTimeToProtoTimeStamps(t, "2022-01-02T23:59:00Z"),
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				Name:     "measure_0",
				SortType: runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_ABS_DELTA,
				Desc:     false,
			},
		},
		Where: expressionpb.NotIn(
			expressionpb.Identifier("domain_parts"),
			[]*runtimev1.Expression{expressionpb.Value(structpb.NewStringValue("yahoo")), expressionpb.Value(structpb.NewStringValue("google"))},
		),
		Exact: true,
	})
	require.NoError(t, err)
	require.Len(t, tr.Rows, 4)
	require.Equal(t, 1, len(tr.Rows[0].MeasureValues))
	require.Equal(t, "instagram", tr.Rows[0].DimensionValue.GetStringValue())
	require.Equal(t, "com", tr.Rows[1].DimensionValue.GetStringValue())
	require.Equal(t, "facebook", tr.Rows[2].DimensionValue.GetStringValue())
	require.Equal(t, "msn", tr.Rows[3].DimensionValue.GetStringValue())
}

func TestServer_MetricsViewComparison_unnested_dimension_expression_like_filter(t *testing.T) {
	t.Parallel()
	srv, instanceId := getMetricsTestServer(t, "ad_bids")

	tr, err := srv.MetricsViewComparison(testCtx(), &runtimev1.MetricsViewComparisonRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		Dimension: &runtimev1.MetricsViewAggregationDimension{
			Name: "domain_parts",
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "measure_0",
			},
		},
		TimeRange: &runtimev1.TimeRange{
			Start: parseTimeToProtoTimeStamps(t, "2022-01-03T00:00:00Z"),
			End:   parseTimeToProtoTimeStamps(t, "2022-01-04T23:59:00Z"),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: parseTimeToProtoTimeStamps(t, "2022-01-01T00:00:00Z"),
			End:   parseTimeToProtoTimeStamps(t, "2022-01-02T23:59:00Z"),
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				Name:     "measure_0",
				SortType: runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_ABS_DELTA,
				Desc:     false,
			},
		},
		Where: expressionpb.NotLike(
			expressionpb.Identifier("domain_parts"),
			expressionpb.Value(structpb.NewStringValue("%yahoo%")),
		),
		Exact: true,
	})
	require.NoError(t, err)
	for _, row := range tr.Rows {
		fmt.Println(row.DimensionValue.GetStringValue())
	}
	require.Len(t, tr.Rows, 6)
	require.Equal(t, 1, len(tr.Rows[0].MeasureValues))
	require.Equal(t, "instagram", tr.Rows[0].DimensionValue.GetStringValue())
	require.Equal(t, "facebook", tr.Rows[1].DimensionValue.GetStringValue())
	require.Equal(t, "com", tr.Rows[2].DimensionValue.GetStringValue())
	require.Equal(t, "google", tr.Rows[3].DimensionValue.GetStringValue())
	require.Equal(t, "news", tr.Rows[4].DimensionValue.GetStringValue())
	require.Equal(t, "msn", tr.Rows[5].DimensionValue.GetStringValue())
}

/*
Source:

|id |timestamp               |publisher|domain   |bid_price|volume|impressions|ad words|clicks|device|
|---|------------------------|---------|---------|---------|------|-----------|--------|------|------|
|0  |2022-01-01T14:49:50.459Z|         |msn.com  |2        |4     |2          |cars    |      |iphone|
|1  |2022-01-02T11:58:12.475Z|Yahoo    |yahoo.com|2        |4     |1          |cars    |1     |      |

Measures:

  - label: "Number of bids"
    expression: count(*)
    description: ""
    format_preset: ""
  - label: "Total volume"
    expression: sum(volume)
    description: ""
    format_preset: ""
  - label: "Total impressions"
    expression: sum(impressions)
  - label: "Total clicks"
    expression: sum(clicks)
*/
func TestServer_MetricsViewComparison_no_comparison(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewComparison(testCtx(), &runtimev1.MetricsViewComparisonRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		Dimension: &runtimev1.MetricsViewAggregationDimension{
			Name: "domain",
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "measure_2",
			},
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				Name: "measure_2",
				Desc: true,
			},
		},
		Exact: true,
	})
	require.NoError(t, err)
	require.Equal(t, 2, len(tr.Rows))

	require.Equal(t, 1, len(tr.Rows[0].MeasureValues))
	require.Equal(t, 1, len(tr.Rows[1].MeasureValues))

	require.Equal(t, "msn.com", tr.Rows[0].DimensionValue.GetStringValue())
	require.Equal(t, 2.0, tr.Rows[0].MeasureValues[0].BaseValue.GetNumberValue())

	require.Equal(t, "yahoo.com", tr.Rows[1].DimensionValue.GetStringValue())
	require.Equal(t, 1.0, tr.Rows[1].MeasureValues[0].BaseValue.GetNumberValue())
}

func TestServer_MetricsViewComparison_no_comparison_quotes(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewComparison(testCtx(), &runtimev1.MetricsViewComparisonRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		Dimension: &runtimev1.MetricsViewAggregationDimension{
			Name: "ad words",
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "measure_0",
			},
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				Name: "measure_0",
			},
		},
		Exact: true,
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Rows))

	require.Equal(t, 1, len(tr.Rows[0].MeasureValues))

	require.Equal(t, "cars", tr.Rows[0].DimensionValue.GetStringValue())
	require.Equal(t, 2.0, tr.Rows[0].MeasureValues[0].BaseValue.GetNumberValue())
}

func TestServer_MetricsViewComparison_no_comparison_numeric_dim(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewComparison(testCtx(), &runtimev1.MetricsViewComparisonRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		Dimension: &runtimev1.MetricsViewAggregationDimension{
			Name: "numeric_dim",
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "measure_0",
			},
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				Name: "measure_0",
			},
		},
		Exact: true,
	})

	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Rows))
	require.Equal(t, 1, len(tr.Rows[0].MeasureValues))

	require.Equal(t, float64(1), tr.Rows[0].DimensionValue.GetNumberValue())
	require.Equal(t, 2.0, tr.Rows[0].MeasureValues[0].BaseValue.GetNumberValue())
}

func Ignore_TestServer_MetricsViewComparison_no_comparison_HugeInt(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewComparison(testCtx(), &runtimev1.MetricsViewComparisonRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		Dimension: &runtimev1.MetricsViewAggregationDimension{
			Name: "id",
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "measure_2",
			},
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				Name: "measure_2",
			},
		},
		Exact: true,
	})
	require.NoError(t, err)
	require.Equal(t, 2, len(tr.Rows))

	require.Equal(t, 2, len(tr.Rows[0].MeasureValues))
	require.Equal(t, 2, len(tr.Rows[1].MeasureValues))

	require.Equal(t, "170141183460469231731687303715884105727", tr.Rows[0].DimensionValue.GetStringValue())
	require.Equal(t, 1.0, tr.Rows[0].MeasureValues[0].BaseValue.GetNumberValue())

	require.Equal(t, "170141183460469231731687303715884105726", tr.Rows[1].DimensionValue.GetStringValue())
	require.Equal(t, 2.0, tr.Rows[1].MeasureValues[0].BaseValue.GetNumberValue())
}

func TestServer_MetricsViewComparison_no_comparison_asc(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewComparison(testCtx(), &runtimev1.MetricsViewComparisonRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		Dimension: &runtimev1.MetricsViewAggregationDimension{
			Name: "domain",
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "measure_2",
			},
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				Name: "measure_2",
				Desc: false,
			},
		},
		Exact: true,
	})
	require.NoError(t, err)
	require.Equal(t, 2, len(tr.Rows))

	require.Equal(t, 1, len(tr.Rows[0].MeasureValues))
	require.Equal(t, 1, len(tr.Rows[1].MeasureValues))

	require.Equal(t, "yahoo.com", tr.Rows[0].DimensionValue.GetStringValue())
	require.Equal(t, 1.0, tr.Rows[0].MeasureValues[0].BaseValue.GetNumberValue())

	require.Equal(t, "msn.com", tr.Rows[1].DimensionValue.GetStringValue())
	require.Equal(t, 2.0, tr.Rows[1].MeasureValues[0].BaseValue.GetNumberValue())
}

func TestServer_MetricsViewComparison_no_comparison_nulls_last(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewComparison(testCtx(), &runtimev1.MetricsViewComparisonRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		Dimension: &runtimev1.MetricsViewAggregationDimension{
			Name: "domain",
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "measure_3",
			},
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				Name: "measure_3",
				Desc: false,
			},
		},
		Exact: true,
	})
	require.NoError(t, err)
	require.Equal(t, 2, len(tr.Rows))

	require.Equal(t, 1, len(tr.Rows[0].MeasureValues))
	require.Equal(t, 1, len(tr.Rows[1].MeasureValues))

	require.Equal(t, "yahoo.com", tr.Rows[0].DimensionValue.GetStringValue())
	require.Equal(t, 1.0, tr.Rows[0].MeasureValues[0].BaseValue.GetNumberValue())

	require.Equal(t, "msn.com", tr.Rows[1].DimensionValue.GetStringValue())
	require.Equal(t, structpb.NullValue(0), tr.Rows[1].MeasureValues[0].BaseValue.GetNullValue())

	tr, err = server.MetricsViewComparison(testCtx(), &runtimev1.MetricsViewComparisonRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		Dimension: &runtimev1.MetricsViewAggregationDimension{
			Name: "domain",
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "measure_3",
			},
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				Name: "measure_3",
				Desc: true,
			},
		},
		Exact: true,
	})

	require.NoError(t, err)
	require.Equal(t, 2, len(tr.Rows))

	require.Equal(t, 1, len(tr.Rows[0].MeasureValues))
	require.Equal(t, 1, len(tr.Rows[1].MeasureValues))

	require.Equal(t, "yahoo.com", tr.Rows[0].DimensionValue.GetStringValue())
	require.Equal(t, 1.0, tr.Rows[0].MeasureValues[0].BaseValue.GetNumberValue())

	require.Equal(t, "msn.com", tr.Rows[1].DimensionValue.GetStringValue())
	require.Equal(t, structpb.NullValue(0), tr.Rows[1].MeasureValues[0].BaseValue.GetNullValue())
}

func TestServer_MetricsViewComparison_no_comparison_asc_limit(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewComparison(testCtx(), &runtimev1.MetricsViewComparisonRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		Dimension: &runtimev1.MetricsViewAggregationDimension{
			Name: "domain",
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "measure_2",
			},
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				Name: "measure_2",
				Desc: false,
			},
		},
		Limit: 1,
		Exact: true,
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Rows))
	require.Equal(t, 1, len(tr.Rows[0].MeasureValues))

	require.Equal(t, "yahoo.com", tr.Rows[0].DimensionValue.GetStringValue())
	require.Equal(t, 1.0, tr.Rows[0].MeasureValues[0].BaseValue.GetNumberValue())
}

func TestServer_MetricsViewComparison_no_comparison_2measures(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewComparison(testCtx(), &runtimev1.MetricsViewComparisonRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		Dimension: &runtimev1.MetricsViewAggregationDimension{
			Name: "domain",
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "measure_0",
			},

			{
				Name: "measure_2",
			},
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				Name: "measure_0",
				Desc: false,
			},
			{
				Name: "measure_2",
				Desc: false,
			},
		},
		Exact: true,
	})
	require.NoError(t, err)
	require.Equal(t, 2, len(tr.Rows))
	require.Equal(t, 2, len(tr.Rows[0].MeasureValues))

	require.Equal(t, "yahoo.com", tr.Rows[0].DimensionValue.GetStringValue())
	require.Equal(t, 1.0, tr.Rows[0].MeasureValues[0].BaseValue.GetNumberValue())
	require.Equal(t, 1.0, tr.Rows[0].MeasureValues[1].BaseValue.GetNumberValue())

	require.Equal(t, "msn.com", tr.Rows[1].DimensionValue.GetStringValue())
	require.Equal(t, 1.0, tr.Rows[1].MeasureValues[0].BaseValue.GetNumberValue())
	require.Equal(t, 2.0, tr.Rows[1].MeasureValues[1].BaseValue.GetNumberValue())
}

func TestServer_MetricsViewComparison_no_comparison_complete_source_sanity_test(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids")

	tr, err := server.MetricsViewComparison(testCtx(), &runtimev1.MetricsViewComparisonRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		Dimension: &runtimev1.MetricsViewAggregationDimension{
			Name: "dom",
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "measure_0",
			},
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				Name: "measure_0",
				Desc: false,
			},
		},
		Where: expressionpb.NotIn(
			expressionpb.Identifier("pub"),
			[]*runtimev1.Expression{expressionpb.Value(structpb.NewStringValue("Yahoo"))},
		),
		Exact: true,
	})
	require.NoError(t, err)
	require.True(t, len(tr.Rows) > 1)
	require.Equal(t, 1, len(tr.Rows[0].MeasureValues))
}

func TestServer_MetricsViewComparison_no_comparison_dimension_expression(t *testing.T) {
	t.Parallel()
	srv, instanceId := getMetricsTestServer(t, "ad_bids")

	tr, err := srv.MetricsViewComparison(testCtx(), &runtimev1.MetricsViewComparisonRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		Dimension: &runtimev1.MetricsViewAggregationDimension{
			Name: "tld",
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "measure_0",
			},
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				Name: "measure_0",
				Desc: false,
			},
		},
		Where: expressionpb.NotLike(
			expressionpb.Identifier("dom"),
			expressionpb.Value(structpb.NewStringValue("%yahoo%")),
		),
		Exact: true,
	})
	require.NoError(t, err)
	require.Len(t, tr.Rows, 4)
	require.Equal(t, 1, len(tr.Rows[0].MeasureValues))
	require.Equal(t, "instagram.com", tr.Rows[0].DimensionValue.GetStringValue())
	require.Equal(t, "msn.com", tr.Rows[1].DimensionValue.GetStringValue())
	require.Equal(t, "facebook.com", tr.Rows[2].DimensionValue.GetStringValue())
	require.Equal(t, "google.com", tr.Rows[3].DimensionValue.GetStringValue())
}

func TestServer_MetricsViewComparison_no_comparison_unnested_dimension_expression_in_filter(t *testing.T) {
	t.Parallel()
	srv, instanceId := getMetricsTestServer(t, "ad_bids")

	tr, err := srv.MetricsViewComparison(testCtx(), &runtimev1.MetricsViewComparisonRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		Dimension: &runtimev1.MetricsViewAggregationDimension{
			Name: "domain_parts",
		},
		Measures: []*runtimev1.MetricsViewAggregationMeasure{
			{
				Name: "measure_0",
			},
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				Name: "measure_0",
				Desc: false,
			},
		},
		Where: expressionpb.NotIn(
			expressionpb.Identifier("domain_parts"),
			[]*runtimev1.Expression{expressionpb.Value(structpb.NewStringValue("yahoo")), expressionpb.Value(structpb.NewStringValue("google"))},
		),
		Exact: true,
	})
	require.NoError(t, err)
	require.Len(t, tr.Rows, 4)
	require.Equal(t, 1, len(tr.Rows[0].MeasureValues))
	require.Equal(t, "instagram", tr.Rows[0].DimensionValue.GetStringValue())
	require.Equal(t, "msn", tr.Rows[1].DimensionValue.GetStringValue())
	require.Equal(t, "facebook", tr.Rows[2].DimensionValue.GetStringValue())
	require.Equal(t, "com", tr.Rows[3].DimensionValue.GetStringValue())
}
