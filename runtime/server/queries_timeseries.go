package server

import (
	"context"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/queries"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.opentelemetry.io/otel/attribute"
)

// Metrics/Timeseries APIs
func (s *Server) ColumnRollupInterval(ctx context.Context, req *runtimev1.ColumnRollupIntervalRequest) (*runtimev1.ColumnRollupIntervalResponse, error) {
	observability.SetRequestAttributes(ctx,
		attribute.String("instance_id", req.InstanceId),
		attribute.String("table", req.TableName),
		attribute.String("column", req.ColumnName),
		attribute.Int("priority", int(req.Priority)),
	)

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

func (s *Server) ColumnTimeSeries(ctx context.Context, req *runtimev1.ColumnTimeSeriesRequest) (*runtimev1.ColumnTimeSeriesResponse, error) {
	observability.SetRequestAttributes(ctx,
		attribute.String("instance_id", req.InstanceId),
		attribute.String("table", req.TableName),
		attribute.StringSlice("measures", marshalProtoSlice(req.Measures)),
		attribute.String("timestamp_column", req.TimestampColumnName),
		attribute.String("time_range", marshalProto(req.TimeRange)),
		attribute.Int("Filters", filterCount(req.Filters)),
		attribute.Int("pixels", int(req.Pixels)),
		attribute.Int("sample_size", int(req.SampleSize)),
		attribute.Int("priority", int(req.Priority)),
	)

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

	return &runtimev1.ColumnTimeSeriesResponse{
		Rollup: &runtimev1.TimeSeriesResponse{
			Results:    q.Result.Results,
			Spark:      q.Result.Spark,
			SampleSize: q.Result.SampleSize,
		},
	}, nil
}
