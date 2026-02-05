package sqlstring

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// ToLiteral converts a Go value to its corresponding SQL literal representation.
// It handles primitive types and lists. All other types are emitted as JSON-encoded strings.
func ToLiteral(val any) string {
	if val == nil {
		return "NULL"
	}

	v := reflect.ValueOf(val)

	// Unwrap pointers
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return "NULL"
		}
		v = v.Elem()
	}

	// Check for time.Time after unwrapping pointers
	if v.Type() == reflect.TypeOf(time.Time{}) {
		return "'" + v.Interface().(time.Time).Format(time.RFC3339Nano) + "'"
	}

	switch v.Kind() {
	case reflect.String:
		return "'" + strings.ReplaceAll(v.String(), "'", "''") + "'"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10)
	case reflect.Float32:
		return strconv.FormatFloat(v.Float(), 'f', -1, 32)
	case reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, 64)
	case reflect.Bool:
		if v.Bool() {
			return "TRUE"
		}
		return "FALSE"
	case reflect.Slice, reflect.Array:
		// []byte → hex literal
		if v.Type().Elem().Kind() == reflect.Uint8 {
			return "X'" + hex.EncodeToString(v.Bytes()) + "'"
		}
		if v.Len() == 0 {
			return "(NULL)"
		}
		parts := make([]string, v.Len())
		for i := range parts {
			parts[i] = ToLiteral(v.Index(i).Interface())
		}
		return "(" + strings.Join(parts, ", ") + ")"
	default:
		// Fallback: JSON-encode and treat as string
		b, err := json.Marshal(v.Interface())
		if err != nil {
			b = fmt.Appendf([]byte{}, "<json error: %s>", err.Error())
		}
		return "'" + strings.ReplaceAll(string(b), "'", "''") + "'"
	}
}
