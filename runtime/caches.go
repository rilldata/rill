package runtime

import (
	"context"
	"errors"
	"fmt"
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
	}
	return ok
}

type catalogCache struct {
	cache map[string]*catalog.Service
	lock  sync.Mutex
}

func newCatalogCache() *catalogCache {
	return &catalogCache{
		cache: make(map[string]*catalog.Service),
	}
}

func (c *catalogCache) get(ctx context.Context, rt *Runtime, instID string) (*catalog.Service, error) {
	// TODO 1: opening a driver shouldn't take too long but we should still have an instance specific lock
	// TODO 2: This is a cache on a cache, which may lead to undefined behavior
	// TODO 3: Use LRU and not a map

	c.lock.Lock()
	defer c.lock.Unlock()

	key := instID

	service, ok := c.cache[key]
	if ok {
		return service, nil
	}

	registry, _ := rt.metastore.RegistryStore()
	inst, err := registry.FindInstance(ctx, instID)
	if err != nil {
		return nil, err
	}

	olapConn, err := rt.connCache.get(ctx, instID, inst.OLAPDriver, inst.OLAPDSN)
	if err != nil {
		return nil, err
	}
	olap, _ := olapConn.OLAPStore()

	var catalogStore drivers.CatalogStore
	if inst.EmbedCatalog {
		conn, err := rt.connCache.get(ctx, inst.ID, inst.OLAPDriver, inst.OLAPDSN)
		if err != nil {
			return nil, err
		}

		store, ok := conn.CatalogStore()
		if !ok {
			return nil, fmt.Errorf("instance cannot embed catalog")
		}

		catalogStore = store
	} else {
		store, ok := rt.metastore.CatalogStore()
		if !ok {
			return nil, fmt.Errorf("metastore cannot serve as catalog")
		}
		catalogStore = store
	}

	repoConn, err := rt.connCache.get(ctx, instID, inst.RepoDriver, inst.RepoDSN)
	if err != nil {
		return nil, err
	}
	repoStore, _ := repoConn.RepoStore()

	service = catalog.NewService(catalogStore, repoStore, olap, registry, instID, rt.logger)
	c.cache[key] = service
	return service, nil
}

func (c *catalogCache) evict(ctx context.Context, instID string) {
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
