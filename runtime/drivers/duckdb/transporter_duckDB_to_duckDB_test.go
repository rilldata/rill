package duckdb

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/rilldata/rill/runtime/drivers"
	activity "github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestDuckDBToDuckDBTransfer(t *testing.T) {
	tempDir := t.TempDir()
	conn, err := Driver{}.Open("default", map[string]any{"path": fmt.Sprintf("%s.db", filepath.Join(tempDir, "tranfser"))}, activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)

	olap, ok := conn.AsOLAP("")
	require.True(t, ok)

	err = olap.Exec(context.Background(), &drivers.Statement{
		Query: "CREATE TABLE foo(bar VARCHAR, baz INTEGER)",
	})
	require.NoError(t, err)

	err = olap.Exec(context.Background(), &drivers.Statement{
		Query: "INSERT INTO foo VALUES ('a', 1), ('a', 2), ('b', 3), ('c', 4)",
	})
	require.NoError(t, err)
	require.NoError(t, conn.Close())

	to, err := Driver{}.Open("default", map[string]any{"dsn": ":memory:"}, activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)

	olap, _ = to.AsOLAP("")

	tr := NewDuckDBToDuckDB(olap, zap.NewNop())

	// transfer once
	err = tr.Transfer(context.Background(), map[string]any{"sql": "SELECT * FROM foo", "db": filepath.Join(tempDir, "tranfser.db")}, map[string]any{"table": "test"}, &drivers.TransferOptions{Progress: drivers.NoOpProgress{}})
	require.NoError(t, err)

	rows, err := olap.Execute(context.Background(), &drivers.Statement{Query: "SELECT COUNT(*) FROM test"})
	require.NoError(t, err)

	var count int
	rows.Next()
	require.NoError(t, rows.Scan(&count))
	require.Equal(t, 4, count)
	require.NoError(t, rows.Close())

	// transfer again
	err = tr.Transfer(context.Background(), map[string]any{"sql": "SELECT * FROM foo", "db": filepath.Join(tempDir, "tranfser.db")}, map[string]any{"table": "test"}, &drivers.TransferOptions{Progress: drivers.NoOpProgress{}})
	require.NoError(t, err)

	rows, err = olap.Execute(context.Background(), &drivers.Statement{Query: "SELECT COUNT(*) FROM test"})
	require.NoError(t, err)

	rows.Next()
	require.NoError(t, rows.Scan(&count))
	require.Equal(t, 4, count)
	require.NoError(t, rows.Close())
}
