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
	parameters := make([]base.Parameter, len(parameterOIDs))
	for i, oid := range parameterOIDs {
		parameters[i].OID = oid
	}
	rows, err := s.Query(ctx, query, parameters, nil)
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
	return true
}

func (r *resolverRows) Values() [][]byte   { return r.values }
func (r *resolverRows) Err() error         { return r.err }
func (r *resolverRows) CommandTag() string { return "" }
func (r *resolverRows) Close() error       { return r.result.Close() }

func fieldsForSchema(schema *runtimev1.StructType, resultFormats []int16) ([]pgproto3.FieldDescription, error) {
	if schema == nil {
		return nil, nil
	}
	fields := make([]pgproto3.FieldDescription, len(schema.Fields))
	for i, field := range schema.Fields {
		oid, size := postgresType(field.Type)
		fields[i] = pgproto3.FieldDescription{
			Name:         []byte(field.Name),
			DataTypeOID:  oid,
			DataTypeSize: size,
			TypeModifier: -1,
		}
	}
	return base.ApplyResultFormats(fields, resultFormats)
}

type postgresTypeInfo struct {
	oid      uint32
	arrayOID uint32
	size     int16
}

var postgresTypes = map[runtimev1.Type_Code]postgresTypeInfo{
	runtimev1.Type_CODE_BOOL:      {pgtype.BoolOID, pgtype.BoolArrayOID, 1},
	runtimev1.Type_CODE_INT8:      {pgtype.Int2OID, pgtype.Int2ArrayOID, 2},
	runtimev1.Type_CODE_INT16:     {pgtype.Int2OID, pgtype.Int2ArrayOID, 2},
	runtimev1.Type_CODE_UINT8:     {pgtype.Int2OID, pgtype.Int2ArrayOID, 2},
	runtimev1.Type_CODE_INT32:     {pgtype.Int4OID, pgtype.Int4ArrayOID, 4},
	runtimev1.Type_CODE_UINT16:    {pgtype.Int4OID, pgtype.Int4ArrayOID, 4},
	runtimev1.Type_CODE_INT64:     {pgtype.Int8OID, pgtype.Int8ArrayOID, 8},
	runtimev1.Type_CODE_UINT32:    {pgtype.Int8OID, pgtype.Int8ArrayOID, 8},
	runtimev1.Type_CODE_INT128:    {pgtype.NumericOID, pgtype.NumericArrayOID, -1},
	runtimev1.Type_CODE_INT256:    {pgtype.NumericOID, pgtype.NumericArrayOID, -1},
	runtimev1.Type_CODE_UINT64:    {pgtype.NumericOID, pgtype.NumericArrayOID, -1},
	runtimev1.Type_CODE_UINT128:   {pgtype.NumericOID, pgtype.NumericArrayOID, -1},
	runtimev1.Type_CODE_UINT256:   {pgtype.NumericOID, pgtype.NumericArrayOID, -1},
	runtimev1.Type_CODE_DECIMAL:   {pgtype.NumericOID, pgtype.NumericArrayOID, -1},
	runtimev1.Type_CODE_FLOAT32:   {pgtype.Float4OID, pgtype.Float4ArrayOID, 4},
	runtimev1.Type_CODE_FLOAT64:   {pgtype.Float8OID, pgtype.Float8ArrayOID, 8},
	runtimev1.Type_CODE_TIMESTAMP: {pgtype.TimestamptzOID, pgtype.TimestamptzArrayOID, 8},
	runtimev1.Type_CODE_DATE:      {pgtype.DateOID, pgtype.DateArrayOID, 4},
	runtimev1.Type_CODE_TIME:      {pgtype.TimeOID, pgtype.TimeArrayOID, 8},
	runtimev1.Type_CODE_BYTES:     {pgtype.ByteaOID, pgtype.ByteaArrayOID, -1},
	runtimev1.Type_CODE_JSON:      {pgtype.JSONBOID, pgtype.JSONBArrayOID, -1},
	runtimev1.Type_CODE_MAP:       {pgtype.JSONBOID, pgtype.JSONBArrayOID, -1},
	runtimev1.Type_CODE_STRUCT:    {pgtype.JSONBOID, pgtype.JSONBArrayOID, -1},
	runtimev1.Type_CODE_UUID:      {pgtype.UUIDOID, pgtype.UUIDArrayOID, 16},
}

func postgresType(typ *runtimev1.Type) (uint32, int16) {
	if typ == nil {
		return pgtype.TextOID, -1
	}
	if typ.Code == runtimev1.Type_CODE_ARRAY {
		if typ.ArrayElementType != nil {
			if info, ok := postgresTypes[typ.ArrayElementType.Code]; ok {
				return info.arrayOID, -1
			}
		}
		return pgtype.TextArrayOID, -1
	}
	if info, ok := postgresTypes[typ.Code]; ok {
		return info.oid, info.size
	}
	return pgtype.TextOID, -1
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

var (
	parameterPattern = regexp.MustCompile(`\$([1-9][0-9]*)(?:::([a-zA-Z0-9_]+(?:\s+precision|\s+with\s+time\s+zone)?))?`)
	parameterOIDs    = map[string]uint32{
		"bool":                     pgtype.BoolOID,
		"boolean":                  pgtype.BoolOID,
		"int2":                     pgtype.Int2OID,
		"smallint":                 pgtype.Int2OID,
		"int4":                     pgtype.Int4OID,
		"int":                      pgtype.Int4OID,
		"integer":                  pgtype.Int4OID,
		"int8":                     pgtype.Int8OID,
		"bigint":                   pgtype.Int8OID,
		"float4":                   pgtype.Float4OID,
		"real":                     pgtype.Float4OID,
		"float8":                   pgtype.Float8OID,
		"double precision":         pgtype.Float8OID,
		"numeric":                  pgtype.NumericOID,
		"decimal":                  pgtype.NumericOID,
		"date":                     pgtype.DateOID,
		"timestamp":                pgtype.TimestampOID,
		"timestamptz":              pgtype.TimestamptzOID,
		"timestamp with time zone": pgtype.TimestamptzOID,
		"uuid":                     pgtype.UUIDOID,
	}
)

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
		result[index-1] = parameterOIDs[strings.ToLower(strings.TrimSpace(match[2]))]
		if result[index-1] == 0 {
			result[index-1] = pgtype.TextOID
		}
	}
	return result
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
