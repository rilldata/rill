package server

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/marcboeker/go-duckdb"
	"github.com/rilldata/rill/runtime/api"
	"github.com/rilldata/rill/runtime/drivers"
)

func getExpressionColumnsFromMeasures(measures []*api.BasicMeasureDefinition) string {
	var result string
	for i, measure := range measures {
		result += measure.Expression + " as " + *measure.SqlName
		if i < len(measures)-1 {
			result += ", "
		}
	}
	return result
}

func getCoalesceStatementsMeasures(measures []*api.BasicMeasureDefinition) string {
	var result string
	for i, measure := range measures {
		result += fmt.Sprintf(`COALESCE(series.%s, 0) as %s`, *measure.SqlName, *measure.SqlName)
		if i < len(measures)-1 {
			result += ", "
		}
	}
	return result
}

func EscapeSingleQuotes(value string) string {
	return strings.ReplaceAll(value, "'", "''")
}

func getFilterFromDimensionValuesFilter(
	dimensionValues []*api.MetricsViewDimensionValue,
	prefix string,
	dimensionJoiner string,
) string {
	if len(dimensionValues) == 0 {
		return ""
	}
	var result string
	conditions := make([]string, 3)
	for i, dv := range dimensionValues {
		escapedName := EscapeSingleQuotes(dv.Name)
		var nulls bool
		var notNulls bool
		for _, iv := range dv.In {
			if iv.GetKind() == nil {
				nulls = true
			} else {
				notNulls = true
			}
		}
		conditions = conditions[:0]
		if notNulls {
			var inClause string = escapedName + " " + prefix + " IN ("
			for j, iv := range dv.In {
				if iv.GetKind() != nil {
					inClause += "'" + EscapeSingleQuotes(iv.GetStringValue()) + "'"
				}
				if j < len(dv.In)-1 {
					inClause += ", "
				}
			}
			inClause += ")"
			conditions = append(conditions, inClause)
		}
		if nulls {
			var nullClause = escapedName + " IS " + prefix + " NULL"
			conditions = append(conditions, nullClause)
		}
		if dv.Like != nil {
			var likeClause string
			for j, lv := range dv.Like.Values {
				if lv.GetKind() == nil {
					continue
				}
				likeClause += escapedName + " " + prefix + " ILIKE '" + EscapeSingleQuotes(lv.GetStringValue()) + "'"
				if j < len(dv.Like.Values)-1 {
					likeClause += " AND "
				}
			}
			conditions = append(conditions, likeClause)
		}
		result += strings.Join(conditions, " "+dimensionJoiner+" ")
		if i < len(dimensionValues)-1 {
			result += " AND "
		}
	}

	return result
}

func getFilterFromMetricsViewFilters(filters *api.MetricsViewRequestFilter) string {
	includeFilters := getFilterFromDimensionValuesFilter(filters.Include, "", "OR")

	excludeFilters := getFilterFromDimensionValuesFilter(filters.Exclude, "NOT", "AND")
	if includeFilters != "" && excludeFilters != "" {
		return includeFilters + " AND " + excludeFilters
	} else if includeFilters != "" {
		return includeFilters
	} else if excludeFilters != "" {
		return excludeFilters
	} else {
		return ""
	}
}

func getFallbackMeasureName(index int, sqlName *string) *string {
	if sqlName == nil || *sqlName == "" {
		s := fmt.Sprintf("measure_%d", index)
		return &s
	} else {
		return sqlName
	}
}

var countName string = "count"

func normaliseMeasures(measures *api.GenerateTimeSeriesRequest_BasicMeasures) *api.GenerateTimeSeriesRequest_BasicMeasures {
	if measures == nil {
		return &api.GenerateTimeSeriesRequest_BasicMeasures{
			BasicMeasures: []*api.BasicMeasureDefinition{
				{
					Expression: "count(*)",
					SqlName:    &countName,
					Id:         "",
				},
			},
		}
	}
	for i, measure := range measures.BasicMeasures {
		measure.SqlName = getFallbackMeasureName(i, measure.SqlName)
	}
	return measures
}

func (s *Server) normaliseTimeRange(ctx context.Context, request *api.GenerateTimeSeriesRequest) (*api.TimeSeriesTimeRange, error) {
	tableName := EscapeDoubleQuotes(request.TableName)
	escapedColumnName := EscapeDoubleQuotes(request.TimestampColumnName)
	rows, err := s.query(ctx, request.InstanceId, &drivers.Statement{
		Query: `SELECT
        	max(` + escapedColumnName + `) - min(` + escapedColumnName + `) as r,
        	max(` + escapedColumnName + `) as max_value,
        	min(` + escapedColumnName + `) as min_value,
        	count(*) as count
        	from ` + tableName,
	})
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	rows.Next()
	var r duckdb.Interval
	var max, min time.Time
	var count int64
	err = rows.Scan(&r, &max, &min, &count)
	if err != nil {
		return nil, err
	}

	const (
		MICROS_SECOND = 1000 * 1000
		MICROS_MINUTE = 1000 * 1000 * 60
		MICROS_HOUR   = 1000 * 1000 * 60 * 60
		MICROS_DAY    = 1000 * 1000 * 60 * 60 * 24
	)

	var rollupInterval api.TimeGrain
	if r.Days == 0 && r.Micros <= MICROS_MINUTE {
		rollupInterval = api.TimeGrain_MILLISECOND
	} else if r.Days == 0 && r.Micros > MICROS_MINUTE && r.Micros <= MICROS_HOUR {
		rollupInterval = api.TimeGrain_SECOND
	} else if r.Days == 0 && r.Micros <= MICROS_DAY {
		rollupInterval = api.TimeGrain_MINUTE
	} else if r.Days <= 7 {
		rollupInterval = api.TimeGrain_HOUR
	} else if r.Days <= 365*20 {
		rollupInterval = api.TimeGrain_DAY
	} else if r.Days <= 365*500 {
		rollupInterval = api.TimeGrain_MONTH
	} else {
		rollupInterval = api.TimeGrain_YEAR
	}

	start := min.Format("2006-01-02 15:04:05")
	end := max.Format("2006-01-02 15:04:05") // todo iso format

	rtr := request.TimeRange
	if rtr == nil {
		rtr = &api.TimeSeriesTimeRange{
			Start: start,
			End:   end,
		}
	}
	if rtr.Start == "" {
		rtr.Start = start
	}
	if rtr.End == "" {
		rtr.End = end
	}
	if rtr.Interval == api.TimeGrain_UNSPECIFIED {
		rtr.Interval = rollupInterval
	}
	return rtr, nil
}

func (s *Server) GenerateTimeSeries(ctx context.Context, request *api.GenerateTimeSeriesRequest) (*api.TimeSeriesRollup, error) {
	timeRange, err := s.normaliseTimeRange(ctx, request)
	if err != nil {
		return createErrResult(request.TimeRange), err
	}
	var measures []*api.BasicMeasureDefinition = normaliseMeasures(request.Measures).BasicMeasures
	var timestampColumn string = request.TimestampColumnName
	var tableName string = request.TableName
	var filter string = getFilterFromMetricsViewFilters(request.Filters)
	var timeGranularity string = timeRange.Interval.Enum().String()
	var tsAlias string
	if timestampColumn == "ts" {
		tsAlias = "_ts"
	} else {
		tsAlias = "ts"
	}
	if filter != "" {
		filter = "WHERE " + filter
	}
	sql := `CREATE TEMPORARY TABLE _ts_ AS (
        -- generate a time series column that has the intended range
        WITH template as (
          SELECT 
            generate_series as ` + tsAlias + `
          FROM 
            generate_series(
              date_trunc('` + timeGranularity + `', TIMESTAMP '` + timeRange.Start + `'), 
              date_trunc('` + timeGranularity + `', TIMESTAMP '` + timeRange.End + `'),
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
          template.` + tsAlias + ` as ts from template
        LEFT OUTER JOIN series ON template.` + tsAlias + ` = series.` + tsAlias + `
        ORDER BY template.` + tsAlias + `
      )`
	rows, err := s.query(ctx, request.InstanceId, &drivers.Statement{
		Query: sql,
	})
	defer s.dropTempTable(ctx, request.InstanceId)
	if err != nil {
		return createErrResult(timeRange), err
	}
	rows.Close()
	rows, err = s.query(ctx, request.InstanceId, &drivers.Statement{
		Query: "SELECT * from _ts_",
	})
	if err != nil {
		return createErrResult(timeRange), err
	}
	results, err := convertRowsToTimeSeriesValues(rows, len(measures)+1)
	if err != nil {
		return createErrResultWithPartial(timeRange, results), err
	}

	return &api.TimeSeriesRollup{
		Rollup: &api.TimeSeriesResponse{
			Results:   results,
			TimeRange: timeRange,
		},
	}, nil
}

func convertRowsToTimeSeriesValues(rows *drivers.Result, rowLength int) ([]*api.TimeSeriesValue, error) {
	results := make([]*api.TimeSeriesValue, 0)
	for rows.Next() {
		value := api.TimeSeriesValue{}
		results = append(results, &value)
		row := make(map[string]interface{}, rowLength)
		err := rows.MapScan(row)
		if err != nil {
			return results, err
		}
		value.Ts = row["ts"].(time.Time).Format("2006-01-02 15:04:05")
		delete(row, "ts")
		value.Records = make(map[string]float64, len(row))
		for k, v := range row {
			value.Records[k] = v.(float64)
		}
	}
	rows.Close()
	return results, nil
}

func createErrResult(timeRange *api.TimeSeriesTimeRange) *api.TimeSeriesRollup {
	return &api.TimeSeriesRollup{
		Rollup: &api.TimeSeriesResponse{
			Results:   []*api.TimeSeriesValue{},
			TimeRange: timeRange,
		},
	}
}

func createErrResultWithPartial(timeRange *api.TimeSeriesTimeRange, results []*api.TimeSeriesValue) *api.TimeSeriesRollup {
	return &api.TimeSeriesRollup{
		Rollup: &api.TimeSeriesResponse{
			Results:   results,
			TimeRange: timeRange,
		},
	}
}

func (s *Server) dropTempTable(ctx context.Context, instanceId string) {
	rs, er := s.query(ctx, instanceId, &drivers.Statement{
		Query: "DROP TABLE _ts_",
	})
	if er == nil {
		rs.Close()
	}
}
