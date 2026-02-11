package queries

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/apache/arrow-go/v18/arrow"
	"github.com/apache/arrow-go/v18/arrow/array"
	"github.com/apache/arrow-go/v18/arrow/memory"
	"github.com/apache/arrow-go/v18/parquet/pqarrow"
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

var ErrForbidden = errors.New("action not allowed")

// resolveMVAndSecurityFromAttributes resolves the metrics view and security policy from the attributes
func resolveMVAndSecurityFromAttributes(ctx context.Context, rt *runtime.Runtime, instanceID, metricsViewName string, claims *runtime.SecurityClaims) (*runtimev1.MetricsViewState, *runtime.ResolvedSecurity, error) {
	res, mv, err := lookupMetricsView(ctx, rt, instanceID, metricsViewName)
	if err != nil {
		return nil, nil, err
	}

	resolvedSecurity, err := rt.ResolveSecurity(ctx, instanceID, claims, res)
	if err != nil {
		return nil, nil, err
	}

	if !resolvedSecurity.CanAccess() {
		return nil, nil, ErrForbidden
	}

	return mv, resolvedSecurity, nil
}

// returns the metrics view and the time the catalog was last updated
func lookupMetricsView(ctx context.Context, rt *runtime.Runtime, instanceID, name string) (*runtimev1.Resource, *runtimev1.MetricsViewState, error) {
	ctrl, err := rt.Controller(ctx, instanceID)
	if err != nil {
		return nil, nil, status.Error(codes.InvalidArgument, err.Error())
	}

	res, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: name}, false)
	if err != nil {
		return nil, nil, status.Error(codes.InvalidArgument, err.Error())
	}

	mv := res.GetMetricsView()
	if mv.State.ValidSpec == nil {
		return nil, nil, status.Errorf(codes.InvalidArgument, "metrics view %q is invalid", name)
	}

	return res, mv.State, nil
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

func WriteCSV(meta []*runtimev1.MetricsViewColumn, data []*structpb.Struct, writer io.Writer) error {
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

func WriteXLSX(meta []*runtimev1.MetricsViewColumn, data []*structpb.Struct, writer io.Writer) error {
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

func WriteParquet(meta []*runtimev1.MetricsViewColumn, data []*structpb.Struct, ioWriter io.Writer) error {
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
		case runtimev1.Type_CODE_STRUCT, runtimev1.Type_CODE_UUID, runtimev1.Type_CODE_ARRAY, runtimev1.Type_CODE_STRING, runtimev1.Type_CODE_MAP, runtimev1.Type_CODE_INTERVAL:
			arrowField.Type = arrow.BinaryTypes.String
		case runtimev1.Type_CODE_TIMESTAMP, runtimev1.Type_CODE_DATE, runtimev1.Type_CODE_TIME:
			arrowField.Type = arrow.FixedWidthTypes.Timestamp_us
		case runtimev1.Type_CODE_BYTES:
			arrowField.Type = arrow.BinaryTypes.Binary
		}
		fields = append(fields, arrowField)
	}
	schema := arrow.NewSchema(fields, nil)

	mem := memory.DefaultAllocator
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
			case runtimev1.Type_CODE_INTERVAL:
				switch v := v.GetKind().(type) {
				case *structpb.Value_NumberValue:
					s := fmt.Sprintf("%f", v.NumberValue)
					recordBuilder.Field(idx).(*array.StringBuilder).Append(s)
				case *structpb.Value_StringValue:
					recordBuilder.Field(idx).(*array.StringBuilder).Append(v.StringValue)
				default:
				}
			}
		}
	}

	parquetwriter, err := pqarrow.NewFileWriter(schema, ioWriter, nil, pqarrow.ArrowWriterProperties{})
	if err != nil {
		return err
	}

	defer parquetwriter.Close()

	rec := recordBuilder.NewRecordBatch()
	err = parquetwriter.Write(rec)
	return err
}

func DuckDBCopyExport(ctx context.Context, w io.Writer, opts *runtime.ExportOptions, sql string, args []any, filename string, olap drivers.OLAPStore, exportFormat runtimev1.ExportFormat) error {
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

	rows, err := olap.Query(ctx, &drivers.Statement{
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
