package server

import (
	"context"
	"fmt"
	"strings"

	"github.com/bufbuild/connect-go"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/queries"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.opentelemetry.io/otel/attribute"
)

// MetricsViewToplist implements QueryService.
func (s *Server) MetricsViewToplist(ctx context.Context, req *connect.Request[runtimev1.MetricsViewToplistRequest]) (*connect.Response[runtimev1.MetricsViewToplistResponse], error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.Msg.InstanceId),
		attribute.String("args.metric_view", req.Msg.MetricsViewName),
		attribute.String("args.dimension", req.Msg.DimensionName),
		attribute.StringSlice("args.measures", req.Msg.MeasureNames),
		attribute.Int64("args.limit", req.Msg.Limit),
		attribute.Int64("args.offset", req.Msg.Offset),
		attribute.Int("args.priority", int(req.Msg.Priority)),
		attribute.String("args.time_start", safeTimeStr(req.Msg.TimeStart)),
		attribute.String("args.time_end", safeTimeStr(req.Msg.TimeEnd)),
		attribute.StringSlice("args.sort.names", marshalMetricsViewSort(req.Msg.Sort)),
		attribute.StringSlice("args.inline_measures", marshalInlineMeasure(req.Msg.InlineMeasures)),
		attribute.Int("args.filter_count", filterCount(req.Msg.Filter)),
	)

	s.addInstanceRequestAttributes(ctx, req.Msg.InstanceId)

	if !auth.GetClaims(ctx).CanInstance(req.Msg.InstanceId, auth.ReadMetrics) {
		return nil, ErrForbidden
	}

	err := validateInlineMeasures(req.Msg.InlineMeasures)
	if err != nil {
		return nil, err
	}

	q := &queries.MetricsViewToplist{
		MetricsViewName: req.Msg.MetricsViewName,
		DimensionName:   req.Msg.DimensionName,
		MeasureNames:    req.Msg.MeasureNames,
		InlineMeasures:  req.Msg.InlineMeasures,
		TimeStart:       req.Msg.TimeStart,
		TimeEnd:         req.Msg.TimeEnd,
		Limit:           &req.Msg.Limit,
		Offset:          req.Msg.Offset,
		Sort:            req.Msg.Sort,
		Filter:          req.Msg.Filter,
	}
	err = s.runtime.Query(ctx, req.Msg.InstanceId, q, int(req.Msg.Priority))
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(q.Result), nil
}

// MetricsViewComparisonToplist implements QueryService.
func (s *Server) MetricsViewComparisonToplist(ctx context.Context, req *connect.Request[runtimev1.MetricsViewComparisonToplistRequest]) (*connect.Response[runtimev1.MetricsViewComparisonToplistResponse], error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.Msg.InstanceId),
		attribute.String("args.metric_view", req.Msg.MetricsViewName),
		attribute.String("args.dimension", req.Msg.DimensionName),
		attribute.StringSlice("args.measures", req.Msg.MeasureNames),
		attribute.StringSlice("args.inline_measures.names", marshalInlineMeasure(req.Msg.InlineMeasures)),
		attribute.StringSlice("args.sort.names", marshalMetricsViewComparisonSort(req.Msg.Sort)),
		attribute.Int("args.filter_count", filterCount(req.Msg.Filter)),
		attribute.Int64("args.limit", req.Msg.Limit),
		attribute.Int64("args.offset", req.Msg.Offset),
		attribute.Int("args.priority", int(req.Msg.Priority)),
	)

	s.addInstanceRequestAttributes(ctx, req.Msg.InstanceId)

	if req.Msg.BaseTimeRange != nil {
		observability.AddRequestAttributes(ctx, attribute.String("args.base_time_range.start", safeTimeStr(req.Msg.BaseTimeRange.Start)))
		observability.AddRequestAttributes(ctx, attribute.String("args.base_time_range.end", safeTimeStr(req.Msg.BaseTimeRange.End)))
	}
	if req.Msg.ComparisonTimeRange != nil {
		observability.AddRequestAttributes(ctx, attribute.String("args.comparison_time_range.start", safeTimeStr(req.Msg.ComparisonTimeRange.Start)))
		observability.AddRequestAttributes(ctx, attribute.String("args.comparison_time_range.end", safeTimeStr(req.Msg.ComparisonTimeRange.End)))
	}

	if !auth.GetClaims(ctx).CanInstance(req.Msg.InstanceId, auth.ReadMetrics) {
		return nil, ErrForbidden
	}

	err := validateInlineMeasures(req.Msg.InlineMeasures)
	if err != nil {
		return nil, err
	}

	q := &queries.MetricsViewComparisonToplist{
		MetricsViewName:     req.Msg.MetricsViewName,
		DimensionName:       req.Msg.DimensionName,
		MeasureNames:        req.Msg.MeasureNames,
		InlineMeasures:      req.Msg.InlineMeasures,
		BaseTimeRange:       req.Msg.BaseTimeRange,
		ComparisonTimeRange: req.Msg.ComparisonTimeRange,
		Limit:               req.Msg.Limit,
		Offset:              req.Msg.Offset,
		Sort:                req.Msg.Sort,
		Filter:              req.Msg.Filter,
	}
	err = s.runtime.Query(ctx, req.Msg.InstanceId, q, int(req.Msg.Priority))
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(q.Result), nil
}

// MetricsViewTimeSeries implements QueryService.
func (s *Server) MetricsViewTimeSeries(ctx context.Context, req *connect.Request[runtimev1.MetricsViewTimeSeriesRequest]) (*connect.Response[runtimev1.MetricsViewTimeSeriesResponse], error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.Msg.InstanceId),
		attribute.String("args.metric_view", req.Msg.MetricsViewName),
		attribute.StringSlice("args.measures", req.Msg.MeasureNames),
		attribute.StringSlice("args.inline_measures.names", marshalInlineMeasure(req.Msg.InlineMeasures)),
		attribute.String("args.time_start", safeTimeStr(req.Msg.TimeStart)),
		attribute.String("args.time_end", safeTimeStr(req.Msg.TimeEnd)),
		attribute.String("args.time_granularity", req.Msg.TimeGranularity.String()),
		attribute.Int("args.filter_count", filterCount(req.Msg.Filter)),
		attribute.Int("args.priority", int(req.Msg.Priority)),
	)

	s.addInstanceRequestAttributes(ctx, req.Msg.InstanceId)

	if !auth.GetClaims(ctx).CanInstance(req.Msg.InstanceId, auth.ReadMetrics) {
		return nil, ErrForbidden
	}

	err := validateInlineMeasures(req.Msg.InlineMeasures)
	if err != nil {
		return nil, err
	}

	q := &queries.MetricsViewTimeSeries{
		MetricsViewName: req.Msg.MetricsViewName,
		MeasureNames:    req.Msg.MeasureNames,
		InlineMeasures:  req.Msg.InlineMeasures,
		TimeStart:       req.Msg.TimeStart,
		TimeEnd:         req.Msg.TimeEnd,
		TimeGranularity: req.Msg.TimeGranularity,
		Filter:          req.Msg.Filter,
		TimeZone:        req.Msg.TimeZone,
	}
	err = s.runtime.Query(ctx, req.Msg.InstanceId, q, int(req.Msg.Priority))
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(q.Result), nil
}

// MetricsViewTotals implements QueryService.
func (s *Server) MetricsViewTotals(ctx context.Context, req *connect.Request[runtimev1.MetricsViewTotalsRequest]) (*connect.Response[runtimev1.MetricsViewTotalsResponse], error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.Msg.InstanceId),
		attribute.String("args.metric_view", req.Msg.MetricsViewName),
		attribute.StringSlice("args.measures", req.Msg.MeasureNames),
		attribute.StringSlice("args.inline_measures.names", marshalInlineMeasure(req.Msg.InlineMeasures)),
		attribute.String("args.time_start", safeTimeStr(req.Msg.TimeStart)),
		attribute.String("args.time_end", safeTimeStr(req.Msg.TimeEnd)),
		attribute.Int("args.filter_count", filterCount(req.Msg.Filter)),
		attribute.Int("args.priority", int(req.Msg.Priority)),
	)

	s.addInstanceRequestAttributes(ctx, req.Msg.InstanceId)

	if !auth.GetClaims(ctx).CanInstance(req.Msg.InstanceId, auth.ReadMetrics) {
		return nil, ErrForbidden
	}

	err := validateInlineMeasures(req.Msg.InlineMeasures)
	if err != nil {
		return nil, err
	}

	q := &queries.MetricsViewTotals{
		MetricsViewName: req.Msg.MetricsViewName,
		MeasureNames:    req.Msg.MeasureNames,
		InlineMeasures:  req.Msg.InlineMeasures,
		TimeStart:       req.Msg.TimeStart,
		TimeEnd:         req.Msg.TimeEnd,
		Filter:          req.Msg.Filter,
	}
	err = s.runtime.Query(ctx, req.Msg.InstanceId, q, int(req.Msg.Priority))
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(q.Result), nil
}

// MetricsViewRows implements QueryService.
func (s *Server) MetricsViewRows(ctx context.Context, req *connect.Request[runtimev1.MetricsViewRowsRequest]) (*connect.Response[runtimev1.MetricsViewRowsResponse], error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.Msg.InstanceId),
		attribute.String("args.metric_view", req.Msg.MetricsViewName),
		attribute.String("args.time_start", safeTimeStr(req.Msg.TimeStart)),
		attribute.String("args.time_end", safeTimeStr(req.Msg.TimeEnd)),
		attribute.String("args.time_granularity", req.Msg.TimeGranularity.String()),
		attribute.Int("args.filter_count", filterCount(req.Msg.Filter)),
		attribute.StringSlice("args.sort.names", marshalMetricsViewSort(req.Msg.Sort)),
		attribute.Int("args.limit", int(req.Msg.Limit)),
		attribute.Int64("args.offset", req.Msg.Offset),
		attribute.Int("args.priority", int(req.Msg.Priority)),
	)

	s.addInstanceRequestAttributes(ctx, req.Msg.InstanceId)

	if !auth.GetClaims(ctx).CanInstance(req.Msg.InstanceId, auth.ReadMetrics) {
		return nil, ErrForbidden
	}

	limit := int64(req.Msg.Limit)

	q := &queries.MetricsViewRows{
		MetricsViewName: req.Msg.MetricsViewName,
		TimeStart:       req.Msg.TimeStart,
		TimeEnd:         req.Msg.TimeEnd,
		TimeGranularity: req.Msg.TimeGranularity,
		Filter:          req.Msg.Filter,
		Sort:            req.Msg.Sort,
		Limit:           &limit,
		Offset:          req.Msg.Offset,
		TimeZone:        req.Msg.TimeZone,
	}
	err := s.runtime.Query(ctx, req.Msg.InstanceId, q, int(req.Msg.Priority))
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(q.Result), nil
}

// MetricsViewTimeRange implements QueryService.
func (s *Server) MetricsViewTimeRange(ctx context.Context, req *connect.Request[runtimev1.MetricsViewTimeRangeRequest]) (*connect.Response[runtimev1.MetricsViewTimeRangeResponse], error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.Msg.InstanceId),
		attribute.String("args.metric_view", req.Msg.MetricsViewName),
		attribute.Int("args.priority", int(req.Msg.Priority)),
	)

	s.addInstanceRequestAttributes(ctx, req.Msg.InstanceId)

	if !auth.GetClaims(ctx).CanInstance(req.Msg.InstanceId, auth.ReadMetrics) {
		return nil, ErrForbidden
	}

	q := &queries.MetricsViewTimeRange{
		MetricsViewName: req.Msg.MetricsViewName,
	}
	err := s.runtime.Query(ctx, req.Msg.InstanceId, q, int(req.Msg.Priority))
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(q.Result), nil
}

// validateInlineMeasures checks that the inline measures are allowed.
// This is to prevent injection of arbitrary SQL from clients with only ReadMetrics access.
// In the future, we should consider allowing arbitrary expressions from people with wider access.
// Currently, only COUNT(*) is allowed.
func validateInlineMeasures(ms []*runtimev1.InlineMeasure) error {
	for _, im := range ms {
		if !strings.EqualFold(im.Expression, "COUNT(*)") {
			return fmt.Errorf("illegal inline measure expression: %q", im.Expression)
		}
	}
	return nil
}
