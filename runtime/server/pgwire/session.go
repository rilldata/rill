// Package pgwire adapts runtime queries to the PostgreSQL wire protocol.
package pgwire

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/jackc/pgx/v5/pgtype"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	base "github.com/rilldata/rill/runtime/pkg/pgwire"
)

// Session executes Metrics SQL for one runtime instance.
type Session struct {
	runtime    *runtime.Runtime
	instanceID string
	claims     *runtime.SecurityClaims
	types      *pgtype.Map
}

// NewSession creates a runtime pgwire session.
func NewSession(rt *runtime.Runtime, instanceID string, claims *runtime.SecurityClaims) (*Session, error) {
	if instanceID == "" {
		return nil, &base.Error{Code: "3D000", Message: "database must be a runtime instance ID"}
	}
	if claims == nil || !claims.Can(runtime.ReadMetrics) {
		return nil, &base.Error{Code: "42501", Message: "permission denied for runtime instance"}
	}
	return &Session{runtime: rt, instanceID: instanceID, claims: claims, types: pgtype.NewMap()}, nil
}

// Close implements base.Session.
func (s *Session) Close() error { return nil }

// Describe implements base.Session.
func (s *Session) Describe(ctx context.Context, query string, parameterOIDs []uint32) (*base.Description, error) {
	parameterOIDs = inferParameterOIDs(query, parameterOIDs)
	// Resolving here is currently necessary because resolver schemas are produced
	// by the underlying OLAP query. The result is closed without consuming rows.
	rows, err := s.Query(ctx, query, nilParameters(parameterOIDs), nil)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return &base.Description{ParameterOIDs: append([]uint32(nil), parameterOIDs...), Fields: rows.Fields()}, nil
}

// Query implements base.Session.
func (s *Session) Query(ctx context.Context, query string, parameters []base.Parameter, resultFormats []int16) (base.Rows, error) {
	var err error
	if len(parameters) != 0 {
		query, err = interpolateParameters(query, parameters, s.types)
		if err != nil {
			return nil, err
		}
	}

	if isCatalogQuery(query) {
		return queryCatalog(ctx, s.runtime, s.instanceID, s.claims, query, resultFormats, s.types)
	}

	res, _, err := s.runtime.Resolve(ctx, &runtime.ResolveOptions{
		InstanceID:         s.instanceID,
		Resolver:           "metrics_sql",
		ResolverProperties: map[string]any{"sql": query},
		Args:               map[string]any{"priority": 1},
		Claims:             s.claims,
	})
	if err != nil {
		return nil, err
	}

	fields, err := fieldsForSchema(res.Schema(), resultFormats)
	if err != nil {
		_ = res.Close()
		return nil, err
	}
	return &resolverRows{result: res, fields: fields, types: s.types}, nil
}

type resolverRows struct {
	result runtime.ResolverResult
	fields []pgproto3.FieldDescription
	types  *pgtype.Map
	values [][]byte
	err    error
	count  int64
}

func (r *resolverRows) Fields() []pgproto3.FieldDescription { return r.fields }

func (r *resolverRows) Next() bool {
	row, err := r.result.Next()
	if err != nil {
		if !errors.Is(err, io.EOF) {
			r.err = err
		}
		return false
	}

	r.values = make([][]byte, len(r.fields))
	for i, field := range r.fields {
		value := row[string(field.Name)]
		value, err = normalizeValue(value, field.DataTypeOID)
		if err != nil {
			r.err = err
			return false
		}
		r.values[i], err = encodeValue(r.types, field.DataTypeOID, field.Format, value)
		if err != nil {
			r.err = fmt.Errorf("failed to encode column %q: %w", field.Name, err)
			return false
		}
	}
	r.count++
	return true
}

func (r *resolverRows) Values() [][]byte { return r.values }
func (r *resolverRows) Err() error       { return r.err }
func (r *resolverRows) CommandTag() string {
	return fmt.Sprintf("SELECT %d", r.count)
}
func (r *resolverRows) Close() error { return r.result.Close() }

func fieldsForSchema(schema *runtimev1.StructType, resultFormats []int16) ([]pgproto3.FieldDescription, error) {
	if schema == nil {
		return nil, nil
	}
	fields := make([]pgproto3.FieldDescription, len(schema.Fields))
	for i, field := range schema.Fields {
		oid, size := postgresType(field.Type)
		format, err := resultFormat(resultFormats, i, len(schema.Fields))
		if err != nil {
			return nil, err
		}
		fields[i] = pgproto3.FieldDescription{
			Name:                 []byte(field.Name),
			DataTypeOID:          oid,
			DataTypeSize:         size,
			TypeModifier:         -1,
			Format:               format,
			TableAttributeNumber: 0,
		}
	}
	return fields, nil
}

func postgresType(typ *runtimev1.Type) (uint32, int16) {
	if typ == nil {
		return pgtype.TextOID, -1
	}
	switch typ.Code {
	case runtimev1.Type_CODE_BOOL:
		return pgtype.BoolOID, 1
	case runtimev1.Type_CODE_INT8, runtimev1.Type_CODE_INT16, runtimev1.Type_CODE_UINT8:
		return pgtype.Int2OID, 2
	case runtimev1.Type_CODE_INT32, runtimev1.Type_CODE_UINT16:
		return pgtype.Int4OID, 4
	case runtimev1.Type_CODE_INT64, runtimev1.Type_CODE_UINT32:
		return pgtype.Int8OID, 8
	case runtimev1.Type_CODE_INT128, runtimev1.Type_CODE_INT256, runtimev1.Type_CODE_UINT64, runtimev1.Type_CODE_UINT128, runtimev1.Type_CODE_UINT256, runtimev1.Type_CODE_DECIMAL:
		return pgtype.NumericOID, -1
	case runtimev1.Type_CODE_FLOAT32:
		return pgtype.Float4OID, 4
	case runtimev1.Type_CODE_FLOAT64:
		return pgtype.Float8OID, 8
	case runtimev1.Type_CODE_TIMESTAMP:
		return pgtype.TimestamptzOID, 8
	case runtimev1.Type_CODE_DATE:
		return pgtype.DateOID, 4
	case runtimev1.Type_CODE_TIME:
		return pgtype.TimeOID, 8
	case runtimev1.Type_CODE_BYTES:
		return pgtype.ByteaOID, -1
	case runtimev1.Type_CODE_JSON, runtimev1.Type_CODE_MAP, runtimev1.Type_CODE_STRUCT:
		return pgtype.JSONBOID, -1
	case runtimev1.Type_CODE_UUID:
		return pgtype.UUIDOID, 16
	case runtimev1.Type_CODE_ARRAY:
		return postgresArrayType(typ.ArrayElementType), -1
	default:
		return pgtype.TextOID, -1
	}
}

func postgresArrayType(typ *runtimev1.Type) uint32 {
	if typ == nil {
		return pgtype.TextArrayOID
	}
	switch typ.Code {
	case runtimev1.Type_CODE_BOOL:
		return pgtype.BoolArrayOID
	case runtimev1.Type_CODE_INT8, runtimev1.Type_CODE_INT16, runtimev1.Type_CODE_UINT8:
		return pgtype.Int2ArrayOID
	case runtimev1.Type_CODE_INT32, runtimev1.Type_CODE_UINT16:
		return pgtype.Int4ArrayOID
	case runtimev1.Type_CODE_INT64, runtimev1.Type_CODE_UINT32:
		return pgtype.Int8ArrayOID
	case runtimev1.Type_CODE_FLOAT32:
		return pgtype.Float4ArrayOID
	case runtimev1.Type_CODE_FLOAT64:
		return pgtype.Float8ArrayOID
	case runtimev1.Type_CODE_TIMESTAMP:
		return pgtype.TimestamptzArrayOID
	case runtimev1.Type_CODE_DATE:
		return pgtype.DateArrayOID
	case runtimev1.Type_CODE_TIME:
		return pgtype.TimeArrayOID
	case runtimev1.Type_CODE_BYTES:
		return pgtype.ByteaArrayOID
	case runtimev1.Type_CODE_JSON, runtimev1.Type_CODE_MAP, runtimev1.Type_CODE_STRUCT:
		return pgtype.JSONBArrayOID
	case runtimev1.Type_CODE_UUID:
		return pgtype.UUIDArrayOID
	case runtimev1.Type_CODE_DECIMAL, runtimev1.Type_CODE_INT128, runtimev1.Type_CODE_INT256, runtimev1.Type_CODE_UINT64, runtimev1.Type_CODE_UINT128, runtimev1.Type_CODE_UINT256:
		return pgtype.NumericArrayOID
	default:
		return pgtype.TextArrayOID
	}
}

func resultFormat(formats []int16, index, fieldCount int) (int16, error) {
	var format int16
	switch len(formats) {
	case 0:
		format = 0
	case 1:
		format = formats[0]
	default:
		if len(formats) != fieldCount {
			return 0, &base.Error{Code: "08P01", Message: fmt.Sprintf("received %d result formats for %d columns", len(formats), fieldCount)}
		}
		format = formats[index]
	}
	if format != 0 && format != 1 {
		return 0, &base.Error{Code: "08P01", Message: fmt.Sprintf("unsupported result format %d", format)}
	}
	return format, nil
}

func encodeValue(types *pgtype.Map, oid uint32, format int16, value any) ([]byte, error) {
	// pgtype emits UTC timestamptz values with a trailing "Z" in text format.
	// PostgreSQL emits a numeric offset, which is required for psycopg2 to parse
	// the value reliably (a trailing "Z" can retain seconds from its input buffer).
	if oid == pgtype.TimestamptzOID && format == pgtype.TextFormatCode {
		if value, ok := value.(time.Time); ok {
			return []byte(value.Format("2006-01-02 15:04:05.999999999-07:00")), nil
		}
	}
	return types.Encode(oid, format, value, nil)
}

func normalizeValue(value any, oid uint32) (any, error) {
	if value == nil {
		return nil, nil
	}
	switch oid {
	case pgtype.TimestamptzOID, pgtype.TimestampOID:
		if value, ok := value.(string); ok {
			return time.Parse(time.RFC3339Nano, value)
		}
	case pgtype.DateOID:
		if value, ok := value.(string); ok {
			return time.Parse(time.DateOnly, value)
		}
	case pgtype.TimeOID:
		if value, ok := value.(string); ok {
			parsed := &pgtype.Time{}
			if err := parsed.Scan(value); err != nil {
				return nil, err
			}
			return parsed, nil
		}
	case pgtype.NumericOID:
		switch value := value.(type) {
		case string:
			parsed := &pgtype.Numeric{}
			if err := parsed.Scan(value); err != nil {
				return nil, err
			}
			return parsed, nil
		case *big.Int:
			parsed := &pgtype.Numeric{}
			if err := parsed.Scan(value.String()); err != nil {
				return nil, err
			}
			return parsed, nil
		}
	case pgtype.JSONOID, pgtype.JSONBOID:
		switch value.(type) {
		case string, []byte:
			return value, nil
		default:
			return json.Marshal(value)
		}
	}
	return value, nil
}

var parameterPattern = regexp.MustCompile(`\$([1-9][0-9]*)(?:::([a-zA-Z0-9_]+(?:\s+precision|\s+with\s+time\s+zone)?))?`)

func inferParameterOIDs(query string, provided []uint32) []uint32 {
	result := append([]uint32(nil), provided...)
	for _, match := range parameterPattern.FindAllStringSubmatch(query, -1) {
		index, _ := strconv.Atoi(match[1])
		for len(result) < index {
			result = append(result, 0)
		}
		if result[index-1] != 0 {
			continue
		}
		switch strings.ToLower(strings.TrimSpace(match[2])) {
		case "bool", "boolean":
			result[index-1] = pgtype.BoolOID
		case "int2", "smallint":
			result[index-1] = pgtype.Int2OID
		case "int4", "int", "integer":
			result[index-1] = pgtype.Int4OID
		case "int8", "bigint":
			result[index-1] = pgtype.Int8OID
		case "float4", "real":
			result[index-1] = pgtype.Float4OID
		case "float8", "double precision":
			result[index-1] = pgtype.Float8OID
		case "numeric", "decimal":
			result[index-1] = pgtype.NumericOID
		case "date":
			result[index-1] = pgtype.DateOID
		case "timestamp":
			result[index-1] = pgtype.TimestampOID
		case "timestamptz", "timestamp with time zone":
			result[index-1] = pgtype.TimestamptzOID
		case "uuid":
			result[index-1] = pgtype.UUIDOID
		default:
			result[index-1] = pgtype.TextOID
		}
	}
	return result
}

func nilParameters(oids []uint32) []base.Parameter {
	parameters := make([]base.Parameter, len(oids))
	for i, oid := range oids {
		parameters[i] = base.Parameter{OID: oid}
	}
	return parameters
}

func interpolateParameters(query string, parameters []base.Parameter, types *pgtype.Map) (string, error) {
	literals := make([]string, len(parameters))
	for i, parameter := range parameters {
		literal, err := parameterLiteral(parameter, types)
		if err != nil {
			return "", fmt.Errorf("invalid parameter $%d: %w", i+1, err)
		}
		literals[i] = literal
	}

	var out strings.Builder
	for i := 0; i < len(query); {
		switch query[i] {
		case '\'':
			start := i
			i++
			for i < len(query) {
				if query[i] == '\'' {
					i++
					if i < len(query) && query[i] == '\'' {
						i++
						continue
					}
					break
				}
				i++
			}
			out.WriteString(query[start:i])
		case '"':
			start := i
			i++
			for i < len(query) {
				if query[i] == '"' {
					i++
					if i < len(query) && query[i] == '"' {
						i++
						continue
					}
					break
				}
				i++
			}
			out.WriteString(query[start:i])
		case '$':
			j := i + 1
			for j < len(query) && query[j] >= '0' && query[j] <= '9' {
				j++
			}
			if j == i+1 {
				out.WriteByte(query[i])
				i++
				continue
			}
			index, _ := strconv.Atoi(query[i+1 : j])
			if index == 0 || index > len(literals) {
				return "", &base.Error{Code: "42P02", Message: fmt.Sprintf("there is no parameter $%d", index)}
			}
			out.WriteString(literals[index-1])
			i = j
		default:
			out.WriteByte(query[i])
			i++
		}
	}
	return out.String(), nil
}

func parameterLiteral(parameter base.Parameter, types *pgtype.Map) (string, error) {
	if parameter.Value == nil {
		return "NULL", nil
	}
	if parameter.OID == 0 {
		if parameter.Format != 0 {
			return "", fmt.Errorf("binary parameter has no type OID")
		}
		return quoteLiteral(string(parameter.Value)), nil
	}

	var value any
	if err := types.Scan(parameter.OID, parameter.Format, parameter.Value, &value); err != nil {
		return "", err
	}
	text, err := types.Encode(parameter.OID, pgtype.TextFormatCode, value, nil)
	if err != nil {
		return "", err
	}
	switch parameter.OID {
	case pgtype.BoolOID, pgtype.Int2OID, pgtype.Int4OID, pgtype.Int8OID, pgtype.Float4OID, pgtype.Float8OID, pgtype.NumericOID:
		return string(text), nil
	case pgtype.ByteaOID:
		return "decode('" + hex.EncodeToString(value.([]byte)) + "', 'hex')", nil
	default:
		typeName := "text"
		if typ, ok := types.TypeForOID(parameter.OID); ok {
			typeName = typ.Name
		}
		return quoteLiteral(string(text)) + "::" + typeName, nil
	}
}

func quoteLiteral(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}
