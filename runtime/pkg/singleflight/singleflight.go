// Package singleflight provides a duplicate function call suppression
// mechanism. Borrowed from golang.org/x/sync/singleflight with fix
// that if first request gets cancelled/timedout the other requests
// non-cancelled requests get the value instead of cancellation signal.
package singleflight

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"runtime"
	"runtime/debug"
	"sync"
	"time"
)

// errGoexit indicates the runtime.Goexit was called in
// the user given function.
var errGoexit = errors.New("runtime.Goexit was called")

// A panicError is an arbitrary value recovered from a panic
// with the stack trace during the execution of given function.
type panicError struct {
	value interface{}
	stack []byte
}

// Error implements error interface.
func (p *panicError) Error() string {
	return fmt.Sprintf("%v\n\n%s", p.value, p.stack)
}

func newPanicError(v interface{}) error {
	stack := debug.Stack()

	// The first line of the stack trace is of the form "goroutine N [status]:"
	// but by the time the panic reaches Do the goroutine may no longer exist
	// and its status will have changed. Trim out the misleading line.
	if line := bytes.IndexByte(stack, '\n'); line >= 0 {
		stack = stack[line+1:]
	}
	return &panicError{value: v, stack: stack}
}

// call is an in-flight or completed singleflight.Do call
type call[V any] struct {
	ctx     context.Context
	cancel  context.CancelFunc
	counter uint
	val     V
	err     error
}

// Group represents a class of work and forms a namespace in
// which units of work can be executed with duplicate suppression.
type Group[K comparable, V any] struct {
	mu sync.Mutex     // protects m
	m  map[K]*call[V] // lazily initialized
}

// Do executes and returns the results of the given function, making
// sure that only one execution is in-flight for a given key at a
// time. If a duplicate comes in, the duplicate caller waits for the
// original to complete and receives the same results.
func (g *Group[K, V]) Do(ctx context.Context, key K, fn func(context.Context) (V, error)) (V, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[K]*call[V])
	}
	c, ok := g.m[key]
	if !ok {
		cctx, cancel := withCancelAndContextValues(ctx)
		c = &call[V]{
			ctx:    cctx,
			cancel: cancel,
		}
		g.m[key] = c
		go func() {
			g.doCall(c, key, fn)
		}()
	}
	c.counter++
	g.mu.Unlock()

	select {
	case <-ctx.Done():
	case <-c.ctx.Done():
	}

	g.mu.Lock()
	c.counter--
	if c.counter == 0 {
		c.cancel()
		delete(g.m, key)
	}
	g.mu.Unlock()

	if ctx.Err() != nil {
		var empty V
		return empty, ctx.Err()
	}

	pErr := &panicError{}
	if errors.As(c.err, &pErr) {
		panic(pErr)
	} else if errors.Is(c.err, errGoexit) {
		runtime.Goexit()
	}
	return c.val, c.err
}

// doCall handles the single call for a key.
func (g *Group[K, V]) doCall(c *call[V], key K, fn func(ctx context.Context) (V, error)) {
	normalReturn := false
	recovered := false

	// use double-defer to distinguish panic from runtime.Goexit,
	// more details see https://golang.org/cl/134395
	defer func() {
		// the given function invoked runtime.Goexit
		if !normalReturn && !recovered {
			c.err = errGoexit
		}

		c.cancel()
	}()

	func() {
		defer func() {
			if !normalReturn {
				// Ideally, we would wait to take a stack trace until we've determined
				// whether this is a panic or a runtime.Goexit.
				//
				// Unfortunately, the only way we can distinguish the two is to see
				// whether the recover stopped the goroutine from terminating, and by
				// the time we know that, the part of the stack trace relevant to the
				// panic has been discarded.
				if r := recover(); r != nil {
					c.err = newPanicError(r)
				}
			}
		}()

		c.val, c.err = fn(c.ctx)
		normalReturn = true
	}()

	if !normalReturn {
		recovered = true
	}
}

// withCancelAndContextValues returns a context whose done channel is closed when the
// returned cancel function is called. It inherits the values of the parent context.
//
// Canceling this context releases resources associated with it, so code should
// call cancel as soon as the operations running in this Context complete.
func withCancelAndContextValues(parent context.Context) (context.Context, context.CancelFunc) {
	if parent == nil {
		panic("cannot create context from nil parent")
	}
	ctx, cancel := context.WithCancel(context.Background())
	return withCancelAndParentValuesCtx{ctx: ctx, parentCtx: parent}, cancel
}

type withCancelAndParentValuesCtx struct {
	ctx       context.Context
	parentCtx context.Context
}

func (c withCancelAndParentValuesCtx) Deadline() (deadline time.Time, ok bool) {
	return c.ctx.Deadline()
}

func (c withCancelAndParentValuesCtx) Done() <-chan struct{} {
	return c.ctx.Done()
}

func (c withCancelAndParentValuesCtx) Err() error {
	return c.ctx.Err()
}

func (c withCancelAndParentValuesCtx) Value(key any) any {
	return c.parentCtx.Value(key)
}
