package duckdb

import (
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/stretchr/testify/require"
)

func TestDatabaseTypeToPB(t *testing.T) {
	tests := []struct {
		input  string
		output *runtimev1.Type
	}{
		{
			input:  "DECIMAL(10,20)",
			output: &runtimev1.Type{Code: runtimev1.Type_CODE_DECIMAL, Nullable: true},
		},
		{
			input: `STRUCT(foo HUGEINT, "bar" STRUCT(a INTEGER, b MAP(INTEGER, BOOLEAN)), baz VARCHAR[])`,
			output: &runtimev1.Type{Code: runtimev1.Type_CODE_STRUCT, Nullable: true, StructType: &runtimev1.StructType{Fields: []*runtimev1.StructType_Field{
				{Name: "foo", Type: &runtimev1.Type{Code: runtimev1.Type_CODE_INT128, Nullable: true}},
				{Name: "bar", Type: &runtimev1.Type{Code: runtimev1.Type_CODE_STRUCT, Nullable: true, StructType: &runtimev1.StructType{Fields: []*runtimev1.StructType_Field{
					{Name: "a", Type: &runtimev1.Type{Code: runtimev1.Type_CODE_INT32, Nullable: true}},
					{Name: "b", Type: &runtimev1.Type{Code: runtimev1.Type_CODE_MAP, Nullable: true, MapType: &runtimev1.MapType{
						KeyType:   &runtimev1.Type{Code: runtimev1.Type_CODE_INT32, Nullable: true},
						ValueType: &runtimev1.Type{Code: runtimev1.Type_CODE_BOOL, Nullable: true},
					}}},
				}}}},
				{Name: "baz", Type: &runtimev1.Type{Code: runtimev1.Type_CODE_ARRAY, Nullable: true, ArrayElementType: &runtimev1.Type{Code: runtimev1.Type_CODE_STRING, Nullable: true}}},
			}}},
		},
		{
			input: `STRUCT("foo ""("" bar" STRUCT("baz ,, \ \"" "" )" INTEGER))`,
			output: &runtimev1.Type{Code: runtimev1.Type_CODE_STRUCT, Nullable: true, StructType: &runtimev1.StructType{Fields: []*runtimev1.StructType_Field{
				{Name: `foo "(" bar`, Type: &runtimev1.Type{Code: runtimev1.Type_CODE_STRUCT, Nullable: true, StructType: &runtimev1.StructType{Fields: []*runtimev1.StructType_Field{
					{Name: `baz ,, \ \" " )`, Type: &runtimev1.Type{Code: runtimev1.Type_CODE_INT32, Nullable: true}},
				}}}},
			}}},
		},
		// Array having struct
		{
			input: `STRUCT(id BIGINT, name VARCHAR)[]`,
			output: &runtimev1.Type{
				Code:     runtimev1.Type_CODE_ARRAY,
				Nullable: true,
				ArrayElementType: &runtimev1.Type{
					Code:     runtimev1.Type_CODE_STRUCT,
					Nullable: true,
					StructType: &runtimev1.StructType{Fields: []*runtimev1.StructType_Field{
						{Name: "id", Type: &runtimev1.Type{Code: runtimev1.Type_CODE_INT64, Nullable: true}},
						{Name: "name", Type: &runtimev1.Type{Code: runtimev1.Type_CODE_STRING, Nullable: true}},
					}},
				},
			},
		},
		// Array of struct having array
		{
			input: `STRUCT(id BIGINT, tags VARCHAR[])[]`,
			output: &runtimev1.Type{
				Code:     runtimev1.Type_CODE_ARRAY,
				Nullable: true,
				ArrayElementType: &runtimev1.Type{
					Code:     runtimev1.Type_CODE_STRUCT,
					Nullable: true,
					StructType: &runtimev1.StructType{Fields: []*runtimev1.StructType_Field{
						{Name: "id", Type: &runtimev1.Type{Code: runtimev1.Type_CODE_INT64, Nullable: true}},
						{Name: "tags", Type: &runtimev1.Type{
							Code:             runtimev1.Type_CODE_ARRAY,
							Nullable:         true,
							ArrayElementType: &runtimev1.Type{Code: runtimev1.Type_CODE_STRING, Nullable: true},
						}},
					}},
				},
			},
		},
	}

	for _, test := range tests {
		output, err := databaseTypeToPB(test.input, true)
		require.NoError(t, err)
		require.Equal(t, test.output, output)
	}
}
