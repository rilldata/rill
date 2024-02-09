package pbutil

import (
	"math/big"
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestToStructCoerceKeys(t *testing.T) {
	cases := []struct {
		Input    map[any]any
		Expected map[string]any
	}{
		{Input: map[any]any{10: 1}, Expected: map[string]any{"10": 1}},
		{Input: map[any]any{big.NewInt(10): 1}, Expected: map[string]any{"10": 1}},
		{Input: map[any]any{3.141: 1}, Expected: map[string]any{"3.141": 1}},
		{Input: map[any]any{3.141: map[any]any{1: 2}}, Expected: map[string]any{"3.141": map[string]any{"1": 2}}},
		{Input: map[any]any{time.Date(2020, 01, 01, 0, 0, 0, 0, time.UTC): 20}, Expected: map[string]any{"2020-01-01T00:00:00Z": 20}},
	}
	for _, tt := range cases {
		expected, err := structpb.NewStruct(tt.Expected)
		require.NoError(t, err)

		actual, err := ToStructCoerceKeys(tt.Input, nil)
		require.NoError(t, err)

		require.True(t, proto.Equal(expected, actual))
	}
}

func TestToStructCoerceKeysUnknown(t *testing.T) {
	cases := []struct {
		Input    any
		Expected map[string]any
		MapType  *runtimev1.MapType
	}{
		{Input: map[int]int{10: 1}, Expected: map[string]any{"10": 1}},
		{Input: map[*big.Int]int{big.NewInt(10): 1}, Expected: map[string]any{"10": 1}},
		{Input: map[float64]int{3.141: 1}, Expected: map[string]any{"3.141": 1}},
		{Input: map[float64]map[int]int{3.141: {1: 2}}, Expected: map[string]any{"3.141": map[string]any{"1": 2}}, MapType: &runtimev1.MapType{ValueType: &runtimev1.Type{MapType: &runtimev1.MapType{}}}},
		{Input: map[time.Time]int{time.Date(2020, 01, 01, 0, 0, 0, 0, time.UTC): 20}, Expected: map[string]any{"2020-01-01T00:00:00Z": 20}},
	}
	for i, tt := range cases {
		expected, err := structpb.NewStruct(tt.Expected)
		require.NoError(t, err)

		actual, err := ToStructCoerceKeysUnknown(tt.Input, tt.MapType)
		require.NoError(t, err, "case %d", i)

		require.True(t, proto.Equal(expected, actual), "expected: %v, actual: %v", expected, actual)
	}
}

func Test_ToListValueUnknown(t *testing.T) {
	cases := []struct {
		Input    any
		Expected *structpb.ListValue
		Type     *runtimev1.Type
	}{
		{Input: []int64{1, 2, 3, 4}, Expected: structpbList([]interface{}{1, 2, 3, 4})},
		{Input: &[]int16{1, 2, 3, 4}, Expected: structpbList([]interface{}{1, 2, 3, 4})},
		{Input: listPtr(&[]int16{1, 2, 3, 4}), Expected: structpbList([]interface{}{1, 2, 3, 4})},
		{Input: &[]*uint8{intPtr(1), intPtr(2), intPtr(3), intPtr(4)}, Expected: structpbList([]interface{}{1, 2, 3, 4})},
		{Input: &[][]int8{{1, 1}, {2, 2}, {3, 3}}, Expected: structpbListList([]any{[]any{1, 1}, []any{2, 2}, []any{3, 3}}), Type: &runtimev1.Type{Code: runtimev1.Type_CODE_ARRAY, ArrayElementType: &runtimev1.Type{Code: runtimev1.Type_CODE_ARRAY, ArrayElementType: &runtimev1.Type{Code: runtimev1.Type_CODE_INT8}}}},
		{Input: nil, Expected: nil, Type: nil},
	}
	for i, tt := range cases {
		actual, err := ToListValueUnknown(tt.Input, tt.Type)
		require.NoError(t, err, "test %v", i+1)

		require.True(t, proto.Equal(actual, tt.Expected), "test %v, actual: %v, expected: %v", i, actual, tt.Expected)
	}
}

func structpbList(v []interface{}) *structpb.ListValue {
	list, err := structpb.NewList(v)
	if err != nil {
		panic(err)
	}
	return list
}

func structpbListList(x []interface{}) *structpb.ListValue {
	list := &structpb.ListValue{Values: make([]*structpb.Value, len(x))}
	var err error
	for i, v := range x {
		list.Values[i], err = structpb.NewValue(v)
		if err != nil {
			panic(err)
		}
	}
	return list
}

func intPtr(v uint8) *uint8 {
	return &v
}

func listPtr(v *[]int16) **[]int16 {
	return &v
}
