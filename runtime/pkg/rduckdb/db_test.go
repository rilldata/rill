package rduckdb

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDB(t *testing.T) {
	dir := t.TempDir()
	ctx := context.Background()
	db, err := NewDB(ctx, &DBOptions{
		LocalPath:     dir,
		Backup:        nil,
		ReadSettings:  map[string]string{"memory_limit": "2GB", "threads": "1"},
		WriteSettings: map[string]string{"memory_limit": "2GB", "threads": "1"},
		InitQueries:   []string{"SET autoinstall_known_extensions=true", "SET autoload_known_extensions=true"},
		Logger:        slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	require.NoError(t, err)

	// create table
	err = db.CreateTableAsSelect(ctx, "test", "SELECT 1 AS id, 'India' AS country", &CreateTableOptions{})
	require.NoError(t, err)

	// query table
	var (
		id      int
		country string
	)
	conn, release, err := db.AcquireReadConnection(ctx)
	require.NoError(t, err)
	err = conn.Connx().QueryRowxContext(ctx, "SELECT id, country FROM test").Scan(&id, &country)
	require.NoError(t, err)
	require.Equal(t, 1, id)
	require.Equal(t, "India", country)
	require.NoError(t, release())

	// rename table
	err = db.RenameTable(ctx, "test", "test2")
	require.NoError(t, err)

	// drop old table
	err = db.DropTable(ctx, "test")
	require.Error(t, err)

	// insert into table
	err = db.InsertTableAsSelect(ctx, "test2", "SELECT 2 AS id, 'US' AS country", &InsertTableOptions{
		Strategy: IncrementalStrategyAppend,
	})
	require.NoError(t, err)

	// merge into table
	err = db.InsertTableAsSelect(ctx, "test2", "SELECT 2 AS id, 'USA' AS country", &InsertTableOptions{
		Strategy:  IncrementalStrategyMerge,
		UniqueKey: []string{"id"},
	})
	require.NoError(t, err)

	// query table
	conn, release, err = db.AcquireReadConnection(ctx)
	require.NoError(t, err)
	err = conn.Connx().QueryRowxContext(ctx, "SELECT id, country FROM test2 where id = 2").Scan(&id, &country)
	require.NoError(t, err)
	require.Equal(t, 2, id)
	require.Equal(t, "USA", country)
	require.NoError(t, release())

	// Add column
	err = db.AddTableColumn(ctx, "test2", "city", "TEXT")
	require.NoError(t, err)

	// drop table
	err = db.DropTable(ctx, "test2")
	require.NoError(t, err)
}
