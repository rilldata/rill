package server

import (
	"context"
	"sync"

	"github.com/hashicorp/golang-lru/simplelru"
	"github.com/rilldata/rill/runtime/drivers"
)

type connectionCache struct {
	cache *simplelru.LRU
	lock  sync.Mutex
}

func newConnectionCache(size int) *connectionCache {
	cache, err := simplelru.NewLRU(size, nil)
	if err != nil {
		panic(err)
	}
	return &connectionCache{cache: cache}
}

func (c *connectionCache) openAndMigrate(ctx context.Context, instanceID string, driver string, dsn string) (drivers.Connection, error) {
	// TODO: This locks for all instances for the duration of Open and Migrate.
	// Adapt to lock only on the lookup, and then on the individual instance's Open and Migrate.

	c.lock.Lock()
	defer c.lock.Unlock()

	key := instanceID
	val, ok := c.cache.Get(key)
	if !ok {
		conn, err := drivers.Open(driver, dsn)
		if err != nil {
			return nil, err
		}

		err = conn.Migrate(ctx)
		if err != nil {
			return nil, err
		}

		c.cache.Add(key, conn)
		return conn, nil
	}

	return val.(drivers.Connection), nil
}
