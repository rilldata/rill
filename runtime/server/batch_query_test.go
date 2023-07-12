package server

import (
	"context"
	"fmt"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestServer_QueryBatch_MetricsViewQueries(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	// Send all types of metrics view requests
	req := &runtimev1.QueryBatchRequest{
		InstanceId: instanceId,
		Queries: []*runtimev1.QueryBatchEntry{
			{
				Key: 0,
				Query: &runtimev1.QueryBatchEntry_MetricsViewToplistRequest{
					MetricsViewToplistRequest: &runtimev1.MetricsViewToplistRequest{
						MetricsViewName: "ad_bids_metrics",
						DimensionName:   "domain",
						MeasureNames:    []string{"measure_2"},
						Sort: []*runtimev1.MetricsViewSort{
							{
								Name: "measure_2",
							},
						},
					},
				},
			},
			{
				Key: 1,
				Query: &runtimev1.QueryBatchEntry_MetricsViewComparisonToplistRequest{
					MetricsViewComparisonToplistRequest: &runtimev1.MetricsViewComparisonToplistRequest{
						MetricsViewName: "ad_bids_metrics",
						DimensionName:   "ad words",
						MeasureNames:    []string{"measure_2"},
						BaseTimeRange: &runtimev1.TimeRange{
							Start: parseTimeToProtoTimeStamps(t, "2022-01-01T00:00:00Z"),
							End:   parseTimeToProtoTimeStamps(t, "2022-01-01T23:59:00Z"),
						},
						ComparisonTimeRange: &runtimev1.TimeRange{
							Start: parseTimeToProtoTimeStamps(t, "2022-01-02T00:00:00Z"),
							End:   parseTimeToProtoTimeStamps(t, "2022-01-02T23:59:00Z"),
						},
						Sort: []*runtimev1.MetricsViewComparisonSort{
							{
								MeasureName: "measure_2",
								Type:        runtimev1.MetricsViewComparisonSortType_METRICS_VIEW_COMPARISON_SORT_TYPE_BASE_VALUE,
								Ascending:   true,
							},
						},
					},
				},
			},
			{
				Key: 2,
				Query: &runtimev1.QueryBatchEntry_MetricsViewTimeSeriesRequest{
					MetricsViewTimeSeriesRequest: &runtimev1.MetricsViewTimeSeriesRequest{
						MetricsViewName: "ad_bids_metrics",
						TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_DAY,
						MeasureNames:    []string{"measure_0", "measure_2"},
					},
				},
			},
			{
				Key: 3,
				Query: &runtimev1.QueryBatchEntry_MetricsViewTotalsRequest{
					MetricsViewTotalsRequest: &runtimev1.MetricsViewTotalsRequest{
						MetricsViewName: "ad_bids_metrics",
						MeasureNames:    []string{"measure_0"},
					},
				},
			},
			{
				Key: 4,
				Query: &runtimev1.QueryBatchEntry_MetricsViewRowsRequest{
					MetricsViewRowsRequest: &runtimev1.MetricsViewRowsRequest{
						MetricsViewName: "ad_bids_metrics",
					},
				},
			},
		},
	}

	batchServer := &fakeBatchServer{
		responses: make([]*runtimev1.QueryBatchResponse, 0),
		ctx:       testCtx(),
	}
	err := server.QueryBatch(req, batchServer)
	require.NoError(t, err)
	require.Equal(t, len(req.Queries), len(batchServer.responses))

	haveResponse := make([]bool, len(req.Queries))
	for _, response := range batchServer.responses {
		require.Empty(t, response.Error, fmt.Sprintf("request errored for %T", req.Queries[response.Key].Query))
		require.False(t, haveResponse[response.Key], fmt.Sprintf("duplicate response for %T", req.Queries[response.Key].Query))
		haveResponse[response.Key] = true

		switch response.Key {
		case 0:
			require.IsType(t, &runtimev1.QueryBatchResponse_MetricsViewToplistResponse{}, response.Result)
			tr := response.GetMetricsViewToplistResponse()
			require.Equal(t, 2, len(tr.Data))
			require.Equal(t, 2, len(tr.Data[0].Fields))
			require.Equal(t, 2, len(tr.Data[1].Fields))

		case 1:
			require.IsType(t, &runtimev1.QueryBatchResponse_MetricsViewComparisonToplistResponse{}, response.Result)
			tr := response.GetMetricsViewComparisonToplistResponse()
			rows := tr.Rows
			require.NoError(t, err)
			require.Equal(t, 1, len(rows))

		case 2:
			require.IsType(t, &runtimev1.QueryBatchResponse_MetricsViewTimeSeriesResponse{}, response.Result)
			tr := response.GetMetricsViewTimeSeriesResponse()
			require.Equal(t, 2, len(tr.Data))
			require.Equal(t, 2, len(tr.Data[0].Records.Fields))

		case 3:
			require.IsType(t, &runtimev1.QueryBatchResponse_MetricsViewTotalsResponse{}, response.Result)
			tr := response.GetMetricsViewTotalsResponse()
			require.Equal(t, 1, len(tr.Data.Fields))

		case 4:
			require.IsType(t, &runtimev1.QueryBatchResponse_MetricsViewRowsResponse{}, response.Result)
			tr := response.GetMetricsViewRowsResponse()
			require.Equal(t, 2, len(tr.Data))
		}
	}
}

func TestServer_QueryBatch_ColumnQueries(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	// Send all types of column query requests
	req := &runtimev1.QueryBatchRequest{
		InstanceId: instanceId,
		Queries: []*runtimev1.QueryBatchEntry{
			{
				Key: 0,
				Query: &runtimev1.QueryBatchEntry_ColumnRollupIntervalRequest{
					ColumnRollupIntervalRequest: &runtimev1.ColumnRollupIntervalRequest{
						TableName:  "ad_bids",
						ColumnName: "timestamp",
					},
				},
			},
			{
				Key: 1,
				Query: &runtimev1.QueryBatchEntry_ColumnTopKRequest{
					ColumnTopKRequest: &runtimev1.ColumnTopKRequest{
						TableName:  "ad_bids",
						ColumnName: "publisher",
						K:          1,
					},
				},
			},
			{
				Key: 2,
				Query: &runtimev1.QueryBatchEntry_ColumnNullCountRequest{
					ColumnNullCountRequest: &runtimev1.ColumnNullCountRequest{
						TableName:  "ad_bids",
						ColumnName: "publisher",
					},
				},
			},
			{
				Key: 3,
				Query: &runtimev1.QueryBatchEntry_ColumnDescriptiveStatisticsRequest{
					ColumnDescriptiveStatisticsRequest: &runtimev1.ColumnDescriptiveStatisticsRequest{
						TableName:  "ad_bids",
						ColumnName: "impressions",
					},
				},
			},
			{
				Key: 4,
				Query: &runtimev1.QueryBatchEntry_ColumnTimeGrainRequest{
					ColumnTimeGrainRequest: &runtimev1.ColumnTimeGrainRequest{
						TableName:  "ad_bids",
						ColumnName: "timestamp",
					},
				},
			},
			{
				Key: 5,
				Query: &runtimev1.QueryBatchEntry_ColumnNumericHistogramRequest{
					ColumnNumericHistogramRequest: &runtimev1.ColumnNumericHistogramRequest{
						TableName:       "ad_bids",
						ColumnName:      "impressions",
						HistogramMethod: runtimev1.HistogramMethod_HISTOGRAM_METHOD_DIAGNOSTIC,
					},
				},
			},
			{
				Key: 6,
				Query: &runtimev1.QueryBatchEntry_ColumnRugHistogramRequest{
					ColumnRugHistogramRequest: &runtimev1.ColumnRugHistogramRequest{
						TableName:  "ad_bids",
						ColumnName: "impressions",
					},
				},
			},
			{
				Key: 7,
				Query: &runtimev1.QueryBatchEntry_ColumnTimeRangeRequest{
					ColumnTimeRangeRequest: &runtimev1.ColumnTimeRangeRequest{
						TableName:  "ad_bids",
						ColumnName: "timestamp",
					},
				},
			},
			{
				Key: 8,
				Query: &runtimev1.QueryBatchEntry_ColumnCardinalityRequest{
					ColumnCardinalityRequest: &runtimev1.ColumnCardinalityRequest{
						TableName:  "ad_bids",
						ColumnName: "domain",
					},
				},
			},
			{
				Key: 9,
				Query: &runtimev1.QueryBatchEntry_ColumnTimeSeriesRequest{
					ColumnTimeSeriesRequest: &runtimev1.ColumnTimeSeriesRequest{
						TableName:           "ad_bids",
						TimestampColumnName: "timestamp",
						TimeRange: &runtimev1.TimeSeriesTimeRange{
							Interval: runtimev1.TimeGrain_TIME_GRAIN_DAY,
						},
					},
				},
			},
		},
	}

	batchServer := &fakeBatchServer{
		responses: make([]*runtimev1.QueryBatchResponse, 0),
		ctx:       testCtx(),
	}
	err := server.QueryBatch(req, batchServer)
	require.NoError(t, err)
	require.Equal(t, len(req.Queries), len(batchServer.responses))

	haveResponse := make([]bool, len(req.Queries))
	for _, response := range batchServer.responses {
		require.Empty(t, response.Error, fmt.Sprintf("request errored for %T", req.Queries[response.Key].Query))
		require.False(t, haveResponse[response.Key], fmt.Sprintf("duplicate response for %T", req.Queries[response.Key].Query))
		haveResponse[response.Key] = true

		switch response.Key {
		case 0:
			require.IsType(t, &runtimev1.QueryBatchResponse_ColumnRollupIntervalResponse{}, response.Result)
			tr := response.GetColumnRollupIntervalResponse()
			require.Equal(t, 1, tr.Start.AsTime().Day())
			require.Equal(t, 2, tr.End.AsTime().Day())

		case 1:
			require.IsType(t, &runtimev1.QueryBatchResponse_ColumnTopKResponse{}, response.Result)
			tr := response.GetColumnTopKResponse()
			require.Equal(t, 1, len(tr.CategoricalSummary.GetTopK().Entries))

		case 2:
			require.IsType(t, &runtimev1.QueryBatchResponse_ColumnNullCountResponse{}, response.Result)
			tr := response.GetColumnNullCountResponse()
			require.Equal(t, 1.0, tr.Count)

		case 3:
			require.IsType(t, &runtimev1.QueryBatchResponse_ColumnDescriptiveStatisticsResponse{}, response.Result)
			tr := response.GetColumnDescriptiveStatisticsResponse()
			require.Equal(t, 1.0, tr.NumericSummary.GetNumericStatistics().Min)
			require.Equal(t, 2.0, tr.NumericSummary.GetNumericStatistics().Max)

		case 4:
			require.IsType(t, &runtimev1.QueryBatchResponse_ColumnTimeGrainResponse{}, response.Result)
			tr := response.GetColumnTimeGrainResponse()
			require.Equal(t, runtimev1.TimeGrain_TIME_GRAIN_MILLISECOND, tr.TimeGrain)

		case 5:
			require.IsType(t, &runtimev1.QueryBatchResponse_ColumnNumericHistogramResponse{}, response.Result)
			tr := response.GetColumnNumericHistogramResponse()
			require.Equal(t, 2, len(tr.NumericSummary.GetNumericHistogramBins().Bins))

		case 6:
			require.IsType(t, &runtimev1.QueryBatchResponse_ColumnRugHistogramResponse{}, response.Result)
			tr := response.GetColumnRugHistogramResponse()
			require.Equal(t, 2, len(tr.NumericSummary.GetNumericOutliers().Outliers))

		case 7:
			require.IsType(t, &runtimev1.QueryBatchResponse_ColumnTimeRangeResponse{}, response.Result)
			tr := response.GetColumnTimeRangeResponse()
			require.Equal(t, 1, tr.TimeRangeSummary.Min.AsTime().Day())
			require.Equal(t, 2, tr.TimeRangeSummary.Max.AsTime().Day())

		case 8:
			require.IsType(t, &runtimev1.QueryBatchResponse_ColumnCardinalityResponse{}, response.Result)
			tr := response.GetColumnCardinalityResponse()
			require.Equal(t, 2.0, tr.CategoricalSummary.GetCardinality())

		case 9:
			require.IsType(t, &runtimev1.QueryBatchResponse_ColumnTimeSeriesResponse{}, response.Result)
			tr := response.GetColumnTimeSeriesResponse()
			require.Equal(t, 2, len(tr.Rollup.Results))
		}
	}
}

func TestServer_QueryBatch_TableQueries(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	// Send all types of column query requests
	req := &runtimev1.QueryBatchRequest{
		InstanceId: instanceId,
		Queries: []*runtimev1.QueryBatchEntry{
			{
				Key: 0,
				Query: &runtimev1.QueryBatchEntry_TableCardinalityRequest{
					TableCardinalityRequest: &runtimev1.TableCardinalityRequest{
						TableName: "ad_bids",
					},
				},
			},
			{
				Key: 1,
				Query: &runtimev1.QueryBatchEntry_TableColumnsRequest{
					TableColumnsRequest: &runtimev1.TableColumnsRequest{
						TableName: "ad_bids",
					},
				},
			},
			{
				Key: 2,
				Query: &runtimev1.QueryBatchEntry_TableRowsRequest{
					TableRowsRequest: &runtimev1.TableRowsRequest{
						TableName: "ad_bids",
					},
				},
			},
		},
	}

	batchServer := &fakeBatchServer{
		responses: make([]*runtimev1.QueryBatchResponse, 0),
		ctx:       testCtx(),
	}
	err := server.QueryBatch(req, batchServer)
	require.NoError(t, err)
	require.Equal(t, len(req.Queries), len(batchServer.responses))

	haveResponse := make([]bool, len(req.Queries))
	for _, response := range batchServer.responses {
		require.Empty(t, response.Error, fmt.Sprintf("request errored for %T", req.Queries[response.Key].Query))
		require.False(t, haveResponse[response.Key], fmt.Sprintf("duplicate response for %T", req.Queries[response.Key].Query))
		haveResponse[response.Key] = true

		switch response.Key {
		case 0:
			require.IsType(t, &runtimev1.QueryBatchResponse_TableCardinalityResponse{}, response.Result)
			tr := response.GetTableCardinalityResponse()
			require.Equal(t, int64(2), tr.Cardinality)

		case 1:
			require.IsType(t, &runtimev1.QueryBatchResponse_TableColumnsResponse{}, response.Result)
			tr := response.GetTableColumnsResponse()
			require.Equal(t, 11, len(tr.ProfileColumns))

		case 2:
			require.IsType(t, &runtimev1.QueryBatchResponse_TableRowsResponse{}, response.Result)
			tr := response.GetTableRowsResponse()
			require.Equal(t, 2, len(tr.Data))
		}
	}
}

func TestServer_QueryBatch_SomeErrors(t *testing.T) {
	t.Parallel()
	server, instanceId := getMetricsTestServer(t, "ad_bids_2rows")

	// Send all types of column query requests
	req := &runtimev1.QueryBatchRequest{
		InstanceId: instanceId,
		Queries: []*runtimev1.QueryBatchEntry{
			{
				Key: 0,
				Query: &runtimev1.QueryBatchEntry_MetricsViewTotalsRequest{
					MetricsViewTotalsRequest: &runtimev1.MetricsViewTotalsRequest{
						MetricsViewName: "ad_bids_metrics",
						MeasureNames:    []string{"measure_0"},
					},
				},
			},
			{
				Key: 1,
				Query: &runtimev1.QueryBatchEntry_ColumnNullCountRequest{
					ColumnNullCountRequest: &runtimev1.ColumnNullCountRequest{
						TableName: "ad_bids",
						// Query on non-existent column (should error out)
						ColumnName: "pub",
					},
				},
			},
			{
				Key: 2,
				Query: &runtimev1.QueryBatchEntry_TableRowsRequest{
					TableRowsRequest: &runtimev1.TableRowsRequest{
						TableName: "ad_bids",
					},
				},
			},
		},
	}

	batchServer := &fakeBatchServer{
		responses: make([]*runtimev1.QueryBatchResponse, 0),
		ctx:       testCtx(),
	}
	err := server.QueryBatch(req, batchServer)
	require.NoError(t, err)
	require.Equal(t, len(req.Queries), len(batchServer.responses))

	haveResponse := make([]bool, len(req.Queries))
	for _, response := range batchServer.responses {
		require.False(t, haveResponse[response.Key], fmt.Sprintf("duplicate response for %T", req.Queries[response.Key].Query))
		haveResponse[response.Key] = true

		switch response.Key {
		case 0:
			require.IsType(t, &runtimev1.QueryBatchResponse_MetricsViewTotalsResponse{}, response.Result)
			tr := response.GetMetricsViewTotalsResponse()
			require.Equal(t, 1, len(tr.Data.Fields))

		case 1:
			require.Contains(t, response.Error, `Referenced column "pub" not found`)

		case 2:
			require.IsType(t, &runtimev1.QueryBatchResponse_TableRowsResponse{}, response.Result)
			tr := response.GetTableRowsResponse()
			require.Equal(t, 2, len(tr.Data))
		}
	}
}

type fakeBatchServer struct {
	grpc.ServerStream
	responses []*runtimev1.QueryBatchResponse
	ctx       context.Context
}

func (s *fakeBatchServer) Send(m *runtimev1.QueryBatchResponse) error {
	s.responses = append(s.responses, m)
	return nil
}

func (s *fakeBatchServer) Context() context.Context {
	return s.ctx
}
