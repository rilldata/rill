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
	results := make(chan int)
	var g errgroup.Group
	concurrency := 4

	pw := New(func(ctx context.Context, i int) error {
		results <- i
		return nil
	}, concurrency)
	pw.Pause() // ensure predictable output

	for i := n; i > 0; i-- {
		priority := i
		g.Go(func() error {
			return pw.Process(context.Background(), priority, priority)
		})
	}

	// We need to check two things -
	// 1. That the concurrency limit is respected and we are running as many jobs as concurrency
	// 2. That the priority is respected
	// Since we can't guarantee the order of the results, we are checking that sum of every n results is equal to sum
	// of topN priorities, where n is the concurrency limit
	// At this point all jobs are already present the priority worker
	// After getting n results we are unPausing and letting next n jobs to run and pausing again unless we get n results
	actual := 0
	expected := 0
	res := 0
	for i := n; i > 0; i-- {
		if i%concurrency == 0 {
			require.Equal(t, expected, actual)
			actual = 0
			expected = 0
			pw.Unpause()
			// give time for topN jobs to start running,
			time.Sleep(100 * time.Millisecond)
			pw.Pause()
			require.Equal(t, concurrency, len(pw.runningJobs))
		}
		res = <-results
		actual += res
		expected += i
	}
	time.Sleep(100 * time.Millisecond)
	require.Equal(t, 0, len(pw.runningJobs))

	err := g.Wait()
	require.NoError(t, err)
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
