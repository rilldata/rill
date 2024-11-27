package duckdb

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	cfg, err := newConfig(map[string]any{})
	require.NoError(t, err)
	require.Equal(t, 2, cfg.PoolSize)

	cfg, err = newConfig(map[string]any{"dsn": ":memory:?memory_limit=2GB"})
	require.NoError(t, err)
	require.Equal(t, 2, cfg.PoolSize)

	cfg, err = newConfig(map[string]any{"dsn": "", "memory_limit_gb": "1", "cpu": 2})
	require.NoError(t, err)
	require.Equal(t, "1", cfg.readSettings()["threads"])
	require.Equal(t, "1", cfg.readSettings()["threads"])
	require.Equal(t, 2, cfg.PoolSize)

	cfg, err = newConfig(map[string]any{"data_dir": "path/to"})
	require.NoError(t, err)
	require.Subset(t, cfg.writeSettings(), map[string]string{"custom_user_agent": "rill"})
	require.Equal(t, 2, cfg.PoolSize)

	cfg, err = newConfig(map[string]any{"data_dir": "path/to", "pool_size": 10})
	require.NoError(t, err)
	require.Equal(t, 10, cfg.PoolSize)

	_, err = newConfig(map[string]any{"dsn": "path/to/duck.db?max_memory=4GB", "pool_size": "abc"})
	require.Error(t, err)

	_, err = newConfig(map[string]any{"dsn": "duck.db"})
	require.NoError(t, err)

	_, err = newConfig(map[string]any{"dsn": "duck.db?rill_pool_size=10"})
	require.NoError(t, err)

	cfg, err = newConfig(map[string]any{"dsn": "duck.db", "memory_limit_gb": "8", "cpu": "2"})
	require.NoError(t, err)
	require.Equal(t, "1", cfg.readSettings()["threads"])
	require.Equal(t, "1", cfg.writeSettings()["threads"])
	require.Equal(t, "4", cfg.readSettings()["max_memory"])
	require.Equal(t, "4", cfg.writeSettings()["max_memory"])
	require.Equal(t, 2, cfg.PoolSize)
}
