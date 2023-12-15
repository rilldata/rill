package server_test

import (
	"context"
	"fmt"
	"sync"
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
		Queries: []*runtimev1.Query{
			{
				Query: &runtimev1.Query_MetricsViewToplistRequest{
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
				Query: &runtimev1.Query_MetricsViewComparisonRequest{
					MetricsViewComparisonRequest: &runtimev1.MetricsViewComparisonRequest{
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
							Start: parseTimeToProtoTimeStamps(t, "2022-01-01T00:00:00Z"),
							End:   parseTimeToProtoTimeStamps(t, "2022-01-01T23:59:00Z"),
						},
						ComparisonTimeRange: &runtimev1.TimeRange{
							Start: parseTimeToProtoTimeStamps(t, "2022-01-02T00:00:00Z"),
							End:   parseTimeToProtoTimeStamps(t, "2022-01-02T23:59:00Z"),
						},
						Sort: []*runtimev1.MetricsViewComparisonSort{
							{
								Name:     "measure_2",
								SortType: runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_BASE_VALUE,
								Desc:     false,
							},
						},
					},
				},
			},
			{
				Query: &runtimev1.Query_MetricsViewTimeSeriesRequest{
					MetricsViewTimeSeriesRequest: &runtimev1.MetricsViewTimeSeriesRequest{
						MetricsViewName: "ad_bids_metrics",
						TimeGranularity: runtimev1.TimeGrain_TIME_GRAIN_DAY,
						MeasureNames:    []string{"measure_0", "measure_2"},
					},
				},
			},
			{
				Query: &runtimev1.Query_MetricsViewTotalsRequest{
					MetricsViewTotalsRequest: &runtimev1.MetricsViewTotalsRequest{
						MetricsViewName: "ad_bids_metrics",
						MeasureNames:    []string{"measure_0"},
					},
				},
			},
			{
				Query: &runtimev1.Query_MetricsViewRowsRequest{
					MetricsViewRowsRequest: &runtimev1.MetricsViewRowsRequest{
						MetricsViewName: "ad_bids_metrics",
					},
				},
			},
		},
	}

	batchServer := newFakeBatchServer()
	err := server.QueryBatch(req, batchServer)
	require.NoError(t, err)
	require.Equal(t, len(req.Queries), len(batchServer.responses))

	haveResponse := make([]bool, len(req.Queries))
	for _, response := range batchServer.responses {
		require.Empty(t, response.Error, fmt.Sprintf("request errored for %T", req.Queries[response.Index].Query))
		require.False(t, haveResponse[response.Index], fmt.Sprintf("duplicate response for %T", req.Queries[response.Index].Query))
		haveResponse[response.Index] = true

		switch response.Index {
		case 0:
			require.IsType(t, &runtimev1.QueryResult_MetricsViewToplistResponse{}, response.Result.Result)
			tr := response.Result.GetMetricsViewToplistResponse()
			require.Equal(t, 2, len(tr.Data))
			require.Equal(t, 2, len(tr.Data[0].Fields))
			require.Equal(t, 2, len(tr.Data[1].Fields))

		case 1:
			require.IsType(t, &runtimev1.QueryResult_MetricsViewComparisonResponse{}, response.Result.Result)
			tr := response.Result.GetMetricsViewComparisonResponse()
			rows := tr.Rows
			require.NoError(t, err)
			require.Equal(t, 1, len(rows))

		case 2:
			require.IsType(t, &runtimev1.QueryResult_MetricsViewTimeSeriesResponse{}, response.Result.Result)
			tr := response.Result.GetMetricsViewTimeSeriesResponse()
			require.Equal(t, 2, len(tr.Data))
			require.Equal(t, 2, len(tr.Data[0].Records.Fields))

		case 3:
			require.IsType(t, &runtimev1.QueryResult_MetricsViewTotalsResponse{}, response.Result.Result)
			tr := response.Result.GetMetricsViewTotalsResponse()
			require.Equal(t, 1, len(tr.Data.Fields))

		case 4:
			require.IsType(t, &runtimev1.QueryResult_MetricsViewRowsResponse{}, response.Result.Result)
			tr := response.Result.GetMetricsViewRowsResponse()
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
		Queries: []*runtimev1.Query{
			{
				Query: &runtimev1.Query_ColumnRollupIntervalRequest{
					ColumnRollupIntervalRequest: &runtimev1.ColumnRollupIntervalRequest{
						TableName:  "ad_bids",
						ColumnName: "timestamp",
					},
				},
			},
			{
				Query: &runtimev1.Query_ColumnTopKRequest{
					ColumnTopKRequest: &runtimev1.ColumnTopKRequest{
						TableName:  "ad_bids",
						ColumnName: "publisher",
						K:          1,
					},
				},
			},
			{
				Query: &runtimev1.Query_ColumnNullCountRequest{
					ColumnNullCountRequest: &runtimev1.ColumnNullCountRequest{
						TableName:  "ad_bids",
						ColumnName: "publisher",
					},
				},
			},
			{
				Query: &runtimev1.Query_ColumnDescriptiveStatisticsRequest{
					ColumnDescriptiveStatisticsRequest: &runtimev1.ColumnDescriptiveStatisticsRequest{
						TableName:  "ad_bids",
						ColumnName: "impressions",
					},
				},
			},
			{
				Query: &runtimev1.Query_ColumnTimeGrainRequest{
					ColumnTimeGrainRequest: &runtimev1.ColumnTimeGrainRequest{
						TableName:  "ad_bids",
						ColumnName: "timestamp",
					},
				},
			},
			{
				Query: &runtimev1.Query_ColumnNumericHistogramRequest{
					ColumnNumericHistogramRequest: &runtimev1.ColumnNumericHistogramRequest{
						TableName:       "ad_bids",
						ColumnName:      "impressions",
						HistogramMethod: runtimev1.HistogramMethod_HISTOGRAM_METHOD_DIAGNOSTIC,
					},
				},
			},
			{
				Query: &runtimev1.Query_ColumnRugHistogramRequest{
					ColumnRugHistogramRequest: &runtimev1.ColumnRugHistogramRequest{
						TableName:  "ad_bids",
						ColumnName: "impressions",
					},
				},
			},
			{
				Query: &runtimev1.Query_ColumnTimeRangeRequest{
					ColumnTimeRangeRequest: &runtimev1.ColumnTimeRangeRequest{
						TableName:  "ad_bids",
						ColumnName: "timestamp",
					},
				},
			},
			{
				Query: &runtimev1.Query_ColumnCardinalityRequest{
					ColumnCardinalityRequest: &runtimev1.ColumnCardinalityRequest{
						TableName:  "ad_bids",
						ColumnName: "domain",
					},
				},
			},
			{
				Query: &runtimev1.Query_ColumnTimeSeriesRequest{
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

	batchServer := newFakeBatchServer()
	err := server.QueryBatch(req, batchServer)
	require.NoError(t, err)
	require.Equal(t, len(req.Queries), len(batchServer.responses))

	haveResponse := make([]bool, len(req.Queries))
	for _, response := range batchServer.responses {
		require.Empty(t, response.Error, fmt.Sprintf("request errored for %T", req.Queries[response.Index].Query))
		require.False(t, haveResponse[response.Index], fmt.Sprintf("duplicate response for %T", req.Queries[response.Index].Query))
		haveResponse[response.Index] = true

		switch response.Index {
		case 0:
			require.IsType(t, &runtimev1.QueryResult_ColumnRollupIntervalResponse{}, response.Result.Result)
			tr := response.Result.GetColumnRollupIntervalResponse()
			require.Equal(t, 1, tr.Start.AsTime().Day())
			require.Equal(t, 2, tr.End.AsTime().Day())

		case 1:
			require.IsType(t, &runtimev1.QueryResult_ColumnTopKResponse{}, response.Result.Result)
			tr := response.Result.GetColumnTopKResponse()
			require.Equal(t, 1, len(tr.CategoricalSummary.GetTopK().Entries))

		case 2:
			require.IsType(t, &runtimev1.QueryResult_ColumnNullCountResponse{}, response.Result.Result)
			tr := response.Result.GetColumnNullCountResponse()
			require.Equal(t, 1.0, tr.Count)

		case 3:
			require.IsType(t, &runtimev1.QueryResult_ColumnDescriptiveStatisticsResponse{}, response.Result.Result)
			tr := response.Result.GetColumnDescriptiveStatisticsResponse()
			require.Equal(t, 1.0, tr.NumericSummary.GetNumericStatistics().Min)
			require.Equal(t, 2.0, tr.NumericSummary.GetNumericStatistics().Max)

		case 4:
			require.IsType(t, &runtimev1.QueryResult_ColumnTimeGrainResponse{}, response.Result.Result)
			tr := response.Result.GetColumnTimeGrainResponse()
			require.Equal(t, runtimev1.TimeGrain_TIME_GRAIN_MILLISECOND, tr.TimeGrain)

		case 5:
			require.IsType(t, &runtimev1.QueryResult_ColumnNumericHistogramResponse{}, response.Result.Result)
			tr := response.Result.GetColumnNumericHistogramResponse()
			require.Equal(t, 2, len(tr.NumericSummary.GetNumericHistogramBins().Bins))

		case 6:
			require.IsType(t, &runtimev1.QueryResult_ColumnRugHistogramResponse{}, response.Result.Result)
			tr := response.Result.GetColumnRugHistogramResponse()
			require.Equal(t, 2, len(tr.NumericSummary.GetNumericOutliers().Outliers))

		case 7:
			require.IsType(t, &runtimev1.QueryResult_ColumnTimeRangeResponse{}, response.Result.Result)
			tr := response.Result.GetColumnTimeRangeResponse()
			require.Equal(t, 1, tr.TimeRangeSummary.Min.AsTime().Day())
			require.Equal(t, 2, tr.TimeRangeSummary.Max.AsTime().Day())

		case 8:
			require.IsType(t, &runtimev1.QueryResult_ColumnCardinalityResponse{}, response.Result.Result)
			tr := response.Result.GetColumnCardinalityResponse()
			require.Equal(t, 2.0, tr.CategoricalSummary.GetCardinality())

		case 9:
			require.IsType(t, &runtimev1.QueryResult_ColumnTimeSeriesResponse{}, response.Result.Result)
			tr := response.Result.GetColumnTimeSeriesResponse()
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
		Queries: []*runtimev1.Query{
			{
				Query: &runtimev1.Query_TableCardinalityRequest{
					TableCardinalityRequest: &runtimev1.TableCardinalityRequest{
						TableName: "ad_bids",
					},
				},
			},
			{
				Query: &runtimev1.Query_TableColumnsRequest{
					TableColumnsRequest: &runtimev1.TableColumnsRequest{
						TableName: "ad_bids",
					},
				},
			},
			{
				Query: &runtimev1.Query_TableRowsRequest{
					TableRowsRequest: &runtimev1.TableRowsRequest{
						TableName: "ad_bids",
					},
				},
			},
		},
	}

	batchServer := newFakeBatchServer()
	err := server.QueryBatch(req, batchServer)
	require.NoError(t, err)
	require.Equal(t, len(req.Queries), len(batchServer.responses))

	haveResponse := make([]bool, len(req.Queries))
	for _, response := range batchServer.responses {
		require.Empty(t, response.Error, fmt.Sprintf("request errored for %T", req.Queries[response.Index].Query))
		require.False(t, haveResponse[response.Index], fmt.Sprintf("duplicate response for %T", req.Queries[response.Index].Query))
		haveResponse[response.Index] = true

		switch response.Index {
		case 0:
			require.IsType(t, &runtimev1.QueryResult_TableCardinalityResponse{}, response.Result.Result)
			tr := response.Result.GetTableCardinalityResponse()
			require.Equal(t, int64(2), tr.Cardinality)

		case 1:
			require.IsType(t, &runtimev1.QueryResult_TableColumnsResponse{}, response.Result.Result)
			tr := response.Result.GetTableColumnsResponse()
			require.Equal(t, 11, len(tr.ProfileColumns))

		case 2:
			require.IsType(t, &runtimev1.QueryResult_TableRowsResponse{}, response.Result.Result)
			tr := response.Result.GetTableRowsResponse()
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
		Queries: []*runtimev1.Query{
			{
				Query: &runtimev1.Query_MetricsViewTotalsRequest{
					MetricsViewTotalsRequest: &runtimev1.MetricsViewTotalsRequest{
						MetricsViewName: "ad_bids_metrics",
						MeasureNames:    []string{"measure_0"},
					},
				},
			},
			{
				Query: &runtimev1.Query_ColumnNullCountRequest{
					ColumnNullCountRequest: &runtimev1.ColumnNullCountRequest{
						TableName: "ad_bids",
						// Query on non-existent column (should error out)
						ColumnName: "pub",
					},
				},
			},
			{
				Query: &runtimev1.Query_TableRowsRequest{
					TableRowsRequest: &runtimev1.TableRowsRequest{
						TableName: "ad_bids",
					},
				},
			},
		},
	}

	batchServer := newFakeBatchServer()
	err := server.QueryBatch(req, batchServer)
	require.NoError(t, err)
	require.Equal(t, len(req.Queries), len(batchServer.responses))

	haveResponse := make([]bool, len(req.Queries))
	for _, response := range batchServer.responses {
		require.False(t, haveResponse[response.Index], fmt.Sprintf("duplicate response for %T", req.Queries[response.Index].Query))
		haveResponse[response.Index] = true

		switch response.Index {
		case 0:
			require.IsType(t, &runtimev1.QueryResult_MetricsViewTotalsResponse{}, response.Result.Result)
			tr := response.Result.GetMetricsViewTotalsResponse()
			require.Equal(t, 1, len(tr.Data.Fields))

		case 1:
			require.Contains(t, response.Error, `Referenced column "pub" not found`)

		case 2:
			require.IsType(t, &runtimev1.QueryResult_TableRowsResponse{}, response.Result.Result)
			tr := response.Result.GetTableRowsResponse()
			require.Equal(t, 2, len(tr.Data))
		}
	}
}

type fakeBatchServer struct {
	grpc.ServerStream
	responses []*runtimev1.QueryBatchResponse
	ctx       context.Context
	lock      sync.Mutex
}

func newFakeBatchServer() *fakeBatchServer {
	return &fakeBatchServer{
		responses: make([]*runtimev1.QueryBatchResponse, 0),
		ctx:       testCtx(),
		lock:      sync.Mutex{},
	}
}

func (s *fakeBatchServer) Send(m *runtimev1.QueryBatchResponse) error {
	// needed since individual batches are run in parallel
	s.lock.Lock()
	defer s.lock.Unlock()
	s.responses = append(s.responses, m)
	return nil
}

func (s *fakeBatchServer) Context() context.Context {
	return s.ctx
}
