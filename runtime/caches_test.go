package runtime

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/rilldata/rill/runtime/drivers/sqlite"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestConnectionCache(t *testing.T) {
	ctx := context.Background()
	id := uuid.NewString()

	c := newConnectionCache(10, zap.NewNop())
	conn1, release, err := c.get(ctx, id, "sqlite", map[string]any{"dsn": ":memory:"}, false)
	require.NoError(t, err)
	release()
	require.NotNil(t, conn1)

	conn2, release, err := c.get(ctx, id, "sqlite", map[string]any{"dsn": ":memory:"}, false)
	require.NoError(t, err)
	release()
	require.NotNil(t, conn2)

	conn3, release, err := c.get(ctx, uuid.NewString(), "sqlite", map[string]any{"dsn": ":memory:"}, false)
	require.NoError(t, err)
	release()
	require.NotNil(t, conn3)

	require.True(t, conn1 == conn2)
	require.False(t, conn2 == conn3)
}

func TestConnectionCacheWithAllShared(t *testing.T) {
	ctx := context.Background()
	id := uuid.NewString()

	c := newConnectionCache(1, zap.NewNop())
	conn1, release, err := c.get(ctx, id, "sqlite", map[string]any{"dsn": ":memory:"}, true)
	require.NoError(t, err)
	require.NotNil(t, conn1)
	defer release()

	conn2, release, err := c.get(ctx, id, "sqlite", map[string]any{"dsn": ":memory:"}, true)
	require.NoError(t, err)
	require.NotNil(t, conn2)
	defer release()

	conn3, release, err := c.get(ctx, uuid.NewString(), "sqlite", map[string]any{"dsn": ":memory:"}, true)
	require.NoError(t, err)
	require.NotNil(t, conn3)
	defer release()

	require.True(t, conn1 == conn2)
	require.True(t, conn2 == conn3)
	require.Equal(t, 1, len(c.cache))
	require.Equal(t, 0, c.lruCache.Len())
}

func TestConnectionCacheWithAllOpen(t *testing.T) {
	ctx := context.Background()

	c := newConnectionCache(1, zap.NewNop())
	conn1, r1, err := c.get(ctx, uuid.NewString(), "sqlite", map[string]any{"dsn": ":memory:"}, false)
	require.NoError(t, err)
	require.NotNil(t, conn1)

	conn2, r2, err := c.get(ctx, uuid.NewString(), "sqlite", map[string]any{"dsn": ":memory:"}, false)
	require.NoError(t, err)
	require.NotNil(t, conn2)

	conn3, r3, err := c.get(ctx, uuid.NewString(), "sqlite", map[string]any{"dsn": ":memory:"}, false)
	require.NoError(t, err)
	require.NotNil(t, conn3)

	require.Equal(t, 3, len(c.cache))
	require.Equal(t, 0, c.lruCache.Len())
	// release all connections
	r1()
	r2()
	r3()
	require.Equal(t, 0, len(c.cache))
	require.Equal(t, 1, c.lruCache.Len())
	_, val, _ := c.lruCache.GetOldest()
	require.True(t, conn3 == val.(*connWithRef).Handle)
}

func TestConnectionCacheParallel(t *testing.T) {
	ctx := context.Background()

	c := newConnectionCache(5, zap.NewNop())
	defer c.Close()

	var wg sync.WaitGroup
	// open 10 connections and do not release
	go func() {
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				conn, _, err := c.get(ctx, uuid.NewString(), "sqlite", map[string]any{"dsn": ":memory:"}, false)
				require.NoError(t, err)
				require.NotNil(t, conn)
				time.Sleep(100 * time.Millisecond)
			}()
		}
	}()

	// open 20 connections and release
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			conn, r, err := c.get(ctx, uuid.NewString(), "sqlite", map[string]any{"dsn": ":memory:"}, false)
			defer r()
			require.NoError(t, err)
			require.NotNil(t, conn)
			time.Sleep(100 * time.Millisecond)
		}()
	}
	wg.Wait()

	// 10 connections were not released so should be present in in-use cache
	require.Equal(t, 10, len(c.cache))
	// 20 connections were released so 15 should be evicted
	require.Equal(t, 5, c.lruCache.Len())
}

func TestConnectionCacheMultipleConfigs(t *testing.T) {
	ctx := context.Background()

	c := newConnectionCache(10, zap.NewNop())
	defer c.Close()
	conn1, r1, err := c.get(ctx, uuid.NewString(), "sqlite", map[string]any{"dsn": ":memory:", "host": "localhost:8080", "allow_host_access": "true"}, true)
	require.NoError(t, err)
	require.NotNil(t, conn1)

	conn2, r2, err := c.get(ctx, uuid.NewString(), "sqlite", map[string]any{"dsn": ":memory:", "host": "localhost:8080", "allow_host_access": "true"}, true)
	require.NoError(t, err)
	require.NotNil(t, conn2)

	conn3, r3, err := c.get(ctx, uuid.NewString(), "sqlite", map[string]any{"dsn": ":memory:", "host": "localhost:8080", "allow_host_access": "true"}, true)
	require.NoError(t, err)
	require.NotNil(t, conn3)

	require.Equal(t, 1, len(c.cache))
	require.Equal(t, 0, c.lruCache.Len())
	// release all connections
	r1()
	r2()
	r3()
	require.Equal(t, 0, len(c.cache))
	require.Equal(t, 1, c.lruCache.Len())
}
