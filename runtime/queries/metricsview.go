package queries

import (
	"context"
	"fmt"
	"math"
	"math/big"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/marcboeker/go-duckdb"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

func lookupMetricsView(ctx context.Context, rt *runtime.Runtime, instanceID, name string) (*runtimev1.MetricsView, error) {
	obj, err := rt.GetCatalogEntry(ctx, instanceID, name)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if obj.GetMetricsView() == nil {
		return nil, status.Errorf(codes.NotFound, "object named '%s' is not a metrics view", name)
	}

	return obj.GetMetricsView(), nil
}

func metricsQuery(ctx context.Context, olap drivers.OLAPStore, priority int, sql string, args []any) ([]*runtimev1.MetricsViewColumn, []*structpb.Struct, error) {
	rows, err := olap.Execute(ctx, &drivers.Statement{
		Query:    sql,
		Args:     args,
		Priority: priority,
	})
	if err != nil {
		return nil, nil, status.Error(codes.InvalidArgument, err.Error())
	}
	defer rows.Close()

	data, err := rowsToData(rows)
	if err != nil {
		return nil, nil, status.Error(codes.Internal, err.Error())
	}

	return structTypeToMetricsViewColumn(rows.Schema), data, nil
}

func rowsToData(rows *drivers.Result) ([]*structpb.Struct, error) {
	var data []*structpb.Struct
	for rows.Next() {
		rowMap := make(map[string]any)
		err := rows.MapScan(rowMap)
		if err != nil {
			return nil, err
		}

		rowStruct, err := mapToPB(rowMap)
		if err != nil {
			return nil, err
		}

		data = append(data, rowStruct)
	}

	err := rows.Err()
	if err != nil {
		return nil, err
	}

	return data, nil
}

func structTypeToMetricsViewColumn(v *runtimev1.StructType) []*runtimev1.MetricsViewColumn {
	res := make([]*runtimev1.MetricsViewColumn, len(v.Fields))
	for i, f := range v.Fields {
		res[i] = &runtimev1.MetricsViewColumn{
			Name:     f.Name,
			Type:     f.Type.Code.String(),
			Nullable: f.Type.Nullable,
		}
	}
	return res
}

// valToPB converts any value to a google.protobuf.Value. It's similar to
// structpb.NewValue, but adds support for a few extra primitive types.
func valToPB(v any) (*structpb.Value, error) {
	switch v := v.(type) {
	// In addition to the extra supported types, we also override handling for
	// maps and lists since we need to use valToPB on nested fields.
	case map[string]interface{}:
		v2, err := mapToPB(v)
		if err != nil {
			return nil, err
		}
		return structpb.NewStructValue(v2), nil
	case []interface{}:
		v2, err := sliceToPB(v)
		if err != nil {
			return nil, err
		}
		return structpb.NewListValue(v2), nil
	// Handle types not handled by structpb.NewValue
	case int8:
		return structpb.NewNumberValue(float64(v)), nil
	case int16:
		return structpb.NewNumberValue(float64(v)), nil
	case uint8:
		return structpb.NewNumberValue(float64(v)), nil
	case uint16:
		return structpb.NewNumberValue(float64(v)), nil
	case time.Time:
		s := v.Format(time.RFC3339Nano)
		return structpb.NewStringValue(s), nil
	case float32:
		// Turning NaNs and Infs into nulls until frontend can deal with them as strings
		// (They don't have a native JSON representation)
		if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
			return structpb.NewNullValue(), nil
		}
		return structpb.NewNumberValue(float64(v)), nil
	case float64:
		// Turning NaNs and Infs into nulls until frontend can deal with them as strings
		// (They don't have a native JSON representation)
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return structpb.NewNullValue(), nil
		}
		return structpb.NewNumberValue(v), nil
	case *big.Int:
		// Evil cast to float until frontend can deal with bigs:
		v2, _ := new(big.Float).SetInt(v).Float64()
		return structpb.NewNumberValue(v2), nil
		// This is what we should do when frontend supports it:
		// s := v.String()
		// return structpb.NewStringValue(s), nil
	case duckdb.Interval:
		m := map[string]any{"months": v.Months, "days": v.Days, "micros": v.Micros}
		v2, err := mapToPB(m)
		if err != nil {
			return nil, err
		}
		return structpb.NewStructValue(v2), nil
	default:
		// Default handling for basic types (ints, string, etc.)
		return structpb.NewValue(v)
	}
}

// mapToPB converts a map to a google.protobuf.Struct. It's similar to
// structpb.NewStruct(), but it recurses on valToPB instead of structpb.NewValue
// to add support for more types.
func mapToPB(v map[string]any) (*structpb.Struct, error) {
	x := &structpb.Struct{Fields: make(map[string]*structpb.Value, len(v))}
	for k, v := range v {
		if !utf8.ValidString(k) {
			return nil, fmt.Errorf("invalid UTF-8 in string: %q", k)
		}
		var err error
		x.Fields[k], err = valToPB(v)
		if err != nil {
			return nil, err
		}
	}
	return x, nil
}

// sliceToPB converts a map to a google.protobuf.List. It's similar to
// structpb.NewList(), but it recurses on valToPB instead of structpb.NewList
// to add support for more types.
func sliceToPB(v []interface{}) (*structpb.ListValue, error) {
	x := &structpb.ListValue{Values: make([]*structpb.Value, len(v))}
	for i, v := range v {
		var err error
		x.Values[i], err = valToPB(v)
		if err != nil {
			return nil, err
		}
	}
	return x, nil
}

// Builds clause and args for runtimev1.MetricsViewFilter
func buildFilterClauseForMetricsViewFilter(filter *runtimev1.MetricsViewFilter) (string, []any, error) {
	whereClause := ""
	var args []any

	if filter != nil && filter.Include != nil {
		clause, clauseArgs, err := buildFilterClauseForConditions(filter.Include, false)
		if err != nil {
			return "", nil, err
		}
		whereClause += clause
		args = append(args, clauseArgs...)
	}

	if filter != nil && filter.Exclude != nil {
		clause, clauseArgs, err := buildFilterClauseForConditions(filter.Exclude, true)
		if err != nil {
			return "", nil, err
		}
		whereClause += clause
		args = append(args, clauseArgs...)
	}

	return whereClause, args, nil
}

func buildFilterClauseForConditions(conds []*runtimev1.MetricsViewFilter_Cond, exclude bool) (string, []any, error) {
	clause := ""
	var args []any

	for _, cond := range conds {
		condClause, condArgs, err := buildFilterClauseForCondition(cond, exclude)
		if err != nil {
			return "", nil, fmt.Errorf("filter error: %w", err)
		}
		if condClause == "" {
			continue
		}
		clause += condClause
		args = append(args, condArgs...)
	}

	return clause, args, nil
}

func buildFilterClauseForCondition(cond *runtimev1.MetricsViewFilter_Cond, exclude bool) (string, []any, error) {
	var clauses []string
	var args []any

	var operatorPrefix string
	var conditionJoiner string
	if exclude {
		operatorPrefix = " NOT "
		conditionJoiner = ") AND ("
	} else {
		operatorPrefix = ""
		conditionJoiner = " OR "
	}

	if len(cond.In) > 0 {
		// null values should be added with IS NULL / IS NOT NULL
		nullCount := 0
		for _, val := range cond.In {
			if _, ok := val.Kind.(*structpb.Value_NullValue); ok {
				nullCount++
				continue
			}
			arg, err := protobufValueToAny(val)
			if err != nil {
				return "", nil, fmt.Errorf("filter error: %w", err)
			}
			args = append(args, arg)
		}

		questionMarks := strings.Join(repeatString("?", len(cond.In)-nullCount), ",")
		// <dimension> (NOT) IN (?,?,...)
		if questionMarks != "" {
			clauses = append(clauses, fmt.Sprintf("%s %s IN (%s)", cond.Name, operatorPrefix, questionMarks))
		}
		if nullCount > 0 {
			// <dimension> IS (NOT) NULL
			clauses = append(clauses, fmt.Sprintf("%s IS %s NULL", cond.Name, operatorPrefix))
		}
	}

	if len(cond.Like) > 0 {
		for _, val := range cond.Like {
			arg, err := protobufValueToAny(val)
			if err != nil {
				return "", nil, fmt.Errorf("filter error: %w", err)
			}
			args = append(args, arg)
			// <dimension> (NOT) ILIKE ?
			clauses = append(clauses, fmt.Sprintf("%s %s ILIKE ?", cond.Name, operatorPrefix))
		}
	}

	clause := ""
	if len(clauses) > 0 {
		clause = fmt.Sprintf(" AND (%s)", strings.Join(clauses, conditionJoiner))
	}
	return clause, args, nil
}

func repeatString(val string, n int) []string {
	res := make([]string, n)
	for i := 0; i < n; i++ {
		res[i] = val
	}
	return res
}

func protobufValueToAny(val *structpb.Value) (any, error) {
	switch v := val.GetKind().(type) {
	case *structpb.Value_StringValue:
		return v.StringValue, nil
	case *structpb.Value_BoolValue:
		return v.BoolValue, nil
	case *structpb.Value_NumberValue:
		return v.NumberValue, nil
	case *structpb.Value_NullValue:
		return nil, nil
	default:
		return nil, fmt.Errorf("value not supported: %v", v)
	}
}
