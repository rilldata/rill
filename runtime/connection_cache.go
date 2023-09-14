package runtime

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/golang-lru/simplelru"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

var errConnectionCacheClosed = errors.New("connectionCache: closed")

const migrateTimeout = 30 * time.Second

// connectionCache is a thread-safe cache for open connections.
// Connections should preferably be opened only via the connection cache.
//
// TODO: It opens connections async, but it will close them sync when evicted. If a handle's close hangs, this can block the cache.
// We should move the closing to the background. However, it must then handle the case of trying to re-open a connection that's currently closing in the background.
type connectionCache struct {
	size             int
	runtime          *Runtime
	logger           *zap.Logger
	activity         activity.Client
	closed           bool
	migrateCtx       context.Context    // ctx used for connection migrations
	migrateCtxCancel context.CancelFunc // cancel all running migrations
	lock             sync.Mutex
	acquired         map[string]*connWithRef // items with non-zero references (in use) which should not be evicted
	lru              *simplelru.LRU          // items with no references (opened, but not in use) ready for eviction
}

type connWithRef struct {
	handle drivers.Handle
	err    error
	refs   int
	ready  chan struct{}
}

func newConnectionCache(size int, logger *zap.Logger, rt *Runtime, ac activity.Client) *connectionCache {
	// LRU cache that closes evicted connections
	lru, err := simplelru.NewLRU(size, func(key interface{}, value interface{}) {
		// Skip if the conn has refs, since the callback also gets called when transferring to acquired cache
		conn := value.(*connWithRef)
		if conn.refs != 0 {
			return
		}
		if conn.handle != nil {
			if err := conn.handle.Close(); err != nil {
				logger.Error("failed closing cached connection", zap.String("key", key.(string)), zap.Error(err))
			}
		}
	})
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &connectionCache{
		size:             size,
		runtime:          rt,
		logger:           logger,
		activity:         ac,
		migrateCtx:       ctx,
		migrateCtxCancel: cancel,
		acquired:         make(map[string]*connWithRef),
		lru:              lru,
	}
}

func (c *connectionCache) Close() error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.closed {
		return errConnectionCacheClosed
	}
	c.closed = true

	// Cancel currently running migrations
	c.migrateCtxCancel()

	var firstErr error
	for _, key := range c.lru.Keys() {
		val, ok := c.lru.Get(key)
		if !ok {
			continue
		}
		conn := val.(*connWithRef)
		if conn.handle == nil {
			continue
		}
		err := conn.handle.Close()
		if err != nil {
			c.logger.Error("failed closing cached connection", zap.Error(err))
			if firstErr == nil {
				firstErr = err
			}
		}
	}

	for _, value := range c.acquired {
		if value.handle == nil {
			continue
		}
		err := value.handle.Close()
		if err != nil {
			c.logger.Error("failed closing cached connection", zap.Error(err))
			if firstErr == nil {
				firstErr = err
			}
		}
	}

	return firstErr
}

func (c *connectionCache) get(ctx context.Context, instanceID, driver string, config map[string]any, shared bool) (drivers.Handle, func(), error) {
	var key string
	if shared {
		// not using instanceID to ensure all instances share the same handle
		key = driver + generateKey(config)
	} else {
		key = instanceID + driver + generateKey(config)
	}

	c.lock.Lock()
	if c.closed {
		c.lock.Unlock()
		return nil, nil, errConnectionCacheClosed
	}

	// Get conn from caches
	conn, ok := c.acquired[key]
	if ok {
		conn.refs++
	} else {
		var val any
		val, ok = c.lru.Get(key)
		if ok {
			// Conn was found in LRU - move to acquired cache
			conn = val.(*connWithRef)
			conn.refs++ // NOTE: Must increment before call to c.lru.remove to avoid closing the conn
			c.lru.Remove(key)
			c.acquired[key] = conn
		}
	}

	// Cached conn not found, open a new one
	if !ok {
		conn = &connWithRef{
			refs:  1, // Since refs is assumed to already have been incremented when checking conn.ready
			ready: make(chan struct{}),
		}
		c.acquired[key] = conn

		if len(c.acquired)+c.lru.Len() > c.size {
			c.logger.Warn("number of connections acquired and in LRU exceed total configured size", zap.Int("acquired", len(c.acquired)), zap.Int("lru", c.lru.Len()))
		}

		// Open and migrate the connection in a separate goroutine (outside lock).
		// Incrementing ref and releasing the conn for this operation separately to cover the case where all waiting goroutines are cancelled before the migration completes.
		conn.refs++
		go func() {
			handle, err := c.openAndMigrate(c.migrateCtx, instanceID, driver, shared, config)
			c.lock.Lock()
			conn.handle = handle
			conn.err = err
			c.releaseConn(key, conn)
			wasClosed := c.closed
			c.lock.Unlock()
			close(conn.ready)

			// The cache might have been closed while the connection was being opened.
			// Since we acquired the lock, the close will have already been completed, so we need to close the connection here.
			if wasClosed && handle != nil {
				_ = handle.Close()
			}
		}()
	}

	// We can now release the lock and wait for the connection to be ready (it might already be)
	c.lock.Unlock()

	// Wait for connection to be ready or context to be cancelled
	var err error
	select {
	case <-conn.ready:
		err = conn.err
	case <-ctx.Done():
		err = ctx.Err() // Will always be non-nil, ensuring releaseConn is called
	}
	if err != nil {
		c.lock.Lock()
		c.releaseConn(key, conn)
		c.lock.Unlock()
		return nil, nil, err
	}

	release := func() {
		c.lock.Lock()
		c.releaseConn(key, conn)
		c.lock.Unlock()
	}

	return conn.handle, release, nil
}

func (c *connectionCache) releaseConn(key string, conn *connWithRef) {
	conn.refs--
	if conn.refs == 0 {
		// No longer referenced. Move from acquired to LRU.
		if !c.closed {
			delete(c.acquired, key)
			c.lru.Add(key, conn)
		}
	}
}

func (c *connectionCache) openAndMigrate(ctx context.Context, instanceID, driver string, shared bool, config map[string]any) (drivers.Handle, error) {
	logger := c.logger
	if instanceID != "default" {
		logger = c.logger.With(zap.String("instance_id", instanceID), zap.String("driver", driver))
	}

	ctx, cancel := context.WithTimeout(ctx, migrateTimeout)
	defer cancel()

	activityClient := c.activity
	if !shared {
		inst, err := c.runtime.FindInstance(ctx, instanceID)
		if err != nil {
			return nil, err
		}

		activityDims := instanceAnnotationsToAttribs(inst)
		if activityClient != nil {
			activityClient = activityClient.With(activityDims...)
		}
	}

	handle, err := drivers.Open(driver, config, shared, activityClient, logger)
	if err != nil {
		return nil, err
	}

	err = handle.Migrate(ctx)
	if err != nil {
		handle.Close()
		return nil, err
	}
	return handle, nil
}

// evict removes the connection from cache and closes the connection.
func (c *connectionCache) evict(ctx context.Context, instanceID, driver string, config map[string]any) bool {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.closed {
		return false
	}

	key := instanceID + driver + generateKey(config)
	conn, ok := c.lru.Get(key)
	if !ok {
		conn, ok = c.acquired[key]
	}
	if ok {
		conn := conn.(*connWithRef)
		if conn.handle != nil {
			err := conn.handle.Close()
			if err != nil {
				c.logger.Error("connection cache: failed to close cached connection", zap.Error(err), zap.String("instance", instanceID), observability.ZapCtx(ctx))
			}
			conn.handle = nil
			conn.err = fmt.Errorf("connection evicted") // Defensive, should never happen
		}
		c.lru.Remove(key)
		delete(c.acquired, key)
	}

	return ok
}

func generateKey(m map[string]any) string {
	sb := strings.Builder{}
	keys := maps.Keys(m)
	slices.Sort(keys)
	for _, key := range keys {
		sb.WriteString(key)
		sb.WriteString(":")
		sb.WriteString(fmt.Sprint(m[key]))
		sb.WriteString(" ")
	}
	return sb.String()
}
