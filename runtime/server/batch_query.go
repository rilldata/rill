package server

import (
	"context"

	"github.com/bufbuild/connect-go"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"golang.org/x/sync/errgroup"
)

func (s *Server) QueryBatch(ctx context.Context, req *connect.Request[runtimev1.QueryBatchRequest], srv *connect.ServerStream[runtimev1.QueryBatchResponse]) error {
	// TODO: Performance improvements:
	//       1. Check for access upfront based on what is in the request
	//       2. Check for cache and return those queries immediately before creating a goroutine
	//       3. Use a goroutine pool with size equal to driver's concurrency to execute the queries

	g, ctx := errgroup.WithContext(ctx)

	for _, query := range req.Msg.Queries {
		query := query // create a closed var for goroutine
		g.Go(func() error {
			resp := s.forwardQuery(ctx, req.Msg, query)
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
		var r *connect.Response[runtimev1.MetricsViewToplistResponse]
		q.MetricsViewToplistRequest.InstanceId = query.InstanceId
		r, err = s.MetricsViewToplist(ctx, connect.NewRequest(q.MetricsViewToplistRequest))
		if err == nil {
			resp.Result = &runtimev1.QueryBatchResponse_MetricsViewToplistResponse{MetricsViewToplistResponse: r.Msg}
		}

	case *runtimev1.QueryBatchEntry_MetricsViewComparisonToplistRequest:
		var r *connect.Response[runtimev1.MetricsViewComparisonToplistResponse]
		q.MetricsViewComparisonToplistRequest.InstanceId = query.InstanceId
		r, err = s.MetricsViewComparisonToplist(ctx, connect.NewRequest(q.MetricsViewComparisonToplistRequest))
		if err == nil {
			resp.Result = &runtimev1.QueryBatchResponse_MetricsViewComparisonToplistResponse{MetricsViewComparisonToplistResponse: r.Msg}
		}

	case *runtimev1.QueryBatchEntry_MetricsViewTimeSeriesRequest:
		var r *connect.Response[runtimev1.MetricsViewTimeSeriesResponse]
		q.MetricsViewTimeSeriesRequest.InstanceId = query.InstanceId
		r, err = s.MetricsViewTimeSeries(ctx, connect.NewRequest(q.MetricsViewTimeSeriesRequest))
		if err == nil {
			resp.Result = &runtimev1.QueryBatchResponse_MetricsViewTimeSeriesResponse{MetricsViewTimeSeriesResponse: r.Msg}
		}

	case *runtimev1.QueryBatchEntry_MetricsViewTotalsRequest:
		var r *connect.Response[runtimev1.MetricsViewTotalsResponse]
		q.MetricsViewTotalsRequest.InstanceId = query.InstanceId
		r, err = s.MetricsViewTotals(ctx, connect.NewRequest(q.MetricsViewTotalsRequest))
		if err == nil {
			resp.Result = &runtimev1.QueryBatchResponse_MetricsViewTotalsResponse{MetricsViewTotalsResponse: r.Msg}
		}

	case *runtimev1.QueryBatchEntry_MetricsViewRowsRequest:
		var r *connect.Response[runtimev1.MetricsViewRowsResponse]
		q.MetricsViewRowsRequest.InstanceId = query.InstanceId
		r, err = s.MetricsViewRows(ctx, connect.NewRequest(q.MetricsViewRowsRequest))
		if err == nil {
			resp.Result = &runtimev1.QueryBatchResponse_MetricsViewRowsResponse{MetricsViewRowsResponse: r.Msg}
		}

	case *runtimev1.QueryBatchEntry_ColumnRollupIntervalRequest:
		var r *connect.Response[runtimev1.ColumnRollupIntervalResponse]
		q.ColumnRollupIntervalRequest.InstanceId = query.InstanceId
		r, err = s.ColumnRollupInterval(ctx, connect.NewRequest(q.ColumnRollupIntervalRequest))
		if err == nil {
			resp.Result = &runtimev1.QueryBatchResponse_ColumnRollupIntervalResponse{ColumnRollupIntervalResponse: r.Msg}
		}

	case *runtimev1.QueryBatchEntry_ColumnTopKRequest:
		var r *connect.Response[runtimev1.ColumnTopKResponse]
		q.ColumnTopKRequest.InstanceId = query.InstanceId
		r, err = s.ColumnTopK(ctx, connect.NewRequest(q.ColumnTopKRequest))
		if err == nil {
			resp.Result = &runtimev1.QueryBatchResponse_ColumnTopKResponse{ColumnTopKResponse: r.Msg}
		}

	case *runtimev1.QueryBatchEntry_ColumnNullCountRequest:
		var r *connect.Response[runtimev1.ColumnNullCountResponse]
		q.ColumnNullCountRequest.InstanceId = query.InstanceId
		r, err = s.ColumnNullCount(ctx, connect.NewRequest(q.ColumnNullCountRequest))
		if err == nil {
			resp.Result = &runtimev1.QueryBatchResponse_ColumnNullCountResponse{ColumnNullCountResponse: r.Msg}
		}

	case *runtimev1.QueryBatchEntry_ColumnDescriptiveStatisticsRequest:
		var r *connect.Response[runtimev1.ColumnDescriptiveStatisticsResponse]
		q.ColumnDescriptiveStatisticsRequest.InstanceId = query.InstanceId
		r, err = s.ColumnDescriptiveStatistics(ctx, connect.NewRequest(q.ColumnDescriptiveStatisticsRequest))
		if err == nil {
			resp.Result = &runtimev1.QueryBatchResponse_ColumnDescriptiveStatisticsResponse{ColumnDescriptiveStatisticsResponse: r.Msg}
		}

	case *runtimev1.QueryBatchEntry_ColumnTimeGrainRequest:
		var r *connect.Response[runtimev1.ColumnTimeGrainResponse]
		q.ColumnTimeGrainRequest.InstanceId = query.InstanceId
		r, err = s.ColumnTimeGrain(ctx, connect.NewRequest(q.ColumnTimeGrainRequest))
		if err == nil {
			resp.Result = &runtimev1.QueryBatchResponse_ColumnTimeGrainResponse{ColumnTimeGrainResponse: r.Msg}
		}

	case *runtimev1.QueryBatchEntry_ColumnNumericHistogramRequest:
		var r *connect.Response[runtimev1.ColumnNumericHistogramResponse]
		q.ColumnNumericHistogramRequest.InstanceId = query.InstanceId
		r, err = s.ColumnNumericHistogram(ctx, connect.NewRequest(q.ColumnNumericHistogramRequest))
		if err == nil {
			resp.Result = &runtimev1.QueryBatchResponse_ColumnNumericHistogramResponse{ColumnNumericHistogramResponse: r.Msg}
		}

	case *runtimev1.QueryBatchEntry_ColumnRugHistogramRequest:
		var r *connect.Response[runtimev1.ColumnRugHistogramResponse]
		q.ColumnRugHistogramRequest.InstanceId = query.InstanceId
		r, err = s.ColumnRugHistogram(ctx, connect.NewRequest(q.ColumnRugHistogramRequest))
		if err == nil {
			resp.Result = &runtimev1.QueryBatchResponse_ColumnRugHistogramResponse{ColumnRugHistogramResponse: r.Msg}
		}

	case *runtimev1.QueryBatchEntry_ColumnTimeRangeRequest:
		var r *connect.Response[runtimev1.ColumnTimeRangeResponse]
		q.ColumnTimeRangeRequest.InstanceId = query.InstanceId
		r, err = s.ColumnTimeRange(ctx, connect.NewRequest(q.ColumnTimeRangeRequest))
		if err == nil {
			resp.Result = &runtimev1.QueryBatchResponse_ColumnTimeRangeResponse{ColumnTimeRangeResponse: r.Msg}
		}

	case *runtimev1.QueryBatchEntry_ColumnCardinalityRequest:
		var r *connect.Response[runtimev1.ColumnCardinalityResponse]
		q.ColumnCardinalityRequest.InstanceId = query.InstanceId
		r, err = s.ColumnCardinality(ctx, connect.NewRequest(q.ColumnCardinalityRequest))
		if err == nil {
			resp.Result = &runtimev1.QueryBatchResponse_ColumnCardinalityResponse{ColumnCardinalityResponse: r.Msg}
		}

	case *runtimev1.QueryBatchEntry_ColumnTimeSeriesRequest:
		var r *connect.Response[runtimev1.ColumnTimeSeriesResponse]
		q.ColumnTimeSeriesRequest.InstanceId = query.InstanceId
		r, err = s.ColumnTimeSeries(ctx, connect.NewRequest(q.ColumnTimeSeriesRequest))
		if err == nil {
			resp.Result = &runtimev1.QueryBatchResponse_ColumnTimeSeriesResponse{ColumnTimeSeriesResponse: r.Msg}
		}

	case *runtimev1.QueryBatchEntry_TableCardinalityRequest:
		var r *connect.Response[runtimev1.TableCardinalityResponse]
		q.TableCardinalityRequest.InstanceId = query.InstanceId
		r, err = s.TableCardinality(ctx, connect.NewRequest(q.TableCardinalityRequest))
		if err == nil {
			resp.Result = &runtimev1.QueryBatchResponse_TableCardinalityResponse{TableCardinalityResponse: r.Msg}
		}

	case *runtimev1.QueryBatchEntry_TableColumnsRequest:
		var r *connect.Response[runtimev1.TableColumnsResponse]
		q.TableColumnsRequest.InstanceId = query.InstanceId
		r, err = s.TableColumns(ctx, connect.NewRequest(q.TableColumnsRequest))
		if err == nil {
			resp.Result = &runtimev1.QueryBatchResponse_TableColumnsResponse{TableColumnsResponse: r.Msg}
		}

	case *runtimev1.QueryBatchEntry_TableRowsRequest:
		var r *connect.Response[runtimev1.TableRowsResponse]
		q.TableRowsRequest.InstanceId = query.InstanceId
		r, err = s.TableRows(ctx, connect.NewRequest(q.TableRowsRequest))
		if err == nil {
			resp.Result = &runtimev1.QueryBatchResponse_TableRowsResponse{TableRowsResponse: r.Msg}
		}
	}

	if err != nil {
		resp.Error = err.Error()
	}

	return resp
}
