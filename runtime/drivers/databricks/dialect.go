package databricks

import (
	"fmt"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/timeutil"
)

type dialect struct {
	drivers.BaseDialect
}

var DialectDatabricks drivers.Dialect = func() drivers.Dialect {
	d := &dialect{}
	d.BaseDialect = drivers.NewBaseDialect(drivers.DialectNameDatabricks, DatabricksEscapeIdentifier, DatabricksEscapeIdentifier)
	return d
}()

func DatabricksEscapeIdentifier(ident string) string {
	if ident == "" {
		return ident
	}
	// Databricks uses backticks for quoting identifiers
	// Replace any backticks inside the identifier with double backticks
	return fmt.Sprintf("`%s`", strings.ReplaceAll(ident, "`", "``"))
}

func (d *dialect) GetTimeDimensionParameter(typeCode runtimev1.Type_Code) string {
	return "?"
}

func (d *dialect) SafeDivideExpression(numExpr, denExpr string) string {
	return fmt.Sprintf("TRY_DIVIDE(%s, CAST(%s AS DOUBLE))", numExpr, denExpr)
}

func (d *dialect) OrderByExpression(name string, desc bool) string {
	res := d.EscapeIdentifier(name)
	if desc {
		res += " DESC"
	}
	res += " NULLS LAST"
	return res
}

func (d *dialect) OrderByAliasExpression(name string, desc bool) string {
	res := d.EscapeAlias(name)
	if desc {
		res += " DESC"
	}
	res += " NULLS LAST"
	return res
}

func (d *dialect) DimensionSelect(escapeTable string, dim *runtimev1.MetricsViewSpec_Dimension) (dimSelect, unnestClause string, err error) {
	colName := d.EscapeIdentifier(dim.Name)
	alias := d.EscapeAlias(dim.Name)
	if !dim.Unnest {
		expr, err := d.MetricsViewDimensionExpression(dim)
		if err != nil {
			return "", "", fmt.Errorf("failed to get dimension expression: %w", err)
		}
		return fmt.Sprintf(`(%s) AS %s`, expr, alias), "", nil
	}

	unnestColName := d.EscapeIdentifier(drivers.TempName(fmt.Sprintf("%s_%s_", "unnested", dim.Name)))
	unnestTableName := drivers.TempName("tbl")
	sel := fmt.Sprintf(`%s AS %s`, unnestColName, alias)
	if dim.Expression == "" {
		return sel, fmt.Sprintf(` LATERAL VIEW EXPLODE(%s.%s) %s AS %s`, escapeTable, colName, unnestTableName, unnestColName), nil
	}
	return sel, fmt.Sprintf(` LATERAL VIEW EXPLODE(%s) %s AS %s`, dim.Expression, unnestTableName, unnestColName), nil
}

func (d *dialect) LateralUnnest(expr, tableAlias, colName string) (tbl string, tupleStyle, auto bool, err error) {
	return fmt.Sprintf(`LATERAL VIEW EXPLODE(%s) %s AS %s`, expr, tableAlias, d.EscapeIdentifier(colName)), false, false, nil
}

func (d *dialect) UnnestSQLSuffix(tbl string) string {
	return fmt.Sprintf(" %s", tbl)
}

func (d *dialect) DateTruncExpr(dim *runtimev1.MetricsViewSpec_Dimension, grain runtimev1.TimeGrain, tz string, firstDayOfWeek, firstMonthOfYear int) (string, error) {
	if tz == "UTC" || tz == "Etc/UTC" {
		tz = ""
	}
	if tz != "" {
		_, err := time.LoadLocation(tz)
		if err != nil {
			return "", fmt.Errorf("invalid time zone %q: %w", tz, err)
		}
	}

	specifier := d.ConvertToDateTruncSpecifier(grain)

	var expr string
	if dim.Expression != "" {
		expr = fmt.Sprintf("(%s)", dim.Expression)
	} else {
		expr = d.EscapeIdentifier(dim.Column)
	}

	// For DATE columns, cast to TIMESTAMP so the result is always TIMESTAMP
	if dim.DataType != nil && dim.DataType.Code == runtimev1.Type_CODE_DATE {
		expr = fmt.Sprintf("CAST(%s AS TIMESTAMP)", expr)
	}

	if tz == "" {
		if grain == runtimev1.TimeGrain_TIME_GRAIN_WEEK && firstDayOfWeek > 1 {
			offset := 8 - firstDayOfWeek
			return fmt.Sprintf("DATEADD(DAY, -%d, DATE_TRUNC('%s', DATEADD(DAY, %d, %s)))", offset, specifier, offset, expr), nil
		}
		if grain == runtimev1.TimeGrain_TIME_GRAIN_YEAR && firstMonthOfYear > 1 {
			offset := 13 - firstMonthOfYear
			return fmt.Sprintf("ADD_MONTHS(DATE_TRUNC('%s', ADD_MONTHS(%s, %d)), -%d)", specifier, expr, offset, offset), nil
		}
		return fmt.Sprintf("DATE_TRUNC('%s', %s)", specifier, expr), nil
	}

	// With timezone: convert UTC to local, truncate, convert back to UTC
	localExpr := fmt.Sprintf("FROM_UTC_TIMESTAMP(%s, '%s')", expr, tz)
	wrapTZ := func(inner string) string {
		return fmt.Sprintf("TO_UTC_TIMESTAMP(%s, '%s')", inner, tz)
	}

	if grain == runtimev1.TimeGrain_TIME_GRAIN_WEEK && firstDayOfWeek > 1 {
		offset := 8 - firstDayOfWeek
		return wrapTZ(fmt.Sprintf("DATEADD(DAY, -%d, DATE_TRUNC('%s', DATEADD(DAY, %d, %s)))", offset, specifier, offset, localExpr)), nil
	}
	if grain == runtimev1.TimeGrain_TIME_GRAIN_YEAR && firstMonthOfYear > 1 {
		offset := 13 - firstMonthOfYear
		return wrapTZ(fmt.Sprintf("ADD_MONTHS(DATE_TRUNC('%s', ADD_MONTHS(%s, %d)), -%d)", specifier, localExpr, offset, offset)), nil
	}
	return wrapTZ(fmt.Sprintf("DATE_TRUNC('%s', %s)", specifier, localExpr)), nil
}

func (d *dialect) DateDiff(grain runtimev1.TimeGrain, t1, t2 time.Time) (string, error) {
	unit := d.ConvertToDateTruncSpecifier(grain)
	return fmt.Sprintf("DATEDIFF(%s, CAST('%s' AS TIMESTAMP), CAST('%s' AS TIMESTAMP))", unit, t1.Format(time.RFC3339), t2.Format(time.RFC3339)), nil
}

func (d *dialect) IntervalSubtract(tsExpr, unitExpr string, grain runtimev1.TimeGrain) (string, error) {
	return fmt.Sprintf("DATEADD(%s, -(%s), %s)", d.ConvertToDateTruncSpecifier(grain), unitExpr, tsExpr), nil
}

func (d *dialect) SelectTimeRangeBins(start, end time.Time, grain runtimev1.TimeGrain, alias string, tz *time.Location, firstDay, firstMonth int) (string, []any, error) {
	g := timeutil.TimeGrainFromAPI(grain)
	start = timeutil.TruncateTime(start, g, tz, firstDay, firstMonth)
	startStr := start.Format(time.RFC3339)
	endStr := end.Format(time.RFC3339)
	tzStr := tz.String()
	// Databricks SEQUENCE does not support INTERVAL 1 QUARTER or INTERVAL 1 WEEK;
	// convert to the equivalent MONTH/DAY intervals.
	var stepExpr string
	switch grain {
	case runtimev1.TimeGrain_TIME_GRAIN_QUARTER:
		stepExpr = "INTERVAL 3 MONTH"
	case runtimev1.TimeGrain_TIME_GRAIN_WEEK:
		stepExpr = "INTERVAL 7 DAY"
	default:
		stepExpr = fmt.Sprintf("INTERVAL 1 %s", d.ConvertToDateTruncSpecifier(grain))
	}

	switch grain {
	case runtimev1.TimeGrain_TIME_GRAIN_MILLISECOND,
		runtimev1.TimeGrain_TIME_GRAIN_SECOND,
		runtimev1.TimeGrain_TIME_GRAIN_MINUTE,
		runtimev1.TimeGrain_TIME_GRAIN_HOUR:
		// Sub-day grains: generate timestamps directly
		return fmt.Sprintf(
			"SELECT ts AS %s FROM (SELECT EXPLODE(SEQUENCE(CAST('%s' AS TIMESTAMP), CAST('%s' AS TIMESTAMP), %s)) AS ts) WHERE ts < CAST('%s' AS TIMESTAMP)",
			d.EscapeAlias(alias), startStr, endStr, stepExpr, endStr,
		), nil, nil
	default:
		// Day+ grains: generate dates in local timezone, convert back to UTC
		return fmt.Sprintf(
			"SELECT TO_UTC_TIMESTAMP(CAST(d AS TIMESTAMP), '%s') AS %s FROM (SELECT EXPLODE(SEQUENCE(CAST(FROM_UTC_TIMESTAMP(CAST('%s' AS TIMESTAMP), '%s') AS DATE), CAST(FROM_UTC_TIMESTAMP(CAST('%s' AS TIMESTAMP), '%s') AS DATE), %s)) AS d) WHERE TO_UTC_TIMESTAMP(CAST(d AS TIMESTAMP), '%s') < CAST('%s' AS TIMESTAMP)",
			tzStr, d.EscapeAlias(alias), startStr, tzStr, endStr, tzStr, stepExpr, tzStr, endStr,
		), nil, nil
	}
}
