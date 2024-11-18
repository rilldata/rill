package rduckdb

import (
	"context"
	"io"
	"log/slog"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"gocloud.dev/blob/fileblob"
)

func TestDB(t *testing.T) {
	db, _, _ := prepareDB(t)
	ctx := context.Background()
	// create table
	err := db.CreateTableAsSelect(ctx, "test", "SELECT 1 AS id, 'India' AS country", &CreateTableOptions{})
	require.NoError(t, err)

	// query table
	var (
		id      int
		country string
	)
	conn, release, err := db.AcquireReadConnection(ctx)
	require.NoError(t, err)

	conn.QueryRowContext(ctx, "SELECT id, country FROM test").Scan(&id, &country)
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
	err = db.MutateTable(ctx, "test2", func(ctx context.Context, conn *sqlx.Conn) error {
		_, err := conn.ExecContext(ctx, "INSERT INTO test2 (id, country) VALUES (2, 'USA')")
		return err
	})
	require.NoError(t, err)

	// query table
	conn, release, err = db.AcquireReadConnection(ctx)
	require.NoError(t, err)
	err = conn.QueryRowxContext(ctx, "SELECT id, country FROM test2 where id = 2").Scan(&id, &country)
	require.NoError(t, err)
	require.Equal(t, 2, id)
	require.Equal(t, "USA", country)
	require.NoError(t, release())

	// Add column
	db.MutateTable(ctx, "test2", func(ctx context.Context, conn *sqlx.Conn) error {
		_, err := conn.ExecContext(ctx, "ALTER TABLE test2 ADD COLUMN city TEXT")
		return err
	})

	// drop table
	err = db.DropTable(ctx, "test2")
	require.NoError(t, err)
	require.NoError(t, db.Close())
}

func TestCreateTable(t *testing.T) {
	db, _, _ := prepareDB(t)
	ctx := context.Background()
	err := db.CreateTableAsSelect(ctx, "test", "SELECT 1 AS id, 'India' AS country", &CreateTableOptions{})
	require.NoError(t, err)
	verifyTable(t, db, "SELECT id, country FROM test", []testData{{ID: 1, Country: "India"}})

	// replace table
	err = db.CreateTableAsSelect(ctx, "test", "SELECT 2 AS id, 'USA' AS country", &CreateTableOptions{})
	require.NoError(t, err)
	verifyTable(t, db, "SELECT id, country FROM test", []testData{{ID: 2, Country: "USA"}})

	// create another table that ingests from first table
	err = db.CreateTableAsSelect(ctx, "test2", "SELECT * FROM test", &CreateTableOptions{})
	require.NoError(t, err)
	verifyTable(t, db, "SELECT id, country FROM test2", []testData{{ID: 2, Country: "USA"}})

	// create view
	err = db.CreateTableAsSelect(ctx, "test_view", "SELECT * FROM test", &CreateTableOptions{View: true})
	require.NoError(t, err)
	verifyTable(t, db, "SELECT id, country FROM test_view", []testData{{ID: 2, Country: "USA"}})

	// view on top of view
	err = db.CreateTableAsSelect(ctx, "pest_view", "SELECT * FROM test_view", &CreateTableOptions{View: true})
	require.NoError(t, err)
	verifyTable(t, db, "SELECT id, country FROM pest_view", []testData{{ID: 2, Country: "USA"}})

	// replace underlying table
	err = db.CreateTableAsSelect(ctx, "test", "SELECT 3 AS id, 'UK' AS country", &CreateTableOptions{})
	require.NoError(t, err)
	verifyTable(t, db, "SELECT id, country FROM test", []testData{{ID: 3, Country: "UK"}})

	// view should reflect the change
	verifyTable(t, db, "SELECT id, country FROM test_view", []testData{{ID: 3, Country: "UK"}})

	// create table that was previously view
	err = db.CreateTableAsSelect(ctx, "test_view", "SELECT 1 AS id, 'India' AS country", &CreateTableOptions{})
	require.NoError(t, err)
	verifyTable(t, db, "SELECT id, country FROM test_view", []testData{{ID: 1, Country: "India"}})

	// create view that was previously table
	err = db.CreateTableAsSelect(ctx, "test", "SELECT * FROM test_view", &CreateTableOptions{View: true})
	require.NoError(t, err)
	verifyTable(t, db, "SELECT id, country FROM test", []testData{{ID: 1, Country: "India"}})
	require.NoError(t, db.Close())
}

func TestDropTable(t *testing.T) {
	db, _, _ := prepareDB(t)
	ctx := context.Background()

	// create table
	err := db.CreateTableAsSelect(ctx, "test", "SELECT 1 AS id, 'India' AS country", &CreateTableOptions{})
	require.NoError(t, err)
	verifyTable(t, db, "SELECT id, country FROM test", []testData{{ID: 1, Country: "India"}})

	// create view
	err = db.CreateTableAsSelect(ctx, "test_view", "SELECT * FROM test", &CreateTableOptions{View: true})
	require.NoError(t, err)
	verifyTable(t, db, "SELECT id, country FROM test_view", []testData{{ID: 1, Country: "India"}})

	// drop view
	err = db.DropTable(ctx, "test_view")
	require.NoError(t, err)

	// verify table data is still there
	verifyTable(t, db, "SELECT id, country FROM test", []testData{{ID: 1, Country: "India"}})

	// drop table
	err = db.DropTable(ctx, "test")
	require.NoError(t, err)
	require.NoError(t, db.Close())
}

func TestMutateTable(t *testing.T) {
	db, _, _ := prepareDB(t)
	ctx := context.Background()

	// create table
	err := db.CreateTableAsSelect(ctx, "test", "SELECT 1 AS id, 'Delhi' AS city", &CreateTableOptions{})
	require.NoError(t, err)

	// insert into table
	err = db.MutateTable(ctx, "test", func(ctx context.Context, conn *sqlx.Conn) error {
		_, err := conn.ExecContext(ctx, "INSERT INTO test (id, city) VALUES (2, 'NY')")
		return err
	})
	require.NoError(t, err)
	verifyTable(t, db, "SELECT id, city FROM test", []testData{{ID: 1, City: "Delhi"}, {ID: 2, City: "NY"}})

	// add column and update existing entries in parallel query existing table
	alterDone := make(chan struct{})
	queryDone := make(chan struct{})
	testDone := make(chan struct{})

	go func() {
		db.MutateTable(ctx, "test", func(ctx context.Context, conn *sqlx.Conn) error {
			_, err := conn.ExecContext(ctx, "ALTER TABLE test ADD COLUMN country TEXT")
			require.NoError(t, err)
			_, err = conn.ExecContext(ctx, "UPDATE test SET country = 'USA' WHERE id = 2")
			require.NoError(t, err)
			_, err = conn.ExecContext(ctx, "UPDATE test SET country = 'India' WHERE id = 1")
			require.NoError(t, err)

			close(alterDone)
			<-queryDone
			return nil
		})
		close(testDone)
	}()

	go func() {
		<-alterDone
		verifyTable(t, db, "SELECT * FROM test", []testData{{ID: 1, City: "Delhi"}, {ID: 2, City: "NY"}})
		close(queryDone)
	}()

	<-testDone
	verifyTable(t, db, "SELECT * FROM test", []testData{{ID: 1, City: "Delhi", Country: "India"}, {ID: 2, City: "NY", Country: "USA"}})
	require.NoError(t, db.Close())
}

func TestResetLocal(t *testing.T) {
	db, localDir, remoteDir := prepareDB(t)
	ctx := context.Background()

	// create table
	err := db.CreateTableAsSelect(ctx, "test", "SELECT 1 AS id, 'India' AS country", &CreateTableOptions{})
	require.NoError(t, err)
	verifyTable(t, db, "SELECT id, country FROM test", []testData{{ID: 1, Country: "India"}})

	// reset local
	require.NoError(t, db.Close())
	require.NoError(t, os.RemoveAll(localDir))

	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	bucket, err := fileblob.OpenBucket(remoteDir, nil)
	require.NoError(t, err)
	db, err = NewDB(ctx, &DBOptions{
		LocalPath:     localDir,
		Remote:        bucket,
		ReadSettings:  map[string]string{"memory_limit": "2GB", "threads": "1"},
		WriteSettings: map[string]string{"memory_limit": "2GB", "threads": "1"},
		InitQueries:   []string{"SET autoinstall_known_extensions=true", "SET autoload_known_extensions=true"},
		Logger:        logger,
	})
	require.NoError(t, err)
	verifyTable(t, db, "SELECT id, country FROM test", []testData{{ID: 1, Country: "India"}})
}

func prepareDB(t *testing.T) (db DB, localDir, remoteDir string) {
	localDir = t.TempDir()
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	remoteDir = t.TempDir()
	bucket, err := fileblob.OpenBucket(remoteDir, nil)
	require.NoError(t, err)
	db, err = NewDB(ctx, &DBOptions{
		LocalPath:     localDir,
		Remote:        bucket,
		ReadSettings:  map[string]string{"memory_limit": "2GB", "threads": "1"},
		WriteSettings: map[string]string{"memory_limit": "2GB", "threads": "1"},
		InitQueries:   []string{"SET autoinstall_known_extensions=true", "SET autoload_known_extensions=true"},
		Logger:        logger,
	})
	require.NoError(t, err)
	return
}

func verifyTable(t *testing.T, db DB, query string, data []testData) {
	ctx := context.Background()
	conn, release, err := db.AcquireReadConnection(ctx)
	require.NoError(t, err)
	defer release()

	var scannedData []testData
	err = conn.SelectContext(ctx, &scannedData, query)
	require.NoError(t, err)
	require.Equal(t, data, scannedData)
}

type testData struct {
	ID      int    `db:"id"`
	Country string `db:"country"`
	City    string `db:"city"`
}
