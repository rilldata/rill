package server

import (
	"context"

	"github.com/bufbuild/connect-go"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/queries"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.opentelemetry.io/otel/attribute"
)

// Metrics/Timeseries APIs
func (s *Server) ColumnRollupInterval(ctx context.Context, req *connect.Request[runtimev1.ColumnRollupIntervalRequest]) (*connect.Response[runtimev1.ColumnRollupIntervalResponse], error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.Msg.InstanceId),
		attribute.String("args.table", req.Msg.TableName),
		attribute.String("args.column", req.Msg.ColumnName),
		attribute.Int("args.priority", int(req.Msg.Priority)),
	)

	s.addInstanceRequestAttributes(ctx, req.Msg.InstanceId)

	if !auth.GetClaims(ctx).CanInstance(req.Msg.InstanceId, auth.ReadProfiling) {
		return nil, ErrForbidden
	}

	q := &queries.RollupInterval{
		TableName:  req.Msg.TableName,
		ColumnName: req.Msg.ColumnName,
	}
	err := s.runtime.Query(ctx, req.Msg.InstanceId, q, int(req.Msg.Priority))
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(q.Result), nil
}

func (s *Server) ColumnTimeSeries(ctx context.Context, req *connect.Request[runtimev1.ColumnTimeSeriesRequest]) (*connect.Response[runtimev1.ColumnTimeSeriesResponse], error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.Msg.InstanceId),
		attribute.String("args.table", req.Msg.TableName),
		attribute.StringSlice("args.measures.ids", marshalColumnTimeSeriesRequestBasicMeasure(req.Msg.Measures)),
		attribute.String("args.timestamp_column", req.Msg.TimestampColumnName),
		attribute.Int("args.pixels", int(req.Msg.Pixels)),
		attribute.Int("args.sample_size", int(req.Msg.SampleSize)),
		attribute.Int("args.priority", int(req.Msg.Priority)),
	)

	s.addInstanceRequestAttributes(ctx, req.Msg.InstanceId)

	if req.Msg.TimeRange != nil {
		observability.AddRequestAttributes(ctx, attribute.String("args.time_range.start", safeTimeStr(req.Msg.TimeRange.Start)))
		observability.AddRequestAttributes(ctx, attribute.String("args.time_range.end", safeTimeStr(req.Msg.TimeRange.End)))
		observability.AddRequestAttributes(ctx, attribute.String("args.time_range.interval", req.Msg.TimeRange.Interval.String()))
	}

	if !auth.GetClaims(ctx).CanInstance(req.Msg.InstanceId, auth.ReadProfiling) {
		return nil, ErrForbidden
	}

	q := &queries.ColumnTimeseries{
		TableName:           req.Msg.TableName,
		TimestampColumnName: req.Msg.TimestampColumnName,
		Measures:            req.Msg.Measures,
		TimeRange:           req.Msg.TimeRange,
		Pixels:              req.Msg.Pixels,
		SampleSize:          req.Msg.SampleSize,
		TimeZone:            req.Msg.TimeZone,
	}
	err := s.runtime.Query(ctx, req.Msg.InstanceId, q, int(req.Msg.Priority))
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&runtimev1.ColumnTimeSeriesResponse{
		Rollup: &runtimev1.TimeSeriesResponse{
			Results:    q.Result.Results,
			Spark:      q.Result.Spark,
			SampleSize: q.Result.SampleSize,
		},
	}), nil
}
