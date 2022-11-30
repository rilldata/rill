package server

import (
	"context"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
)

func getTestServer(t *testing.T) (*Server, string) {
	metastore, err := drivers.Open("sqlite", "file:rill?mode=memory&cache=shared")
	require.NoError(t, err)

	err = metastore.Migrate(context.Background())
	require.NoError(t, err)

	server, err := NewServer(&ServerOptions{
		ConnectionCacheSize: 100,
	}, metastore, nil)
	require.NoError(t, err)

	resp, err := server.CreateInstance(context.Background(), &runtimev1.CreateInstanceRequest{
		OlapDriver:   "duckdb",
		OlapDsn:      "",
		RepoDriver:   "file",
		RepoDsn:      t.TempDir(),
		EmbedCatalog: true,
	})
	require.NoError(t, err)
	require.NotEmpty(t, resp.Instance.InstanceId)

	return server, resp.Instance.InstanceId
}

func getTestServerWithData(t *testing.T) (*Server, string) {
	server, instanceId := getTestServer(t)

	_, err := server.QueryDirect(context.Background(), &runtimev1.QueryDirectRequest{
		InstanceId: instanceId,
		Sql: `CREATE TABLE test AS (
			SELECT 'abc' AS col, 1 AS val, TIMESTAMP '2022-11-01 00:00:00' AS times, DATE '2007-04-01' AS dates
			UNION ALL 
			SELECT 'def' AS col, 5 AS val, TIMESTAMP '2022-11-02 00:00:00' AS times, DATE '2009-06-01' AS dates
			UNION ALL 
			SELECT 'abc' AS col, 3 AS val, TIMESTAMP '2022-11-03 00:00:00' AS times, DATE '2010-04-11' AS dates
			UNION ALL 
			SELECT null AS col, 1 AS val, TIMESTAMP '2022-11-03 00:00:00' AS times, DATE '2010-11-21' AS dates
			UNION ALL 
			SELECT 12 AS col, 1 AS val, TIMESTAMP '2022-11-03 00:00:00' AS times, DATE '2011-06-30' AS dates
			)`,
		Args: nil,
	})
	require.NoError(t, err)

	return server, instanceId
}
