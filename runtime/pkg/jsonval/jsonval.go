package jsonval

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"net"
	"reflect"
	"strings"
	"time"

	"github.com/duckdb/duckdb-go/v2"
	"github.com/google/uuid"
	"github.com/paulmach/orb"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers/clickhouse"
)

// ToValue converts a value scanned from a database/sql driver to a Go type that can be marshaled to JSON.
// If v is a complex type, it may be mutated in-place.
func ToValue(v any, t *runtimev1.Type) (any, error) {
	switch v := v.(type) {
	case nil:
		return nil, nil
	case int:
		return int64(v), nil
	case int8:
		return int64(v), nil
	case int16:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case uint:
		return uint64(v), nil
	case uint8:
		return uint64(v), nil
	case uint16:
		return uint64(v), nil
	case uint32:
		return uint64(v), nil
	case float32:
		// Turning NaNs and Infs into nulls until frontend can deal with them as strings
		// (They don't have a native JSON representation)
		if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
			return nil, nil
		}
		return float64(v), nil
	case float64:
		// Turning NaNs and Infs into nulls until frontend can deal with them as strings
		// (They don't have a native JSON representation)
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return nil, nil
		}
		return v, nil
	case time.Time:
		if t != nil && t.Code == runtimev1.Type_CODE_DATE {
			return v.In(time.UTC).Format(time.DateOnly), nil
		}
		return v, nil
	case string:
		if t != nil {
			switch t.Code {
			case runtimev1.Type_CODE_DECIMAL:
				// Evil cast to float until frontend can deal with bigs:
				v2, ok := new(big.Float).SetString(v)
				if ok {
					f, _ := v2.Float64()
					return f, nil
				}
			case runtimev1.Type_CODE_INTERVAL:
				// ClickHouse currently returns INTERVALs as strings.
				// Our current policy is to convert INTERVALs to milliseconds, treating one month as 30 days.
				v2, ok := clickhouse.ParseIntervalToMillis(v)
				if ok {
					return v2, nil
				}
			}
		}
		return strings.ToValidUTF8(v, "�"), nil
	case []byte:
		if t != nil && t.Code == runtimev1.Type_CODE_UUID {
			uid, err := uuid.FromBytes(v)
			if err == nil {
				return uid.String(), nil
			}
		}
		// Falling through for default handling
	case big.Int:
		// Evil cast to float until frontend can deal with bigs:
		v2, _ := new(big.Float).SetInt(&v).Float64()
		return v2, nil
		// This is what we should do when frontend supports it:
		// return v.String(), nil
	case *big.Int:
		// Evil cast to float until frontend can deal with bigs:
		v2, _ := new(big.Float).SetInt(v).Float64()
		return v2, nil
		// This is what we should do when frontend supports it:
		// return v.String(), nil
	case duckdb.Decimal:
		// Evil cast to float until frontend can deal with bigs:
		denom := big.NewInt(10)
		denom.Exp(denom, big.NewInt(int64(v.Scale)), nil)
		v2, _ := new(big.Rat).SetFrac(v.Value, denom).Float64()
		return v2, nil
	case map[string]any:
		var t2 *runtimev1.StructType
		if t != nil {
			t2 = t.StructType
		}
		return toMap(v, t2)
	case []any:
		return toSlice(v, t)
	case map[any]any:
		var t2 *runtimev1.MapType
		if t != nil {
			t2 = t.MapType
		}
		return toMapCoerceKeys(v, t2)
	case duckdb.Map:
		return ToValue(map[any]any(v), t)
	case duckdb.Interval:
		// Our current policy is to convert INTERVALs to milliseconds, treating one month as 30 days.
		ms := v.Micros / 1000
		ms += int64(v.Days) * 24 * 60 * 60 * 1000
		ms += int64(v.Months) * 30 * 24 * 60 * 60 * 1000
		return ms, nil
	case net.IP:
		return v.String(), nil
	case orb.Point:
		return []any{v[0], v[1]}, nil
	case *net.IP:
		if v != nil {
			return ToValue(*v, t)
		}
		return nil, nil
	case *bool:
		if v != nil {
			return *v, nil
		}
		return nil, nil
	case *int:
		if v != nil {
			return int64(*v), nil
		}
		return nil, nil
	case *int8:
		if v != nil {
			return int64(*v), nil
		}
		return nil, nil
	case *int16:
		if v != nil {
			return int64(*v), nil
		}
		return nil, nil
	case *int32:
		if v != nil {
			return int64(*v), nil
		}
		return nil, nil
	case *int64:
		if v != nil {
			return *v, nil
		}
		return nil, nil
	case *uint:
		if v != nil {
			return uint64(*v), nil
		}
		return nil, nil
	case *uint8:
		if v != nil {
			return uint64(*v), nil
		}
		return nil, nil
	case *uint16:
		if v != nil {
			return uint64(*v), nil
		}
		return nil, nil
	case *uint32:
		if v != nil {
			return uint64(*v), nil
		}
		return nil, nil
	case *uint64:
		if v != nil {
			return *v, nil
		}
		return nil, nil
	case *float32:
		if v != nil {
			return ToValue(*v, t)
		}
		return nil, nil
	case *float64:
		if v != nil {
			return ToValue(*v, t)
		}
		return nil, nil
	case *time.Time:
		if v != nil {
			return ToValue(*v, t)
		}
	case *string:
		if v != nil {
			return ToValue(*v, t)
		}
		return nil, nil
	default:
	}
	if t != nil && t.ArrayElementType != nil {
		return toSliceUnknown(v, t)
	}
	if t != nil && t.MapType != nil {
		return toMapCoerceKeysUnknown(v, t.MapType)
	}
	return v, nil
}

// toMap converts a map. Providing t as a type hint is optional.
func toMap(v map[string]any, t *runtimev1.StructType) (map[string]any, error) {
	if t == nil {
		for k, v2 := range v {
			var err error
			v[strings.ToValidUTF8(k, "�")], err = ToValue(v2, nil)
			if err != nil {
				return nil, err
			}
		}
	} else {
		for _, f := range t.Fields {
			var err error
			v[f.Name], err = ToValue(v[f.Name], f.Type)
			if err != nil {
				return nil, err
			}
		}
	}
	return v, nil
}

// toMapCoerceKeys converts a map with non-string keys to a map[string]any.
// It attempts to coerce the keys to JSON strings. Providing t as a type hint is optional.
func toMapCoerceKeys(v map[any]any, t *runtimev1.MapType) (map[string]any, error) {
	var keyType, valType *runtimev1.Type
	if t != nil {
		keyType = t.KeyType
		valType = t.ValueType
	}

	x := make(map[string]any, len(v))
	for k1, v := range v {
		k2, ok := k1.(string)
		if !ok {
			// Encode k1 using ToValue (to correctly coerce time, big numbers, etc.) and then to JSON.
			// This yields more idiomatic/consistent strings than using fmt.Sprintf("%v", k1).
			val, err := ToValue(k1, keyType)
			if err != nil {
				return nil, err
			}

			data, err := json.Marshal(val)
			if err != nil {
				return nil, err
			}

			// Remove surrounding quotes returned by MarshalJSON for strings
			k2 = TrimQuotes(string(data))
		}

		var err error
		x[k2], err = ToValue(v, valType)
		if err != nil {
			return nil, err
		}
	}
	return x, nil
}

// toMapCoerceKeysUnknown is similar to toMapCoerceKeys, but when type of map is unknown.
// It uses reflection so should not be used when toMapCoerceKeys can be used.
func toMapCoerceKeysUnknown(s any, t *runtimev1.MapType) (map[string]any, error) {
	if s == nil { // defensive check, nil is already handled upstream
		return nil, nil
	}
	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Ptr {
		return toMapCoerceKeysUnknown(v.Elem().Interface(), t)
	}
	if v.Kind() != reflect.Map {
		return nil, fmt.Errorf("received invalid type %T, expected map", s)
	}

	var keyType, valType *runtimev1.Type
	if t != nil {
		keyType = t.KeyType
		valType = t.ValueType
	}
	x := make(map[string]any, v.Len())
	iter := v.MapRange()
	for iter.Next() {
		k1 := iter.Key()
		k2, ok := k1.Interface().(string)
		if !ok {
			// Encode k1 using ToValue (to correctly coerce time, big numbers, etc.) and then to JSON.
			// This yields more idiomatic/consistent strings than using fmt.Sprintf("%v", k1).
			val, err := ToValue(k1.Interface(), keyType)
			if err != nil {
				return nil, err
			}

			data, err := json.Marshal(val)
			if err != nil {
				return nil, err
			}

			// Remove surrounding quotes returned by MarshalJSON for strings
			k2 = TrimQuotes(string(data))
		}

		var err error
		x[k2], err = ToValue(iter.Value().Interface(), valType)
		if err != nil {
			return nil, err
		}
	}
	return x, nil
}

// toSlice converts a slice. Providing t as a type hint is optional.
func toSlice(v []any, t *runtimev1.Type) ([]any, error) {
	var elemType *runtimev1.Type
	if t != nil {
		elemType = t.ArrayElementType
	}
	for i, v2 := range v {
		var err error
		v[i], err = ToValue(v2, elemType)
		if err != nil {
			return nil, err
		}
	}
	return v, nil
}

// toSliceUnknown converts a slice similar to toSlice, but when the type of slice is unknown.
// It uses reflection so should not be used when toSlice can be used.
func toSliceUnknown(s any, t *runtimev1.Type) ([]any, error) {
	if s == nil { // defensive check, nil is already handled upstream
		return nil, nil
	}
	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Ptr {
		return toSliceUnknown(v.Elem().Interface(), t)
	}
	if v.Kind() != reflect.Slice {
		return nil, fmt.Errorf("received invalid type %T, expected slice", s)
	}
	var elemType *runtimev1.Type
	if t != nil {
		elemType = t.ArrayElementType
	}
	x := make([]any, v.Len())
	for i := 0; i < v.Len(); i++ {
		var err error
		x[i], err = ToValue(v.Index(i).Interface(), elemType)
		if err != nil {
			return nil, err
		}
	}
	return x, nil
}

// TrimQuotes removes surrounding double quotes from a string, if present.
// Examples: `"10"` -> `10` and `10` -> `10`.
func TrimQuotes(s string) string {
	if len(s) >= 2 {
		if s[0] == '"' && s[len(s)-1] == '"' {
			return s[1 : len(s)-1]
		}
	}
	return s
}
