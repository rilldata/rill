package duckdb

import (
	"fmt"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/timeutil"
)

type dialect struct {
	drivers.BaseDialect
}

var DialectDuckDB drivers.Dialect = func() drivers.Dialect {
	d := &dialect{}
	d.InitBase(d)
	return d
}()

func (d *dialect) String() string { return drivers.DialectNameDuckDB }

func (d *dialect) CanPivot() bool { return true }

// EscapeTable for DuckDB only uses the table name (no db/schema prefix).
func (d *dialect) EscapeTable(_, _, table string) string {
	return d.EscapeIdentifier(table)
}

func (d *dialect) RequiresArrayContainsForInOperator() bool { return true }

func (d *dialect) GetArrayContainsFunction() (string, error) { return "list_has_any", nil }

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

	var shift string
	if grain == runtimev1.TimeGrain_TIME_GRAIN_WEEK && firstDayOfWeek > 1 {
		offset := 8 - firstDayOfWeek
		shift = fmt.Sprintf("%d DAY", offset)
	} else if grain == runtimev1.TimeGrain_TIME_GRAIN_YEAR && firstMonthOfYear > 1 {
		offset := 13 - firstMonthOfYear
		shift = fmt.Sprintf("%d MONTH", offset)
	}

	if tz == "" {
		if shift == "" {
			return fmt.Sprintf("date_trunc('%s', %s::TIMESTAMP)::TIMESTAMP", specifier, expr), nil
		}
		return fmt.Sprintf("date_trunc('%s', %s::TIMESTAMP + INTERVAL %s)::TIMESTAMP - INTERVAL %s", specifier, expr, shift, shift), nil
	}

	// Optimization: date_trunc is faster for day+ granularity.
	switch grain {
	case runtimev1.TimeGrain_TIME_GRAIN_DAY, runtimev1.TimeGrain_TIME_GRAIN_WEEK, runtimev1.TimeGrain_TIME_GRAIN_MONTH, runtimev1.TimeGrain_TIME_GRAIN_QUARTER, runtimev1.TimeGrain_TIME_GRAIN_YEAR:
		if shift == "" {
			return fmt.Sprintf("timezone('%s', date_trunc('%s', timezone('%s', %s::TIMESTAMPTZ)))::TIMESTAMP", tz, specifier, tz, expr), nil
		}
		return fmt.Sprintf("timezone('%s', date_trunc('%s', timezone('%s', %s::TIMESTAMPTZ) + INTERVAL %s) - INTERVAL %s)::TIMESTAMP", tz, specifier, tz, expr, shift, shift), nil
	}

	if shift == "" {
		return fmt.Sprintf("time_bucket(INTERVAL '1 %s', %s::TIMESTAMPTZ, '%s')", specifier, expr, tz), nil
	}
	return fmt.Sprintf("time_bucket(INTERVAL '1 %s', %s::TIMESTAMPTZ + INTERVAL %s, '%s') - INTERVAL %s", specifier, expr, shift, tz, shift), nil
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
	// first convert start and end to the target timezone as the application sends UTC representation of the time, so it will send `2024-03-12T18:30:00Z` for the 13th day of March in Asia/Kolkata timezone (`2024-03-13T00:00:00Z`)
	// then let duckdb range over it and then convert back to the target timezone
	return fmt.Sprintf("SELECT range AT TIME ZONE '%s' AS %s FROM range('%s'::TIMESTAMPTZ AT TIME ZONE '%s', '%s'::TIMESTAMPTZ AT TIME ZONE '%s', INTERVAL '1 %s')",
		tz.String(), d.EscapeAlias(alias),
		start.Format(time.RFC3339), tz.String(),
		end.Format(time.RFC3339), tz.String(),
		d.ConvertToDateTruncSpecifier(grain),
	), nil, nil
}

func (d *dialect) SelectInlineResults(result *drivers.Result) (string, []any, []any, error) {
	for _, f := range result.Schema.Fields {
		if !drivers.CheckTypeCompatibility(f) {
			return "", nil, nil, fmt.Errorf("select inline: schema field type not supported %q: %w", f.Type.Code, drivers.ErrOptimizationFailure)
		}
	}

	values := make([]any, len(result.Schema.Fields))
	valuePtrs := make([]any, len(result.Schema.Fields))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	var dimVals []any
	var args []any
	rows := 0
	prefix := ""
	suffix := ""

	for result.Next() {
		if err := result.Scan(valuePtrs...); err != nil {
			return "", nil, nil, fmt.Errorf("select inline: failed to scan value: %w", err)
		}
		// format: SELECT * FROM (VALUES (?,?,...), ...) t(a, b, ...)
		if rows == 0 {
			prefix = "SELECT * FROM (VALUES "
			suffix = "t("
		}
		if rows > 0 {
			prefix += ", "
		}

		dimVals = append(dimVals, values[0])
		for i, v := range values {
			if i == 0 {
				prefix += "("
			} else {
				prefix += ", "
			}
			if rows == 0 {
				suffix += d.EscapeIdentifier(result.Schema.Fields[i].Name)
				if i != len(result.Schema.Fields)-1 {
					suffix += ", "
				}
			}
			argExpr, argVal, err := d.GetArgExpr(v, result.Schema.Fields[i].Type.Code)
			if err != nil {
				return "", nil, nil, fmt.Errorf("select inline: failed to get argument expression: %w", err)
			}
			prefix += argExpr
			args = append(args, argVal)
		}
		prefix += ")"
		if rows == 0 {
			suffix += ")"
		}
		rows++
	}
	if err := result.Err(); err != nil {
		return "", nil, nil, err
	}
	prefix += ") "
	return prefix + suffix, args, dimVals, nil
}

func (d *dialect) GetDateTimeExpr(t time.Time) (bool, string) {
	return true, fmt.Sprintf("CAST('%s' AS TIMESTAMP)", t.Format(time.RFC3339Nano))
}

func (d *dialect) GetDateExpr(t time.Time) (bool, string) {
	return true, fmt.Sprintf("CAST('%s' AS DATE)", t.Format(time.DateOnly))
}

func (d *dialect) ColumnCardinality(db, dbSchema, table, column string) (string, error) {
	return fmt.Sprintf("SELECT approx_count_distinct(%s) AS count FROM %s", d.EscapeIdentifier(column), d.EscapeTable(db, dbSchema, table)), nil
}

func (d *dialect) ColumnDescriptiveStatistics(db, dbSchema, table, column string) (string, error) {
	return fmt.Sprintf("SELECT "+
		"min(%[1]s)::DOUBLE as min, "+
		"approx_quantile(%[1]s, 0.25)::DOUBLE as q25, "+
		"approx_quantile(%[1]s, 0.5)::DOUBLE as q50, "+
		"approx_quantile(%[1]s, 0.75)::DOUBLE as q75, "+
		"max(%[1]s)::DOUBLE as max, "+
		"avg(%[1]s)::DOUBLE as mean, "+
		"'NaN'::DOUBLE as sd "+
		"FROM %[2]s WHERE NOT isinf(%[1]s) ",
		d.EscapeIdentifier(column),
		d.EscapeTable(db, dbSchema, table)), nil
}

func (d *dialect) IsNonNullFinite(floatColumn string) string {
	sanitizedFloatColumn := d.EscapeIdentifier(floatColumn)
	return fmt.Sprintf("%s IS NOT NULL AND NOT isinf(%s)", sanitizedFloatColumn, sanitizedFloatColumn)
}

func (d dialect) ColumnNumericHistogramBucket(db, dbSchema, table, column string) (string, error) {
	sanitizedColumnName := d.EscapeIdentifier(column)
	return fmt.Sprintf("SELECT (approx_quantile(%s, 0.75)-approx_quantile(%s, 0.25))::DOUBLE AS iqr, approx_count_distinct(%s) AS count, (max(%s) - min(%s))::DOUBLE AS range FROM %s",
		sanitizedColumnName,
		sanitizedColumnName,
		sanitizedColumnName,
		sanitizedColumnName,
		sanitizedColumnName,
		d.EscapeTable(db, dbSchema, table)), nil
}
