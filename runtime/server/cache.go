package server

import (
	"context"
	"sync"
	"time"

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

type catalogCache struct {
	cache *simplelru.LRU
	lock  sync.Mutex
	ttl   time.Duration
}

type catalogCacheEntry struct {
	objects   []*drivers.CatalogObject
	refreshed time.Time
}

func newCatalogCache(size int, ttl time.Duration) *catalogCache {
	cache, err := simplelru.NewLRU(size, nil)
	if err != nil {
		panic(err)
	}
	return &catalogCache{cache: cache, ttl: ttl}
}

func (c *catalogCache) allObjects(ctx context.Context, instanceID string, catalog drivers.CatalogStore) []*drivers.CatalogObject {
	// TODO: Same locking problem as connectionCache

	c.lock.Lock()
	defer c.lock.Unlock()

	key := instanceID
	val, ok := c.cache.Get(key)
	if !ok {
		objs := catalog.FindObjects(ctx, instanceID, drivers.CatalogObjectTypeUnspecified)
		entry := &catalogCacheEntry{objects: objs, refreshed: time.Now()}
		c.cache.Add(key, entry)
		return entry.objects
	}

	entry := val.(*catalogCacheEntry)
	if time.Since(entry.refreshed) > c.ttl {
		entry.objects = catalog.FindObjects(ctx, instanceID, drivers.CatalogObjectTypeUnspecified)
		entry.refreshed = time.Now()
	}

	return entry.objects
}

func (c *catalogCache) reset(instanceID string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.cache.Remove(instanceID)
}
