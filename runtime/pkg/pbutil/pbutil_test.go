package pbutil

import (
	"math/big"
	"testing"
	"time"

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

		actual, err := ToStructCoerceKeys(tt.Input)
		require.NoError(t, err)

		require.True(t, proto.Equal(expected, actual))
	}
}
