package runtime

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"sync"

	"github.com/dgraph-io/ristretto"
	"github.com/hashicorp/golang-lru/simplelru"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/singleflight"
	"github.com/rilldata/rill/runtime/services/catalog"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

var errConnectionCacheClosed = errors.New("connectionCache: closed")

// init registers the protobuf types with gob so they can be encoded.
func init() {
	gob.Register(structpb.Value_BoolValue{})
	gob.Register(structpb.Value_NumberValue{})
	gob.Register(structpb.Value_StringValue{})
	gob.Register(structpb.Value_NullValue{})
	gob.Register(structpb.Value_ListValue{})
	gob.Register(structpb.Value_StructValue{})
}

// cache for instance specific connections only
// all instance specific connections should be opened via connection cache only
type connectionCache struct {
	cache  *simplelru.LRU
	lock   sync.Mutex
	closed bool
	logger *zap.Logger
}

func newConnectionCache(size int, logger *zap.Logger) *connectionCache {
	cache, err := simplelru.NewLRU(size, func(key interface{}, value interface{}) {
		// close the evicted connection
		if err := value.(drivers.Connection).Close(); err != nil {
			logger.Error("failed closing cached connection for ", zap.String("key", key.(string)), zap.Error(err))
		}
	})
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
	cache *simplelru.LRU
	lock  sync.Mutex
}

func newMigrationMetaCache(size int) *migrationMetaCache {
	cache, err := simplelru.NewLRU(size, nil)
	if err != nil {
		panic(err)
	}

	return &migrationMetaCache{cache: cache}
}

func (c *migrationMetaCache) get(instID string) *catalog.MigrationMeta {
	c.lock.Lock()
	defer c.lock.Unlock()
	if val, ok := c.cache.Get(instID); ok {
		return val.(*catalog.MigrationMeta)
	}

	meta := catalog.NewMigrationMeta()
	c.cache.Add(instID, meta)
	return meta
}

func (c *migrationMetaCache) evict(ctx context.Context, instID string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.cache.Remove(instID)
}

type queryCache struct {
	cache *ristretto.Cache
	group *singleflight.Group
}

func newQueryCache(sizeInBytes int64) *queryCache {
	if sizeInBytes <= 0 {
		panic(fmt.Sprintf("invalid cache size should be greater than 0 : %v", sizeInBytes))
	}
	cache, err := ristretto.NewCache(&ristretto.Config{
		// Use 5% of cache memory for storing counters. Each counter takes roughly 3 bytes.
		// Recommended value is 10x the number of items in cache when full.
		// Tune this again based on metrics.
		NumCounters: int64(float64(sizeInBytes) * 0.05 / 3),
		MaxCost:     int64(float64(sizeInBytes) * 0.95),
		BufferItems: 64,
		Metrics:     false,
		Cost: func(val interface{}) int64 {
			if val == nil {
				return 0
			}
			b := new(bytes.Buffer)
			if err := gob.NewEncoder(b).Encode(val); err != nil {
				panic(err)
			}
			return int64(b.Len())
		},
	})
	if err != nil {
		panic(err)
	}
	return &queryCache{cache: cache}
}

// getOrLoad gets the key from cache if present. If absent, it looks up the key using the loadFn and puts it into cache before returning value.
// NOTE:: Due to limitation of the underlying caching library, key can only be one of int(signed/unsgined),string or byte array.
func (c *queryCache) getOrLoad(key any, loadFn func() (any, error)) (any, bool, error) {
	if val, ok := c.cache.Get(key); ok {
		return val, true, nil
	}

	val, err := c.group.Do(key, loadFn)
	if err != nil {
		return nil, false, err
	}

	c.cache.Set(key, val, 0)
	return val, false, nil
}

// nolint:unused // use in tests
func (c *queryCache) add(key, val any) bool {
	return c.cache.Set(key, val, 0)
}

// nolint:unused // use in tests
func (c *queryCache) get(key any) (any, bool) {
	return c.cache.Get(key)
}
