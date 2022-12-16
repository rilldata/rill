package priorityqueue

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

func TestSemaphoreSimple(t *testing.T) {
	s := NewSemaphore(1)
	err := s.Acquire(context.Background(), 1)
	require.NoError(t, err)
	s.Release()
	err = s.Acquire(context.Background(), 1)
	require.NoError(t, err)
	s.Release()
	require.True(t, s.TryAcquire())
	require.False(t, s.TryAcquire())
	s.Release()

	s = NewSemaphore(2)
	err = s.Acquire(context.Background(), 1)
	require.NoError(t, err)
	err = s.Acquire(context.Background(), 1)
	require.NoError(t, err)
	s.Release()
	s.Release()
	require.Panics(t, func() {
		s.Release()
	})
}

func TestSemaphorePriority(t *testing.T) {
	// Prepare
	n := 10
	var results []int
	var g errgroup.Group
	s := NewSemaphore(1)

	// Acquire to block
	err := s.Acquire(context.Background(), 1)
	require.NoError(t, err)
	require.False(t, s.TryAcquire())

	// Fill up queue
	for i := 0; i <= n; i++ {
		priority := i
		g.Go(func() error {
			err := s.Acquire(context.Background(), priority)
			require.NoError(t, err)
			results = append(results, priority)
			s.Release()
			return nil
		})
	}

	// Wait a bit to ensure queue fills up
	time.Sleep(time.Second)
	s.Release()

	// Wait for processing
	err = g.Wait()
	require.NoError(t, err)

	// Check results evaluated in priority order
	for i := 0; i <= n; i++ {
		require.Equal(t, i, results[n-i])
	}
}

func TestSemaphoreCancel(t *testing.T) {
	// Prepare
	n := 100
	size := 4
	cancelIdx := 50
	results := make(chan int, n-1)
	var g errgroup.Group
	s := NewSemaphore(size)

	// Acquire up to size
	for i := 0; i < size; i++ {
		err := s.Acquire(context.Background(), 1)
		require.NoError(t, err)
	}
	require.False(t, s.TryAcquire())

	// Fill up queue
	for i := 0; i < n; i++ {
		priority := i
		g.Go(func() error {
			ctx := context.Background()
			if priority == cancelIdx {
				cctx, cancel := context.WithCancel(ctx)
				ctx = cctx
				cancel()
			}

			err := s.Acquire(ctx, priority)
			if priority == cancelIdx {
				require.Error(t, err)
				return nil
			}

			results <- priority
			s.Release()
			return nil
		})
	}

	// Release to unblock for processing
	for i := 0; i < size; i++ {
		s.Release()
	}

	// Wait for processing
	err := g.Wait()
	require.NoError(t, err)

	for i := 0; i < n-1; i++ {
		require.NotEqual(t, cancelIdx, <-results)
	}
}
