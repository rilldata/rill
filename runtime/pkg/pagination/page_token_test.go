package pagination

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPageToken(t *testing.T) {
	tc := func(args []any, out ...any) {
		token := MarshalPageToken(args...)
		err := UnmarshalPageToken(token, out...)
		require.NoError(t, err)
		require.Equal(t, len(args), len(out))
		for i := range args {
			require.Equal(t, args[i], reflect.ValueOf(out[i]).Elem().Interface())
		}
	}

	var a1, b1, c1 string
	tc([]any{"a", "b", "c"}, &a1, &b1, &c1)

	var a2 string
	var b2 int
	tc([]any{"a", int(10)}, &a2, &b2)

	var a3 string
	var b3 time.Time
	tm, _ := time.Parse(time.DateOnly, "2024-01-01")
	tc([]any{"a", tm}, &a3, &b3)

	var a4 any
	var b4 bool
	tc([]any{nil, true}, &a4, &b4)
}
