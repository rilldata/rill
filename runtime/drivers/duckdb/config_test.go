package duckdb

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/rilldata/rill/runtime/drivers"
	activity "github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestConfig(t *testing.T) {
	cfg, err := newConfig(map[string]any{})
	require.NoError(t, err)
	require.Equal(t, "?custom_user_agent=rill", cfg.DSN)
	require.Equal(t, 2, cfg.PoolSize)

	cfg, err = newConfig(map[string]any{"dsn": ":memory:?memory_limit=2GB"})
	require.NoError(t, err)
	require.Equal(t, "?custom_user_agent=rill&memory_limit=2GB", cfg.DSN)
	require.Equal(t, 2, cfg.PoolSize)

	cfg, err = newConfig(map[string]any{"dsn": "", "memory_limit_gb": "1", "cpu": 2})
	require.NoError(t, err)
	require.Equal(t, "?custom_user_agent=rill&max_memory=1GB&threads=2", cfg.DSN)
	require.Equal(t, 2, cfg.PoolSize)
	require.Equal(t, true, cfg.ExtTableStorage)

	cfg, err = newConfig(map[string]any{"data_dir": "path/to"})
	require.NoError(t, err)
	require.Subset(t, cfg.WriteSettings, map[string]string{"custom_user_agent": "rill"})
	require.Equal(t, 2, cfg.PoolSize)

	cfg, err = newConfig(map[string]any{"data_dir": "path/to", "pool_size": 10})
	require.NoError(t, err)
	require.Subset(t, cfg.WriteSettings, map[string]string{"custom_user_agent": "rill"})
	require.Equal(t, 10, cfg.PoolSize)

	cfg, err = newConfig(map[string]any{"data_dir": "path/to", "pool_size": "10"})
	require.NoError(t, err)
	require.Equal(t, 10, cfg.PoolSize)

	cfg, err = newConfig(map[string]any{"data_dir": "path/to", "dsn": "?rill_pool_size=4", "pool_size": "10"})
	require.NoError(t, err)
	require.Equal(t, 4, cfg.PoolSize)

	cfg, err = newConfig(map[string]any{"dsn": "path/to/duck.db?rill_pool_size=10"})
	require.NoError(t, err)
	require.Equal(t, "path/to/duck.db?custom_user_agent=rill", cfg.DSN)
	require.Equal(t, 10, cfg.PoolSize)

	cfg, err = newConfig(map[string]any{"dsn": "path/to/duck.db?max_memory=4GB&rill_pool_size=10"})
	require.NoError(t, err)
	require.Equal(t, "path/to/duck.db?custom_user_agent=rill&max_memory=4GB", cfg.DSN)
	require.Equal(t, 10, cfg.PoolSize)

	_, err = newConfig(map[string]any{"dsn": "path/to/duck.db?max_memory=4GB", "pool_size": "abc"})
	require.Error(t, err)

	_, err = newConfig(map[string]any{"dsn": "duck.db"})
	require.NoError(t, err)

	_, err = newConfig(map[string]any{"dsn": "duck.db?rill_pool_size=10"})
	require.NoError(t, err)

	cfg, err = newConfig(map[string]any{"dsn": "duck.db", "memory_limit_gb": "8", "cpu": "2"})
	require.NoError(t, err)
	require.Equal(t, "duck.db?custom_user_agent=rill&max_memory=8GB&threads=2", cfg.DSN)
	require.Equal(t, 2, cfg.PoolSize)

	cfg, err = newConfig(map[string]any{"dsn": "duck.db?max_memory=2GB&rill_pool_size=4"})
	require.NoError(t, err)
	require.Equal(t, "duck.db?custom_user_agent=rill&max_memory=2GB", cfg.DSN)
	require.Equal(t, 4, cfg.PoolSize)
}

func Test_specialCharInPath(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "let's t@st \"weird\" dirs")
	err := os.Mkdir(path, fs.ModePerm)
	require.NoError(t, err)

	dbFile := filepath.Join(path, "st@g3's.db")
	conn, err := Driver{}.Open("default", map[string]any{"path": dbFile, "memory_limit_gb": "4", "cpu": "1", "external_table_storage": false}, activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	config := conn.(*connection).config
	require.Equal(t, dbFile+"?custom_user_agent=rill&max_memory=4GB&threads=1", config.DSN)
	require.Equal(t, 2, config.PoolSize)

	olap, ok := conn.AsOLAP("")
	require.True(t, ok)

	res, err := olap.Execute(context.Background(), &drivers.Statement{Query: "SELECT 1"})
	require.NoError(t, err)
	require.NoError(t, res.Close())
	require.NoError(t, conn.Close())
}
