package runtime

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestQueryLimiter(t *testing.T) {
	ctx := t.Context()
	const short = 50 * time.Millisecond

	t.Run("instance limit", func(t *testing.T) {
		l := newQueryLimiter()
		r1, err := l.acquire(ctx, "i1", "u1", 2, 0, short)
		require.NoError(t, err)
		r2, err := l.acquire(ctx, "i1", "u2", 2, 0, short)
		require.NoError(t, err)

		_, err = l.acquire(ctx, "i1", "u3", 2, 0, short)
		require.ErrorIs(t, err, ErrTooManyConcurrentQueries)
		require.ErrorContains(t, err, "per project")

		// Other instances are not affected.
		r3, err := l.acquire(ctx, "i2", "u3", 2, 0, short)
		require.NoError(t, err)

		r1()
		r4, err := l.acquire(ctx, "i1", "u3", 2, 0, short)
		require.NoError(t, err)

		r2()
		r3()
		r4()
		require.Empty(t, l.instances)
	})

	t.Run("user limit", func(t *testing.T) {
		l := newQueryLimiter()
		r1, err := l.acquire(ctx, "i1", "u1", 4, 1, short)
		require.NoError(t, err)

		_, err = l.acquire(ctx, "i1", "u1", 4, 1, short)
		require.ErrorIs(t, err, ErrTooManyConcurrentQueries)
		require.ErrorContains(t, err, "per user")

		// Other users and callers without a user ID still get slots.
		r2, err := l.acquire(ctx, "i1", "u2", 4, 1, short)
		require.NoError(t, err)
		r3, err := l.acquire(ctx, "i1", "", 4, 1, short)
		require.NoError(t, err)
		r4, err := l.acquire(ctx, "i1", "", 4, 1, short)
		require.NoError(t, err)

		r1()
		r2()
		r3()
		r4()
		require.Empty(t, l.instances)
	})

	t.Run("no limits", func(t *testing.T) {
		l := newQueryLimiter()
		var releases []func()
		for range 10 {
			r, err := l.acquire(ctx, "i1", "u1", 0, 0, short)
			require.NoError(t, err)
			releases = append(releases, r)
		}
		for _, r := range releases {
			r()
		}
		require.Empty(t, l.instances)
	})

	t.Run("release wakes a waiter", func(t *testing.T) {
		l := newQueryLimiter()
		r1, err := l.acquire(ctx, "i1", "u1", 1, 0, short)
		require.NoError(t, err)

		acquired := make(chan error, 1)
		go func() {
			r, err := l.acquire(ctx, "i1", "u2", 1, 0, 10*time.Second)
			if err == nil {
				r()
			}
			acquired <- err
		}()

		select {
		case err := <-acquired:
			t.Fatalf("waiter acquired a slot before it was released: %v", err)
		case <-time.After(short):
		}

		r1()
		select {
		case err := <-acquired:
			require.NoError(t, err)
		case <-time.After(5 * time.Second):
			t.Fatal("waiter did not acquire the released slot")
		}
		require.Empty(t, l.instances)
	})

	t.Run("release is idempotent", func(t *testing.T) {
		l := newQueryLimiter()
		r1, err := l.acquire(ctx, "i1", "u1", 1, 1, short)
		require.NoError(t, err)
		r2, err := l.acquire(ctx, "i2", "u1", 1, 1, short)
		require.NoError(t, err)

		r1()
		r1()
		require.NotContains(t, l.instances, "i1")
		require.Equal(t, 1, l.instances["i2"].total)

		r2()
		require.Empty(t, l.instances)
	})

	t.Run("context cancellation", func(t *testing.T) {
		l := newQueryLimiter()
		r1, err := l.acquire(ctx, "i1", "u1", 1, 0, short)
		require.NoError(t, err)
		defer r1()

		cctx, cancel := context.WithCancel(ctx)
		cancel()
		_, err = l.acquire(cctx, "i1", "u2", 1, 0, time.Minute)
		require.ErrorIs(t, err, context.Canceled)
		require.NotErrorIs(t, err, ErrTooManyConcurrentQueries)
	})
}
