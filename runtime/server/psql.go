package server

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	wire "github.com/jeroenrinzema/psql-wire"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/server/auth"
	"github.com/rilldata/rill/runtime/server/psql"
)

// psqlQueryHandler handles incoming queries on psql wire endpoint.
// The metadata queries are redirected to ResolvePSQLQuery.
// The metrics_sql queries are redirected to metrics_sql resolver.
func (s *Server) psqlQueryHandler(ctx context.Context, query string) (wire.PreparedStatements, error) {
	clientParams := wire.ClientParameters(ctx)
	instanceID := clientParams[wire.ParamDatabase]

	// The logic to identify metadata queries is somewhat hacky.
	// Identify metadata queries based on whether query contains names of common postgres catalogs like pg_attribute, pg_catalog etc.
	// Also consider queries that do not have a FROM clause as metadata queries since querying from metrics_view requires a FROM clause.
	if strings.Contains(query, "pg_attribute") || strings.Contains(query, "pg_catalog") || strings.Contains(query, "pg_type") || strings.Contains(query, "pg_namespace") || !strings.Contains(strings.ToUpper(query), "FROM") {
		data, schema, err := psql.ResolvePSQLQuery(ctx, &psql.PSQLQueryOpts{
			SQL:            query,
			Runtime:        s.runtime,
			InstanceID:     instanceID,
			UserAttributes: auth.GetClaims(ctx).Attributes(),
			Priority:       1,
			Logger:         s.logger,
		})
		if err != nil {
			return nil, err
		}

		handle := func(ctx context.Context, writer wire.DataWriter, parameters []wire.Parameter) error {
			for i := 0; i < len(data); i++ {
				if err := writer.Row(data[i]); err != nil {
					return fmt.Errorf("failed to write row: %w", err)
				}
			}
			return writer.Complete("OK")
		}
		return wire.Prepared(wire.NewStatement(handle, wire.WithColumns(convertSchema(schema)))), nil
	}

	// todo handle normal SQL. Some ideas to handle that
	// 1. Approach similar to BigQuery which has #legacySQL
	// 2. We can parse for a -- @connector: xxx comment, and if found, use that connector directly instead of metrics SQL
	res, err := s.runtime.Resolve(ctx, &runtime.ResolveOptions{
		InstanceID:         instanceID,
		Resolver:           "metrics_sql",
		ResolverProperties: map[string]any{"sql": query},
		Args:               map[string]any{"priority": 1},
		UserAttributes:     auth.GetClaims(ctx).Attributes(),
	})
	if err != nil {
		return nil, err
	}

	handle := func(ctx context.Context, writer wire.DataWriter, parameters []wire.Parameter) error {
		var rawData []map[string]interface{}
		err := json.Unmarshal(res.Data, &rawData)
		if err != nil {
			return err
		}

		row := make([]any, len(res.Schema.Fields))
		for i := 0; i < len(rawData); i++ {
			// get the values in the same order as schema
			for j := 0; j < len(res.Schema.Fields); j++ {
				if v, ok := rawData[i][res.Schema.Fields[j].Name]; ok {
					row[j] = v
				} else {
					row[j] = nil
				}
			}
			if err := writer.Row(row); err != nil {
				return fmt.Errorf("failed to write row: %w", err)
			}
		}
		return writer.Complete("OK")
	}
	return wire.Prepared(wire.NewStatement(handle, wire.WithColumns(convertSchema(res.Schema)))), nil
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
		col.Oid = pgtype.NumericOID
		col.Width = -1
	case runtimev1.Type_CODE_JSON:
		col.Oid = pgtype.JSONOID
		col.Width = -1
	case runtimev1.Type_CODE_UUID:
		col.Oid = pgtype.UUIDOID
		col.Width = 16
	case runtimev1.Type_CODE_ARRAY:
		// TODO : array will mostly fail for metrics_sql since in JSON they will be present as string but pgx protocol expects to be []type for array
		// metrics_sql will not output array right now.
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
		col.Oid = pgtype.UnknownOID
	}
	return col
}
