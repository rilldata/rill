package snowflake

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/sqldialect"
	"github.com/rilldata/rill/runtime/pkg/timeutil"
)

// snowflakeSpecialCharsRegex matches any character that requires quoting in Snowflake identifiers.
// NOTE: it does not handle cases when identifier is a reserved keyword.
var snowflakeSpecialCharsRegex = regexp.MustCompile(`[^A-Za-z0-9_]|^\d`)

type dialect struct {
	sqldialect.Base
}

func newDialect() *dialect {
	d := &dialect{}
	d.InitBase(d)
	return d
}

func (d *dialect) String() string { return "snowflake" }

func (d *dialect) EscapeIdentifier(ident string) string {
	if ident == "" {
		return ident
	}
	// Snowflake stores unquoted identifiers as uppercase; only quote if necessary.
	if snowflakeSpecialCharsRegex.MatchString(ident) {
		return fmt.Sprintf(`"%s"`, strings.ReplaceAll(ident, `"`, `""`)) // nolint:gocritic
	}
	return ident
}

// EscapeAlias always quotes in Snowflake to prevent silent uppercasing.
func (d *dialect) EscapeAlias(alias string) string {
	return fmt.Sprintf(`"%s"`, strings.ReplaceAll(alias, `"`, `""`)) // nolint:gocritic
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

func (d *dialect) GetDateTimeExpr(_ time.Time) (bool, string) {
	return false, ""
}

func (d *dialect) GetDateExpr(_ time.Time) (bool, string) {
	return false, ""
}

func (d *dialect) CastToDataType(typ runtimev1.Type_Code) (string, error) {
	switch typ {
	case runtimev1.Type_CODE_TIMESTAMP:
		return "TIMESTAMP", nil
	default:
		return "", fmt.Errorf("unsupported cast type %q for dialect %q", typ.String(), d.String())
	}
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
			return fmt.Sprintf("date_trunc('%s', %s::TIMESTAMP)", specifier, expr), nil
		}
		return fmt.Sprintf("date_trunc('%s', %s::TIMESTAMP + INTERVAL '%s') - INTERVAL '%s'", specifier, expr, shift, shift), nil
	}

	// CONVERT_TIMEZONE('source_tz', 'target_tz', ts) converts from source to target.
	if shift == "" {
		return fmt.Sprintf("CONVERT_TIMEZONE('%s', 'UTC', date_trunc('%s', CONVERT_TIMEZONE('UTC', '%s', %s::TIMESTAMP)))", tz, specifier, tz, expr), nil
	}
	return fmt.Sprintf("CONVERT_TIMEZONE('%s', 'UTC', date_trunc('%s', CONVERT_TIMEZONE('UTC', '%s', %s::TIMESTAMP) + INTERVAL '%s') - INTERVAL '%s')", tz, specifier, tz, expr, shift, shift), nil
}

func (d *dialect) DateDiff(grain runtimev1.TimeGrain, t1, t2 time.Time) (string, error) {
	unit := d.ConvertToDateTruncSpecifier(grain)
	return fmt.Sprintf("DATEDIFF('%s', CAST('%s' AS TIMESTAMP), CAST('%s' AS TIMESTAMP))", unit, t1.Format(time.RFC3339), t2.Format(time.RFC3339)), nil
}

func (d *dialect) IntervalSubtract(tsExpr, unitExpr string, grain runtimev1.TimeGrain) (string, error) {
	return fmt.Sprintf("DATEADD('%s', -1 * (%s), %s::TIMESTAMP)", d.ConvertToDateTruncSpecifier(grain), unitExpr, tsExpr), nil
}

func (d *dialect) SelectTimeRangeBins(start, end time.Time, grain runtimev1.TimeGrain, alias string, tz *time.Location, firstDay, firstMonth int) (string, []any, error) {
	g := timeutil.TimeGrainFromAPI(grain)
	start = timeutil.TruncateTime(start, g, tz, firstDay, firstMonth)
	// Snowflake uses UNION ALL for generating time series.
	var sb strings.Builder
	first := true
	for t := start; t.Before(end); t = timeutil.OffsetTime(t, g, 1, tz) {
		if !first {
			sb.WriteString(" UNION ALL ")
		}
		fmt.Fprintf(&sb, "SELECT CAST('%s' AS TIMESTAMP) AS %s", t.Format(time.RFC3339), d.EscapeAlias(alias))
		first = false
	}
	return sb.String(), nil, nil
}

func (d *dialect) SelectInlineResults(result *drivers.Result) (string, []any, []any, error) {
	for _, f := range result.Schema.Fields {
		if !sqldialect.CheckTypeCompatibility(f) {
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
	prefix := ""

	for result.Next() {
		if err := result.Scan(valuePtrs...); err != nil {
			return "", nil, nil, fmt.Errorf("select inline: failed to scan value: %w", err)
		}
		// format: SELECT ? AS a, ? AS b UNION ALL SELECT ...
		if prefix != "" {
			prefix += " UNION ALL "
		}
		prefix += "SELECT "

		dimVals = append(dimVals, values[0])
		for i, v := range values {
			if i > 0 {
				prefix += ", "
			}
			prefix += fmt.Sprintf("%s AS %s", "?", d.EscapeIdentifier(result.Schema.Fields[i].Name))
			args = append(args, v)
		}
	}
	if err := result.Err(); err != nil {
		return "", nil, nil, err
	}
	return prefix, args, dimVals, nil
}
