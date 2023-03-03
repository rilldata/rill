package server

import (
	"context"
	"fmt"
	"sync"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/pbutil"
	"github.com/rilldata/rill/runtime/server/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

// Query implements QueryService.
func (s *Server) Query(ctx context.Context, req *runtimev1.QueryRequest) (*runtimev1.QueryResponse, error) {
	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.ReadOLAP) {
		return nil, ErrForbidden
	}

	args := make([]any, len(req.Args))
	for i, arg := range req.Args {
		args[i] = arg.AsInterface()
	}

	olap, err := s.runtime.OLAP(ctx, req.InstanceId)
	if err != nil {
		return nil, err
	}

	res, err := olap.Execute(ctx, &drivers.Statement{
		Query:    req.Sql,
		Args:     args,
		DryRun:   req.DryRun,
		Priority: int(req.Priority),
	})
	if err != nil {
		// TODO: Parse error to determine error code
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// NOTE: Currently, query returns nil res for successful dry-run queries
	if req.DryRun {
		// TODO: Return a meta object for dry-run queries
		return &runtimev1.QueryResponse{}, nil
	}

	defer res.Close()

	data, err := rowsToData(res)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &runtimev1.QueryResponse{
		Meta: res.Schema,
		Data: data,
	}

	return resp, nil
}

func (s *Server) QueryBatch(req *runtimev1.QueryBatchRequest, srv runtimev1.QueryService_QueryBatchServer) error {
	var wg sync.WaitGroup

	for _, query := range req.Queries {
		wg.Add(1)
		go func(query *runtimev1.QueryBatchSingleRequest) {
			defer wg.Done()

			resp := s.forwardQuery(srv.Context(), query)

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
	case runtimev1.QueryBatchType_MetricsViewToplist:
		var r *runtimev1.MetricsViewToplistResponse
		r, err = s.MetricsViewToplist(ctx, query.GetMetricsViewToplistRequest())
		if err == nil {
			resp.Query = &runtimev1.QueryBatchResponse_MetricsViewToplistResponse{MetricsViewToplistResponse: r}
		}

	case runtimev1.QueryBatchType_MetricsViewTimeSeries:
		var r *runtimev1.MetricsViewTimeSeriesResponse
		r, err = s.MetricsViewTimeSeries(ctx, query.GetMetricsViewTimeSeriesRequest())
		if err == nil {
			resp.Query = &runtimev1.QueryBatchResponse_MetricsViewTimeSeriesResponse{MetricsViewTimeSeriesResponse: r}
		}

	case runtimev1.QueryBatchType_MetricsViewTotals:
		var r *runtimev1.MetricsViewTotalsResponse
		r, err = s.MetricsViewTotals(ctx, query.GetMetricsViewTotalsRequest())
		if err == nil {
			resp.Query = &runtimev1.QueryBatchResponse_MetricsViewTotalsResponse{MetricsViewTotalsResponse: r}
		}

	case runtimev1.QueryBatchType_ColumnRollupInterval:
		var r *runtimev1.ColumnRollupIntervalResponse
		r, err = s.ColumnRollupInterval(ctx, query.GetColumnRollupIntervalRequest())
		if err == nil {
			resp.Query = &runtimev1.QueryBatchResponse_ColumnRollupIntervalResponse{ColumnRollupIntervalResponse: r}
		}

	case runtimev1.QueryBatchType_ColumnTopK:
		var r *runtimev1.ColumnTopKResponse
		r, err = s.ColumnTopK(ctx, query.GetColumnTopKRequest())
		if err == nil {
			resp.Query = &runtimev1.QueryBatchResponse_ColumnTopKResponse{ColumnTopKResponse: r}
		}

	case runtimev1.QueryBatchType_ColumnNullCount:
		var r *runtimev1.ColumnNullCountResponse
		r, err = s.ColumnNullCount(ctx, query.GetColumnNullCountRequest())
		if err == nil {
			resp.Query = &runtimev1.QueryBatchResponse_ColumnNullCountResponse{ColumnNullCountResponse: r}
		}

	case runtimev1.QueryBatchType_ColumnDescriptiveStatistics:
		var r *runtimev1.ColumnDescriptiveStatisticsResponse
		r, err = s.ColumnDescriptiveStatistics(ctx, query.GetColumnDescriptiveStatisticsRequest())
		if err == nil {
			resp.Query = &runtimev1.QueryBatchResponse_ColumnDescriptiveStatisticsResponse{ColumnDescriptiveStatisticsResponse: r}
		}

	case runtimev1.QueryBatchType_ColumnTimeGrain:
		var r *runtimev1.ColumnTimeGrainResponse
		r, err = s.ColumnTimeGrain(ctx, query.GetColumnTimeGrainRequest())
		if err == nil {
			resp.Query = &runtimev1.QueryBatchResponse_ColumnTimeGrainResponse{ColumnTimeGrainResponse: r}
		}

	case runtimev1.QueryBatchType_ColumnNumericHistogram:
		var r *runtimev1.ColumnNumericHistogramResponse
		r, err = s.ColumnNumericHistogram(ctx, query.GetColumnNumericHistogramRequest())
		if err == nil {
			resp.Query = &runtimev1.QueryBatchResponse_ColumnNumericHistogramResponse{ColumnNumericHistogramResponse: r}
		}

	case runtimev1.QueryBatchType_ColumnRugHistogram:
		var r *runtimev1.ColumnRugHistogramResponse
		r, err = s.ColumnRugHistogram(ctx, query.GetColumnRugHistogramRequest())
		if err == nil {
			resp.Query = &runtimev1.QueryBatchResponse_ColumnRugHistogramResponse{ColumnRugHistogramResponse: r}
		}

	case runtimev1.QueryBatchType_ColumnTimeRange:
		var r *runtimev1.ColumnTimeRangeResponse
		r, err = s.ColumnTimeRange(ctx, query.GetColumnTimeRangeRequest())
		if err == nil {
			resp.Query = &runtimev1.QueryBatchResponse_ColumnTimeRangeResponse{ColumnTimeRangeResponse: r}
		}

	case runtimev1.QueryBatchType_ColumnCardinality:
		var r *runtimev1.ColumnCardinalityResponse
		r, err = s.ColumnCardinality(ctx, query.GetColumnCardinalityRequest())
		if err == nil {
			resp.Query = &runtimev1.QueryBatchResponse_ColumnCardinalityResponse{ColumnCardinalityResponse: r}
		}

	case runtimev1.QueryBatchType_ColumnTimeSeries:
		var r *runtimev1.ColumnTimeSeriesResponse
		r, err = s.ColumnTimeSeries(ctx, query.GetColumnTimeSeriesRequest())
		if err == nil {
			resp.Query = &runtimev1.QueryBatchResponse_ColumnTimeSeriesResponse{ColumnTimeSeriesResponse: r}
		}

	case runtimev1.QueryBatchType_TableCardinality:
		var r *runtimev1.TableCardinalityResponse
		r, err = s.TableCardinality(ctx, query.GetTableCardinalityRequest())
		if err == nil {
			resp.Query = &runtimev1.QueryBatchResponse_TableCardinalityResponse{TableCardinalityResponse: r}
		}

	case runtimev1.QueryBatchType_TableColumns:
		var r *runtimev1.TableColumnsResponse
		r, err = s.TableColumns(ctx, query.GetTableColumnsRequest())
		if err == nil {
			resp.Query = &runtimev1.QueryBatchResponse_TableColumnsResponse{TableColumnsResponse: r}
		}

	case runtimev1.QueryBatchType_TableRows:
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

func rowsToData(rows *drivers.Result) ([]*structpb.Struct, error) {
	var data []*structpb.Struct
	for rows.Next() {
		rowMap := make(map[string]any)
		err := rows.MapScan(rowMap)
		if err != nil {
			return nil, err
		}

		rowStruct, err := pbutil.ToStruct(rowMap)
		if err != nil {
			return nil, err
		}

		data = append(data, rowStruct)
	}

	err := rows.Err()
	if err != nil {
		return nil, err
	}

	return data, nil
}
