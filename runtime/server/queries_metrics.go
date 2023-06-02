package server

import (
	"context"
	"fmt"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/queries"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.opentelemetry.io/otel/attribute"
)

// MetricsViewToplist implements QueryService.
func (s *Server) MetricsViewToplist(ctx context.Context, req *runtimev1.MetricsViewToplistRequest) (*runtimev1.MetricsViewToplistResponse, error) {
	observability.SetRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.metric_view", req.MetricsViewName),
		attribute.String("args.dimension", req.DimensionName),
		attribute.StringSlice("args.measures", req.MeasureNames),
		attribute.Int64("args.limit", req.Limit),
		attribute.Int64("args.offset", req.Offset),
		attribute.Int("args.priority", int(req.Priority)),
		attribute.String("args.time_start", safeTimeStr(req.TimeStart)),
		attribute.String("args.time_end", safeTimeStr(req.TimeEnd)),
		attribute.StringSlice("args.sort.names", marshalMetricsViewSort(req.Sort)),
		attribute.StringSlice("args.inline_measures", marshalInlineMeasure(req.InlineMeasures)),
		attribute.Int("args.filter_count", filterCount(req.Filter)),
	)

	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.ReadMetrics) {
		return nil, ErrForbidden
	}

	err := validateInlineMeasures(req.InlineMeasures)
	if err != nil {
		return nil, err
	}

	q := &queries.MetricsViewToplist{
		MetricsViewName: req.MetricsViewName,
		DimensionName:   req.DimensionName,
		MeasureNames:    req.MeasureNames,
		InlineMeasures:  req.InlineMeasures,
		TimeStart:       req.TimeStart,
		TimeEnd:         req.TimeEnd,
		Limit:           req.Limit,
		Offset:          req.Offset,
		Sort:            req.Sort,
		Filter:          req.Filter,
	}
	err = s.runtime.Query(ctx, req.InstanceId, q, int(req.Priority))
	if err != nil {
		return nil, err
	}

	return q.Result, nil
}

// MetricsViewComparisonToplist implements QueryService.
func (s *Server) MetricsViewComparisonToplist(ctx context.Context, req *runtimev1.MetricsViewComparisonToplistRequest) (*runtimev1.MetricsViewComparisonToplistResponse, error) {
	observability.SetRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.metric_view", req.MetricsViewName),
		attribute.String("args.dimension", req.DimensionName),
		attribute.StringSlice("args.measures", req.MeasureNames),
		attribute.StringSlice("args.inline_measures.names", marshalInlineMeasure(req.InlineMeasures)),
		attribute.StringSlice("args.sort.names", marshalMetricsViewComparisonSort(req.Sort)),
		attribute.Int("args.filter_count", filterCount(req.Filter)),
		attribute.Int64("args.limit", req.Limit),
		attribute.Int64("args.offset", req.Offset),
		attribute.Int("args.priority", int(req.Priority)),
	)
	if req.BaseTimeRange != nil {
		observability.SetRequestAttributes(ctx, attribute.String("args.base_time_range.start", safeTimeStr(req.BaseTimeRange.Start)))
		observability.SetRequestAttributes(ctx, attribute.String("args.base_time_range.end", safeTimeStr(req.BaseTimeRange.End)))
	}
	if req.ComparisonTimeRange != nil {
		observability.SetRequestAttributes(ctx, attribute.String("args.comparison_time_range.start", safeTimeStr(req.ComparisonTimeRange.Start)))
		observability.SetRequestAttributes(ctx, attribute.String("args.comparison_time_range.end", safeTimeStr(req.ComparisonTimeRange.End)))
	}

	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.ReadMetrics) {
		return nil, ErrForbidden
	}

	err := validateInlineMeasures(req.InlineMeasures)
	if err != nil {
		return nil, err
	}

	q := &queries.MetricsViewComparisonToplist{
		MetricsViewName:     req.MetricsViewName,
		DimensionName:       req.DimensionName,
		MeasureNames:        req.MeasureNames,
		InlineMeasures:      req.InlineMeasures,
		BaseTimeRange:       req.BaseTimeRange,
		ComparisonTimeRange: req.ComparisonTimeRange,
		Limit:               req.Limit,
		Offset:              req.Offset,
		Sort:                req.Sort,
		Filter:              req.Filter,
	}
	err = s.runtime.Query(ctx, req.InstanceId, q, int(req.Priority))
	if err != nil {
		return nil, err
	}

	return q.Result, nil
}

// MetricsViewTimeSeries implements QueryService.
func (s *Server) MetricsViewTimeSeries(ctx context.Context, req *runtimev1.MetricsViewTimeSeriesRequest) (*runtimev1.MetricsViewTimeSeriesResponse, error) {
	observability.SetRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.metric_view", req.MetricsViewName),
		attribute.StringSlice("args.measures", req.MeasureNames),
		attribute.StringSlice("args.inline_measures.names", marshalInlineMeasure(req.InlineMeasures)),
		attribute.String("args.time_start", safeTimeStr(req.TimeStart)),
		attribute.String("args.time_end", safeTimeStr(req.TimeEnd)),
		attribute.String("args.time_granularity", req.TimeGranularity.String()),
		attribute.Int("args.filter_count", filterCount(req.Filter)),
		attribute.Int("args.priority", int(req.Priority)),
	)

	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.ReadMetrics) {
		return nil, ErrForbidden
	}

	err := validateInlineMeasures(req.InlineMeasures)
	if err != nil {
		return nil, err
	}

	q := &queries.MetricsViewTimeSeries{
		MetricsViewName: req.MetricsViewName,
		MeasureNames:    req.MeasureNames,
		InlineMeasures:  req.InlineMeasures,
		TimeStart:       req.TimeStart,
		TimeEnd:         req.TimeEnd,
		TimeGranularity: req.TimeGranularity,
		Filter:          req.Filter,
	}
	err = s.runtime.Query(ctx, req.InstanceId, q, int(req.Priority))
	if err != nil {
		return nil, err
	}
	return q.Result, nil
}

// MetricsViewTotals implements QueryService.
func (s *Server) MetricsViewTotals(ctx context.Context, req *runtimev1.MetricsViewTotalsRequest) (*runtimev1.MetricsViewTotalsResponse, error) {
	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.ReadMetrics) {
		return nil, ErrForbidden
	}

	err := validateInlineMeasures(req.InlineMeasures)
	if err != nil {
		return nil, err
	}

	q := &queries.MetricsViewTotals{
		MetricsViewName: req.MetricsViewName,
		MeasureNames:    req.MeasureNames,
		InlineMeasures:  req.InlineMeasures,
		TimeStart:       req.TimeStart,
		TimeEnd:         req.TimeEnd,
		Filter:          req.Filter,
	}
	err = s.runtime.Query(ctx, req.InstanceId, q, int(req.Priority))
	if err != nil {
		return nil, err
	}
	return q.Result, nil
}

// MetricsViewRows implements QueryService.
func (s *Server) MetricsViewRows(ctx context.Context, req *runtimev1.MetricsViewRowsRequest) (*runtimev1.MetricsViewRowsResponse, error) {
	observability.SetRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.metric_view", req.MetricsViewName),
		attribute.String("args.time_start", safeTimeStr(req.TimeStart)),
		attribute.String("args.time_end", safeTimeStr(req.TimeEnd)),
		attribute.Int("args.filter_count", filterCount(req.Filter)),
		attribute.StringSlice("args.sort.names", marshalMetricsViewSort(req.Sort)),
		attribute.Int("args.limit", int(req.Limit)),
		attribute.Int64("args.offset", req.Offset),
		attribute.Int("args.priority", int(req.Priority)),
	)

	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.ReadMetrics) {
		return nil, ErrForbidden
	}

	q := &queries.MetricsViewRows{
		MetricsViewName: req.MetricsViewName,
		TimeStart:       req.TimeStart,
		TimeEnd:         req.TimeEnd,
		Filter:          req.Filter,
		Sort:            req.Sort,
		Limit:           req.Limit,
		Offset:          req.Offset,
	}
	err := s.runtime.Query(ctx, req.InstanceId, q, int(req.Priority))
	if err != nil {
		return nil, err
	}

	return q.Result, nil
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
