package queries

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/server/pbutil"
	"google.golang.org/protobuf/types/known/structpb"
)

const IsoFormat string = "2006-01-02T15:04:05.000Z"

type ColumnTimeseriesResult struct {
	Meta []*runtimev1.MetricsViewColumn
	Data *runtimev1.TimeSeriesResponse
}

type ColumnTimeseries struct {
	TableName           string                                              `json:"table_name"`
	Measures            []*runtimev1.GenerateTimeSeriesRequest_BasicMeasure `json:"measures"`
	TimestampColumnName string                                              `json:"timestamp_column_name"`
	TimeRange           *runtimev1.TimeSeriesTimeRange                      `json:"time_range"`
	Filters             *runtimev1.MetricsViewFilter                        `json:"filters"`
	Pixels              int32                                               `json:"pixels"`
	SampleSize          int32                                               `json:"sample_size"`
	Result              *ColumnTimeseriesResult                             `json:"-"`
}

var _ runtime.Query = &ColumnTimeseries{}

func (q *ColumnTimeseries) Key() string {
	r, err := json.Marshal(q)
	if err != nil {
		panic(err)
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
	res, ok := v.(*ColumnTimeseriesResult)
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

	return olap.WithConnection(ctx, priority, func(ctx context.Context, ensuredCtx context.Context) error {
		timeRange, err := q.resolveNormaliseTimeRange(ctx, rt, instanceID, priority)
		if err != nil {
			return err
		}

		if timeRange.Interval == runtimev1.TimeGrain_TIME_GRAIN_UNSPECIFIED {
			q.Result = &ColumnTimeseriesResult{
				Data: &runtimev1.TimeSeriesResponse{},
			}
			return nil
		}

		filter, args, err := buildFilterClauseForMetricsViewFilter(q.Filters)
		if err != nil {
			return err
		}
		if filter != "" {
			filter = "WHERE 1=1 " + filter
		}

		measures := normaliseMeasures(q.Measures, q.Pixels != 0)
		dateTruncSpecifier := convertToDateTruncSpecifier(timeRange.Interval)
		tsAlias := tempName("_ts_")
		temporaryTableName := tempName("_timeseries_")
		sql := `CREATE TEMPORARY TABLE ` + temporaryTableName + ` AS (
			-- generate a time series column that has the intended range
			WITH template as (
			SELECT
				generate_series as ` + tsAlias + `
			FROM
				generate_series(
				date_trunc('` + dateTruncSpecifier + `', TIMESTAMP '` + timeRange.Start.AsTime().Format(IsoFormat) + `'),
				date_trunc('` + dateTruncSpecifier + `', TIMESTAMP '` + timeRange.End.AsTime().Format(IsoFormat) + `'),
				interval '1 ` + dateTruncSpecifier + `')
			),
			-- transform the original data, and optionally sample it.
			series AS (
			SELECT
				date_trunc('` + dateTruncSpecifier + `', ` + safeName(q.TimestampColumnName) + `) as ` + tsAlias + `,` + getExpressionColumnsFromMeasures(measures) + `
			FROM ` + safeName(q.TableName) + ` ` + filter + `
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

		err = olap.Exec(ctx, &drivers.Statement{
			Query:    sql,
			Args:     args,
			Priority: priority,
		})
		if err != nil {
			return err
		}
		defer func() {
			// NOTE: Using ensuredCtx
			_ = olap.Exec(ensuredCtx, &drivers.Statement{
				Query:    `DROP TABLE "` + temporaryTableName + `"`,
				Priority: priority,
			})
		}()

		rows, err := olap.Execute(ctx, &drivers.Statement{
			Query:    fmt.Sprintf(`SELECT %s as ts, * EXCLUDE(%s) FROM %s`, tsAlias, tsAlias, temporaryTableName),
			Priority: priority,
		})
		if err != nil {
			return err
		}

		results, err := rowsToData(rows)
		meta := structTypeToMetricsViewColumn(rows.Schema)
		rows.Close()
		if err != nil {
			return err
		}

		var sparkValues []*structpb.Struct
		if q.Pixels != 0 {
			sparkValues, err = q.createTimestampRollupReduction(ctx, rt, olap, instanceID, priority, temporaryTableName, tsAlias, "count")
			if err != nil {
				return err
			}
		}

		q.Result = &ColumnTimeseriesResult{
			Meta: meta,
			Data: &runtimev1.TimeSeriesResponse{
				Results:   results,
				TimeRange: timeRange,
				Spark:     sparkValues,
			},
		}
		return nil
	})
}

func (q *ColumnTimeseries) resolveNormaliseTimeRange(ctx context.Context, rt *runtime.Runtime, instanceID string, priority int) (*runtimev1.TimeSeriesTimeRange, error) {
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
) ([]*structpb.Struct, error) {
	safeTimestampColumnName := safeName(timestampColumnName)
	tc := &TableCardinality{
		TableName: tableName,
	}
	err := tc.Resolve(ctx, rt, instanceID, priority)
	if err != nil {
		return nil, err
	}

	if tc.Result < int64(q.Pixels*4) {
		rows, err := olap.Execute(ctx, &drivers.Statement{
			Query:    `SELECT ` + safeTimestampColumnName + ` as ts, "` + valueColumn + `" as count FROM "` + tableName + `"`,
			Priority: priority,
		})
		if err != nil {
			return nil, err
		}

		defer rows.Close()
		results, err := rowsToData(rows)
		if err != nil {
			return nil, err
		}

		return results, nil
	}

	sql := ` -- extract unix time
      WITH Q as (
        SELECT extract('epoch' from ` + safeTimestampColumnName + `) as t, "` + valueColumn + `" as v FROM "` + tableName + `"
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
		Priority: priority,
	})
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	aggs, err := rowsToData(rows)
	if err != nil {
		return nil, err
	}

	results := make([]*structpb.Struct, 0, len(aggs)*4)
	for _, v := range aggs {
		addStruct(v, &results, "min_t", "argmin_tv")
		addStruct(v, &results, "argmin_vt", "min_v")
		addStruct(v, &results, "argmax_vt", "max_v")
		addStruct(v, &results, "max_t", "argmax_tv")
		if v.Fields["argmin_vt"].GetNumberValue() > v.Fields["argmax_vt"].GetNumberValue() {
			i := len(results)
			results[i-3], results[i-2] = results[i-2], results[i-3]
		}
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return results, nil
}

func addStruct(v *structpb.Struct, results *[]*structpb.Struct, key, value string) {
	s := &structpb.Struct{
		Fields: make(map[string]*structpb.Value, 3),
	}

	ts, err := pbutil.ToValue(time.UnixMilli(int64(v.Fields[key].GetNumberValue())))
	if err != nil {
		panic(err)
	}

	s.Fields["ts"] = ts
	s.Fields["count"] = v.Fields[value]
	s.Fields["bin"] = v.Fields["bin"]
	*results = append(*results, s)
}

// normaliseMeasures is called before this method so measure.SqlName will be non empty
func getExpressionColumnsFromMeasures(measures []*runtimev1.GenerateTimeSeriesRequest_BasicMeasure) string {
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
func getCoalesceStatementsMeasures(measures []*runtimev1.GenerateTimeSeriesRequest_BasicMeasure) string {
	var result string
	for i, measure := range measures {
		result += fmt.Sprintf(`series.%s as %s`, safeName(measure.SqlName), safeName(measure.SqlName))
		if i < len(measures)-1 {
			result += ", "
		}
	}
	return result
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
		if measure.SqlName == "" {
			measure.SqlName = fmt.Sprintf("measure_%d", i)
		}

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
