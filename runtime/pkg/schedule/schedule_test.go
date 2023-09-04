package schedule

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSchedule(t *testing.T) {
	t1 := time.Unix(10, 0)
	t2 := time.Unix(20, 0)
	t3 := time.Unix(30, 0)
	t4 := time.Unix(40, 0)

	s := New(func(i int) int { return i })
	s.Set(1, t1)
	s.Set(2, t2)
	s.Set(3, t3)
	s.Set(4, t4)
	s.Set(1, t4)
	require.Equal(t, 4, s.Len())

	i, ts := s.Peek()
	require.Equal(t, 2, i)
	require.Equal(t, t2, ts)
	i = s.Pop()
	require.Equal(t, 2, i)
	require.Equal(t, 3, s.Len())

	s.Remove(4)
	require.Equal(t, 2, s.Len())

	i, ts = s.Peek()
	require.Equal(t, 3, i)
	require.Equal(t, t3, ts)
	i = s.Pop()
	require.Equal(t, 3, i)
	require.Equal(t, 1, s.Len())

	i, ts = s.Peek()
	require.Equal(t, 1, i)
	require.Equal(t, t4, ts)
	i = s.Pop()
	require.Equal(t, 1, i)
	require.Equal(t, 0, s.Len())

	i, ts = s.Peek()
	require.Equal(t, 0, i)
	require.True(t, ts.IsZero())
}
