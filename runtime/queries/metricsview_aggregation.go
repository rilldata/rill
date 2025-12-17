package queries

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/metricsview/executor"
	"github.com/rilldata/rill/runtime/metricsview/metricssql"
	"github.com/rilldata/rill/runtime/pkg/pbutil"
	"google.golang.org/protobuf/types/known/structpb"
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
	IncludeTargets      bool                                           `json:"include_targets,omitempty"`

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

	var userAttrs map[string]any
	if q.SecurityClaims != nil {
		userAttrs = q.SecurityClaims.UserAttributes
	}

	e, err := executor.New(ctx, rt, instanceID, mv.ValidSpec, mv.Streaming, security, priority, userAttrs)
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

	// Query targets for measures that have targets configured (only if include_targets is true)
	if q.IncludeTargets {
		targetValues, err := q.queryTargets(ctx, rt, instanceID, mv.ValidSpec, qry, priority, userAttrs, security, data)
		if err != nil {
			// Log error but don't fail the query if targets fail
			// TODO: Add proper logging
			_ = err
			// Initialize targets to empty slice on error
			q.Result.Targets = []*runtimev1.MetricsViewTargetValue{}
		} else {
			q.Result.Targets = targetValues
		}
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

	var userAttrs map[string]any
	if q.SecurityClaims != nil {
		userAttrs = q.SecurityClaims.UserAttributes
	}

	e, err := executor.New(ctx, rt, instanceID, mv.ValidSpec, mv.Streaming, security, opts.Priority, userAttrs)
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

func (q *MetricsViewAggregation) queryTargets(
	ctx context.Context,
	rt *runtime.Runtime,
	instanceID string,
	mv *runtimev1.MetricsViewSpec,
	qry *metricsview.Query,
	priority int,
	userAttrs map[string]any,
	security *runtime.ResolvedSecurity,
	queryResultData []*structpb.Struct,
) ([]*runtimev1.MetricsViewTargetValue, error) {
	// Check if any measures have targets
	measureTargets := make(map[string][]*runtimev1.MetricsViewSpec_Target)
	for _, target := range mv.Targets {
		for _, measure := range target.Measures {
			// Check if this measure is in the query
			for _, qm := range qry.Measures {
				if getMeasureName(qm) == measure {
					measureTargets[measure] = append(measureTargets[measure], target)
					break
				}
			}
		}
	}

	if len(measureTargets) == 0 {
		// Return empty slice instead of nil to distinguish between "no targets" and "error"
		return []*runtimev1.MetricsViewTargetValue{}, nil
	}

	// Extract time grain from query
	timeGrain := extractTimeGrainFromQuery(qry)

	// Query targets for each measure
	var targetValues []*runtimev1.MetricsViewTargetValue
	e, err := executor.New(ctx, rt, instanceID, mv, false, security, priority, userAttrs)
	if err != nil {
		return nil, err
	}
	defer e.Close()

	// Build targets query
	targetsQuery := &metricsview.TargetsQuery{
		MetricsView: q.MetricsViewName,
		Measures:    make([]string, 0, len(measureTargets)),
		TimeRange:   qry.TimeRange,
		TimeZone:    qry.TimeZone,
		TimeGrain:   timeGrain,
		Priority:    priority,
	}

	for measure := range measureTargets {
		targetsQuery.Measures = append(targetsQuery.Measures, measure)
	}

	// Resolve time range if needed
	if targetsQuery.TimeRange != nil && (targetsQuery.TimeRange.Start.IsZero() || targetsQuery.TimeRange.End.IsZero()) {
		tsRes, err := ResolveTimestampResult(ctx, rt, instanceID, q.MetricsViewName, targetsQuery.TimeRange.TimeDimension, nil, priority)
		if err != nil {
			return nil, err
		}

		if targetsQuery.TimeRange.Start.IsZero() {
			targetsQuery.TimeRange.Start = tsRes.Min
		}
		if targetsQuery.TimeRange.End.IsZero() {
			targetsQuery.TimeRange.End = tsRes.Max
		}

		err = e.BindTargetsQuery(ctx, targetsQuery, tsRes)
		if err != nil {
			return nil, err
		}
	}

	// Execute targets query
	targetRows, err := e.Targets(ctx, targetsQuery)
	if err != nil {
		return nil, err
	}

	// Group target rows by measure and series
	targetsByMeasureSeries := make(map[string]map[string][]map[string]any)
	for _, row := range targetRows {
		var forMeasures []string
		if forMeasuresVal, ok := row["for_measures"]; ok {
			switch v := forMeasuresVal.(type) {
			case []string:
				forMeasures = v
			case []interface{}:
				forMeasures = make([]string, 0, len(v))
				for _, item := range v {
					if s, ok := item.(string); ok {
						forMeasures = append(forMeasures, s)
					}
				}
			}
		}
		if len(forMeasures) == 0 {
			continue
		}

		targetIdentifier, ok := row["target"].(string)
		if !ok {
			continue
		}

		for _, measure := range forMeasures {
			if _, ok := targetsByMeasureSeries[measure]; !ok {
				targetsByMeasureSeries[measure] = make(map[string][]map[string]any)
			}
			targetsByMeasureSeries[measure][targetIdentifier] = append(targetsByMeasureSeries[measure][targetIdentifier], row)
		}
	}

	// Build target values response
	// If query has no dimensions, aggregate targets into a single value per series
	hasDimensions := len(qry.Dimensions) > 0

	for measure, targetMap := range targetsByMeasureSeries {
		for targetIdentifier, rows := range targetMap {
			var values []*structpb.Struct

			// Get display name for target (if available)
			targetDisplayName := targetIdentifier
			if len(rows) > 0 {
				if dn, ok := rows[0]["target_name"].(string); ok && dn != "" {
					targetDisplayName = dn
				}
			}

			if !hasDimensions {
				// Aggregate targets into a single value
				// Sum values for simple/unspecified measures
				var aggregatedValue float64

				for _, row := range rows {
					// Sum the value - handle various numeric types that might come from database
					val, exists := row["value"]
					if !exists {
						continue
					}
					if val == nil {
						continue
					}

					var valueToAdd float64
					// Use reflection to handle numeric types more robustly
					rv := reflect.ValueOf(val)
					if !rv.IsValid() {
						continue
					}
					switch rv.Kind() {
					case reflect.Float64, reflect.Float32:
						valueToAdd = rv.Float()
					case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
						valueToAdd = float64(rv.Int())
					case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
						valueToAdd = float64(rv.Uint())
					default:
						// Try to convert using fmt.Sprintf and strconv as fallback
						if str := fmt.Sprintf("%v", val); str != "" && str != "<nil>" {
							if f, err := strconv.ParseFloat(str, 64); err == nil {
								valueToAdd = f
							} else {
								// If conversion fails, skip this row
								continue
							}
						} else {
							continue
						}
					}
					aggregatedValue += valueToAdd
				}

				// Create single aggregated struct
				aggregatedStruct := map[string]any{
					"target":      targetIdentifier,
					"target_name": targetDisplayName,
					"value":       aggregatedValue,
				}

				st, err := pbutil.ToStruct(aggregatedStruct, nil)
				if err != nil {
					return nil, fmt.Errorf("failed to convert aggregated target to struct: %w", err)
				}
				values = []*structpb.Struct{st}
			} else {
				// Match targets to query result rows by time period when time dimension is present
				timeDim, queryTimeGrain := getTimeDimensionFromQuery(qry)
				if timeDim != "" && len(queryResultData) > 0 {
					// Build a map of target rows by time
					// Use the coarsest grain that matches (target grain or query grain, whichever is coarser)
					tz := time.UTC
					if qry.TimeZone != "" {
						var err error
						tz, err = time.LoadLocation(qry.TimeZone)
						if err != nil {
							return nil, fmt.Errorf("invalid timezone %q: %w", qry.TimeZone, err)
						}
					}

					// Determine the effective grain to use for matching
					// Try to infer target grain from the data, or use query grain as fallback
					// If targets have daily data and query is hourly, we should match at day level
					effectiveGrain := queryTimeGrain
					if len(rows) > 1 {
						// Try to detect target grain by looking at time intervals
						var targetTimes []time.Time
						for _, targetRow := range rows {
							if targetTimeVal, ok := targetRow["time"]; ok {
								if targetTime, err := parseTargetTime(targetTimeVal); err == nil {
									targetTimes = append(targetTimes, targetTime)
								}
							}
						}
						if len(targetTimes) >= 2 {
							// Check if times are spaced by days (approximately 24 hours)
							dayCount := 0
							for i := 1; i < len(targetTimes); i++ {
								interval := targetTimes[i].Sub(targetTimes[i-1])
								// Check if interval is approximately 24 hours (allowing for some variance)
								hours := interval.Hours()
								if hours >= 23 && hours <= 25 {
									dayCount++
								}
							}
							// If most intervals are ~24 hours, targets are daily
							if dayCount > len(targetTimes)/2 {
								effectiveGrain = metricsview.TimeGrainDay
							}
						}
					}
					// Use the coarser grain between target grain and query grain for matching
					// This ensures daily targets match to hourly query rows at the day level
					// If detected target grain (e.g., Day) is coarser than query grain (e.g., Hour), use Day for matching
					if isGrainCoarserThan(queryTimeGrain, effectiveGrain) {
						// Query grain is coarser, use it
					} else if isGrainCoarserThan(effectiveGrain, queryTimeGrain) {
						// Detected target grain is coarser, keep it (e.g., Day when query is Hour)
					} else {
						// Same or can't determine, use query grain
						effectiveGrain = queryTimeGrain
					}

					targetMapByTime := make(map[string]map[string]any)
					for _, targetRow := range rows {
						targetTimeVal, ok := targetRow["time"]
						if !ok {
							continue
						}
						targetTime, err := parseTargetTime(targetTimeVal)
						if err != nil {
							continue
						}
						// Truncate target time to the effective grain for matching
						targetTimeKey := truncateTimeToGrain(targetTime, effectiveGrain, tz).Format(time.RFC3339)

						// Store the target row by its truncated time key
						targetMapByTime[targetTimeKey] = targetRow
					}

					// Build a set of unique time periods from query results at the effective grain
					// This ensures we return one target row per unique time period, not per query result row
					uniqueTimeKeys := make(map[string]bool)
					for _, resultRow := range queryResultData {
						resultTimeVal, ok := resultRow.Fields[timeDim]
						if !ok {
							continue
						}
						resultTime, err := parseTimeFromStructValue(resultTimeVal)
						if err != nil {
							continue
						}
						// Truncate result time to the effective grain (same as targets)
						resultTimeKey := truncateTimeToGrain(resultTime, effectiveGrain, tz).Format(time.RFC3339)
						uniqueTimeKeys[resultTimeKey] = true
					}

					// Match targets to unique time periods from query results
					values = make([]*structpb.Struct, 0, len(uniqueTimeKeys))
					for timeKey := range uniqueTimeKeys {
						targetRow, found := targetMapByTime[timeKey]
						if found {
							// Create a struct from the matching target row
							structRow := make(map[string]any)
							for k, v := range targetRow {
								if k != "for_measures" {
									structRow[k] = v
								}
							}
							st, err := pbutil.ToStruct(structRow, nil)
							if err != nil {
								return nil, fmt.Errorf("failed to convert target row to struct: %w", err)
							}
							values = append(values, st)
						}
					}
					// Sort values by time to ensure consistent ordering
					if len(values) > 0 && len(uniqueTimeKeys) > 0 {
						// Extract time from each value and sort
						type valueWithTime struct {
							value *structpb.Struct
							time  time.Time
						}
						valuesWithTime := make([]valueWithTime, 0, len(values))
						for _, v := range values {
							if timeVal, ok := v.Fields["time"]; ok {
								if t, err := parseTimeFromStructValue(timeVal); err == nil {
									valuesWithTime = append(valuesWithTime, valueWithTime{value: v, time: t})
								}
							}
						}
						// Sort by time
						for i := 0; i < len(valuesWithTime)-1; i++ {
							for j := i + 1; j < len(valuesWithTime); j++ {
								if valuesWithTime[i].time.After(valuesWithTime[j].time) {
									valuesWithTime[i], valuesWithTime[j] = valuesWithTime[j], valuesWithTime[i]
								}
							}
						}
						// Rebuild values in sorted order
						values = make([]*structpb.Struct, len(valuesWithTime))
						for i, vwt := range valuesWithTime {
							values[i] = vwt.value
						}
					}
				} else {
					// No time dimension or no query result data - return all target rows as-is
					values = make([]*structpb.Struct, 0, len(rows))
					for _, row := range rows {
						// Create a struct from the row, excluding for_measures
						structRow := make(map[string]any)
						for k, v := range row {
							if k != "for_measures" {
								structRow[k] = v
							}
						}
						st, err := pbutil.ToStruct(structRow, nil)
						if err != nil {
							return nil, fmt.Errorf("failed to convert target row to struct: %w", err)
						}
						values = append(values, st)
					}
				}
			}

			targetValues = append(targetValues, &runtimev1.MetricsViewTargetValue{
				Measure: measure,
				Target: &runtimev1.MetricsViewTargetInfo{
					Name:       targetIdentifier,
					TargetName: targetDisplayName,
				},
				Values: values,
			})
		}
	}

	return targetValues, nil
}

func extractTimeGrainFromQuery(qry *metricsview.Query) metricsview.TimeGrain {
	// Check dimensions for time grain
	for _, dim := range qry.Dimensions {
		if dim.Compute != nil && dim.Compute.TimeFloor != nil {
			return dim.Compute.TimeFloor.Grain
		}
	}

	// Fall back to time range round to grain
	if qry.TimeRange != nil {
		return qry.TimeRange.RoundToGrain
	}

	return metricsview.TimeGrainUnspecified
}

// getTimeDimensionFromQuery extracts the time dimension name and grain from the query
func getTimeDimensionFromQuery(qry *metricsview.Query) (string, metricsview.TimeGrain) {
	for _, dim := range qry.Dimensions {
		if dim.Compute != nil && dim.Compute.TimeFloor != nil {
			return dim.Compute.TimeFloor.Dimension, dim.Compute.TimeFloor.Grain
		}
	}
	return "", metricsview.TimeGrainUnspecified
}

// parseTimeFromStructValue parses a time value from a structpb.Value
func parseTimeFromStructValue(v *structpb.Value) (time.Time, error) {
	switch vv := v.Kind.(type) {
	case *structpb.Value_StringValue:
		t, err := time.Parse(time.RFC3339, vv.StringValue)
		if err != nil {
			return time.Time{}, fmt.Errorf("failed to parse time string: %w", err)
		}
		return t, nil
	case *structpb.Value_NumberValue:
		// Unix timestamp in seconds
		return time.Unix(int64(vv.NumberValue), 0), nil
	default:
		return time.Time{}, fmt.Errorf("unexpected time value type: %T", vv)
	}
}

// parseTargetTime parses the time value from a target row (helper function that matches executor signature)
func parseTargetTime(timeVal any) (time.Time, error) {
	switch v := timeVal.(type) {
	case time.Time:
		return v, nil
	case string:
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return time.Time{}, fmt.Errorf("failed to parse time: %w", err)
		}
		return t, nil
	default:
		return time.Time{}, fmt.Errorf("unexpected time type: %T", v)
	}
}

// truncateTimeToGrain truncates a time to the specified grain (matches executor function)
func truncateTimeToGrain(t time.Time, grain metricsview.TimeGrain, tz *time.Location) time.Time {
	// Convert to the timezone first
	t = t.In(tz)
	switch grain {
	case metricsview.TimeGrainDay:
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, tz)
	case metricsview.TimeGrainWeek:
		// Find the start of the week (Monday)
		weekday := int(t.Weekday())
		if weekday == 0 {
			weekday = 7 // Sunday becomes 7
		}
		daysFromMonday := weekday - 1
		return time.Date(t.Year(), t.Month(), t.Day()-daysFromMonday, 0, 0, 0, 0, tz)
	case metricsview.TimeGrainMonth:
		return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, tz)
	case metricsview.TimeGrainQuarter:
		quarter := (int(t.Month()) - 1) / 3
		month := time.Month(quarter*3 + 1)
		return time.Date(t.Year(), month, 1, 0, 0, 0, 0, tz)
	case metricsview.TimeGrainYear:
		return time.Date(t.Year(), 1, 1, 0, 0, 0, 0, tz)
	case metricsview.TimeGrainHour:
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, tz)
	case metricsview.TimeGrainMinute:
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 0, 0, tz)
	default:
		// For other grains or unspecified, return as-is
		return t
	}
}

// isGrainCoarserThan returns true if grain1 is coarser (larger) than grain2
func isGrainCoarserThan(grain1, grain2 metricsview.TimeGrain) bool {
	grainOrder := map[metricsview.TimeGrain]int{
		metricsview.TimeGrainMillisecond: 1,
		metricsview.TimeGrainSecond:      2,
		metricsview.TimeGrainMinute:      3,
		metricsview.TimeGrainHour:        4,
		metricsview.TimeGrainDay:         5,
		metricsview.TimeGrainWeek:        6,
		metricsview.TimeGrainMonth:       7,
		metricsview.TimeGrainQuarter:     8,
		metricsview.TimeGrainYear:        9,
	}
	order1, ok1 := grainOrder[grain1]
	order2, ok2 := grainOrder[grain2]
	if !ok1 || !ok2 {
		return false
	}
	return order1 > order2
}

func getMeasureName(m metricsview.Measure) string {
	if m.Compute == nil {
		return m.Name
	}
	// For computed measures, we need to get the underlying measure name
	// This is a simplified version - full implementation would handle all compute types
	switch {
	case m.Compute.ComparisonValue != nil:
		return m.Compute.ComparisonValue.Measure
	case m.Compute.ComparisonDelta != nil:
		return m.Compute.ComparisonDelta.Measure
	case m.Compute.ComparisonRatio != nil:
		return m.Compute.ComparisonRatio.Measure
	case m.Compute.PercentOfTotal != nil:
		return m.Compute.PercentOfTotal.Measure
	default:
		return m.Name
	}
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
