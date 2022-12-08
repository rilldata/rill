package priorityworker

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

func TestPriorityQueue(t *testing.T) {
	n := 100
	results := make(chan int, n)
	var g errgroup.Group

	pw := New(func(ctx context.Context, i int) error {
		results <- i
		return nil
	}, 4)
	pw.Pause() // ensure predictable output

	for i := n; i > 0; i-- {
		priority := i
		g.Go(func() error {
			return pw.Process(context.Background(), priority, priority)
		})
	}

	// give the queue plenty of time to fill up, then unpause
	time.Sleep(100 * time.Millisecond)
	pw.Unpause()

	err := g.Wait()
	require.NoError(t, err)

	for i := n; i > 0; i-- {
		x := <-results
		require.Equal(t, i, x)
	}
}

func TestCancel(t *testing.T) {
	n := 100
	cancelIdx := 50
	results := make(chan int, n)
	var g errgroup.Group

	pw := New(func(ctx context.Context, i int) error {
		time.Sleep(2 * time.Millisecond)
		results <- i
		return nil
	}, 4)
	pw.Pause() // ensure predictable output

	for i := n; i > 0; i-- {
		priority := i
		g.Go(func() error {
			ctx := context.Background()
			if priority == cancelIdx {
				cctx, cancel := context.WithCancel(ctx)
				ctx = cctx
				go func() {
					time.Sleep(10 * time.Millisecond)
					cancel()
				}()
			}

			err := pw.Process(ctx, priority, priority)
			if priority == cancelIdx {
				require.Error(t, err)
				return nil
			}
			return err
		})
	}

	// we can unpause immediately
	pw.Unpause()

	err := g.Wait()
	require.NoError(t, err)

	for i := n; i > 1; i-- {
		require.NotEqual(t, cancelIdx, <-results)
	}
}

func TestStop(t *testing.T) {
	n := 100
	results := make(chan int, n)
	var g errgroup.Group

	pw := New(func(ctx context.Context, i int) error {
		time.Sleep(2 * time.Millisecond)
		results <- i
		return nil
	}, 4)
	pw.Pause() // ensure predictable output

	for i := n; i > 0; i-- {
		priority := i
		g.Go(func() error {
			return pw.Process(context.Background(), priority, priority)
		})
	}

	// let it get started, then unpause
	time.Sleep(10 * time.Millisecond)
	pw.Unpause()

	g.Go(func() error {
		pw.Stop()
		return nil
	})

	err := g.Wait()
	require.Equal(t, ErrStopped, err)

	require.Greater(t, len(results), 0)
	require.Less(t, len(results), n)
}
