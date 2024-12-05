package duckdb

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/rilldata/rill/runtime/drivers"
	activity "github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestConfig(t *testing.T) {
	cfg, err := newConfig(map[string]any{}, "")
	require.NoError(t, err)
	require.Equal(t, 2, cfg.PoolSize)

	cfg, err = newConfig(map[string]any{"dsn": "", "cpu": 2}, "")
	require.NoError(t, err)
	require.Equal(t, "2", cfg.readSettings()["threads"])
	require.Equal(t, "", cfg.writeSettings()["threads"])
	require.Equal(t, 2, cfg.PoolSize)

	cfg, err = newConfig(map[string]any{}, "path/to")
	require.NoError(t, err)
	require.Subset(t, cfg.writeSettings(), map[string]string{"custom_user_agent": "rill"})
	require.Equal(t, 2, cfg.PoolSize)

	cfg, err = newConfig(map[string]any{"pool_size": 10}, "path/to")
	require.NoError(t, err)
	require.Equal(t, 10, cfg.PoolSize)

	cfg, err = newConfig(map[string]any{"pool_size": "10"}, "path/to")
	require.NoError(t, err)
	require.Equal(t, 10, cfg.PoolSize)

	cfg, err = newConfig(map[string]any{"dsn": "?rill_pool_size=4", "pool_size": "10"}, "path/to")
	require.NoError(t, err)
	require.Equal(t, 4, cfg.PoolSize)

	cfg, err = newConfig(map[string]any{"dsn": "path/to/duck.db?rill_pool_size=10"}, "path/to")
	require.NoError(t, err)
	// require.Equal(t, "path/to/duck.db?custom_user_agent=rill", cfg.DSN)
	// require.Equal(t, "path/to/duck.db", cfg.DBFilePath)
	require.Equal(t, 10, cfg.PoolSize)

	cfg, err = newConfig(map[string]any{"dsn": "path/to/duck.db?max_memory=4GB&rill_pool_size=10"}, "path/to")
	require.NoError(t, err)
	// require.Equal(t, "path/to/duck.db?custom_user_agent=rill&max_memory=4GB", cfg.DSN)
	require.Equal(t, 10, cfg.PoolSize)
	// require.Equal(t, "path/to/duck.db", cfg.DBFilePath)

	_, err = newConfig(map[string]any{"dsn": "path/to/duck.db?max_memory=4GB", "pool_size": "abc"}, "path/to")
	require.Error(t, err)

	cfg, err = newConfig(map[string]any{"dsn": "duck.db"}, "path/to")
	require.NoError(t, err)

	cfg, err = newConfig(map[string]any{"dsn": "duck.db?rill_pool_size=10"}, "path/to")
	require.NoError(t, err)

	cfg, err = newConfig(map[string]any{"dsn": "duck.db", "memory_limit_gb": "8", "cpu": "2"}, "path/to")
	require.NoError(t, err)
	require.Equal(t, "2", cfg.readSettings()["threads"])
	require.Equal(t, "", cfg.writeSettings()["threads"])
	require.Equal(t, "8GB", cfg.readSettings()["max_memory"])
	require.Equal(t, "", cfg.writeSettings()["max_memory"])
	require.Equal(t, 2, cfg.PoolSize)

	cfg, err = newConfig(map[string]any{"dsn": "duck.db?max_memory=2GB&rill_pool_size=4"}, "path/to")
	require.NoError(t, err)
	// require.Equal(t, "duck.db", cfg.DBFilePath)
	// require.Equal(t, "duck.db?custom_user_agent=rill&max_memory=2GB", cfg.DSN)
	require.Equal(t, 4, cfg.PoolSize)
}

func Test_specialCharInPath(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "let's t@st \"weird\" dirs")
	err := os.Mkdir(path, fs.ModePerm)
	require.NoError(t, err)

	dbFile := filepath.Join(path, "st@g3's.db")
	conn, err := Driver{}.Open("default", map[string]any{"init_sql": fmt.Sprintf("ATTACH %s", safeSQLString(dbFile))}, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)

	olap, ok := conn.AsOLAP("")
	require.True(t, ok)

	res, err := olap.Execute(context.Background(), &drivers.Statement{Query: "SELECT 1"})
	require.NoError(t, err)
	require.NoError(t, res.Close())
	require.NoError(t, conn.Close())
}
