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

	if cfg.MetricsNullFillingImplementation == "pushdown" && qry.TimeRange != nil && !qry.TimeRange.Start.IsZero() && !qry.TimeRange.End.IsZero() {
		qry.Spine = &metricsview.Spine{
			TimeRange: &metricsview.TimeSpine{
				Start: qry.TimeRange.Start,
				End:   qry.TimeRange.End,
			},
		}
	}

	e, err := executor.New(ctx, rt, instanceID, mv.ValidSpec, mv.Streaming, security, priority)
	if err != nil {
		return err
	}
	defer e.Close()

	res, err := e.Query(ctx, qry, nil)
	if err != nil {
		return err
	}
	defer res.Close()

	return q.populateResult(res, timeDim, mv.ValidSpec, cfg.MetricsNullFillingImplementation)
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
