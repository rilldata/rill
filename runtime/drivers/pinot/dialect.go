package pinot

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

func newDialect() *dialect {
	d := &dialect{}
	d.InitBase(d)
	return d
}

func (d *dialect) String() string { return "pinot" }

func (d *dialect) EscapeIdentifier(ident string) string {
	if ident == "" {
		return ident
	}
	return fmt.Sprintf(`"%s"`, strings.ReplaceAll(ident, `"`, `""`)) // nolint:gocritic
}

func (d *dialect) SupportsILike() bool               { return false }
func (d *dialect) GetTimeDimensionParameter() string { return "CAST(? AS TIMESTAMP)" }

func (d *dialect) LateralUnnest(_, _, _ string) (tbl string, tupleStyle, auto bool, err error) {
	return "", false, true, nil
}

func (d *dialect) UnnestSQLSuffix(_ string) string {
	panic("Pinot auto unnests")
}

func (d *dialect) GetDateTimeExpr(t time.Time) (bool, string) {
	return true, fmt.Sprintf("CAST(%d AS TIMESTAMP)", t.UnixMilli())
}

func (d *dialect) GetDateExpr(t time.Time) (bool, string) {
	return true, fmt.Sprintf("CAST(%d AS DATE)", t.UnixMilli())
}

func (d *dialect) CastToDataType(typ runtimev1.Type_Code) (string, error) {
	switch typ {
	case runtimev1.Type_CODE_TIMESTAMP:
		return "TIMESTAMP", nil
	default:
		return "", fmt.Errorf("unsupported cast type %q for dialect %q", typ.String(), d.String())
	}
}

func (d *dialect) DateTruncExpr(dim *runtimev1.MetricsViewSpec_Dimension, grain runtimev1.TimeGrain, tz string, _, _ int) (string, error) {
	// TODO: Handle tz instead of ignoring it.
	// TODO: Handle firstDayOfWeek and firstMonthOfYear (currently errored in runtime/validate.go).
	if tz == "UTC" || tz == "Etc/UTC" {
		tz = ""
	}

	specifier := d.ConvertToDateTruncSpecifier(grain)

	var expr string
	if dim.Expression != "" {
		expr = fmt.Sprintf("(%s)", dim.Expression)
	} else {
		expr = d.EscapeIdentifier(dim.Column)
	}

	// Adding a cast to TIMESTAMP to ensure the output type is TIMESTAMP (not long).
	if tz == "" {
		return fmt.Sprintf("CAST(date_trunc('%s', %s, 'MILLISECONDS') AS TIMESTAMP)", specifier, expr), nil
	}
	return fmt.Sprintf("CAST(date_trunc('%s', %s, 'MILLISECONDS', '%s') AS TIMESTAMP)", specifier, expr, tz), nil
}

func (d *dialect) DateDiff(grain runtimev1.TimeGrain, t1, t2 time.Time) (string, error) {
	unit := d.ConvertToDateTruncSpecifier(grain)
	return fmt.Sprintf("DATEDIFF('%s', %d, %d)", unit, t1.UnixMilli(), t2.UnixMilli()), nil
}

func (d *dialect) IntervalSubtract(tsExpr, unitExpr string, grain runtimev1.TimeGrain) (string, error) {
	return fmt.Sprintf("CAST((dateAdd('%s', -1 * %s, %s)) AS TIMESTAMP)", d.ConvertToDateTruncSpecifier(grain), unitExpr, tsExpr), nil
}

func (d *dialect) SelectTimeRangeBins(start, end time.Time, grain runtimev1.TimeGrain, alias string, tz *time.Location, firstDay, firstMonth int) (string, []any, error) {
	g := timeutil.TimeGrainFromAPI(grain)
	start = timeutil.TruncateTime(start, g, tz, firstDay, firstMonth)
	var sb strings.Builder
	var args []any
	sb.WriteString("SELECT * FROM (VALUES ")
	for t := start; t.Before(end); t = timeutil.OffsetTime(t, g, 1, tz) {
		if t != start {
			sb.WriteString(", ")
		}
		sb.WriteString("(CAST(? AS TIMESTAMP))")
		args = append(args, t)
	}
	sb.WriteString(fmt.Sprintf(") t (%s)", d.EscapeAlias(alias)))
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
	rows := 0
	prefix := ""
	suffix := ""

	for result.Next() {
		if err := result.Scan(valuePtrs...); err != nil {
			return "", nil, nil, fmt.Errorf("select inline: failed to scan value: %w", err)
		}
		// format: SELECT * FROM (VALUES (val, ...), ...) t(a, b, ...)
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
			ok, expr, err := d.GetValExpr(v, result.Schema.Fields[i].Type.Code)
			if err != nil {
				return "", nil, nil, fmt.Errorf("select inline: failed to get value expression: %w", err)
			}
			if !ok {
				return "", nil, nil, fmt.Errorf("select inline: unsupported value type %q: %w", result.Schema.Fields[i].Type.Code, drivers.ErrOptimizationFailure)
			}
			prefix += expr
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
	return prefix + suffix, nil, dimVals, nil
}
