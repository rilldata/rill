package priorityqueue

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPriorityQueue(t *testing.T) {
	pq := New[int]()
	require.Equal(t, 0, pq.Len())
	pq.Push(1, 1)
	pq.Push(2, 2)
	itm := pq.Push(3, 3)
	require.Equal(t, 3, pq.Len())
	require.True(t, pq.Contains(itm))
	pq.Remove(itm)
	require.False(t, pq.Contains(itm))
	require.Equal(t, 2, pq.Pop())
	require.Equal(t, 1, pq.Len())
	require.Equal(t, 1, pq.Pop())
	require.Equal(t, 0, pq.Len())
}
