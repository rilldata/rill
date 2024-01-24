package mysql

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

type mapper interface {
	runtimeType(st reflect.Type) (*runtimev1.Type, error)
	dest(st reflect.Type) (any, error)
}

// refer https://github.com/go-sql-driver/mysql/blob/master/fields.go for base types
func getDBTypeNameToMapperMap() map[string]mapper {
	m := make(map[string]mapper)

	// bit
	m["BIT"] = &bitMapper{}

	// numeric
	m["TINYINT"] = &numericMapper{}
	m["SMALLINT"] = &numericMapper{}
	m["MEDIUMINT"] = &numericMapper{}
	m["INT"] = &numericMapper{}
	m["UNSIGNED TINYINT"] = &numericMapper{}
	m["UNSIGNED SMALLINT"] = &numericMapper{}
	m["UNSIGNED INT"] = &numericMapper{}
	m["UNSIGNED BIGINT"] = &numericMapper{}
	m["BIGINT"] = &numericMapper{}
	m["DOUBLE"] = &numericMapper{}
	m["FLOAT"] = &numericMapper{}
	// MySQL stores DECIMAL value in binary format
	// It might be stored as string without losing precision
	m["DECIMAL"] = &charMapper{}

	// string
	m["CHAR"] = &charMapper{}
	m["LONGTEXT"] = &charMapper{}
	m["MEDIUMTEXT"] = &charMapper{}
	m["TEXT"] = &charMapper{}
	m["TINYTEXT"] = &charMapper{}
	m["VARCHAR"] = &charMapper{}

	// binary
	m[("BINARY")] = &byteMapper{}
	m[("TINYBLOB")] = &byteMapper{}
	m[("BLOB")] = &byteMapper{}
	m[("LONGBLOB")] = &byteMapper{}
	m[("MEDIUMBLOB")] = &byteMapper{}
	m[("VARBINARY")] = &byteMapper{}

	// date and time
	m[("DATE")] = &dateMapper{}
	m[("DATETIME")] = &dateMapper{}
	m[("TIMESTAMP")] = &dateMapper{}
	m[("YEAR")] = &int16Mapper{}
	// TIME is scanned as bytes and can be converted to string
	m[("TIME")] = &charMapper{}

	// json
	m[("JSON")] = &jsonMapper{}

	return m
}

var (
	scanTypeFloat32   = reflect.TypeOf(float32(0))
	scanTypeFloat64   = reflect.TypeOf(float64(0))
	scanTypeInt8      = reflect.TypeOf(int8(0))
	scanTypeInt16     = reflect.TypeOf(int16(0))
	scanTypeInt32     = reflect.TypeOf(int32(0))
	scanTypeInt64     = reflect.TypeOf(int64(0))
	scanTypeNullFloat = reflect.TypeOf(sql.NullFloat64{})
	scanTypeNullInt   = reflect.TypeOf(sql.NullInt64{})
	scanTypeUint8     = reflect.TypeOf(uint8(0))
	scanTypeUint16    = reflect.TypeOf(uint16(0))
	scanTypeUint32    = reflect.TypeOf(uint32(0))
	scanTypeUint64    = reflect.TypeOf(uint64(0))
)

type numericMapper struct{}

func (c *numericMapper) runtimeType(st reflect.Type) (*runtimev1.Type, error) {
	switch st {
	case scanTypeInt8:
		return &runtimev1.Type{Code: runtimev1.Type_CODE_INT8}, nil
	case scanTypeInt16:
		return &runtimev1.Type{Code: runtimev1.Type_CODE_INT16}, nil
	case scanTypeInt32:
		return &runtimev1.Type{Code: runtimev1.Type_CODE_INT32}, nil
	case scanTypeInt64:
		return &runtimev1.Type{Code: runtimev1.Type_CODE_INT64}, nil
	case scanTypeUint8:
		return &runtimev1.Type{Code: runtimev1.Type_CODE_UINT8}, nil
	case scanTypeUint16:
		return &runtimev1.Type{Code: runtimev1.Type_CODE_UINT16}, nil
	case scanTypeUint32:
		return &runtimev1.Type{Code: runtimev1.Type_CODE_UINT32}, nil
	case scanTypeUint64:
		return &runtimev1.Type{Code: runtimev1.Type_CODE_UINT64}, nil
	case scanTypeNullInt:
		return &runtimev1.Type{Code: runtimev1.Type_CODE_INT64}, nil
	case scanTypeFloat32:
		return &runtimev1.Type{Code: runtimev1.Type_CODE_FLOAT32}, nil
	case scanTypeFloat64:
		return &runtimev1.Type{Code: runtimev1.Type_CODE_FLOAT64}, nil
	case scanTypeNullFloat:
		return &runtimev1.Type{Code: runtimev1.Type_CODE_FLOAT64}, nil
	default:
		return nil, fmt.Errorf("numericMapper: unsupported scan type %v", st.Name())
	}
}

func (c *numericMapper) dest(st reflect.Type) (any, error) {
	switch st {
	case scanTypeInt8:
		return new(int8), nil
	case scanTypeInt16:
		return new(int16), nil
	case scanTypeInt32:
		return new(int32), nil
	case scanTypeInt64:
		return new(int64), nil
	case scanTypeUint8:
		return new(uint8), nil
	case scanTypeUint16:
		return new(uint16), nil
	case scanTypeUint32:
		return new(uint32), nil
	case scanTypeUint64:
		return new(uint64), nil
	case scanTypeNullInt:
		return new(sql.NullInt64), nil
	case scanTypeFloat32:
		return new(float32), nil
	case scanTypeFloat64:
		return new(float64), nil
	case scanTypeNullFloat:
		return new(sql.NullFloat64), nil
	default:
		return nil, fmt.Errorf("numericMapper: unsupported scan type %v", st.Name())
	}
}

type bitMapper struct{}

func (m *bitMapper) runtimeType(reflect.Type) (*runtimev1.Type, error) {
	return &runtimev1.Type{Code: runtimev1.Type_CODE_STRING}, nil
}

func (m *bitMapper) dest(reflect.Type) (any, error) {
	return &[]byte{}, nil
}

func (m *bitMapper) value(v any) (any, error) {
	switch bs := v.(type) {
	case *[]byte:
		if *bs == nil {
			return nil, nil
		}
		str := strings.Builder{}
		for _, b := range *bs {
			str.WriteString(fmt.Sprintf("%08b ", b))
		}
		s := str.String()[:len(*bs)]
		return s, nil
	default:
		return nil, fmt.Errorf("bitMapper: unsupported type %v", bs)
	}
}

type charMapper struct{}

func (m *charMapper) runtimeType(reflect.Type) (*runtimev1.Type, error) {
	return &runtimev1.Type{Code: runtimev1.Type_CODE_STRING}, nil
}

func (m *charMapper) dest(reflect.Type) (any, error) {
	return new(sql.NullString), nil
}

type byteMapper struct{}

func (m *byteMapper) runtimeType(reflect.Type) (*runtimev1.Type, error) {
	return &runtimev1.Type{Code: runtimev1.Type_CODE_BYTES}, nil
}

func (m *byteMapper) dest(reflect.Type) (any, error) {
	return &[]byte{}, nil
}

type dateMapper struct{}

func (m *dateMapper) runtimeType(reflect.Type) (*runtimev1.Type, error) {
	return &runtimev1.Type{Code: runtimev1.Type_CODE_DATE}, nil
}

func (m *dateMapper) dest(reflect.Type) (any, error) {
	return new(sql.NullTime), nil
}

type int16Mapper struct{}

func (m *int16Mapper) runtimeType(reflect.Type) (*runtimev1.Type, error) {
	return &runtimev1.Type{Code: runtimev1.Type_CODE_INT16}, nil
}

func (m *int16Mapper) dest(reflect.Type) (any, error) {
	return new(sql.NullInt16), nil
}

type jsonMapper struct{}

func (m *jsonMapper) runtimeType(reflect.Type) (*runtimev1.Type, error) {
	return &runtimev1.Type{Code: runtimev1.Type_CODE_JSON}, nil
}

func (m *jsonMapper) dest(reflect.Type) (any, error) {
	return new(sql.NullString), nil
}
