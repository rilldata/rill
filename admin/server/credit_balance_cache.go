package server

import (
	"context"
	"fmt"
	"time"

	lru "github.com/hashicorp/golang-lru"
)

const (
	// creditBalanceCacheSize bounds the number of orgs whose balance is kept in memory.
	// Only orgs on the credit trial (or that just ran out of trial credits) query their balance.
	creditBalanceCacheSize = 500
	// creditBalanceTTL is how long a cached balance is served to callers that allow cached values,
	// without querying the billing provider.
	creditBalanceTTL = 15 * time.Minute
	// creditBalanceFetchTimeout limits how long callers that allow cached values wait for the billing provider.
	creditBalanceFetchTimeout = 2 * time.Second
)

// creditBalanceCache caches billing credit balances per org in memory.
// It lets frequently loaded UI surfaces, such as the trial banner, avoid querying the billing provider on every page load.
// Each admin server has its own cache, so the cache is best-effort and not shared across replicas.
type creditBalanceCache struct {
	cache *lru.Cache // org ID -> creditBalanceEntry
}

type creditBalanceEntry struct {
	balance   float64
	fetchedOn time.Time
}

func newCreditBalanceCache() (*creditBalanceCache, error) {
	cache, err := lru.New(creditBalanceCacheSize)
	if err != nil {
		return nil, fmt.Errorf("error creating credit balance cache: %w", err)
	}
	return &creditBalanceCache{cache: cache}, nil
}

// get returns the credit balance for an org, calling fetch to query the billing provider when needed.
// If allowCached is true, it returns a balance cached within creditBalanceTTL without calling fetch,
// and otherwise calls fetch with a timeout of creditBalanceFetchTimeout.
// If allowCached is false, it always calls fetch.
// Every successful fetch updates the cache.
func (c *creditBalanceCache) get(ctx context.Context, orgID string, allowCached bool, fetch func(ctx context.Context) (float64, error)) (float64, error) {
	if allowCached {
		if v, ok := c.cache.Get(orgID); ok {
			e := v.(creditBalanceEntry)
			if time.Since(e.fetchedOn) < creditBalanceTTL {
				return e.balance, nil
			}
		}

		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, creditBalanceFetchTimeout)
		defer cancel()
	}

	balance, err := fetch(ctx)
	if err != nil {
		return 0, err
	}
	c.cache.Add(orgID, creditBalanceEntry{balance: balance, fetchedOn: time.Now()})
	return balance, nil
}

// invalidate removes the cached balance for an org, e.g. after credits are granted.
func (c *creditBalanceCache) invalidate(orgID string) {
	c.cache.Remove(orgID)
}
