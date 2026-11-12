package runtime_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/queries"
	"github.com/rilldata/rill/runtime/storage"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestQueryNotCachedForPerUserOAuth(t *testing.T) {
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "olap_connector: dbx",
			"connectors/dbx.yaml": `
type: connector
driver: query_cache_test_duckdb
mode: readwrite
init_sql: CREATE TABLE t AS SELECT 1 AS x
`,
			"connectors/plain.yaml": `
type: connector
driver: query_cache_test_duckdb
mode: readwrite
init_sql: CREATE TABLE t AS SELECT 1 AS x
`,
			// A model named like the tables, so the queries have a dependency to cache on
			"models/t.yaml": `
type: model
connector: duckdb
sql: SELECT 1 AS x
output:
  connector: duckdb
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 4, 0, 0)

	query := func(connector string) {
		q := &queries.TableCardinality{Connector: connector, TableName: "t"}
		require.NoError(t, rt.Query(t.Context(), instanceID, q, 1))
		require.Equal(t, int64(1), q.Result)
	}

	// Other connectors cache results; the cache applies writes asynchronously.
	cached := false
	for range 100 {
		n := olapQueries.get("plain")
		query("plain")
		if olapQueries.get("plain") == n {
			cached = true
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	require.True(t, cached)

	// Per-user OAuth connectors don't read the cached result, though their queries have the same cache key and dependencies.
	// The empty connector is the instance's default OLAP connector, dbx.
	n := olapQueries.get("dbx")
	for _, connector := range []string{"", "", "dbx", "dbx"} {
		query(connector)
	}
	require.Equal(t, n+4, olapQueries.get("dbx"))
}

func init() {
	drivers.Register("query_cache_test_duckdb", queryCacheTestDriver{drivers.Drivers["duckdb"]})
}

// olapQueries counts the OLAP queries per connector of query_cache_test_duckdb.
var olapQueries = &queryCounter{counts: make(map[string]int)}

type queryCounter struct {
	mu     sync.Mutex
	counts map[string]int
}

func (c *queryCounter) get(connector string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.counts[connector]
}

// queryCacheTestDriver wraps DuckDB; its connector named "dbx" uses per-user OAuth.
type queryCacheTestDriver struct {
	drivers.Driver
}

func (d queryCacheTestDriver) Open(ctx context.Context, connectorName, instanceID string, config map[string]any, st *storage.Client, ac *activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	h, err := d.Driver.Open(ctx, connectorName, instanceID, config, st, ac, logger)
	if err != nil {
		return nil, err
	}
	return queryCacheTestHandle{Handle: h, name: connectorName}, nil
}

type queryCacheTestHandle struct {
	drivers.Handle
	name string
}

func (h queryCacheTestHandle) AsOLAP(instanceID string) (drivers.OLAPStore, bool) {
	olap, ok := h.Handle.AsOLAP(instanceID)
	if !ok {
		return nil, false
	}
	return queryCacheTestOLAP{OLAPStore: olap, name: h.name}, true
}

type queryCacheTestOLAP struct {
	drivers.OLAPStore
	name string
}

func (o queryCacheTestOLAP) OAuthConfig() *drivers.OAuthConfig {
	if o.name == "dbx" {
		return &drivers.OAuthConfig{}
	}
	return nil
}

func (o queryCacheTestOLAP) Query(ctx context.Context, stmt *drivers.Statement) (*drivers.Result, error) {
	olapQueries.mu.Lock()
	olapQueries.counts[o.name]++
	olapQueries.mu.Unlock()
	return o.OLAPStore.Query(ctx, stmt)
}
