package conncache

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/hashicorp/golang-lru/simplelru"
	"go.opentelemetry.io/otel/metric"
)

// Cache is a concurrency-safe cache of stateful connection objects.
// It differs from a connection pool in that it's designed for caching heterogenous connections.
// The cache will at most open one connection per key, even under concurrent access.
// The cache automatically evicts connections that are not in use ("acquired") using a least-recently-used policy.
type Cache interface {
	// Acquire retrieves or opens a connection for the given config. The returned ReleaseFunc must be called when the connection is no longer needed.
	// While a connection is acquired, it will not be closed unless EvictWhere or Close is called.
	// If Acquire is called while the underlying connection is being evicted, it will wait for the close to complete and then open a new connection.
	// If opening the connection fails, Acquire may return the error on subsequent calls without trying to open again until the entry is evicted.
	Acquire(ctx context.Context, cfg any) (Connection, ReleaseFunc, error)

	// EvictWhere closes the connections that match the predicate.
	// It immediately starts closing the connections, even those that are currently acquired.
	// It returns quickly and does not wait for connections to finish closing in the background.
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
	// MaxIdleConnections is the maximum number of non-acquired connections that will be kept open.
	MaxIdleConnections int
	// OpenTimeout is the maximum amount of time to wait for a connection to open.
	OpenTimeout time.Duration
	// CloseTimeout is the maximum amount of time to wait for a connection to close.
	CloseTimeout time.Duration
	// CheckHangingInterval is the interval at which to check for hanging open/close calls.
	CheckHangingInterval time.Duration
	// OpenFunc opens a connection.
	OpenFunc func(ctx context.Context, cfg any) (Connection, error)
	// KeyFunc computes a comparable key for a connection config.
	KeyFunc func(cfg any) string
	// HangingFunc is called when an open or close exceeds its timeout and does not respond to context cancellation.
	HangingFunc func(cfg any, open bool)
	// Metrics are optional instruments for observability.
	Metrics Metrics
}

// Metrics are optional instruments for observability. If an instrument is nil, it will not be collected.
type Metrics struct {
	Opens          metric.Int64Counter
	Closes         metric.Int64Counter
	SizeTotal      metric.Int64UpDownCounter
	SizeLRU        metric.Int64UpDownCounter
	OpenLatencyMS  metric.Int64Histogram
	CloseLatencyMS metric.Int64Histogram
}

var _ Cache = (*cacheImpl)(nil)

// cacheImpl implements Cache. Implementation notes:
// - It uses an LRU to pool unused connections and eventually close them.
// - It leverages a singleflight pattern to ensure at most one open/close action runs against a connection at a time.
// - It directly implements a singleflight (instead of using a library) because it needs to use the same mutex for the singleflight and the map/LRU to avoid race conditions.
// - An entry will only have entryStatusOpening or entryStatusClosing if a singleflight call is currently running for it.
// - Any code that keeps a reference to an entry after the mutex is released must call retainEntry/releaseEntry.
// - If the ctx for an open call is cancelled, the entry will continue opening in the background (and will be put in the LRU).
// - If attempting to open a closing entry, or close an opening entry, we wait for the singleflight to complete and then retry once. To avoid infinite loops, we don't retry more than once.
type cacheImpl struct {
	opts         Options
	closed       bool
	mu           sync.Mutex
	entries      map[string]*entry
	lru          *simplelru.LRU
	singleflight map[string]chan struct{}
	ctx          context.Context
	cancel       context.CancelFunc
}

type entry struct {
	cfg               any
	refs              int
	status            entryStatus
	since             time.Time
	closeAfterOpening bool
	handle            Connection
	err               error
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
		singleflight: make(map[string]chan struct{}),
		ctx:          ctx,
		cancel:       cancel,
	}

	var err error
	c.lru, err = simplelru.NewLRU(opts.MaxIdleConnections, c.lruEvictionHandler)
	if err != nil {
		panic(err)
	}

	if opts.CheckHangingInterval != 0 {
		go c.periodicallyCheckHangingConnections()
	}

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
		if c.opts.Metrics.SizeTotal != nil {
			c.opts.Metrics.SizeTotal.Add(c.ctx, 1)
		}
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

	ch, ok := c.singleflight[k]

	if ok && e.status == entryStatusClosing {
		c.mu.Unlock()
		select {
		case <-ch:
		case <-ctx.Done():
			c.mu.Lock()
			c.releaseEntry(k, e)
			c.mu.Unlock()
			return nil, nil, ctx.Err()
		}
		c.mu.Lock()

		// Since we released the lock, need to check c.closed and e.status again.
		if c.closed {
			c.releaseEntry(k, e)
			c.mu.Unlock()
			return nil, nil, errors.New("conncache: closed")
		}

		if e.status == entryStatusOpen {
			defer c.mu.Unlock()
			if e.err != nil {
				c.releaseEntry(k, e)
				return nil, nil, e.err
			}
			return e.handle, c.releaseFunc(k, e), nil
		}

		ch, ok = c.singleflight[k]
	}

	if !ok {
		c.retainEntry(k, e) // Retain again to count the goroutine's reference independently (in case ctx is cancelled while the Open continues in the background)

		ch = make(chan struct{})
		c.singleflight[k] = ch

		e.status = entryStatusOpening
		e.since = time.Now()
		e.handle = nil
		e.err = nil

		go func() {
			start := time.Now()
			var handle Connection
			var err error
			if c.opts.OpenTimeout == 0 {
				handle, err = c.opts.OpenFunc(c.ctx, cfg)
			} else {
				ctx, cancel := context.WithTimeout(c.ctx, c.opts.OpenTimeout)
				handle, err = c.opts.OpenFunc(ctx, cfg)
				cancel()
			}

			if c.opts.Metrics.Opens != nil {
				c.opts.Metrics.Opens.Add(c.ctx, 1)
			}
			if c.opts.Metrics.OpenLatencyMS != nil {
				c.opts.Metrics.OpenLatencyMS.Record(c.ctx, time.Since(start).Milliseconds())
			}

			c.mu.Lock()
			defer c.mu.Unlock()

			e.status = entryStatusOpen
			e.since = time.Now()
			e.handle = handle
			e.err = err

			delete(c.singleflight, k)
			close(ch)

			if e.closeAfterOpening {
				e.closeAfterOpening = false
				c.beginClose(k, e)
			}

			c.releaseEntry(k, e)
		}()
	}

	c.mu.Unlock()

	select {
	case <-ch:
	case <-ctx.Done():
		c.mu.Lock()
		c.releaseEntry(k, e)
		c.mu.Unlock()
		return nil, nil, ctx.Err()
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if e.status != entryStatusOpen {
		c.releaseEntry(k, e)
		return nil, nil, errors.New("conncache: connection was immediately closed after being opened")
	}

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
		c.mu.Lock()
		var anyCh chan struct{}
		for _, ch := range c.singleflight {
			anyCh = ch
			break
		}
		c.mu.Unlock()

		if anyCh == nil {
			// all entries are closed, we can return
			return nil
		}

		select {
		case <-anyCh:
			// continue
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// beginClose must be called while c.mu is held.
func (c *cacheImpl) beginClose(k string, e *entry) {
	if e.status == entryStatusClosing || e.status == entryStatusClosed {
		return
	}

	if e.status == entryStatusOpening {
		e.closeAfterOpening = true
		return
	}

	c.retainEntry(k, e)

	ch, ok := c.singleflight[k]
	if ok {
		// Should never happen, but checking since it would be pretty bad.
		panic(errors.New("conncache: singleflight exists for entry that is neither opening nor closing"))
	}
	ch = make(chan struct{})
	c.singleflight[k] = ch

	e.status = entryStatusClosing
	e.since = time.Now()

	go func() {
		start := time.Now()
		var err error
		if e.handle != nil {
			err = e.handle.Close()
		}
		if err == nil {
			err = errors.New("conncache: connection closed")
		}

		if c.opts.Metrics.Closes != nil {
			c.opts.Metrics.Closes.Add(c.ctx, 1)
		}
		if c.opts.Metrics.CloseLatencyMS != nil {
			c.opts.Metrics.CloseLatencyMS.Record(c.ctx, time.Since(start).Milliseconds())
		}

		c.mu.Lock()
		defer c.mu.Unlock()

		e.status = entryStatusClosed
		e.since = time.Now()
		e.handle = nil
		e.err = err

		delete(c.singleflight, k)
		close(ch)

		c.releaseEntry(k, e)
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
		if c.opts.Metrics.SizeLRU != nil {
			c.opts.Metrics.SizeLRU.Add(c.ctx, -1)
		}
	}
}

func (c *cacheImpl) releaseEntry(key string, e *entry) {
	e.refs--
	if e.refs == 0 {
		// If open, keep entry and put in LRU. Else remove entirely.
		if e.status != entryStatusClosing && e.status != entryStatusClosed {
			c.lru.Add(key, e)
			if c.opts.Metrics.SizeLRU != nil {
				c.opts.Metrics.SizeLRU.Add(c.ctx, 1)
			}
		} else {
			delete(c.entries, key)
			if c.opts.Metrics.SizeTotal != nil {
				c.opts.Metrics.SizeTotal.Add(c.ctx, -1)
			}
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

func (c *cacheImpl) periodicallyCheckHangingConnections() {
	ticker := time.NewTicker(c.opts.CheckHangingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.mu.Lock()
			for k := range c.singleflight {
				e := c.entries[k]
				if c.opts.OpenTimeout != 0 && e.status == entryStatusOpening && time.Since(e.since) > c.opts.OpenTimeout {
					c.opts.HangingFunc(e.cfg, true)
				}
				if c.opts.CloseTimeout != 0 && e.status == entryStatusClosing && time.Since(e.since) > c.opts.CloseTimeout {
					c.opts.HangingFunc(e.cfg, false)
				}
			}
			c.mu.Unlock()
		case <-c.ctx.Done():
			return
		}
	}
}
