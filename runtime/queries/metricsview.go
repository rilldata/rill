package queries

import (
	"context"
	"fmt"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/server/pbutil"
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

		rowStruct, err := pbutil.ToStruct(rowMap)
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
			arg, err := pbutil.FromValue(val)
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
			arg, err := pbutil.FromValue(val)
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
