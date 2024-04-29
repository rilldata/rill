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
	"github.com/rilldata/rill/runtime/pkg/duration"
	"github.com/rilldata/rill/runtime/pkg/pbutil"
	"github.com/rilldata/rill/runtime/pkg/timeutil"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type MetricsViewTimeSeries struct {
	MetricsViewName    string                               `json:"metrics_view_name,omitempty"`
	MeasureNames       []string                             `json:"measure_names,omitempty"`
	InlineMeasures     []*runtimev1.InlineMeasure           `json:"inline_measures,omitempty"`
	TimeStart          *timestamppb.Timestamp               `json:"time_start,omitempty"`
	TimeEnd            *timestamppb.Timestamp               `json:"time_end,omitempty"`
	Limit              int64                                `json:"limit,omitempty"`
	Offset             int64                                `json:"offset,omitempty"`
	Sort               []*runtimev1.MetricsViewSort         `json:"sort,omitempty"`
	Where              *runtimev1.Expression                `json:"where,omitempty"`
	Having             *runtimev1.Expression                `json:"having,omitempty"`
	TimeGranularity    runtimev1.TimeGrain                  `json:"time_granularity,omitempty"`
	TimeZone           string                               `json:"time_zone,omitempty"`
	MetricsView        *runtimev1.MetricsViewSpec           `json:"-"`
	ResolvedMVSecurity *runtime.ResolvedMetricsViewSecurity `json:"security"`

	// backwards compatibility
	Filter *runtimev1.MetricsViewFilter `json:"filter,omitempty"`

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
	if q.MetricsView.TimeDimension == "" {
		return fmt.Errorf("metrics view '%s' does not have a time dimension", q.MetricsViewName)
	}

	olap, release, err := rt.OLAP(ctx, instanceID, q.MetricsView.Connector)
	if err != nil {
		return err
	}
	defer release()

	ctrl, err := rt.Controller(ctx, instanceID)
	if err != nil {
		return err
	}

	r, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: q.MetricsViewName}, false)
	if err != nil {
		return err
	}

	// backwards compatibility
	if q.Filter != nil {
		if q.Where != nil {
			return fmt.Errorf("both filter and where is provided")
		}
		q.Where = convertFilterToExpression(q.Filter)
	}

	mv := r.GetMetricsView().Spec
	sql, tsAlias, args, err := q.buildMetricsTimeseriesSQL(olap, mv, q.ResolvedMVSecurity)
	if err != nil {
		return fmt.Errorf("error building query: %w", err)
	}

	rows, err := olap.Execute(ctx, &drivers.Statement{
		Query:            sql,
		Args:             args,
		Priority:         priority,
		ExecutionTimeout: defaultExecutionTimeout,
	})
	if err != nil {
		return err
	}
	defer rows.Close()

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
			if olap.Dialect() != drivers.DialectPinot {
				panic(fmt.Sprintf("unexpected type for timestamp column: %T", v))
			}
			t = time.UnixMilli(v)
		default:
			panic(fmt.Sprintf("unexpected type for timestamp column: %T", v))
		}
		delete(rowMap, tsAlias)

		records, err := pbutil.ToStruct(rowMap, schema)
		if err != nil {
			return err
		}

		if zeroTime.Equal(start) {
			if q.TimeStart != nil {
				start = timeutil.TruncateTime(q.TimeStart.AsTime(), convTimeGrain(q.TimeGranularity), tz, int(fdow), int(fmoy))
				data = addNulls(data, nullRecords, start, t, dur, tz)
			}
		} else {
			data = addNulls(data, nullRecords, start, t, dur, tz)
		}

		data = append(data, &runtimev1.TimeSeriesValue{
			Ts:      timestamppb.New(t),
			Records: records,
		})
		start = addTo(t, dur, tz)
	}
	if q.TimeEnd != nil && nullRecords != nil {
		if start.Equal(zeroTime) && q.TimeStart != nil {
			start = q.TimeStart.AsTime()
		}

		if !start.Equal(zeroTime) {
			data = addNulls(data, nullRecords, start, q.TimeEnd.AsTime(), dur, tz)
		}
	}

	meta := structTypeToMetricsViewColumn(rows.Schema)

	q.Result = &runtimev1.MetricsViewTimeSeriesResponse{
		Meta: meta,
		Data: data,
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

	mv := r.GetMetricsView().Spec

	if opts.PreWriteHook != nil {
		err = opts.PreWriteHook(q.generateFilename())
		if err != nil {
			return err
		}
	}

	tmp := make([]*structpb.Struct, 0, len(q.Result.Data))
	meta := append([]*runtimev1.MetricsViewColumn{{
		Name: mv.TimeDimension,
	}}, q.Result.Meta...)
	for _, dt := range q.Result.Data {
		dt.Records.Fields[mv.TimeDimension] = structpb.NewStringValue(dt.Ts.AsTime().Format(time.RFC3339Nano))
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

func (q *MetricsViewTimeSeries) generateFilename() string {
	filename := strings.ReplaceAll(q.MetricsViewName, `"`, `_`)
	if q.TimeStart != nil || q.TimeEnd != nil || q.Where != nil || q.Having != nil {
		filename += "_filtered"
	}
	return filename
}

func (q *MetricsViewTimeSeries) buildMetricsTimeseriesSQL(olap drivers.OLAPStore, mv *runtimev1.MetricsViewSpec, policy *runtime.ResolvedMetricsViewSecurity) (string, string, []any, error) {
	ms, err := resolveMeasures(mv, q.InlineMeasures, q.MeasureNames)
	if err != nil {
		return "", "", nil, err
	}

	selectCols := []string{}
	for _, m := range ms {
		expr := fmt.Sprintf(`%s as "%s"`, m.Expression, m.Name)
		selectCols = append(selectCols, expr)
	}

	td := safeName(mv.TimeDimension)
	if olap.Dialect() == drivers.DialectDuckDB {
		td = fmt.Sprintf("%s::TIMESTAMP", td)
	}

	whereClause := "1=1"
	args := []any{}
	if q.TimeStart != nil {
		whereClause += fmt.Sprintf(" AND %s >= ?", td)
		args = append(args, q.TimeStart.AsTime())
	}
	if q.TimeEnd != nil {
		whereClause += fmt.Sprintf(" AND %s < ?", td)
		args = append(args, q.TimeEnd.AsTime())
	}

	if q.Where != nil {
		builder := &ExpressionBuilder{
			mv:      mv,
			dialect: olap.Dialect(),
		}
		clause, clauseArgs, err := builder.buildExpression(q.Where)
		if err != nil {
			return "", "", nil, err
		}
		if strings.TrimSpace(clause) != "" {
			whereClause += fmt.Sprintf(" AND (%s)", clause)
		}
		args = append(args, clauseArgs...)
	}

	if policy != nil && policy.RowFilter != "" {
		whereClause += fmt.Sprintf(" AND (%s)", policy.RowFilter)
	}

	havingClause := ""
	if q.Having != nil {
		builder := &ExpressionBuilder{
			mv:      mv,
			dialect: olap.Dialect(),
			having:  true,
		}
		clause, clauseArgs, err := builder.buildExpression(q.Having)
		if err != nil {
			return "", "", nil, err
		}
		if strings.TrimSpace(clause) != "" {
			havingClause = " HAVING " + clause
		}
		args = append(args, clauseArgs...)
	}

	tsAlias := tempName("_ts_")
	timezone := "UTC"
	if q.TimeZone != "" {
		timezone = q.TimeZone
	}

	var sql string
	switch olap.Dialect() {
	case drivers.DialectDuckDB:
		sql = q.buildDuckDBSQL(mv, tsAlias, selectCols, whereClause, havingClause, timezone)
	case drivers.DialectDruid:
		args = append([]any{timezone}, args...)
		sql = q.buildDruidSQL(mv, tsAlias, selectCols, whereClause, havingClause)
	case drivers.DialectPinot:
		args = append([]any{timezone}, args...)
		sql = q.buildPinotSQL(mv, tsAlias, selectCols, whereClause, havingClause)
	case drivers.DialectClickHouse:
		sql = q.buildClickHouseSQL(mv, tsAlias, selectCols, whereClause, havingClause, timezone)
	default:
		return "", "", nil, fmt.Errorf("not available for dialect '%s'", olap.Dialect())
	}

	return sql, tsAlias, args, nil
}

func (q *MetricsViewTimeSeries) buildPinotSQL(mv *runtimev1.MetricsViewSpec, tsAlias string, selectCols []string, whereClause, havingClause string) string {
	dateTruncSpecifier := drivers.DialectPinot.ConvertToDateTruncSpecifier(q.TimeGranularity)

	// TODO: handle shift, currently we add validation error for this, see runtime/validate.go

	timeClause := fmt.Sprintf("DATETRUNC('%s', %s,'MILLISECONDS', ?)", dateTruncSpecifier, safeName(mv.TimeDimension))
	sql := fmt.Sprintf(
		`SELECT %s AS %s, %s FROM %s WHERE %s GROUP BY 1 %s ORDER BY 1`,
		timeClause,
		tsAlias,
		strings.Join(selectCols, ", "),
		safeName(mv.Table),
		whereClause,
		havingClause,
	)

	return sql
}

func (q *MetricsViewTimeSeries) buildDruidSQL(mv *runtimev1.MetricsViewSpec, tsAlias string, selectCols []string, whereClause, havingClause string) string {
	tsSpecifier := convertToDruidTimeFloorSpecifier(q.TimeGranularity)

	timeClause := fmt.Sprintf("time_floor(%s, '%s', null, CAST(? AS VARCHAR))", safeName(mv.TimeDimension), tsSpecifier)
	if q.TimeGranularity == runtimev1.TimeGrain_TIME_GRAIN_WEEK && mv.FirstDayOfWeek > 1 {
		dayOffset := 8 - mv.FirstDayOfWeek
		timeClause = fmt.Sprintf("time_shift(time_floor(time_shift(%[1]s, 'P1D', %[3]d), '%[2]s', null, CAST(? AS VARCHAR)), 'P1D', -%[3]d)", safeName(mv.TimeDimension), tsSpecifier, dayOffset)
	} else if q.TimeGranularity == runtimev1.TimeGrain_TIME_GRAIN_YEAR && mv.FirstMonthOfYear > 1 {
		monthOffset := 13 - mv.FirstMonthOfYear
		timeClause = fmt.Sprintf("time_shift(time_floor(time_shift(%[1]s, 'P1M', %[3]d), '%[2]s', null, CAST(? AS VARCHAR)), 'P1M', -%[3]d)", safeName(mv.TimeDimension), tsSpecifier, monthOffset)
	}

	sql := fmt.Sprintf(
		`SELECT %s AS %s, %s FROM %s WHERE %s GROUP BY 1 %s ORDER BY 1`,
		timeClause,
		tsAlias,
		strings.Join(selectCols, ", "),
		escapeMetricsViewTable(drivers.DialectDruid, mv),
		whereClause,
		havingClause,
	)

	return sql
}

func (q *MetricsViewTimeSeries) buildClickHouseSQL(mv *runtimev1.MetricsViewSpec, tsAlias string, selectCols []string, whereClause, havingClause, timezone string) string {
	dateTruncSpecifier := drivers.DialectClickHouse.ConvertToDateTruncSpecifier(q.TimeGranularity)

	shift := "" // shift to accommodate FirstDayOfWeek or FirstMonthOfYear
	if q.TimeGranularity == runtimev1.TimeGrain_TIME_GRAIN_WEEK && mv.FirstDayOfWeek > 1 {
		offset := 8 - mv.FirstDayOfWeek
		shift = fmt.Sprintf("%d day", offset)
	} else if q.TimeGranularity == runtimev1.TimeGrain_TIME_GRAIN_YEAR && mv.FirstMonthOfYear > 1 {
		offset := 13 - mv.FirstMonthOfYear
		shift = fmt.Sprintf("%d month", offset)
	}

	sql := ""
	if shift == "" {
		sql = fmt.Sprintf(
			`
					SELECT
					toTimeZone(date_trunc('%[1]s', toTimeZone(%[2]s::DateTime64, '%[7]s'))::DateTime64, '%[7]s') as %[3]s,
					%[4]s
					FROM %[5]s
					WHERE %[6]s
					GROUP BY %[3]s
					%[8]s
					ORDER BY %[3]s`,
			dateTruncSpecifier,             // 1
			safeName(mv.TimeDimension),     // 2
			tsAlias,                        // 3
			strings.Join(selectCols, ", "), // 4
			escapeMetricsViewTable(drivers.DialectClickHouse, mv), // 5
			whereClause,  // 6
			timezone,     // 7
			havingClause, // 8
		)
	} else {
		sql = fmt.Sprintf(
			`
				SELECT
					toTimeZone(date_trunc('%[1]s', toTimeZone(%[2]s::DateTime64, '%[7]s') + INTERVAL %[8]s) - (INTERVAL %[8]s), '%[7]s') as %[3]s,
				%[4]s
				FROM %[5]s
				WHERE %[6]s
				GROUP BY %[3]s
				%[9]s
				ORDER BY %[3]s`,
			dateTruncSpecifier,             // 1
			safeName(mv.TimeDimension),     // 2
			tsAlias,                        // 3
			strings.Join(selectCols, ", "), // 4
			escapeMetricsViewTable(drivers.DialectClickHouse, mv), // 5
			whereClause,  // 6
			timezone,     // 7
			shift,        // 8
			havingClause, // 9
		)
	}

	return sql
}

func (q *MetricsViewTimeSeries) buildDuckDBSQL(mv *runtimev1.MetricsViewSpec, tsAlias string, selectCols []string, whereClause, havingClause, timezone string) string {
	dateTruncSpecifier := drivers.DialectDuckDB.ConvertToDateTruncSpecifier(q.TimeGranularity)

	shift := "" // shift to accommodate FirstDayOfWeek or FirstMonthOfYear
	if q.TimeGranularity == runtimev1.TimeGrain_TIME_GRAIN_WEEK && mv.FirstDayOfWeek > 1 {
		offset := 8 - mv.FirstDayOfWeek
		shift = fmt.Sprintf("%d DAY", offset)
	} else if q.TimeGranularity == runtimev1.TimeGrain_TIME_GRAIN_YEAR && mv.FirstMonthOfYear > 1 {
		offset := 13 - mv.FirstMonthOfYear
		shift = fmt.Sprintf("%d MONTH", offset)
	}

	sql := ""
	if shift == "" {
		if q.TimeGranularity == runtimev1.TimeGrain_TIME_GRAIN_HOUR ||
			q.TimeGranularity == runtimev1.TimeGrain_TIME_GRAIN_MINUTE ||
			q.TimeGranularity == runtimev1.TimeGrain_TIME_GRAIN_SECOND {
			sql = fmt.Sprintf(
				`
					SELECT
						time_bucket(INTERVAL '1 %[1]s', %[2]s::TIMESTAMPTZ, '%[7]s') as %[3]s,
						%[4]s
					FROM %[5]s
					WHERE %[6]s
					GROUP BY 1
					%[8]s
					ORDER BY 1`,
				dateTruncSpecifier,             // 1
				safeName(mv.TimeDimension),     // 2
				tsAlias,                        // 3
				strings.Join(selectCols, ", "), // 4
				escapeMetricsViewTable(drivers.DialectDuckDB, mv), // 5
				whereClause,  // 6
				timezone,     // 7
				havingClause, // 8
			)
		} else { // date_trunc is faster than time_bucket for year, month, week
			sql = fmt.Sprintf(
				`
					SELECT
					timezone('%[7]s', date_trunc('%[1]s', timezone('%[7]s', %[2]s::TIMESTAMPTZ))) as %[3]s,
					%[4]s
					FROM %[5]s
					WHERE %[6]s
					GROUP BY 1
					%[8]s
					ORDER BY 1`,
				dateTruncSpecifier,             // 1
				safeName(mv.TimeDimension),     // 2
				tsAlias,                        // 3
				strings.Join(selectCols, ", "), // 4
				escapeMetricsViewTable(drivers.DialectDuckDB, mv), // 5
				whereClause,  // 6
				timezone,     // 7
				havingClause, // 8
			)
		}
	} else {
		sql = fmt.Sprintf(
			`
				SELECT
					timezone('%[7]s', date_trunc('%[1]s', timezone('%[7]s', %[2]s::TIMESTAMPTZ) + INTERVAL %[8]s) - (INTERVAL %[8]s)) as %[3]s,
				%[4]s
				FROM %[5]s
				WHERE %[6]s
				GROUP BY 1
				%[9]s
				ORDER BY 1`,
			dateTruncSpecifier,             // 1
			safeName(mv.TimeDimension),     // 2
			tsAlias,                        // 3
			strings.Join(selectCols, ", "), // 4
			escapeMetricsViewTable(drivers.DialectDuckDB, mv), // 5
			whereClause,  // 6
			timezone,     // 7
			shift,        // 8
			havingClause, // 9
		)
	}

	return sql
}

func generateNullRecords(schema *runtimev1.StructType) *structpb.Struct {
	nullStruct := structpb.Struct{Fields: make(map[string]*structpb.Value, len(schema.Fields))}
	for _, f := range schema.Fields {
		nullStruct.Fields[f.Name] = structpb.NewNullValue()
	}
	return &nullStruct
}

func addNulls(data []*runtimev1.TimeSeriesValue, nullRecords *structpb.Struct, start, end time.Time, d duration.Duration, tz *time.Location) []*runtimev1.TimeSeriesValue {
	for start.Before(end) {
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
