package server

import (
	"context"
	"github.com/rilldata/rill/runtime/api"
	"github.com/rilldata/rill/runtime/drivers"
)

func GetTestServer() (*Server, string, error) {
	metastore, err := drivers.Open("sqlite", "file:rill?mode=memory&cache=shared")
	if err != nil {
		return nil, "", err
	}
	err = metastore.Migrate(context.Background())
	if err != nil {
		return nil, "", err
	}

	server, err := NewServer(&ServerOptions{
		ConnectionCacheSize: 100,
	}, metastore, nil)
	if err != nil {
		return nil, "", err
	}

	resp, err := server.CreateInstance(context.Background(), &api.CreateInstanceRequest{
		Driver:       "duckdb",
		Dsn:          "",
		Exposed:      true,
		EmbedCatalog: true,
	})
	if err != nil {
		return nil, "", err
	}

	return server, resp.InstanceId, nil
}
