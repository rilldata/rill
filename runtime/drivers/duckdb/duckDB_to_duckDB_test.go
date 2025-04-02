package duckdb

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/rilldata/rill/runtime/drivers"
	activity "github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	_ "github.com/marcboeker/go-duckdb"
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

	duckDB, err := drivers.Open("duckdb", "default", map[string]any{"data_dir": t.TempDir()}, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)

	opts := &drivers.ModelExecutorOptions{
		InputHandle:     duckDB,
		InputConnector:  "duckdb",
		OutputHandle:    duckDB,
		OutputConnector: "duckdb",
		Env: &drivers.ModelEnv{
			AllowHostAccess: false,
			StageChanges:    true,
		},
		PreliminaryInputProperties: map[string]any{
			"sql": "SELECT * FROM foo;",
			"db":  dbFile,
		},
		PreliminaryOutputProperties: map[string]any{
			"table": "sink",
		},
	}

	me, ok := duckDB.AsModelExecutor("default", opts)
	require.True(t, ok)

	execOpts := &drivers.ModelExecuteOptions{
		ModelExecutorOptions: opts,
		InputProperties:      opts.PreliminaryInputProperties,
		OutputProperties:     opts.PreliminaryOutputProperties,
	}
	_, err = me.Execute(context.Background(), execOpts)
	require.NoError(t, err)

	rows, err := duckDB.(*connection).Execute(context.Background(), &drivers.Statement{Query: "SELECT COUNT(*) FROM sink"})
	require.NoError(t, err)

	var count int
	rows.Next()
	require.NoError(t, rows.Scan(&count))
	require.Equal(t, 4, count)
	require.NoError(t, rows.Close())

	// transfer again
	_, err = me.Execute(context.Background(), execOpts)
	require.NoError(t, err)

	rows, err = duckDB.(*connection).Execute(context.Background(), &drivers.Statement{Query: "SELECT COUNT(*) FROM sink"})
	require.NoError(t, err)

	rows.Next()
	require.NoError(t, rows.Scan(&count))
	require.Equal(t, 4, count)
	require.NoError(t, rows.Close())
}
