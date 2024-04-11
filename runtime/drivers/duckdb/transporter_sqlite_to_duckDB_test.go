package duckdb

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/rilldata/rill/runtime/drivers"
	_ "github.com/rilldata/rill/runtime/drivers/sqlite"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	_ "modernc.org/sqlite"
)

func Test_sqliteToDuckDB_Transfer(t *testing.T) {
	tempDir := t.TempDir()

	dbPath := fmt.Sprintf("%s.db", tempDir)
	db, err := sql.Open("sqlite", dbPath)
	require.NoError(t, err)

	_, err = db.Exec(`
	drop table if exists t;
	create table t(i);
	insert into t values(42), (314);
	`)
	require.NoError(t, err)
	db.Close()

	to, err := drivers.Open("duckdb", map[string]any{"dsn": ":memory:"}, false, activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	olap, _ := to.AsOLAP("")

	tr := &duckDBToDuckDB{
		to:     olap,
		logger: zap.NewNop(),
	}
	query := fmt.Sprintf("SELECT * FROM sqlite_scan('%s', 't');", dbPath)
	err = tr.Transfer(context.Background(), map[string]any{"sql": query}, map[string]any{"table": "test"}, &drivers.TransferOptions{Progress: drivers.NoOpProgress{}})
	require.NoError(t, err)

	res, err := olap.Execute(context.Background(), &drivers.Statement{Query: "SELECT count(*) from test"})
	require.NoError(t, err)
	res.Next()
	var count int
	err = res.Scan(&count)
	require.NoError(t, err)
	require.NoError(t, res.Close())
	require.Equal(t, 2, count)
}
