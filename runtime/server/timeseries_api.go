package server

import (
	"context"
	"fmt"
	"strings"
	"time"

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

func (s *Server) GenerateTimeSeries(ctx context.Context, request *api.GenerateTimeSeriesRequest) (*api.TimeSeriesRollup, error) {
	var timeRange *api.TimeSeriesTimeRange = request.TimeRange
	var measures []*api.BasicMeasureDefinition = normaliseMeasures(request.Measures).BasicMeasures
	var timestampColumn string = request.TimestampColumnName
	var tableName string = request.TableName
	var filter string = getFilterFromMetricsViewFilters(request.Filters)
	var timeGranularity string = strings.Split(timeRange.Interval, " ")[1]
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
              interval '` + timeRange.Interval + `')
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
	if err != nil {
		rs, er := s.query(ctx, request.InstanceId, &drivers.Statement{
			Query: "DROP TABLE _ts_",
		})
		if er == nil {
			rs.Close()
		}
		return &api.TimeSeriesRollup{
			Rollup: &api.TimeSeriesResponse{
				Results:   []*api.TimeSeriesValue{},
				TimeRange: request.TimeRange, // todo return the generated time range
				// SampleSize: *request.SampleSize, todo
			}, // todo review
		}, err
	}
	rows.Close()
	rows, _ = s.query(ctx, request.InstanceId, &drivers.Statement{
		Query: "SELECT * from _ts_",
	})
	results := make([]*api.TimeSeriesValue, 0)
	for rows.Next() {
		value := api.TimeSeriesValue{}
		results = append(results, &value)
		row := make(map[string]interface{}, len(measures))
		rows.MapScan(row)
		value.Ts = row["ts"].(time.Time).Format("2006-01-02 15:04:05")
		delete(row, "ts")
		value.Records = make(map[string]float64, len(row))
		for k, v := range row {
			value.Records[k] = v.(float64)
		}
	}
	rows.Close()
	rows, _ = s.query(ctx, request.InstanceId, &drivers.Statement{
		Query: "DROP TABLE _ts_",
	})
	rows.Close()

	return &api.TimeSeriesRollup{
		Rollup: &api.TimeSeriesResponse{
			Results:   results,
			TimeRange: request.TimeRange, // todo return the generated time range
			// SampleSize: *request.SampleSize,
		},
	}, nil
}
