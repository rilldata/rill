package queries

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/pbutil"
	"github.com/xuri/excelize/v2"
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

// resolveMeasures returns the selected measures
func resolveMeasures(mv *runtimev1.MetricsView, inlines []*runtimev1.InlineMeasure, selectedNames []string) ([]*runtimev1.MetricsView_Measure, error) {
	// Build combined measures
	ms := make([]*runtimev1.MetricsView_Measure, len(selectedNames))
	for i, n := range selectedNames {
		found := false
		// Search in the inlines (take precedence)
		for _, m := range inlines {
			if m.Name == n {
				ms[i] = &runtimev1.MetricsView_Measure{
					Name:       m.Name,
					Expression: m.Expression,
				}
				found = true
				break
			}
		}
		if found {
			continue
		}
		// Search in the metrics view
		for _, m := range mv.Measures {
			if m.Name == n {
				ms[i] = m
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("measure does not exist: '%s'", n)
		}
	}

	return ms, nil
}

func metricsQuery(ctx context.Context, olap drivers.OLAPStore, priority int, sql string, args []any) ([]*runtimev1.MetricsViewColumn, []*structpb.Struct, error) {
	rows, err := olap.Execute(ctx, &drivers.Statement{
		Query:            sql,
		Args:             args,
		Priority:         priority,
		ExecutionTimeout: defaultExecutionTimeout,
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

// buildFilterClauseForMetricsViewFilter builds a SQL string of conditions joined with AND.
// Unless the result is empty, it is prefixed with "AND".
// I.e. it has the format "AND (...) AND (...) ...".
func buildFilterClauseForMetricsViewFilter(filter *runtimev1.MetricsViewFilter, dialect drivers.Dialect) (string, []any, error) {
	var clauses []string
	var args []any

	if filter != nil && filter.Include != nil {
		clause, clauseArgs, err := buildFilterClauseForConditions(filter.Include, false, dialect)
		if err != nil {
			return "", nil, err
		}
		clauses = append(clauses, clause)
		args = append(args, clauseArgs...)
	}

	if filter != nil && filter.Exclude != nil {
		clause, clauseArgs, err := buildFilterClauseForConditions(filter.Exclude, true, dialect)
		if err != nil {
			return "", nil, err
		}
		clauses = append(clauses, clause)
		args = append(args, clauseArgs...)
	}

	return strings.Join(clauses, " "), args, nil
}

// buildFilterClauseForConditions returns a string with the format "AND (...) AND (...) ..."
func buildFilterClauseForConditions(conds []*runtimev1.MetricsViewFilter_Cond, exclude bool, dialect drivers.Dialect) (string, []any, error) {
	var clauses []string
	var args []any

	for _, cond := range conds {
		condClause, condArgs, err := buildFilterClauseForCondition(cond, exclude, dialect)
		if err != nil {
			return "", nil, err
		}
		if condClause == "" {
			continue
		}
		clauses = append(clauses, condClause)
		args = append(args, condArgs...)
	}

	return strings.Join(clauses, " "), args, nil
}

// buildFilterClauseForCondition returns a string with the format "AND (...)"
func buildFilterClauseForCondition(cond *runtimev1.MetricsViewFilter_Cond, exclude bool, dialect drivers.Dialect) (string, []any, error) {
	var clauses []string
	var args []any

	name := safeName(cond.Name)
	notKeyword := ""
	if exclude {
		notKeyword = "NOT"
	}

	// Tracks if we found NULL(s) in cond.In
	inHasNull := false

	// Build "dim [NOT] IN (?, ?, ...)" clause
	if len(cond.In) > 0 {
		// Add to args, skipping nulls
		for _, val := range cond.In {
			if _, ok := val.Kind.(*structpb.Value_NullValue); ok {
				inHasNull = true
				continue // Handled later using "dim IS [NOT] NULL" clause
			}
			arg, err := pbutil.FromValue(val)
			if err != nil {
				return "", nil, fmt.Errorf("filter error: %w", err)
			}
			args = append(args, arg)
		}

		// If there were non-null args, add a "dim [NOT] IN (...)" clause
		if len(args) > 0 {
			questionMarks := strings.Join(repeatString("?", len(args)), ",")
			clause := fmt.Sprintf("%s %s IN (%s)", name, notKeyword, questionMarks)
			clauses = append(clauses, clause)
		}
	}

	// Build "dim [NOT] ILIKE ?"
	if len(cond.Like) > 0 {
		for _, val := range cond.Like {
			var clause string
			if dialect == drivers.DialectDruid {
				// Druid does not support ILIKE
				clause = fmt.Sprintf("LOWER(%s) %s LIKE LOWER(?)", name, notKeyword)
			} else {
				clause = fmt.Sprintf("%s %s ILIKE ?", name, notKeyword)
			}

			args = append(args, val)
			clauses = append(clauses, clause)
		}
	}

	// Add null check
	// NOTE: DuckDB doesn't handle NULL values in an "IN" expression. They must be checked with a "dim IS [NOT] NULL" clause.
	if inHasNull {
		clauses = append(clauses, fmt.Sprintf("%s IS %s NULL", name, notKeyword))
	}

	// If no checks were added, exit
	if len(clauses) == 0 {
		return "", nil, nil
	}

	// Join conditions
	var condJoiner string
	if exclude {
		condJoiner = " AND "
	} else {
		condJoiner = " OR "
	}
	condsClause := strings.Join(clauses, condJoiner)

	// When you have "dim NOT IN (a, b, ...)", then NULL values are always excluded, even if NULL is not in the list.
	// E.g. this returns zero rows: "select * from (select 1 as a union select null as a) where a not in (1)"
	// We need to explicitly include it.
	if exclude && !inHasNull && len(condsClause) > 0 {
		condsClause += fmt.Sprintf(" OR %s IS NULL", name)
	}

	// Done
	return fmt.Sprintf("AND (%s) ", condsClause), args, nil
}

func repeatString(val string, n int) []string {
	res := make([]string, n)
	for i := 0; i < n; i++ {
		res[i] = val
	}
	return res
}

func writeCSV(meta []*runtimev1.MetricsViewColumn, data []*structpb.Struct, writer io.Writer) error {
	w := csv.NewWriter(writer)

	record := make([]string, 0, len(meta))
	for _, field := range meta {
		record = append(record, field.Name)
	}
	if err := w.Write(record); err != nil {
		return err
	}
	record = record[:0]

	for _, structs := range data {
		for _, field := range meta {
			pbvalue := structs.Fields[field.Name]
			switch pbvalue.GetKind().(type) {
			case *structpb.Value_StructValue:
				bts, err := json.Marshal(pbvalue)
				if err != nil {
					return err
				}

				record = append(record, string(bts))
			case *structpb.Value_NullValue:
				record = append(record, "")
			default:
				record = append(record, fmt.Sprintf("%v", pbvalue.AsInterface()))
			}
		}

		if err := w.Write(record); err != nil {
			return err
		}

		record = record[:0]
	}

	w.Flush()

	return nil
}

func writeXLSX(meta []*runtimev1.MetricsViewColumn, data []*structpb.Struct, writer io.Writer) error {
	f := excelize.NewFile()
	defer func() {
		_ = f.Close()
	}()

	sw, err := f.NewStreamWriter("Sheet1")
	if err != nil {
		return err
	}

	headers := make([]interface{}, 0, len(meta))
	for _, v := range meta {
		headers = append(headers, v.Name)
	}

	if err := sw.SetRow("A1", headers, excelize.RowOpts{Height: 45, Hidden: false}); err != nil {
		return err
	}

	row := make([]interface{}, 0, len(meta))
	for i, s := range data {
		for _, f := range s.Fields {
			row = append(row, f.AsInterface())
		}

		cell, err := excelize.CoordinatesToCellName(1, i+2) // 1-based, and +1 for headers
		if err != nil {
			return err
		}

		if err := sw.SetRow(cell, row); err != nil {
			return err
		}
	}

	if err := sw.Flush(); err != nil {
		return err
	}

	err = f.Write(writer)

	return err
}
