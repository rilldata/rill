package server

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	wire "github.com/jeroenrinzema/psql-wire"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/server/auth"
)

func QueryHandler(s *Server) func(ctx context.Context, query string) (wire.PreparedStatements, error) {
	return func(ctx context.Context, query string) (wire.PreparedStatements, error) {
		var schema wire.Columns
		handle := func(ctx context.Context, writer wire.DataWriter, parameters []wire.Parameter) error {
			clientParams := wire.ClientParameters(ctx)
			instanceID := clientParams[wire.ParamDatabase]
			// todo how to handle normal SQL ?
			api, err := s.runtime.APIForName(ctx, instanceID, "builtin_metrics_sql")
			if err != nil {
				return err
			}

			// Resolve the API to JSON data
			res, err := s.runtime.Resolve(ctx, &runtime.ResolveOptions{
				InstanceID:         instanceID,
				Resolver:           api.Spec.Resolver,
				ResolverProperties: map[string]any{"sql": query, "priority": 1},
				Args:               nil,
				UserAttributes:     auth.GetClaims(ctx).Attributes(),
				ResolverInteractiveOptions: &runtime.ResolverInteractiveOptions{Format: "OBJECTS"},
			})
			if err != nil {
				return err
			}

			schema = convertSchema(res.Schema)
			for _, row := range res.Rows {
				if err := writer.Row(row); err != nil {
					return err
				}
			}
			return writer.Complete("OK")
		}

		return wire.Prepared(wire.NewStatement(handle, wire.WithColumns(schema))), nil
	}
}

// func (p *PostgresHandler) handleStartup() error {
// 	startupMessage, err := p.backend.ReceiveStartupMessage()
// 	if err != nil {
// 		return fmt.Errorf("error receiving startup message: %w", err)
// 	}

// 	switch msg := startupMessage.(type) {
// 	case *pgproto3.PasswordMessage:
// 		_, err = auth.ParseClaims(ctx, p.server.aud, msg.Password)
// 		if err != nil {
// 			p.buf = (&pgproto3.ErrorResponse{}).Encode(p.buf)
// 			err = p.writeBuffer()
// 			if err != nil {
// 				return fmt.Errorf("error sending authentication error response: %w", err)
// 			}
// 			return p.handleStartup()
// 		}

// 		p.buf = (&pgproto3.AuthenticationOk{}).Encode(p.buf)
// 		p.buf = (&pgproto3.ReadyForQuery{TxStatus: 'I'}).Encode(p.buf)
// 		err = p.writeBuffer()
// 		if err != nil {
// 			return fmt.Errorf("error sending ready for query: %w", err)
// 		}
// 	case *pgproto3.StartupMessage:
// 		db, ok := msg.Parameters["database"]
// 		if !ok {
// 			return fmt.Errorf("database name not specified in startup message")
// 		}
// 		p.instanceID = db

// 		p.buf = (&pgproto3.AuthenticationCleartextPassword{}).Encode(p.buf)
// 		err = p.writeBuffer()
// 		if err != nil {
// 			return fmt.Errorf("error sending ready for query: %w", err)
// 		}
// 	case *pgproto3.SSLRequest:
// 		_, err = p.conn.Write([]byte("N"))
// 		if err != nil {
// 			return fmt.Errorf("error sending deny SSL request: %w", err)
// 		}
// 		return p.handleStartup()
// 	default:
// 		return fmt.Errorf("unknown startup message: %#v", startupMessage)
// 	}
// 	return nil
// }

func convertSchema(schema *runtimev1.StructType) wire.Columns {
	columns := make(wire.Columns, len(schema.Fields))
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
	default:
		col.Oid = pgtype.UnknownOID
		col.Width = -1
		// Type_CODE_ARRAY       Type_Code = 19
		// Type_CODE_STRUCT      Type_Code = 20
		// Type_CODE_MAP         Type_Code = 21
	}
	return col
}
