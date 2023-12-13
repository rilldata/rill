package conncache

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/hashicorp/golang-lru/simplelru"
	"github.com/rilldata/rill/runtime/pkg/singleflight"
)

// Cache is a concurrency-safe cache of stateful connection objects.
// It differs from a connection pool in that it's designed for caching heterogenous connections.
// The cache will at most open one connection per key, even under concurrent access.
// The cache automatically evicts connections that are not in use ("acquired") using a least-recently-used policy.
type Cache interface {
	// Acquire retrieves or opens a connection for the given key. The returned ReleaseFunc must be called when the connection is no longer needed.
	// While a connection is acquired, it will not be closed unless Evict or Close is called.
	// If Acquire is called while the underlying connection is being evicted, it will wait for the close to complete and then open a new connection.
	// If opening the connection fails, Acquire may return the error on subsequent calls without trying to open again until the entry is evicted.
	Acquire(ctx context.Context, cfg any) (Connection, ReleaseFunc, error)

	// EvictWhere closes the connections that match the predicate.
	// It immediately closes the connections, even those that are currently acquired.
	// It returns immediately and does not wait for the connections to finish closing.
	EvictWhere(predicate func(cfg any) bool)

	// Close closes all open connections and prevents new connections from being acquired.
	// It returns when all cached connections have been closed or when the provided ctx is cancelled.
	Close(ctx context.Context) error
}

// Connection is a connection that may be cached.
type Connection interface {
	Close() error
}

// ReleaseFunc is a function that must be called when an acquired connection is no longer needed.
type ReleaseFunc func()

// Options configures a new connection cache.
type Options struct {
	// MaxConnectionsIdle is the maximum number of non-acquired connections that will be kept open.
	MaxConnectionsIdle int
	// OpenTimeout is the maximum amount of time to wait for a connection to open.
	OpenTimeout time.Duration
	// CloseTimeout is the maximum amount of time to wait for a connection to close.
	CloseTimeout time.Duration
	// OpenFunc opens a connection.
	OpenFunc func(ctx context.Context, cfg any) (Connection, error)
	// KeyFunc computes a comparable key for a connection configuration.
	KeyFunc func(cfg any) string
	// HangingFunc is called when an open or close exceeds its timeout and does not respond to context cancellation.
	HangingFunc func(cfg any, open bool)
}

var _ Cache = (*cacheImpl)(nil)

// cacheImpl implements Cache.
// It leverages a singleflight to ensure at most one open/close action runs against a connection at a time.
// It also uses an LRU to pool unused connections and eventually close them.
// The implementation heavily depends on implementation details of singleflight.Group.
// Notably, it will in different places invoke the singleflight with different callbacks for the same key.
// It also relies on singleflight.Do always invoking the callback even if the passed ctx is already cancelled.
type cacheImpl struct {
	opts         Options
	closed       bool
	mu           sync.Mutex
	entries      map[string]*entry
	lru          *simplelru.LRU
	singleflight *singleflight.Group[string, any]
	ctx          context.Context
	cancel       context.CancelFunc
}

type entry struct {
	cfg      any
	refs     int
	status   entryStatus
	since    time.Time
	closedCh chan struct{} // Not set for regular evictions, only used when Cache.Close() is called.
	handle   Connection
	err      error
}

type entryStatus int

const (
	entryStatusUnspecified entryStatus = iota
	entryStatusOpening
	entryStatusOpen // Also used for cases where open errored (i.e. entry.err != nil)
	entryStatusClosing
	entryStatusClosed
)

func New(opts Options) Cache {
	ctx, cancel := context.WithCancel(context.Background())
	c := &cacheImpl{
		opts:         opts,
		entries:      make(map[string]*entry),
		singleflight: &singleflight.Group[string, any]{},
		ctx:          ctx,
		cancel:       cancel,
	}

	var err error
	c.lru, err = simplelru.NewLRU(opts.MaxConnectionsIdle, c.lruEvictionHandler)
	if err != nil {
		panic(err)
	}

	go c.periodicallyCheckHangingConnections()

	return c
}

func (c *cacheImpl) Acquire(ctx context.Context, cfg any) (Connection, ReleaseFunc, error) {
	k := c.opts.KeyFunc(cfg)

	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil, nil, errors.New("conncache: closed")
	}

	e, ok := c.entries[k]
	if !ok {
		e = &entry{cfg: cfg, since: time.Now()}
		c.entries[k] = e
	}

	c.retainEntry(k, e)

	if e.status == entryStatusOpen {
		defer c.mu.Unlock()
		if e.err != nil {
			c.releaseEntry(k, e)
			return nil, nil, e.err
		}
		return e.handle, c.releaseFunc(k, e), nil
	}

	c.mu.Unlock()

	for attempt := 0; attempt < 2; attempt++ {
		_, err := c.singleflight.Do(ctx, k, func(_ context.Context) (any, error) {
			c.mu.Lock()
			if c.closed {
				c.mu.Unlock()
				return nil, errors.New("conncache: closed")
			}
			if e.status == entryStatusOpen {
				c.mu.Unlock()
				return nil, nil
			}
			c.retainEntry(k, e)
			e.status = entryStatusOpening
			e.since = time.Now()
			e.handle = nil
			e.err = nil
			c.mu.Unlock()

			var handle Connection
			var err error
			if c.opts.OpenTimeout == 0 {
				handle, err = c.opts.OpenFunc(ctx, cfg)
			} else {
				ctx, cancel := context.WithTimeout(c.ctx, c.opts.OpenTimeout)
				handle, err = c.opts.OpenFunc(ctx, cfg)
				cancel()
			}

			c.mu.Lock()
			c.releaseEntry(k, e)
			e.status = entryStatusOpen
			e.since = time.Now()
			e.handle = handle
			e.err = err
			c.mu.Unlock()

			return nil, nil
		})
		if err != nil {
			// TODO: could be a caught panic. Should we handle panics?
			return nil, nil, err
		}

		c.mu.Lock()
		if e.status == entryStatusOpen {
			break
		}
		c.mu.Unlock()
	}

	defer c.mu.Unlock()

	if e.err != nil {
		c.releaseEntry(k, e)
		return nil, nil, e.err
	}
	return e.handle, c.releaseFunc(k, e), nil
}

func (c *cacheImpl) EvictWhere(predicate func(cfg any) bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, e := range c.entries {
		if predicate(e.cfg) {
			c.beginClose(k, e)
		}
	}
}

func (c *cacheImpl) Close(ctx context.Context) error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return errors.New("conncache: already closed")
	}
	c.closed = true

	c.cancel()

	for k, e := range c.entries {
		c.beginClose(k, e)
	}

	c.mu.Unlock()

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		c.mu.Lock()
		var anyE *entry
		for _, e := range c.entries {
			if e.status != entryStatusClosed {
				anyE = e
				break
			}
		}

		if anyE == nil {
			c.mu.Unlock()
			// all entries are closed, we can return
			break
		}

		anyE.closedCh = make(chan struct{})
		c.mu.Unlock()

		<-anyE.closedCh
	}

	return nil
}

// beginClose must be called while c.mu is held.
func (c *cacheImpl) beginClose(k string, e *entry) {
	if e.status == entryStatusClosing || e.status == entryStatusClosed {
		return
	}

	c.retainEntry(k, e)

	go func() {
		for attempt := 0; attempt < 2; attempt++ {
			_, _ = c.singleflight.Do(context.Background(), k, func(_ context.Context) (any, error) {
				c.mu.Lock()
				if e.status == entryStatusClosed {
					c.mu.Unlock()
					return nil, nil
				}
				e.status = entryStatusClosing
				e.since = time.Now()
				c.mu.Unlock()

				var err error
				if e.handle != nil {
					err = e.handle.Close()
				}
				if err == nil {
					err = errors.New("conncache: connection closed")
				}

				c.mu.Lock()
				e.status = entryStatusClosed
				e.since = time.Now()
				if e.closedCh != nil {
					close(e.closedCh)
				}
				e.handle = nil
				e.err = err
				c.mu.Unlock()

				return nil, nil
			})
			// TODO: can return err on panic in Close. Should we handle panics?

			c.mu.Lock()
			if e.status == entryStatusClosed {
				c.mu.Unlock()
				break
			}
			c.mu.Unlock()
		}

		c.mu.Lock()
		c.releaseEntry(k, e)
		c.mu.Unlock()
	}()
}

func (c *cacheImpl) lruEvictionHandler(key, value any) {
	k := key.(string)
	e := value.(*entry)

	// The callback also gets called when removing from LRU during acquisition.
	// We use conn.refs != 0 to signal that its being acquired and should not be closed.
	if e.refs == 0 {
		c.beginClose(k, e)
	}
}

func (c *cacheImpl) retainEntry(key string, e *entry) {
	e.refs++
	if e.refs == 1 {
		// NOTE: lru.Remove is safe even if it's not in the LRU (should only happen if the entry is acquired for the first time)
		_ = c.lru.Remove(key)
	}
}

func (c *cacheImpl) releaseEntry(key string, e *entry) {
	e.refs--
	if e.refs == 0 {
		// If open, keep entry and put in LRU. Else remove entirely.
		if e.status != entryStatusClosing && e.status != entryStatusClosed {
			c.lru.Add(key, e)
		} else {
			delete(c.entries, key)
		}
	}
}

func (c *cacheImpl) releaseFunc(key string, e *entry) ReleaseFunc {
	return func() {
		c.mu.Lock()
		c.releaseEntry(key, e)
		c.mu.Unlock()
	}
}

var checkHangingInterval = time.Minute

func (c *cacheImpl) periodicallyCheckHangingConnections() {
	ticker := time.NewTicker(checkHangingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.mu.Lock()
			for _, e := range c.entries {
				if c.opts.OpenTimeout != 0 && e.status == entryStatusOpening && time.Since(e.since) >= c.opts.OpenTimeout {
					c.opts.HangingFunc(e.cfg, true)
				}
				if c.opts.CloseTimeout != 0 && e.status == entryStatusClosing && time.Since(e.since) >= c.opts.CloseTimeout {
					c.opts.HangingFunc(e.cfg, false)
				}
			}
			c.mu.Unlock()
		case <-c.ctx.Done():
			return
		}
	}
}
