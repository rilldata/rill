package queries

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	structpb "github.com/golang/protobuf/ptypes/struct"
	"github.com/google/uuid"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
)

type ColumnTimeseries struct {
	TableName           string                                              `json:"table_name"`
	Measures            []*runtimev1.GenerateTimeSeriesRequest_BasicMeasure `json:"measures"`
	TimestampColumnName string                                              `json:"timestamp_column_name"`
	TimeRange           *runtimev1.TimeSeriesTimeRange                      `json:"time_range"`
	Filters             *runtimev1.MetricsViewRequestFilter                 `json:"filters"`
	Pixels              int32                                               `json:"pixels"`
	SampleSize          int32                                               `json:"sample_size"`
	Result              *runtimev1.TimeSeriesResponse                       `json:"-"`
}

var _ runtime.Query = &ColumnTimeseries{}

func (q *ColumnTimeseries) Key() string {
	r, err := json.Marshal(q)
	if err != nil {
		panic(fmt.Errorf("ColumnTimeseries: failed to marshal: %w", err))
	}
	return fmt.Sprintf("ColumnTimeseries:%s", string(r))
}

func (q *ColumnTimeseries) Deps() []string {
	return []string{q.TableName}
}

func (q *ColumnTimeseries) MarshalResult() any {
	return q.Result
}

func (q *ColumnTimeseries) UnmarshalResult(v any) error {
	res, ok := v.(*runtimev1.TimeSeriesResponse)
	if !ok {
		return fmt.Errorf("ColumnTimeseries: mismatched unmarshal input")
	}
	q.Result = res
	return nil
}

func (q *ColumnTimeseries) Resolve(ctx context.Context, rt *runtime.Runtime, instanceID string, priority int) error {
	olap, err := rt.OLAP(ctx, instanceID)
	if err != nil {
		return err
	}

	if olap.Dialect() != drivers.DialectDuckDB {
		return fmt.Errorf("not available for dialect '%s'", olap.Dialect())
	}

	timeRange, err := q.normaliseTimeRange(ctx, rt, instanceID, priority)
	if err != nil {
		return err
	}
	if timeRange.Interval == runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED {
		q.Result = &runtimev1.TimeSeriesResponse{}
		return nil
	}
	var measures = normaliseMeasures(q.Measures, true)
	var timestampColumn = q.TimestampColumnName
	var tableName = q.TableName
	var filter string
	if q.Filters != nil {
		filter = getFilterFromMetricsViewFilters(q.Filters)
	}
	var timeGranularity = convertToDateTruncSpecifier(timeRange.Interval)
	tsAlias := "_ts_" + ReplaceHyphen(uuid.New().String())
	if filter != "" {
		filter = "WHERE " + filter
	}
	temporaryTableName := "_timeseries_" + uuid.New().String()
	sql := `CREATE TEMPORARY TABLE "` + temporaryTableName + `" AS (
        -- generate a time series column that has the intended range
        WITH template as (
          SELECT 
            generate_series as ` + tsAlias + `
          FROM 
            generate_series(
              date_trunc('` + timeGranularity + `', TIMESTAMP '` + timeRange.Start.AsTime().Format(IsoFormat) + `'),
              date_trunc('` + timeGranularity + `', TIMESTAMP '` + timeRange.End.AsTime().Format(IsoFormat) + `'),
              interval '1 ` + timeGranularity + `')
        ),
        -- transform the original data, and optionally sample it.
        series AS (
          SELECT 
            date_trunc('` + timeGranularity + `', "` + EscapeDoubleQuotes(timestampColumn) + `") as ` + tsAlias + `,` + getExpressionColumnsFromMeasures(measures) + `
          FROM "` + EscapeDoubleQuotes(tableName) + `" ` + filter + `
          GROUP BY ` + tsAlias + ` ORDER BY ` + tsAlias + `
        )
        -- join the transformed data with the generated time series column,
        -- coalescing the first value to get the 0-default when the rolled up data
        -- does not have that value.
        SELECT 
          ` + getCoalesceStatementsMeasures(measures) + `,
          template.` + tsAlias + ` from template
        LEFT OUTER JOIN series ON template.` + tsAlias + ` = series.` + tsAlias + `
        ORDER BY template.` + tsAlias + `
      )`

	rows, err := olap.Execute(ctx, &drivers.Statement{
		Query:    sql,
		Priority: priority,
	})
	defer DropTempTable(olap, priority, temporaryTableName)
	if err != nil {
		return err
	}
	rows.Close()
	rows, err = olap.Execute(ctx, &drivers.Statement{
		Query:    `SELECT * from "` + temporaryTableName + `"`,
		Priority: priority,
	})
	if err != nil {
		return err
	}
	results, err := convertRowsToTimeSeriesValues(rows, len(measures)+1, tsAlias)
	if err != nil {
		return err
	}
	var sparkValues []*runtimev1.TimeSeriesValue
	if q.Pixels != 0 {
		sparkValues, err = q.createTimestampRollupReduction(ctx, rt, olap, instanceID, priority, temporaryTableName, tsAlias, "count")
		if err != nil {
			return err
		}
	}

	q.Result = &runtimev1.TimeSeriesResponse{
		Results:   results,
		TimeRange: timeRange,
		Spark:     sparkValues,
	}
	return nil
}

func (q *ColumnTimeseries) normaliseTimeRange(ctx context.Context, rt *runtime.Runtime, instanceID string, priority int) (*runtimev1.TimeSeriesTimeRange, error) {
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
			End:      r.End,
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
			End:      tr.Max,
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
func (q *ColumnTimeseries) createTimestampRollupReduction(
	ctx context.Context,
	rt *runtime.Runtime,
	olap drivers.OLAPStore,
	instanceID string,
	priority int,
	tableName string,
	timestampColumnName string,
	valueColumn string,
) ([]*runtimev1.TimeSeriesValue, error) {
	escapedTimestampColumn := EscapeDoubleQuotes(timestampColumnName)
	tc := &TableCardinality{
		TableName: tableName,
	}
	err := tc.Resolve(ctx, rt, instanceID, int(priority))
	if err != nil {
		return nil, err
	}

	if tc.Result < int64(q.Pixels*4) {
		rows, err := olap.Execute(ctx, &drivers.Statement{
			Query:    `SELECT ` + escapedTimestampColumn + ` as ts, "` + valueColumn + `" as count FROM "` + tableName + `"`,
			Priority: int(priority),
		})
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		results := make([]*runtimev1.TimeSeriesValue, 0, (q.Pixels+1)*4)
		for rows.Next() {
			var ts time.Time
			var count float64
			err = rows.Scan(&ts, &count)
			if err != nil {
				return nil, err
			}
			results = append(results, &runtimev1.TimeSeriesValue{
				Ts:      ts.Format(IsoFormat),
				Records: sMap("count", count),
			})
		}
		return results, nil
	}

	sql := ` -- extract unix time
      WITH Q as (
        SELECT extract('epoch' from ` + escapedTimestampColumn + `) as t, "` + valueColumn + `" as v FROM "` + tableName + `"
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
		Query:    sql,
		Priority: int(priority),
	})
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	results := make([]*runtimev1.TimeSeriesValue, 0, (q.Pixels+1)*4)
	for rows.Next() {
		var minT, maxT, argminVT, argmaxVT int64
		var argminTV, argmaxTV, minV, maxV float64
		var bin float64
		err = rows.Scan(&minT, &argminTV, &maxT, &argmaxTV, &minV, &argminVT, &maxV, &argmaxVT, &bin)
		if err != nil {
			return nil, err
		}
		results = append(results, &runtimev1.TimeSeriesValue{
			Ts:      time.UnixMilli(minT).Format(IsoFormat),
			Bin:     &bin,
			Records: sMap("count", argminTV),
		})
		results = append(results, &runtimev1.TimeSeriesValue{
			Ts:      time.UnixMilli(argminVT).Format(IsoFormat),
			Bin:     &bin,
			Records: sMap("count", minV),
		})

		results = append(results, &runtimev1.TimeSeriesValue{
			Ts:      time.UnixMilli(argmaxVT).Format(IsoFormat),
			Bin:     &bin,
			Records: sMap("count", maxV),
		})

		results = append(results, &runtimev1.TimeSeriesValue{
			Ts:      time.UnixMilli(maxT).Format(IsoFormat),
			Bin:     &bin,
			Records: sMap("count", argmaxTV),
		})
		if argminVT > argmaxVT {
			i := len(results)
			results[i-3], results[i-2] = results[i-2], results[i-3]
		}
	}
	return results, nil
}

// normaliseMeasures is called before this method so measure.SqlName will be non empty
func getExpressionColumnsFromMeasures(measures []*runtimev1.GenerateTimeSeriesRequest_BasicMeasure) string {
	var result string
	for i, measure := range measures {
		result += measure.Expression + " as " + measure.SqlName
		if i < len(measures)-1 {
			result += ", "
		}
	}
	return result
}

// normaliseMeasures is called before this method so measure.SqlName will be non empty
func getCoalesceStatementsMeasures(measures []*runtimev1.GenerateTimeSeriesRequest_BasicMeasure) string {
	var result string
	for i, measure := range measures {
		result += fmt.Sprintf(`COALESCE(series.%s, 0) as %s`, measure.SqlName, measure.SqlName)
		if i < len(measures)-1 {
			result += ", "
		}
	}
	return result
}

func getFilterFromDimensionValuesFilter(
	dimensionValues []*runtimev1.MetricsViewDimensionValue,
	prefix string,
	dimensionJoiner string,
) string {
	if len(dimensionValues) == 0 {
		return ""
	}
	var result string
	conditions := make([]string, 3)
	if len(dimensionValues) > 0 {
		result += " ( "
	}
	for i, dv := range dimensionValues {
		escapedName := EscapeSingleQuotes(dv.Name)
		var nulls bool
		var notNulls bool
		for _, iv := range dv.In {
			if _, ok := iv.Kind.(*structpb.Value_NullValue); ok {
				nulls = true
			} else {
				notNulls = true
			}
		}
		conditions = conditions[:0]
		if notNulls {
			var inClause = escapedName + " " + prefix + " IN ("
			for j, iv := range dv.In {
				if _, ok := iv.Kind.(*structpb.Value_NullValue); !ok {
					inClause += "'" + EscapeSingleQuotes(iv.GetStringValue()) + "'"
					if j < len(dv.In)-1 {
						inClause += ", "
					}
				}
			}
			inClause += ")"
			conditions = append(conditions, inClause)
		}
		if nulls {
			var nullClause = escapedName + " IS " + prefix + " NULL"
			conditions = append(conditions, nullClause)
		}
		if len(dv.Like) > 0 {
			var likeClause string
			for j, lv := range dv.Like {
				if lv.GetKind() == nil {
					continue
				}
				likeClause += escapedName + " " + prefix + " ILIKE '" + EscapeSingleQuotes(lv.GetStringValue()) + "'"
				if j < len(dv.Like)-1 {
					likeClause += " OR "
				}
			}
			conditions = append(conditions, likeClause)
		}
		result += strings.Join(conditions, " "+dimensionJoiner+" ")
		if i < len(dimensionValues)-1 {
			result += ") AND ("
		}
	}
	result += " ) "

	return result
}

func getFilterFromMetricsViewFilters(filters *runtimev1.MetricsViewRequestFilter) string {
	includeFilters := getFilterFromDimensionValuesFilter(filters.Include, "", "OR")
	excludeFilters := getFilterFromDimensionValuesFilter(filters.Exclude, "NOT", "AND")
	if includeFilters != "" && excludeFilters != "" {
		return " ( " + includeFilters + ") AND (" + excludeFilters + ")"
	} else if includeFilters != "" {
		return includeFilters
	} else if excludeFilters != "" {
		return excludeFilters
	} else {
		return ""
	}
}

func getFallbackMeasureName(index int, sqlName string) string {
	if sqlName == "" {
		s := fmt.Sprintf("measure_%d", index)
		return s
	} else {
		return sqlName
	}
}

func normaliseMeasures(measures []*runtimev1.GenerateTimeSeriesRequest_BasicMeasure, generateCount bool) []*runtimev1.GenerateTimeSeriesRequest_BasicMeasure {
	if len(measures) == 0 {
		return []*runtimev1.GenerateTimeSeriesRequest_BasicMeasure{
			{
				Expression: "count(*)",
				SqlName:    "count",
				Id:         "",
			},
		}
	}
	var countExists bool
	for i, measure := range measures {
		measure.SqlName = getFallbackMeasureName(i, measure.SqlName)
		if measure.SqlName == "count" {
			countExists = true
		}
	}
	if !countExists && generateCount {
		measures = append(measures, &runtimev1.GenerateTimeSeriesRequest_BasicMeasure{
			Expression: "count(*)",
			SqlName:    "count",
			Id:         "",
		})
	}
	return measures
}

const IsoFormat string = "2006-01-02T15:04:05.000Z"

func sMap(k string, v float64) map[string]float64 {
	m := make(map[string]float64, 1)
	m[k] = v
	return m
}

func convertToDateTruncSpecifier(specifier runtimev1.TimeGrain) string {
	switch specifier {
	case runtimev1.TimeGrain_TIME_GRAIN_MILLISECOND:
		return "MILLISECOND"
	case runtimev1.TimeGrain_TIME_GRAIN_SECOND:
		return "SECOND"
	case runtimev1.TimeGrain_TIME_GRAIN_MINUTE:
		return "MINUTE"
	case runtimev1.TimeGrain_TIME_GRAIN_HOUR:
		return "HOUR"
	case runtimev1.TimeGrain_TIME_GRAIN_DAY:
		return "DAY"
	case runtimev1.TimeGrain_TIME_GRAIN_WEEK:
		return "WEEK"
	case runtimev1.TimeGrain_TIME_GRAIN_MONTH:
		return "MONTH"
	case runtimev1.TimeGrain_TIME_GRAIN_YEAR:
		return "YEAR"
	}
	panic(fmt.Errorf("unconvertable time grain specifier: %v", specifier))
}

func convertRowsToTimeSeriesValues(rows *drivers.Result, rowLength int, tsAlias string) ([]*runtimev1.TimeSeriesValue, error) {
	results := make([]*runtimev1.TimeSeriesValue, 0)
	defer rows.Close()
	var converr error
	for rows.Next() {
		value := runtimev1.TimeSeriesValue{}
		results = append(results, &value)
		row := make(map[string]interface{}, rowLength)
		err := rows.MapScan(row)
		if err != nil {
			return results, err
		}
		value.Ts = row[tsAlias].(time.Time).Format(IsoFormat)
		value.Records = make(map[string]float64, len(row))
		for k, v := range row {
			if k == tsAlias {
				continue
			}
			switch x := v.(type) {
			case int32:
				value.Records[k] = float64(x)
			case int64:
				value.Records[k] = float64(x)
			case float32:
				value.Records[k] = float64(x)
			case float64:
				value.Records[k] = x
			default:
				return nil, fmt.Errorf("unknown type %T ", v)
			}
		}
	}
	return results, converr
}
