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

	for _, query := range req.Queries {
		query := query // create a closed var for goroutine
		g.Go(func() error {
			resp := s.forwardQuery(ctx, req, query)
			return srv.Send(resp)
		})
	}

	return g.Wait()
}

func (s *Server) forwardQuery(ctx context.Context, query *runtimev1.QueryBatchRequest, queryEntry *runtimev1.QueryBatchEntry) *runtimev1.QueryBatchResponse {
	resp := &runtimev1.QueryBatchResponse{
		Key: queryEntry.Key,
	}

	var err error
	switch q := queryEntry.Query.(type) {
	case *runtimev1.QueryBatchEntry_MetricsViewToplistRequest:
		var r *runtimev1.MetricsViewToplistResponse
		q.MetricsViewToplistRequest.InstanceId = query.InstanceId
		r, err = s.MetricsViewToplist(ctx, q.MetricsViewToplistRequest)
		if err == nil {
			resp.Result = &runtimev1.QueryBatchResponse_MetricsViewToplistResponse{MetricsViewToplistResponse: r}
		}

	case *runtimev1.QueryBatchEntry_MetricsViewComparisonToplistRequest:
		var r *runtimev1.MetricsViewComparisonToplistResponse
		q.MetricsViewComparisonToplistRequest.InstanceId = query.InstanceId
		r, err = s.MetricsViewComparisonToplist(ctx, q.MetricsViewComparisonToplistRequest)
		if err == nil {
			resp.Result = &runtimev1.QueryBatchResponse_MetricsViewComparisonToplistResponse{MetricsViewComparisonToplistResponse: r}
		}

	case *runtimev1.QueryBatchEntry_MetricsViewTimeSeriesRequest:
		var r *runtimev1.MetricsViewTimeSeriesResponse
		q.MetricsViewTimeSeriesRequest.InstanceId = query.InstanceId
		r, err = s.MetricsViewTimeSeries(ctx, q.MetricsViewTimeSeriesRequest)
		if err == nil {
			resp.Result = &runtimev1.QueryBatchResponse_MetricsViewTimeSeriesResponse{MetricsViewTimeSeriesResponse: r}
		}

	case *runtimev1.QueryBatchEntry_MetricsViewTotalsRequest:
		var r *runtimev1.MetricsViewTotalsResponse
		q.MetricsViewTotalsRequest.InstanceId = query.InstanceId
		r, err = s.MetricsViewTotals(ctx, q.MetricsViewTotalsRequest)
		if err == nil {
			resp.Result = &runtimev1.QueryBatchResponse_MetricsViewTotalsResponse{MetricsViewTotalsResponse: r}
		}

	case *runtimev1.QueryBatchEntry_MetricsViewRowsRequest:
		var r *runtimev1.MetricsViewRowsResponse
		q.MetricsViewRowsRequest.InstanceId = query.InstanceId
		r, err = s.MetricsViewRows(ctx, q.MetricsViewRowsRequest)
		if err == nil {
			resp.Result = &runtimev1.QueryBatchResponse_MetricsViewRowsResponse{MetricsViewRowsResponse: r}
		}

	case *runtimev1.QueryBatchEntry_ColumnRollupIntervalRequest:
		var r *runtimev1.ColumnRollupIntervalResponse
		q.ColumnRollupIntervalRequest.InstanceId = query.InstanceId
		r, err = s.ColumnRollupInterval(ctx, q.ColumnRollupIntervalRequest)
		if err == nil {
			resp.Result = &runtimev1.QueryBatchResponse_ColumnRollupIntervalResponse{ColumnRollupIntervalResponse: r}
		}

	case *runtimev1.QueryBatchEntry_ColumnTopKRequest:
		var r *runtimev1.ColumnTopKResponse
		q.ColumnTopKRequest.InstanceId = query.InstanceId
		r, err = s.ColumnTopK(ctx, q.ColumnTopKRequest)
		if err == nil {
			resp.Result = &runtimev1.QueryBatchResponse_ColumnTopKResponse{ColumnTopKResponse: r}
		}

	case *runtimev1.QueryBatchEntry_ColumnNullCountRequest:
		var r *runtimev1.ColumnNullCountResponse
		q.ColumnNullCountRequest.InstanceId = query.InstanceId
		r, err = s.ColumnNullCount(ctx, q.ColumnNullCountRequest)
		if err == nil {
			resp.Result = &runtimev1.QueryBatchResponse_ColumnNullCountResponse{ColumnNullCountResponse: r}
		}

	case *runtimev1.QueryBatchEntry_ColumnDescriptiveStatisticsRequest:
		var r *runtimev1.ColumnDescriptiveStatisticsResponse
		q.ColumnDescriptiveStatisticsRequest.InstanceId = query.InstanceId
		r, err = s.ColumnDescriptiveStatistics(ctx, q.ColumnDescriptiveStatisticsRequest)
		if err == nil {
			resp.Result = &runtimev1.QueryBatchResponse_ColumnDescriptiveStatisticsResponse{ColumnDescriptiveStatisticsResponse: r}
		}

	case *runtimev1.QueryBatchEntry_ColumnTimeGrainRequest:
		var r *runtimev1.ColumnTimeGrainResponse
		q.ColumnTimeGrainRequest.InstanceId = query.InstanceId
		r, err = s.ColumnTimeGrain(ctx, q.ColumnTimeGrainRequest)
		if err == nil {
			resp.Result = &runtimev1.QueryBatchResponse_ColumnTimeGrainResponse{ColumnTimeGrainResponse: r}
		}

	case *runtimev1.QueryBatchEntry_ColumnNumericHistogramRequest:
		var r *runtimev1.ColumnNumericHistogramResponse
		q.ColumnNumericHistogramRequest.InstanceId = query.InstanceId
		r, err = s.ColumnNumericHistogram(ctx, q.ColumnNumericHistogramRequest)
		if err == nil {
			resp.Result = &runtimev1.QueryBatchResponse_ColumnNumericHistogramResponse{ColumnNumericHistogramResponse: r}
		}

	case *runtimev1.QueryBatchEntry_ColumnRugHistogramRequest:
		var r *runtimev1.ColumnRugHistogramResponse
		q.ColumnRugHistogramRequest.InstanceId = query.InstanceId
		r, err = s.ColumnRugHistogram(ctx, q.ColumnRugHistogramRequest)
		if err == nil {
			resp.Result = &runtimev1.QueryBatchResponse_ColumnRugHistogramResponse{ColumnRugHistogramResponse: r}
		}

	case *runtimev1.QueryBatchEntry_ColumnTimeRangeRequest:
		var r *runtimev1.ColumnTimeRangeResponse
		q.ColumnTimeRangeRequest.InstanceId = query.InstanceId
		r, err = s.ColumnTimeRange(ctx, q.ColumnTimeRangeRequest)
		if err == nil {
			resp.Result = &runtimev1.QueryBatchResponse_ColumnTimeRangeResponse{ColumnTimeRangeResponse: r}
		}

	case *runtimev1.QueryBatchEntry_ColumnCardinalityRequest:
		var r *runtimev1.ColumnCardinalityResponse
		q.ColumnCardinalityRequest.InstanceId = query.InstanceId
		r, err = s.ColumnCardinality(ctx, q.ColumnCardinalityRequest)
		if err == nil {
			resp.Result = &runtimev1.QueryBatchResponse_ColumnCardinalityResponse{ColumnCardinalityResponse: r}
		}

	case *runtimev1.QueryBatchEntry_ColumnTimeSeriesRequest:
		var r *runtimev1.ColumnTimeSeriesResponse
		q.ColumnTimeSeriesRequest.InstanceId = query.InstanceId
		r, err = s.ColumnTimeSeries(ctx, q.ColumnTimeSeriesRequest)
		if err == nil {
			resp.Result = &runtimev1.QueryBatchResponse_ColumnTimeSeriesResponse{ColumnTimeSeriesResponse: r}
		}

	case *runtimev1.QueryBatchEntry_TableCardinalityRequest:
		var r *runtimev1.TableCardinalityResponse
		q.TableCardinalityRequest.InstanceId = query.InstanceId
		r, err = s.TableCardinality(ctx, q.TableCardinalityRequest)
		if err == nil {
			resp.Result = &runtimev1.QueryBatchResponse_TableCardinalityResponse{TableCardinalityResponse: r}
		}

	case *runtimev1.QueryBatchEntry_TableColumnsRequest:
		var r *runtimev1.TableColumnsResponse
		q.TableColumnsRequest.InstanceId = query.InstanceId
		r, err = s.TableColumns(ctx, q.TableColumnsRequest)
		if err == nil {
			resp.Result = &runtimev1.QueryBatchResponse_TableColumnsResponse{TableColumnsResponse: r}
		}

	case *runtimev1.QueryBatchEntry_TableRowsRequest:
		var r *runtimev1.TableRowsResponse
		q.TableRowsRequest.InstanceId = query.InstanceId
		r, err = s.TableRows(ctx, q.TableRowsRequest)
		if err == nil {
			resp.Result = &runtimev1.QueryBatchResponse_TableRowsResponse{TableRowsResponse: r}
		}
	}

	if err != nil {
		resp.Error = err.Error()
	}

	return resp
}
