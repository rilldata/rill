package queries

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"slices"
	"strconv"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/pbutil"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	// Load IANA time zone data
	_ "time/tzdata"
)

const IsoFormat string = "2006-01-02T15:04:05.000Z"

type ColumnTimeseriesResult struct {
	Meta       []*runtimev1.MetricsViewColumn
	Results    []*runtimev1.TimeSeriesValue
	Spark      []*runtimev1.TimeSeriesValue
	TimeRange  *runtimev1.TimeSeriesTimeRange
	SampleSize int32
}

type ColumnTimeseries struct {
	TableName           string                                            `json:"table_name"`
	Measures            []*runtimev1.ColumnTimeSeriesRequest_BasicMeasure `json:"measures"`
	TimestampColumnName string                                            `json:"timestamp_column_name"`
	TimeRange           *runtimev1.TimeSeriesTimeRange                    `json:"time_range"`
	Pixels              int32                                             `json:"pixels"`
	SampleSize          int32                                             `json:"sample_size"`
	TimeZone            string                                            `json:"time_zone,omitempty"`
	Result              *ColumnTimeseriesResult                           `json:"-"`
	FirstDayOfWeek      uint32
	FirstMonthOfYear    uint32

	// MetricsView-related fields. These can be removed when MetricsViewTimeSeries is refactored to a standalone implementation.
	MetricsView       *runtimev1.MetricsViewSpec           `json:"-"`
	MetricsViewFilter *runtimev1.MetricsViewFilter         `json:"filters"`
	MetricsViewPolicy *runtime.ResolvedMetricsViewSecurity `json:"security"`
}

var _ runtime.Query = &ColumnTimeseries{}

func (q *ColumnTimeseries) Key() string {
	r, err := json.Marshal(q)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("ColumnTimeseries:%s", r)
}

func (q *ColumnTimeseries) Deps() []*runtimev1.ResourceName {
	return []*runtimev1.ResourceName{
		{Kind: runtime.ResourceKindSource, Name: q.TableName},
		{Kind: runtime.ResourceKindModel, Name: q.TableName},
	}
}

func (q *ColumnTimeseries) MarshalResult() *runtime.QueryResult {
	return &runtime.QueryResult{
		Value: q.Result,
		Bytes: approxSize(q.Result),
	}
}

func (q *ColumnTimeseries) UnmarshalResult(v any) error {
	res, ok := v.(*ColumnTimeseriesResult)
	if !ok {
		return fmt.Errorf("ColumnTimeseries: mismatched unmarshal input")
	}
	q.Result = res
	return nil
}

func (q *ColumnTimeseries) Resolve(ctx context.Context, rt *runtime.Runtime, instanceID string, priority int) error {
	olap, release, err := rt.OLAP(ctx, instanceID)
	if err != nil {
		return err
	}
	defer release()

	if olap.Dialect() != drivers.DialectDuckDB {
		return fmt.Errorf("not available for dialect '%s'", olap.Dialect())
	}

	timeRange, err := q.ResolveNormaliseTimeRange(ctx, rt, instanceID, priority)
	if err != nil {
		return err
	}

	if timeRange.Interval == runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED {
		q.Result = &ColumnTimeseriesResult{}
		return nil
	}

	return olap.WithConnection(ctx, priority, false, false, func(ctx context.Context, ensuredCtx context.Context, _ *sql.Conn) error {
		filter := ""
		var args []any

		measures := normaliseMeasures(q.Measures, q.Pixels != 0)
		dateTruncSpecifier := convertToDateTruncSpecifier(timeRange.Interval)
		tsAlias := tempName("_ts_")
		temporaryTableName := tempName("_timeseries_")

		timezone := "UTC"
		if q.TimeZone != "" {
			timezone = q.TimeZone
		}

		_, err = time.LoadLocation(timezone)
		if err != nil {
			return err
		}

		args = append([]any{timezone, timeRange.Start.AsTime(), timezone, timeRange.End.AsTime(), timezone}, args...)
		args = append(args, timezone)

		if q.FirstDayOfWeek > 7 || q.FirstDayOfWeek <= 0 {
			q.FirstDayOfWeek = 1
		}

		if q.FirstMonthOfYear > 12 || q.FirstMonthOfYear <= 0 {
			q.FirstMonthOfYear = 1
		}

		timeOffsetClause1 := ""
		timeOffsetClause2 := ""
		if timeRange.Interval == runtimev1.TimeGrain_TIME_GRAIN_WEEK && q.FirstDayOfWeek > 1 {
			dayOffset := 8 - q.FirstDayOfWeek
			timeOffsetClause1 = fmt.Sprintf(" + INTERVAL '%d DAY'", dayOffset)
			timeOffsetClause2 = fmt.Sprintf(" - INTERVAL '%d DAY'", dayOffset)
		} else if timeRange.Interval == runtimev1.TimeGrain_TIME_GRAIN_YEAR && q.FirstMonthOfYear > 1 {
			monthOffset := 13 - q.FirstMonthOfYear
			timeOffsetClause1 = fmt.Sprintf(" + INTERVAL '%d MONTH'", monthOffset)
			timeOffsetClause2 = fmt.Sprintf(" - INTERVAL '%d MONTH'", monthOffset)
		}

		querySQL := `CREATE TEMPORARY TABLE ` + temporaryTableName + ` AS (
			-- generate a time series column that has the intended range
			WITH template as (
				SELECT
					range as ` + tsAlias + `
				FROM
					range(
					date_trunc('` + dateTruncSpecifier + `', timezone(?, ?::TIMESTAMPTZ) ` + timeOffsetClause1 + `) ` + timeOffsetClause2 + `,
					date_trunc('` + dateTruncSpecifier + `', timezone(?, ?::TIMESTAMPTZ) ` + timeOffsetClause1 + `) ` + timeOffsetClause2 + `,
					INTERVAL '1 ` + dateTruncSpecifier + `')
			),
			-- transform the original data, and optionally sample it.
			series AS (
			SELECT
				date_trunc('` + dateTruncSpecifier + `', timezone(?, ` + safeName(q.TimestampColumnName) + `::TIMESTAMPTZ) ` + timeOffsetClause1 + `) ` + timeOffsetClause2 + ` as ` + tsAlias + `,` + getExpressionColumnsFromMeasures(measures) + `
			FROM ` + safeName(q.TableName) + ` ` + filter + `
			GROUP BY ` + tsAlias + ` ORDER BY ` + tsAlias + `
			)
			-- an additional grouping is required for time zone DST (see unit tests for examples)
			SELECT ` + tsAlias + `,` + getCoalesceStatementsMeasuresLast(measures) + ` FROM (
				-- join the transformed data with the generated time series column,
				-- coalescing the first value to get the 0-default when the rolled up data
				-- does not have that value.
				SELECT
				` + getCoalesceStatementsMeasures(measures) + `,
				timezone(?, template.` + tsAlias + `) as ` + tsAlias + ` from template
				LEFT OUTER JOIN series ON template.` + tsAlias + ` = series.` + tsAlias + `
				ORDER BY template.` + tsAlias + `
			) GROUP BY 1 ORDER BY 1
		)`

		err = olap.Exec(ctx, &drivers.Statement{
			Query:            querySQL,
			Args:             args,
			Priority:         priority,
			ExecutionTimeout: defaultExecutionTimeout,
		})
		if err != nil {
			return err
		}
		defer func() {
			// NOTE: Using ensuredCtx
			_ = olap.Exec(ensuredCtx, &drivers.Statement{
				Query:            `DROP TABLE "` + temporaryTableName + `"`,
				Priority:         priority,
				ExecutionTimeout: defaultExecutionTimeout,
			})
		}()

		rows, err := olap.Execute(ctx, &drivers.Statement{
			Query:            fmt.Sprintf(`SELECT * FROM %q`, temporaryTableName),
			Priority:         priority,
			ExecutionTimeout: defaultExecutionTimeout,
		})
		if err != nil {
			return err
		}

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

		var data []*runtimev1.TimeSeriesValue
		rowMap := make(map[string]any)
		for rows.Next() {
			err := rows.MapScan(rowMap)
			if err != nil {
				rows.Close()
				return err
			}

			var t time.Time
			switch v := rowMap[tsAlias].(type) {
			case time.Time:
				t = v
			default:
				rows.Close()
				panic(fmt.Sprintf("unexpected type for timestamp column: %T", v))
			}
			delete(rowMap, tsAlias)

			records, err := pbutil.ToStruct(rowMap, schema)
			if err != nil {
				rows.Close()
				return err
			}

			tpb := timestamppb.New(t)
			if err := tpb.CheckValid(); err != nil {
				rows.Close()
				return err
			}

			data = append(data, &runtimev1.TimeSeriesValue{
				Ts:      tpb,
				Records: records,
			})
		}
		meta := structTypeToMetricsViewColumn(rows.Schema)
		rows.Close()

		var sparkValues []*runtimev1.TimeSeriesValue
		if q.Pixels != 0 {
			sparkValues, err = q.CreateTimestampRollupReduction(ctx, rt, olap, instanceID, priority, temporaryTableName, tsAlias, "count")
			if err != nil {
				return err
			}
		}

		q.Result = &ColumnTimeseriesResult{
			Meta:    meta,
			Results: data,
			Spark:   sparkValues,
		}
		return nil
	})
}

func (q *ColumnTimeseries) Export(ctx context.Context, rt *runtime.Runtime, instanceID string, w io.Writer, opts *runtime.ExportOptions) error {
	return ErrExportNotSupported
}

func (q *ColumnTimeseries) ResolveNormaliseTimeRange(ctx context.Context, rt *runtime.Runtime, instanceID string, priority int) (*runtimev1.TimeSeriesTimeRange, error) {
	rtr := q.TimeRange
	if rtr == nil {
		rtr = &runtimev1.TimeSeriesTimeRange{}
	}

	var result runtimev1.TimeSeriesTimeRange
	if rtr.Interval == runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED {
		q := &RollupInterval{
			TableName:  q.TableName,
			ColumnName: q.TimestampColumnName,
		}
		err := rt.Query(ctx, instanceID, q, priority)
		if err != nil {
			return nil, err
		}

		r := q.Result
		if r == nil || r.Interval == runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED {
			return &result, nil
		}

		result = runtimev1.TimeSeriesTimeRange{
			Interval: r.Interval,
			Start:    r.Start,
			End:      timestamppb.New(addInterval(r.End.AsTime(), r.Interval)),
		}
	} else if rtr.Start == nil || rtr.End == nil {
		q := &ColumnTimeRange{
			TableName:  q.TableName,
			ColumnName: q.TimestampColumnName,
		}
		err := rt.Query(ctx, instanceID, q, priority)
		if err != nil {
			return nil, err
		}

		tr := q.Result
		result = runtimev1.TimeSeriesTimeRange{
			Interval: rtr.Interval,
			Start:    tr.Min,
			End:      timestamppb.New(addInterval(tr.Max.AsTime(), rtr.Interval)),
		}
	}

	if rtr.Start != nil {
		result.Start = rtr.Start
	}

	if rtr.End != nil {
		result.End = rtr.End
	}

	if rtr.Interval != runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED {
		result.Interval = rtr.Interval
	}

	return &result, nil
}

/**
 * Contains an as-of-this-commit unpublished algorithm for an M4-like line density reduction.
 * This will take in an n-length time series and produce a pixels * 4 reduction of the time series
 * that preserves the shape and trends.
 *
 * This algorithm expects the source table to have a timestamp column and some kind of value column,
 * meaning it expects the data to essentially already be aggregated.
 *
 * It's important to note that this implemention is NOT the original M4 aggregation method, but a method
 * that has the same basic understanding but is much faster.
 *
 * Nonetheless, we mostly use this to reduce a many-thousands-point-long time series to about 120 * 4 pixels.
 * Importantly, this function runs very fast. For more information about the original M4 method,
 * see http://www.vldb.org/pvldb/vol7/p797-jugel.pdf
 */
func (q *ColumnTimeseries) CreateTimestampRollupReduction(
	ctx context.Context,
	rt *runtime.Runtime,
	olap drivers.OLAPStore,
	instanceID string,
	priority int,
	tableName string,
	timestampColumnName string,
	valueColumn string,
) ([]*runtimev1.TimeSeriesValue, error) {
	safeTimestampColumnName := safeName(timestampColumnName)

	rowCount, err := q.resolveRowCount(ctx, tableName, olap, priority)
	if err != nil {
		return nil, err
	}

	if rowCount < int64(q.Pixels*4) {
		rows, err := olap.Execute(ctx, &drivers.Statement{
			Query:            `SELECT ` + safeTimestampColumnName + ` as ts, "` + valueColumn + `"::DOUBLE as count FROM "` + tableName + `"`,
			Priority:         priority,
			ExecutionTimeout: defaultExecutionTimeout,
		})
		if err != nil {
			return nil, err
		}

		defer rows.Close()

		results := make([]*runtimev1.TimeSeriesValue, 0, (q.Pixels+1)*4)
		for rows.Next() {
			var ts time.Time
			var count sql.NullFloat64
			err = rows.Scan(&ts, &count)
			if err != nil {
				return nil, err
			}

			tsv := &runtimev1.TimeSeriesValue{
				Ts: timestamppb.New(ts),
				Records: &structpb.Struct{
					Fields: make(map[string]*structpb.Value),
				},
			}

			if count.Valid {
				tsv.Records.Fields["count"] = structpb.NewNumberValue(count.Float64)
			} else {
				tsv.Records.Fields["count"] = structpb.NewNullValue()
			}

			results = append(results, tsv)
		}

		return results, nil
	}

	querySQL := ` -- extract unix time
      WITH Q as (
        SELECT extract('epoch' from ` + safeTimestampColumnName + `)::BIGINT as t, "` + valueColumn + `"::DOUBLE as v FROM "` + tableName + `"
      ),
      -- generate bounds
      M as (
        SELECT min(t) as t1, max(t) as t2, max(t) - min(t) as diff FROM Q
      )
      -- core logic
      SELECT 
        -- left boundary point
        min(t) * 1000  as min_t, 
        arg_min(v, t) as argmin_tv, 

        -- right boundary point
        max(t) * 1000 as max_t, 
        arg_max(v, t) as argmax_tv,

        -- smallest point within boundary
        min(v) as min_v, 
        arg_min(t, v) * 1000  as argmin_vt,

        -- largest point within boundary
        max(v) as max_v, 
        arg_max(t, v) * 1000  as argmax_vt,

        round(` + strconv.FormatInt(int64(q.Pixels), 10) + ` * (t - (SELECT t1 FROM M)) / (SELECT diff FROM M)) AS bin
  
      FROM Q GROUP BY bin
      ORDER BY bin
    `

	rows, err := olap.Execute(ctx, &drivers.Statement{
		Query:            querySQL,
		Priority:         priority,
		ExecutionTimeout: defaultExecutionTimeout,
	})
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	toTSV := func(ts int64, value sql.NullFloat64, bin float64) *runtimev1.TimeSeriesValue {
		tsv := &runtimev1.TimeSeriesValue{
			Records: &structpb.Struct{
				Fields: make(map[string]*structpb.Value),
			},
		}
		tsv.Ts = timestamppb.New(time.UnixMilli(ts))
		tsv.Bin = bin
		if value.Valid {
			tsv.Records.Fields["count"] = structpb.NewNumberValue(value.Float64)
		} else {
			tsv.Records.Fields["count"] = structpb.NewNullValue()
		}
		return tsv
	}

	results := make([]*runtimev1.TimeSeriesValue, 0, (q.Pixels+1)*4)
	for rows.Next() {
		var minT, maxT int64
		var argminVT, argmaxVT sql.NullInt64
		var argminTV, argmaxTV, minV, maxV sql.NullFloat64
		var bin float64
		err = rows.Scan(&minT, &argminTV, &maxT, &argmaxTV, &minV, &argminVT, &maxV, &argmaxVT, &bin)
		if err != nil {
			return nil, err
		}

		argminVTSafe := minT
		if argminVT.Valid {
			argminVTSafe = argminVT.Int64
		}
		argmaxVTSafe := maxT
		if argmaxVT.Valid {
			argmaxVTSafe = argmaxVT.Int64
		}
		results = append(results, toTSV(minT, argminTV, bin), toTSV(argminVTSafe, minV, bin), toTSV(argmaxVTSafe, maxV, bin), toTSV(maxT, argmaxTV, bin))

		if argminVT.Int64 > argmaxVT.Int64 {
			i := len(results)
			results[i-3], results[i-2] = results[i-2], results[i-3]
		}
	}

	return results, nil
}

func (q *ColumnTimeseries) resolveRowCount(ctx context.Context, tableName string, olap drivers.OLAPStore, priority int) (int64, error) {
	rows, err := olap.Execute(ctx, &drivers.Statement{
		Query:    fmt.Sprintf("SELECT count(*) AS count FROM %s", safeName(tableName)),
		Priority: priority,
	})
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var count int64
	for rows.Next() {
		err = rows.Scan(&count)
		if err != nil {
			return 0, err
		}
	}

	err = rows.Err()
	if err != nil {
		return 0, err
	}

	return count, nil
}

// normaliseMeasures is called before this method so measure.SqlName will be non empty
func getExpressionColumnsFromMeasures(measures []*runtimev1.ColumnTimeSeriesRequest_BasicMeasure) string {
	var result string
	for i, measure := range measures {
		result += measure.Expression + " as " + safeName(measure.SqlName)
		if i < len(measures)-1 {
			result += ", "
		}
	}
	return result
}

// normaliseMeasures is called before this method so measure.SqlName will be non empty
func getCoalesceStatementsMeasures(measures []*runtimev1.ColumnTimeSeriesRequest_BasicMeasure) string {
	var result string
	for i, measure := range measures {
		result += fmt.Sprintf(`series.%[1]s as %[1]s`, safeName(measure.SqlName))
		if i < len(measures)-1 {
			result += ", "
		}
	}
	return result
}

func getCoalesceStatementsMeasuresLast(measures []*runtimev1.ColumnTimeSeriesRequest_BasicMeasure) string {
	var result string
	for i, measure := range measures {
		result += fmt.Sprintf(`last(%[1]s) as %[1]s`, safeName(measure.SqlName))
		if i < len(measures)-1 {
			result += ", "
		}
	}
	return result
}

func normaliseMeasures(measures []*runtimev1.ColumnTimeSeriesRequest_BasicMeasure, generateCount bool) []*runtimev1.ColumnTimeSeriesRequest_BasicMeasure {
	if len(measures) == 0 {
		return []*runtimev1.ColumnTimeSeriesRequest_BasicMeasure{
			{
				Expression: "count(*)",
				SqlName:    "count",
				Id:         "",
			},
		}
	}

	var countExists bool
	for i, measure := range measures {
		if measure.SqlName == "" {
			measure.SqlName = fmt.Sprintf("measure_%d", i)
		}

		if measure.SqlName == "count" {
			countExists = true
		}
	}

	if !countExists && generateCount {
		measures = append(measures, &runtimev1.ColumnTimeSeriesRequest_BasicMeasure{
			Expression: "count(*)",
			SqlName:    "count",
			Id:         "",
		})
	}

	return measures
}

func approxSize(c *ColumnTimeseriesResult) int64 {
	var size int64
	if len(c.Meta) > 0 {
		size += sizeProtoMessage(c.Meta[0]) * int64(len(c.Meta))
	}
	if len(c.Results) > 0 {
		size += sizeProtoMessage(c.Results[0]) * int64(len(c.Results))
	}
	if len(c.Spark) > 0 {
		size += sizeProtoMessage(c.Spark[0]) * int64(len(c.Spark))
	}
	size += sizeProtoMessage(c.TimeRange)
	size += int64(reflect.TypeOf(c.SampleSize).Size())
	return size
}
