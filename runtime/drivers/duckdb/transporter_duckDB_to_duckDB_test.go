package duckdb

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	_ "github.com/marcboeker/go-duckdb"
	"github.com/rilldata/rill/runtime/drivers"
	activity "github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestDuckDBToDuckDBTransfer(t *testing.T) {
	tempDir := t.TempDir()
	dbFile := filepath.Join(tempDir, "transfer.db")
	db, err := sql.Open("duckdb", dbFile)
	require.NoError(t, err)

	_, err = db.ExecContext(context.Background(), "CREATE TABLE foo(bar VARCHAR, baz INTEGER)")
	require.NoError(t, err)

	_, err = db.ExecContext(context.Background(), "INSERT INTO foo VALUES ('a', 1), ('a', 2), ('b', 3), ('c', 4)")
	require.NoError(t, err)
	require.NoError(t, db.Close())

	to, err := Driver{}.Open("default", map[string]any{}, storage.MustNew(tempDir, nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)

	tr := newDuckDBToDuckDB(to.(*connection), "duckdb", zap.NewNop())

	// transfer once
	err = tr.Transfer(context.Background(), map[string]any{"sql": "SELECT * FROM foo", "db": dbFile}, map[string]any{"table": "test"}, &drivers.TransferOptions{})
	require.NoError(t, err)

	rows, err := to.(*connection).Execute(context.Background(), &drivers.Statement{Query: "SELECT COUNT(*) FROM test"})
	require.NoError(t, err)

	var count int
	rows.Next()
	require.NoError(t, rows.Scan(&count))
	require.Equal(t, 4, count)
	require.NoError(t, rows.Close())

	// transfer again
	err = tr.Transfer(context.Background(), map[string]any{"sql": "SELECT * FROM foo", "db": dbFile}, map[string]any{"table": "test"}, &drivers.TransferOptions{})
	require.NoError(t, err)

	rows, err = to.(*connection).Execute(context.Background(), &drivers.Statement{Query: "SELECT COUNT(*) FROM test"})
	require.NoError(t, err)

	rows.Next()
	require.NoError(t, rows.Scan(&count))
	require.Equal(t, 4, count)
	require.NoError(t, rows.Close())
}
