package druid

import (
	"fmt"
	"math"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/timeutil"
)

type dialect struct {
	drivers.BaseDialect
}

var DialectDruid drivers.Dialect = func() drivers.Dialect {
	d := &dialect{}
	d.BaseDialect = drivers.NewBaseDialect(drivers.DialectNameDruid, drivers.DoubleQuotesEscapeIdentifier, drivers.DoubleQuotesEscapeIdentifier)
	return d
}()

func (d *dialect) SupportsILike() bool { return false }

func (d *dialect) SupportsRegexMatch() bool { return true }

func (d *dialect) GetRegexMatchFunction() (string, error) { return "REGEXP_LIKE", nil }

// DimensionSelect for Druid skips unnesting even when dim.Unnest is true.
func (d *dialect) DimensionSelect(_ string, dim *runtimev1.MetricsViewSpec_Dimension) (dimSelect, unnestClause string, err error) {
	alias := d.EscapeAlias(dim.Name)
	expr, err := d.MetricsViewDimensionExpression(dim)
	if err != nil {
		return "", "", fmt.Errorf("failed to get dimension expression: %w", err)
	}
	return fmt.Sprintf(`(%s) AS %s`, expr, alias), "", nil
}

func (d *dialect) LateralUnnest(_, _, _ string) (tbl string, tupleStyle, auto bool, err error) {
	return "", false, true, nil
}

func (d *dialect) UnnestSQLSuffix(_ string) string {
	panic("Druid auto unnests")
}

func (d *dialect) MinDimensionExpression(expr string) string {
	return fmt.Sprintf("EARLIEST(%s)", expr) // MIN on string columns is not supported in Druid
}

func (d *dialect) MaxDimensionExpression(expr string) string {
	return fmt.Sprintf("LATEST(%s)", expr) // MAX on string columns is not supported in Druid
}

func (d *dialect) SafeDivideExpression(numExpr, denExpr string) string {
	return fmt.Sprintf("SAFE_DIVIDE(%s, CAST(%s AS DOUBLE))", numExpr, denExpr)
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

	var specifier string
	if tz != "" {
		specifier = druidTimeFloorSpecifier(grain)
	} else {
		specifier = d.ConvertToDateTruncSpecifier(grain)
	}

	var expr string
	if dim.Expression != "" {
		expr = fmt.Sprintf("(%s)", dim.Expression)
	} else {
		expr = d.EscapeIdentifier(dim.Column)
	}

	var shift int
	var shiftPeriod string
	if grain == runtimev1.TimeGrain_TIME_GRAIN_WEEK && firstDayOfWeek > 1 {
		shift = 8 - firstDayOfWeek
		shiftPeriod = "P1D"
	} else if grain == runtimev1.TimeGrain_TIME_GRAIN_YEAR && firstMonthOfYear > 1 {
		shift = 13 - firstMonthOfYear
		shiftPeriod = "P1M"
	}

	if tz == "" {
		if shift == 0 {
			return fmt.Sprintf("date_trunc('%s', %s)", specifier, expr), nil
		}
		return fmt.Sprintf("time_shift(date_trunc('%s', time_shift(%s, '%s', %d)), '%s', -%d)", specifier, expr, shiftPeriod, shift, shiftPeriod, shift), nil
	}

	if shift == 0 {
		return fmt.Sprintf("time_floor(%s, '%s', null, '%s')", expr, specifier, tz), nil
	}
	return fmt.Sprintf("time_shift(time_floor(time_shift(%s, '%s', %d), '%s', null, '%s'), '%s', -%d)", expr, shiftPeriod, shift, specifier, tz, shiftPeriod, shift), nil
}

func (d *dialect) DateDiff(grain runtimev1.TimeGrain, t1, t2 time.Time) (string, error) {
	unit := d.ConvertToDateTruncSpecifier(grain)
	return fmt.Sprintf("TIMESTAMPDIFF(%q, TIME_PARSE('%s'), TIME_PARSE('%s'))", unit, t1.Format(time.RFC3339), t2.Format(time.RFC3339)), nil
}

func (d *dialect) IntervalSubtract(tsExpr, unitExpr string, grain runtimev1.TimeGrain) (string, error) {
	return fmt.Sprintf("(%s - INTERVAL (%s) %s)", tsExpr, unitExpr, d.ConvertToDateTruncSpecifier(grain)), nil
}

func (d *dialect) SelectTimeRangeBins(start, end time.Time, grain runtimev1.TimeGrain, alias string, tz *time.Location, firstDay, firstMonth int) (string, []any, error) {
	g := timeutil.TimeGrainFromAPI(grain)
	start = timeutil.TruncateTime(start, g, tz, firstDay, firstMonth)
	// generate select like - SELECT * FROM (
	//  VALUES
	//  (CAST('2006-01-02T15:04:05Z' AS TIMESTAMP)),
	//  (CAST('2006-01-02T15:04:05Z' AS TIMESTAMP))
	// ) t (time)
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
		// format: SELECT * FROM (VALUES (val, val, ...), ...) t(a, b, ...)
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
			ok, expr, err := getValExpr(v, result.Schema.Fields[i].Type.Code)
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

func druidTimeFloorSpecifier(grain runtimev1.TimeGrain) string {
	switch grain {
	case runtimev1.TimeGrain_TIME_GRAIN_MILLISECOND:
		return "PT0.001S"
	case runtimev1.TimeGrain_TIME_GRAIN_SECOND:
		return "PT1S"
	case runtimev1.TimeGrain_TIME_GRAIN_MINUTE:
		return "PT1M"
	case runtimev1.TimeGrain_TIME_GRAIN_HOUR:
		return "PT1H"
	case runtimev1.TimeGrain_TIME_GRAIN_DAY:
		return "P1D"
	case runtimev1.TimeGrain_TIME_GRAIN_WEEK:
		return "P1W"
	case runtimev1.TimeGrain_TIME_GRAIN_MONTH:
		return "P1M"
	case runtimev1.TimeGrain_TIME_GRAIN_QUARTER:
		return "P3M"
	case runtimev1.TimeGrain_TIME_GRAIN_YEAR:
		return "P1Y"
	}
	panic(fmt.Errorf("invalid time grain enum value %d", int(grain)))
}

func getValExpr(val any, typ runtimev1.Type_Code) (bool, string, error) {
	if val == nil {
		ok, expr := getNullExpr(typ)
		if ok {
			return true, expr, nil
		}
		return false, "", fmt.Errorf("could not get null expr for type %q", typ)
	}
	switch typ {
	case runtimev1.Type_CODE_STRING:
		if s, ok := val.(string); ok {
			return true, drivers.EscapeStringValue(s), nil
		}
		return false, "", fmt.Errorf("could not cast value %v to string type", val)
	case runtimev1.Type_CODE_INT8, runtimev1.Type_CODE_INT16, runtimev1.Type_CODE_INT32, runtimev1.Type_CODE_INT64,
		runtimev1.Type_CODE_UINT8, runtimev1.Type_CODE_UINT16, runtimev1.Type_CODE_UINT32, runtimev1.Type_CODE_UINT64,
		runtimev1.Type_CODE_FLOAT32, runtimev1.Type_CODE_FLOAT64:
		// check NaN and Inf
		if f, ok := val.(float64); ok && (math.IsNaN(f) || math.IsInf(f, 0)) {
			return true, "NULL", nil
		}
		return true, fmt.Sprintf("%v", val), nil
	case runtimev1.Type_CODE_BOOL:
		return true, fmt.Sprintf("%v", val), nil
	case runtimev1.Type_CODE_TIME, runtimev1.Type_CODE_TIMESTAMP:
		if t, ok := val.(time.Time); ok {
			if ok, expr := getDateTimeExpr(t); ok {
				return true, expr, nil
			}
			return false, "", fmt.Errorf("cannot get time expr for this dialect")
		}
		return false, "", fmt.Errorf("unsupported time type %q", typ)
	case runtimev1.Type_CODE_DATE:
		if t, ok := val.(time.Time); ok {
			if ok, expr := getDateExpr(t); ok {
				return true, expr, nil
			}
			return false, "", fmt.Errorf("cannot get date expr for this dialect")
		}
		return false, "", fmt.Errorf("unsupported date type %q", typ)
	default:
		return false, "", fmt.Errorf("unsupported type %q", typ)
	}
}

func getNullExpr(typ runtimev1.Type_Code) (bool, string) {
	switch typ {
	case runtimev1.Type_CODE_STRING:
		return true, "CAST(NULL AS VARCHAR)"
	case runtimev1.Type_CODE_INT8, runtimev1.Type_CODE_INT16, runtimev1.Type_CODE_INT32, runtimev1.Type_CODE_INT64,
		runtimev1.Type_CODE_INT128, runtimev1.Type_CODE_INT256,
		runtimev1.Type_CODE_UINT8, runtimev1.Type_CODE_UINT16, runtimev1.Type_CODE_UINT32, runtimev1.Type_CODE_UINT64,
		runtimev1.Type_CODE_UINT128, runtimev1.Type_CODE_UINT256:
		return true, "CAST(NULL AS INTEGER)"
	case runtimev1.Type_CODE_FLOAT32, runtimev1.Type_CODE_FLOAT64, runtimev1.Type_CODE_DECIMAL:
		return true, "CAST(NULL AS DOUBLE)"
	case runtimev1.Type_CODE_BOOL:
		return true, "CAST(NULL AS BOOLEAN)"
	case runtimev1.Type_CODE_TIME, runtimev1.Type_CODE_DATE, runtimev1.Type_CODE_TIMESTAMP:
		return true, "CAST(NULL AS TIMESTAMP)"
	default:
		return false, ""
	}
}

func getDateTimeExpr(t time.Time) (bool, string) {
	return true, fmt.Sprintf("CAST('%s' AS TIMESTAMP)", t.Format(time.RFC3339Nano))
}

func getDateExpr(t time.Time) (bool, string) {
	return true, fmt.Sprintf("CAST('%s' AS DATE)", t.Format(time.DateOnly))
}
