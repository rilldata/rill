// Package sqldialect provides a base struct with default implementations for the
// drivers.Dialect interface. Each OLAP driver embeds Base in its concrete dialect
// and calls InitBase to wire up virtual dispatch.
package sqldialect

import (
	"fmt"
	"math"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

// Dispatcher is the subset of drivers.Dialect that Base delegates back to the concrete
// dialect for virtual dispatch. Defined locally to avoid a circular import with drivers.
type Dispatcher interface {
	EscapeIdentifier(string) string
	EscapeAlias(string) string
	EscapeQualifiedIdentifier(string) string
	EscapeTable(string, string, string) string
	MetricsViewDimensionExpression(*runtimev1.MetricsViewSpec_Dimension) (string, error)
	LookupExpr(string, string, string, string) (string, error)
	GetNullExpr(runtimev1.Type_Code) (bool, string)
	GetDateTimeExpr(time.Time) (bool, string)
	GetDateExpr(time.Time) (bool, string)
}

// Base provides default implementations for the drivers.Dialect interface.
// Embed it in a concrete dialect struct and call InitBase to wire up virtual dispatch.
type Base struct {
	self Dispatcher
}

// InitBase wires up virtual dispatch. Must be called in the concrete dialect's constructor.
func (b *Base) InitBase(self Dispatcher) {
	b.self = self
}

func (b *Base) EscapeStringValue(s string) string {
	return fmt.Sprintf("'%s'", strings.ReplaceAll(s, "'", "''"))
}

func (b *Base) EscapeAlias(alias string) string {
	return b.self.EscapeIdentifier(alias)
}

func (b *Base) EscapeQualifiedIdentifier(name string) string {
	if name == "" {
		return name
	}
	parts := strings.Split(name, ".")
	for i, part := range parts {
		parts[i] = b.self.EscapeIdentifier(part)
	}
	return strings.Join(parts, ".")
}

func (b *Base) EscapeTable(db, schema, table string) string {
	var sb strings.Builder
	if db != "" {
		sb.WriteString(b.self.EscapeIdentifier(db))
		sb.WriteString(".")
	}
	if schema != "" {
		sb.WriteString(b.self.EscapeIdentifier(schema))
		sb.WriteString(".")
	}
	sb.WriteString(b.self.EscapeIdentifier(table))
	return sb.String()
}

func (b *Base) EscapeMember(tbl, name string) string {
	if tbl == "" {
		return b.self.EscapeIdentifier(name)
	}
	return fmt.Sprintf("%s.%s", b.self.EscapeIdentifier(tbl), b.self.EscapeIdentifier(name))
}

func (b *Base) EscapeMemberAlias(tbl, alias string) string {
	if tbl == "" {
		return b.self.EscapeAlias(alias)
	}
	return fmt.Sprintf("%s.%s", b.self.EscapeIdentifier(tbl), b.self.EscapeAlias(alias))
}

func (b *Base) ConvertToDateTruncSpecifier(grain runtimev1.TimeGrain) string {
	switch grain {
	case runtimev1.TimeGrain_TIME_GRAIN_MILLISECOND:
		return "MILLISECOND"
	case runtimev1.TimeGrain_TIME_GRAIN_SECOND:
		return "SECOND"
	case runtimev1.TimeGrain_TIME_GRAIN_MINUTE:
		return "MINUTE"
	case runtimev1.TimeGrain_TIME_GRAIN_HOUR:
		return "HOUR"
	case runtimev1.TimeGrain_TIME_GRAIN_DAY:
		return "DAY"
	case runtimev1.TimeGrain_TIME_GRAIN_WEEK:
		return "WEEK"
	case runtimev1.TimeGrain_TIME_GRAIN_MONTH:
		return "MONTH"
	case runtimev1.TimeGrain_TIME_GRAIN_QUARTER:
		return "QUARTER"
	case runtimev1.TimeGrain_TIME_GRAIN_YEAR:
		return "YEAR"
	}
	return ""
}

func (b *Base) CanPivot() bool                           { return false }
func (b *Base) SupportsILike() bool                      { return true }
func (b *Base) GetCastExprForLike() string               { return "" }
func (b *Base) SupportsRegexMatch() bool                 { return false }
func (b *Base) GetRegexMatchFunction() string            { panic("regex match not supported for this dialect") }
func (b *Base) RequiresArrayContainsForInOperator() bool { return false }
func (b *Base) GetArrayContainsFunction() string {
	panic("array contains not supported for this dialect")
}
func (b *Base) AnyValueExpression(expr string) string     { return fmt.Sprintf("ANY_VALUE(%s)", expr) }
func (b *Base) MinDimensionExpression(expr string) string { return fmt.Sprintf("MIN(%s)", expr) }
func (b *Base) MaxDimensionExpression(expr string) string { return fmt.Sprintf("MAX(%s)", expr) }
func (b *Base) GetTimeDimensionParameter() string         { return "?" }

func (b *Base) SafeDivideExpression(numExpr, denExpr string) string {
	return fmt.Sprintf("(%s)/CAST(%s AS DOUBLE)", numExpr, denExpr)
}

func (b *Base) OrderByExpression(name string, desc bool) string {
	res := b.self.EscapeIdentifier(name)
	if desc {
		res += " DESC"
	}
	return res
}

func (b *Base) OrderByAliasExpression(name string, desc bool) string {
	res := b.self.EscapeAlias(name)
	if desc {
		res += " DESC"
	}
	return res
}

func (b *Base) JoinOnExpression(lhs, rhs string) string {
	return fmt.Sprintf("%s IS NOT DISTINCT FROM %s", lhs, rhs)
}

func (b *Base) SanitizeQueryForLogging(sql string) string { return sql }

func (b *Base) MetricsViewDimensionExpression(dimension *runtimev1.MetricsViewSpec_Dimension) (string, error) {
	if dimension.LookupTable != "" {
		var keyExpr string
		if dimension.Column != "" {
			keyExpr = b.self.EscapeIdentifier(dimension.Column)
		} else if dimension.Expression != "" {
			keyExpr = dimension.Expression
		} else {
			return "", fmt.Errorf("dimension %q has a lookup table but no column or expression defined", dimension.Name)
		}
		return b.self.LookupExpr(dimension.LookupTable, dimension.LookupValueColumn, keyExpr, dimension.LookupDefaultExpression)
	}
	if dimension.Expression != "" {
		return dimension.Expression, nil
	}
	if dimension.Column != "" {
		return b.self.EscapeIdentifier(dimension.Column), nil
	}
	// Backwards compatibility: column may be absent for projects that haven't re-reconciled.
	return b.self.EscapeIdentifier(dimension.Name), nil
}

func (b *Base) DimensionSelect(db, dbSchema, table string, dim *runtimev1.MetricsViewSpec_Dimension) (dimSelect, unnestClause string, err error) {
	colName := b.self.EscapeIdentifier(dim.Name)
	alias := b.self.EscapeAlias(dim.Name)
	if !dim.Unnest {
		expr, err := b.self.MetricsViewDimensionExpression(dim)
		if err != nil {
			return "", "", fmt.Errorf("failed to get dimension expression: %w", err)
		}
		return fmt.Sprintf(`(%s) AS %s`, expr, alias), "", nil
	}

	unnestColName := b.self.EscapeIdentifier(TempName(fmt.Sprintf("%s_%s_", "unnested", dim.Name)))
	unnestTableName := TempName("tbl")
	sel := fmt.Sprintf(`%s AS %s`, unnestColName, alias)
	if dim.Expression == "" {
		return sel, fmt.Sprintf(`, LATERAL UNNEST(%s.%s) %s(%s)`, b.self.EscapeTable(db, dbSchema, table), colName, unnestTableName, unnestColName), nil
	}
	return sel, fmt.Sprintf(`, LATERAL UNNEST(%s) %s(%s)`, dim.Expression, unnestTableName, unnestColName), nil
}

func (b *Base) DimensionSelectPair(db, dbSchema, table string, dim *runtimev1.MetricsViewSpec_Dimension) (expr, alias, unnestClause string, err error) {
	colAlias := b.self.EscapeAlias(dim.Name)
	if !dim.Unnest {
		ex, err := b.self.MetricsViewDimensionExpression(dim)
		if err != nil {
			return "", "", "", fmt.Errorf("failed to get dimension expression: %w", err)
		}
		return ex, colAlias, "", nil
	}

	unnestColName := b.self.EscapeIdentifier(TempName(fmt.Sprintf("%s_%s_", "unnested", dim.Name)))
	unnestTableName := TempName("tbl")
	if dim.Expression == "" {
		return unnestColName, colAlias, fmt.Sprintf(`, LATERAL UNNEST(%s.%s) %s(%s)`, b.self.EscapeTable(db, dbSchema, table), colAlias, unnestTableName, unnestColName), nil
	}
	return unnestColName, colAlias, fmt.Sprintf(`, LATERAL UNNEST(%s) %s(%s)`, dim.Expression, unnestTableName, unnestColName), nil
}

func (b *Base) LateralUnnest(expr, tableAlias, colName string) (tbl string, tupleStyle, auto bool, err error) {
	return fmt.Sprintf(`LATERAL UNNEST(%s) %s(%s)`, expr, tableAlias, b.self.EscapeIdentifier(colName)), true, false, nil
}

func (b *Base) UnnestSQLSuffix(tbl string) string {
	return fmt.Sprintf(", %s", tbl)
}

func (b *Base) GetNullExpr(_ runtimev1.Type_Code) (bool, string) {
	return true, "NULL"
}

func (b *Base) GetDateTimeExpr(_ time.Time) (bool, string) {
	return false, ""
}

func (b *Base) GetDateExpr(_ time.Time) (bool, string) {
	return false, ""
}

func (b *Base) GetArgExpr(val any, typ runtimev1.Type_Code) (string, any, error) {
	if typ == runtimev1.Type_CODE_DATE {
		t, ok := val.(time.Time)
		if !ok {
			return "", nil, fmt.Errorf("could not cast value %v to time.Time for date type", val)
		}
		return "CAST(? AS DATE)", t.Format(time.DateOnly), nil
	}
	return "?", val, nil
}

func (b *Base) GetValExpr(val any, typ runtimev1.Type_Code) (bool, string, error) {
	if val == nil {
		ok, expr := b.self.GetNullExpr(typ)
		if ok {
			return true, expr, nil
		}
		return false, "", fmt.Errorf("could not get null expr for type %q", typ)
	}
	switch typ {
	case runtimev1.Type_CODE_STRING:
		if s, ok := val.(string); ok {
			return true, b.EscapeStringValue(s), nil
		}
		return false, "", fmt.Errorf("could not cast value %v to string type", val)
	case runtimev1.Type_CODE_INT8, runtimev1.Type_CODE_INT16, runtimev1.Type_CODE_INT32, runtimev1.Type_CODE_INT64,
		runtimev1.Type_CODE_UINT8, runtimev1.Type_CODE_UINT16, runtimev1.Type_CODE_UINT32, runtimev1.Type_CODE_UINT64,
		runtimev1.Type_CODE_FLOAT32, runtimev1.Type_CODE_FLOAT64:
		if f, ok := val.(float64); ok && (math.IsNaN(f) || math.IsInf(f, 0)) {
			return true, "NULL", nil
		}
		return true, fmt.Sprintf("%v", val), nil
	case runtimev1.Type_CODE_BOOL:
		return true, fmt.Sprintf("%v", val), nil
	case runtimev1.Type_CODE_TIME, runtimev1.Type_CODE_TIMESTAMP:
		if t, ok := val.(time.Time); ok {
			if ok, expr := b.self.GetDateTimeExpr(t); ok {
				return true, expr, nil
			}
			return false, "", fmt.Errorf("cannot get time expr for this dialect")
		}
		return false, "", fmt.Errorf("unsupported time type %q", typ)
	case runtimev1.Type_CODE_DATE:
		if t, ok := val.(time.Time); ok {
			if ok, expr := b.self.GetDateExpr(t); ok {
				return true, expr, nil
			}
			return false, "", fmt.Errorf("cannot get date expr for this dialect")
		}
		return false, "", fmt.Errorf("unsupported date type %q", typ)
	default:
		return false, "", fmt.Errorf("unsupported type %q", typ)
	}
}

func (b *Base) LookupExpr(_, _, _, _ string) (string, error) {
	return "", fmt.Errorf("lookup tables are not supported for this dialect")
}

func (b *Base) LookupSelectExpr(_, _ string) (string, error) {
	return "", fmt.Errorf("lookup tables are not supported for this dialect")
}

func (b *Base) CastToDataType(typ runtimev1.Type_Code) (string, error) {
	return "", fmt.Errorf("CastToDataType not implemented for this dialect (type: %q)", typ.String())
}

func (b *Base) DateTruncExpr(_ *runtimev1.MetricsViewSpec_Dimension, _ runtimev1.TimeGrain, _ string, _, _ int) (string, error) {
	return "", fmt.Errorf("DateTruncExpr not implemented for this dialect")
}

func (b *Base) DateDiff(_ runtimev1.TimeGrain, _, _ time.Time) (string, error) {
	return "", fmt.Errorf("DateDiff not implemented for this dialect")
}

func (b *Base) IntervalSubtract(_, _ string, _ runtimev1.TimeGrain) (string, error) {
	return "", fmt.Errorf("IntervalSubtract not implemented for this dialect")
}

func (b *Base) SelectTimeRangeBins(_, _ time.Time, _ runtimev1.TimeGrain, _ string, _ *time.Location, _, _ int) (string, []any, error) {
	return "", nil, fmt.Errorf("SelectTimeRangeBins not implemented for this dialect")
}
