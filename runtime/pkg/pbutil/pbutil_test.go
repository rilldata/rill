package pbutil

import (
	"math/big"
	"net"
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
		{Input: []*uint8{intPtr(1), intPtr(2), intPtr(3), intPtr(4)}, Expected: structpbList([]interface{}{1, 2, 3, 4})},
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

func TestToValue_NilPointers(t *testing.T) {
	expected := structpb.NewNullValue()

	cases := []struct {
		name  string
		input any
	}{
		{name: "*bool", input: (*bool)(nil)},
		{name: "*int", input: (*int)(nil)},
		{name: "*int32", input: (*int32)(nil)},
		{name: "*int64", input: (*int64)(nil)},
		{name: "*uint", input: (*uint)(nil)},
		{name: "*uint32", input: (*uint32)(nil)},
		{name: "*uint64", input: (*uint64)(nil)},
		{name: "*string", input: (*string)(nil)},
		{name: "*int8", input: (*int8)(nil)},
		{name: "*int16", input: (*int16)(nil)},
		{name: "*uint8", input: (*uint8)(nil)},
		{name: "*uint16", input: (*uint16)(nil)},
		{name: "*time.Time", input: (*time.Time)(nil)},
		{name: "*float32", input: (*float32)(nil)},
		{name: "*float64", input: (*float64)(nil)},
		{name: "*big.Int", input: (*big.Int)(nil)},
		{name: "*net.IP", input: (*net.IP)(nil)},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := ToValue(tt.input, nil)
			require.NoError(t, err)
			require.Equal(t, expected, actual)
		})
	}
}

func TestToValue_PointerValues(t *testing.T) {
	now := time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC)

	boolVal := true
	intVal := int(42)
	int32Val := int32(43)
	int64Val := int64(44)
	uintVal := uint(45)
	uint32Val := uint32(46)
	uint64Val := uint64(47)
	strVal := "hello"
	int8Val := int8(48)
	int16Val := int16(49)
	uint8Val := uint8(50)
	uint16Val := uint16(51)
	timeVal := now
	float32Val := float32(1.5)
	float64Val := float64(2.5)
	ipVal := net.ParseIP("192.168.0.1")
	bigIntVal := *big.NewInt(12345)

	cases := []struct {
		name     string
		input    any
		expected *structpb.Value
	}{
		{name: "*bool", input: &boolVal, expected: structpb.NewBoolValue(true)},
		{name: "*int", input: &intVal, expected: structpb.NewNumberValue(42)},
		{name: "*int32", input: &int32Val, expected: structpb.NewNumberValue(43)},
		{name: "*int64", input: &int64Val, expected: structpb.NewNumberValue(44)},
		{name: "*uint", input: &uintVal, expected: structpb.NewNumberValue(45)},
		{name: "*uint32", input: &uint32Val, expected: structpb.NewNumberValue(46)},
		{name: "*uint64", input: &uint64Val, expected: structpb.NewNumberValue(47)},
		{name: "*string", input: &strVal, expected: structpb.NewStringValue("hello")},
		{name: "*int8", input: &int8Val, expected: structpb.NewNumberValue(48)},
		{name: "*int16", input: &int16Val, expected: structpb.NewNumberValue(49)},
		{name: "*uint8", input: &uint8Val, expected: structpb.NewNumberValue(50)},
		{name: "*uint16", input: &uint16Val, expected: structpb.NewNumberValue(51)},
		{name: "*time.Time", input: &timeVal, expected: structpb.NewStringValue(now.In(time.UTC).Format(time.RFC3339Nano))},
		{name: "*float32", input: &float32Val, expected: structpb.NewNumberValue(float64(float32Val))},
		{name: "*float64", input: &float64Val, expected: structpb.NewNumberValue(float64Val)},
		{name: "*net.IP", input: &ipVal, expected: structpb.NewStringValue(ipVal.String())},
		{name: "*big.Int", input: &bigIntVal, expected: structpb.NewNumberValue(func() float64 { f, _ := new(big.Float).SetInt(&bigIntVal).Float64(); return f }())},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := ToValue(tt.input, nil)
			require.NoError(t, err)
			require.True(t, proto.Equal(tt.expected, actual), "expected: %v, actual: %v", tt.expected, actual)
		})
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
