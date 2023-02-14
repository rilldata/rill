package container

import (
	"testing"

	"github.com/rilldata/rill/runtime/pkg/arrayutil"
	"github.com/stretchr/testify/require"
)

func TestFIFOOverCapacity(t *testing.T) {

	c, err := NewFIFO[int](32, nil)
	require.NoError(t, err)

	i := 0
	for ; i < 50 && !c.Full(); i += 1 {
		c.Add(i)
	}
	require.Equal(t, 50, i)
	require.Equal(t, arrayutil.RangeInt(18, 50, true), c.Items())
}

func TestFIFOUnderCapacity(t *testing.T) {
	c, err := NewFIFO[int](32, nil)
	require.NoError(t, err)

	i := 0
	for ; i < 16 && !c.Full(); i += 1 {
		c.Add(i)
	}
	require.Equal(t, 16, i)
	require.Equal(t, arrayutil.RangeInt(0, 16, true), c.Items())
}

func TestFIFOMatchCapacity(t *testing.T) {
	c, err := NewFIFO[int](32, nil)
	require.NoError(t, err)

	i := 0
	for ; i < 32 && !c.Full(); i += 1 {
		c.Add(i)
	}
	require.Equal(t, 32, i)
	require.Equal(t, arrayutil.RangeInt(0, 32, true), c.Items())
}

func TestFIFOWithCleanup(t *testing.T) {
	cleanupcalled := 0
	c, err := NewFIFO(32, func(item int) { cleanupcalled++ })
	require.NoError(t, err)

	i := 0
	for ; i < 50 && !c.Full(); i += 1 {
		c.Add(i)
	}
	require.Equal(t, 18, cleanupcalled)
}

func TestFIFOError(t *testing.T) {
	c, err := NewFIFO[int](-1, nil)
	require.Error(t, err)
	require.Nil(t, c)
}
