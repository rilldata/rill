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
