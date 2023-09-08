package duckdb

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	cfg, err := newConfig(map[string]any{})
	require.NoError(t, err)
	require.Equal(t, "", cfg.DSN)
	require.Equal(t, 1, cfg.PoolSize)

	cfg, err = newConfig(map[string]any{"dsn": "path/to/duck.db"})
	require.NoError(t, err)
	require.Equal(t, "path/to/duck.db", cfg.DSN)
	require.Equal(t, "path/to/duck.db", cfg.DBFilePath)
	require.Equal(t, 1, cfg.PoolSize)

	cfg, err = newConfig(map[string]any{"dsn": "path/to/duck.db", "pool_size": 10})
	require.NoError(t, err)
	require.Equal(t, "path/to/duck.db", cfg.DSN)
	require.Equal(t, "path/to/duck.db", cfg.DBFilePath)
	require.Equal(t, 10, cfg.PoolSize)

	cfg, err = newConfig(map[string]any{"dsn": "path/to/duck.db", "pool_size": "10"})
	require.NoError(t, err)
	require.Equal(t, 10, cfg.PoolSize)

	cfg, err = newConfig(map[string]any{"dsn": "path/to/duck.db?rill_pool_size=4", "pool_size": "10"})
	require.NoError(t, err)
	require.Equal(t, 4, cfg.PoolSize)

	cfg, err = newConfig(map[string]any{"dsn": "path/to/duck.db?rill_pool_size=10"})
	require.NoError(t, err)
	require.Equal(t, "path/to/duck.db", cfg.DSN)
	require.Equal(t, "path/to/duck.db", cfg.DBFilePath)
	require.Equal(t, 10, cfg.PoolSize)

	cfg, err = newConfig(map[string]any{"dsn": "path/to/duck.db?max_memory=4GB&rill_pool_size=10"})
	require.NoError(t, err)
	require.Equal(t, "path/to/duck.db?max_memory=4GB", cfg.DSN)
	require.Equal(t, 10, cfg.PoolSize)
	require.Equal(t, "path/to/duck.db", cfg.DBFilePath)

	_, err = newConfig(map[string]any{"dsn": "path/to/duck.db?max_memory=4GB", "pool_size": "abc"})
	require.Error(t, err)

	_, err = newConfig(map[string]any{"dsn": "path/to/duck.db?max_memory=4GB", "pool_size": 0})
	require.Error(t, err)

	cfg, err = newConfig(map[string]any{"dsn": "duck.db"})
	require.NoError(t, err)
	require.Equal(t, "duck.db", cfg.DBFilePath)

	cfg, err = newConfig(map[string]any{"dsn": "duck.db?rill_pool_size=10"})
	require.NoError(t, err)
	require.Equal(t, "duck.db", cfg.DBFilePath)
}
