package runtime

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	_ "github.com/rilldata/rill/runtime/drivers/sqlite"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestConnectionCache(t *testing.T) {
	ctx := context.Background()
	id := "default"

	rt := NewTestRunTimeWithInst(t)
	c := newConnectionCache(10, zap.NewNop(), rt, activity.NewNoopClient())
	conn1, release, err := c.get(ctx, id, "sqlite", map[string]any{"dsn": ":memory:"}, false)
	require.NoError(t, err)
	release()
	require.NotNil(t, conn1)

	conn2, release, err := c.get(ctx, id, "sqlite", map[string]any{"dsn": ":memory:"}, false)
	require.NoError(t, err)
	release()
	require.NotNil(t, conn2)

	inst := &drivers.Instance{
		ID:            "default1",
		OLAPConnector: "duckdb",
		RepoConnector: "repo",
		EmbedCatalog:  true,
		Connectors: []*runtimev1.Connector{
			{
				Type:   "file",
				Name:   "repo",
				Config: map[string]string{"dsn": ""},
			},
			{
				Type:   "duckdb",
				Name:   "duckdb",
				Config: map[string]string{"dsn": ""},
			},
		},
	}
	require.NoError(t, rt.CreateInstance(context.Background(), inst))

	conn3, release, err := c.get(ctx, "default1", "sqlite", map[string]any{"dsn": ":memory:"}, false)
	require.NoError(t, err)
	release()
	require.NotNil(t, conn3)

	require.True(t, conn1 == conn2)
	require.False(t, conn2 == conn3)
}

func TestConnectionCacheWithAllShared(t *testing.T) {
	ctx := context.Background()
	id := "default"

	c := newConnectionCache(1, zap.NewNop(), NewTestRunTimeWithInst(t), activity.NewNoopClient())
	conn1, release, err := c.get(ctx, id, "sqlite", map[string]any{"dsn": ":memory:"}, true)
	require.NoError(t, err)
	require.NotNil(t, conn1)
	defer release()

	conn2, release, err := c.get(ctx, id, "sqlite", map[string]any{"dsn": ":memory:"}, true)
	require.NoError(t, err)
	require.NotNil(t, conn2)
	defer release()

	conn3, release, err := c.get(ctx, "default", "sqlite", map[string]any{"dsn": ":memory:"}, true)
	require.NoError(t, err)
	require.NotNil(t, conn3)
	defer release()

	require.True(t, conn1 == conn2)
	require.True(t, conn2 == conn3)
	require.Equal(t, 1, len(c.acquired))
	require.Equal(t, 0, c.lru.Len())
}

func TestConnectionCacheWithAllOpen(t *testing.T) {
	ctx := context.Background()

	rt := NewTestRunTimeWithInst(t)
	c := newConnectionCache(1, zap.NewNop(), rt, activity.NewNoopClient())
	conn1, r1, err := c.get(ctx, "default", "sqlite", map[string]any{"dsn": ":memory:"}, false)
	require.NoError(t, err)
	require.NotNil(t, conn1)

	createInstance(t, rt, "default1")
	conn2, r2, err := c.get(ctx, "default1", "sqlite", map[string]any{"dsn": ":memory:"}, false)
	require.NoError(t, err)
	require.NotNil(t, conn2)

	createInstance(t, rt, "default2")
	conn3, r3, err := c.get(ctx, "default2", "sqlite", map[string]any{"dsn": ":memory:"}, false)
	require.NoError(t, err)
	require.NotNil(t, conn3)

	require.Equal(t, 3, len(c.acquired))
	require.Equal(t, 0, c.lru.Len())
	// release all connections
	r1()
	r2()
	r3()
	require.Equal(t, 0, len(c.acquired))
	require.Equal(t, 1, c.lru.Len())
	_, val, _ := c.lru.GetOldest()
	require.True(t, conn3 == val.(*connWithRef).handle)
}

func TestConnectionCacheParallel(t *testing.T) {
	ctx := context.Background()

	rt := NewTestRunTimeWithInst(t)
	c := newConnectionCache(5, zap.NewNop(), rt, activity.NewNoopClient())
	defer c.Close()

	var wg sync.WaitGroup
	wg.Add(30)
	// open 10 connections and do not release
	go func() {
		for i := 0; i < 10; i++ {
			j := i
			go func() {
				defer wg.Done()
				id := fmt.Sprintf("default%v", 100+j)
				createInstance(t, rt, id)
				conn, _, err := c.get(ctx, id, "sqlite", map[string]any{"dsn": ":memory:"}, false)
				require.NoError(t, err)
				require.NotNil(t, conn)
				time.Sleep(100 * time.Millisecond)
			}()
		}
	}()

	// open 20 connections and release
	for i := 0; i < 20; i++ {
		j := i
		go func() {
			defer wg.Done()
			id := fmt.Sprintf("default%v", 200+j)
			createInstance(t, rt, id)
			conn, r, err := c.get(ctx, id, "sqlite", map[string]any{"dsn": ":memory:"}, false)
			defer r()
			require.NoError(t, err)
			require.NotNil(t, conn)
			time.Sleep(100 * time.Millisecond)
		}()
	}
	wg.Wait()

	// 10 connections were not released so should be present in in-use cache
	require.Equal(t, 10, len(c.acquired))
	// 20 connections were released so 15 should be evicted
	require.Equal(t, 5, c.lru.Len())
}

func TestConnectionCacheMultipleConfigs(t *testing.T) {
	ctx := context.Background()

	c := newConnectionCache(10, zap.NewNop(), NewTestRunTimeWithInst(t), activity.NewNoopClient())
	defer c.Close()
	conn1, r1, err := c.get(ctx, "default", "sqlite", map[string]any{"dsn": ":memory:", "host": "localhost:8080", "allow_host_access": "true"}, true)
	require.NoError(t, err)
	require.NotNil(t, conn1)

	conn2, r2, err := c.get(ctx, "default", "sqlite", map[string]any{"dsn": ":memory:", "host": "localhost:8080", "allow_host_access": "true"}, true)
	require.NoError(t, err)
	require.NotNil(t, conn2)

	conn3, r3, err := c.get(ctx, "default", "sqlite", map[string]any{"dsn": ":memory:", "host": "localhost:8080", "allow_host_access": "true"}, true)
	require.NoError(t, err)
	require.NotNil(t, conn3)

	require.Equal(t, 1, len(c.acquired))
	require.Equal(t, 0, c.lru.Len())
	// release all connections
	r1()
	r2()
	r3()
	require.Equal(t, 0, len(c.acquired))
	require.Equal(t, 1, c.lru.Len())
}

func TestConnectionCacheParallelCalls(t *testing.T) {
	ctx := context.Background()

	c := newConnectionCache(10, zap.NewNop(), NewTestRunTimeWithInst(t), activity.NewNoopClient())
	defer c.Close()

	m := &mockDriver{}
	drivers.Register("mock_driver", m)
	defer func() {
		delete(drivers.Drivers, "mock_driver")
	}()

	var wg sync.WaitGroup
	wg.Add(10)
	// open 10 connections and verify no error
	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			conn, _, err := c.get(ctx, "default", "mock_driver", map[string]any{"sleep": int64(100)}, false)
			require.NoError(t, err)
			require.NotNil(t, conn)
		}()
	}
	wg.Wait()

	require.Equal(t, int32(1), m.opened.Load())
	require.Equal(t, 1, len(c.acquired))
}

func TestConnectionCacheBlockingCalls(t *testing.T) {
	ctx := context.Background()

	rt := NewTestRunTimeWithInst(t)
	c := newConnectionCache(10, zap.NewNop(), rt, activity.NewNoopClient())
	defer c.Close()

	m := &mockDriver{}
	drivers.Register("mock_driver", m)
	defer func() {
		delete(drivers.Drivers, "mock_driver")
	}()

	var wg sync.WaitGroup
	wg.Add(12)
	// open 1 slow connection
	go func() {
		defer wg.Done()
		conn, _, err := c.get(ctx, "default", "mock_driver", map[string]any{"sleep": int64(1000)}, false)
		require.NoError(t, err)
		require.NotNil(t, conn)
	}()

	// open 10 fast different connections(takes 10-20 ms to open) and verify not blocked
	for i := 0; i < 10; i++ {
		j := i
		go func() {
			defer wg.Done()
			conn, _, err := c.get(ctx, "default", "mock_driver", map[string]any{"sleep": int64(j + 10)}, false)
			require.NoError(t, err)
			require.NotNil(t, conn)
		}()
	}

	// verify that after 100 ms 11 connections have been opened
	go func() {
		time.Sleep(100 * time.Millisecond)
		require.Equal(t, int32(11), m.opened.Load())
		wg.Done()
	}()
	wg.Wait()

	require.Equal(t, int32(11), m.opened.Load())
}

type mockDriver struct {
	opened atomic.Int32
}

// Drop implements drivers.Driver.
func (*mockDriver) Drop(config map[string]any, logger *zap.Logger) error {
	panic("unimplemented")
}

// HasAnonymousSourceAccess implements drivers.Driver.
func (*mockDriver) HasAnonymousSourceAccess(ctx context.Context, src map[string]any, logger *zap.Logger) (bool, error) {
	panic("unimplemented")
}

// Open implements drivers.Driver.
func (m *mockDriver) Open(config map[string]any, shared bool, client activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	m.opened.Add(1)
	sleep := config["sleep"].(int64)
	time.Sleep(time.Duration(sleep) * time.Millisecond)
	return &mockHandle{}, nil
}

// Spec implements drivers.Driver.
func (*mockDriver) Spec() drivers.Spec {
	panic("unimplemented")
}

var _ drivers.Driver = &mockDriver{}

type mockHandle struct {
}

// AsCatalogStore implements drivers.Handle.
func (*mockHandle) AsCatalogStore(instanceID string) (drivers.CatalogStore, bool) {
	panic("unimplemented")
}

// AsFileStore implements drivers.Handle.
func (*mockHandle) AsFileStore() (drivers.FileStore, bool) {
	panic("unimplemented")
}

// AsOLAP implements drivers.Handle.
func (*mockHandle) AsOLAP(instanceID string) (drivers.OLAPStore, bool) {
	panic("unimplemented")
}

// AsObjectStore implements drivers.Handle.
func (*mockHandle) AsObjectStore() (drivers.ObjectStore, bool) {
	panic("unimplemented")
}

// AsRegistry implements drivers.Handle.
func (*mockHandle) AsRegistry() (drivers.RegistryStore, bool) {
	panic("unimplemented")
}

// AsRepoStore implements drivers.Handle.
func (*mockHandle) AsRepoStore(instanceID string) (drivers.RepoStore, bool) {
	panic("unimplemented")
}

// AsSQLStore implements drivers.Handle.
func (*mockHandle) AsSQLStore() (drivers.SQLStore, bool) {
	panic("unimplemented")
}

// AsTransporter implements drivers.Handle.
func (*mockHandle) AsTransporter(from drivers.Handle, to drivers.Handle) (drivers.Transporter, bool) {
	panic("unimplemented")
}

// Close implements drivers.Handle.
func (*mockHandle) Close() error {
	return nil
}

// Config implements drivers.Handle.
func (*mockHandle) Config() map[string]any {
	panic("unimplemented")
}

// Driver implements drivers.Handle.
func (*mockHandle) Driver() string {
	panic("unimplemented")
}

// Migrate implements drivers.Handle.
func (*mockHandle) Migrate(ctx context.Context) error {
	return nil
}

// MigrationStatus implements drivers.Handle.
func (*mockHandle) MigrationStatus(ctx context.Context) (current int, desired int, err error) {
	panic("unimplemented")
}

var _ drivers.Handle = &mockHandle{}

func NewTestRunTimeWithInst(t *testing.T) *Runtime {
	rt := NewTestRunTime(t)
	createInstance(t, rt, "default")
	return rt
}

func createInstance(t *testing.T, rt *Runtime, instanceId string) {
	inst := &drivers.Instance{
		ID:            instanceId,
		OLAPConnector: "duckdb",
		RepoConnector: "repo",
		EmbedCatalog:  true,
		Connectors: []*runtimev1.Connector{
			{
				Type:   "file",
				Name:   "repo",
				Config: map[string]string{"dsn": ""},
			},
			{
				Type:   "duckdb",
				Name:   "duckdb",
				Config: map[string]string{"dsn": ""},
			},
		},
	}
	require.NoError(t, rt.CreateInstance(context.Background(), inst))
}
