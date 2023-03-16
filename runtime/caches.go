package runtime

import (
	"context"
	"errors"
	"sync"

	lru "github.com/hashicorp/golang-lru"
	"github.com/hashicorp/golang-lru/simplelru"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/services/catalog"
	"go.uber.org/zap"
)

var errConnectionCacheClosed = errors.New("connectionCache: closed")

type connectionCache struct {
	cache  *simplelru.LRU
	lock   sync.Mutex
	closed bool
	logger *zap.Logger
}

func newConnectionCache(size int, logger *zap.Logger) *connectionCache {
	cache, err := simplelru.NewLRU(size, nil)
	if err != nil {
		panic(err)
	}
	return &connectionCache{cache: cache, logger: logger}
}

func (c *connectionCache) Close() error {
	c.lock.Lock()
	if c.closed {
		c.lock.Unlock()
		return errConnectionCacheClosed
	}
	c.closed = true
	c.lock.Unlock()

	var firstErr error
	for _, key := range c.cache.Keys() {
		val, _ := c.cache.Get(key)
		err := val.(drivers.Connection).Close()
		if err != nil {
			c.logger.Error("failed closing cached connection", zap.Error(err))
			if firstErr == nil {
				firstErr = err
			}
		}
	}

	return firstErr
}

func (c *connectionCache) get(ctx context.Context, instanceID, driver, dsn string) (drivers.Connection, error) {
	// TODO: This locks for all instances for the duration of Open and Migrate.
	// Adapt to lock only on the lookup, and then on the individual instance's Open and Migrate.

	c.lock.Lock()
	defer c.lock.Unlock()

	if c.closed {
		return nil, errConnectionCacheClosed
	}

	key := instanceID + driver + dsn
	val, ok := c.cache.Get(key)
	if !ok {
		conn, err := drivers.Open(driver, dsn, c.logger)
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

// evict removes the connection from cache and closes the connection
func (c *connectionCache) evict(ctx context.Context, instanceID, driver, dsn string) bool {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.closed {
		return false
	}

	key := instanceID + driver + dsn
	conn, ok := c.cache.Get(key)
	if ok {
		// closing this would mean that any running query might also fail
		conn.(drivers.Connection).Close()
		c.cache.Remove(key)
	}
	return ok
}

type migrationMetaCache struct {
	cache map[string]*catalog.MigrationMeta
	lock  sync.Mutex
}

func newMigrationMetadataCache() *migrationMetaCache {
	return &migrationMetaCache{
		cache: make(map[string]*catalog.MigrationMeta),
	}
}

func (c *migrationMetaCache) get(instID string) *catalog.MigrationMeta {
	// TODO : Use LRU and not a map
	c.lock.Lock()
	defer c.lock.Unlock()

	if val, ok := c.cache[instID]; ok {
		return val
	}

	meta := catalog.NewMigrationMeta()
	c.cache[instID] = meta
	return meta
}

func (c *migrationMetaCache) evict(ctx context.Context, instID string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	delete(c.cache, instID)
}

type queryCache struct {
	cache *lru.Cache
}

func newQueryCache(size int) *queryCache {
	cache, err := lru.New(size)
	if err != nil {
		panic(err)
	}
	return &queryCache{cache: cache}
}

func (c *queryCache) get(key queryCacheKey) (any, bool) {
	return c.cache.Get(key)
}

func (c *queryCache) add(key queryCacheKey, value any) bool {
	return c.cache.Add(key, value)
}
