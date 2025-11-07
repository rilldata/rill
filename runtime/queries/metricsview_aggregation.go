package queries

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/metricsview/executor"
	"github.com/rilldata/rill/runtime/metricsview/metricssql"
)

type MetricsViewAggregation struct {
	MetricsViewName     string                                         `json:"metrics_view,omitempty"`
	Dimensions          []*runtimev1.MetricsViewAggregationDimension   `json:"dimensions,omitempty"`
	Measures            []*runtimev1.MetricsViewAggregationMeasure     `json:"measures,omitempty"`
	Sort                []*runtimev1.MetricsViewAggregationSort        `json:"sort,omitempty"`
	TimeRange           *runtimev1.TimeRange                           `json:"time_range,omitempty"`
	ComparisonTimeRange *runtimev1.TimeRange                           `json:"comparison_time_range,omitempty"`
	Where               *runtimev1.Expression                          `json:"where,omitempty"`
	WhereSQL            string                                         `json:"where_sql,omitempty"`
	Having              *runtimev1.Expression                          `json:"having,omitempty"`
	HavingSQL           string                                         `json:"having_sql,omitempty"`
	Filter              *runtimev1.MetricsViewFilter                   `json:"filter,omitempty"` // Backwards compatibility
	Priority            int32                                          `json:"priority,omitempty"`
	Limit               *int64                                         `json:"limit,omitempty"`
	Offset              int64                                          `json:"offset,omitempty"`
	PivotOn             []string                                       `json:"pivot_on,omitempty"`
	SecurityClaims      *runtime.SecurityClaims                        `json:"security_claims,omitempty"`
	Aliases             []*runtimev1.MetricsViewComparisonMeasureAlias `json:"aliases,omitempty"`
	Exact               bool                                           `json:"exact,omitempty"`
	FillMissing         bool                                           `json:"fill_missing,omitempty"`
	Rows                bool                                           `json:"rows,omitempty"`
	ExecutionTime       *time.Time                                     `json:"execution_time,omitempty"`

	Result    *runtimev1.MetricsViewAggregationResponse `json:"-"`
	Exporting bool                                      `json:"-"` // Deprecated: Remove when tests call Export directly
}

var _ runtime.Query = &MetricsViewAggregation{}

func (q *MetricsViewAggregation) Key() string {
	r, err := json.Marshal(q)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("MetricsViewAggregation:%s", string(r))
}

func (q *MetricsViewAggregation) Deps() []*runtimev1.ResourceName {
	return []*runtimev1.ResourceName{
		{Kind: runtime.ResourceKindMetricsView, Name: q.MetricsViewName},
	}
}

func (q *MetricsViewAggregation) MarshalResult() *runtime.QueryResult {
	return &runtime.QueryResult{
		Value: q.Result,
		Bytes: sizeProtoMessage(q.Result),
	}
}

func (q *MetricsViewAggregation) UnmarshalResult(v any) error {
	res, ok := v.(*runtimev1.MetricsViewAggregationResponse)
	if !ok {
		return fmt.Errorf("MetricsViewAggregation: mismatched unmarshal input")
	}
	q.Result = res
	return nil
}

func (q *MetricsViewAggregation) Resolve(ctx context.Context, rt *runtime.Runtime, instanceID string, priority int) error {
	mv, security, err := resolveMVAndSecurityFromAttributes(ctx, rt, instanceID, q.MetricsViewName, q.SecurityClaims)
	if err != nil {
		return err
	}

	qry, err := q.rewriteToMetricsViewQuery(q.Exporting)
	if err != nil {
		return fmt.Errorf("error rewriting to metrics query: %w", err)
	}

	e, err := executor.New(ctx, rt, instanceID, mv.ValidSpec, mv.Streaming, security, priority)
	if err != nil {
		return err
	}
	defer e.Close()

	res, err := e.Query(ctx, qry, q.ExecutionTime)
	if err != nil {
		return err
	}
	defer res.Close()

	data, err := rowsToData(res)
	if err != nil {
		return err
	}

	q.Result = &runtimev1.MetricsViewAggregationResponse{
		Schema: res.Schema,
		Data:   data,
	}
	return nil
}

func (q *MetricsViewAggregation) Export(ctx context.Context, rt *runtime.Runtime, instanceID string, w io.Writer, opts *runtime.ExportOptions) error {
	filename := strings.ReplaceAll(q.MetricsViewName, `"`, `_`)
	if !isTimeRangeNil(q.TimeRange) || q.Where != nil || q.Having != nil || q.WhereSQL != "" || q.HavingSQL != "" {
		filename += "_filtered"
	}

	// Resolve metrics view
	mv, security, err := resolveMVAndSecurityFromAttributes(ctx, rt, instanceID, q.MetricsViewName, q.SecurityClaims)
	if err != nil {
		return err
	}

	// Route to metricsview executor
	qry, err := q.rewriteToMetricsViewQuery(true)
	if err != nil {
		return fmt.Errorf("error rewriting to metrics query: %w", err)
	}

	e, err := executor.New(ctx, rt, instanceID, mv.ValidSpec, mv.Streaming, security, opts.Priority)
	if err != nil {
		return err
	}
	defer e.Close()

	if mv.ValidSpec.TimeDimension != "" {
		tsRes, err := ResolveTimestampResult(ctx, rt, instanceID, q.MetricsViewName, q.TimeRange.TimeDimension, q.SecurityClaims, opts.Priority)
		if err != nil {
			return err
		}

		err = e.BindQuery(ctx, qry, tsRes)
		if err != nil {
			return err
		}
	}

	var format drivers.FileFormat
	switch opts.Format {
	case runtimev1.ExportFormat_EXPORT_FORMAT_CSV:
		format = drivers.FileFormatCSV
	case runtimev1.ExportFormat_EXPORT_FORMAT_XLSX:
		format = drivers.FileFormatXLSX
	case runtimev1.ExportFormat_EXPORT_FORMAT_PARQUET:
		format = drivers.FileFormatParquet
	default:
		return fmt.Errorf("unsupported format: %s", opts.Format.String())
	}

	headers, err := q.generateExportHeaders(ctx, rt, instanceID, opts, qry)
	if err != nil {
		return err
	}

	path, err := e.Export(ctx, qry, q.ExecutionTime, format, headers)
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(path) }()

	err = opts.PreWriteHook(filename)
	if err != nil {
		return err
	}

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(w, f)
	if err != nil {
		return err
	}

	return nil
}

// ResolveTimestampResult resolves the time range for a metrics view and returns the min, max, and watermark timestamps.
// timeDimension is optional and can be used to specify which time dimension to use for the time range query otherwise it will use the default time dimension of the metrics view.
func ResolveTimestampResult(ctx context.Context, rt *runtime.Runtime, instanceID, metricsViewName, timeDimension string, security *runtime.SecurityClaims, priority int) (metricsview.TimestampsResult, error) {
	res, err := rt.Resolve(ctx, &runtime.ResolveOptions{
		InstanceID: instanceID,
		Resolver:   "metrics_time_range",
		ResolverProperties: map[string]any{
			"metrics_view": metricsViewName,
		},
		Args: map[string]any{
			"priority":       priority,
			"time_dimension": timeDimension,
		},
		Claims: security,
	})
	if err != nil {
		return metricsview.TimestampsResult{}, err
	}
	defer res.Close()

	row, err := res.Next()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return metricsview.TimestampsResult{}, errors.New("time range query returned no results")
		}
		return metricsview.TimestampsResult{}, err
	}

	tsRes := metricsview.TimestampsResult{}

	tsRes.Min, err = anyToTime(row["min"])
	if err != nil {
		return tsRes, err
	}
	tsRes.Max, err = anyToTime(row["max"])
	if err != nil {
		return tsRes, err
	}
	tsRes.Watermark, err = anyToTime(row["watermark"])
	if err != nil {
		return tsRes, err
	}

	return tsRes, nil
}

func (q *MetricsViewAggregation) rewriteToMetricsViewQuery(export bool) (*metricsview.Query, error) {
	qry := &metricsview.Query{MetricsView: q.MetricsViewName}

	for _, d := range q.Dimensions {
		res := metricsview.Dimension{Name: d.Name}
		if d.Alias != "" {
			res.Name = d.Alias
		}
		if d.TimeZone != "" {
			qry.TimeZone = d.TimeZone
		}
		if d.TimeGrain != runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED {
			res.Compute = &metricsview.DimensionCompute{
				TimeFloor: &metricsview.DimensionComputeTimeFloor{
					Dimension: d.Name,
					Grain:     metricsview.TimeGrainFromProto(d.TimeGrain),
				},
			}
		}
		qry.Dimensions = append(qry.Dimensions, res)
	}

	var measureFilter *runtimev1.Expression

	for _, m := range q.Measures {
		res := metricsview.Measure{Name: m.Name}
		switch m.BuiltinMeasure {
		case runtimev1.BuiltinMeasure_BUILTIN_MEASURE_COUNT:
			res.Compute = &metricsview.MeasureCompute{Count: true}
		case runtimev1.BuiltinMeasure_BUILTIN_MEASURE_COUNT_DISTINCT:
			res.Compute = &metricsview.MeasureCompute{CountDistinct: &metricsview.MeasureComputeCountDistinct{
				Dimension: m.BuiltinMeasureArgs[0].GetStringValue(),
			}}
		}

		if m.Filter != nil {
			if len(q.Measures) > 1 {
				return nil, fmt.Errorf("measure-level filter is not supported when multiple measures are present")
			}
			measureFilter = m.Filter
		}

		if m.Compute != nil {
			switch c := m.Compute.(type) {
			case *runtimev1.MetricsViewAggregationMeasure_ComparisonValue:
				res.Compute = &metricsview.MeasureCompute{ComparisonValue: &metricsview.MeasureComputeComparisonValue{
					Measure: c.ComparisonValue.Measure,
				}}
			case *runtimev1.MetricsViewAggregationMeasure_ComparisonDelta:
				res.Compute = &metricsview.MeasureCompute{ComparisonDelta: &metricsview.MeasureComputeComparisonDelta{
					Measure: c.ComparisonDelta.Measure,
				}}
			case *runtimev1.MetricsViewAggregationMeasure_ComparisonRatio:
				res.Compute = &metricsview.MeasureCompute{ComparisonRatio: &metricsview.MeasureComputeComparisonRatio{
					Measure: c.ComparisonRatio.Measure,
				}}
			case *runtimev1.MetricsViewAggregationMeasure_PercentOfTotal:
				res.Compute = &metricsview.MeasureCompute{PercentOfTotal: &metricsview.MeasureComputePercentOfTotal{
					Measure: c.PercentOfTotal.Measure,
				}}
			case *runtimev1.MetricsViewAggregationMeasure_Uri:
				res.Compute = &metricsview.MeasureCompute{URI: &metricsview.MeasureComputeURI{
					Dimension: c.Uri.Dimension,
				}}
			case *runtimev1.MetricsViewAggregationMeasure_ComparisonTime:
				res.Compute = &metricsview.MeasureCompute{ComparisonTime: &metricsview.MeasureComputeComparisonTime{
					Dimension: c.ComparisonTime.Dimension,
				}}
			}
		}

		qry.Measures = append(qry.Measures, res)
	}

	qry.PivotOn = q.PivotOn

	for _, s := range q.Sort {
		qry.Sort = append(qry.Sort, metricsview.Sort{
			Name: s.Name,
			Desc: s.Desc,
		})
	}

	if q.TimeRange != nil {
		res := &metricsview.TimeRange{}
		if q.TimeRange.Start != nil {
			res.Start = q.TimeRange.Start.AsTime()
		}
		if q.TimeRange.End != nil {
			res.End = q.TimeRange.End.AsTime()
		}
		res.Expression = q.TimeRange.Expression
		res.IsoDuration = q.TimeRange.IsoDuration
		res.IsoOffset = q.TimeRange.IsoOffset
		res.RoundToGrain = metricsview.TimeGrainFromProto(q.TimeRange.RoundToGrain)
		if q.TimeRange.TimeZone != "" {
			qry.TimeZone = q.TimeRange.TimeZone
		}
		res.TimeDimension = q.TimeRange.TimeDimension
		qry.TimeRange = res
	}

	if q.ComparisonTimeRange != nil {
		res := &metricsview.TimeRange{}
		if q.ComparisonTimeRange.Start != nil {
			res.Start = q.ComparisonTimeRange.Start.AsTime()
		}
		if q.ComparisonTimeRange.End != nil {
			res.End = q.ComparisonTimeRange.End.AsTime()
		}
		res.Expression = q.ComparisonTimeRange.Expression
		res.IsoDuration = q.ComparisonTimeRange.IsoDuration
		res.IsoOffset = q.ComparisonTimeRange.IsoOffset
		res.RoundToGrain = metricsview.TimeGrainFromProto(q.ComparisonTimeRange.RoundToGrain)
		if q.ComparisonTimeRange.TimeZone != "" {
			if qry.TimeZone != "" && qry.TimeZone != q.ComparisonTimeRange.TimeZone {
				return nil, fmt.Errorf("comparison_time_range has a different time zone")
			}
			qry.TimeZone = q.ComparisonTimeRange.TimeZone
		}
		res.TimeDimension = q.ComparisonTimeRange.TimeDimension
		qry.ComparisonTimeRange = res
	}

	if q.Filter != nil { // Backwards compatibility
		if q.Where != nil {
			return nil, fmt.Errorf("both filter and where is provided")
		}
		q.Where = convertFilterToExpression(q.Filter)
	}

	var err error
	qry.Where, err = metricViewExpression(q.Where, q.WhereSQL)
	if err != nil {
		return nil, err
	}

	// If a measure-level filter is present, we set qry.Where as the spine, and use (qry.Where AND measuresFilter) as the new where clause
	if measureFilter != nil {
		qry.Spine = &metricsview.Spine{Where: &metricsview.WhereSpine{Expression: qry.Where}}

		if qry.Where == nil {
			qry.Where = metricsview.NewExpressionFromProto(measureFilter)
		} else {
			qry.Where = &metricsview.Expression{
				Condition: &metricsview.Condition{
					Operator: metricsview.OperatorAnd,
					Expressions: []*metricsview.Expression{
						qry.Where,
						metricsview.NewExpressionFromProto(measureFilter),
					},
				},
			}
		}
	}

	// If there is only one time dimension and null fill is enabled, we set the spine to the time range
	if q.FillMissing {
		if qry.Spine != nil {
			// should we silently ignore instead of error ?
			return nil, fmt.Errorf("cannot have both where and time spine")
		}
		if (q.TimeRange == nil) || ((q.TimeRange.Start == nil || q.TimeRange.End == nil) && (q.TimeRange.IsoDuration == "")) {
			return nil, fmt.Errorf("time range is required for null fill")
		}

		// this will be resolved later in executor_rewrite_time.go after the time range is resolved
		qry.Spine = &metricsview.Spine{}
		qry.Spine.TimeRange = &metricsview.TimeSpine{}
	}

	qry.Having, err = metricViewExpression(q.Having, q.HavingSQL)
	if err != nil {
		return nil, err
	}

	if q.Limit != nil {
		if *q.Limit == 0 {
			q.Limit = nil
		}
		qry.Limit = q.Limit
	}

	if q.Offset != 0 {
		qry.Offset = &q.Offset
	}

	if len(q.PivotOn) > 0 {
		qry.PivotOn = q.PivotOn
	}

	qry.UseDisplayNames = export
	qry.Rows = q.Rows

	return qry, nil
}

func (q *MetricsViewAggregation) generateExportHeaders(ctx context.Context, rt *runtime.Runtime, instanceID string, opts *runtime.ExportOptions, qry *metricsview.Query) ([]string, error) {
	if !opts.IncludeHeader {
		return nil, nil
	}

	// Get org and project name from instance annotations.
	var org, project string
	inst, err := rt.Instance(ctx, instanceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get instance: %w", err)
	}
	if inst.Annotations != nil {
		org = inst.Annotations["organization_name"]
		project = inst.Annotations["project_name"]
	}

	var headers []string

	// Build title
	var parts []string
	if org != "" && project != "" {
		parts = append(parts, org, project)
	}
	var dashboardDisplayName string
	if opts.OriginDashboard != nil {
		dashboardDisplayName, err = q.getDisplayName(ctx, rt, instanceID, opts.OriginDashboard)
		if err != nil {
			return nil, fmt.Errorf("failed to get dashboard display name: %w", err)
		}
	}
	if dashboardDisplayName != "" {
		parts = append(parts, dashboardDisplayName)
	}
	title := "Report by Rill Data"
	if len(parts) > 0 {
		title += " – " + strings.Join(parts, " / ")
	}
	headers = append(headers, title)

	// Build date range
	if !qry.TimeRange.Start.IsZero() || !qry.TimeRange.End.IsZero() {
		timeRange := fmt.Sprintf("Date range: %s to %s", qry.TimeRange.Start.Format(time.RFC3339), qry.TimeRange.End.Format(time.RFC3339))
		headers = append(headers, timeRange)
	}

	// Build filters
	expStr, err := metricsview.ExpressionToExport(qry.Where)
	if err != nil {
		return nil, err
	}
	headers = append(headers, fmt.Sprintf("Filters: %s", expStr))

	// Add URL to dashboard
	if opts.OriginURL != "" && dashboardDisplayName != "" {
		headers = append(headers, fmt.Sprintf("Go to dashboard: %s", opts.OriginURL))
	}

	// Always add blank line at end
	headers = append(headers, "")

	return headers, nil
}

func (q *MetricsViewAggregation) getDisplayName(ctx context.Context, rt *runtime.Runtime, instanceID string, resourceName *runtimev1.ResourceName) (string, error) {
	c, err := rt.Controller(ctx, instanceID)
	if err != nil {
		return "", fmt.Errorf("failed to get controller: %w", err)
	}

	res, err := c.Get(ctx, resourceName, false)
	if err != nil {
		if errors.Is(err, drivers.ErrResourceNotFound) {
			return resourceName.Name, nil
		}
		return "", fmt.Errorf("failed to get resource: %w", err)
	}

	// Try to get DisplayName for known resource types
	switch resourceName.Kind {
	case runtime.ResourceKindExplore:
		explore := res.GetExplore()
		if explore != nil && explore.State != nil && explore.State.ValidSpec != nil && explore.State.ValidSpec.DisplayName != "" {
			return explore.State.ValidSpec.DisplayName, nil
		}
	case runtime.ResourceKindCanvas:
		canvas := res.GetCanvas()
		if canvas != nil && canvas.State != nil && canvas.State.ValidSpec != nil && canvas.State.ValidSpec.DisplayName != "" {
			return canvas.State.ValidSpec.DisplayName, nil
		}
	}
	return resourceName.Name, nil
}

func metricViewExpression(expr *runtimev1.Expression, sql string) (*metricsview.Expression, error) {
	if expr != nil && sql != "" {
		sqlExpr, err := metricssql.ParseFilter(sql)
		if err != nil {
			return nil, err
		}
		return &metricsview.Expression{
			Condition: &metricsview.Condition{
				Operator: metricsview.OperatorAnd,
				Expressions: []*metricsview.Expression{
					metricsview.NewExpressionFromProto(expr),
					sqlExpr,
				},
			},
		}, nil
	}
	if expr != nil {
		return metricsview.NewExpressionFromProto(expr), nil
	}
	if sql != "" {
		return metricssql.ParseFilter(sql)
	}
	return nil, nil
}

func anyToTime(tm any) (time.Time, error) {
	if tm == nil {
		return time.Time{}, nil
	}

	tmStr, ok := tm.(string)
	if !ok {
		t, ok := tm.(time.Time)
		if !ok {
			return time.Time{}, fmt.Errorf("unable to convert type %T to Time", tm)
		}
		return t, nil
	}
	return time.Parse(time.RFC3339Nano, tmStr)
}
