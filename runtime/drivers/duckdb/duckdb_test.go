package duckdb

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestOpenDrop(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tmp.db")
	walpath := path + ".wal"
	dsn := path + "?rill_pool_size=2"

	handle, err := Driver{}.Open(dsn, zap.NewNop())
	require.NoError(t, err)

	olap, ok := handle.OLAPStore()
	require.True(t, ok)

	err = olap.Exec(context.Background(), &drivers.Statement{Query: "CREATE TABLE foo (bar INTEGER)"})
	require.NoError(t, err)

	err = handle.Close()
	require.NoError(t, err)
	require.FileExists(t, path)
	require.FileExists(t, walpath)

	err = Driver{}.Drop(dsn, zap.NewNop())
	require.NoError(t, err)
	require.NoFileExists(t, path)
	require.NoFileExists(t, walpath)
}
