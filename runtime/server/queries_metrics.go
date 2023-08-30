package server

import (
	"context"
	"fmt"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/queries"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MetricsViewToplist implements QueryService.
func (s *Server) MetricsViewToplist(ctx context.Context, req *runtimev1.MetricsViewToplistRequest) (*runtimev1.MetricsViewToplistResponse, error) {
	observability.AddRequestAttributes(ctx,
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

	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.ReadMetrics) {
		return nil, ErrForbidden
	}

	mv, policy, err := resolveMVAndPolicy(ctx, s.runtime, req.InstanceId, req.MetricsViewName)
	if err != nil {
		return nil, err
	}

	if !checkFieldAccess(req.DimensionName, policy) {
		return nil, ErrForbidden
	}

	// validate measures access
	for _, m := range req.MeasureNames {
		if !checkFieldAccess(m, policy) {
			return nil, ErrForbidden
		}
	}

	err = validateInlineMeasures(req.InlineMeasures)
	if err != nil {
		return nil, err
	}

	q := &queries.MetricsViewToplist{
		MetricsViewName:  req.MetricsViewName,
		DimensionName:    req.DimensionName,
		MeasureNames:     req.MeasureNames,
		InlineMeasures:   req.InlineMeasures,
		TimeStart:        req.TimeStart,
		TimeEnd:          req.TimeEnd,
		Limit:            &req.Limit,
		Offset:           req.Offset,
		Sort:             req.Sort,
		Filter:           req.Filter,
		MetricsView:      mv,
		ResolvedMVPolicy: policy,
	}
	err = s.runtime.Query(ctx, req.InstanceId, q, int(req.Priority))
	if err != nil {
		return nil, err
	}

	return q.Result, nil
}

// MetricsViewComparisonToplist implements QueryService.
func (s *Server) MetricsViewComparisonToplist(ctx context.Context, req *runtimev1.MetricsViewComparisonToplistRequest) (*runtimev1.MetricsViewComparisonToplistResponse, error) {
	observability.AddRequestAttributes(ctx,
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

	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	if req.BaseTimeRange != nil {
		observability.AddRequestAttributes(ctx, attribute.String("args.base_time_range.start", safeTimeStr(req.BaseTimeRange.Start)))
		observability.AddRequestAttributes(ctx, attribute.String("args.base_time_range.end", safeTimeStr(req.BaseTimeRange.End)))
	}
	if req.ComparisonTimeRange != nil {
		observability.AddRequestAttributes(ctx, attribute.String("args.comparison_time_range.start", safeTimeStr(req.ComparisonTimeRange.Start)))
		observability.AddRequestAttributes(ctx, attribute.String("args.comparison_time_range.end", safeTimeStr(req.ComparisonTimeRange.End)))
	}

	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.ReadMetrics) {
		return nil, ErrForbidden
	}

	mv, policy, err := resolveMVAndPolicy(ctx, s.runtime, req.InstanceId, req.MetricsViewName)
	if err != nil {
		return nil, err
	}

	if !checkFieldAccess(req.DimensionName, policy) {
		return nil, ErrForbidden
	}

	// validate measures access
	for _, m := range req.MeasureNames {
		if !checkFieldAccess(m, policy) {
			return nil, ErrForbidden
		}
	}

	err = validateInlineMeasures(req.InlineMeasures)
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
		MetricsView:         mv,
		ResolvedMVPolicy:    policy,
	}
	err = s.runtime.Query(ctx, req.InstanceId, q, int(req.Priority))
	if err != nil {
		return nil, err
	}

	return q.Result, nil
}

// MetricsViewTimeSeries implements QueryService.
func (s *Server) MetricsViewTimeSeries(ctx context.Context, req *runtimev1.MetricsViewTimeSeriesRequest) (*runtimev1.MetricsViewTimeSeriesResponse, error) {
	observability.AddRequestAttributes(ctx,
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

	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.ReadMetrics) {
		return nil, ErrForbidden
	}

	mv, policy, err := resolveMVAndPolicy(ctx, s.runtime, req.InstanceId, req.MetricsViewName)
	if err != nil {
		return nil, err
	}

	// validate measures access
	for _, m := range req.MeasureNames {
		if !checkFieldAccess(m, policy) {
			return nil, ErrForbidden
		}
	}

	err = validateInlineMeasures(req.InlineMeasures)
	if err != nil {
		return nil, err
	}

	q := &queries.MetricsViewTimeSeries{
		MetricsViewName:  req.MetricsViewName,
		MeasureNames:     req.MeasureNames,
		InlineMeasures:   req.InlineMeasures,
		TimeStart:        req.TimeStart,
		TimeEnd:          req.TimeEnd,
		TimeGranularity:  req.TimeGranularity,
		Filter:           req.Filter,
		TimeZone:         req.TimeZone,
		MetricsView:      mv,
		ResolvedMVPolicy: policy,
	}
	err = s.runtime.Query(ctx, req.InstanceId, q, int(req.Priority))
	if err != nil {
		return nil, err
	}
	return q.Result, nil
}

// MetricsViewTotals implements QueryService.
func (s *Server) MetricsViewTotals(ctx context.Context, req *runtimev1.MetricsViewTotalsRequest) (*runtimev1.MetricsViewTotalsResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.metric_view", req.MetricsViewName),
		attribute.StringSlice("args.measures", req.MeasureNames),
		attribute.StringSlice("args.inline_measures.names", marshalInlineMeasure(req.InlineMeasures)),
		attribute.String("args.time_start", safeTimeStr(req.TimeStart)),
		attribute.String("args.time_end", safeTimeStr(req.TimeEnd)),
		attribute.Int("args.filter_count", filterCount(req.Filter)),
		attribute.Int("args.priority", int(req.Priority)),
	)

	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.ReadMetrics) {
		return nil, ErrForbidden
	}

	mv, policy, err := resolveMVAndPolicy(ctx, s.runtime, req.InstanceId, req.MetricsViewName)
	if err != nil {
		return nil, err
	}

	// validate measures access
	for _, m := range req.MeasureNames {
		if !checkFieldAccess(m, policy) {
			return nil, ErrForbidden
		}
	}

	err = validateInlineMeasures(req.InlineMeasures)
	if err != nil {
		return nil, err
	}

	q := &queries.MetricsViewTotals{
		MetricsViewName:  req.MetricsViewName,
		MeasureNames:     req.MeasureNames,
		InlineMeasures:   req.InlineMeasures,
		TimeStart:        req.TimeStart,
		TimeEnd:          req.TimeEnd,
		Filter:           req.Filter,
		MetricsView:      mv,
		ResolvedMVPolicy: policy,
	}
	err = s.runtime.Query(ctx, req.InstanceId, q, int(req.Priority))
	if err != nil {
		return nil, err
	}
	return q.Result, nil
}

// MetricsViewRows implements QueryService.
func (s *Server) MetricsViewRows(ctx context.Context, req *runtimev1.MetricsViewRowsRequest) (*runtimev1.MetricsViewRowsResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.metric_view", req.MetricsViewName),
		attribute.String("args.time_start", safeTimeStr(req.TimeStart)),
		attribute.String("args.time_end", safeTimeStr(req.TimeEnd)),
		attribute.String("args.time_granularity", req.TimeGranularity.String()),
		attribute.Int("args.filter_count", filterCount(req.Filter)),
		attribute.StringSlice("args.sort.names", marshalMetricsViewSort(req.Sort)),
		attribute.Int("args.limit", int(req.Limit)),
		attribute.Int64("args.offset", req.Offset),
		attribute.Int("args.priority", int(req.Priority)),
	)

	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.ReadMetrics) {
		return nil, ErrForbidden
	}

	mv, policy, err := resolveMVAndPolicy(ctx, s.runtime, req.InstanceId, req.MetricsViewName)
	if err != nil {
		return nil, err
	}

	limit := int64(req.Limit)

	q := &queries.MetricsViewRows{
		MetricsViewName:  req.MetricsViewName,
		TimeStart:        req.TimeStart,
		TimeEnd:          req.TimeEnd,
		TimeGranularity:  req.TimeGranularity,
		Filter:           req.Filter,
		Sort:             req.Sort,
		Limit:            &limit,
		Offset:           req.Offset,
		TimeZone:         req.TimeZone,
		MetricsView:      mv,
		ResolvedMVPolicy: policy,
	}
	err = s.runtime.Query(ctx, req.InstanceId, q, int(req.Priority))
	if err != nil {
		return nil, err
	}

	return q.Result, nil
}

// MetricsViewTimeRange implements QueryService.
func (s *Server) MetricsViewTimeRange(ctx context.Context, req *runtimev1.MetricsViewTimeRangeRequest) (*runtimev1.MetricsViewTimeRangeResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.metric_view", req.MetricsViewName),
		attribute.Int("args.priority", int(req.Priority)),
	)

	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.ReadMetrics) {
		return nil, ErrForbidden
	}

	mv, policy, err := resolveMVAndPolicy(ctx, s.runtime, req.InstanceId, req.MetricsViewName)
	if err != nil {
		return nil, err
	}

	q := &queries.MetricsViewTimeRange{
		MetricsViewName:  req.MetricsViewName,
		MetricsView:      mv,
		ResolvedMVPolicy: policy,
	}
	err = s.runtime.Query(ctx, req.InstanceId, q, int(req.Priority))
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

func resolveMVAndPolicy(ctx context.Context, rt *runtime.Runtime, instanceID, metricsViewName string) (*runtimev1.MetricsView, *runtime.ResolvedMetricsViewPolicy, error) {
	mv, lastUpdatedOn, err := lookupMetricsView(ctx, rt, instanceID, metricsViewName)
	if err != nil {
		return nil, nil, err
	}

	resolvedPolicy, err := rt.ResolveMetricsViewPolicy(auth.GetClaims(ctx).Attributes(), instanceID, mv, lastUpdatedOn)
	if err != nil {
		return nil, nil, err
	}
	if resolvedPolicy != nil && !resolvedPolicy.HasAccess {
		return nil, nil, ErrForbidden
	}

	return mv, resolvedPolicy, nil
}

// returns the metrics view and the time the catalog was last updated
func lookupMetricsView(ctx context.Context, rt *runtime.Runtime, instanceID, name string) (*runtimev1.MetricsView, time.Time, error) {
	obj, err := rt.GetCatalogEntry(ctx, instanceID, name)
	if err != nil {
		return nil, time.Time{}, status.Error(codes.InvalidArgument, err.Error())
	}

	mv := obj.GetMetricsView()

	return mv, obj.UpdatedOn, nil
}

func checkFieldAccess(field string, policy *runtime.ResolvedMetricsViewPolicy) bool {
	if policy != nil {
		if !policy.HasAccess {
			return false
		}

		if len(policy.Include) > 0 {
			for _, include := range policy.Include {
				if include == field {
					return true
				}
			}
		} else if len(policy.Exclude) > 0 {
			for _, exclude := range policy.Exclude {
				if exclude == field {
					return false
				}
			}
		} else {
			// if no include/exclude is specified, then all fields are allowed
			return true
		}
	}
	return true
}
