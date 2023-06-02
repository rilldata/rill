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
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.table", req.TableName),
		attribute.String("args.column", req.ColumnName),
		attribute.Int("args.priority", int(req.Priority)),
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
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.table", req.TableName),
		attribute.StringSlice("args.measures.ids", marshalColumnTimeSeriesRequestBasicMeasure(req.Measures)),
		attribute.String("args.timestamp_column", req.TimestampColumnName),
		attribute.Int("args.filter_count", filterCount(req.Filters)),
		attribute.Int("args.pixels", int(req.Pixels)),
		attribute.Int("args.sample_size", int(req.SampleSize)),
		attribute.Int("args.priority", int(req.Priority)),
	)
	if req.TimeRange != nil {
		observability.SetRequestAttributes(ctx, attribute.String("args.time_range.start", safeTimeStr(req.TimeRange.Start)))
		observability.SetRequestAttributes(ctx, attribute.String("args.time_range.end", safeTimeStr(req.TimeRange.End)))
		observability.SetRequestAttributes(ctx, attribute.String("args.time_range.interval", req.TimeRange.Interval.String()))
	}

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
