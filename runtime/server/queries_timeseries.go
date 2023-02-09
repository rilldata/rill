package server

import (
	"context"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/queries"
)

// Metrics/Timeseries APIs
func (s *Server) ColumnRollupInterval(ctx context.Context, request *runtimev1.ColumnRollupIntervalRequest) (*runtimev1.ColumnRollupIntervalResponse, error) {
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

func (s *Server) ColumnTimeSeries(ctx context.Context, request *runtimev1.ColumnTimeSeriesRequest) (*runtimev1.ColumnTimeSeriesResponse, error) {
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

	return &runtimev1.ColumnTimeSeriesResponse{
		Rollup: &runtimev1.TimeSeriesResponse{
			Results:    q.Result.Results,
			Spark:      q.Result.Spark,
			TimeRange:  q.Result.TimeRange,
			SampleSize: q.Result.SampleSize,
		},
	}, nil
}
