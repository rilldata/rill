package runtime

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/golang-lru/simplelru"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
)

var errConnectionCacheClosed = errors.New("connectionCache: closed")

var errConnectionClosed = errors.New("connectionCache: connection closed")

const migrateTimeout = 2 * time.Minute

const hangingTimeout = 5 * time.Minute

// connectionCache is a thread-safe cache for open connections.
// Connections should preferably be opened only via the connection cache.
type connectionCache struct {
	size      int
	runtime   *Runtime
	logger    *zap.Logger
	activity  activity.Client
	closed    bool
	ctx       context.Context    // ctx used for background tasks
	ctxCancel context.CancelFunc // cancel background ctx
	lock      sync.Mutex
	entries   map[string]*connectionCacheEntry
	lru       *simplelru.LRU // entries with no references (opened, but not in use) ready for eviction
}

type connectionCacheEntry struct {
	instanceID   string
	refs         int
	working      bool
	workingCh    chan struct{}
	workingSince time.Time
	handle       drivers.Handle
	err          error
	closed       bool
}

func newConnectionCache(size int, logger *zap.Logger, rt *Runtime, ac activity.Client) *connectionCache {
	ctx, cancel := context.WithCancel(context.Background())
	c := &connectionCache{
		size:      size,
		runtime:   rt,
		logger:    logger,
		activity:  ac,
		ctx:       ctx,
		ctxCancel: cancel,
		entries:   make(map[string]*connectionCacheEntry),
	}

	var err error
	c.lru, err = simplelru.NewLRU(size, c.lruEvictionHandler)
	if err != nil {
		panic(err)
	}

	go c.periodicallyCheckHangingConnections()

	return c
}

// Close closes all connections in the cache.
// While not strictly necessary, it's probably best to call this after all connections have been released.
func (c *connectionCache) Close() error {
	c.lock.Lock()

	// Set closed
	if c.closed {
		c.lock.Unlock()
		return errConnectionCacheClosed
	}
	c.closed = true

	// Cancel currently running migrations
	c.ctxCancel()

	// Start closing all connections (will close in the background)
	for key, entry := range c.entries {
		c.closeEntry(key, entry)
	}

	// Clear the LRU - might not be needed, but just to be sure
	c.lru.Purge()

	// Unlock to allow entries to close and remove themselves from c.entries in the background
	c.lock.Unlock()

	// Wait for c.entries to become empty
	for {
		c.lock.Lock()
		var anyEntry *connectionCacheEntry
		for _, e := range c.entries {
			anyEntry = e
			break
		}
		c.lock.Unlock()

		if anyEntry == nil {
			// c.entries is empty, we can return
			break
		}

		<-anyEntry.workingCh
	}

	return nil
}

// Get opens and caches a connection.
// The caller should call the returned release function when done with the connection.
func (c *connectionCache) Get(ctx context.Context, instanceID, driver string, config map[string]any, shared bool) (drivers.Handle, func(), error) {
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

	// Get or create conn
	entry, ok := c.entries[key]
	if !ok {
		// Cached conn not found, open a new one
		entry = &connectionCacheEntry{instanceID: instanceID}
		c.entries[key] = entry

		c.openEntry(key, entry, driver, shared, config)

		if len(c.entries) >= 2*c.size {
			c.logger.Warn("connection cache: the number of open connections exceeds the cache size by more than 2x", zap.Int("entries", len(c.entries)))
		}
	}

	// Acquire the entry
	c.acquireEntry(key, entry)

	// We can now release the lock and wait for the connection to be ready (it might already be)
	c.lock.Unlock()

	// Wait for connection to be ready or context to be cancelled
	var err error
	stop := false
	for !stop {
		select {
		case <-entry.workingCh:
			c.lock.Lock()

			// The entry was closed right after being opened, we must loop to check c.workingCh again.
			if entry.working {
				c.lock.Unlock()
				continue
			}

			// We acquired the entry as it was closing, let's reopen it.
			if entry.closed {
				c.openEntry(key, entry, driver, shared, config)
				c.lock.Unlock()
				continue
			}

			stop = true
		case <-ctx.Done():
			c.lock.Lock()
			err = ctx.Err() // Will always be non-nil, ensuring releaseEntry is called
			stop = true
		}
	}

	// We've got the lock now and know entry.working is false
	defer c.lock.Unlock()

	if err == nil {
		err = entry.err
	}

	if err != nil {
		c.releaseEntry(key, entry)
		return nil, nil, err
	}

	release := func() {
		c.lock.Lock()
		c.releaseEntry(key, entry)
		c.lock.Unlock()
	}

	return entry.handle, release, nil
}

// EvictAll closes all connections for an instance.
func (c *connectionCache) EvictAll(ctx context.Context, instanceID string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.closed {
		return
	}

	for key, entry := range c.entries {
		if entry.instanceID != instanceID {
			continue
		}

		c.closeEntry(key, entry)
	}
}

// acquireEntry increments an entry's refs and moves it out of the LRU if it's there.
// It should be called when holding the lock.
func (c *connectionCache) acquireEntry(key string, entry *connectionCacheEntry) {
	entry.refs++
	if entry.refs == 1 {
		// NOTE: lru.Remove is safe even if it's not in the LRU (should only happen if the entry is acquired for the first time)
		_ = c.lru.Remove(key)
	}
}

// releaseEntry decrements an entry's refs and moves it to the LRU if nothing references it.
// It should be called when holding the lock.
func (c *connectionCache) releaseEntry(key string, entry *connectionCacheEntry) {
	entry.refs--
	if entry.refs == 0 {
		// No longer referenced. Move to LRU unless conn and/or cache is closed.
		delete(c.entries, key)
		if !c.closed && !entry.closed {
			c.lru.Add(key, entry)
		}
	}
}

// lruEvictionHandler is called by the LRU when evicting an entry.
// Note that the LRU only holds entries with refs == 0 (unless the entry is currently being moved to the acquired cache).
// Note also that this handler is called sync by the LRU, i.e. c.lock will be held.
func (c *connectionCache) lruEvictionHandler(key, value interface{}) {
	entry := value.(*connectionCacheEntry)

	// The callback also gets called when removing from LRU during acquisition.
	// We use conn.refs != 0 to signal that its being acquired and should not be closed.
	if entry.refs != 0 {
		return
	}

	// Close the connection
	c.closeEntry(key.(string), entry)
}

// openEntry opens an entry's connection. It's safe to call for a previously closed entry.
// It's NOT safe to call for an entry that's currently working.
// It should be called when holding the lock (but the actual open and migrate will happen in the background).
func (c *connectionCache) openEntry(key string, entry *connectionCacheEntry, driver string, shared bool, config map[string]any) {
	// Since whatever code that called openEntry may get cancelled/return before the connection is opened, we get our own reference to it.
	c.acquireEntry(key, entry)

	// Reset entry and set working
	entry.working = true
	entry.workingCh = make(chan struct{})
	entry.workingSince = time.Now()
	entry.handle = nil
	entry.err = nil
	entry.closed = false

	// Open in the background
	// NOTE: If closeEntry is called while it's opening, closeEntry will wait for the open to complete, so we don't need to handle that case here.
	go func() {
		handle, err := c.openAndMigrate(c.ctx, entry.instanceID, driver, shared, config)

		c.lock.Lock()
		entry.working = false
		close(entry.workingCh)
		entry.workingSince = time.Time{}
		entry.handle = handle
		entry.err = err
		c.releaseEntry(key, entry)
		c.lock.Unlock()
	}()
}

// closeEntry closes an entry's connection. It's safe to call for an entry that's currently being closed/already closed.
// It should be called when holding the lock (but the actual close will happen in the background).
func (c *connectionCache) closeEntry(key string, entry *connectionCacheEntry) {
	if entry.closed {
		return
	}

	c.acquireEntry(key, entry)

	wasWorking := entry.working
	if !wasWorking {
		entry.working = true
		entry.workingCh = make(chan struct{})
		entry.workingSince = time.Now()
	}

	go func() {
		// If the entry was working when closeEntry was called, wait for it to finish before continuing.
		if wasWorking {
			stop := false
			for !stop {
				<-entry.workingCh
				c.lock.Lock()

				// Bad luck, something else started working on the entry. Loop and wait again.
				if entry.working {
					c.lock.Unlock()
					continue
				}

				// Good luck, something else closed the entry. We're done.
				if entry.closed {
					c.lock.Unlock()
					return
				}

				// Our turn to start working it
				entry.working = true
				entry.workingCh = make(chan struct{})
				entry.workingSince = time.Now()
				c.lock.Unlock()
				stop = true
			}
		}

		// Close the connection
		if entry.handle != nil {
			err := entry.handle.Close()
			if err != nil {
				c.logger.Error("failed closing cached connection", zap.String("key", key), zap.Error(err))
			}
		}

		// Mark closed
		c.lock.Lock()
		entry.working = false
		close(entry.workingCh)
		entry.workingSince = time.Time{}
		entry.handle = nil
		entry.err = errConnectionClosed
		entry.closed = true
		c.releaseEntry(key, entry)
		c.lock.Unlock()
	}()
}

// openAndMigrate opens a connection and migrates it.
func (c *connectionCache) openAndMigrate(ctx context.Context, instanceID, driver string, shared bool, config map[string]any) (drivers.Handle, error) {
	logger := c.logger
	if instanceID != "default" {
		logger = c.logger.With(zap.String("instance_id", instanceID), zap.String("driver", driver))
	}

	ctx, cancel := context.WithTimeout(ctx, migrateTimeout)
	defer cancel()

	activityClient := c.activity
	if !shared {
		inst, err := c.runtime.Instance(ctx, instanceID)
		if err != nil {
			return nil, err
		}

		activityDims := instanceAnnotationsToAttribs(inst)
		if activityClient != nil {
			activityClient = activityClient.With(activityDims...)
		}
	}

	handle, err := drivers.Open(driver, config, shared, activityClient, logger)
	if err == nil && ctx.Err() != nil {
		err = fmt.Errorf("timed out while opening driver %q", driver)
	}
	if err != nil {
		return nil, err
	}

	err = handle.Migrate(ctx)
	if err != nil {
		handle.Close()
		if errors.Is(err, ctx.Err()) {
			err = fmt.Errorf("timed out while migrating driver %q: %w", driver, err)
		}
		return nil, err
	}
	return handle, nil
}

// periodicallyCheckHangingConnections periodically checks for connection opens or closes that have been working for too long.
func (c *connectionCache) periodicallyCheckHangingConnections() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.lock.Lock()
			for key, entry := range c.entries {
				if entry.working && time.Since(entry.workingSince) > hangingTimeout {
					c.logger.Error("connection cache: connection open or close has been working for too long", zap.String("key", key), zap.Duration("duration", time.Since(entry.workingSince)))
				}
			}
			c.lock.Unlock()
		case <-c.ctx.Done():
			return
		}
	}
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
