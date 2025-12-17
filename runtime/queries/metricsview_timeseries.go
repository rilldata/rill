package queries

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/metricsview/executor"
	"github.com/rilldata/rill/runtime/pkg/duration"
	"github.com/rilldata/rill/runtime/pkg/pbutil"
	"github.com/rilldata/rill/runtime/pkg/timeutil"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type MetricsViewTimeSeries struct {
	MetricsViewName string                       `json:"metrics_view_name,omitempty"`
	MeasureNames    []string                     `json:"measure_names,omitempty"`
	TimeStart       *timestamppb.Timestamp       `json:"time_start,omitempty"`
	TimeEnd         *timestamppb.Timestamp       `json:"time_end,omitempty"`
	Limit           int64                        `json:"limit,omitempty"`
	Offset          int64                        `json:"offset,omitempty"`
	Sort            []*runtimev1.MetricsViewSort `json:"sort,omitempty"`
	Where           *runtimev1.Expression        `json:"where,omitempty"`
	WhereSQL        string                       `json:"where_sql,omitempty"`
	Filter          *runtimev1.MetricsViewFilter `json:"filter,omitempty"` // backwards compatibility
	Having          *runtimev1.Expression        `json:"having,omitempty"`
	HavingSQL       string                       `json:"having_sql,omitempty"`
	TimeGranularity runtimev1.TimeGrain          `json:"time_granularity,omitempty"`
	TimeZone        string                       `json:"time_zone,omitempty"`
	SecurityClaims  *runtime.SecurityClaims      `json:"security_claims,omitempty"`
	TimeDimension   string                       `json:"time_dimension,omitempty"`
	IncludeTargets  bool                         `json:"include_targets,omitempty"`

	Result *runtimev1.MetricsViewTimeSeriesResponse `json:"-"`
}

var _ runtime.Query = &MetricsViewTimeSeries{}

func (q *MetricsViewTimeSeries) Key() string {
	r, err := json.Marshal(q)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("MetricsViewTimeSeries:%s", r)
}

func (q *MetricsViewTimeSeries) Deps() []*runtimev1.ResourceName {
	return []*runtimev1.ResourceName{
		{Kind: runtime.ResourceKindMetricsView, Name: q.MetricsViewName},
	}
}

func (q *MetricsViewTimeSeries) MarshalResult() *runtime.QueryResult {
	return &runtime.QueryResult{
		Value: q.Result,
		Bytes: sizeProtoMessage(q.Result),
	}
}

func (q *MetricsViewTimeSeries) UnmarshalResult(v any) error {
	res, ok := v.(*runtimev1.MetricsViewTimeSeriesResponse)
	if !ok {
		return fmt.Errorf("MetricsViewTimeSeries: mismatched unmarshal input")
	}
	q.Result = res
	return nil
}

func (q *MetricsViewTimeSeries) Resolve(ctx context.Context, rt *runtime.Runtime, instanceID string, priority int) error {
	mv, security, err := resolveMVAndSecurityFromAttributes(ctx, rt, instanceID, q.MetricsViewName, q.SecurityClaims)
	if err != nil {
		return err
	}

	timeDim := mv.ValidSpec.TimeDimension
	if q.TimeDimension != "" {
		timeDim = q.TimeDimension
	}

	if timeDim == "" {
		return fmt.Errorf("no time dimension specified for metrics view %q", q.MetricsViewName)
	}

	qry, err := q.rewriteToMetricsViewQuery(timeDim)
	if err != nil {
		return fmt.Errorf("error rewriting to metrics query: %w", err)
	}

	cfg, err := rt.InstanceConfig(ctx, instanceID)
	if err != nil {
		return err
	}

	nullImpl := cfg.MetricsNullFillingImplementation

	if nullImpl == "pushdown" {
		if qry.TimeRange != nil && !qry.TimeRange.Start.IsZero() && !qry.TimeRange.End.IsZero() {
			qry.Spine = &metricsview.Spine{
				TimeRange: &metricsview.TimeSpine{
					Start: qry.TimeRange.Start,
					End:   qry.TimeRange.End,
				},
			}
		} else {
			nullImpl = "" // cannot be pushed down so use legacy method
		}
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

	res, err := e.Query(ctx, qry, nil)
	if err != nil {
		return err
	}
	defer res.Close()

	err = q.populateResult(res, timeDim, mv.ValidSpec, nullImpl)
	if err != nil {
		return err
	}

	// Query targets for measures that have targets configured (only if include_targets is true)
	if q.IncludeTargets {
		targetValues, err := q.queryTargets(ctx, rt, instanceID, mv.ValidSpec, qry, timeDim, priority, userAttrs, security, q.Result.Data)
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

func (q *MetricsViewTimeSeries) Export(ctx context.Context, rt *runtime.Runtime, instanceID string, w io.Writer, opts *runtime.ExportOptions) error {
	err := q.Resolve(ctx, rt, instanceID, opts.Priority)
	if err != nil {
		return err
	}

	ctrl, err := rt.Controller(ctx, instanceID)
	if err != nil {
		return err
	}

	r, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: q.MetricsViewName}, false)
	if err != nil {
		return err
	}

	spec := r.GetMetricsView().State.ValidSpec
	if spec == nil {
		return fmt.Errorf("metrics view spec is not valid")
	}

	if opts.PreWriteHook != nil {
		err = opts.PreWriteHook(q.generateFilename())
		if err != nil {
			return err
		}
	}

	tmp := make([]*structpb.Struct, 0, len(q.Result.Data))
	meta := append([]*runtimev1.MetricsViewColumn{{
		Name: spec.TimeDimension,
	}}, q.Result.Meta...)
	for _, dt := range q.Result.Data {
		dt.Records.Fields[spec.TimeDimension] = structpb.NewStringValue(dt.Ts.AsTime().Format(time.RFC3339Nano))
		tmp = append(tmp, dt.Records)
	}

	switch opts.Format {
	case runtimev1.ExportFormat_EXPORT_FORMAT_UNSPECIFIED:
		return fmt.Errorf("unspecified format")
	case runtimev1.ExportFormat_EXPORT_FORMAT_CSV:
		return WriteCSV(meta, tmp, w)
	case runtimev1.ExportFormat_EXPORT_FORMAT_XLSX:
		return WriteXLSX(meta, tmp, w)
	case runtimev1.ExportFormat_EXPORT_FORMAT_PARQUET:
		return WriteParquet(meta, tmp, w)
	}

	return nil
}

func (q *MetricsViewTimeSeries) populateResult(rows *drivers.Result, tsAlias string, mv *runtimev1.MetricsViewSpec, nullFillingImplementation string) error {
	// Omit the time value from the result schema
	schema := rows.Schema
	if schema != nil {
		for i, f := range schema.Fields {
			if f.Name == tsAlias {
				schema.Fields = slices.Delete(schema.Fields, i, i+1)
				break
			}
		}
	}

	tz := time.UTC
	if q.TimeZone != "" {
		var err error
		tz, err = time.LoadLocation(q.TimeZone)
		if err != nil {
			return fmt.Errorf("invalid timezone '%s': %w", q.TimeZone, err)
		}
	}

	fdow := mv.FirstDayOfWeek
	if mv.FirstDayOfWeek > 7 || mv.FirstDayOfWeek <= 0 {
		fdow = 1
	}

	fmoy := mv.FirstMonthOfYear
	if mv.FirstMonthOfYear > 12 || mv.FirstMonthOfYear <= 0 {
		fmoy = 1
	}

	dur := timeGrainToDuration(q.TimeGranularity)

	var start time.Time
	var zeroTime time.Time
	var data []*runtimev1.TimeSeriesValue
	nullRecords := generateNullRecords(schema)
	rowMap := make(map[string]any)
	for rows.Next() {
		err := rows.MapScan(rowMap)
		if err != nil {
			return err
		}

		var t time.Time
		switch v := rowMap[tsAlias].(type) {
		case time.Time:
			t = v
		case *time.Time:
			if v != nil {
				t = *v
			}
		case int64:
			t = time.UnixMilli(v)
		default:
			if v != nil {
				panic(fmt.Sprintf("unexpected type for timestamp column: %T", v))
			}
		}
		delete(rowMap, tsAlias)

		records, err := pbutil.ToStruct(rowMap, schema)
		if err != nil {
			return err
		}

		if nullFillingImplementation == "" || nullFillingImplementation == "new" {
			if zeroTime.Equal(start) {
				if q.TimeStart != nil {
					start = timeutil.TruncateTime(q.TimeStart.AsTime(), timeutil.TimeGrainFromAPI(q.TimeGranularity), tz, int(fdow), int(fmoy))
					data = addNulls(data, nullRecords, start, t, dur, tz, nullFillingImplementation == "new")
				}
			} else {
				data = addNulls(data, nullRecords, start, t, dur, tz, nullFillingImplementation == "new")
			}
		}

		data = append(data, &runtimev1.TimeSeriesValue{
			Ts:      timestamppb.New(t),
			Records: records,
		})
		start = addTo(t, dur, tz)
	}
	err := rows.Err()
	if err != nil {
		return err
	}
	if q.TimeEnd != nil && nullRecords != nil {
		if start.Equal(zeroTime) && q.TimeStart != nil {
			start = q.TimeStart.AsTime()
		}

		if nullFillingImplementation == "" || nullFillingImplementation == "new" {
			if !start.Equal(zeroTime) {
				data = addNulls(data, nullRecords, start, q.TimeEnd.AsTime(), dur, tz, nullFillingImplementation == "new")
			}
		}
	}

	meta := structTypeToMetricsViewColumn(rows.Schema)

	q.Result = &runtimev1.MetricsViewTimeSeriesResponse{
		Meta: meta,
		Data: data,
	}

	return nil
}

func (q *MetricsViewTimeSeries) generateFilename() string {
	filename := strings.ReplaceAll(q.MetricsViewName, `"`, `_`)
	if q.TimeStart != nil || q.TimeEnd != nil || q.Where != nil || q.Having != nil || q.WhereSQL != "" || q.HavingSQL != "" {
		filename += "_filtered"
	}
	return filename
}

func generateNullRecords(schema *runtimev1.StructType) *structpb.Struct {
	nullStruct := structpb.Struct{Fields: make(map[string]*structpb.Value, len(schema.Fields))}
	for _, f := range schema.Fields {
		nullStruct.Fields[f.Name] = structpb.NewNullValue()
	}
	return &nullStruct
}

func addNulls(data []*runtimev1.TimeSeriesValue, nullRecords *structpb.Struct, start, end time.Time, d duration.Duration, tz *time.Location, newImplementation bool) []*runtimev1.TimeSeriesValue {
	if newImplementation {
		return addNullsNew(data, nullRecords, start, end, d, tz)
	}

	i := 0
	for start.Before(end) {
		if i > 5000 {
			break // safety break
		}
		i++
		data = append(data, &runtimev1.TimeSeriesValue{
			Ts:      timestamppb.New(start),
			Records: nullRecords,
		})
		start = addTo(start, d, tz)
	}
	return data
}

func addTo(t time.Time, d duration.Duration, tz *time.Location) time.Time {
	sd := d.(duration.StandardDuration)
	if sd.Hour > 0 || sd.Minute > 0 || sd.Second > 0 {
		return d.Add(t)
	}
	return d.Add(t.In(tz)).In(time.UTC)
}

func addNullsNew(data []*runtimev1.TimeSeriesValue, nullRecords *structpb.Struct, start, end time.Time, d duration.Duration, tz *time.Location) []*runtimev1.TimeSeriesValue {
	i := 0
	for start.Before(end) {
		if i > 5000 {
			break // safety break
		}
		i++
		data = append(data, &runtimev1.TimeSeriesValue{
			Ts:      timestamppb.New(start),
			Records: nullRecords,
		})
		newStart := addToNew(start, d, tz)
		// Defensive check: ensure time is progressing forward to prevent infinite loops
		// This can happen during DST transitions if timezone handling is incorrect
		if !newStart.After(start) {
			// Time didn't progress - break to prevent infinite loop and memory exhaustion
			break
		}
		start = newStart
	}
	return data
}

func addToNew(t time.Time, d duration.Duration, tz *time.Location) time.Time {
	sd := d.(duration.StandardDuration)
	if sd.Hour > 0 || sd.Minute > 0 || sd.Second > 0 {
		// For hours/minutes/seconds, add in UTC to get elapsed time
		// But ensure the input time is in UTC first to avoid DST issues
		return d.Add(t.In(time.UTC))
	}
	// For days/weeks/months, respect timezone to handle DST transitions
	return d.Add(t.In(tz)).In(time.UTC)
}

func (q *MetricsViewTimeSeries) rewriteToMetricsViewQuery(timeDimension string) (*metricsview.Query, error) {
	qry := &metricsview.Query{MetricsView: q.MetricsViewName}

	for _, m := range q.MeasureNames {
		qry.Measures = append(qry.Measures, metricsview.Measure{Name: m})
	}

	res := &metricsview.TimeRange{}
	if q.TimeStart != nil {
		res.Start = q.TimeStart.AsTime()
	}
	if q.TimeEnd != nil {
		res.End = q.TimeEnd.AsTime()
	}
	res.TimeDimension = timeDimension
	qry.TimeRange = res

	if q.Limit != 0 {
		qry.Limit = &q.Limit
	}

	if q.Offset != 0 {
		qry.Offset = &q.Offset
	}

	for _, s := range q.Sort {
		qry.Sort = append(qry.Sort, metricsview.Sort{
			Name: s.Name,
			Desc: !s.Ascending,
		})
	}

	if len(q.Sort) == 0 {
		qry.Sort = append(qry.Sort, metricsview.Sort{
			Name: timeDimension,
			Desc: false,
		})
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
		return nil, fmt.Errorf("error converting where clause: %w", err)
	}

	qry.Having, err = metricViewExpression(q.Having, q.HavingSQL)
	if err != nil {
		return nil, fmt.Errorf("error converting having clause: %w", err)
	}

	qry.Dimensions = append(qry.Dimensions, metricsview.Dimension{
		Name: timeDimension,
		Compute: &metricsview.DimensionCompute{
			TimeFloor: &metricsview.DimensionComputeTimeFloor{
				Dimension: timeDimension,
				Grain:     metricsview.TimeGrainFromProto(q.TimeGranularity),
			},
		},
	})

	qry.TimeZone = q.TimeZone

	return qry, nil
}

func (q *MetricsViewTimeSeries) queryTargets(
	ctx context.Context,
	rt *runtime.Runtime,
	instanceID string,
	mv *runtimev1.MetricsViewSpec,
	qry *metricsview.Query,
	timeDim string,
	priority int,
	userAttrs map[string]any,
	security *runtime.ResolvedSecurity,
	timeSeriesData []*runtimev1.TimeSeriesValue,
) ([]*runtimev1.MetricsViewTargetValue, error) {
	// Check if any measures have targets
	measureTargets := make(map[string][]*runtimev1.MetricsViewSpec_Target)
	for _, target := range mv.Targets {
		for _, measure := range target.Measures {
			// Check if this measure is in the query
			for _, measureName := range q.MeasureNames {
				if measureName == measure {
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
	queryTimeGrain := metricsview.TimeGrainFromProto(q.TimeGranularity)

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
		TimeGrain:   queryTimeGrain,
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

	// Group target rows by measure and target identifier
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

	// Match targets to time series data points by time period
	tz := time.UTC
	if q.TimeZone != "" {
		var err error
		tz, err = time.LoadLocation(q.TimeZone)
		if err != nil {
			return nil, fmt.Errorf("invalid timezone %q: %w", q.TimeZone, err)
		}
	}

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

			// Determine the effective grain to use for matching
			effectiveGrain := queryTimeGrain
			if len(rows) > 1 {
				// Try to detect target grain by looking at time intervals
				var targetTimes []time.Time
				for _, targetRow := range rows {
					if targetTimeVal, ok := targetRow["time"]; ok {
						if targetTime, err := parseTargetTimeForTimeSeries(targetTimeVal); err == nil {
							targetTimes = append(targetTimes, targetTime)
						}
					}
				}
				if len(targetTimes) >= 2 {
					// Check if times are spaced by days (approximately 24 hours)
					dayCount := 0
					for i := 1; i < len(targetTimes); i++ {
						interval := targetTimes[i].Sub(targetTimes[i-1])
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
			if isGrainCoarserThan(queryTimeGrain, effectiveGrain) {
				// Query grain is coarser, use it
			} else if isGrainCoarserThan(effectiveGrain, queryTimeGrain) {
				// Detected target grain is coarser, keep it (e.g., Day when query is Hour)
			} else {
				// Same or can't determine, use query grain
				effectiveGrain = queryTimeGrain
			}

			// Build a map of target rows by time (truncated to effective grain)
			targetMapByTime := make(map[string]map[string]any)
			for _, targetRow := range rows {
				targetTimeVal, ok := targetRow["time"]
				if !ok {
					continue
				}
				targetTime, err := parseTargetTimeForTimeSeries(targetTimeVal)
				if err != nil {
					continue
				}
				// Truncate target time to the effective grain for matching
				targetTimeKey := truncateTimeToGrainForTimeSeries(targetTime, effectiveGrain, tz).Format(time.RFC3339)

				// Store the target row by its truncated time key
				targetMapByTime[targetTimeKey] = targetRow
			}

			// Build a set of unique time periods from time series data at the effective grain
			uniqueTimeKeys := make(map[string]bool)
			for _, tsValue := range timeSeriesData {
				if tsValue.Ts == nil {
					continue
				}
				resultTime := tsValue.Ts.AsTime()
				// Truncate result time to the effective grain (same as targets)
				resultTimeKey := truncateTimeToGrainForTimeSeries(resultTime, effectiveGrain, tz).Format(time.RFC3339)
				uniqueTimeKeys[resultTimeKey] = true
			}

			// Match targets to unique time periods from time series data
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
			if len(values) > 0 {
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

// parseTargetTimeForTimeSeries parses the time value from a target row for timeseries
func parseTargetTimeForTimeSeries(timeVal any) (time.Time, error) {
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

// truncateTimeToGrainForTimeSeries truncates a time to the specified grain for timeseries
func truncateTimeToGrainForTimeSeries(t time.Time, grain metricsview.TimeGrain, tz *time.Location) time.Time {
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
