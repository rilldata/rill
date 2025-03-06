package conncache

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type mockConn struct {
	cfg         string
	closeDelay  time.Duration
	closeCalled atomic.Bool
}

func (c *mockConn) Driver() string {
	return "mock"
}

func (c *mockConn) Close() error {
	c.closeCalled.Store(true)
	time.Sleep(c.closeDelay)
	return nil
}

func TestBasic(t *testing.T) {
	opens := atomic.Int64{}

	c := New(Options{
		MaxIdleConnections: 2,
		OpenFunc: func(ctx context.Context, cfg any) (Connection, error) {
			opens.Add(1)
			return &mockConn{cfg: cfg.(string)}, nil
		},
		KeyFunc: func(cfg any) string {
			return cfg.(string)
		},
	})

	// Get "foo"
	m1, r1, err := c.Acquire(context.Background(), "foo")
	require.NoError(t, err)
	require.Equal(t, int64(1), opens.Load())

	// Get "foo" again
	m2, r2, err := c.Acquire(context.Background(), "foo")
	require.NoError(t, err)
	require.Equal(t, int64(1), opens.Load())

	// Check that they're the same
	require.Equal(t, m1, m2)

	// Release the "foo"s and get "foo" again, check it's the same
	r1()
	r2()
	m3, r3, err := c.Acquire(context.Background(), "foo")
	require.NoError(t, err)
	require.Equal(t, int64(1), opens.Load())
	require.Equal(t, m1, m3)
	r3()

	// Open and release two more conns, check "foo" is closed (since LRU size is 2)
	for i := 0; i < 2; i++ {
		_, r, err := c.Acquire(context.Background(), fmt.Sprintf("bar%d", i))
		require.NoError(t, err)
		require.Equal(t, int64(1+i+1), opens.Load())
		r()
	}
	time.Sleep(time.Second)
	require.Equal(t, true, m1.(*mockConn).closeCalled.Load())

	// Close cache
	require.NoError(t, c.Close(context.Background()))
}

func TestConcurrentOpen(t *testing.T) {
	opens := atomic.Int64{}

	c := New(Options{
		MaxIdleConnections: 2,
		OpenFunc: func(ctx context.Context, cfg any) (Connection, error) {
			opens.Add(1)
			time.Sleep(time.Second)
			return &mockConn{cfg: cfg.(string)}, nil
		},
		KeyFunc: func(cfg any) string {
			return cfg.(string)
		},
	})

	var m1, m2 Connection

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		m, _, err := c.Acquire(context.Background(), "foo")
		require.NoError(t, err)
		m1 = m
	}()
	go func() {
		defer wg.Done()
		m, _, err := c.Acquire(context.Background(), "foo")
		require.NoError(t, err)
		m2 = m
	}()

	wg.Wait()
	require.NotNil(t, m1)
	require.Equal(t, m1, m2)
	require.Equal(t, int64(1), opens.Load())

	// Close cache
	require.NoError(t, c.Close(context.Background()))
}

func TestOpenDuringClose(t *testing.T) {
	opens := atomic.Int64{}

	c := New(Options{
		MaxIdleConnections: 2,
		OpenFunc: func(ctx context.Context, cfg any) (Connection, error) {
			opens.Add(1)
			return &mockConn{
				cfg:        cfg.(string),
				closeDelay: time.Second, // Closes hang for 1s
			}, nil
		},
		KeyFunc: func(cfg any) string {
			return cfg.(string)
		},
	})

	// Create conn
	m1, r1, err := c.Acquire(context.Background(), "foo")
	require.NoError(t, err)
	require.Equal(t, int64(1), opens.Load())
	r1()

	// Evict it so it starts closing
	c.EvictWhere(func(cfg any) bool { return true })
	// closeCalled is set before mockConn.Close hangs, but it will take 1s to actually close
	time.Sleep(100 * time.Millisecond)
	require.True(t, m1.(*mockConn).closeCalled.Load())

	// Open again, check it takes ~1s to do so
	start := time.Now()
	m2, r2, err := c.Acquire(context.Background(), "foo")
	require.NoError(t, err)
	require.Greater(t, time.Since(start), 500*time.Millisecond)
	require.Equal(t, int64(2), opens.Load())
	require.NotEqual(t, m1, m2)
	r2()

	// Close cache
	require.NoError(t, c.Close(context.Background()))
}

func TestCloseDuringOpen(t *testing.T) {
	opens := atomic.Int64{}
	m := &mockConn{cfg: "foo"}

	c := New(Options{
		MaxIdleConnections: 2,
		OpenFunc: func(ctx context.Context, cfg any) (Connection, error) {
			time.Sleep(time.Second)
			opens.Add(1)
			return m, nil
		},
		KeyFunc: func(cfg any) string {
			return cfg.(string)
		},
	})

	// Start opening
	go func() {
		_, _, err := c.Acquire(context.Background(), "foo")
		require.NoError(t, err)
		require.Equal(t, int64(2), opens.Load())
	}()

	// Evict it so it starts closing
	time.Sleep(100 * time.Millisecond) // Give it time to start opening
	c.EvictWhere(func(cfg any) bool { return true })

	// It will let the open finish before closing it, so will take ~1s
	time.Sleep(2 * time.Second)
	require.True(t, m.closeCalled.Load())

	// Close cache
	require.NoError(t, c.Close(context.Background()))
}

func TestCloseInUse(t *testing.T) {
	opens := atomic.Int64{}

	c := New(Options{
		MaxIdleConnections: 2,
		OpenFunc: func(ctx context.Context, cfg any) (Connection, error) {
			opens.Add(1)
			return &mockConn{cfg: cfg.(string)}, nil
		},
		KeyFunc: func(cfg any) string {
			return cfg.(string)
		},
	})

	// Open conn "foo"
	m1, r1, err := c.Acquire(context.Background(), "foo")
	require.NoError(t, err)
	require.Equal(t, int64(1), opens.Load())

	// Evict it, check it's closed even though still in use (r1 not called)
	c.EvictWhere(func(cfg any) bool { return true })
	time.Sleep(time.Second)
	require.Equal(t, true, m1.(*mockConn).closeCalled.Load())

	// Open "foo" again, check it opens a new one
	m2, r2, err := c.Acquire(context.Background(), "foo")
	require.NoError(t, err)
	require.Equal(t, int64(2), opens.Load())
	require.NotEqual(t, m1, m2)

	// Check that releasing m1 doesn't fail (though it's been closed)
	r1()
	r2()
}

func TestHanging(t *testing.T) {
	hangingOpens := atomic.Int64{}
	hangingCloses := atomic.Int64{}

	c := New(Options{
		MaxIdleConnections:   2,
		OpenTimeout:          100 * time.Millisecond,
		CloseTimeout:         100 * time.Millisecond,
		ErrTTL:               100 * time.Second,
		CheckHangingInterval: 100 * time.Millisecond,
		OpenFunc: func(ctx context.Context, cfg any) (Connection, error) {
			time.Sleep(time.Second)
			return &mockConn{
				cfg:        cfg.(string),
				closeDelay: time.Second, // Make closes hang for 1s
			}, nil
		},
		KeyFunc: func(cfg any) string {
			return cfg.(string)
		},
		HangingFunc: func(cfg any, open bool) {
			if open {
				hangingOpens.Add(1)
			} else {
				hangingCloses.Add(1)
			}
		},
	})

	// Open conn "foo"
	m1, r1, err := c.Acquire(context.Background(), "foo")
	require.NoError(t, err)
	require.GreaterOrEqual(t, hangingOpens.Load(), int64(1))
	r1()

	// Evict it, check it's closed even though still in use (r1 not called)
	c.EvictWhere(func(cfg any) bool { return true })
	time.Sleep(time.Second)
	require.Equal(t, true, m1.(*mockConn).closeCalled.Load())
	require.GreaterOrEqual(t, hangingCloses.Load(), int64(1))
}

func TestAcquireCloseAfterOpening(t *testing.T) {
	c := New(Options{
		MaxIdleConnections: 2,
		OpenFunc: func(ctx context.Context, cfg any) (Connection, error) {
			time.Sleep(time.Second)
			return &mockConn{cfg: cfg.(string)}, nil
		},
		KeyFunc: func(cfg any) string {
			return cfg.(string)
		},
	})

	// Acquire a conn that takes 1s to open, but cancellation after 100ms.
	// The conn will continue opening in the background after the cancellation (for 1s).
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	_, _, err := c.Acquire(ctx, "foo")
	require.ErrorIs(t, err, ctx.Err())

	// Evict all connections, including the one that's still opening.
	// Internally, it will get closeAfterOpening set to true.
	c.EvictWhere(func(cfg any) bool { return true })

	// Now try to acquire it again. We expect it to finish opening (), then be closed (per the eviction), then be opened again and returned here.
	_, r1, err := c.Acquire(context.Background(), "foo")
	require.NoError(t, err)
	r1()

	// Close cache
	require.NoError(t, c.Close(context.Background()))
}

func TestErrorTTL(t *testing.T) {
	opens := atomic.Int64{}
	shouldError := true
	c := New(Options{
		MaxIdleConnections: 2,
		ErrTTL:             250 * time.Millisecond,
		OpenFunc: func(ctx context.Context, cfg any) (Connection, error) {
			opens.Add(1)
			if shouldError {
				return nil, errors.New("errored!")
			}
			return &mockConn{cfg: cfg.(string)}, nil
		},
		KeyFunc: func(cfg any) string {
			return cfg.(string)
		},
	})

	_, _, err := c.Acquire(context.Background(), "foo")
	require.NotNil(t, err)
	require.Equal(t, int64(1), opens.Load())

	_, _, err = c.Acquire(context.Background(), "foo")
	require.NotNil(t, err)
	require.Equal(t, int64(1), opens.Load())

	time.Sleep(500 * time.Millisecond)
	shouldError = false

	_, r1, err := c.Acquire(context.Background(), "foo")
	require.Nil(t, err)
	require.Equal(t, int64(2), opens.Load())
	r1()
}
