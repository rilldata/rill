package starrocks

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

var DialectStarrocks drivers.Dialect = func() drivers.Dialect {
	d := &dialect{}
	d.InitBase(d)
	return d
}()

func (d *dialect) String() string { return "starrocks" }

func (d *dialect) EscapeIdentifier(ident string) string {
	if ident == "" {
		return ident
	}
	// StarRocks uses backticks for quoting identifiers
	// Replace any backticks inside the identifier with double backticks.
	return fmt.Sprintf("`%s`", strings.ReplaceAll(ident, "`", "``"))
}

func (d *dialect) SupportsILike() bool {
	return false
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

func (d *dialect) JoinOnExpression(lhs, rhs string) string {
	// StarRocks uses MySQL's NULL-safe equal operator.
	return fmt.Sprintf("%s <=> %s", lhs, rhs)
}

func (d *dialect) GetDateTimeExpr(t time.Time) (bool, string) {
	return true, fmt.Sprintf("CAST('%s' AS DATETIME)", t.Format(time.DateTime))
}

func (d *dialect) GetDateExpr(t time.Time) (bool, string) {
	return true, fmt.Sprintf("CAST('%s' AS DATE)", t.Format(time.DateOnly))
}

func (d *dialect) DateTruncExpr(dim *runtimev1.MetricsViewSpec_Dimension, grain runtimev1.TimeGrain, tz string, _, _ int) (string, error) {
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

	if tz == "" {
		return fmt.Sprintf("date_trunc('%s', %s)", specifier, expr), nil
	}
	// Convert to target timezone, truncate, then convert back to UTC.
	return fmt.Sprintf("CONVERT_TZ(date_trunc('%s', CONVERT_TZ(%s, 'UTC', '%s')), '%s', 'UTC')", specifier, expr, tz, tz), nil
}

func (d *dialect) DateDiff(grain runtimev1.TimeGrain, t1, t2 time.Time) (string, error) {
	unit := d.ConvertToDateTruncSpecifier(grain)
	return fmt.Sprintf("DATEDIFF('%s', TIMESTAMP '%s', TIMESTAMP '%s')", unit, t1.Format(time.RFC3339), t2.Format(time.RFC3339)), nil
}

func (d *dialect) IntervalSubtract(tsExpr, unitExpr string, grain runtimev1.TimeGrain) (string, error) {
	return fmt.Sprintf("(%s - INTERVAL (%s) %s)", tsExpr, unitExpr, d.ConvertToDateTruncSpecifier(grain)), nil
}

func (d *dialect) SelectTimeRangeBins(start, end time.Time, grain runtimev1.TimeGrain, alias string, tz *time.Location, firstDay, firstMonth int) (string, []any, error) {
	g := timeutil.TimeGrainFromAPI(grain)
	start = timeutil.TruncateTime(start, g, tz, firstDay, firstMonth)
	// StarRocks uses UNION ALL for generating time series.
	var sb strings.Builder
	first := true
	for t := start; t != end; t = timeutil.OffsetTime(t, g, 1, tz) {
		if !first {
			sb.WriteString(" UNION ALL ")
		}
		sb.WriteString(fmt.Sprintf("SELECT CAST('%s' AS DATETIME) AS %s", t.Format(time.DateTime), d.EscapeAlias(alias)))
		first = false
	}
	return sb.String(), nil, nil
}

func (d *dialect) ColumnCardinalitySQL(db, dbSchema, table, column string) (string, error) {
	return fmt.Sprintf("SELECT approx_count_distinct(%s) AS count FROM %s", d.EscapeIdentifier(column), d.EscapeTable(db, dbSchema, table)), nil
}

func (d *dialect) ColumnDescriptiveStatistics(db, dbSchema, table, column string) (string, error) {
	return fmt.Sprintf("SELECT "+
		"CAST(min(%[1]s) AS DOUBLE) as min, "+
		"CAST(percentile_approx(%[1]s, 0.25) AS DOUBLE) as q25, "+
		"CAST(percentile_approx(%[1]s, 0.5) AS DOUBLE) as q50, "+
		"CAST(percentile_approx(%[1]s, 0.75) AS DOUBLE) as q75, "+
		"CAST(max(%[1]s) AS DOUBLE) as max, "+
		"CAST(avg(%[1]s) AS DOUBLE) as mean, "+
		"CAST(stddev_samp(%[1]s) AS DOUBLE) as sd "+
		"FROM %[2]s WHERE %[1]s IS NOT NULL",
		d.EscapeIdentifier(column),
		d.EscapeTable(db, dbSchema, table)), nil
}

func (d *dialect) IsNonNullFinite(floatColumn string) string {
	sanitizedFloatColumn := d.EscapeIdentifier(floatColumn)
	// StarRocks doesn't have isinf(), use range check to filter Infinity
	// -1e308 to 1e308 covers all finite DOUBLE values
	return fmt.Sprintf("%s IS NOT NULL AND %s > -1e308 AND %s < 1e308", sanitizedFloatColumn, sanitizedFloatColumn, sanitizedFloatColumn)
}

func (d dialect) ColumnNumericHistogram(db, dbSchema, table, column string) (string, error) {
	sanitizedColumnName := d.EscapeIdentifier(column)
	return fmt.Sprintf("SELECT (percentile_approx(%s, 0.75)-percentile_approx(%s, 0.25)) AS iqr, approx_count_distinct(%s) AS count, (max(%s) - min(%s)) AS `range` FROM %s",
		sanitizedColumnName,
		sanitizedColumnName,
		sanitizedColumnName,
		sanitizedColumnName,
		sanitizedColumnName,
		d.EscapeTable(db, dbSchema, table)), nil
}
