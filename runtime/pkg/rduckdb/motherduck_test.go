package rduckdb

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/rilldata/rill/runtime/testruntime/testmode"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestMain(m *testing.M) {
	// Load .env file at the repo root (if any)
	_, currentFile, _, _ := runtime.Caller(0)
	envPath := filepath.Join(currentFile, "..", "..", "..", "..", ".env")
	_, err := os.Stat(envPath)
	if err == nil {
		err = godotenv.Load(envPath)
		if err != nil {
			panic(fmt.Sprintf("Error loading .env file: %v", err))
		}
	}
	_ = m.Run()
}

func TestMotherDuckDB(t *testing.T) {
	t.Parallel()
	testmode.Expensive(t)
	db := prepareMotherDuckDB(t)
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

func TestMotherDuckCreateTable(t *testing.T) {
	t.Parallel()
	testmode.Expensive(t)
	db := prepareMotherDuckDB(t)
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

func TestMotherDuckDropTable(t *testing.T) {
	t.Parallel()
	testmode.Expensive(t)
	db := prepareMotherDuckDB(t)
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

func TestMotherDuckMutateTable(t *testing.T) {
	t.Parallel()
	testmode.Expensive(t)
	db := prepareMotherDuckDB(t)
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
	db.MutateTable(ctx, "test", nil, func(ctx context.Context, conn *sqlx.Conn) error {
		_, err := conn.ExecContext(ctx, "ALTER TABLE test ADD COLUMN country TEXT")
		require.NoError(t, err)
		_, err = conn.ExecContext(ctx, "UPDATE test SET country = 'USA' WHERE id = 2")
		require.NoError(t, err)
		_, err = conn.ExecContext(ctx, "UPDATE test SET country = 'India' WHERE id = 1")
		require.NoError(t, err)
		return nil
	})

	verifyTable(t, db, "SELECT * FROM test", []testData{{ID: 1, City: "Delhi", Country: "India"}, {ID: 2, City: "NY", Country: "USA"}})
	require.NoError(t, db.Close())
}

func TestOtherSchema(t *testing.T) {
	t.Parallel()
	testmode.Expensive(t)
	tempDir := t.TempDir()
	randomDB := provisionDatabase(t)
	db, err := NewGeneric(context.Background(), &GenericOptions{
		DBInitQueries:      []string{"INSTALL motherduck; LOAD motherduck; SET motherduck_token = '" + os.Getenv("RILL_RUNTIME_MOTHERDUCK_TEST_TOKEN") + "'"},
		Path:               fmt.Sprintf("md:%s", randomDB),
		LocalDataDir:       tempDir,
		LocalMemoryLimitGB: 2,
		LocalCPU:           1,
		Logger:             zap.NewNop(),
		SchemaName:         "other",
	})
	require.NoError(t, err)
	ctx := context.Background()

	// create table in other schema
	_, err = db.CreateTableAsSelect(ctx, "test_other", "SELECT * FROM test", &CreateTableOptions{})
	require.NoError(t, err)

	// query table
	var greeting string
	conn, release, err := db.AcquireReadConnection(ctx)
	require.NoError(t, err)

	conn.QueryRowContext(ctx, "SELECT * FROM test_other").Scan(&greeting)
	require.Equal(t, "hello", greeting)
	require.NoError(t, release())

	require.NoError(t, db.Close())
}

func prepareMotherDuckDB(t *testing.T) DB {
	tempDir := t.TempDir()
	randomDB := provisionDatabase(t)
	db, err := NewGeneric(context.Background(), &GenericOptions{
		DBInitQueries:      []string{"INSTALL motherduck; LOAD motherduck; SET motherduck_token = '" + os.Getenv("RILL_RUNTIME_MOTHERDUCK_TEST_TOKEN") + "'"},
		Path:               fmt.Sprintf("md:%s", randomDB),
		LocalDataDir:       tempDir,
		LocalMemoryLimitGB: 2,
		LocalCPU:           1,
		Logger:             zap.NewNop(),
	})
	require.NoError(t, err)
	// Just to test that connections are closed properly
	db.(*generic).db.SetMaxOpenConns(1)
	return db
}

func provisionDatabase(t *testing.T) string {
	db, err := sql.Open("duckdb", "md:my_db?motherduck_token="+os.Getenv("RILL_RUNTIME_MOTHERDUCK_TEST_TOKEN"))
	require.NoError(t, err)

	name := "db" + uuid.NewString()[:8]
	_, err = db.Exec("CREATE DATABASE " + name)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, err = db.Exec("DROP DATABASE " + name)
		require.NoError(t, err)
		require.NoError(t, db.Close())
	})

	_, err = db.Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s.other", name))
	require.NoError(t, err)

	_, err = db.Exec(fmt.Sprintf("CREATE OR REPLACE TABLE %s.other.test AS SELECT 'hello' AS greeting", name))
	require.NoError(t, err)
	return name
}
