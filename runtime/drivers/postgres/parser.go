package postgres

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

type mapper interface {
	runtimeType() *runtimev1.Type
	value(pgxVal any) (any, error)
}

func register(oidToMapperMap map[string]mapper, typ string, m mapper) {
	oidToMapperMap[typ] = m
	// array of base type
	oidToMapperMap[fmt.Sprintf("_%s", typ)] = &arrayMapper{baseMapper: m}
}

// refer https://github.com/jackc/pgx/blob/master/pgtype/pgtype_default.go for base types
func getOidToMapperMap() map[string]mapper {
	m := make(map[string]mapper)
	register(m, "bit", &bitMapper{})
	register(m, "bool", &boolMapper{})
	register(m, "bpchar", &charMapper{})
	register(m, "bytea", &byteMapper{})
	register(m, "char", &charMapper{})
	register(m, "date", &dateMapper{})
	register(m, "float4", &float32Mapper{})
	register(m, "float8", &float64Mapper{})
	register(m, "int2", &int16Mapper{})
	register(m, "int4", &int32Mapper{})
	register(m, "int8", &int64Mapper{})
	register(m, "numeric", &numericMapper{})
	register(m, "text", &charMapper{})
	register(m, "time", &timeMapper{})
	register(m, "timestamp", &timeStampMapper{})
	register(m, "timestamptz", &timeStampMapper{})
	register(m, "uuid", &uuidMapper{})
	register(m, "varbit", &bitMapper{})
	register(m, "varchar", &charMapper{})
	register(m, "json", &jsonMapper{})
	register(m, "jsonb", &jsonMapper{})
	return m
}

type bitMapper struct{}

func (m *bitMapper) runtimeType() *runtimev1.Type {
	// use bitstring once appender supports it
	return &runtimev1.Type{Code: runtimev1.Type_CODE_STRING}
}

func (m *bitMapper) value(pgxVal any) (any, error) {
	switch b := pgxVal.(type) {
	case pgtype.Bits:
		str := strings.Builder{}
		for _, n := range b.Bytes {
			str.WriteString(fmt.Sprintf("%08b ", n))
		}
		return str.String()[:b.Len], nil
	default:
		return nil, fmt.Errorf("bitMapper: unsupported type %v", b)
	}
}

type boolMapper struct{}

func (m *boolMapper) runtimeType() *runtimev1.Type {
	return &runtimev1.Type{Code: runtimev1.Type_CODE_BOOL}
}

func (m *boolMapper) value(pgxVal any) (any, error) {
	switch b := pgxVal.(type) {
	case bool:
		return b, nil
	default:
		return nil, fmt.Errorf("boolMapper: unsupported type %v", b)
	}
}

type charMapper struct{}

func (m *charMapper) runtimeType() *runtimev1.Type {
	return &runtimev1.Type{Code: runtimev1.Type_CODE_STRING}
}

func (m *charMapper) value(pgxVal any) (any, error) {
	switch b := pgxVal.(type) {
	case string:
		return b, nil
	default:
		return nil, fmt.Errorf("charMapper: unsupported type %v", b)
	}
}

type byteMapper struct{}

func (m *byteMapper) runtimeType() *runtimev1.Type {
	return &runtimev1.Type{Code: runtimev1.Type_CODE_BYTES}
}

func (m *byteMapper) value(pgxVal any) (any, error) {
	switch b := pgxVal.(type) {
	case []byte:
		return b, nil
	default:
		return nil, fmt.Errorf("byteMapper: unsupported type %v", b)
	}
}

type dateMapper struct{}

func (m *dateMapper) runtimeType() *runtimev1.Type {
	// Use runtimev1.Type_CODE_DATE once DATE is supported by DuckDB appender
	return &runtimev1.Type{Code: runtimev1.Type_CODE_TIMESTAMP}
}

func (m *dateMapper) value(pgxVal any) (any, error) {
	switch b := pgxVal.(type) {
	case time.Time:
		return b, nil
	default:
		return nil, fmt.Errorf("dateMapper: unsupported type %v", b)
	}
}

type float32Mapper struct{}

func (m *float32Mapper) runtimeType() *runtimev1.Type {
	return &runtimev1.Type{Code: runtimev1.Type_CODE_FLOAT32}
}

func (m *float32Mapper) value(pgxVal any) (any, error) {
	switch b := pgxVal.(type) {
	case float32:
		return b, nil
	default:
		return nil, fmt.Errorf("float32Mapper: unsupported type %v", b)
	}
}

type float64Mapper struct{}

func (m *float64Mapper) runtimeType() *runtimev1.Type {
	return &runtimev1.Type{Code: runtimev1.Type_CODE_FLOAT64}
}

func (m *float64Mapper) value(pgxVal any) (any, error) {
	switch b := pgxVal.(type) {
	case float64:
		return b, nil
	default:
		return nil, fmt.Errorf("float64Mapper: unsupported type %v", b)
	}
}

type int16Mapper struct{}

func (m *int16Mapper) runtimeType() *runtimev1.Type {
	return &runtimev1.Type{Code: runtimev1.Type_CODE_INT16}
}

func (m *int16Mapper) value(pgxVal any) (any, error) {
	switch b := pgxVal.(type) {
	case int16:
		return b, nil
	default:
		return nil, fmt.Errorf("int16Mapper: unsupported type %v", b)
	}
}

type int32Mapper struct{}

func (m *int32Mapper) runtimeType() *runtimev1.Type {
	return &runtimev1.Type{Code: runtimev1.Type_CODE_INT32}
}

func (m *int32Mapper) value(pgxVal any) (any, error) {
	switch b := pgxVal.(type) {
	case int32:
		return b, nil
	default:
		return nil, fmt.Errorf("int32Mapper: unsupported type %v", b)
	}
}

type int64Mapper struct{}

func (m *int64Mapper) runtimeType() *runtimev1.Type {
	return &runtimev1.Type{Code: runtimev1.Type_CODE_INT64}
}

func (m *int64Mapper) value(pgxVal any) (any, error) {
	switch b := pgxVal.(type) {
	case int64:
		return b, nil
	default:
		return nil, fmt.Errorf("int64Mapper: unsupported type %v", b)
	}
}

type timeMapper struct{}

func (m *timeMapper) runtimeType() *runtimev1.Type {
	// Use runtimev1.Type_CODE_TIME once DATE is supported by DuckDB appender
	return &runtimev1.Type{Code: runtimev1.Type_CODE_TIMESTAMP}
}

func (m *timeMapper) value(pgxVal any) (any, error) {
	switch b := pgxVal.(type) {
	case pgtype.Time:
		midnight := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.UTC)
		duration := time.Duration(b.Microseconds) * time.Microsecond
		midnight = midnight.Add(duration)
		return midnight, nil
	default:
		return nil, fmt.Errorf("timeMapper: unsupported type %v", b)
	}
}

type timeStampMapper struct{}

func (m *timeStampMapper) runtimeType() *runtimev1.Type {
	return &runtimev1.Type{Code: runtimev1.Type_CODE_TIMESTAMP}
}

func (m *timeStampMapper) value(pgxVal any) (any, error) {
	switch b := pgxVal.(type) {
	case time.Time:
		return b, nil
	default:
		return nil, fmt.Errorf("timeStampMapper: unsupported type %v", b)
	}
}

type uuidMapper struct{}

func (m *uuidMapper) runtimeType() *runtimev1.Type {
	return &runtimev1.Type{Code: runtimev1.Type_CODE_UUID}
}

func (m *uuidMapper) value(pgxVal any) (any, error) {
	switch b := pgxVal.(type) {
	case [16]byte:
		id, err := uuid.FromBytes(b[:])
		if err != nil {
			return nil, err
		}
		return id.String(), nil
	default:
		return nil, fmt.Errorf("uuidMapper: unsupported type %v", b)
	}
}

type numericMapper struct{}

func (m *numericMapper) runtimeType() *runtimev1.Type {
	return &runtimev1.Type{Code: runtimev1.Type_CODE_STRING}
}

func (m *numericMapper) value(pgxVal any) (any, error) {
	switch b := pgxVal.(type) {
	case pgtype.NumericValuer:
		f, err := b.NumericValue()
		if err != nil {
			return nil, err
		}
		bytes, err := f.MarshalJSON()
		if err != nil {
			return nil, err
		}
		return string(bytes), nil
	case pgtype.Float64Valuer:
		f, err := b.Float64Value()
		if err != nil {
			return nil, err
		}
		return fmt.Sprint(f.Float64), nil
	case pgtype.Int64Valuer:
		f, err := b.Int64Value()
		if err != nil {
			return nil, err
		}
		return fmt.Sprint(f.Int64), nil
	default:
		return nil, fmt.Errorf("numericMapper: unsupported type %v", b)
	}
}

type jsonMapper struct{}

func (m *jsonMapper) runtimeType() *runtimev1.Type {
	return &runtimev1.Type{Code: runtimev1.Type_CODE_JSON}
}

func (m *jsonMapper) value(pgxVal any) (any, error) {
	switch b := pgxVal.(type) {
	case []byte:
		return string(b), nil
	case map[string]any:
		enc, err := json.Marshal(b)
		if err != nil {
			return nil, err
		}
		return string(enc), nil
	default:
		return nil, fmt.Errorf("jsonMapper: unsupported type %v", b)
	}
}

type arrayMapper struct {
	baseMapper mapper
}

func (m *arrayMapper) runtimeType() *runtimev1.Type {
	return &runtimev1.Type{Code: runtimev1.Type_CODE_JSON}
}

func (m *arrayMapper) value(pgxVal any) (any, error) {
	switch b := pgxVal.(type) {
	case []interface{}:
		arr := make([]any, len(b))
		for i, val := range b {
			res, err := m.baseMapper.value(val)
			if err != nil {
				return nil, err
			}
			arr[i] = res
		}
		enc, err := json.Marshal(arr)
		if err != nil {
			return nil, err
		}
		return string(enc), nil
	default:
		return nil, fmt.Errorf("arrayMapper: unsupported type %v", b)
	}
}
