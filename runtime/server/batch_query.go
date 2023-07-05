package server

import (
	"context"
	"sync"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

func (s *Server) QueryBatch(req *runtimev1.QueryBatchRequest, srv runtimev1.QueryService_QueryBatchServer) error {
	ctx, cancel := context.WithCancelCause(srv.Context())
	cancelled := false

	// TODO: Performance improvements:
	//       1. Check for access upfront based on what is in the request
	//       2. Check for cache and return those queries immediately before creating a goroutine
	//       3. Use a goroutine pool with size equal to driver's concurrency to execute the queries

	var wg sync.WaitGroup

	for _, query := range req.Queries {
		wg.Add(1)
		go func(queryEntry *runtimev1.QueryBatchEntry) {
			defer wg.Done()

			resp := s.forwardQuery(ctx, req, queryEntry)

			if err := srv.Send(resp); err != nil && !cancelled {
				// if we failed to send response, cancel the context so any pending queries are cancelled
				cancel(err)
				cancelled = true
			}
		}(query)
	}

	wg.Wait()
	return nil
}

func (s *Server) forwardQuery(ctx context.Context, query *runtimev1.QueryBatchRequest, queryEntry *runtimev1.QueryBatchEntry) *runtimev1.QueryBatchResponse {
	resp := &runtimev1.QueryBatchResponse{
		Key: queryEntry.Key,
	}

	var err error
	switch typedQueryEntry := queryEntry.Query.(type) {
	case *runtimev1.QueryBatchEntry_MetricsViewToplistRequest:
		var r *runtimev1.MetricsViewToplistResponse
		if typedQueryEntry.MetricsViewToplistRequest.Priority == 0 {
			typedQueryEntry.MetricsViewToplistRequest.Priority = query.Priority
		}
		typedQueryEntry.MetricsViewToplistRequest.InstanceId = query.InstanceId
		r, err = s.MetricsViewToplist(ctx, typedQueryEntry.MetricsViewToplistRequest)
		if err == nil {
			resp.Query = &runtimev1.QueryBatchResponse_MetricsViewToplistResponse{MetricsViewToplistResponse: r}
		}

	case *runtimev1.QueryBatchEntry_MetricsViewComparisonToplistRequest:
		var r *runtimev1.MetricsViewComparisonToplistResponse
		if typedQueryEntry.MetricsViewComparisonToplistRequest.Priority == 0 {
			typedQueryEntry.MetricsViewComparisonToplistRequest.Priority = query.Priority
		}
		typedQueryEntry.MetricsViewComparisonToplistRequest.InstanceId = query.InstanceId
		r, err = s.MetricsViewComparisonToplist(ctx, typedQueryEntry.MetricsViewComparisonToplistRequest)
		if err == nil {
			resp.Query = &runtimev1.QueryBatchResponse_MetricsViewComparisonToplistResponse{MetricsViewComparisonToplistResponse: r}
		}

	case *runtimev1.QueryBatchEntry_MetricsViewTimeSeriesRequest:
		var r *runtimev1.MetricsViewTimeSeriesResponse
		if typedQueryEntry.MetricsViewTimeSeriesRequest.Priority == 0 {
			typedQueryEntry.MetricsViewTimeSeriesRequest.Priority = query.Priority
		}
		typedQueryEntry.MetricsViewTimeSeriesRequest.InstanceId = query.InstanceId
		r, err = s.MetricsViewTimeSeries(ctx, typedQueryEntry.MetricsViewTimeSeriesRequest)
		if err == nil {
			resp.Query = &runtimev1.QueryBatchResponse_MetricsViewTimeSeriesResponse{MetricsViewTimeSeriesResponse: r}
		}

	case *runtimev1.QueryBatchEntry_MetricsViewTotalsRequest:
		var r *runtimev1.MetricsViewTotalsResponse
		if typedQueryEntry.MetricsViewTotalsRequest.Priority == 0 {
			typedQueryEntry.MetricsViewTotalsRequest.Priority = query.Priority
		}
		typedQueryEntry.MetricsViewTotalsRequest.InstanceId = query.InstanceId
		r, err = s.MetricsViewTotals(ctx, typedQueryEntry.MetricsViewTotalsRequest)
		if err == nil {
			resp.Query = &runtimev1.QueryBatchResponse_MetricsViewTotalsResponse{MetricsViewTotalsResponse: r}
		}

	case *runtimev1.QueryBatchEntry_MetricsViewRowsRequest:
		var r *runtimev1.MetricsViewRowsResponse
		if typedQueryEntry.MetricsViewRowsRequest.Priority == 0 {
			typedQueryEntry.MetricsViewRowsRequest.Priority = query.Priority
		}
		typedQueryEntry.MetricsViewRowsRequest.InstanceId = query.InstanceId
		r, err = s.MetricsViewRows(ctx, typedQueryEntry.MetricsViewRowsRequest)
		if err == nil {
			resp.Query = &runtimev1.QueryBatchResponse_MetricsViewRowsResponse{MetricsViewRowsResponse: r}
		}

	case *runtimev1.QueryBatchEntry_ColumnRollupIntervalRequest:
		var r *runtimev1.ColumnRollupIntervalResponse
		if typedQueryEntry.ColumnRollupIntervalRequest.Priority == 0 {
			typedQueryEntry.ColumnRollupIntervalRequest.Priority = query.Priority
		}
		typedQueryEntry.ColumnRollupIntervalRequest.InstanceId = query.InstanceId
		r, err = s.ColumnRollupInterval(ctx, typedQueryEntry.ColumnRollupIntervalRequest)
		if err == nil {
			resp.Query = &runtimev1.QueryBatchResponse_ColumnRollupIntervalResponse{ColumnRollupIntervalResponse: r}
		}

	case *runtimev1.QueryBatchEntry_ColumnTopKRequest:
		var r *runtimev1.ColumnTopKResponse
		if typedQueryEntry.ColumnTopKRequest.Priority == 0 {
			typedQueryEntry.ColumnTopKRequest.Priority = query.Priority
		}
		typedQueryEntry.ColumnTopKRequest.InstanceId = query.InstanceId
		r, err = s.ColumnTopK(ctx, typedQueryEntry.ColumnTopKRequest)
		if err == nil {
			resp.Query = &runtimev1.QueryBatchResponse_ColumnTopKResponse{ColumnTopKResponse: r}
		}

	case *runtimev1.QueryBatchEntry_ColumnNullCountRequest:
		var r *runtimev1.ColumnNullCountResponse
		if typedQueryEntry.ColumnNullCountRequest.Priority == 0 {
			typedQueryEntry.ColumnNullCountRequest.Priority = query.Priority
		}
		typedQueryEntry.ColumnNullCountRequest.InstanceId = query.InstanceId
		r, err = s.ColumnNullCount(ctx, typedQueryEntry.ColumnNullCountRequest)
		if err == nil {
			resp.Query = &runtimev1.QueryBatchResponse_ColumnNullCountResponse{ColumnNullCountResponse: r}
		}

	case *runtimev1.QueryBatchEntry_ColumnDescriptiveStatisticsRequest:
		var r *runtimev1.ColumnDescriptiveStatisticsResponse
		if typedQueryEntry.ColumnDescriptiveStatisticsRequest.Priority == 0 {
			typedQueryEntry.ColumnDescriptiveStatisticsRequest.Priority = query.Priority
		}
		typedQueryEntry.ColumnDescriptiveStatisticsRequest.InstanceId = query.InstanceId
		r, err = s.ColumnDescriptiveStatistics(ctx, typedQueryEntry.ColumnDescriptiveStatisticsRequest)
		if err == nil {
			resp.Query = &runtimev1.QueryBatchResponse_ColumnDescriptiveStatisticsResponse{ColumnDescriptiveStatisticsResponse: r}
		}

	case *runtimev1.QueryBatchEntry_ColumnTimeGrainRequest:
		var r *runtimev1.ColumnTimeGrainResponse
		if typedQueryEntry.ColumnTimeGrainRequest.Priority == 0 {
			typedQueryEntry.ColumnTimeGrainRequest.Priority = query.Priority
		}
		typedQueryEntry.ColumnTimeGrainRequest.InstanceId = query.InstanceId
		r, err = s.ColumnTimeGrain(ctx, typedQueryEntry.ColumnTimeGrainRequest)
		if err == nil {
			resp.Query = &runtimev1.QueryBatchResponse_ColumnTimeGrainResponse{ColumnTimeGrainResponse: r}
		}

	case *runtimev1.QueryBatchEntry_ColumnNumericHistogramRequest:
		var r *runtimev1.ColumnNumericHistogramResponse
		if typedQueryEntry.ColumnNumericHistogramRequest.Priority == 0 {
			typedQueryEntry.ColumnNumericHistogramRequest.Priority = query.Priority
		}
		typedQueryEntry.ColumnNumericHistogramRequest.InstanceId = query.InstanceId
		r, err = s.ColumnNumericHistogram(ctx, typedQueryEntry.ColumnNumericHistogramRequest)
		if err == nil {
			resp.Query = &runtimev1.QueryBatchResponse_ColumnNumericHistogramResponse{ColumnNumericHistogramResponse: r}
		}

	case *runtimev1.QueryBatchEntry_ColumnRugHistogramRequest:
		var r *runtimev1.ColumnRugHistogramResponse
		if typedQueryEntry.ColumnRugHistogramRequest.Priority == 0 {
			typedQueryEntry.ColumnRugHistogramRequest.Priority = query.Priority
		}
		typedQueryEntry.ColumnRugHistogramRequest.InstanceId = query.InstanceId
		r, err = s.ColumnRugHistogram(ctx, typedQueryEntry.ColumnRugHistogramRequest)
		if err == nil {
			resp.Query = &runtimev1.QueryBatchResponse_ColumnRugHistogramResponse{ColumnRugHistogramResponse: r}
		}

	case *runtimev1.QueryBatchEntry_ColumnTimeRangeRequest:
		var r *runtimev1.ColumnTimeRangeResponse
		if typedQueryEntry.ColumnTimeRangeRequest.Priority == 0 {
			typedQueryEntry.ColumnTimeRangeRequest.Priority = query.Priority
		}
		typedQueryEntry.ColumnTimeRangeRequest.InstanceId = query.InstanceId
		r, err = s.ColumnTimeRange(ctx, typedQueryEntry.ColumnTimeRangeRequest)
		if err == nil {
			resp.Query = &runtimev1.QueryBatchResponse_ColumnTimeRangeResponse{ColumnTimeRangeResponse: r}
		}

	case *runtimev1.QueryBatchEntry_ColumnCardinalityRequest:
		var r *runtimev1.ColumnCardinalityResponse
		if typedQueryEntry.ColumnCardinalityRequest.Priority == 0 {
			typedQueryEntry.ColumnCardinalityRequest.Priority = query.Priority
		}
		typedQueryEntry.ColumnCardinalityRequest.InstanceId = query.InstanceId
		r, err = s.ColumnCardinality(ctx, typedQueryEntry.ColumnCardinalityRequest)
		if err == nil {
			resp.Query = &runtimev1.QueryBatchResponse_ColumnCardinalityResponse{ColumnCardinalityResponse: r}
		}

	case *runtimev1.QueryBatchEntry_ColumnTimeSeriesRequest:
		var r *runtimev1.ColumnTimeSeriesResponse
		if typedQueryEntry.ColumnTimeSeriesRequest.Priority == 0 {
			typedQueryEntry.ColumnTimeSeriesRequest.Priority = query.Priority
		}
		typedQueryEntry.ColumnTimeSeriesRequest.InstanceId = query.InstanceId
		r, err = s.ColumnTimeSeries(ctx, typedQueryEntry.ColumnTimeSeriesRequest)
		if err == nil {
			resp.Query = &runtimev1.QueryBatchResponse_ColumnTimeSeriesResponse{ColumnTimeSeriesResponse: r}
		}

	case *runtimev1.QueryBatchEntry_TableCardinalityRequest:
		var r *runtimev1.TableCardinalityResponse
		if typedQueryEntry.TableCardinalityRequest.Priority == 0 {
			typedQueryEntry.TableCardinalityRequest.Priority = query.Priority
		}
		typedQueryEntry.TableCardinalityRequest.InstanceId = query.InstanceId
		r, err = s.TableCardinality(ctx, typedQueryEntry.TableCardinalityRequest)
		if err == nil {
			resp.Query = &runtimev1.QueryBatchResponse_TableCardinalityResponse{TableCardinalityResponse: r}
		}

	case *runtimev1.QueryBatchEntry_TableColumnsRequest:
		var r *runtimev1.TableColumnsResponse
		if typedQueryEntry.TableColumnsRequest.Priority == 0 {
			typedQueryEntry.TableColumnsRequest.Priority = query.Priority
		}
		typedQueryEntry.TableColumnsRequest.InstanceId = query.InstanceId
		r, err = s.TableColumns(ctx, typedQueryEntry.TableColumnsRequest)
		if err == nil {
			resp.Query = &runtimev1.QueryBatchResponse_TableColumnsResponse{TableColumnsResponse: r}
		}

	case *runtimev1.QueryBatchEntry_TableRowsRequest:
		var r *runtimev1.TableRowsResponse
		if typedQueryEntry.TableRowsRequest.Priority == 0 {
			typedQueryEntry.TableRowsRequest.Priority = query.Priority
		}
		typedQueryEntry.TableRowsRequest.InstanceId = query.InstanceId
		r, err = s.TableRows(ctx, typedQueryEntry.TableRowsRequest)
		if err == nil {
			resp.Query = &runtimev1.QueryBatchResponse_TableRowsResponse{TableRowsResponse: r}
		}
	}

	if err != nil {
		resp.Error = err.Error()
	}

	return resp
}
