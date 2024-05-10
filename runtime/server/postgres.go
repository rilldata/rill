package server

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/jackc/pgx/v5/pgtype"
	wire "github.com/jeroenrinzema/psql-wire"
	"github.com/marcboeker/go-duckdb"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

func QueryHandler(s *Server) func(ctx context.Context, query string) (wire.PreparedStatements, error) {
	return func(ctx context.Context, query string) (wire.PreparedStatements, error) {
		clientParams := wire.ClientParameters(ctx)
		instanceID := clientParams[wire.ParamDatabase]
		// todo how to handle normal SQL ?
		api, err := s.runtime.APIForName(ctx, instanceID, "metrics-sql")
		if err != nil {
			return nil, err
		}

		// Resolve the API to JSON data
		res, err := s.runtime.Resolve(ctx, &runtime.ResolveOptions{
			InstanceID:                instanceID,
			Resolver:                  api.Spec.Resolver,
			ResolverProperties:        api.Spec.ResolverProperties.AsMap(),
			Args:                      map[string]any{"sql": query, "priority": 1},
			UserAttributes:            nil,
			ResolveInteractiveOptions: &runtime.ResolverInteractiveOptions{Format: runtime.GOOBJECTS},
		})
		if err != nil {
			return nil, err
		}

		handle := func(ctx context.Context, writer wire.DataWriter, parameters []wire.Parameter) error {
			if len(res.Rows) == 0 {
				return writer.Empty()
			}
			for i := 0; i < len(res.Rows); i++ {
				if err := convert(res.Rows[i], res.Schema); err != nil {
					return fmt.Errorf("data conversion failed")
				}
				if err := writer.Row(res.Rows[i]); err != nil {
					s.logger.Warn("failed to write row", zap.Error(err))
					return err
				}
			}
			return writer.Complete("OK")
		}
		return wire.Prepared(wire.NewStatement(handle, wire.WithColumns(convertSchema(res.Schema)))), nil
	}
}

func convert(row []any, schema *runtimev1.StructType) error {
	for i := 0; i < len(row); i++ {
		code := schema.Fields[i].Type.Code
		switch code {
		case runtimev1.Type_CODE_INT128, runtimev1.Type_CODE_INT256, runtimev1.Type_CODE_UINT128, runtimev1.Type_CODE_UINT256:
			if v, ok := row[i].(*big.Int); ok {
				row[i] = decimal.NewFromBigInt(v, 0)
			}
		case runtimev1.Type_CODE_UINT64:
			if v, ok := row[i].(uint64); ok {
				row[i] = decimal.NewFromUint64(v)
			}
		case runtimev1.Type_CODE_DECIMAL:
			switch v := row[i].(type) {
			case duckdb.Decimal: // may be this duckdb specific handling should go in resolver ?
				row[i] = decimal.NewFromBigInt(v.Value, 0)
			case uint64:
				row[i] = decimal.NewFromUint64(v)
			}
		case runtimev1.Type_CODE_MAP:
			val, err := json.Marshal(row[i])
			if err != nil {
				return err
			}
			row[i] = string(val)
		case runtimev1.Type_CODE_STRUCT:
			val, err := json.Marshal(row[i])
			if err != nil {
				return err
			}
			row[i] = string(val)
		case runtimev1.Type_CODE_ARRAY:
			elemType := schema.Fields[i].Type.ArrayElementType
			if elemType != nil && (elemType.Code == runtimev1.Type_CODE_ARRAY || elemType.Code == runtimev1.Type_CODE_STRUCT || elemType.Code == runtimev1.Type_CODE_MAP) {
				val, err := json.Marshal(row[i])
				if err != nil {
					return err
				}
				row[i] = string(val)
			}
		}
	}
	return nil
}

func convertSchema(schema *runtimev1.StructType) wire.Columns {
	columns := make([]wire.Column, len(schema.Fields))
	for i, field := range schema.Fields {
		columns[i] = columnForType(field)
	}
	return columns
}

func columnForType(field *runtimev1.StructType_Field) wire.Column {
	col := wire.Column{
		Name: field.Name,
	}

	switch field.Type.Code {
	case runtimev1.Type_CODE_UNSPECIFIED:
		col.Oid = pgtype.UnknownOID
		col.Width = -1
	case runtimev1.Type_CODE_BOOL:
		col.Oid = pgtype.BoolOID
		col.Width = 1
	case runtimev1.Type_CODE_INT8, runtimev1.Type_CODE_INT16, runtimev1.Type_CODE_UINT8:
		col.Oid = pgtype.Int2OID
		col.Width = 2
	case runtimev1.Type_CODE_INT32, runtimev1.Type_CODE_UINT16:
		col.Oid = pgtype.Int4OID
		col.Width = 4
	case runtimev1.Type_CODE_INT64, runtimev1.Type_CODE_UINT32:
		col.Oid = pgtype.Int8OID
		col.Width = 8
	case runtimev1.Type_CODE_INT128, runtimev1.Type_CODE_INT256, runtimev1.Type_CODE_UINT64, runtimev1.Type_CODE_UINT128, runtimev1.Type_CODE_UINT256:
		col.Oid = pgtype.NumericOID
		col.Width = -1
	case runtimev1.Type_CODE_FLOAT32:
		col.Oid = pgtype.Float4OID
		col.Width = 4
	case runtimev1.Type_CODE_FLOAT64:
		col.Oid = pgtype.Float8OID
		col.Width = 8
	case runtimev1.Type_CODE_TIMESTAMP:
		col.Oid = pgtype.TimestampOID
		col.Width = 8
	case runtimev1.Type_CODE_DATE:
		col.Oid = pgtype.DateOID
		col.Width = 8
	case runtimev1.Type_CODE_TIME:
		col.Oid = pgtype.TimeOID
		col.Width = 8
	case runtimev1.Type_CODE_STRING:
		col.Oid = pgtype.VarcharOID
		col.Width = -1
	case runtimev1.Type_CODE_BYTES:
		col.Oid = pgtype.ByteaOID
		col.Width = -1
	case runtimev1.Type_CODE_DECIMAL:
		col.Oid = pgtype.NumericArrayOID
		col.Width = -1
	case runtimev1.Type_CODE_JSON:
		col.Oid = pgtype.JSONOID
		col.Width = -1
	case runtimev1.Type_CODE_UUID:
		col.Oid = pgtype.UUIDOID
		col.Width = 16
	case runtimev1.Type_CODE_ARRAY:
		return columnForArrayType(field)
	case runtimev1.Type_CODE_MAP:
		col.Oid = pgtype.JSONOID
		col.Width = -1
	case runtimev1.Type_CODE_STRUCT:
		col.Oid = pgtype.JSONOID
		col.Width = -1
	default:
		col.Oid = pgtype.UnknownOID
		col.Width = -1
	}
	return col
}

func columnForArrayType(field *runtimev1.StructType_Field) wire.Column {
	col := wire.Column{
		Name:  field.Name,
		Width: -1,
	}

	if field.Type.ArrayElementType == nil {
		col.Oid = pgtype.UnknownOID
	}
	switch field.Type.ArrayElementType.Code {
	case runtimev1.Type_CODE_BOOL:
		col.Oid = pgtype.BoolArrayOID
	case runtimev1.Type_CODE_INT8, runtimev1.Type_CODE_INT16, runtimev1.Type_CODE_UINT8:
		col.Oid = pgtype.Int2ArrayOID
	case runtimev1.Type_CODE_INT32, runtimev1.Type_CODE_UINT16:
		col.Oid = pgtype.Int4ArrayOID
	case runtimev1.Type_CODE_INT64, runtimev1.Type_CODE_UINT32:
		col.Oid = pgtype.Int8ArrayOID
	case runtimev1.Type_CODE_INT128, runtimev1.Type_CODE_INT256, runtimev1.Type_CODE_UINT64, runtimev1.Type_CODE_UINT128, runtimev1.Type_CODE_UINT256:
		col.Oid = pgtype.NumericArrayOID
	case runtimev1.Type_CODE_FLOAT32:
		col.Oid = pgtype.Float4ArrayOID
	case runtimev1.Type_CODE_FLOAT64:
		col.Oid = pgtype.Float8ArrayOID
	case runtimev1.Type_CODE_TIMESTAMP:
		col.Oid = pgtype.TimestampArrayOID
	case runtimev1.Type_CODE_DATE:
		col.Oid = pgtype.DateArrayOID
	case runtimev1.Type_CODE_TIME:
		col.Oid = pgtype.TimeArrayOID
	case runtimev1.Type_CODE_STRING:
		col.Oid = pgtype.VarcharArrayOID
	case runtimev1.Type_CODE_BYTES:
		col.Oid = pgtype.ByteaArrayOID
	case runtimev1.Type_CODE_DECIMAL:
		col.Oid = pgtype.NumericArrayOID
	case runtimev1.Type_CODE_JSON:
		col.Oid = pgtype.JSONArrayOID
	case runtimev1.Type_CODE_UUID:
		col.Oid = pgtype.UUIDArrayOID
	default:
		// todo handle this
		col.Oid = pgtype.UnknownOID
	}
	return col
}
