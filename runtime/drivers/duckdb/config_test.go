package duckdb

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	cfg, err := newConfig("")
	require.NoError(t, err)
	require.Equal(t, "", cfg.DSN)
	require.Equal(t, 1, cfg.PoolSize)

	cfg, err = newConfig("path/to/duck.db")
	require.NoError(t, err)
	require.Equal(t, "path/to/duck.db", cfg.DSN)
	require.Equal(t, "path/to/duck.db", cfg.DBFilePath)
	require.Equal(t, 1, cfg.PoolSize)

	cfg, err = newConfig("path/to/duck.db?rill_pool_size=10")
	require.NoError(t, err)
	require.Equal(t, "path/to/duck.db", cfg.DSN)
	require.Equal(t, "path/to/duck.db", cfg.DBFilePath)
	require.Equal(t, 10, cfg.PoolSize)

	cfg, err = newConfig("path/to/duck.db?rill_pool_size=10&hello=world")
	require.NoError(t, err)
	require.Equal(t, "path/to/duck.db?hello=world", cfg.DSN)
	require.Equal(t, 10, cfg.PoolSize)
	require.Equal(t, "path/to/duck.db", cfg.DBFilePath)

	cfg, err = newConfig("path/to/duck.db?rill_pool_size=abc&hello=world")
	require.Error(t, err)

	cfg, err = newConfig("path/to/duck.db?rill_pool_size=0&hello=world")
	require.Error(t, err)

	cfg, err = newConfig("duck.db")
	require.NoError(t, err)
	require.Equal(t, "duck.db", cfg.DBFilePath)

	cfg, err = newConfig("duck.db?rill_pool_size=10")
	require.NoError(t, err)
	require.Equal(t, "duck.db", cfg.DBFilePath)
}
