package mapstructureutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestStrToTime(t *testing.T) {
	ts := time.Now()
	in := map[string]any{"Val": ts.Format(time.RFC3339Nano)}
	out := struct{ Val time.Time }{}
	err := WeakDecode(in, &out)
	require.NoError(t, err)
	require.True(t, ts.Equal(out.Val))
}

func TestStrToTimePtr(t *testing.T) {
	in := map[string]any{}
	out := struct{ Val *time.Time }{}
	err := WeakDecode(in, &out)
	require.NoError(t, err)
	require.Nil(t, out.Val)

	ts := time.Now()
	in["Val"] = ts.Format(time.RFC3339Nano)
	err = WeakDecode(in, &out)
	require.NoError(t, err)
	require.True(t, ts.Equal(*out.Val))
}

func TestWeakDecodeWithWarnings(t *testing.T) {
	type target struct {
		Name string `mapstructure:"name"`
		Age  int    `mapstructure:"age"`
	}

	in := map[string]any{
		"name":    "Alice",
		"age":     "30",
		"unknown": "value",
		"extra":   42,
	}
	out := &target{}
	unused, err := WeakDecodeWithWarnings(in, out)
	require.NoError(t, err)
	require.Equal(t, "Alice", out.Name)
	require.Equal(t, 30, out.Age) // weakly typed: string "30" -> int 30
	require.ElementsMatch(t, []string{"unknown", "extra"}, unused)
}

func TestWeakDecodeWithWarnings_NoUnused(t *testing.T) {
	type target struct {
		Name string `mapstructure:"name"`
	}

	in := map[string]any{"name": "Bob"}
	out := &target{}
	unused, err := WeakDecodeWithWarnings(in, out)
	require.NoError(t, err)
	require.Equal(t, "Bob", out.Name)
	require.Empty(t, unused)
}
