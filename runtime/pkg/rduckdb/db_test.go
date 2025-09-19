package rduckdb

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gocloud.dev/blob/fileblob"
)

func TestDB(t *testing.T) {
	db, _, _ := prepareDB(t)
	ctx := context.Background()
	// create table
	_, err := db.CreateTableAsSelect(ctx, "test", "SELECT 1 AS id, 'India' AS country", &CreateTableOptions{})
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
	_, err = db.MutateTable(ctx, "test2", nil, func(ctx context.Context, conn *sqlx.Conn) error {
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
	_, err = db.MutateTable(ctx, "test2", nil, func(ctx context.Context, conn *sqlx.Conn) error {
		_, err := conn.ExecContext(ctx, "ALTER TABLE test2 ADD COLUMN city TEXT")
		return err
	})
	require.NoError(t, err)

	// drop table
	err = db.DropTable(ctx, "test2")
	require.NoError(t, err)
	require.NoError(t, db.Close())
}

func TestCreateTable(t *testing.T) {
	db, _, _ := prepareDB(t)
	ctx := context.Background()
	_, err := db.CreateTableAsSelect(ctx, "test", "SELECT 1 AS id, 'India' AS country", &CreateTableOptions{})
	require.NoError(t, err)
	verifyTable(t, db, "SELECT id, country FROM test", []testData{{ID: 1, Country: "India"}})

	// replace table
	_, err = db.CreateTableAsSelect(ctx, "test", "SELECT 2 AS id, 'USA' AS country", &CreateTableOptions{})
	require.NoError(t, err)
	verifyTable(t, db, "SELECT id, country FROM test", []testData{{ID: 2, Country: "USA"}})

	// create another table that ingests from first table
	_, err = db.CreateTableAsSelect(ctx, "test2", "SELECT * FROM test", &CreateTableOptions{})
	require.NoError(t, err)
	verifyTable(t, db, "SELECT id, country FROM test2", []testData{{ID: 2, Country: "USA"}})

	// create view
	_, err = db.CreateTableAsSelect(ctx, "test_view", "SELECT * FROM test", &CreateTableOptions{View: true})
	require.NoError(t, err)
	verifyTable(t, db, "SELECT id, country FROM test_view", []testData{{ID: 2, Country: "USA"}})

	// view on top of view
	_, err = db.CreateTableAsSelect(ctx, "pest_view", "SELECT * FROM test_view", &CreateTableOptions{View: true})
	require.NoError(t, err)
	verifyTable(t, db, "SELECT id, country FROM pest_view", []testData{{ID: 2, Country: "USA"}})

	// replace underlying table
	_, err = db.CreateTableAsSelect(ctx, "test", "SELECT 3 AS id, 'UK' AS country", &CreateTableOptions{})
	require.NoError(t, err)
	verifyTable(t, db, "SELECT id, country FROM test", []testData{{ID: 3, Country: "UK"}})

	// view should reflect the change
	verifyTable(t, db, "SELECT id, country FROM test_view", []testData{{ID: 3, Country: "UK"}})

	// create table that was previously view
	_, err = db.CreateTableAsSelect(ctx, "test_view", "SELECT 1 AS id, 'India' AS country", &CreateTableOptions{})
	require.NoError(t, err)
	verifyTable(t, db, "SELECT id, country FROM test_view", []testData{{ID: 1, Country: "India"}})

	// create view that was previously table
	_, err = db.CreateTableAsSelect(ctx, "test", "SELECT * FROM test_view", &CreateTableOptions{View: true})
	require.NoError(t, err)
	verifyTable(t, db, "SELECT id, country FROM test", []testData{{ID: 1, Country: "India"}})
	require.NoError(t, db.Close())
}

func TestDropTable(t *testing.T) {
	db, _, _ := prepareDB(t)
	ctx := context.Background()

	// create table
	_, err := db.CreateTableAsSelect(ctx, "test", "SELECT 1 AS id, 'India' AS country", &CreateTableOptions{})
	require.NoError(t, err)
	verifyTable(t, db, "SELECT id, country FROM test", []testData{{ID: 1, Country: "India"}})

	// create view
	_, err = db.CreateTableAsSelect(ctx, "test_view", "SELECT * FROM test", &CreateTableOptions{View: true})
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
	_, err := db.CreateTableAsSelect(ctx, "test", "SELECT 1 AS id, 'Delhi' AS city", &CreateTableOptions{})
	require.NoError(t, err)

	// create dependent view
	_, err = db.CreateTableAsSelect(ctx, "test_view", "SELECT * FROM test", &CreateTableOptions{View: true})
	require.NoError(t, err)

	// insert into table
	_, err = db.MutateTable(ctx, "test", nil, func(ctx context.Context, conn *sqlx.Conn) error {
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
		db.MutateTable(ctx, "test", nil, func(ctx context.Context, conn *sqlx.Conn) error {
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
	_, err := db.CreateTableAsSelect(ctx, "test", "SELECT 1 AS id, 'India' AS country", &CreateTableOptions{})
	require.NoError(t, err)
	verifyTable(t, db, "SELECT id, country FROM test", []testData{{ID: 1, Country: "India"}})

	// reset local
	require.NoError(t, db.Close())
	require.NoError(t, os.RemoveAll(localDir))

	logger := zap.NewNop()
	bucket, err := fileblob.OpenBucket(remoteDir, nil)
	require.NoError(t, err)
	db, err = NewDB(ctx, &DBOptions{
		LocalPath:      localDir,
		Remote:         bucket,
		MemoryLimitGB:  2,
		CPU:            1,
		ReadWriteRatio: 0.5,
		DBInitQueries:  []string{"SET autoinstall_known_extensions=true", "SET autoload_known_extensions=true"},
		Logger:         logger,
	})
	require.NoError(t, err)

	// acquire connection
	conn, release, err := db.AcquireReadConnection(ctx)
	require.NoError(t, err)

	// drop table
	err = db.DropTable(ctx, "test")
	require.NoError(t, err)

	// verify table is still accessible
	verifyTableForConn(t, conn, "SELECT id, country FROM test", []testData{{ID: 1, Country: "India"}})
	require.NoError(t, release())

	// verify table is now dropped
	err = db.DropTable(ctx, "test")
	require.ErrorContains(t, err, "not found")

	require.NoError(t, db.Close())
}

func TestResetSelectiveLocal(t *testing.T) {
	db, localDir, remoteDir := prepareDB(t)
	ctx := context.Background()

	// create table
	_, err := db.CreateTableAsSelect(ctx, "test", "SELECT 1 AS id, 'India' AS country", &CreateTableOptions{})
	require.NoError(t, err)
	verifyTable(t, db, "SELECT id, country FROM test", []testData{{ID: 1, Country: "India"}})

	// create two views on this
	_, err = db.CreateTableAsSelect(ctx, "test_view", "SELECT * FROM test", &CreateTableOptions{View: true})
	require.NoError(t, err)
	_, err = db.CreateTableAsSelect(ctx, "test_view2", "SELECT * FROM test", &CreateTableOptions{View: true})
	require.NoError(t, err)

	// create another table
	_, err = db.CreateTableAsSelect(ctx, "test2", "SELECT 2 AS id, 'USA' AS country", &CreateTableOptions{})
	require.NoError(t, err)

	// create views on this
	_, err = db.CreateTableAsSelect(ctx, "test2_view", "SELECT * FROM test2", &CreateTableOptions{View: true})
	require.NoError(t, err)

	// reset local for some tables
	require.NoError(t, db.Close())
	require.NoError(t, os.RemoveAll(filepath.Join(localDir, "test2")))
	require.NoError(t, os.RemoveAll(filepath.Join(localDir, "test_view2")))

	logger := zap.NewNop()
	bucket, err := fileblob.OpenBucket(remoteDir, nil)
	require.NoError(t, err)
	db, err = NewDB(ctx, &DBOptions{
		LocalPath:      localDir,
		Remote:         bucket,
		MemoryLimitGB:  2,
		CPU:            1,
		ReadWriteRatio: 0.5,
		DBInitQueries:  []string{"SET autoinstall_known_extensions=true", "SET autoload_known_extensions=true"},
		Logger:         logger,
	})
	require.NoError(t, err)
	verifyTable(t, db, "SELECT id, country FROM test2_view", []testData{{ID: 2, Country: "USA"}})
	verifyTable(t, db, "SELECT id, country FROM test_view2", []testData{{ID: 1, Country: "India"}})
	require.NoError(t, db.Close())
}

func TestResetTablesRemote(t *testing.T) {
	db, localDir, remoteDir := prepareDB(t)
	ctx := context.Background()

	// create table
	_, err := db.CreateTableAsSelect(ctx, "test", "SELECT 1 AS id, 'India' AS country", &CreateTableOptions{})
	require.NoError(t, err)

	require.NoError(t, db.Close())

	// remove remote data
	require.NoError(t, os.RemoveAll(remoteDir))

	logger := zap.NewNop()
	bucket, err := fileblob.OpenBucket(remoteDir, &fileblob.Options{CreateDir: true})
	require.NoError(t, err)
	db, err = NewDB(ctx, &DBOptions{
		LocalPath:      localDir,
		Remote:         bucket,
		MemoryLimitGB:  2,
		CPU:            1,
		ReadWriteRatio: 0.5,
		DBInitQueries:  []string{"SET autoinstall_known_extensions=true", "SET autoload_known_extensions=true"},
		Logger:         logger,
	})
	require.NoError(t, err)
	require.ErrorContains(t, db.DropTable(ctx, "test"), "not found")
	require.NoError(t, db.Close())
}

func TestResetSelectiveTablesRemote(t *testing.T) {
	db, localDir, remoteDir := prepareDB(t)
	ctx := context.Background()

	// create table
	_, err := db.CreateTableAsSelect(ctx, "test", "SELECT 1 AS id, 'India' AS country", &CreateTableOptions{})
	require.NoError(t, err)

	// create two views on this
	_, err = db.CreateTableAsSelect(ctx, "test_view", "SELECT * FROM test", &CreateTableOptions{View: true})
	require.NoError(t, err)
	_, err = db.CreateTableAsSelect(ctx, "test_view2", "SELECT * FROM test", &CreateTableOptions{View: true})
	require.NoError(t, err)

	// create another table
	_, err = db.CreateTableAsSelect(ctx, "test2", "SELECT 2 AS id, 'USA' AS country", &CreateTableOptions{})
	require.NoError(t, err)

	// create views on this
	_, err = db.CreateTableAsSelect(ctx, "test2_view", "SELECT * FROM test2", &CreateTableOptions{View: true})
	require.NoError(t, err)

	require.NoError(t, db.Close())

	// remove remote data for some tables
	require.NoError(t, os.RemoveAll(filepath.Join(remoteDir, "test2")))
	require.NoError(t, os.RemoveAll(filepath.Join(remoteDir, "test_view2")))

	logger := zap.NewNop()
	bucket, err := fileblob.OpenBucket(remoteDir, nil)
	require.NoError(t, err)
	db, err = NewDB(ctx, &DBOptions{
		LocalPath:      localDir,
		Remote:         bucket,
		MemoryLimitGB:  2,
		CPU:            1,
		ReadWriteRatio: 0.5,
		DBInitQueries:  []string{"SET autoinstall_known_extensions=true", "SET autoload_known_extensions=true"},
		Logger:         logger,
	})
	require.NoError(t, err)
	verifyTable(t, db, "SELECT id, country FROM test", []testData{{ID: 1, Country: "India"}})
	verifyTable(t, db, "SELECT id, country FROM test_view", []testData{{ID: 1, Country: "India"}})
	require.NoError(t, db.Close())
}

func TestConcurrentReads(t *testing.T) {
	testDB, _, _ := prepareDB(t)
	ctx := context.Background()

	// create table
	_, err := testDB.CreateTableAsSelect(ctx, "pest", "SELECT 2 AS id, 'USA' AS country", &CreateTableOptions{})
	require.NoError(t, err)

	// create test table
	_, err = testDB.CreateTableAsSelect(ctx, "test", "SELECT 1 AS id, 'India' AS country", &CreateTableOptions{})
	require.NoError(t, err)

	// acquire connection
	conn1, release1, err1 := testDB.AcquireReadConnection(ctx)
	require.NoError(t, err1)

	// replace with a view
	_, err = testDB.CreateTableAsSelect(ctx, "test", "SELECT * FROM pest", &CreateTableOptions{View: true})
	require.NoError(t, err)

	// acquire connection
	conn2, release2, err2 := testDB.AcquireReadConnection(ctx)
	require.NoError(t, err2)

	// drop table
	err = testDB.DropTable(ctx, "test")

	// verify both tables are still accessible
	verifyTableForConn(t, conn1, "SELECT id, country FROM test", []testData{{ID: 1, Country: "India"}})
	require.NoError(t, release1())
	verifyTableForConn(t, conn2, "SELECT id, country FROM test", []testData{{ID: 2, Country: "USA"}})
	require.NoError(t, release2())

	// acquire connection to see that table is now dropped
	conn3, release3, err3 := testDB.AcquireReadConnection(ctx)
	require.NoError(t, err3)
	var id int
	var country string
	err = conn3.QueryRowContext(ctx, "SELECT id, country FROM test").Scan(&id, &country)
	require.Error(t, err)
	require.NoError(t, release3())
}

func TestInconsistentSchema(t *testing.T) {
	testDB, _, _ := prepareDB(t)
	ctx := context.Background()

	// create table
	_, err := testDB.CreateTableAsSelect(ctx, "test", "SELECT 2 AS id, 'USA' AS country", &CreateTableOptions{})
	require.NoError(t, err)

	// create view
	_, err = testDB.CreateTableAsSelect(ctx, "test_view", "SELECT id, country FROM test", &CreateTableOptions{View: true})
	require.NoError(t, err)
	verifyTable(t, testDB, "SELECT * FROM test_view", []testData{{ID: 2, Country: "USA"}})

	// replace underlying table
	_, err = testDB.CreateTableAsSelect(ctx, "test", "SELECT 20 AS id, 'USB' AS city", &CreateTableOptions{})
	require.NoError(t, err)

	conn, release, err := testDB.AcquireReadConnection(ctx)
	require.NoError(t, err)
	defer release()

	var (
		id      int
		country string
	)
	err = conn.QueryRowxContext(ctx, "SELECT * FROM test_view").Scan(&id, &country)
	require.Error(t, err)

	// but querying from table should work
	err = conn.QueryRowxContext(ctx, "SELECT * FROM test").Scan(&id, &country)
	require.NoError(t, err)
	require.Equal(t, 20, id)
	require.Equal(t, "USB", country)
}

func TestViews(t *testing.T) {
	testDB, _, _ := prepareDB(t)
	ctx := context.Background()

	// create view
	_, err := testDB.CreateTableAsSelect(ctx, "parent_view", "SELECT 1 AS id, 'India' AS country", &CreateTableOptions{View: true})
	require.NoError(t, err)

	// create dependent view
	_, err = testDB.CreateTableAsSelect(ctx, "child_view", "SELECT * FROM parent_view", &CreateTableOptions{View: true})
	require.NoError(t, err)
	verifyTable(t, testDB, "SELECT id, country FROM child_view", []testData{{ID: 1, Country: "India"}})

	// replace parent view
	_, err = testDB.CreateTableAsSelect(ctx, "parent_view", "SELECT 2 AS id, 'USA' AS country", &CreateTableOptions{View: true})
	require.NoError(t, err)
	verifyTable(t, testDB, "SELECT id, country FROM child_view", []testData{{ID: 2, Country: "USA"}})

	// rename child view
	err = testDB.RenameTable(ctx, "child_view", "view0")
	require.NoError(t, err)
	verifyTable(t, testDB, "SELECT id, country FROM view0", []testData{{ID: 2, Country: "USA"}})

	// old child view does not exist
	err = testDB.DropTable(ctx, "child_view")
	require.Error(t, err)

	// create a chain of views
	for i := 1; i <= 10; i++ {
		_, err = testDB.CreateTableAsSelect(ctx, fmt.Sprintf("view%d", i), fmt.Sprintf("SELECT * FROM view%d", i-1), &CreateTableOptions{View: true})
		require.NoError(t, err)
	}
	verifyTable(t, testDB, "SELECT id, country FROM view10", []testData{{ID: 2, Country: "USA"}})

	require.NoError(t, testDB.Close())
}

func TestCloseDB(t *testing.T) {
	localDir := t.TempDir()
	db, err := NewDB(t.Context(), &DBOptions{
		LocalPath:      localDir,
		ReadWriteRatio: 0.5,
		Remote:         nil,
		Logger:         zap.NewNop(),
	})
	require.NoError(t, err)

	// create view
	_, err = db.CreateTableAsSelect(t.Context(), "view1", "SELECT 1 AS id, 'India' AS country", &CreateTableOptions{View: true})
	require.NoError(t, err)

	// create another view
	_, err = db.CreateTableAsSelect(t.Context(), "view2", "SELECT 2 AS id, 'USA' AS country", &CreateTableOptions{View: true})
	require.NoError(t, err)

	// drop view1
	err = db.DropTable(t.Context(), "view1")
	require.NoError(t, err)

	// wait for async delete versions to complete
	time.Sleep(2 * time.Second)
	// close DB
	require.NoError(t, db.Close())

	// reopen DB
	db, err = NewDB(t.Context(), &DBOptions{
		LocalPath:      localDir,
		ReadWriteRatio: 0.5,
		Remote:         nil,
		Logger:         zap.NewNop(),
	})
	require.NoError(t, err)
	defer func() {
		require.NoError(t, db.Close())
	}()

	conn, release, err := db.AcquireReadConnection(t.Context())
	require.NoError(t, err)
	defer func() {
		require.NoError(t, release())
	}()

	var id int
	var country string
	err = conn.QueryRowxContext(t.Context(), "SELECT id, country FROM view1").Scan(id, country)
	require.Error(t, err, "view1 should not exist")

	// view2 should still exist
	err = conn.QueryRowxContext(t.Context(), "SELECT id, country FROM view2").Scan(&id, &country)
	require.NoError(t, err)
	require.Equal(t, 2, id)
	require.Equal(t, "USA", country)
}

func TestDiskCleanupOnFailures(t *testing.T) {
	testDB, localDir, _ := prepareDB(t)
	ctx := context.Background()

	// create a table that fails
	_, err := testDB.CreateTableAsSelect(ctx, "test_table_fail", "SELECT 1 AS id, 'India AS country", &CreateTableOptions{})
	require.Error(t, err)

	// verify directory is cleaned up
	entries, err := os.ReadDir(filepath.Join(localDir, "test_table_fail"))
	require.NoError(t, err)
	require.Len(t, entries, 0)

	// create a view that fails
	_, err = testDB.CreateTableAsSelect(ctx, "test_view_fail", "SELECT 1 AS id, 'India AS country", &CreateTableOptions{View: true})
	require.Error(t, err)

	// verify directory is cleaned up
	entries, err = os.ReadDir(filepath.Join(localDir, "test_view_fail"))
	require.NoError(t, err)
	require.Len(t, entries, 0)

	require.NoError(t, testDB.Close())
}

func prepareDB(t *testing.T) (db DB, localDir, remoteDir string) {
	localDir = t.TempDir()
	ctx := context.Background()
	remoteDir = t.TempDir()
	bucket, err := fileblob.OpenBucket(remoteDir, nil)
	require.NoError(t, err)
	db, err = NewDB(ctx, &DBOptions{
		LocalPath:      localDir,
		Remote:         bucket,
		MemoryLimitGB:  2,
		CPU:            1,
		ReadWriteRatio: 0.5,
		DBInitQueries:  []string{"SET autoinstall_known_extensions=true", "SET autoload_known_extensions=true"},
		Logger:         zap.NewNop(),
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

func verifyTableForConn(t *testing.T, conn *sqlx.Conn, query string, data []testData) {
	ctx := context.Background()
	var scannedData []testData
	err := conn.SelectContext(ctx, &scannedData, query)
	require.NoError(t, err)
	require.Equal(t, data, scannedData)
}

type testData struct {
	ID      int    `db:"id"`
	Country string `db:"country"`
	City    string `db:"city"`
}
