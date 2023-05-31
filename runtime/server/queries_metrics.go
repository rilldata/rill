package server

import (
	"context"
	"fmt"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/queries"
	"github.com/rilldata/rill/runtime/server/auth"
)

// MetricsViewToplist implements QueryService.
func (s *Server) MetricsViewToplist(ctx context.Context, req *runtimev1.MetricsViewToplistRequest) (*runtimev1.MetricsViewToplistResponse, error) {
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
func ValidateInlineMeasures(ms []*runtimev1.InlineMeasure) error {
	for _, im := range ms {
		if !strings.EqualFold(im.Expression, "COUNT(*)") {
			return fmt.Errorf("illegal inline measure expression: %q", im.Expression)
		}
	}
	return nil
}
