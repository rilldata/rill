package dag2

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAcyclic(t *testing.T) {
	d := New(hash)
	require.True(t, d.Add(1, 2))
	require.True(t, d.Add(2, 3, 4))
	require.True(t, d.Add(3, 4))
	require.False(t, d.Add(4, 1))
	require.Len(t, d.vertices, 4)
	require.True(t, d.Add(4))
	require.ElementsMatch(t, []int{1, 2, 3}, d.DeepChildren(4))
}

func TestRetention(t *testing.T) {
	d := New(hash)
	require.True(t, d.Add(1, 2))
	require.True(t, d.Add(2, 3))
	require.Len(t, d.vertices, 3)

	require.True(t, d.Add(3))
	require.Len(t, d.vertices, 3)
	require.ElementsMatch(t, []int{3}, d.Parents(2, true))
	require.ElementsMatch(t, []int{3}, d.Parents(2, false))

	d.Remove(2)
	require.Len(t, d.vertices, 3)
	require.ElementsMatch(t, []int{}, d.Children(3))
	require.ElementsMatch(t, []int{}, d.Parents(1, true))
	require.ElementsMatch(t, []int{2}, d.Parents(1, false))

	d.Remove(1)
	require.Len(t, d.vertices, 1)
	require.ElementsMatch(t, []int{}, d.Children(3))
}

func TestPanics(t *testing.T) {
	// Already exists
	d := New(hash)
	require.True(t, d.Add(1))
	require.Panics(t, func() { d.Add(1) })

	// Doesn't exist
	d = New(hash)
	require.True(t, d.Add(1))
	require.Panics(t, func() { d.Remove(2) })
	require.Panics(t, func() { d.Parents(2, false) })
	require.Panics(t, func() { d.Children(2) })
}

func hash(i int) int {
	return i
}
