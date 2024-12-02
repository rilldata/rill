package trino

import (
	"database/sql"
	"fmt"
	"reflect"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

type mapper interface {
	runtimeType(st reflect.Type) (*runtimev1.Type, error)
	// dest returns a pointer to a destination value that can be used in Rows.Scan
	dest(st reflect.Type) (any, error)
	// value dereferences a pointer created by dest
	value(p any) (any, error)
}

// refer https://trino.io/docs/current/language/types.html# for base types some complex date type like
// Map, Array, Row, HyperLogLog, P4HyperLogLog, SetDigest, QDigest and TDigest are not supported
func getDBTypeNameToMapperMap() map[string]mapper {
	m := make(map[string]mapper)

	boolean := booleanMapper{}
	numeric := numericMapper{}
	char := charMapper{}
	bytes := byteMapper{}
	date := dateMapper{}
	json := jsonMapper{}

	// boolean
	m["BOOLEAN"] = boolean

	// numeric
	m["TINYINT"] = numeric
	m["SMALLINT"] = numeric
	m["INTEGER"] = numeric
	m["BIGINT"] = numeric

	m["REAL"] = numeric
	m["DOUBLE"] = numeric
	// Trino DECIMAL value have Precision up to 38 digits
	// It might be stored as string without losing precision
	m["DECIMAL"] = char

	// string
	m["CHAR"] = char
	m["VARCHAR"] = char

	// date and time
	m["DATE"] = date

	m["TIMESTAMP"] = date
	m["TIMESTAMP WITH TIME ZONE"] = date

	// TIME is scanned as bytes and can be converted to string
	m["TIME"] = date
	m["TIME WITH TIME ZONE"] = date

	m["INTERVAL YEAR TO MONTH"] = char
	m["INTERVAL DAY TO SECOND"] = char

	// binary
	m["VARBINARY"] = bytes

	// json
	m["JSON"] = json

	m["IPADDRESS"] = char

	m["UUID"] = char

	return m
}

var (
	scanTypeNullInt16   = reflect.TypeOf(sql.NullInt16{})
	scanTypeNullInt32   = reflect.TypeOf(sql.NullInt32{})
	scanTypeNullInt64   = reflect.TypeOf(sql.NullInt64{})
	scanTypeNullFloat64 = reflect.TypeOf(sql.NullFloat64{})
)

type numericMapper struct{}

func (m numericMapper) runtimeType(st reflect.Type) (*runtimev1.Type, error) {
	switch st {
	case scanTypeNullInt16:
		return &runtimev1.Type{Code: runtimev1.Type_CODE_INT16}, nil
	case scanTypeNullInt32:
		return &runtimev1.Type{Code: runtimev1.Type_CODE_INT32}, nil
	case scanTypeNullInt64:
		return &runtimev1.Type{Code: runtimev1.Type_CODE_INT64}, nil
	case scanTypeNullFloat64:
		return &runtimev1.Type{Code: runtimev1.Type_CODE_FLOAT64}, nil
	default:
		return nil, fmt.Errorf("numericMapper: unsupported scan type %v", st.Name())
	}
}

func (m numericMapper) dest(st reflect.Type) (any, error) {
	switch st {
	case scanTypeNullInt16:
		return new(sql.NullInt16), nil
	case scanTypeNullInt32:
		return new(sql.NullInt32), nil
	case scanTypeNullInt64:
		return new(sql.NullInt64), nil
	case scanTypeNullFloat64:
		return new(sql.NullFloat64), nil
	default:
		return nil, fmt.Errorf("numericMapper: unsupported scan type %v", st.Name())
	}
}

func (m numericMapper) value(p any) (any, error) {
	switch v := p.(type) {
	case *sql.NullInt16:
		vl, err := v.Value()
		if err != nil {
			return nil, err
		}
		return vl, nil
	case *sql.NullInt32:
		vl, err := v.Value()
		if err != nil {
			return nil, err
		}
		return vl, nil
	case *sql.NullInt64:
		vl, err := v.Value()
		if err != nil {
			return nil, err
		}
		return vl, nil
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

type booleanMapper struct{}

func (m booleanMapper) runtimeType(reflect.Type) (*runtimev1.Type, error) {
	return &runtimev1.Type{Code: runtimev1.Type_CODE_BOOL}, nil
}

func (m booleanMapper) dest(reflect.Type) (any, error) {
	return new(sql.NullBool), nil
}

func (m booleanMapper) value(p any) (any, error) {
	// Interpret the value from the destination
	switch b := p.(type) {
	case *sql.NullBool:
		vl, err := b.Value()
		if err != nil {
			return nil, err
		}
		return vl, nil
	default:
		return nil, fmt.Errorf("boolMapper: unsupported value type %v", b)
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
	return &runtimev1.Type{Code: runtimev1.Type_CODE_TIMESTAMP}, nil
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
