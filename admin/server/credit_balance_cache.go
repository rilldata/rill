package server

import (
	"context"
	"fmt"
	"time"

	lru "github.com/hashicorp/golang-lru"
	"github.com/rilldata/rill/runtime/pkg/singleflight"
	"go.uber.org/zap"
)

const (
	// creditBalanceCacheSize bounds the number of orgs whose balance is kept in memory.
	creditBalanceCacheSize = 10000
	// creditBalanceFreshTTL is how long a cached balance is served to callers that allow cached values,
	// without querying the billing provider.
	creditBalanceFreshTTL = 15 * time.Minute
	// creditBalanceStaleTTL is how long a cached balance is used as a fallback for callers that allow cached values,
	// when querying the billing provider fails or times out.
	creditBalanceStaleTTL = time.Hour
	// creditBalanceFetchTimeout limits how long callers that allow cached values wait for the billing provider.
	creditBalanceFetchTimeout = 2 * time.Second
)

// creditBalanceCache caches billing credit balances per org in memory.
// It lets frequently loaded UI surfaces, such as the trial banner, avoid querying the billing provider on every page load.
// Each admin server has its own cache, so the cache is best-effort and not shared across replicas.
type creditBalanceCache struct {
	logger *zap.Logger
	cache  *lru.Cache // org ID -> creditBalanceEntry
	group  singleflight.Group[string, float64]
}

type creditBalanceEntry struct {
	balance   float64
	fetchedOn time.Time
}

func newCreditBalanceCache(logger *zap.Logger) (*creditBalanceCache, error) {
	cache, err := lru.New(creditBalanceCacheSize)
	if err != nil {
		return nil, fmt.Errorf("error creating credit balance cache: %w", err)
	}
	return &creditBalanceCache{logger: logger, cache: cache}, nil
}

// get returns the credit balance for an org, calling fetch to query the billing provider when needed.
// If allowCached is false, it always calls fetch.
// If allowCached is true, it returns a cached balance fetched within creditBalanceFreshTTL without calling fetch;
// otherwise it calls fetch with a timeout and falls back to a cached balance fetched within creditBalanceStaleTTL if fetch fails.
// Every successful fetch updates the cache.
func (c *creditBalanceCache) get(ctx context.Context, orgID string, allowCached bool, fetch func(ctx context.Context) (float64, error)) (float64, error) {
	if !allowCached {
		balance, err := fetch(ctx)
		if err != nil {
			return 0, err
		}
		c.cache.Add(orgID, creditBalanceEntry{balance: balance, fetchedOn: time.Now()})
		return balance, nil
	}

	var cached *creditBalanceEntry
	if v, ok := c.cache.Get(orgID); ok {
		e := v.(creditBalanceEntry)
		cached = &e
	}
	if cached != nil && time.Since(cached.fetchedOn) < creditBalanceFreshTTL {
		return cached.balance, nil
	}

	// Deduplicate concurrent fetches for the same org, e.g. when several pages load at once with an empty cache.
	balance, err := c.group.Do(ctx, orgID, func(ctx context.Context) (float64, error) {
		ctx, cancel := context.WithTimeout(ctx, creditBalanceFetchTimeout)
		defer cancel()
		balance, err := fetch(ctx)
		if err != nil {
			return 0, err
		}
		c.cache.Add(orgID, creditBalanceEntry{balance: balance, fetchedOn: time.Now()})
		return balance, nil
	})
	if err == nil {
		return balance, nil
	}

	if cached != nil && time.Since(cached.fetchedOn) < creditBalanceStaleTTL {
		c.logger.Warn("failed to fetch credit balance; returning cached balance", zap.String("org_id", orgID), zap.Time("fetched_on", cached.fetchedOn), zap.Error(err))
		return cached.balance, nil
	}
	return 0, err
}

// invalidate removes the cached balance for an org, e.g. after credits are granted.
func (c *creditBalanceCache) invalidate(orgID string) {
	c.cache.Remove(orgID)
}
