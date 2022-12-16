package server

import (
	"context"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/queries"
)

// Metrics/Timeseries APIs
func (s *Server) EstimateRollupInterval(ctx context.Context, request *runtimev1.EstimateRollupIntervalRequest) (*runtimev1.EstimateRollupIntervalResponse, error) {
	q := &queries.RollupInterval{
		TableName:  request.TableName,
		ColumnName: request.ColumnName,
	}
	err := s.runtime.Query(ctx, request.InstanceId, q, int(request.Priority))
	if err != nil {
		return nil, err
	}

	return q.Result, nil
}

func (s *Server) GenerateTimeSeries(ctx context.Context, request *runtimev1.GenerateTimeSeriesRequest) (*runtimev1.GenerateTimeSeriesResponse, error) {
<<<<<<< HEAD
	q := &queries.ColumnTimeseries{
		TableName:           request.TableName,
		TimestampColumnName: request.TimestampColumnName,
		Measures:            request.Measures,
		Filters:             request.Filters,
		TimeRange:           request.TimeRange,
		Pixels:              request.Pixels,
		SampleSize:          request.SampleSize,
	}
	err := s.runtime.Query(ctx, request.InstanceId, q, int(request.Priority))
	if err != nil {
		return nil, err
	}

	return &runtimev1.GenerateTimeSeriesResponse{
		Rollup: q.Result,
	}, nil
}
