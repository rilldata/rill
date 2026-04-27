package bigquery

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

var DialectBigQuery drivers.Dialect = func() drivers.Dialect {
	d := &dialect{}
	d.BaseDialect = drivers.NewBaseDialect(drivers.DialectNameBigQuery, BigQueryEscapeIdentifier, BigQueryEscapeIdentifier)
	return d
}()

func BigQueryEscapeIdentifier(ident string) string {
	if ident == "" {
		return ident
	}
	// Bigquery uses backticks for quoting identifiers
	// Replace any backticks inside the identifier with double backticks
	return fmt.Sprintf("`%s`", strings.ReplaceAll(ident, "`", "``"))
}

func (d *dialect) SupportsILike() bool { return false }

func (d *dialect) GetTimeDimensionParameter(typeCode runtimev1.Type_Code) string {
	if typeCode == runtimev1.Type_CODE_DATE {
		return "DATE(?)"
	}
	return "?"
}

func (d *dialect) SafeDivideExpression(numExpr, denExpr string) string {
	return fmt.Sprintf("SAFE_DIVIDE(%s, CAST(%s AS FLOAT64))", numExpr, denExpr)
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
	// BigQuery requires plain equality for FULL joins
	return fmt.Sprintf("coalesce(CAST(%s AS STRING), '__rill_sentinel__') = coalesce(CAST(%s AS STRING), '__rill_sentinel__')", lhs, rhs)
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

	// BigQuery: TIMESTAMP_TRUNC(expr, SPECIFIER [, 'timezone'])
	// For DATE columns, explicitly cast to TIMESTAMP first to ensure the result is TIMESTAMP (not DATETIME).
	if dim.DataType != nil && dim.DataType.Code == runtimev1.Type_CODE_DATE {
		expr = fmt.Sprintf("TIMESTAMP(%s)", expr)
	}
	if tz == "" {
		if grain == runtimev1.TimeGrain_TIME_GRAIN_WEEK && firstDayOfWeek > 1 {
			offset := 8 - firstDayOfWeek
			return fmt.Sprintf("TIMESTAMP_SUB(TIMESTAMP_TRUNC(TIMESTAMP_ADD(%s, INTERVAL %d DAY), %s), INTERVAL %d DAY)", expr, offset, specifier, offset), nil
		}
		if grain == runtimev1.TimeGrain_TIME_GRAIN_YEAR && firstMonthOfYear > 1 {
			offset := 13 - firstMonthOfYear
			// TIMESTAMP_ADD/TIMESTAMP_SUB don't support MONTH; wrap TIMESTAMP_TRUNC result in DATETIME() before DATETIME_SUB.
			return fmt.Sprintf("TIMESTAMP(DATETIME_SUB(DATETIME(TIMESTAMP_TRUNC(TIMESTAMP(DATETIME_ADD(DATETIME(%s, 'UTC'), INTERVAL %d MONTH), 'UTC'), %s), 'UTC'), INTERVAL %d MONTH), 'UTC')", expr, offset, specifier, offset), nil
		}
		return fmt.Sprintf("TIMESTAMP_TRUNC(%s, %s)", expr, specifier), nil
	}
	// TIMESTAMP_TRUNC natively accepts a timezone argument.
	if grain == runtimev1.TimeGrain_TIME_GRAIN_WEEK && firstDayOfWeek > 1 {
		offset := 8 - firstDayOfWeek
		return fmt.Sprintf("TIMESTAMP_SUB(TIMESTAMP_TRUNC(TIMESTAMP_ADD(%s, INTERVAL %d DAY), %s, '%s'), INTERVAL %d DAY)", expr, offset, specifier, tz, offset), nil
	}
	if grain == runtimev1.TimeGrain_TIME_GRAIN_YEAR && firstMonthOfYear > 1 {
		offset := 13 - firstMonthOfYear
		// TIMESTAMP_ADD/TIMESTAMP_SUB don't support MONTH; wrap TIMESTAMP_TRUNC result in DATETIME() before DATETIME_SUB.
		return fmt.Sprintf("TIMESTAMP(DATETIME_SUB(DATETIME(TIMESTAMP_TRUNC(TIMESTAMP(DATETIME_ADD(DATETIME(%s, 'UTC'), INTERVAL %d MONTH), 'UTC'), %s, '%s'), 'UTC'), INTERVAL %d MONTH), 'UTC')", expr, offset, specifier, tz, offset), nil
	}
	return fmt.Sprintf("TIMESTAMP_TRUNC(%s, %s, '%s')", expr, specifier, tz), nil
}

func (d *dialect) DateDiff(grain runtimev1.TimeGrain, t1, t2 time.Time) (string, error) {
	unit := d.ConvertToDateTruncSpecifier(grain)
	return fmt.Sprintf("DATETIME_DIFF(DATETIME(CAST('%s' AS TIMESTAMP), 'UTC'), DATETIME(CAST('%s' AS TIMESTAMP), 'UTC'), %s)", t2.Format(time.RFC3339), t1.Format(time.RFC3339), unit), nil
}

func (d *dialect) IntervalSubtract(tsExpr, unitExpr string, grain runtimev1.TimeGrain) (string, error) {
	return fmt.Sprintf("TIMESTAMP(DATETIME_SUB(DATETIME(%s, 'UTC'), INTERVAL (%s) %s), 'UTC')", tsExpr, unitExpr, d.ConvertToDateTruncSpecifier(grain)), nil
}

func (d *dialect) SelectTimeRangeBins(start, end time.Time, grain runtimev1.TimeGrain, alias string, tz *time.Location, firstDay, firstMonth int) (string, []any, error) {
	g := timeutil.TimeGrainFromAPI(grain)
	start = timeutil.TruncateTime(start, g, tz, firstDay, firstMonth)
	startStr := start.Format(time.RFC3339)
	endStr := end.Format(time.RFC3339)
	tzStr := tz.String()
	switch grain {
	case runtimev1.TimeGrain_TIME_GRAIN_MILLISECOND,
		runtimev1.TimeGrain_TIME_GRAIN_SECOND,
		runtimev1.TimeGrain_TIME_GRAIN_MINUTE,
		runtimev1.TimeGrain_TIME_GRAIN_HOUR:
		return fmt.Sprintf(
			"SELECT ts AS %s FROM UNNEST(GENERATE_TIMESTAMP_ARRAY(TIMESTAMP '%s', TIMESTAMP '%s', INTERVAL 1 %s)) AS ts WHERE ts < TIMESTAMP '%s'",
			d.EscapeAlias(alias), startStr, endStr, d.ConvertToDateTruncSpecifier(grain), endStr,
		), nil, nil
	default:
		return fmt.Sprintf(
			"SELECT TIMESTAMP(d, '%s') AS %s FROM UNNEST(GENERATE_DATE_ARRAY(DATE(TIMESTAMP '%s', '%s'), DATE(TIMESTAMP '%s', '%s'), INTERVAL 1 %s)) AS d WHERE TIMESTAMP(d, '%s') < TIMESTAMP '%s'",
			tzStr, d.EscapeAlias(alias), startStr, tzStr, endStr, tzStr, d.ConvertToDateTruncSpecifier(grain), tzStr, endStr,
		), nil, nil
	}
}
