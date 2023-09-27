package server

import (
	"context"
	"fmt"
	"regexp"
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

// MetricsViewAggregation implements QueryService.
func (s *Server) MetricsViewAggregation(ctx context.Context, req *runtimev1.MetricsViewAggregationRequest) (*runtimev1.MetricsViewAggregationResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.metric_view", req.MetricsView),
		attribute.StringSlice("args.dimensions.names", marshalMetricsViewAggregationDimension(req.Dimensions)),
		attribute.StringSlice("args.measures.names", marshalMetricsViewAggregationMeasures(req.Measures)),
		attribute.StringSlice("args.sort.names", marshalMetricsViewAggregationSort(req.Sort)),
		attribute.String("args.time_start", safeTimeStr(req.TimeStart)),
		attribute.String("args.time_end", safeTimeStr(req.TimeEnd)),
		attribute.Int("args.filter_count", filterCount(req.Filter)),
		attribute.Int64("args.limit", req.Limit),
		attribute.Int64("args.offset", req.Offset),
		attribute.Int("args.priority", int(req.Priority)),
	)
	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	if !auth.GetClaims(ctx).CanInstance(req.InstanceId, auth.ReadMetrics) {
		return nil, ErrForbidden
	}

	mv, security, err := resolveMVAndSecurity(ctx, s.runtime, req.InstanceId, req.MetricsView)
	if err != nil {
		return nil, err
	}

	for _, dim := range req.Dimensions {
		if dim.Name == mv.TimeDimension {
			// checkFieldAccess doesn't currently check the time dimension
			continue
		}
		if !checkFieldAccess(dim.Name, security) {
			return nil, ErrForbidden
		}
	}

	for _, m := range req.Measures {
		if m.BuiltinMeasure != runtimev1.BuiltinMeasure_BUILTIN_MEASURE_UNSPECIFIED {
			continue
		}
		if !checkFieldAccess(m.Name, security) {
			return nil, ErrForbidden
		}
	}

	q := &queries.MetricsViewAggregation{
		MetricsViewName:    req.MetricsView,
		Dimensions:         req.Dimensions,
		Measures:           req.Measures,
		Sort:               req.Sort,
		TimeStart:          req.TimeStart,
		TimeEnd:            req.TimeEnd,
		Filter:             req.Filter,
		Limit:              &req.Limit,
		Offset:             req.Offset,
		MetricsView:        mv,
		ResolvedMVSecurity: security,
	}
	err = s.runtime.Query(ctx, req.InstanceId, q, int(req.Priority))
	if err != nil {
		return nil, err
	}

	return q.Result, nil
}

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

	mv, security, err := resolveMVAndSecurity(ctx, s.runtime, req.InstanceId, req.MetricsViewName)
	if err != nil {
		return nil, err
	}

	if !checkFieldAccess(req.DimensionName, security) {
		return nil, ErrForbidden
	}

	// validate measures access
	for _, m := range req.MeasureNames {
		if !checkFieldAccess(m, security) {
			return nil, ErrForbidden
		}
	}

	err = validateInlineMeasures(req.InlineMeasures)
	if err != nil {
		return nil, err
	}

	q := &queries.MetricsViewToplist{
		MetricsViewName:    req.MetricsViewName,
		DimensionName:      req.DimensionName,
		MeasureNames:       req.MeasureNames,
		InlineMeasures:     req.InlineMeasures,
		TimeStart:          req.TimeStart,
		TimeEnd:            req.TimeEnd,
		Limit:              &req.Limit,
		Offset:             req.Offset,
		Sort:               req.Sort,
		Filter:             req.Filter,
		MetricsView:        mv,
		ResolvedMVSecurity: security,
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

	mv, security, err := resolveMVAndSecurity(ctx, s.runtime, req.InstanceId, req.MetricsViewName)
	if err != nil {
		return nil, err
	}

	if !checkFieldAccess(req.DimensionName, security) {
		return nil, ErrForbidden
	}

	// validate measures access
	for _, m := range req.MeasureNames {
		if !checkFieldAccess(m, security) {
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
		ResolvedMVSecurity:  security,
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

	mv, security, err := resolveMVAndSecurity(ctx, s.runtime, req.InstanceId, req.MetricsViewName)
	if err != nil {
		return nil, err
	}

	// validate measures access
	for _, m := range req.MeasureNames {
		if !checkFieldAccess(m, security) {
			return nil, ErrForbidden
		}
	}

	err = validateInlineMeasures(req.InlineMeasures)
	if err != nil {
		return nil, err
	}

	q := &queries.MetricsViewTimeSeries{
		MetricsViewName:    req.MetricsViewName,
		MeasureNames:       req.MeasureNames,
		InlineMeasures:     req.InlineMeasures,
		TimeStart:          req.TimeStart,
		TimeEnd:            req.TimeEnd,
		TimeGranularity:    req.TimeGranularity,
		Filter:             req.Filter,
		TimeZone:           req.TimeZone,
		MetricsView:        mv,
		ResolvedMVSecurity: security,
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

	mv, security, err := resolveMVAndSecurity(ctx, s.runtime, req.InstanceId, req.MetricsViewName)
	if err != nil {
		return nil, err
	}

	// validate measures access
	for _, m := range req.MeasureNames {
		if !checkFieldAccess(m, security) {
			return nil, ErrForbidden
		}
	}

	err = validateInlineMeasures(req.InlineMeasures)
	if err != nil {
		return nil, err
	}

	q := &queries.MetricsViewTotals{
		MetricsViewName:    req.MetricsViewName,
		MeasureNames:       req.MeasureNames,
		InlineMeasures:     req.InlineMeasures,
		TimeStart:          req.TimeStart,
		TimeEnd:            req.TimeEnd,
		Filter:             req.Filter,
		MetricsView:        mv,
		ResolvedMVSecurity: security,
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

	mv, security, err := resolveMVAndSecurity(ctx, s.runtime, req.InstanceId, req.MetricsViewName)
	if err != nil {
		return nil, err
	}

	limit := int64(req.Limit)

	q := &queries.MetricsViewRows{
		MetricsViewName:    req.MetricsViewName,
		TimeStart:          req.TimeStart,
		TimeEnd:            req.TimeEnd,
		TimeGranularity:    req.TimeGranularity,
		Filter:             req.Filter,
		Sort:               req.Sort,
		Limit:              &limit,
		Offset:             req.Offset,
		TimeZone:           req.TimeZone,
		MetricsView:        mv,
		ResolvedMVSecurity: security,
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

	mv, security, err := resolveMVAndSecurity(ctx, s.runtime, req.InstanceId, req.MetricsViewName)
	if err != nil {
		return nil, err
	}

	q := &queries.MetricsViewTimeRange{
		MetricsViewName:    req.MetricsViewName,
		MetricsView:        mv,
		ResolvedMVSecurity: security,
	}
	err = s.runtime.Query(ctx, req.InstanceId, q, int(req.Priority))
	if err != nil {
		return nil, err
	}

	return q.Result, nil
}

// inlineMeasureRegexp is used by validateInlineMeasures.
var inlineMeasureRegexp = regexp.MustCompile(`(?i)^COUNT\((DISTINCT)? *.+\)$`)

// validateInlineMeasures checks that the inline measures are allowed.
// This is to prevent injection of arbitrary SQL from clients with only ReadMetrics access.
// In the future, we should consider allowing arbitrary expressions from people with wider access.
// Currently, only COUNT(*) and COUNT(DISTINCT name) is allowed.
func validateInlineMeasures(ms []*runtimev1.InlineMeasure) error {
	for _, im := range ms {
		if !inlineMeasureRegexp.MatchString(im.Expression) {
			return fmt.Errorf("illegal inline measure expression: %q", im.Expression)
		}
	}
	return nil
}

func resolveMVAndSecurity(ctx context.Context, rt *runtime.Runtime, instanceID, metricsViewName string) (*runtimev1.MetricsView, *runtime.ResolvedMetricsViewSecurity, error) {
	mv, lastUpdatedOn, err := lookupMetricsView(ctx, rt, instanceID, metricsViewName)
	if err != nil {
		return nil, nil, err
	}

	resolvedSecurity, err := rt.ResolveMetricsViewSecurity(auth.GetClaims(ctx).Attributes(), instanceID, mv, lastUpdatedOn)
	if err != nil {
		return nil, nil, err
	}
	if resolvedSecurity != nil && !resolvedSecurity.Access {
		return nil, nil, ErrForbidden
	}

	return mv, resolvedSecurity, nil
}

func resolveMVAndSecurityFromAttributes(ctx context.Context, rt *runtime.Runtime, instanceID, metricsViewName string, attrs map[string]any) (*runtimev1.MetricsView, *runtime.ResolvedMetricsViewSecurity, error) {
	mv, lastUpdatedOn, err := lookupMetricsView(ctx, rt, instanceID, metricsViewName)
	if err != nil {
		return nil, nil, err
	}

	resolvedSecurity, err := rt.ResolveMetricsViewSecurity(attrs, instanceID, mv, lastUpdatedOn)
	if err != nil {
		return nil, nil, err
	}

	if resolvedSecurity != nil && !resolvedSecurity.Access {
		return nil, nil, ErrForbidden
	}

	return mv, resolvedSecurity, nil
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

func checkFieldAccess(field string, policy *runtime.ResolvedMetricsViewSecurity) bool {
	if policy != nil {
		if !policy.Access {
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
