package server

import (
	"context"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"golang.org/x/sync/errgroup"
)

func (s *Server) QueryBatch(req *runtimev1.QueryBatchRequest, srv runtimev1.QueryService_QueryBatchServer) error {
	// TODO: Performance improvements:
	//       1. Check for access upfront based on what is in the request
	//       2. Check for cache and return those queries immediately before creating a goroutine
	//       3. Use a goroutine pool with size equal to driver's concurrency to execute the queries

	g, ctx := errgroup.WithContext(srv.Context())

	for idx, qry := range req.Queries {
		idx := idx
		qry := qry
		g.Go(func() error {
			resp := s.forwardQuery(ctx, req.InstanceId, idx, qry)
			return srv.Send(resp)
		})
	}

	return g.Wait()
}

func (s *Server) forwardQuery(ctx context.Context, instID string, idx int, qry *runtimev1.Query) *runtimev1.QueryBatchResponse {
	var err error
	res := &runtimev1.QueryResult{}
	switch q := qry.Query.(type) {
	case *runtimev1.Query_MetricsViewAggregationRequest:
		var r *runtimev1.MetricsViewAggregationResponse
		q.MetricsViewAggregationRequest.InstanceId = instID
		r, err = s.MetricsViewAggregation(ctx, q.MetricsViewAggregationRequest)
		if err == nil {
			res.Result = &runtimev1.QueryResult_MetricsViewAggregationResponse{MetricsViewAggregationResponse: r}
		}

	case *runtimev1.Query_MetricsViewToplistRequest:
		var r *runtimev1.MetricsViewToplistResponse
		q.MetricsViewToplistRequest.InstanceId = instID
		r, err = s.MetricsViewToplist(ctx, q.MetricsViewToplistRequest)
		if err == nil {
			res.Result = &runtimev1.QueryResult_MetricsViewToplistResponse{MetricsViewToplistResponse: r}
		}

	case *runtimev1.Query_MetricsViewComparisonRequest:
		var r *runtimev1.MetricsViewComparisonResponse
		q.MetricsViewComparisonRequest.InstanceId = instID
		r, err = s.MetricsViewComparison(ctx, q.MetricsViewComparisonRequest)
		if err == nil {
			res.Result = &runtimev1.QueryResult_MetricsViewComparisonResponse{MetricsViewComparisonResponse: r}
		}

	case *runtimev1.Query_MetricsViewTimeSeriesRequest:
		var r *runtimev1.MetricsViewTimeSeriesResponse
		q.MetricsViewTimeSeriesRequest.InstanceId = instID
		r, err = s.MetricsViewTimeSeries(ctx, q.MetricsViewTimeSeriesRequest)
		if err == nil {
			res.Result = &runtimev1.QueryResult_MetricsViewTimeSeriesResponse{MetricsViewTimeSeriesResponse: r}
		}

	case *runtimev1.Query_MetricsViewTotalsRequest:
		var r *runtimev1.MetricsViewTotalsResponse
		q.MetricsViewTotalsRequest.InstanceId = instID
		r, err = s.MetricsViewTotals(ctx, q.MetricsViewTotalsRequest)
		if err == nil {
			res.Result = &runtimev1.QueryResult_MetricsViewTotalsResponse{MetricsViewTotalsResponse: r}
		}

	case *runtimev1.Query_MetricsViewRowsRequest:
		var r *runtimev1.MetricsViewRowsResponse
		q.MetricsViewRowsRequest.InstanceId = instID
		r, err = s.MetricsViewRows(ctx, q.MetricsViewRowsRequest)
		if err == nil {
			res.Result = &runtimev1.QueryResult_MetricsViewRowsResponse{MetricsViewRowsResponse: r}
		}

	case *runtimev1.Query_ColumnRollupIntervalRequest:
		var r *runtimev1.ColumnRollupIntervalResponse
		q.ColumnRollupIntervalRequest.InstanceId = instID
		r, err = s.ColumnRollupInterval(ctx, q.ColumnRollupIntervalRequest)
		if err == nil {
			res.Result = &runtimev1.QueryResult_ColumnRollupIntervalResponse{ColumnRollupIntervalResponse: r}
		}

	case *runtimev1.Query_ColumnTopKRequest:
		var r *runtimev1.ColumnTopKResponse
		q.ColumnTopKRequest.InstanceId = instID
		r, err = s.ColumnTopK(ctx, q.ColumnTopKRequest)
		if err == nil {
			res.Result = &runtimev1.QueryResult_ColumnTopKResponse{ColumnTopKResponse: r}
		}

	case *runtimev1.Query_ColumnNullCountRequest:
		var r *runtimev1.ColumnNullCountResponse
		q.ColumnNullCountRequest.InstanceId = instID
		r, err = s.ColumnNullCount(ctx, q.ColumnNullCountRequest)
		if err == nil {
			res.Result = &runtimev1.QueryResult_ColumnNullCountResponse{ColumnNullCountResponse: r}
		}

	case *runtimev1.Query_ColumnDescriptiveStatisticsRequest:
		var r *runtimev1.ColumnDescriptiveStatisticsResponse
		q.ColumnDescriptiveStatisticsRequest.InstanceId = instID
		r, err = s.ColumnDescriptiveStatistics(ctx, q.ColumnDescriptiveStatisticsRequest)
		if err == nil {
			res.Result = &runtimev1.QueryResult_ColumnDescriptiveStatisticsResponse{ColumnDescriptiveStatisticsResponse: r}
		}

	case *runtimev1.Query_ColumnTimeGrainRequest:
		var r *runtimev1.ColumnTimeGrainResponse
		q.ColumnTimeGrainRequest.InstanceId = instID
		r, err = s.ColumnTimeGrain(ctx, q.ColumnTimeGrainRequest)
		if err == nil {
			res.Result = &runtimev1.QueryResult_ColumnTimeGrainResponse{ColumnTimeGrainResponse: r}
		}

	case *runtimev1.Query_ColumnNumericHistogramRequest:
		var r *runtimev1.ColumnNumericHistogramResponse
		q.ColumnNumericHistogramRequest.InstanceId = instID
		r, err = s.ColumnNumericHistogram(ctx, q.ColumnNumericHistogramRequest)
		if err == nil {
			res.Result = &runtimev1.QueryResult_ColumnNumericHistogramResponse{ColumnNumericHistogramResponse: r}
		}

	case *runtimev1.Query_ColumnRugHistogramRequest:
		var r *runtimev1.ColumnRugHistogramResponse
		q.ColumnRugHistogramRequest.InstanceId = instID
		r, err = s.ColumnRugHistogram(ctx, q.ColumnRugHistogramRequest)
		if err == nil {
			res.Result = &runtimev1.QueryResult_ColumnRugHistogramResponse{ColumnRugHistogramResponse: r}
		}

	case *runtimev1.Query_ColumnTimeRangeRequest:
		var r *runtimev1.ColumnTimeRangeResponse
		q.ColumnTimeRangeRequest.InstanceId = instID
		r, err = s.ColumnTimeRange(ctx, q.ColumnTimeRangeRequest)
		if err == nil {
			res.Result = &runtimev1.QueryResult_ColumnTimeRangeResponse{ColumnTimeRangeResponse: r}
		}

	case *runtimev1.Query_ColumnCardinalityRequest:
		var r *runtimev1.ColumnCardinalityResponse
		q.ColumnCardinalityRequest.InstanceId = instID
		r, err = s.ColumnCardinality(ctx, q.ColumnCardinalityRequest)
		if err == nil {
			res.Result = &runtimev1.QueryResult_ColumnCardinalityResponse{ColumnCardinalityResponse: r}
		}

	case *runtimev1.Query_ColumnTimeSeriesRequest:
		var r *runtimev1.ColumnTimeSeriesResponse
		q.ColumnTimeSeriesRequest.InstanceId = instID
		r, err = s.ColumnTimeSeries(ctx, q.ColumnTimeSeriesRequest)
		if err == nil {
			res.Result = &runtimev1.QueryResult_ColumnTimeSeriesResponse{ColumnTimeSeriesResponse: r}
		}

	case *runtimev1.Query_TableCardinalityRequest:
		var r *runtimev1.TableCardinalityResponse
		q.TableCardinalityRequest.InstanceId = instID
		r, err = s.TableCardinality(ctx, q.TableCardinalityRequest)
		if err == nil {
			res.Result = &runtimev1.QueryResult_TableCardinalityResponse{TableCardinalityResponse: r}
		}

	case *runtimev1.Query_TableColumnsRequest:
		var r *runtimev1.TableColumnsResponse
		q.TableColumnsRequest.InstanceId = instID
		r, err = s.TableColumns(ctx, q.TableColumnsRequest)
		if err == nil {
			res.Result = &runtimev1.QueryResult_TableColumnsResponse{TableColumnsResponse: r}
		}

	case *runtimev1.Query_TableRowsRequest:
		var r *runtimev1.TableRowsResponse
		q.TableRowsRequest.InstanceId = instID
		r, err = s.TableRows(ctx, q.TableRowsRequest)
		if err == nil {
			res.Result = &runtimev1.QueryResult_TableRowsResponse{TableRowsResponse: r}
		}
	}

	if err != nil {
		return &runtimev1.QueryBatchResponse{Index: uint32(idx), Error: err.Error()}
	}

	return &runtimev1.QueryBatchResponse{Index: uint32(idx), Result: res}
}
