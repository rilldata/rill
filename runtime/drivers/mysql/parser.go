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
	// dest returns a pointer to a destination value that can be used in Rows.Scan
	dest(st reflect.Type) (any, error)
	// value dereferences a pointer created by dest
	value(p any) (any, error)
}

// refer https://github.com/go-sql-driver/mysql/blob/master/fields.go for base types
func getDBTypeNameToMapperMap() map[string]mapper {
	m := make(map[string]mapper)

	bit := bitMapper{}
	numeric := numericMapper{}
	char := charMapper{}
	bytes := byteMapper{}
	date := dateMapper{}
	json := jsonMapper{}

	// bit
	m["BIT"] = bit

	// numeric
	m["TINYINT"] = numeric
	m["SMALLINT"] = numeric
	m["MEDIUMINT"] = numeric
	m["INT"] = numeric
	m["UNSIGNED TINYINT"] = numeric
	m["UNSIGNED SMALLINT"] = numeric
	m["UNSIGNED INT"] = numeric
	m["UNSIGNED BIGINT"] = numeric
	m["BIGINT"] = numeric
	m["DOUBLE"] = numeric
	m["FLOAT"] = numeric
	// MySQL stores DECIMAL value in binary format
	// It might be stored as string without losing precision
	m["DECIMAL"] = char

	// string
	m["CHAR"] = char
	m["LONGTEXT"] = char
	m["MEDIUMTEXT"] = char
	m["TEXT"] = char
	m["TINYTEXT"] = char
	m["VARCHAR"] = char

	// binary
	m["BINARY"] = bytes
	m["TINYBLOB"] = bytes
	m["BLOB"] = bytes
	m["LONGBLOB"] = bytes
	m["MEDIUMBLOB"] = bytes
	m["VARBINARY"] = bytes

	// date and time
	m["DATE"] = date
	m["DATETIME"] = date
	m["TIMESTAMP"] = date
	m["YEAR"] = numeric
	// TIME is scanned as bytes and can be converted to string
	m["TIME"] = char

	// json
	m["JSON"] = json

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

func (m numericMapper) runtimeType(st reflect.Type) (*runtimev1.Type, error) {
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

func (m numericMapper) dest(st reflect.Type) (any, error) {
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

func (m numericMapper) value(p any) (any, error) {
	switch v := p.(type) {
	case *int8:
		return *v, nil
	case *int16:
		return *v, nil
	case *int32:
		return *v, nil
	case *int64:
		return *v, nil
	case *uint8:
		return *v, nil
	case *uint16:
		return *v, nil
	case *uint32:
		return *v, nil
	case *uint64:
		return *v, nil
	case *sql.NullInt64:
		vl, err := v.Value()
		if err != nil {
			return nil, err
		}
		return vl, nil
	case *float32:
		return *v, nil
	case *float64:
		return *v, nil
	case *sql.NullFloat64:
		vl, err := v.Value()
		if err != nil {
			return nil, err
		}
		return vl, nil
	default:
		return nil, fmt.Errorf("numericMapper: unsupported value type %v", p)
	}
}

type bitMapper struct{}

func (m bitMapper) runtimeType(reflect.Type) (*runtimev1.Type, error) {
	return &runtimev1.Type{Code: runtimev1.Type_CODE_STRING}, nil
}

func (m bitMapper) dest(reflect.Type) (any, error) {
	return &[]byte{}, nil
}

func (m bitMapper) value(p any) (any, error) {
	switch bs := p.(type) {
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
		return nil, fmt.Errorf("bitMapper: unsupported value type %v", bs)
	}
}

type charMapper struct{}

func (m charMapper) runtimeType(reflect.Type) (*runtimev1.Type, error) {
	return &runtimev1.Type{Code: runtimev1.Type_CODE_STRING}, nil
}

func (m charMapper) dest(reflect.Type) (any, error) {
	return new(sql.NullString), nil
}

func (m charMapper) value(p any) (any, error) {
	switch v := p.(type) {
	case *sql.NullString:
		vl, err := v.Value()
		if err != nil {
			return nil, err
		}
		return vl, nil
	default:
		return nil, fmt.Errorf("charMapper: unsupported value type %v", v)
	}
}

type byteMapper struct{}

func (m byteMapper) runtimeType(reflect.Type) (*runtimev1.Type, error) {
	return &runtimev1.Type{Code: runtimev1.Type_CODE_BYTES}, nil
}

func (m byteMapper) dest(reflect.Type) (any, error) {
	return &[]byte{}, nil
}

func (m byteMapper) value(p any) (any, error) {
	switch v := p.(type) {
	case *[]byte:
		if *v == nil {
			return nil, nil
		}
		return *v, nil
	default:
		return nil, fmt.Errorf("byteMapper: unsupported value type %v", v)
	}
}

type dateMapper struct{}

func (m dateMapper) runtimeType(reflect.Type) (*runtimev1.Type, error) {
	return &runtimev1.Type{Code: runtimev1.Type_CODE_DATE}, nil
}

func (m dateMapper) dest(reflect.Type) (any, error) {
	return new(sql.NullTime), nil
}

func (m dateMapper) value(p any) (any, error) {
	switch v := p.(type) {
	case *sql.NullTime:
		vl, err := v.Value()
		if err != nil {
			return nil, err
		}
		return vl, nil
	default:
		return nil, fmt.Errorf("dateMapper: unsupported value type %v", v)
	}
}

type jsonMapper struct{}

func (m jsonMapper) runtimeType(reflect.Type) (*runtimev1.Type, error) {
	return &runtimev1.Type{Code: runtimev1.Type_CODE_JSON}, nil
}

func (m jsonMapper) dest(reflect.Type) (any, error) {
	return new(sql.NullString), nil
}

func (m jsonMapper) value(p any) (any, error) {
	switch v := p.(type) {
	case *sql.NullString:
		vl, err := v.Value()
		if err != nil {
			return nil, err
		}
		return vl, nil
	default:
		return nil, fmt.Errorf("jsonMapper: unsupported value type %v", v)
	}
}
