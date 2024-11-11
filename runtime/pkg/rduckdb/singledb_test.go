package rduckdb

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSingleDB_test(t *testing.T) {
	ctx := context.Background()
	db, err := NewSingleDB(ctx, &SingleDBOptions{
		DSN: "",
	})
	require.NoError(t, err)

	// create table
	rw, release, err := db.AcquireWriteConnection(ctx)
	require.NoError(t, err)

	err = rw.CreateTableAsSelect(ctx, "test-2", "SELECT 1 AS id, 'India' AS country", &CreateTableOptions{})
	require.NoError(t, err)

	// rename table
	err = rw.RenameTable(ctx, "test-2", "test")
	require.NoError(t, err)

	// insert into table
	err = rw.InsertTableAsSelect(ctx, "test", "SELECT 2 AS id, 'USA' AS country", &InsertTableOptions{
		Strategy: IncrementalStrategyAppend,
	})
	require.NoError(t, err)

	// add column
	err = rw.AddTableColumn(ctx, "test", "currency_score", "INT")
	require.NoError(t, err)

	// alter column
	err = rw.AlterTableColumn(ctx, "test", "currency_score", "FLOAT")
	require.NoError(t, err)
	require.NoError(t, release())

	// select from table
	conn, release, err := db.AcquireReadConnection(ctx)
	require.NoError(t, err)

	var (
		id            int
		country       string
		currencyScore sql.NullFloat64
	)

	err = conn.Connx().QueryRowxContext(ctx, "SELECT id, country, currency_score FROM test WHERE id = 2").Scan(&id, &country, &currencyScore)
	require.NoError(t, err)
	require.Equal(t, 2, id)
	require.Equal(t, "USA", country)
	require.Equal(t, false, currencyScore.Valid)

	err = release()
	require.NoError(t, err)

	// drop table
	err = db.DropTable(ctx, "test")
	require.NoError(t, err)
}

func TestSingleDB_testRenameExisting(t *testing.T) {
	ctx := context.Background()
	db, err := NewSingleDB(ctx, &SingleDBOptions{
		DSN: "",
	})
	require.NoError(t, err)

	// create table
	err = db.CreateTableAsSelect(ctx, "test-2", "SELECT 1 AS id, 'India' AS country", nil)
	require.NoError(t, err)

	// create another table
	err = db.CreateTableAsSelect(ctx, "test-3", "SELECT 2 AS id, 'USA' AS country", nil)
	require.NoError(t, err)

	// rename table
	err = db.RenameTable(ctx, "test-2", "test-3")
	require.NoError(t, err)

	// select from table
	conn, release, err := db.AcquireReadConnection(ctx)
	require.NoError(t, err)

	var (
		id      int
		country string
	)

	err = conn.Connx().QueryRowxContext(ctx, "SELECT id, country FROM \"test-3\" WHERE id = 1").Scan(&id, &country)
	require.NoError(t, err)
	require.Equal(t, 1, id)
	require.Equal(t, "India", country)

	err = release()
	require.NoError(t, err)
}
