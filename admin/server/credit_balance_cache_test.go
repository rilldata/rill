package server

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
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
		{name: "expired fresh cache fetches", cachedAge: 20 * time.Minute, allowCached: true, wantBalance: 200, wantFetch: true},
		{name: "empty cache fetches", allowCached: true, wantBalance: 200, wantFetch: true},
		{name: "fetch failure falls back to cache under an hour old", cachedAge: 30 * time.Minute, allowCached: true, fetchErr: errFetch, wantBalance: 100, wantFetch: true},
		{name: "fetch failure with cache over an hour old errors", cachedAge: 2 * time.Hour, allowCached: true, fetchErr: errFetch, wantErr: true, wantFetch: true},
		{name: "fetch failure with empty cache errors", allowCached: true, fetchErr: errFetch, wantErr: true, wantFetch: true},
		{name: "cache is ignored when not allowed", cachedAge: time.Minute, allowCached: false, wantBalance: 200, wantFetch: true},
		{name: "fetch failure does not fall back when cache not allowed", cachedAge: time.Minute, allowCached: false, fetchErr: errFetch, wantErr: true, wantFetch: true},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			c, err := newCreditBalanceCache(zap.NewNop())
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
	c, err := newCreditBalanceCache(zap.NewNop())
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
	c, err := newCreditBalanceCache(zap.NewNop())
	require.NoError(t, err)
	c.cache.Add("org", creditBalanceEntry{balance: 100, fetchedOn: time.Now().Add(-30 * time.Minute)})

	// A slow fetch is cut off by the timeout and the cached balance is returned.
	balance, err := c.get(t.Context(), "org", true, func(ctx context.Context) (float64, error) {
		<-ctx.Done()
		return 0, ctx.Err()
	})
	require.NoError(t, err)
	require.Equal(t, 100.0, balance)
}
