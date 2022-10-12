package test_utils

import (
	"github.com/rilldata/rill/runtime/drivers"
	_ "github.com/rilldata/rill/runtime/drivers/duckdb"
	_ "github.com/rilldata/rill/runtime/drivers/sqlite"
	"github.com/rilldata/rill/runtime/server"
)

func GetTestServer(dsn string) (*server.Server, error) {
	metastore, err := drivers.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	return server.NewServer(nil, metastore, nil)
}
