package container

import (
	"testing"

	"github.com/rilldata/rill/runtime/pkg/arrayutil"
	"github.com/stretchr/testify/require"
)

func TestOverCapacity(t *testing.T) {
	c, err := NewBoundedContainer[int](32)
	require.NoError(t, err)

	i := 0
	for ; i < 50 && !c.IsFull(); i += 1 {
		c.Add(i)
	}
	require.Equal(t, 32, i)
	require.Equal(t, arrayutil.RangeInt(0, 32, false), c.Items())
}

func TestUnderCapacity(t *testing.T) {
	c, err := NewBoundedContainer[int](32)
	require.NoError(t, err)

	i := 0
	for ; i < 16 && !c.IsFull(); i += 1 {
		c.Add(i)
	}
	require.Equal(t, 16, i)
	require.Equal(t, arrayutil.RangeInt(0, 16, false), c.Items())
}

func TestMatchCapacity(t *testing.T) {
	c, err := NewBoundedContainer[int](32)
	require.NoError(t, err)

	i := 0
	for ; i < 32 && !c.IsFull(); i += 1 {
		c.Add(i)
	}
	require.Equal(t, 32, i)
	require.Equal(t, arrayutil.RangeInt(0, 32, false), c.Items())
}

func TestError(t *testing.T) {
	c, err := NewBoundedContainer[int](-1)
	require.Error(t, err)
	require.Nil(t, c)
}
