package file

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/apache/arrow/go/v14/arrow"
	"github.com/apache/arrow/go/v14/arrow/array"
	"github.com/apache/arrow/go/v14/arrow/memory"
	"github.com/apache/arrow/go/v14/parquet/pqarrow"
	"github.com/mitchellh/mapstructure"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/jsonval"
	"github.com/xuri/excelize/v2"
)

type olapToSelfExecutor struct {
	c    *connection
	olap drivers.OLAPStore
	opts *drivers.ModelExecutorOptions
}

var _ drivers.ModelExecutor = &olapToSelfExecutor{}

func (e *olapToSelfExecutor) Execute(ctx context.Context) (*drivers.ModelResult, error) {
	// Parse SQL from input properties
	inputProps := &struct {
		SQL  string `mapstructure:"sql"`
		Args []any  `mapstructure:"args"`
	}{}
	if err := mapstructure.WeakDecode(e.opts.InputProperties, inputProps); err != nil {
		return nil, fmt.Errorf("failed to parse input properties: %w", err)
	}
	if inputProps.SQL == "" {
		return nil, errors.New("missing SQL in input properties")
	}

	// Parse output properties
	outputProps := &ModelOutputProperties{}
	if err := mapstructure.WeakDecode(e.opts.OutputProperties, outputProps); err != nil {
		return nil, fmt.Errorf("failed to parse output properties: %w", err)
	}
	if err := outputProps.Validate(); err != nil {
		return nil, fmt.Errorf("invalid output properties: %w", err)
	}

	// Execute the SQL
	res, err := e.olap.Execute(ctx, &drivers.Statement{
		Query:    inputProps.SQL,
		Args:     inputProps.Args,
		Priority: e.opts.Priority,
	})
	if err != nil {
		return nil, err
	}
	defer res.Close()

	switch outputProps.Format {
	case drivers.FileFormatParquet:
		err = writeParquet(res, outputProps.Path)
	case drivers.FileFormatCSV:
		err = writeCSV(res, outputProps.Path)
	case drivers.FileFormatJSON:
		return nil, errors.New("json file output not currently supported")
	case drivers.FileFormatXLSX:
		err = writeXLSX(res, outputProps.Path)
	default:
		return nil, fmt.Errorf("unsupported output format %q", outputProps.Format)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to write format %q: %w", outputProps.Format, err)
	}

	// Build result props
	resultProps := &ModelResultProperties{
		Path:   outputProps.Path,
		Format: outputProps.Format,
	}
	resultPropsMap := map[string]any{}
	err = mapstructure.WeakDecode(resultProps, &resultPropsMap)
	if err != nil {
		return nil, fmt.Errorf("failed to encode result properties: %w", err)
	}
	return &drivers.ModelResult{
		Connector:  e.opts.OutputConnector,
		Properties: resultPropsMap,
	}, nil
}

func writeCSV(res *drivers.Result, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)

	strs := make([]string, len(res.Schema.Fields))
	for i, f := range res.Schema.Fields {
		strs[i] = f.Name
	}
	err = w.Write(strs)
	if err != nil {
		return err
	}

	vals := make([]any, len(res.Schema.Fields))
	for i := range vals {
		vals[i] = new(any)
	}

	for res.Next() {
		err := res.Scan(vals...)
		if err != nil {
			return err
		}

		for i, v := range vals {
			v := *(v.(*any))

			v, err := jsonval.ToValue(v, res.Schema.Fields[i].Type)
			if err != nil {
				return fmt.Errorf("failed to convert to JSON value: %w", err)
			}

			var s string
			if v != nil {
				tmp, err := json.Marshal(v)
				if err != nil {
					return fmt.Errorf("failed to marshal JSON value: %w", err)
				}
				s = string(tmp)
			}

			strs[i] = s
		}

		err = w.Write(strs)
		if err != nil {
			return err
		}
	}
	if res.Err() != nil {
		return res.Err()
	}

	w.Flush()
	return nil
}

func writeXLSX(res *drivers.Result, path string) error {
	xf := excelize.NewFile()
	defer func() { _ = xf.Close() }()

	sw, err := xf.NewStreamWriter("Sheet1")
	if err != nil {
		return err
	}

	row := make([]any, len(res.Schema.Fields))
	for i, f := range res.Schema.Fields {
		row[i] = f.Name
	}
	if err := sw.SetRow("A1", row, excelize.RowOpts{Height: 45, Hidden: false}); err != nil {
		return err
	}

	vals := make([]any, len(res.Schema.Fields))
	for i := range vals {
		vals[i] = new(any)
	}

	idx := 2 // 1-based, and +1 for headers
	for res.Next() {
		err := res.Scan(vals...)
		if err != nil {
			return err
		}

		for i, v := range vals {
			v := *(v.(*any))
			res, err := jsonval.ToValue(v, res.Schema.Fields[i].Type)
			if err != nil {
				return fmt.Errorf("failed to convert to JSON value: %w", err)
			}

			switch res.(type) {
			case nil:
				res = ""
			case []any, map[string]any:
				res, err = json.Marshal(res)
				if err != nil {
					return fmt.Errorf("failed to marshal JSON value: %w", err)
				}
			}

			row[i] = res
		}

		cell, err := excelize.CoordinatesToCellName(1, idx)
		if err != nil {
			return err
		}
		if err := sw.SetRow(cell, row); err != nil {
			return err
		}

		idx++
	}
	if res.Err() != nil {
		return res.Err()
	}

	if err := sw.Flush(); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	err = xf.Write(f)
	if err != nil {
		return err
	}
	return nil
}

func writeParquet(res *drivers.Result, path string) error {
	fields := make([]arrow.Field, 0, len(res.Schema.Fields))
	for _, f := range res.Schema.Fields {
		arrowField := arrow.Field{}
		arrowField.Name = f.Name
		switch f.Type.Code {
		case runtimev1.Type_CODE_BOOL:
			arrowField.Type = arrow.FixedWidthTypes.Boolean
		case runtimev1.Type_CODE_INT8, runtimev1.Type_CODE_INT16, runtimev1.Type_CODE_INT32, runtimev1.Type_CODE_INT64:
			arrowField.Type = arrow.PrimitiveTypes.Int64
		case runtimev1.Type_CODE_INT128, runtimev1.Type_CODE_INT256:
			arrowField.Type = arrow.PrimitiveTypes.Float64
		case runtimev1.Type_CODE_UINT8, runtimev1.Type_CODE_UINT16, runtimev1.Type_CODE_UINT32, runtimev1.Type_CODE_UINT64:
			arrowField.Type = arrow.PrimitiveTypes.Uint64
		case runtimev1.Type_CODE_UINT128, runtimev1.Type_CODE_UINT256:
			arrowField.Type = arrow.PrimitiveTypes.Float64
		case runtimev1.Type_CODE_FLOAT32, runtimev1.Type_CODE_FLOAT64:
			arrowField.Type = arrow.PrimitiveTypes.Float64
		case runtimev1.Type_CODE_DECIMAL:
			arrowField.Type = arrow.PrimitiveTypes.Float64
		case runtimev1.Type_CODE_TIMESTAMP, runtimev1.Type_CODE_TIME:
			arrowField.Type = arrow.FixedWidthTypes.Timestamp_us
		case runtimev1.Type_CODE_STRING, runtimev1.Type_CODE_DATE, runtimev1.Type_CODE_ARRAY, runtimev1.Type_CODE_STRUCT, runtimev1.Type_CODE_MAP, runtimev1.Type_CODE_JSON, runtimev1.Type_CODE_UUID:
			arrowField.Type = arrow.BinaryTypes.String
		case runtimev1.Type_CODE_BYTES:
			arrowField.Type = arrow.BinaryTypes.Binary
		}
		fields = append(fields, arrowField)
	}
	schema := arrow.NewSchema(fields, nil)
	mem := memory.NewCheckedAllocator(memory.NewGoAllocator())
	recordBuilder := array.NewRecordBuilder(mem, schema)
	defer recordBuilder.Release()

	vals := make([]any, len(res.Schema.Fields))
	for i := range vals {
		vals[i] = new(any)
	}

	for res.Next() {
		err := res.Scan(vals...)
		if err != nil {
			return err
		}

		for i, v := range vals {
			t := res.Schema.Fields[i].Type
			v := *(v.(*any))
			v, err := jsonval.ToValue(v, res.Schema.Fields[i].Type)
			if err != nil {
				return fmt.Errorf("failed to convert to JSON value: %w", err)
			}

			switch t.Code {
			case runtimev1.Type_CODE_BOOL:
				v, _ := v.(bool)
				recordBuilder.Field(i).(*array.BooleanBuilder).Append(v)
			case runtimev1.Type_CODE_INT8, runtimev1.Type_CODE_INT16, runtimev1.Type_CODE_INT32, runtimev1.Type_CODE_INT64:
				v, _ := v.(int64)
				recordBuilder.Field(i).(*array.Int64Builder).Append(v)
			case runtimev1.Type_CODE_INT128, runtimev1.Type_CODE_INT256:
				v, _ := v.(float64)
				recordBuilder.Field(i).(*array.Float64Builder).Append(v)
			case runtimev1.Type_CODE_UINT8, runtimev1.Type_CODE_UINT16, runtimev1.Type_CODE_UINT32, runtimev1.Type_CODE_UINT64:
				v, _ := v.(uint64)
				recordBuilder.Field(i).(*array.Uint64Builder).Append(v)
			case runtimev1.Type_CODE_UINT128, runtimev1.Type_CODE_UINT256:
				v, _ := v.(float64)
				recordBuilder.Field(i).(*array.Float64Builder).Append(v)
			case runtimev1.Type_CODE_FLOAT32, runtimev1.Type_CODE_FLOAT64:
				v, _ := v.(float64)
				recordBuilder.Field(i).(*array.Float64Builder).Append(v)
			case runtimev1.Type_CODE_DECIMAL:
				v, _ := v.(float64)
				recordBuilder.Field(i).(*array.Float64Builder).Append(v)
			case runtimev1.Type_CODE_TIMESTAMP, runtimev1.Type_CODE_TIME:
				v, _ := v.(time.Time)
				tmp, err := arrow.TimestampFromTime(v, arrow.Microsecond)
				if err != nil {
					return err
				}
				recordBuilder.Field(i).(*array.TimestampBuilder).Append(tmp)
			case runtimev1.Type_CODE_STRING, runtimev1.Type_CODE_DATE, runtimev1.Type_CODE_ARRAY, runtimev1.Type_CODE_STRUCT, runtimev1.Type_CODE_MAP, runtimev1.Type_CODE_JSON, runtimev1.Type_CODE_UUID:
				res, err := json.Marshal(v)
				if err != nil {
					return fmt.Errorf("failed to convert to JSON value: %w", err)
				}
				recordBuilder.Field(i).(*array.StringBuilder).Append(string(res))
			case runtimev1.Type_CODE_BYTES:
				v, _ := v.([]byte)
				recordBuilder.Field(i).(*array.BinaryBuilder).Append(v)
			}
		}
	}
	if res.Err() != nil {
		return res.Err()
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	parquetwriter, err := pqarrow.NewFileWriter(schema, f, nil, pqarrow.ArrowWriterProperties{})
	if err != nil {
		return err
	}
	defer parquetwriter.Close()

	return parquetwriter.Write(recordBuilder.NewRecord())
}
