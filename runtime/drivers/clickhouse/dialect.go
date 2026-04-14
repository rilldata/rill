package clickhouse

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/timeutil"
)

var dictPwdRegex = regexp.MustCompile(`PASSWORD\s+'[^']*'`)

type dialect struct {
	drivers.BaseDialect
}

var DialectClickhouse drivers.Dialect = func() drivers.Dialect {
	d := &dialect{}
	d.BaseDialect = drivers.NewBaseDialect(drivers.DialectNameClickHouse, drivers.DoubleQuotesEscapeIdentifier, drivers.DoubleQuotesEscapeIdentifier)
	return d
}()

func (d *dialect) GetCastExprForLike() string { return "::Nullable(TEXT)" }

func (d *dialect) ConvertToDateTruncSpecifier(grain runtimev1.TimeGrain) string {
	return strings.ToLower(d.BaseDialect.ConvertToDateTruncSpecifier(grain))
}

func (d *dialect) DimensionSelect(escapeTable string, dim *runtimev1.MetricsViewSpec_Dimension) (dimSelect, unnestClause string, err error) {
	alias := d.EscapeAlias(dim.Name)
	if !dim.Unnest {
		expr, err := d.MetricsViewDimensionExpression(dim)
		if err != nil {
			return "", "", fmt.Errorf("failed to get dimension expression: %w", err)
		}
		return fmt.Sprintf(`(%s) AS %s`, expr, alias), "", nil
	}
	expr, err := d.MetricsViewDimensionExpression(dim)
	if err != nil {
		return "", "", fmt.Errorf("failed to get dimension expression: %w", err)
	}
	return fmt.Sprintf(`arrayJoin(%s) AS %s`, expr, alias), "", nil
}

func (d *dialect) MetricsViewDimensionExpression(dimension *runtimev1.MetricsViewSpec_Dimension) (string, error) {
	if dimension.LookupTable != "" {
		var keyExpr string
		if dimension.Column != "" {
			keyExpr = d.EscapeIdentifier(dimension.Column)
		} else if dimension.Expression != "" {
			keyExpr = dimension.Expression
		} else {
			return "", fmt.Errorf("dimension %q has a lookup table but no column or expression defined", dimension.Name)
		}
		return lookupExpr(dimension.LookupTable, dimension.LookupValueColumn, keyExpr, dimension.LookupDefaultExpression)
	}
	if dimension.Expression != "" {
		return dimension.Expression, nil
	}
	if dimension.Column != "" {
		return d.EscapeIdentifier(dimension.Column), nil
	}
	// Backwards compatibility for older projects that have not run reconcile on this metrics view.
	// In that case `column` will not be present.
	return d.EscapeIdentifier(dimension.Name), nil
}

func (d *dialect) LateralUnnest(expr, _, colName string) (tbl string, tupleStyle, auto bool, err error) {
	// using LEFT ARRAY JOIN instead of ARRAY JOIN to include empty arrays with zero values
	return fmt.Sprintf("LEFT ARRAY JOIN %s AS %s", expr, d.EscapeIdentifier(colName)), false, false, nil
}

func (d *dialect) UnnestSQLSuffix(tbl string) string {
	return fmt.Sprintf(" %s", tbl)
}

func (d *dialect) RequiresArrayContainsForInOperator() bool { return true }

func (d *dialect) GetArrayContainsFunction() (string, error) { return "hasAny", nil }

func (d *dialect) CastToDataType(typ runtimev1.Type_Code) (string, error) {
	switch typ {
	case runtimev1.Type_CODE_TIMESTAMP:
		return "DateTime64", nil
	default:
		return "", fmt.Errorf("unsupported cast type %q for dialect %q", typ.String(), d.String())
	}
}

func (d *dialect) JoinOnExpression(lhs, rhs string) string {
	return fmt.Sprintf("isNotDistinctFrom(%s, %s)", lhs, rhs)
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
			return fmt.Sprintf("date_trunc('%s', %s)::DateTime64", specifier, expr), nil
		}
		return fmt.Sprintf("date_trunc('%s', %s + INTERVAL %s)::DateTime64 - INTERVAL %s", specifier, expr, shift, shift), nil
	}

	if shift == "" {
		return fmt.Sprintf("date_trunc('%s', %s::DateTime64(6, '%s'))::DateTime64(6, '%s')", specifier, expr, tz, tz), nil
	}
	return fmt.Sprintf("date_trunc('%s', %s::DateTime64(6, '%s') + INTERVAL %s)::DateTime64(6, '%s') - INTERVAL %s", specifier, expr, tz, shift, tz, shift), nil
}

func (d *dialect) DateDiff(grain runtimev1.TimeGrain, t1, t2 time.Time) (string, error) {
	unit := d.ConvertToDateTruncSpecifier(grain)
	return fmt.Sprintf("DATEDIFF('%s', parseDateTimeBestEffort('%s'), parseDateTimeBestEffort('%s'))", unit, t1.Format(time.RFC3339), t2.Format(time.RFC3339)), nil
}

func (d *dialect) IntervalSubtract(tsExpr, unitExpr string, grain runtimev1.TimeGrain) (string, error) {
	return fmt.Sprintf("(%s - INTERVAL (%s) %s)", tsExpr, unitExpr, d.ConvertToDateTruncSpecifier(grain)), nil
}

func (d *dialect) SelectTimeRangeBins(start, end time.Time, grain runtimev1.TimeGrain, alias string, tz *time.Location, firstDay, firstMonth int) (string, []any, error) {
	g := timeutil.TimeGrainFromAPI(grain)
	start = timeutil.TruncateTime(start, g, tz, firstDay, firstMonth)
	// format: SELECT c1 AS "alias" FROM VALUES(toDateTime(...), ...)
	var sb strings.Builder
	var args []any
	sb.WriteString(fmt.Sprintf("SELECT c1 AS %s FROM VALUES(", d.EscapeAlias(alias)))
	for t := start; t.Before(end); t = timeutil.OffsetTime(t, g, 1, tz) {
		if t != start {
			sb.WriteString(", ")
		}
		sb.WriteString("?")
		args = append(args, t)
	}
	sb.WriteString(")")
	return sb.String(), args, nil
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
		// format: SELECT c1 AS a, c2 AS b FROM VALUES((v1, v2), (v1, v2), ...)
		if rows == 0 {
			prefix = "SELECT "
			suffix = " FROM VALUES ("
		}
		if rows > 0 {
			suffix += ", "
		}

		dimVals = append(dimVals, values[0])
		for i, v := range values {
			if i == 0 {
				suffix += "("
			} else {
				suffix += ", "
			}
			if rows == 0 {
				prefix += fmt.Sprintf("c%d AS %s", i+1, d.EscapeIdentifier(result.Schema.Fields[i].Name))
				if i != len(result.Schema.Fields)-1 {
					prefix += ", "
				}
			}
			argExpr, argVal, err := getArgExpr(v, result.Schema.Fields[i].Type.Code)
			if err != nil {
				return "", nil, nil, fmt.Errorf("select inline: failed to get argument expression: %w", err)
			}
			suffix += argExpr
			args = append(args, argVal)
		}
		suffix += ")"
		rows++
	}
	if err := result.Err(); err != nil {
		return "", nil, nil, err
	}
	suffix += ")"
	return prefix + suffix, args, dimVals, nil
}

func (d *dialect) LookupSelectExpr(lookupTable, lookupKeyColumn string) (string, error) {
	return fmt.Sprintf("SELECT %s FROM %s", d.EscapeIdentifier(lookupKeyColumn), d.EscapeQualifiedIdentifier(lookupTable)), nil
}

func (d *dialect) SanitizeQueryForLogging(sql string) string {
	// replace inline "PASSWORD 'pwd'" for dict source with "PASSWORD '***'"
	return dictPwdRegex.ReplaceAllString(sql, "PASSWORD '***'")
}

func (d *dialect) ColumnCardinality(db, dbSchema, table, column string) (string, error) {
	return fmt.Sprintf("SELECT uniq(%s) AS count FROM %s", d.EscapeIdentifier(column), d.EscapeTable(db, dbSchema, table)), nil
}

func (d *dialect) ColumnDescriptiveStatistics(db, dbSchema, table, column string) (string, error) {
	return fmt.Sprintf(`SELECT
			min(%[1]s)::DOUBLE as min,
			quantileTDigest(0.25)(%[1]s)::DOUBLE as q25,
			quantileTDigest(0.5)(%[1]s)::DOUBLE as q50,
			quantileTDigest(0.75)(%[1]s)::DOUBLE as q75,
			max(%[1]s)::DOUBLE as max,
			avg(%[1]s)::DOUBLE as mean,
			stddevSamp(%[1]s)::DOUBLE as sd
			FROM %[2]s WHERE `+d.IsNonNullFinite(column)+``,
		d.EscapeIdentifier(column),
		d.EscapeTable(db, dbSchema, table)), nil
}

func (d *dialect) IsNonNullFinite(floatColumn string) string {
	sanitizedFloatColumn := d.EscapeIdentifier(floatColumn)
	return fmt.Sprintf("%s IS NOT NULL AND isFinite(%s)", sanitizedFloatColumn, sanitizedFloatColumn)
}

func (d dialect) ColumnNumericHistogramBucket(db, dbSchema, table, column string) (string, error) {
	sanitizedColumnName := d.EscapeIdentifier(column)
	return fmt.Sprintf("SELECT (quantileTDigest(0.75)(%s)-quantileTDigest(0.25)(%s)) AS iqr, uniq(%s) AS count, (max(%s) - min(%s)) AS range FROM %s",
		sanitizedColumnName,
		sanitizedColumnName,
		sanitizedColumnName,
		sanitizedColumnName,
		sanitizedColumnName,
		d.EscapeTable(db, dbSchema, table)), nil
}

func getArgExpr(val any, typ runtimev1.Type_Code) (string, any, error) {
	if typ == runtimev1.Type_CODE_DATE {
		t, ok := val.(time.Time)
		if !ok {
			return "", nil, fmt.Errorf("could not cast value %v to time.Time for date type", val)
		}
		return "toDate(?)", t.Format(time.DateOnly), nil
	}
	return "?", val, nil
}

func lookupExpr(lookupTable, lookupValueColumn, lookupKeyExpr, lookupDefaultExpression string) (string, error) {
	if lookupDefaultExpression != "" {
		return fmt.Sprintf("dictGetOrDefault('%s', '%s', %s, %s)", lookupTable, lookupValueColumn, lookupKeyExpr, lookupDefaultExpression), nil
	}
	return fmt.Sprintf("dictGet('%s', '%s', %s)", lookupTable, lookupValueColumn, lookupKeyExpr), nil
}
