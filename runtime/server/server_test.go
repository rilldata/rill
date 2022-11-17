package server

import (
	"context"
	"testing"

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
		ConnectionCacheSize: 100,
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
