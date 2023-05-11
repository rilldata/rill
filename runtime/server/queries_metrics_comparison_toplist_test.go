package server

import (
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

func getMetricsTestServer(t *testing.T, projectName string) (*Server, string) {
	rt, instanceID := testruntime.NewInstanceForProject(t, projectName)

	server, err := NewServer(&Options{}, rt, nil)
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
func TestServer_MetricsViewComparisonToplist(t *testing.T) {
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewComparisonToplist(testCtx(), &runtimev1.MetricsViewComparisonToplistRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		DimensionName:   "ad words",
		MeasureNames:    []string{"measure_2"},
		BaseTimeRange: &runtimev1.TimeRange{
			Start: parseTime(t, "2022-01-01T00:00:00Z"),
			End:   parseTime(t, "2022-01-01T23:59:00Z"),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: parseTime(t, "2022-01-02T00:00:00Z"),
			End:   parseTime(t, "2022-01-02T23:59:00Z"),
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				MeasureName: "measure_2",
				Type:        runtimev1.ComparisonSortType_COMPARISON_SORT_TYPE_BASE_VALUE,
				Ascending:   true,
			},
		},
	})

	rows := tr.Rows
	require.NoError(t, err)
	require.Equal(t, 1, len(rows))
	require.Equal(t, "cars", rows[0].DimensionValue.GetStringValue())

	require.Equal(t, 2.0, rows[0].MeasureValues[0].BaseValue.GetNumberValue())
	require.Equal(t, 1.0, rows[0].MeasureValues[0].ComparisonValue.GetNumberValue())
	require.Equal(t, -1.0, rows[0].MeasureValues[0].DeltaAbs.GetNumberValue())
	require.Equal(t, -0.5, rows[0].MeasureValues[0].DeltaRel.GetNumberValue())
}

func TestServer_MetricsViewComparisonToplist_nulls(t *testing.T) {
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewComparisonToplist(testCtx(), &runtimev1.MetricsViewComparisonToplistRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		DimensionName:   "domain",
		MeasureNames:    []string{"measure_2"},
		BaseTimeRange: &runtimev1.TimeRange{
			Start: parseTime(t, "2022-01-01T00:00:00Z"),
			End:   parseTime(t, "2022-01-01T23:59:00Z"),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: parseTime(t, "2022-01-02T00:00:00Z"),
			End:   parseTime(t, "2022-01-02T23:59:00Z"),
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				MeasureName: "measure_2",
				Type:        runtimev1.ComparisonSortType_COMPARISON_SORT_TYPE_BASE_VALUE,
				Ascending:   true,
			},
		},
	})

	rows := tr.Rows
	require.NoError(t, err)
	require.Equal(t, 2, len(rows))

	require.Equal(t, "msn.com", rows[0].DimensionValue.GetStringValue())
	require.Equal(t, 2.0, rows[0].MeasureValues[0].BaseValue.GetNumberValue())
	require.Equal(t, structpb.NullValue(0), rows[0].MeasureValues[0].ComparisonValue.GetNullValue())
	require.Equal(t, structpb.NullValue(0), rows[0].MeasureValues[0].DeltaAbs.GetNullValue())
	require.Equal(t, structpb.NullValue(0), rows[0].MeasureValues[0].DeltaRel.GetNullValue())

	require.Equal(t, "yahoo.com", rows[1].DimensionValue.GetStringValue())
	require.Equal(t, structpb.NullValue(0), rows[1].MeasureValues[0].BaseValue.GetNullValue())
	require.Equal(t, 1.0, rows[1].MeasureValues[0].ComparisonValue.GetNumberValue())
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
func TestServer_MetricsViewComparisonToplist_sort_by_base(t *testing.T) {
	server, instanceId := getMetricsTestServer(t, "ad_bids")

	tr, err := server.MetricsViewComparisonToplist(testCtx(), &runtimev1.MetricsViewComparisonToplistRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_mini_metrics",
		DimensionName:   "domain",
		MeasureNames:    []string{"measure_2"},
		BaseTimeRange: &runtimev1.TimeRange{
			Start: parseTime(t, "2022-01-01T00:00:00Z"),
			End:   parseTime(t, "2022-01-02T23:59:00Z"),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: parseTime(t, "2022-01-03T00:00:00Z"),
			End:   parseTime(t, "2022-01-04T23:59:00Z"),
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				MeasureName: "measure_2",
				Type:        runtimev1.ComparisonSortType_COMPARISON_SORT_TYPE_BASE_VALUE,
				Ascending:   false,
			},
		},
	})

	rows := tr.Rows
	require.NoError(t, err)
	require.Equal(t, 2, len(rows))

	require.Equal(t, "msn.com", rows[0].DimensionValue.GetStringValue())
	require.Equal(t, 2.0, rows[0].MeasureValues[0].BaseValue.GetNumberValue())
	require.Equal(t, 1.0, rows[0].MeasureValues[0].ComparisonValue.GetNumberValue())
	require.Equal(t, -1.0, rows[0].MeasureValues[0].DeltaAbs.GetNumberValue())
	require.Equal(t, -0.5, rows[0].MeasureValues[0].DeltaRel.GetNumberValue())

	require.Equal(t, "yahoo.com", rows[1].DimensionValue.GetStringValue())
	require.Equal(t, 1.0, rows[1].MeasureValues[0].BaseValue.GetNumberValue())
	require.Equal(t, 2.0, rows[1].MeasureValues[0].ComparisonValue.GetNumberValue())
	require.Equal(t, 1.0, rows[1].MeasureValues[0].DeltaAbs.GetNumberValue())
	require.Equal(t, 1.0, rows[1].MeasureValues[0].DeltaRel.GetNumberValue())
}

/*
the result should be:

|domain                  |base |comparison|delta|rel    |
|------------------------|-----|----------|-----|-------|
|yahoo.com               |1    |2         |1    |1      |
|msn.com                 |2    |1         | -1  |-0.5   |
*/
func TestServer_MetricsViewComparisonToplist_sort_by_comparison(t *testing.T) {
	server, instanceId := getMetricsTestServer(t, "ad_bids")

	tr, err := server.MetricsViewComparisonToplist(testCtx(), &runtimev1.MetricsViewComparisonToplistRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_mini_metrics",
		DimensionName:   "domain",
		MeasureNames:    []string{"measure_2"},
		BaseTimeRange: &runtimev1.TimeRange{
			Start: parseTime(t, "2022-01-01T00:00:00Z"),
			End:   parseTime(t, "2022-01-02T23:59:00Z"),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: parseTime(t, "2022-01-03T00:00:00Z"),
			End:   parseTime(t, "2022-01-04T23:59:00Z"),
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				MeasureName: "measure_2",
				Type:        runtimev1.ComparisonSortType_COMPARISON_SORT_TYPE_COMPARISON_VALUE,
				Ascending:   false,
			},
		},
	})

	rows := tr.Rows
	require.NoError(t, err)
	require.Equal(t, 2, len(rows))

	require.Equal(t, "yahoo.com", rows[0].DimensionValue.GetStringValue())
	require.Equal(t, 1.0, rows[0].MeasureValues[0].BaseValue.GetNumberValue())
	require.Equal(t, 2.0, rows[0].MeasureValues[0].ComparisonValue.GetNumberValue())
	require.Equal(t, 1.0, rows[0].MeasureValues[0].DeltaAbs.GetNumberValue())
	require.Equal(t, 1.0, rows[0].MeasureValues[0].DeltaRel.GetNumberValue())

	require.Equal(t, "msn.com", rows[1].DimensionValue.GetStringValue())
	require.Equal(t, 2.0, rows[1].MeasureValues[0].BaseValue.GetNumberValue())
	require.Equal(t, 1.0, rows[1].MeasureValues[0].ComparisonValue.GetNumberValue())
	require.Equal(t, -1.0, rows[1].MeasureValues[0].DeltaAbs.GetNumberValue())
	require.Equal(t, -0.5, rows[1].MeasureValues[0].DeltaRel.GetNumberValue())
}

/*
the result should be:

|domain                  |base |comparison|delta|rel    |
|------------------------|-----|----------|-----|-------|
|yahoo.com               |100  |200       |100  |1      |
|msn.com                 |1    |10        |9    |9      |
*/

func TestServer_MetricsViewComparisonToplist_sort_by_abs_delta(t *testing.T) {
	server, instanceId := getMetricsTestServer(t, "ad_bids")

	tr, err := server.MetricsViewComparisonToplist(testCtx(), &runtimev1.MetricsViewComparisonToplistRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_mini_metrics",
		DimensionName:   "domain",
		MeasureNames:    []string{"measure_1"},
		BaseTimeRange: &runtimev1.TimeRange{
			Start: parseTime(t, "2022-01-01T00:00:00Z"),
			End:   parseTime(t, "2022-01-02T23:59:00Z"),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: parseTime(t, "2022-01-03T00:00:00Z"),
			End:   parseTime(t, "2022-01-04T23:59:00Z"),
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				MeasureName: "measure_1",
				Type:        runtimev1.ComparisonSortType_COMPARISON_SORT_TYPE_ABS_DELTA,
				Ascending:   false,
			},
		},
	})

	rows := tr.Rows
	require.NoError(t, err)
	require.Equal(t, 2, len(rows))

	require.Equal(t, "yahoo.com", rows[0].DimensionValue.GetStringValue())
	require.Equal(t, 100.0, rows[0].MeasureValues[0].BaseValue.GetNumberValue())
	require.Equal(t, 200.0, rows[0].MeasureValues[0].ComparisonValue.GetNumberValue())
	require.Equal(t, 100.0, rows[0].MeasureValues[0].DeltaAbs.GetNumberValue())
	require.Equal(t, 1.0, rows[0].MeasureValues[0].DeltaRel.GetNumberValue())

	require.Equal(t, "msn.com", rows[1].DimensionValue.GetStringValue())
	require.Equal(t, 1.0, rows[1].MeasureValues[0].BaseValue.GetNumberValue())
	require.Equal(t, 10.0, rows[1].MeasureValues[0].ComparisonValue.GetNumberValue())
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
func TestServer_MetricsViewComparisonToplist_sort_by_rel_delta(t *testing.T) {
	server, instanceId := getMetricsTestServer(t, "ad_bids")

	tr, err := server.MetricsViewComparisonToplist(testCtx(), &runtimev1.MetricsViewComparisonToplistRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_mini_metrics",
		DimensionName:   "domain",
		MeasureNames:    []string{"measure_1"},
		BaseTimeRange: &runtimev1.TimeRange{
			Start: parseTime(t, "2022-01-01T00:00:00Z"),
			End:   parseTime(t, "2022-01-02T23:59:00Z"),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: parseTime(t, "2022-01-03T00:00:00Z"),
			End:   parseTime(t, "2022-01-04T23:59:00Z"),
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				MeasureName: "measure_1",
				Type:        runtimev1.ComparisonSortType_COMPARISON_SORT_TYPE_REL_DELTA,
				Ascending:   false,
			},
		},
	})

	rows := tr.Rows
	require.NoError(t, err)
	require.Equal(t, 2, len(rows))

	require.Equal(t, "msn.com", rows[0].DimensionValue.GetStringValue())
	require.Equal(t, 1.0, rows[0].MeasureValues[0].BaseValue.GetNumberValue())
	require.Equal(t, 10.0, rows[0].MeasureValues[0].ComparisonValue.GetNumberValue())
	require.Equal(t, 9.0, rows[0].MeasureValues[0].DeltaAbs.GetNumberValue())
	require.Equal(t, 9.0, rows[0].MeasureValues[0].DeltaRel.GetNumberValue())

	require.Equal(t, "yahoo.com", rows[1].DimensionValue.GetStringValue())
	require.Equal(t, 100.0, rows[1].MeasureValues[0].BaseValue.GetNumberValue())
	require.Equal(t, 200.0, rows[1].MeasureValues[0].ComparisonValue.GetNumberValue())
	require.Equal(t, 100.0, rows[1].MeasureValues[0].DeltaAbs.GetNumberValue())
	require.Equal(t, 1.0, rows[1].MeasureValues[0].DeltaRel.GetNumberValue())
}

func TestServer_MetricsViewComparisonToplist_sort_error(t *testing.T) {
	server, instanceId := getMetricsTestServer(t, "ad_bids")

	_, err := server.MetricsViewComparisonToplist(testCtx(), &runtimev1.MetricsViewComparisonToplistRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_mini_metrics",
		DimensionName:   "domain",
		MeasureNames:    []string{"measure_2"},
		BaseTimeRange: &runtimev1.TimeRange{
			Start: parseTime(t, "2022-01-01T00:00:00Z"),
			End:   parseTime(t, "2022-01-02T23:59:00Z"),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: parseTime(t, "2022-01-03T00:00:00Z"),
			End:   parseTime(t, "2022-01-04T23:59:00Z"),
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				MeasureName: "measure_1",
				Type:        runtimev1.ComparisonSortType_COMPARISON_SORT_TYPE_ABS_DELTA,
				Ascending:   false,
			},
		},
	})
	require.Error(t, err)
}

/*
the result should be:

|domain                  |base |comparison|delta|rel    |
|------------------------|-----|----------|-----|-------|
|yahoo.com               |1    |2         |1    |1      |
*/

func TestServer_MetricsViewComparisonToplist_sort_by_delta_limit_1(t *testing.T) {
	server, instanceId := getMetricsTestServer(t, "ad_bids")

	tr, err := server.MetricsViewComparisonToplist(testCtx(), &runtimev1.MetricsViewComparisonToplistRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_mini_metrics",
		DimensionName:   "domain",
		MeasureNames:    []string{"measure_2"},
		BaseTimeRange: &runtimev1.TimeRange{
			Start: parseTime(t, "2022-01-01T00:00:00Z"),
			End:   parseTime(t, "2022-01-02T23:59:00Z"),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: parseTime(t, "2022-01-03T00:00:00Z"),
			End:   parseTime(t, "2022-01-04T23:59:00Z"),
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				MeasureName: "measure_2",
				Type:        runtimev1.ComparisonSortType_COMPARISON_SORT_TYPE_ABS_DELTA,
				Ascending:   false,
			},
		},
		Limit: 1,
	})

	rows := tr.Rows
	require.NoError(t, err)
	require.Equal(t, 1, len(rows))

	require.Equal(t, "yahoo.com", rows[0].DimensionValue.GetStringValue())
	require.Equal(t, 1.0, rows[0].MeasureValues[0].BaseValue.GetNumberValue())
	require.Equal(t, 2.0, rows[0].MeasureValues[0].ComparisonValue.GetNumberValue())
	require.Equal(t, 1.0, rows[0].MeasureValues[0].DeltaAbs.GetNumberValue())
	require.Equal(t, 1.0, rows[0].MeasureValues[0].DeltaRel.GetNumberValue())
}

func TestServer_MetricsViewComparisonToplist_sort_by_base_limit_1(t *testing.T) {
	server, instanceId := getMetricsTestServer(t, "ad_bids")

	tr, err := server.MetricsViewComparisonToplist(testCtx(), &runtimev1.MetricsViewComparisonToplistRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_mini_metrics",
		DimensionName:   "domain",
		MeasureNames:    []string{"measure_2"},
		BaseTimeRange: &runtimev1.TimeRange{
			Start: parseTime(t, "2022-01-01T00:00:00Z"),
			End:   parseTime(t, "2022-01-02T23:59:00Z"),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: parseTime(t, "2022-01-03T00:00:00Z"),
			End:   parseTime(t, "2022-01-04T23:59:00Z"),
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				MeasureName: "measure_2",
				Type:        runtimev1.ComparisonSortType_COMPARISON_SORT_TYPE_BASE_VALUE,
				Ascending:   false,
			},
		},
		Limit: 1,
	})

	rows := tr.Rows
	require.NoError(t, err)
	require.Equal(t, 1, len(rows))

	require.Equal(t, "msn.com", rows[0].DimensionValue.GetStringValue())
	require.Equal(t, 2.0, rows[0].MeasureValues[0].BaseValue.GetNumberValue())
	require.Equal(t, 1.0, rows[0].MeasureValues[0].ComparisonValue.GetNumberValue())
	require.Equal(t, -1.0, rows[0].MeasureValues[0].DeltaAbs.GetNumberValue())
	require.Equal(t, -0.5, rows[0].MeasureValues[0].DeltaRel.GetNumberValue())
}

func TestServer_MetricsViewComparisonToplist_sort_by_base_filter(t *testing.T) {
	server, instanceId := getMetricsTestServer(t, "ad_bids")

	tr, err := server.MetricsViewComparisonToplist(testCtx(), &runtimev1.MetricsViewComparisonToplistRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_mini_metrics",
		DimensionName:   "domain",
		MeasureNames:    []string{"measure_2"},
		BaseTimeRange: &runtimev1.TimeRange{
			Start: parseTime(t, "2022-01-01T00:00:00Z"),
			End:   parseTime(t, "2022-01-02T23:59:00Z"),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: parseTime(t, "2022-01-03T00:00:00Z"),
			End:   parseTime(t, "2022-01-04T23:59:00Z"),
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				MeasureName: "measure_2",
				Type:        runtimev1.ComparisonSortType_COMPARISON_SORT_TYPE_BASE_VALUE,
				Ascending:   false,
			},
		},
		Filter: &runtimev1.MetricsViewFilter{
			Exclude: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "domain",
					In:   []*structpb.Value{structpb.NewStringValue("yahoo.com")},
				},
			},
		},
	})

	rows := tr.Rows
	require.NoError(t, err)
	require.Equal(t, 1, len(rows))

	require.Equal(t, "msn.com", rows[0].DimensionValue.GetStringValue())
	require.Equal(t, 2.0, rows[0].MeasureValues[0].BaseValue.GetNumberValue())
	require.Equal(t, 1.0, rows[0].MeasureValues[0].ComparisonValue.GetNumberValue())
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
func TestServer_MetricsViewComparisonToplist_2_measures(t *testing.T) {
	server, instanceId := getMetricsTestServer(t, "ad_bids")

	tr, err := server.MetricsViewComparisonToplist(testCtx(), &runtimev1.MetricsViewComparisonToplistRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_mini_metrics",
		DimensionName:   "domain",
		MeasureNames:    []string{"measure_1", "measure_2"},
		BaseTimeRange: &runtimev1.TimeRange{
			Start: parseTime(t, "2022-01-01T00:00:00Z"),
			End:   parseTime(t, "2022-01-02T23:59:00Z"),
		},
		ComparisonTimeRange: &runtimev1.TimeRange{
			Start: parseTime(t, "2022-01-03T00:00:00Z"),
			End:   parseTime(t, "2022-01-04T23:59:00Z"),
		},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				MeasureName: "measure_2",
				Type:        runtimev1.ComparisonSortType_COMPARISON_SORT_TYPE_ABS_DELTA,
				Ascending:   false,
			},
		},
	})

	rows := tr.Rows
	require.NoError(t, err)
	require.Equal(t, 2, len(rows))

	require.Equal(t, "yahoo.com", rows[0].DimensionValue.GetStringValue())
	require.Equal(t, 100.0, rows[0].MeasureValues[0].BaseValue.GetNumberValue())
	require.Equal(t, 200.0, rows[0].MeasureValues[0].ComparisonValue.GetNumberValue())
	require.Equal(t, 100.0, rows[0].MeasureValues[0].DeltaAbs.GetNumberValue())
	require.Equal(t, 1.0, rows[0].MeasureValues[0].DeltaRel.GetNumberValue())
	require.Equal(t, 1.0, rows[0].MeasureValues[1].BaseValue.GetNumberValue())
	require.Equal(t, 2.0, rows[0].MeasureValues[1].ComparisonValue.GetNumberValue())
	require.Equal(t, 1.0, rows[0].MeasureValues[1].DeltaAbs.GetNumberValue())
	require.Equal(t, 1.0, rows[0].MeasureValues[1].DeltaRel.GetNumberValue())

	require.Equal(t, "msn.com", rows[1].DimensionValue.GetStringValue())
	require.Equal(t, 1.0, rows[1].MeasureValues[0].BaseValue.GetNumberValue())
	require.Equal(t, 10.0, rows[1].MeasureValues[0].ComparisonValue.GetNumberValue())
	require.Equal(t, 9.0, rows[1].MeasureValues[0].DeltaAbs.GetNumberValue())
	require.Equal(t, 9.0, rows[1].MeasureValues[0].DeltaRel.GetNumberValue())
	require.Equal(t, 2.0, rows[1].MeasureValues[1].BaseValue.GetNumberValue())
	require.Equal(t, 1.0, rows[1].MeasureValues[1].ComparisonValue.GetNumberValue())
	require.Equal(t, -1.0, rows[1].MeasureValues[1].DeltaAbs.GetNumberValue())
	require.Equal(t, -0.5, rows[1].MeasureValues[1].DeltaRel.GetNumberValue())
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
func TestServer_MetricsViewComparisonToplist_no_comparison(t *testing.T) {
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewComparisonToplist(testCtx(), &runtimev1.MetricsViewComparisonToplistRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		DimensionName:   "domain",
		MeasureNames:    []string{"measure_2"},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				MeasureName: "measure_2",
				Ascending:   false,
			},
		},
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

func TestServer_MetricsViewComparisonToplist_no_comparison_quotes(t *testing.T) {
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewComparisonToplist(testCtx(), &runtimev1.MetricsViewComparisonToplistRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		DimensionName:   "ad words",
		MeasureNames:    []string{"measure_0"},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				MeasureName: "measure_0",
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Rows))

	require.Equal(t, 1, len(tr.Rows[0].MeasureValues))

	require.Equal(t, "cars", tr.Rows[0].DimensionValue.GetStringValue())
	require.Equal(t, 2.0, tr.Rows[0].MeasureValues[0].BaseValue.GetNumberValue())
}

func TestServer_MetricsViewComparisonToplist_no_comparison_numeric_dim(t *testing.T) {
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewComparisonToplist(testCtx(), &runtimev1.MetricsViewComparisonToplistRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		DimensionName:   "numeric_dim",
		MeasureNames:    []string{"measure_0"},
	})

	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Rows))
	require.Equal(t, 1, len(tr.Rows[0].MeasureValues))

	require.Equal(t, float64(1), tr.Rows[0].DimensionValue.GetNumberValue())
	require.Equal(t, 2.0, tr.Rows[0].MeasureValues[0].BaseValue.GetNumberValue())
}

func Ignore_TestServer_MetricsViewComparisonToplist_no_comparison_HugeInt(t *testing.T) {
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewComparisonToplist(testCtx(), &runtimev1.MetricsViewComparisonToplistRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		DimensionName:   "id",
		MeasureNames:    []string{"measure_2"},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				MeasureName: "measure_2",
			},
		},
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

func TestServer_MetricsViewComparisonToplist_no_comparison_asc(t *testing.T) {
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewComparisonToplist(testCtx(), &runtimev1.MetricsViewComparisonToplistRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		DimensionName:   "domain",
		MeasureNames:    []string{"measure_2"},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				MeasureName: "measure_2",
				Ascending:   true,
			},
		},
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

func TestServer_MetricsViewComparisonToplist_no_comparison_nulls_last(t *testing.T) {
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewComparisonToplist(testCtx(), &runtimev1.MetricsViewComparisonToplistRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		DimensionName:   "domain",
		MeasureNames:    []string{"measure_3"},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				MeasureName: "measure_3",
				Ascending:   true,
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, 2, len(tr.Rows))

	require.Equal(t, 1, len(tr.Rows[0].MeasureValues))
	require.Equal(t, 1, len(tr.Rows[1].MeasureValues))

	require.Equal(t, "yahoo.com", tr.Rows[0].DimensionValue.GetStringValue())
	require.Equal(t, 1.0, tr.Rows[0].MeasureValues[0].BaseValue.GetNumberValue())

	require.Equal(t, "msn.com", tr.Rows[1].DimensionValue.GetStringValue())
	require.Equal(t, structpb.NullValue(0), tr.Rows[1].MeasureValues[0].BaseValue.GetNullValue())

	tr, err = server.MetricsViewComparisonToplist(testCtx(), &runtimev1.MetricsViewComparisonToplistRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		DimensionName:   "domain",
		MeasureNames:    []string{"measure_3"},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				MeasureName: "measure_3",
				Ascending:   false,
			},
		},
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

func TestServer_MetricsViewComparisonToplist_no_comparison_asc_limit(t *testing.T) {
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewComparisonToplist(testCtx(), &runtimev1.MetricsViewComparisonToplistRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		DimensionName:   "domain",
		MeasureNames:    []string{"measure_2"},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				MeasureName: "measure_2",
				Ascending:   true,
			},
		},
		Limit: 1,
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(tr.Rows))
	require.Equal(t, 1, len(tr.Rows[0].MeasureValues))

	require.Equal(t, "yahoo.com", tr.Rows[0].DimensionValue.GetStringValue())
	require.Equal(t, 1.0, tr.Rows[0].MeasureValues[0].BaseValue.GetNumberValue())
}

func TestServer_MetricsViewComparisonToplist_no_comparison_2measures(t *testing.T) {
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	tr, err := server.MetricsViewComparisonToplist(testCtx(), &runtimev1.MetricsViewComparisonToplistRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		DimensionName:   "domain",
		MeasureNames:    []string{"measure_0", "measure_2"},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				MeasureName: "measure_0",
				Ascending:   true,
			},
			{
				MeasureName: "measure_2",
				Ascending:   true,
			},
		},
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

func TestServer_MetricsViewComparisonToplist_no_comparison_complete_source_sanity_test(t *testing.T) {
	server, instanceId := getMetricsTestServer(t, "ad_bids")

	tr, err := server.MetricsViewComparisonToplist(testCtx(), &runtimev1.MetricsViewComparisonToplistRequest{
		InstanceId:      instanceId,
		MetricsViewName: "ad_bids_metrics",
		DimensionName:   "domain",
		MeasureNames:    []string{"measure_0"},
		Sort: []*runtimev1.MetricsViewComparisonSort{
			{
				MeasureName: "measure_0",
				Ascending:   true,
			},
		},
		Filter: &runtimev1.MetricsViewFilter{
			Exclude: []*runtimev1.MetricsViewFilter_Cond{
				{
					Name: "publisher",
					In: []*structpb.Value{
						structpb.NewStringValue("Yahoo"),
					},
				},
			},
		},
	})
	require.NoError(t, err)
	require.True(t, len(tr.Rows) > 1)
	require.Equal(t, 1, len(tr.Rows[0].MeasureValues))
}
