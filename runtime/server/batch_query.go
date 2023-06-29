package server

import (
	"context"
	"fmt"
	"sync"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/server/auth"
)

func (s *Server) QueryBatch(req *runtimev1.QueryBatchRequest, srv runtimev1.QueryService_QueryBatchServer) error {
	ctx := srv.Context()

	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.ReadOLAP) {
		return ErrForbidden
	}
	var wg sync.WaitGroup

	for _, query := range req.Queries {
		wg.Add(1)
		go func(query *runtimev1.QueryBatchSingleRequest) {
			defer wg.Done()

			resp := s.forwardQuery(ctx, query)

			if err := srv.Send(resp); err != nil {
				s.logger.Debug(fmt.Sprintf("Profiling Query Error: %v", err))
			}
		}(query)
	}

	wg.Wait()
	return nil
}

func (s *Server) forwardQuery(ctx context.Context, query *runtimev1.QueryBatchSingleRequest) *runtimev1.QueryBatchResponse {
	resp := &runtimev1.QueryBatchResponse{
		Id:   query.Id,
		Type: query.Type,
	}

	var err error
	switch query.Type {
	case runtimev1.QueryBatchType_QUERY_BATCH_TYPE_METRICS_VIEW_TOPLIST:
		var r *runtimev1.MetricsViewToplistResponse
		r, err = s.MetricsViewToplist(ctx, query.GetMetricsViewToplistRequest())
		if err == nil {
			resp.Query = &runtimev1.QueryBatchResponse_MetricsViewToplistResponse{MetricsViewToplistResponse: r}
		}

	case runtimev1.QueryBatchType_QUERY_BATCH_TYPE_METRICS_VIEW_COMPARISON_TOPLIST:
		var r *runtimev1.MetricsViewComparisonToplistResponse
		r, err = s.MetricsViewComparisonToplist(ctx, query.GetMetricsViewComparisonToplistRequest())
		if err == nil {
			resp.Query = &runtimev1.QueryBatchResponse_MetricsViewComparisonToplistResponse{MetricsViewComparisonToplistResponse: r}
		}

	case runtimev1.QueryBatchType_QUERY_BATCH_TYPE_METRICS_VIEW_TIMESERIES:
		var r *runtimev1.MetricsViewTimeSeriesResponse
		r, err = s.MetricsViewTimeSeries(ctx, query.GetMetricsViewTimeSeriesRequest())
		if err == nil {
			resp.Query = &runtimev1.QueryBatchResponse_MetricsViewTimeSeriesResponse{MetricsViewTimeSeriesResponse: r}
		}

	case runtimev1.QueryBatchType_QUERY_BATCH_TYPE_METRICS_VIEW_TOTALS:
		var r *runtimev1.MetricsViewTotalsResponse
		r, err = s.MetricsViewTotals(ctx, query.GetMetricsViewTotalsRequest())
		if err == nil {
			resp.Query = &runtimev1.QueryBatchResponse_MetricsViewTotalsResponse{MetricsViewTotalsResponse: r}
		}

	case runtimev1.QueryBatchType_QUERY_BATCH_TYPE_METRICS_VIEW_ROWS:
		var r *runtimev1.MetricsViewRowsResponse
		r, err = s.MetricsViewRows(ctx, query.GetMetricsViewRowsRequest())
		if err == nil {
			resp.Query = &runtimev1.QueryBatchResponse_MetricsViewRowsResponse{MetricsViewRowsResponse: r}
		}

	case runtimev1.QueryBatchType_QUERY_BATCH_TYPE_COLUMN_ROLLUP_INTERVAL:
		var r *runtimev1.ColumnRollupIntervalResponse
		r, err = s.ColumnRollupInterval(ctx, query.GetColumnRollupIntervalRequest())
		if err == nil {
			resp.Query = &runtimev1.QueryBatchResponse_ColumnRollupIntervalResponse{ColumnRollupIntervalResponse: r}
		}

	case runtimev1.QueryBatchType_QUERY_BATCH_TYPE_COLUMN_TOPK:
		var r *runtimev1.ColumnTopKResponse
		r, err = s.ColumnTopK(ctx, query.GetColumnTopKRequest())
		if err == nil {
			resp.Query = &runtimev1.QueryBatchResponse_ColumnTopKResponse{ColumnTopKResponse: r}
		}

	case runtimev1.QueryBatchType_QUERY_BATCH_TYPE_COLUMN_NULL_COUNT:
		var r *runtimev1.ColumnNullCountResponse
		r, err = s.ColumnNullCount(ctx, query.GetColumnNullCountRequest())
		if err == nil {
			resp.Query = &runtimev1.QueryBatchResponse_ColumnNullCountResponse{ColumnNullCountResponse: r}
		}

	case runtimev1.QueryBatchType_QUERY_BATCH_TYPE_COLUMN_DESCRIPTIVE_STATISTICS:
		var r *runtimev1.ColumnDescriptiveStatisticsResponse
		r, err = s.ColumnDescriptiveStatistics(ctx, query.GetColumnDescriptiveStatisticsRequest())
		if err == nil {
			resp.Query = &runtimev1.QueryBatchResponse_ColumnDescriptiveStatisticsResponse{ColumnDescriptiveStatisticsResponse: r}
		}

	case runtimev1.QueryBatchType_QUERY_BATCH_TYPE_COLUMN_TIME_GRAIN:
		var r *runtimev1.ColumnTimeGrainResponse
		r, err = s.ColumnTimeGrain(ctx, query.GetColumnTimeGrainRequest())
		if err == nil {
			resp.Query = &runtimev1.QueryBatchResponse_ColumnTimeGrainResponse{ColumnTimeGrainResponse: r}
		}

	case runtimev1.QueryBatchType_QUERY_BATCH_TYPE_COLUMN_NUMERIC_HISTOGRAM:
		var r *runtimev1.ColumnNumericHistogramResponse
		r, err = s.ColumnNumericHistogram(ctx, query.GetColumnNumericHistogramRequest())
		if err == nil {
			resp.Query = &runtimev1.QueryBatchResponse_ColumnNumericHistogramResponse{ColumnNumericHistogramResponse: r}
		}

	case runtimev1.QueryBatchType_QUERY_BATCH_TYPE_COLUMN_RUG_HISTOGRAM:
		var r *runtimev1.ColumnRugHistogramResponse
		r, err = s.ColumnRugHistogram(ctx, query.GetColumnRugHistogramRequest())
		if err == nil {
			resp.Query = &runtimev1.QueryBatchResponse_ColumnRugHistogramResponse{ColumnRugHistogramResponse: r}
		}

	case runtimev1.QueryBatchType_QUERY_BATCH_TYPE_COLUMN_TIME_RANGE:
		var r *runtimev1.ColumnTimeRangeResponse
		r, err = s.ColumnTimeRange(ctx, query.GetColumnTimeRangeRequest())
		if err == nil {
			resp.Query = &runtimev1.QueryBatchResponse_ColumnTimeRangeResponse{ColumnTimeRangeResponse: r}
		}

	case runtimev1.QueryBatchType_QUERY_BATCH_TYPE_COLUMN_CARDINALITY:
		var r *runtimev1.ColumnCardinalityResponse
		r, err = s.ColumnCardinality(ctx, query.GetColumnCardinalityRequest())
		if err == nil {
			resp.Query = &runtimev1.QueryBatchResponse_ColumnCardinalityResponse{ColumnCardinalityResponse: r}
		}

	case runtimev1.QueryBatchType_QUERY_BATCH_TYPE_COLUMN_TIMESERIES:
		var r *runtimev1.ColumnTimeSeriesResponse
		r, err = s.ColumnTimeSeries(ctx, query.GetColumnTimeSeriesRequest())
		if err == nil {
			resp.Query = &runtimev1.QueryBatchResponse_ColumnTimeSeriesResponse{ColumnTimeSeriesResponse: r}
		}

	case runtimev1.QueryBatchType_QUERY_BATCH_TYPE_TABLE_CARDINALITY:
		var r *runtimev1.TableCardinalityResponse
		r, err = s.TableCardinality(ctx, query.GetTableCardinalityRequest())
		if err == nil {
			resp.Query = &runtimev1.QueryBatchResponse_TableCardinalityResponse{TableCardinalityResponse: r}
		}

	case runtimev1.QueryBatchType_QUERY_BATCH_TYPE_TABLE_COLUMNS:
		var r *runtimev1.TableColumnsResponse
		r, err = s.TableColumns(ctx, query.GetTableColumnsRequest())
		if err == nil {
			resp.Query = &runtimev1.QueryBatchResponse_TableColumnsResponse{TableColumnsResponse: r}
		}

	case runtimev1.QueryBatchType_QUERY_BATCH_TYPE_TABLE_ROWS:
		var r *runtimev1.TableRowsResponse
		r, err = s.TableRows(ctx, query.GetTableRowsRequest())
		if err == nil {
			resp.Query = &runtimev1.QueryBatchResponse_TableRowsResponse{TableRowsResponse: r}
		}
	}

	if err != nil {
		resp.Error = err.Error()
	}

	return resp
}
