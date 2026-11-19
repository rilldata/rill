package server

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCreditBalanceCache(t *testing.T) {
	errFetch := errors.New("billing provider unavailable")

	tt := []struct {
		name        string
		cachedAge   time.Duration // zero means nothing is cached
		allowCached bool
		fetchErr    error
		wantBalance float64
		wantErr     bool
		wantFetch   bool
	}{
		{name: "fresh cache is served without fetching", cachedAge: time.Minute, allowCached: true, wantBalance: 100},
		{name: "expired cache fetches", cachedAge: 20 * time.Minute, allowCached: true, wantBalance: 200, wantFetch: true},
		{name: "empty cache fetches", allowCached: true, wantBalance: 200, wantFetch: true},
		{name: "fetch failure with expired cache errors", cachedAge: 20 * time.Minute, allowCached: true, fetchErr: errFetch, wantErr: true, wantFetch: true},
		{name: "cache is ignored when not allowed", cachedAge: time.Minute, allowCached: false, wantBalance: 200, wantFetch: true},
		{name: "fetch failure errors when cache not allowed", cachedAge: time.Minute, allowCached: false, fetchErr: errFetch, wantErr: true, wantFetch: true},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			c, err := newCreditBalanceCache()
			require.NoError(t, err)
			if tc.cachedAge != 0 {
				c.cache.Add("org", creditBalanceEntry{balance: 100, fetchedOn: time.Now().Add(-tc.cachedAge)})
			}

			fetched := false
			balance, err := c.get(t.Context(), "org", tc.allowCached, func(ctx context.Context) (float64, error) {
				fetched = true
				return 200, tc.fetchErr
			})

			require.Equal(t, tc.wantFetch, fetched)
			if tc.wantErr {
				require.ErrorIs(t, err, errFetch)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.wantBalance, balance)
		})
	}
}

func TestCreditBalanceCacheUpdatesAndInvalidates(t *testing.T) {
	c, err := newCreditBalanceCache()
	require.NoError(t, err)

	// A fetch that bypasses the cache still updates it for later cached reads.
	_, err = c.get(t.Context(), "org", false, func(ctx context.Context) (float64, error) { return 150, nil })
	require.NoError(t, err)
	balance, err := c.get(t.Context(), "org", true, func(ctx context.Context) (float64, error) {
		t.Fatal("unexpected fetch")
		return 0, nil
	})
	require.NoError(t, err)
	require.Equal(t, 150.0, balance)

	// After invalidation, a cached read fetches again.
	c.invalidate("org")
	balance, err = c.get(t.Context(), "org", true, func(ctx context.Context) (float64, error) { return 300, nil })
	require.NoError(t, err)
	require.Equal(t, 300.0, balance)
}

func TestCreditBalanceCacheFetchTimeout(t *testing.T) {
	c, err := newCreditBalanceCache()
	require.NoError(t, err)

	// A slow fetch is cut off by the timeout when cached values are allowed.
	_, err = c.get(t.Context(), "org", true, func(ctx context.Context) (float64, error) {
		<-ctx.Done()
		return 0, ctx.Err()
	})
	require.ErrorIs(t, err, context.DeadlineExceeded)
}
