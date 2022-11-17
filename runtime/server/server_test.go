package server

import (
	"context"
	"testing"
	"time"

	"github.com/rilldata/rill/runtime/api"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
)

func getTestServer(t *testing.T) (*Server, string, error) {
	metastore, err := drivers.Open("sqlite", "file:rill?mode=memory&cache=shared")
	require.NoError(t, err)

	err = metastore.Migrate(context.Background())
	require.NoError(t, err)

	server, err := NewServer(&ServerOptions{
		ConnectionCacheSize:  100,
		CatalogCacheSize:     100,
		CatalogCacheDuration: 10 * time.Second,
	}, metastore, nil)
	require.NoError(t, err)

	resp, err := server.CreateInstance(context.Background(), &api.CreateInstanceRequest{
		Driver:       "duckdb",
		Dsn:          "",
		Exposed:      true,
		EmbedCatalog: true,
	})
	require.NoError(t, err)
	require.NotEmpty(t, resp.InstanceId)

	return server, resp.InstanceId, nil
}

func getTestServerWithData(t *testing.T) (*Server, string, error) {
	server, instanceId, err := getTestServer(t)

	_, err = server.QueryDirect(context.Background(), &api.QueryDirectRequest{
		InstanceId: instanceId,
		Sql: `CREATE TABLE test AS (
			SELECT 'abc' AS col, 1 AS val, TIMESTAMP '2022-11-01 00:00:00' AS times 
			UNION ALL 
			SELECT 'def' AS col, 5 AS val, TIMESTAMP '2022-11-02 00:00:00' AS times
			UNION ALL 
			SELECT 'abc' AS col, 3 AS val, TIMESTAMP '2022-11-03 00:00:00' AS times
			UNION ALL 
			SELECT null AS col, 1 AS val, TIMESTAMP '2022-11-03 00:00:00' AS times
			)`,
		Args: nil,
	})
	require.NoError(t, err)

	return server, instanceId, nil
}
