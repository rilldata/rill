package container

import (
	"testing"

	"github.com/rilldata/rill/runtime/pkg/arrayutil"
	"github.com/stretchr/testify/require"
)

func TestTailOverCapacity(t *testing.T) {
	cleanupcalled := 0
	c, err := NewTailContainer(32, func(int) { cleanupcalled++ })
	require.NoError(t, err)

	i := 0
	for ; i < 50 && !c.IsFull(); i += 1 {
		c.Add(i)
	}
	require.Equal(t, 18, cleanupcalled)
	require.Equal(t, 50, i)
	require.Equal(t, arrayutil.RangeInt(18, 50, true), c.Items())
}

func TestTailUnderCapacity(t *testing.T) {
	c, err := NewTailContainer(32, func(int) {})
	require.NoError(t, err)

	i := 0
	for ; i < 16 && !c.IsFull(); i += 1 {
		c.Add(i)
	}
	require.Equal(t, 16, i)
	require.Equal(t, arrayutil.RangeInt(0, 16, true), c.Items())
}

func TestTailMatchCapacity(t *testing.T) {
	c, err := NewTailContainer(32, func(int) {})
	require.NoError(t, err)

	i := 0
	for ; i < 32 && !c.IsFull(); i += 1 {
		c.Add(i)
	}
	require.Equal(t, 32, i)
	require.Equal(t, arrayutil.RangeInt(0, 32, true), c.Items())
}

func TestTailError(t *testing.T) {
	c, err := NewTailContainer(-1, func(int) {})
	require.Error(t, err)
	require.Nil(t, c)
}
