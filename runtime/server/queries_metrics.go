package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/pkg/mapstructureutil"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/pkg/pbutil"
	"github.com/rilldata/rill/runtime/pkg/rilltime"
	"github.com/rilldata/rill/runtime/pkg/timeutil"
	"github.com/rilldata/rill/runtime/queries"
	"github.com/rilldata/rill/runtime/server/auth"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
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
		attribute.String("args.time_range", marshalTimeRange(req.TimeRange)),
		attribute.String("args.comparison_time_range", marshalTimeRange(req.ComparisonTimeRange)),
		attribute.Int("args.filter_count", filterCount(req.Where)),
		attribute.Int64("args.limit", req.Limit),
		attribute.Int64("args.offset", req.Offset),
		attribute.Int("args.priority", int(req.Priority)),
	)
	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	claims := auth.GetClaims(ctx, req.InstanceId)
	if !claims.Can(runtime.ReadMetrics) {
		return nil, ErrForbidden
	}

	tr := req.TimeRange
	if req.TimeStart != nil || req.TimeEnd != nil {
		tr = &runtimev1.TimeRange{
			Start: req.TimeStart,
			End:   req.TimeEnd,
		}
	}

	q := &queries.MetricsViewAggregation{
		MetricsViewName:     req.MetricsView,
		Dimensions:          req.Dimensions,
		Measures:            req.Measures,
		Sort:                req.Sort,
		TimeRange:           tr,
		ComparisonTimeRange: req.ComparisonTimeRange,
		Where:               req.Where,
		WhereSQL:            req.WhereSql,
		Having:              req.Having,
		HavingSQL:           req.HavingSql,
		Filter:              req.Filter,
		Limit:               &req.Limit,
		Offset:              req.Offset,
		PivotOn:             req.PivotOn,
		SecurityClaims:      claims,
		Exact:               req.Exact,
		Aliases:             req.Aliases,
		FillMissing:         req.FillMissing,
		Rows:                req.Rows,
	}
	err := s.runtime.Query(ctx, req.InstanceId, q, int(req.Priority))
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
		attribute.Int("args.filter_count", filterCount(req.Where)),
		attribute.StringSlice("args.sort.names", marshalMetricsViewSort(req.Sort)),
	)

	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	claims := auth.GetClaims(ctx, req.InstanceId)
	if !claims.Can(runtime.ReadMetrics) {
		return nil, ErrForbidden
	}

	if req.Limit == 0 {
		req.Limit = 100
	}

	q := &queries.MetricsViewToplist{
		MetricsViewName: req.MetricsViewName,
		DimensionName:   req.DimensionName,
		MeasureNames:    req.MeasureNames,
		TimeStart:       req.TimeStart,
		TimeEnd:         req.TimeEnd,
		Limit:           &req.Limit,
		Offset:          req.Offset,
		Sort:            req.Sort,
		Where:           req.Where,
		WhereSQL:        req.WhereSql,
		Having:          req.Having,
		HavingSQL:       req.HavingSql,
		Filter:          req.Filter,
		SecurityClaims:  claims,
	}
	err := s.runtime.Query(ctx, req.InstanceId, q, int(req.Priority))
	if err != nil {
		return nil, err
	}

	return q.Result, nil
}

// MetricsViewComparison implements QueryService.
func (s *Server) MetricsViewComparison(ctx context.Context, req *runtimev1.MetricsViewComparisonRequest) (*runtimev1.MetricsViewComparisonResponse, error) {
	measureNames := make([]string, 0, len(req.Measures))
	for _, m := range req.Measures {
		measureNames = append(measureNames, m.Name)
	}
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.metric_view", req.MetricsViewName),
		attribute.String("args.dimension", req.Dimension.Name),
		attribute.StringSlice("args.measures", measureNames),
		attribute.StringSlice("args.comparison_measures", req.ComparisonMeasures),
		attribute.StringSlice("args.sort.names", marshalMetricsViewComparisonSort(req.Sort)),
		attribute.Int("args.filter_count", filterCount(req.Where)),
		attribute.Int64("args.limit", req.Limit),
		attribute.Int64("args.offset", req.Offset),
		attribute.Int("args.priority", int(req.Priority)),
	)

	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	if req.TimeRange != nil {
		observability.AddRequestAttributes(ctx, attribute.String("args.base_time_range.start", safeTimeStr(req.TimeRange.Start)))
		observability.AddRequestAttributes(ctx, attribute.String("args.base_time_range.end", safeTimeStr(req.TimeRange.End)))
	}
	if req.ComparisonTimeRange != nil {
		observability.AddRequestAttributes(ctx, attribute.String("args.comparison_time_range.start", safeTimeStr(req.ComparisonTimeRange.Start)))
		observability.AddRequestAttributes(ctx, attribute.String("args.comparison_time_range.end", safeTimeStr(req.ComparisonTimeRange.End)))
	}

	claims := auth.GetClaims(ctx, req.InstanceId)
	if !claims.Can(runtime.ReadMetrics) {
		return nil, ErrForbidden
	}

	if req.Limit == 0 {
		req.Limit = 100
	}

	q := &queries.MetricsViewComparison{
		MetricsViewName:     req.MetricsViewName,
		DimensionName:       req.Dimension.Name,
		Measures:            req.Measures,
		ComparisonMeasures:  req.ComparisonMeasures,
		TimeRange:           req.TimeRange,
		ComparisonTimeRange: req.ComparisonTimeRange,
		Limit:               req.Limit,
		Offset:              req.Offset,
		Sort:                req.Sort,
		Where:               req.Where,
		WhereSQL:            req.WhereSql,
		Having:              req.Having,
		HavingSQL:           req.HavingSql,
		Exact:               req.Exact,
		Filter:              req.Filter,
		SecurityClaims:      claims,
	}
	err := s.runtime.Query(ctx, req.InstanceId, q, int(req.Priority))
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
		attribute.String("args.time_start", safeTimeStr(req.TimeStart)),
		attribute.String("args.time_end", safeTimeStr(req.TimeEnd)),
		attribute.String("args.time_granularity", req.TimeGranularity.String()),
		attribute.Int("args.filter_count", filterCount(req.Where)),
		attribute.Int("args.priority", int(req.Priority)),
		attribute.String("args.time_dimension", req.TimeDimension),
	)

	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	claims := auth.GetClaims(ctx, req.InstanceId)
	if !claims.Can(runtime.ReadMetrics) {
		return nil, ErrForbidden
	}

	q := &queries.MetricsViewTimeSeries{
		MetricsViewName: req.MetricsViewName,
		MeasureNames:    req.MeasureNames,
		TimeStart:       req.TimeStart,
		TimeEnd:         req.TimeEnd,
		TimeGranularity: req.TimeGranularity,
		Where:           req.Where,
		WhereSQL:        req.WhereSql,
		Having:          req.Having,
		HavingSQL:       req.HavingSql,
		TimeZone:        req.TimeZone,
		Filter:          req.Filter,
		SecurityClaims:  claims,
		TimeDimension:   req.TimeDimension,
	}
	err := s.runtime.Query(ctx, req.InstanceId, q, int(req.Priority))
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
		attribute.String("args.time_start", safeTimeStr(req.TimeStart)),
		attribute.String("args.time_end", safeTimeStr(req.TimeEnd)),
		attribute.Int("args.filter_count", filterCount(req.Where)),
		attribute.Int("args.priority", int(req.Priority)),
		attribute.String("args.time_dimension", req.TimeDimension),
	)

	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	claims := auth.GetClaims(ctx, req.InstanceId)
	if !claims.Can(runtime.ReadMetrics) {
		return nil, ErrForbidden
	}

	q := &queries.MetricsViewTotals{
		MetricsViewName: req.MetricsViewName,
		MeasureNames:    req.MeasureNames,
		TimeStart:       req.TimeStart,
		TimeEnd:         req.TimeEnd,
		Where:           req.Where,
		WhereSQL:        req.WhereSql,
		Filter:          req.Filter,
		SecurityClaims:  claims,
		TimeDimension:   req.TimeDimension,
	}
	err := s.runtime.Query(ctx, req.InstanceId, q, int(req.Priority))
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
		attribute.Int("args.filter_count", filterCount(req.Where)),
		attribute.StringSlice("args.sort.names", marshalMetricsViewSort(req.Sort)),
		attribute.Int("args.limit", int(req.Limit)),
		attribute.Int64("args.offset", req.Offset),
		attribute.Int("args.priority", int(req.Priority)),
		attribute.String("args.time_dimension", req.TimeDimension),
	)

	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	claims := auth.GetClaims(ctx, req.InstanceId)
	if !claims.Can(runtime.ReadMetrics) {
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
		Where:              req.Where,
		Sort:               req.Sort,
		Limit:              &limit,
		Offset:             req.Offset,
		TimeZone:           req.TimeZone,
		MetricsView:        mv.ValidSpec,
		ResolvedMVSecurity: security,
		Streaming:          mv.Streaming,
		Filter:             req.Filter,
		TimeDimension:      req.TimeDimension,
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
		attribute.String("args.time_dimension", req.TimeDimension),
	)

	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	claims := auth.GetClaims(ctx, req.InstanceId)
	if !claims.Can(runtime.ReadMetrics) {
		return nil, ErrForbidden
	}

	ts, err := queries.ResolveTimestampResult(ctx, s.runtime, req.InstanceId, req.MetricsViewName, req.TimeDimension, claims, int(req.Priority))
	if err != nil {
		return nil, err
	}

	return &runtimev1.MetricsViewTimeRangeResponse{
		TimeRangeSummary: &runtimev1.TimeRangeSummary{
			// JS identifies no time present using a null check instead of `IsZero`
			// So send null when time.IsZero
			Min:       valOrNullTime(ts.Min),
			Max:       valOrNullTime(ts.Max),
			Watermark: valOrNullTime(ts.Watermark),
		},
	}, nil
}

func (s *Server) MetricsViewSchema(ctx context.Context, req *runtimev1.MetricsViewSchemaRequest) (*runtimev1.MetricsViewSchemaResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.metric_view", req.MetricsViewName),
		attribute.Int("args.priority", int(req.Priority)),
	)

	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	claims := auth.GetClaims(ctx, req.InstanceId)
	if !claims.Can(runtime.ReadMetrics) {
		return nil, ErrForbidden
	}

	q := &queries.MetricsViewSchema{
		MetricsViewName: req.MetricsViewName,
		SecurityClaims:  claims,
	}
	err := s.runtime.Query(ctx, req.InstanceId, q, int(req.Priority))
	if err != nil {
		return nil, err
	}

	return q.Result, nil
}

func (s *Server) MetricsViewSearch(ctx context.Context, req *runtimev1.MetricsViewSearchRequest) (*runtimev1.MetricsViewSearchResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.metric_view", req.MetricsViewName),
		attribute.StringSlice("args.dimensions.names", req.Dimensions),
		attribute.String("args.search", req.Search),
		attribute.String("args.time_range", marshalTimeRange(req.TimeRange)),
		attribute.Int("args.filter_count", filterCount(req.Where)),
		attribute.Int("args.priority", int(req.Priority)),
	)

	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	claims := auth.GetClaims(ctx, req.InstanceId)
	if !claims.Can(runtime.ReadMetrics) {
		return nil, ErrForbidden
	}

	limit := int64(req.Limit)
	q := &queries.MetricsViewSearch{
		MetricsViewName: req.MetricsViewName,
		Dimensions:      req.Dimensions,
		Search:          req.Search,
		TimeRange:       req.TimeRange,
		Where:           req.Where,
		Having:          req.Having,
		Priority:        req.Priority,
		Limit:           &limit,
		SecurityClaims:  claims,
	}
	err := s.runtime.Query(ctx, req.InstanceId, q, int(req.Priority))
	if err != nil {
		return nil, err
	}

	return q.Result, nil
}

func (s *Server) MetricsViewTimeRanges(ctx context.Context, req *runtimev1.MetricsViewTimeRangesRequest) (*runtimev1.MetricsViewTimeRangesResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.metric_view", req.MetricsViewName),
		attribute.StringSlice("args.expressions", req.Expressions),
		attribute.String("args.time_dimension", req.TimeDimension),
	)
	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	claims := auth.GetClaims(ctx, req.InstanceId)
	if !claims.Can(runtime.ReadMetrics) {
		return nil, ErrForbidden
	}

	mv, _, err := resolveMVAndSecurity(ctx, s.runtime, req.InstanceId, req.MetricsViewName)
	if err != nil {
		return nil, err
	}

	// to keep results consistent
	now := time.Now()

	ts, err := queries.ResolveTimestampResult(ctx, s.runtime, req.InstanceId, req.MetricsViewName, req.TimeDimension, claims, int(req.Priority))
	if err != nil {
		return nil, err
	}
	if req.ExecutionTime != nil {
		// If override is set, we use it for every ref except `min`
		ts = metricsview.TimestampsResult{
			Min:       ts.Min,
			Max:       req.ExecutionTime.AsTime(),
			Watermark: req.ExecutionTime.AsTime(),
		}
		now = req.ExecutionTime.AsTime()
	}

	timeDimension := req.TimeDimension
	if timeDimension == "" && mv.ValidSpec != nil {
		timeDimension = mv.ValidSpec.TimeDimension
	}

	var tz *time.Location
	if req.TimeZone != "" {
		tz, err = time.LoadLocation(req.TimeZone)
		if err != nil {
			return nil, err
		}
	}

	timeRanges := make([]*runtimev1.ResolvedTimeRange, len(req.Expressions))
	for i, tr := range req.Expressions {
		expr, err := rilltime.Parse(tr, rilltime.ParseOptions{
			TimeZoneOverride: tz,
			SmallestGrain:    timeutil.TimeGrainFromAPI(mv.ValidSpec.SmallestTimeGrain),
		})
		if err != nil {
			return nil, fmt.Errorf("error parsing time range %s: %w", tr, err)
		}

		start, end, grain := expr.Eval(rilltime.EvalOptions{
			Now:        now,
			MinTime:    ts.Min,
			MaxTime:    ts.Max,
			Watermark:  ts.Watermark,
			FirstDay:   int(mv.ValidSpec.FirstDayOfWeek),
			FirstMonth: int(mv.ValidSpec.FirstMonthOfYear),
		})

		timeRanges[i] = &runtimev1.ResolvedTimeRange{
			Start:         timestamppb.New(start),
			End:           timestamppb.New(end),
			Grain:         timeutil.TimeGrainToAPI(grain),
			TimeDimension: timeDimension,
			TimeZone:      req.TimeZone,
			Expression:    tr,
		}
	}

	backwardsCompatibleRanges := make([]*runtimev1.TimeRange, len(req.Expressions))
	for i, r := range timeRanges {
		backwardsCompatibleRanges[i] = &runtimev1.TimeRange{
			Start:         r.Start,
			End:           r.End,
			TimeDimension: r.TimeDimension,
			IsoDuration:   "",
			IsoOffset:     "",
			Expression:    r.Expression,
			RoundToGrain:  r.Grain,
		}
	}

	return &runtimev1.MetricsViewTimeRangesResponse{
		FullTimeRange: &runtimev1.TimeRangeSummary{
			Min:       valOrNullTime(ts.Min),
			Max:       valOrNullTime(ts.Max),
			Watermark: valOrNullTime(ts.Watermark),
		},
		ResolvedTimeRanges: timeRanges,
		TimeRanges:         backwardsCompatibleRanges,
	}, nil
}

func (s *Server) MetricsViewAnnotations(ctx context.Context, req *runtimev1.MetricsViewAnnotationsRequest) (*runtimev1.MetricsViewAnnotationsResponse, error) {
	observability.AddRequestAttributes(ctx,
		attribute.String("args.instance_id", req.InstanceId),
		attribute.String("args.metric_view", req.MetricsViewName),
		attribute.String("args.measures", strings.Join(req.Measures, ",")),
	)

	claims := auth.GetClaims(ctx, req.InstanceId)
	if !claims.Can(runtime.ReadMetrics) {
		return nil, ErrForbidden
	}
	s.addInstanceRequestAttributes(ctx, req.InstanceId)

	var limit *int64
	if req.Limit != 0 {
		limit = &req.Limit
	}

	var offset *int64
	if req.Offset != 0 {
		limit = &req.Offset
	}

	qry := metricsview.AnnotationsQuery{
		MetricsView: req.MetricsViewName,
		Measures:    req.Measures,
		Limit:       limit,
		Offset:      offset,
		TimeZone:    req.TimeZone,
		TimeGrain:   metricsview.TimeGrainFromProto(req.TimeGrain),
		Priority:    int(req.Priority),
	}

	if req.TimeRange != nil {
		qry.TimeRange = &metricsview.TimeRange{
			Start:         req.TimeRange.Start.AsTime(),
			End:           req.TimeRange.End.AsTime(),
			Expression:    req.TimeRange.Expression,
			IsoDuration:   req.TimeRange.IsoDuration,
			IsoOffset:     req.TimeRange.IsoOffset,
			RoundToGrain:  metricsview.TimeGrainFromProto(req.TimeRange.RoundToGrain),
			TimeDimension: req.TimeRange.TimeDimension,
		}
	}

	props, err := qry.AsMap()
	if err != nil {
		return nil, err
	}

	res, err := s.runtime.Resolve(ctx, &runtime.ResolveOptions{
		InstanceID:         req.InstanceId,
		Resolver:           "metrics_annotations",
		ResolverProperties: props,
		Claims:             claims,
	})
	if err != nil {
		return nil, err
	}
	defer res.Close()

	rows := make([]*runtimev1.MetricsViewAnnotationsResponse_Annotation, 0)
	for {
		row, err := res.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		var ann annotation
		err = mapstructureutil.WeakDecode(row, &ann)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		additionalFieldsPb, err := pbutil.ToStruct(ann.AdditionalFields, res.Schema())
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		var timeEnd *timestamppb.Timestamp
		if !ann.TimeEnd.IsZero() {
			timeEnd = timestamppb.New(ann.TimeEnd)
		}

		var dur *string
		if ann.Duration != "" {
			dur = &ann.Duration
		}

		rows = append(rows, &runtimev1.MetricsViewAnnotationsResponse_Annotation{
			Time:             timestamppb.New(ann.Time),
			TimeEnd:          timeEnd,
			Description:      ann.Description,
			Duration:         dur,
			ForMeasures:      ann.ForMeasures,
			AdditionalFields: additionalFieldsPb,
		})
	}

	return &runtimev1.MetricsViewAnnotationsResponse{
		Rows: rows,
	}, nil
}

type annotation struct {
	Time             time.Time      `mapstructure:"time"`
	TimeEnd          time.Time      `mapstructure:"time_end"`
	Description      string         `mapstructure:"description"`
	Duration         string         `mapstructure:"duration"`
	ForMeasures      []string       `mapstructure:"for_measures"`
	AdditionalFields map[string]any `mapstructure:",remain"`
}

func resolveMVAndSecurity(ctx context.Context, rt *runtime.Runtime, instanceID, metricsViewName string) (*runtimev1.MetricsViewState, *runtime.ResolvedSecurity, error) {
	res, mv, err := lookupMetricsView(ctx, rt, instanceID, metricsViewName)
	if err != nil {
		return nil, nil, err
	}

	resolvedSecurity, err := rt.ResolveSecurity(ctx, instanceID, auth.GetClaims(ctx, instanceID), res)
	if err != nil {
		return nil, nil, err
	}
	if !resolvedSecurity.CanAccess() {
		return nil, nil, ErrForbidden
	}

	return mv, resolvedSecurity, nil
}

func resolveMVAndSecurityFromAttributes(ctx context.Context, rt *runtime.Runtime, instanceID, metricsViewName string, claims *runtime.SecurityClaims) (*runtimev1.MetricsViewState, *runtime.ResolvedSecurity, error) {
	res, mv, err := lookupMetricsView(ctx, rt, instanceID, metricsViewName)
	if err != nil {
		return nil, nil, err
	}

	resolvedSecurity, err := rt.ResolveSecurity(ctx, instanceID, claims, res)
	if err != nil {
		return nil, nil, err
	}

	if !resolvedSecurity.CanAccess() {
		return nil, nil, ErrForbidden
	}

	return mv, resolvedSecurity, nil
}

// returns the metrics view and the time the catalog was last updated
func lookupMetricsView(ctx context.Context, rt *runtime.Runtime, instanceID, name string) (*runtimev1.Resource, *runtimev1.MetricsViewState, error) {
	ctrl, err := rt.Controller(ctx, instanceID)
	if err != nil {
		return nil, nil, status.Error(codes.InvalidArgument, err.Error())
	}

	res, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: name}, false)
	if err != nil {
		return nil, nil, status.Error(codes.InvalidArgument, err.Error())
	}

	mv := res.GetMetricsView()
	spec := mv.State.ValidSpec
	if spec == nil {
		return nil, nil, status.Errorf(codes.InvalidArgument, "metrics view %q is invalid", name)
	}

	return res, mv.State, nil
}

func valOrNullTime(v time.Time) *timestamppb.Timestamp {
	if v.IsZero() {
		return nil
	}
	return timestamppb.New(v)
}
