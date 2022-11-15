package server

import (
	"context"
	"fmt"
)

func (s *Server) GenerateTimeseries(ctx context.Context) {
	var tsAlias string
	var timeGranularity string
	var timeRangeStart string
	var timeRangeEnd string
	var timeRangeInterval string
	var measures string
	var timestampColumn string
	var tableName string
	var filter string
	if timestampColumn == "ts" {
		tsAlias = "_ts"
	} else {
		tsAlias = "ts"
	}
	sql := `CREATE TEMPORARY TABLE _ts_ AS (
        -- generate a time series column that has the intended range
        WITH template as (
          SELECT 
            generate_series as ` + tsAlias + `
          FROM 
            generate_series(
              date_trunc('` +
		timeGranularity + `', 
                TIMESTAMP '` + timeRangeStart + `'
              ), 
              date_trunc('` +
		timeGranularity + `', 
                TIMESTAMP '` + timeRangeEnd + `'
              ),
              interval '` + timeRangeInterval + `')
        ),
        -- transform the original data, and optionally sample it.
        series AS (
          SELECT 
            date_trunc('` + timeGranularity + `', ` + EscapeColumn(timestampColumn) + `)}) as ` + tsAlias + `,` + getExpressionColumnsFromMeasures(measures) + `
          FROM "` + tableName + `" ` + filter + `
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
	fmt.Println(sql)
}
