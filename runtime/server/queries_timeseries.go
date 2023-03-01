package server

import (
	"context"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/queries"
	"github.com/rilldata/rill/runtime/server/auth"
)

// Metrics/Timeseries APIs
func (s *Server) EstimateRollupInterval(ctx context.Context, req *runtimev1.EstimateRollupIntervalRequest) (*runtimev1.EstimateRollupIntervalResponse, error) {
	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.ReadProfiling) {
		return nil, ErrForbidden
	}

	q := &queries.RollupInterval{
		TableName:  req.TableName,
		ColumnName: req.ColumnName,
	}
	err := s.runtime.Query(ctx, req.InstanceId, q, int(req.Priority))
	if err != nil {
		return nil, err
	}

	return q.Result, nil
}

func (s *Server) GenerateTimeSeries(ctx context.Context, req *runtimev1.GenerateTimeSeriesRequest) (*runtimev1.GenerateTimeSeriesResponse, error) {
	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.ReadProfiling) {
		return nil, ErrForbidden
	}

	q := &queries.ColumnTimeseries{
		TableName:           req.TableName,
		TimestampColumnName: req.TimestampColumnName,
		Measures:            req.Measures,
		Filters:             req.Filters,
		TimeRange:           req.TimeRange,
		Pixels:              req.Pixels,
		SampleSize:          req.SampleSize,
	}
	err := s.runtime.Query(ctx, req.InstanceId, q, int(req.Priority))
	if err != nil {
		return nil, err
	}

	return &runtimev1.GenerateTimeSeriesResponse{
		Rollup: &runtimev1.TimeSeriesResponse{
			Results:    q.Result.Results,
			Spark:      q.Result.Spark,
			TimeRange:  q.Result.TimeRange,
			SampleSize: q.Result.SampleSize,
		},
	}, nil
}
