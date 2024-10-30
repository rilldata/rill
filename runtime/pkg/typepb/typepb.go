package typepb

import (
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

// InferFromValue attempts to infer type from a value.
func InferFromValue(val any) *runtimev1.Type {
	switch val := val.(type) {
	case bool:
		return &runtimev1.Type{Code: runtimev1.Type_CODE_BOOL}
	case int8:
		return &runtimev1.Type{Code: runtimev1.Type_CODE_INT8}
	case int16:
		return &runtimev1.Type{Code: runtimev1.Type_CODE_INT16}
	case int32:
		return &runtimev1.Type{Code: runtimev1.Type_CODE_INT32}
	case int, int64:
		return &runtimev1.Type{Code: runtimev1.Type_CODE_INT64}
	case uint8:
		return &runtimev1.Type{Code: runtimev1.Type_CODE_UINT8}
	case uint16:
		return &runtimev1.Type{Code: runtimev1.Type_CODE_UINT16}
	case uint32:
		return &runtimev1.Type{Code: runtimev1.Type_CODE_UINT32}
	case uint, uint64:
		return &runtimev1.Type{Code: runtimev1.Type_CODE_UINT64}
	case float32:
		return &runtimev1.Type{Code: runtimev1.Type_CODE_FLOAT32}
	case float64:
		return &runtimev1.Type{Code: runtimev1.Type_CODE_FLOAT64}
	case time.Time:
		return &runtimev1.Type{Code: runtimev1.Type_CODE_TIMESTAMP}
	case string:
		return &runtimev1.Type{Code: runtimev1.Type_CODE_STRING}
	case []byte:
		return &runtimev1.Type{Code: runtimev1.Type_CODE_BYTES, Nullable: true}
	case []any:
		var elemType *runtimev1.Type
		if len(val) > 0 {
			elemType = InferFromValue(val[0])
		} else {
			elemType = &runtimev1.Type{Code: runtimev1.Type_CODE_UNSPECIFIED}
		}
		return &runtimev1.Type{Code: runtimev1.Type_CODE_ARRAY, Nullable: true, ArrayElementType: elemType}
	case []string:
		return &runtimev1.Type{Code: runtimev1.Type_CODE_ARRAY, Nullable: true, ArrayElementType: &runtimev1.Type{Code: runtimev1.Type_CODE_STRING, Nullable: true}}
	case []int:
		return &runtimev1.Type{Code: runtimev1.Type_CODE_ARRAY, Nullable: true, ArrayElementType: &runtimev1.Type{Code: runtimev1.Type_CODE_INT64}}
	case map[string]any:
		t := &runtimev1.StructType{}
		for k, v := range val {
			t.Fields = append(t.Fields, &runtimev1.StructType_Field{Name: k, Type: InferFromValue(v)})
		}
		return &runtimev1.Type{Code: runtimev1.Type_CODE_STRUCT, Nullable: true, StructType: t}
	default:
		return &runtimev1.Type{Code: runtimev1.Type_CODE_UNSPECIFIED}
	}
}
