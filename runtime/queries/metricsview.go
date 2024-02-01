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
	"github.com/rilldata/rill/runtime/pkg/expressionpb"
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

func columnIdentifierExpression(mv *runtimev1.MetricsViewSpec, aliases []*runtimev1.MetricsViewComparisonMeasureAlias, name string, dialect drivers.Dialect) (string, bool) {
	// check if identifier is a dimension
	for _, dim := range mv.Dimensions {
		if dim.Name == name {
			return metricsViewDimensionExpression(dim), true
		}
	}

	// check if identifier is passed as an alias
	for _, alias := range aliases {
		if alias.Alias == name {
			switch alias.Type {
			case runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_UNSPECIFIED,
				runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_BASE_VALUE:
				return safeName(alias.Name), true
			case runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_COMPARISON_VALUE:
				return safeName(alias.Name + "__previous"), true
			case runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_ABS_DELTA:
				return safeName(alias.Name + "__delta_abs"), true
			case runtimev1.MetricsViewComparisonMeasureType_METRICS_VIEW_COMPARISON_MEASURE_TYPE_REL_DELTA:
				return safeName(alias.Name + "__delta_rel"), true
			}
		}
	}

	// check if identifier is measure but not passed as alias
	for _, mes := range mv.Measures {
		if mes.Name == name {
			return safeName(mes.Name), true
		}
	}

	return "", false
}

func identifierIsUnnest(mv *runtimev1.MetricsViewSpec, expr *runtimev1.Expression) bool {
	ident, isIdent := expr.Expression.(*runtimev1.Expression_Ident)
	if isIdent {
		for _, dim := range mv.Dimensions {
			if dim.Name == ident.Ident {
				return dim.Unnest
			}
		}
	}
	return false
}

func dimensionSelect(table string, dim *runtimev1.MetricsViewSpec_DimensionV2, dialect drivers.Dialect) (dimSelect, unnestClause string) {
	colName := safeName(dim.Name)
	if !dim.Unnest || dialect == drivers.DialectDruid {
		return fmt.Sprintf(`(%s) as %s`, metricsViewDimensionExpression(dim), colName), ""
	}

	unnestColName := safeName(tempName(fmt.Sprintf("%s_%s_", "unnested", dim.Name)))
	sel := fmt.Sprintf(`%s as %s`, unnestColName, colName)
	if dim.Expression == "" {
		// select "unnested_colName" as "colName" ... FROM "mv_table", LATERAL UNNEST("mv_table"."colName") tbl("unnested_colName") ...
		return sel, fmt.Sprintf(`, LATERAL UNNEST(%s.%s) tbl(%s)`, safeName(table), colName, unnestColName)
	}

	return sel, fmt.Sprintf(`, LATERAL UNNEST(%s) tbl(%s)`, dim.Expression, unnestColName)
}

func buildExpression(mv *runtimev1.MetricsViewSpec, expr *runtimev1.Expression, aliases []*runtimev1.MetricsViewComparisonMeasureAlias, dialect drivers.Dialect) (string, []any, error) {
	if expr == nil {
		return "", nil, nil
	}

	switch e := expr.Expression.(type) {
	case *runtimev1.Expression_Val:
		arg, err := pbutil.FromValue(e.Val)
		if err != nil {
			return "", nil, err
		}
		return "?", []any{arg}, nil

	case *runtimev1.Expression_Ident:
		expr, isIdent := columnIdentifierExpression(mv, aliases, e.Ident, dialect)
		if !isIdent {
			return "", nil, fmt.Errorf("unknown column filter: %s", e.Ident)
		}
		return expr, nil, nil

	case *runtimev1.Expression_Cond:
		return buildConditionExpression(mv, e.Cond, aliases, dialect)
	}

	return "", nil, nil
}

func buildConditionExpression(mv *runtimev1.MetricsViewSpec, cond *runtimev1.Condition, aliases []*runtimev1.MetricsViewComparisonMeasureAlias, dialect drivers.Dialect) (string, []any, error) {
	switch cond.Op {
	case runtimev1.Operation_OPERATION_LIKE, runtimev1.Operation_OPERATION_NLIKE:
		return buildLikeExpression(mv, cond, aliases, dialect)

	case runtimev1.Operation_OPERATION_IN, runtimev1.Operation_OPERATION_NIN:
		return buildInExpression(mv, cond, aliases, dialect)

	case runtimev1.Operation_OPERATION_AND:
		return buildAndOrExpressions(mv, cond, aliases, dialect, " AND ")

	case runtimev1.Operation_OPERATION_OR:
		return buildAndOrExpressions(mv, cond, aliases, dialect, " OR ")

	default:
		leftExpr, args, err := buildExpression(mv, cond.Exprs[0], aliases, dialect)
		if err != nil {
			return "", nil, err
		}

		rightExpr, subArgs, err := buildExpression(mv, cond.Exprs[1], aliases, dialect)
		if err != nil {
			return "", nil, err
		}
		args = append(args, subArgs...)

		return fmt.Sprintf("(%s) %s (%s)", leftExpr, conditionExpressionOperation(cond.Op), rightExpr), args, nil
	}
}

func buildLikeExpression(mv *runtimev1.MetricsViewSpec, cond *runtimev1.Condition, aliases []*runtimev1.MetricsViewComparisonMeasureAlias, dialect drivers.Dialect) (string, []any, error) {
	if len(cond.Exprs) != 2 {
		return "", nil, fmt.Errorf("like/not like expression should have exactly 2 sub expressions")
	}

	leftExpr, args, err := buildExpression(mv, cond.Exprs[0], aliases, dialect)
	if err != nil {
		return "", nil, err
	}

	rightExpr, subArgs, err := buildExpression(mv, cond.Exprs[1], aliases, dialect)
	if err != nil {
		return "", nil, err
	}
	args = append(args, subArgs...)

	notKeyword := ""
	if cond.Op == runtimev1.Operation_OPERATION_NLIKE {
		notKeyword = "NOT"
	}

	// identify if immediate identifier has unnest
	unnest := identifierIsUnnest(mv, cond.Exprs[0])

	var clause string
	// Build [NOT] len(list_filter("dim", x -> x ILIKE ?)) > 0
	if unnest && dialect != drivers.DialectDruid {
		clause = fmt.Sprintf("%s len(list_filter((%s), x -> x ILIKE %s)) > 0", notKeyword, leftExpr, rightExpr)
	} else {
		if dialect == drivers.DialectDruid {
			// Druid does not support ILIKE
			clause = fmt.Sprintf("LOWER(%s) %s LIKE LOWER(CAST(%s AS VARCHAR))", leftExpr, notKeyword, rightExpr)
		} else {
			clause = fmt.Sprintf("(%s) %s ILIKE %s", leftExpr, notKeyword, rightExpr)
		}
	}

	// When you have "dim NOT ILIKE '...'", then NULL values are always excluded.
	// We need to explicitly include it.
	if cond.Op == runtimev1.Operation_OPERATION_NLIKE {
		clause += fmt.Sprintf(" OR (%s) IS NULL", leftExpr)
	}

	return clause, args, nil
}

func buildInExpression(mv *runtimev1.MetricsViewSpec, cond *runtimev1.Condition, aliases []*runtimev1.MetricsViewComparisonMeasureAlias, dialect drivers.Dialect) (string, []any, error) {
	if len(cond.Exprs) <= 1 {
		return "", nil, fmt.Errorf("in/not in expression should have atleast 2 sub expressions")
	}

	leftExpr, args, err := buildExpression(mv, cond.Exprs[0], aliases, dialect)
	if err != nil {
		return "", nil, err
	}

	notKeyword := ""
	exclude := cond.Op == runtimev1.Operation_OPERATION_NIN
	if exclude {
		notKeyword = "NOT"
	}

	inHasNull := false
	var valClauses []string
	// Add to args, skipping nulls
	for _, subExpr := range cond.Exprs[1:] {
		if v, isVal := subExpr.Expression.(*runtimev1.Expression_Val); isVal {
			if _, isNull := v.Val.Kind.(*structpb.Value_NullValue); isNull {
				inHasNull = true
				continue // Handled later using "dim IS [NOT] NULL" clause
			}
		}
		inVal, subArgs, err := buildExpression(mv, subExpr, aliases, dialect)
		if err != nil {
			return "", nil, err
		}
		args = append(args, subArgs...)
		valClauses = append(valClauses, inVal)
	}

	// identify if immediate identifier has unnest
	unnest := identifierIsUnnest(mv, cond.Exprs[0])

	clauses := make([]string, 0)

	// If there were non-null args, add a "dim [NOT] IN (...)" clause
	if len(valClauses) > 0 {
		questionMarks := strings.Join(valClauses, ",")
		var clause string
		// Build [NOT] list_has_any("dim", ARRAY[?, ?, ...])
		if unnest && dialect != drivers.DialectDruid {
			clause = fmt.Sprintf("%s list_has_any((%s), ARRAY[%s])", notKeyword, leftExpr, questionMarks)
		} else {
			clause = fmt.Sprintf("(%s) %s IN (%s)", leftExpr, notKeyword, questionMarks)
		}
		clauses = append(clauses, clause)
	}

	if inHasNull {
		// Add null check
		// NOTE: DuckDB doesn't handle NULL values in an "IN" expression. They must be checked with a "dim IS [NOT] NULL" clause.
		clauses = append(clauses, fmt.Sprintf("(%s) IS %s NULL", leftExpr, notKeyword))
	}
	var condsClause string
	if exclude {
		condsClause = strings.Join(clauses, " AND ")
	} else {
		condsClause = strings.Join(clauses, " OR ")
	}
	if exclude && !inHasNull && len(clauses) > 0 {
		// When you have "dim NOT IN (a, b, ...)", then NULL values are always excluded, even if NULL is not in the list.
		// E.g. this returns zero rows: "select * from (select 1 as a union select null as a) where a not in (1)"
		// We need to explicitly include it.
		condsClause += fmt.Sprintf(" OR (%s) IS NULL", leftExpr)
	}

	return condsClause, args, nil
}

func buildAndOrExpressions(mv *runtimev1.MetricsViewSpec, cond *runtimev1.Condition, aliases []*runtimev1.MetricsViewComparisonMeasureAlias, dialect drivers.Dialect, joiner string) (string, []any, error) {
	if len(cond.Exprs) == 0 {
		return "", nil, fmt.Errorf("or/and expression should have atleast 1 sub expressions")
	}

	clauses := make([]string, 0)
	var args []any
	for _, expr := range cond.Exprs {
		clause, subArgs, err := buildExpression(mv, expr, aliases, dialect)
		if err != nil {
			return "", nil, err
		}
		args = append(args, subArgs...)
		clauses = append(clauses, fmt.Sprintf("(%s)", clause))
	}
	return strings.Join(clauses, joiner), args, nil
}

func conditionExpressionOperation(oprn runtimev1.Operation) string {
	switch oprn {
	case runtimev1.Operation_OPERATION_EQ:
		return "="
	case runtimev1.Operation_OPERATION_NEQ:
		return "!="
	case runtimev1.Operation_OPERATION_LT:
		return "<"
	case runtimev1.Operation_OPERATION_LTE:
		return "<="
	case runtimev1.Operation_OPERATION_GT:
		return ">"
	case runtimev1.Operation_OPERATION_GTE:
		return ">="
	}
	panic(fmt.Sprintf("unknown condition operation: %v", oprn))
}

func convertFilterToExpression(filter *runtimev1.MetricsViewFilter) *runtimev1.Expression {
	var exprs []*runtimev1.Expression

	if len(filter.Include) > 0 {
		for _, cond := range filter.Include {
			domExpr := convertDimensionFilterToExpression(cond, false)
			if domExpr != nil {
				exprs = append(exprs, domExpr)
			}
		}
	}

	if len(filter.Exclude) > 0 {
		for _, cond := range filter.Exclude {
			domExpr := convertDimensionFilterToExpression(cond, true)
			if domExpr != nil {
				exprs = append(exprs, domExpr)
			}
		}
	}

	if len(exprs) == 1 {
		return exprs[0]
	} else if len(exprs) > 1 {
		return expressionpb.And(exprs)
	}
	return nil
}

func convertDimensionFilterToExpression(cond *runtimev1.MetricsViewFilter_Cond, exclude bool) *runtimev1.Expression {
	var inExpr *runtimev1.Expression
	if len(cond.In) > 0 {
		var inExprs []*runtimev1.Expression
		for _, inVal := range cond.In {
			inExprs = append(inExprs, expressionpb.Value(inVal))
		}
		if exclude {
			inExpr = expressionpb.NotIn(expressionpb.Identifier(cond.Name), inExprs)
		} else {
			inExpr = expressionpb.In(expressionpb.Identifier(cond.Name), inExprs)
		}
	}

	var likeExpr *runtimev1.Expression
	if len(cond.Like) == 1 {
		if exclude {
			likeExpr = expressionpb.NotLike(expressionpb.Identifier(cond.Name), expressionpb.Value(structpb.NewStringValue(cond.Like[0])))
		} else {
			likeExpr = expressionpb.Like(expressionpb.Identifier(cond.Name), expressionpb.Value(structpb.NewStringValue(cond.Like[0])))
		}
	} else if len(cond.Like) > 1 {
		var likeExprs []*runtimev1.Expression
		for _, l := range cond.Like {
			col := expressionpb.Identifier(cond.Name)
			val := expressionpb.Value(structpb.NewStringValue(l))
			if exclude {
				likeExprs = append(likeExprs, expressionpb.NotLike(col, val))
			} else {
				likeExprs = append(likeExprs, expressionpb.Like(col, val))
			}
		}
		if exclude {
			likeExpr = expressionpb.And(likeExprs)
		} else {
			likeExpr = expressionpb.Or(likeExprs)
		}
	}

	if inExpr != nil && likeExpr != nil {
		if exclude {
			return expressionpb.And([]*runtimev1.Expression{inExpr, likeExpr})
		}
		return expressionpb.Or([]*runtimev1.Expression{inExpr, likeExpr})
	} else if inExpr != nil {
		return inExpr
	} else if likeExpr != nil {
		return likeExpr
	}

	return nil
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

func metricsViewDimension(mv *runtimev1.MetricsViewSpec, dimName string) (*runtimev1.MetricsViewSpec_DimensionV2, error) {
	for _, dimension := range mv.Dimensions {
		if strings.EqualFold(dimension.Name, dimName) {
			return dimension, nil
		}
	}
	return nil, fmt.Errorf("dimension %s not found", dimName)
}

func metricsViewDimensionExpression(dimension *runtimev1.MetricsViewSpec_DimensionV2) string {
	if dimension.Expression != "" {
		return dimension.Expression
	}
	if dimension.Column != "" {
		return safeName(dimension.Column)
	}
	// backwards compatibility for older projects that have not run reconcile on this dashboard
	// in that case `column` will not be present
	return safeName(dimension.Name)
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
	if q.TimeStart != nil || q.TimeEnd != nil || q.Where != nil {
		filename += "_filtered"
	}
	return filename
}
