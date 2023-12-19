package queries

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/apache/arrow/go/v14/arrow"
	"github.com/apache/arrow/go/v14/arrow/array"
	"github.com/apache/arrow/go/v14/arrow/memory"
	"github.com/apache/arrow/go/v14/parquet/pqarrow"
	"github.com/google/uuid"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/pbutil"
	"github.com/xuri/excelize/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

// resolveMeasures returns the selected measures
func resolveMeasures(mv *runtimev1.MetricsViewSpec, inlines []*runtimev1.InlineMeasure, selectedNames []string) ([]*runtimev1.MetricsViewSpec_MeasureV2, error) {
	// Build combined measures
	ms := make([]*runtimev1.MetricsViewSpec_MeasureV2, len(selectedNames))
	for i, n := range selectedNames {
		found := false
		// Search in the inlines (take precedence)
		for _, m := range inlines {
			if m.Name == n {
				ms[i] = &runtimev1.MetricsViewSpec_MeasureV2{
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

func olapQuery(ctx context.Context, olap drivers.OLAPStore, priority int, sql string, args []any) (*runtimev1.StructType, []*structpb.Struct, error) {
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

	return rows.Schema, data, nil
}

func rowsToData(rows *drivers.Result) ([]*structpb.Struct, error) {
	var data []*structpb.Struct
	for rows.Next() {
		rowMap := make(map[string]any)
		err := rows.MapScan(rowMap)
		if err != nil {
			return nil, err
		}

		rowStruct, err := pbutil.ToStruct(rowMap, rows.Schema)
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
func buildFilterClauseForMetricsViewFilter(mv *runtimev1.MetricsViewSpec, filter *runtimev1.MetricsViewFilter, dialect drivers.Dialect, policy *runtime.ResolvedMetricsViewSecurity) (string, []any, error) {
	var clauses []string
	var args []any

	if filter != nil && filter.Include != nil {
		clause, clauseArgs, err := buildFilterClauseForConditions(mv, filter.Include, false, dialect)
		if err != nil {
			return "", nil, err
		}
		clauses = append(clauses, clause)
		args = append(args, clauseArgs...)
	}

	if filter != nil && filter.Exclude != nil {
		clause, clauseArgs, err := buildFilterClauseForConditions(mv, filter.Exclude, true, dialect)
		if err != nil {
			return "", nil, err
		}
		clauses = append(clauses, clause)
		args = append(args, clauseArgs...)
	}

	if policy != nil && policy.RowFilter != "" {
		clauses = append(clauses, "AND "+policy.RowFilter)
	}

	return strings.Join(clauses, " "), args, nil
}

// buildFilterClauseForConditions returns a string with the format "AND (...) AND (...) ..."
func buildFilterClauseForConditions(mv *runtimev1.MetricsViewSpec, conds []*runtimev1.MetricsViewFilter_Cond, exclude bool, dialect drivers.Dialect) (string, []any, error) {
	var clauses []string
	var args []any

	for _, cond := range conds {
		condClause, condArgs, err := buildFilterClauseForCondition(mv, cond, exclude, dialect)
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
func buildFilterClauseForCondition(mv *runtimev1.MetricsViewSpec, cond *runtimev1.MetricsViewFilter_Cond, exclude bool, dialect drivers.Dialect) (string, []any, error) {
	var clauses []string
	var args []any

	// NOTE: Looking up for dimension like this will lead to O(nm).
	//       Ideal way would be to create a map, but we need to find a clean solution down the line
	dim, err := metricsViewDimension(mv, cond.Name)
	if err != nil {
		return "", nil, err
	}
	name := safeName(metricsViewDimensionColumn(dim))

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
			var clause string
			// Build [NOT] list_has_any("dim", ARRAY[?, ?, ...])
			if dim.Unnest && dialect != drivers.DialectDruid {
				clause = fmt.Sprintf("%s list_has_any(%s, ARRAY[%s])", notKeyword, name, questionMarks)
			} else {
				clause = fmt.Sprintf("%s %s IN (%s)", name, notKeyword, questionMarks)
			}
			clauses = append(clauses, clause)
		}
	}

	// Build "dim [NOT] ILIKE ?"
	if len(cond.Like) > 0 {
		for _, val := range cond.Like {
			var clause string
			// Build [NOT] len(list_filter("dim", x -> x ILIKE ?)) > 0
			if dim.Unnest && dialect != drivers.DialectDruid {
				clause = fmt.Sprintf("%s len(list_filter(%s, x -> x %s ILIKE ?)) > 0", notKeyword, name, notKeyword)
			} else {
				if dialect == drivers.DialectDruid {
					// Druid does not support ILIKE
					clause = fmt.Sprintf("LOWER(%s) %s LIKE LOWER(CAST(? AS VARCHAR))", name, notKeyword)
				} else {
					clause = fmt.Sprintf("%s %s ILIKE ?", name, notKeyword)
				}
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

func convertToString(pbvalue *structpb.Value) (string, error) {
	switch pbvalue.GetKind().(type) {
	case *structpb.Value_StructValue:
		bts, err := protojson.Marshal(pbvalue)
		if err != nil {
			return "", err
		}

		return string(bts), nil
	case *structpb.Value_NullValue:
		return "", nil
	default:
		return fmt.Sprintf("%v", pbvalue.AsInterface()), nil
	}
}

func convertToXLSXValue(pbvalue *structpb.Value) (interface{}, error) {
	switch pbvalue.GetKind().(type) {
	case *structpb.Value_StructValue:
		bts, err := protojson.Marshal(pbvalue)
		if err != nil {
			return "", err
		}

		return string(bts), nil
	case *structpb.Value_NullValue:
		return "", nil
	default:
		return pbvalue.AsInterface(), nil
	}
}

func metricsViewDimensionToSafeColumn(mv *runtimev1.MetricsViewSpec, dimName string) (string, error) {
	dimName = strings.ToLower(dimName)
	dimension, err := metricsViewDimension(mv, dimName)
	if err != nil {
		return "", err
	}
	return safeName(metricsViewDimensionColumn(dimension)), nil
}

func metricsViewDimension(mv *runtimev1.MetricsViewSpec, dimName string) (*runtimev1.MetricsViewSpec_DimensionV2, error) {
	for _, dimension := range mv.Dimensions {
		if strings.EqualFold(dimension.Name, dimName) {
			return dimension, nil
		}
	}
	return nil, fmt.Errorf("dimension %s not found", dimName)
}

func metricsViewDimensionColumn(dimension *runtimev1.MetricsViewSpec_DimensionV2) string {
	if dimension.Column != "" {
		return dimension.Column
	}
	// backwards compatibility for older projects that have not run reconcile on this dashboard
	// in that case `column` will not be present
	return dimension.Name
}

func metricsViewMeasureExpression(mv *runtimev1.MetricsViewSpec, measureName string) (string, error) {
	for _, measure := range mv.Measures {
		if strings.EqualFold(measure.Name, measureName) {
			return measure.Expression, nil
		}
	}
	return "", fmt.Errorf("measure %s not found", measureName)
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
			str, err := convertToString(pbvalue)
			if err != nil {
				return err
			}

			record = append(record, str)
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
	for i, structs := range data {
		for _, field := range meta {
			pbvalue := structs.Fields[field.Name]
			value, err := convertToXLSXValue(pbvalue)
			if err != nil {
				return err
			}

			row = append(row, value)
		}

		cell, err := excelize.CoordinatesToCellName(1, i+2) // 1-based, and +1 for headers
		if err != nil {
			return err
		}

		if err := sw.SetRow(cell, row); err != nil {
			return err
		}

		row = row[:0]
	}

	if err := sw.Flush(); err != nil {
		return err
	}

	err = f.Write(writer)

	return err
}

func writeParquet(meta []*runtimev1.MetricsViewColumn, data []*structpb.Struct, ioWriter io.Writer) error {
	fields := make([]arrow.Field, 0, len(meta))
	for _, f := range meta {
		arrowField := arrow.Field{}
		arrowField.Name = f.Name
		typeCode := runtimev1.Type_Code(runtimev1.Type_Code_value[f.Type])
		switch typeCode {
		case runtimev1.Type_CODE_BOOL:
			arrowField.Type = arrow.FixedWidthTypes.Boolean
		case runtimev1.Type_CODE_INT8:
			arrowField.Type = arrow.PrimitiveTypes.Int8
		case runtimev1.Type_CODE_INT16:
			arrowField.Type = arrow.PrimitiveTypes.Int16
		case runtimev1.Type_CODE_INT32:
			arrowField.Type = arrow.PrimitiveTypes.Int32
		case runtimev1.Type_CODE_INT64:
			arrowField.Type = arrow.PrimitiveTypes.Int64
		case runtimev1.Type_CODE_INT128:
			arrowField.Type = arrow.PrimitiveTypes.Float64
		case runtimev1.Type_CODE_UINT8:
			arrowField.Type = arrow.PrimitiveTypes.Uint8
		case runtimev1.Type_CODE_UINT16:
			arrowField.Type = arrow.PrimitiveTypes.Uint16
		case runtimev1.Type_CODE_UINT32:
			arrowField.Type = arrow.PrimitiveTypes.Uint32
		case runtimev1.Type_CODE_UINT64:
			arrowField.Type = arrow.PrimitiveTypes.Uint64
		case runtimev1.Type_CODE_DECIMAL:
			arrowField.Type = arrow.PrimitiveTypes.Float64
		case runtimev1.Type_CODE_FLOAT32:
			arrowField.Type = arrow.PrimitiveTypes.Float32
		case runtimev1.Type_CODE_FLOAT64:
			arrowField.Type = arrow.PrimitiveTypes.Float64
		case runtimev1.Type_CODE_STRUCT, runtimev1.Type_CODE_UUID, runtimev1.Type_CODE_ARRAY, runtimev1.Type_CODE_STRING, runtimev1.Type_CODE_MAP:
			arrowField.Type = arrow.BinaryTypes.String
		case runtimev1.Type_CODE_TIMESTAMP, runtimev1.Type_CODE_DATE, runtimev1.Type_CODE_TIME:
			arrowField.Type = arrow.FixedWidthTypes.Timestamp_us
		case runtimev1.Type_CODE_BYTES:
			arrowField.Type = arrow.BinaryTypes.Binary
		}
		fields = append(fields, arrowField)
	}
	schema := arrow.NewSchema(fields, nil)

	mem := memory.NewCheckedAllocator(memory.NewGoAllocator())
	recordBuilder := array.NewRecordBuilder(mem, schema)
	defer recordBuilder.Release()
	for _, s := range data {
		for idx, t := range meta {
			v := s.Fields[t.Name]
			typeCode := runtimev1.Type_Code(runtimev1.Type_Code_value[t.Type])
			switch typeCode {
			case runtimev1.Type_CODE_BOOL:
				recordBuilder.Field(idx).(*array.BooleanBuilder).Append(v.GetBoolValue())
			case runtimev1.Type_CODE_INT8:
				recordBuilder.Field(idx).(*array.Int8Builder).Append(int8(v.GetNumberValue()))
			case runtimev1.Type_CODE_INT16:
				recordBuilder.Field(idx).(*array.Int16Builder).Append(int16(v.GetNumberValue()))
			case runtimev1.Type_CODE_INT32:
				recordBuilder.Field(idx).(*array.Int32Builder).Append(int32(v.GetNumberValue()))
			case runtimev1.Type_CODE_INT64:
				recordBuilder.Field(idx).(*array.Int64Builder).Append(int64(v.GetNumberValue()))
			case runtimev1.Type_CODE_UINT8:
				recordBuilder.Field(idx).(*array.Uint8Builder).Append(uint8(v.GetNumberValue()))
			case runtimev1.Type_CODE_UINT16:
				recordBuilder.Field(idx).(*array.Uint16Builder).Append(uint16(v.GetNumberValue()))
			case runtimev1.Type_CODE_UINT32:
				recordBuilder.Field(idx).(*array.Uint32Builder).Append(uint32(v.GetNumberValue()))
			case runtimev1.Type_CODE_UINT64:
				recordBuilder.Field(idx).(*array.Uint64Builder).Append(uint64(v.GetNumberValue()))
			case runtimev1.Type_CODE_INT128:
				recordBuilder.Field(idx).(*array.Float64Builder).Append(v.GetNumberValue())
			case runtimev1.Type_CODE_FLOAT32:
				recordBuilder.Field(idx).(*array.Float32Builder).Append(float32(v.GetNumberValue()))
			case runtimev1.Type_CODE_FLOAT64, runtimev1.Type_CODE_DECIMAL:
				recordBuilder.Field(idx).(*array.Float64Builder).Append(v.GetNumberValue())
			case runtimev1.Type_CODE_STRING, runtimev1.Type_CODE_UUID:
				recordBuilder.Field(idx).(*array.StringBuilder).Append(v.GetStringValue())
			case runtimev1.Type_CODE_TIMESTAMP, runtimev1.Type_CODE_DATE, runtimev1.Type_CODE_TIME:
				tmp, err := arrow.TimestampFromString(v.GetStringValue(), arrow.Microsecond)
				if err != nil {
					return err
				}

				recordBuilder.Field(idx).(*array.TimestampBuilder).Append(tmp)
			case runtimev1.Type_CODE_ARRAY, runtimev1.Type_CODE_MAP, runtimev1.Type_CODE_STRUCT:
				bts, err := protojson.Marshal(v)
				if err != nil {
					return err
				}

				recordBuilder.Field(idx).(*array.StringBuilder).Append(string(bts))
			}
		}
	}

	parquetwriter, err := pqarrow.NewFileWriter(schema, ioWriter, nil, pqarrow.ArrowWriterProperties{})
	if err != nil {
		return err
	}

	defer parquetwriter.Close()

	rec := recordBuilder.NewRecord()
	err = parquetwriter.Write(rec)
	return err
}

func duckDBCopyExport(ctx context.Context, w io.Writer, opts *runtime.ExportOptions, sql string, args []any, filename string, olap drivers.OLAPStore, exportFormat runtimev1.ExportFormat) error {
	var extension string
	switch exportFormat {
	case runtimev1.ExportFormat_EXPORT_FORMAT_PARQUET:
		extension = "parquet"
	case runtimev1.ExportFormat_EXPORT_FORMAT_CSV:
		extension = "csv"
	}

	tmpPath := fmt.Sprintf("export_%s.%s", uuid.New().String(), extension)
	tmpPath = filepath.Join(os.TempDir(), tmpPath)
	defer os.Remove(tmpPath)

	sql = fmt.Sprintf("COPY (%s) TO '%s'", sql, tmpPath)
	if extension == "csv" {
		sql += " (FORMAT CSV, HEADER)"
	}

	rows, err := olap.Execute(ctx, &drivers.Statement{
		Query:            sql,
		Args:             args,
		Priority:         opts.Priority,
		ExecutionTimeout: defaultExecutionTimeout,
	})
	if err != nil {
		return err
	}
	defer rows.Close()

	if opts.PreWriteHook != nil {
		err = opts.PreWriteHook(filename)
		if err != nil {
			return err
		}
	}

	f, err := os.Open(tmpPath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(w, f)
	return err
}

func (q *MetricsViewRows) generateFilename(mv *runtimev1.MetricsViewSpec) string {
	filename := strings.ReplaceAll(mv.Table, `"`, `_`)
	if q.TimeStart != nil || q.TimeEnd != nil || q.Filter != nil && (len(q.Filter.Include) > 0 || len(q.Filter.Exclude) > 0) {
		filename += "_filtered"
	}
	return filename
}
