package duckdb

import (
	"context"
	"path/filepath"
	"sync"
	"testing"
	"time"

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

func TestFatalErr(t *testing.T) {
	// NOTE: Using this issue to create a fatal error: https://github.com/duckdb/duckdb/issues/7905

	path := filepath.Join(t.TempDir(), "tmp.db")
	dsn := path + "?rill_pool_size=2"

	handle, err := Driver{}.Open(dsn, zap.NewNop())
	require.NoError(t, err)

	olap, ok := handle.OLAPStore()
	require.True(t, ok)

	qry := `
		CREATE TABLE a(
			a1 VARCHAR,
		);

		CREATE TABLE b(
			b1 VARCHAR,
			b2 TIMESTAMP,
			b3 TIMESTAMP,
			b4 VARCHAR,
			b5 VARCHAR,
			b6 VARCHAR,
			b7 TIMESTAMP,
			b8 TIMESTAMP,
			b9 VARCHAR,
			b10 VARCHAR,
			b11 VARCHAR,
			b12 VARCHAR,
			b13 VARCHAR,
			b14 VARCHAR,
		);

		INSERT INTO b VALUES (NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL);

		CREATE TABLE c(
			c1 VARCHAR,
		);

		CREATE TABLE d(
			d1 VARCHAR,
			d2 VARCHAR,
		);

		SELECT *
		FROM a
		LEFT JOIN b ON b.b14 = a.a1 
		LEFT JOIN c ON b.b13 = c.c1
		LEFT JOIN d ON b.b12 = d.d1
		WHERE d.d2 IN ('');
	`

	err = olap.Exec(context.Background(), &drivers.Statement{Query: qry})
	require.ErrorContains(t, err, "INTERNAL Error")

	err = olap.Exec(context.Background(), &drivers.Statement{Query: "SELECT * FROM a"})
	require.NoError(t, err)

	err = handle.Close()
	require.NoError(t, err)
}

func TestFatalErrConcurrent(t *testing.T) {
	// NOTE: Using this issue to create a fatal error: https://github.com/duckdb/duckdb/issues/7905

	path := filepath.Join(t.TempDir(), "tmp.db")
	dsn := path + "?rill_pool_size=3"

	handle, err := Driver{}.Open(dsn, zap.NewNop())
	require.NoError(t, err)

	olap, ok := handle.OLAPStore()
	require.True(t, ok)

	qry := `
		CREATE TABLE a(
			a1 VARCHAR,
		);

		CREATE TABLE b(
			b1 VARCHAR,
			b2 TIMESTAMP,
			b3 TIMESTAMP,
			b4 VARCHAR,
			b5 VARCHAR,
			b6 VARCHAR,
			b7 TIMESTAMP,
			b8 TIMESTAMP,
			b9 VARCHAR,
			b10 VARCHAR,
			b11 VARCHAR,
			b12 VARCHAR,
			b13 VARCHAR,
			b14 VARCHAR,
		);

		INSERT INTO b VALUES (NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL);

		CREATE TABLE c(
			c1 VARCHAR,
		);

		CREATE TABLE d(
			d1 VARCHAR,
			d2 VARCHAR,
		);
	`
	err = olap.Exec(context.Background(), &drivers.Statement{Query: qry})
	require.NoError(t, err)

	wg := sync.WaitGroup{}

	// Func 1 acquires conn immediately, runs query after 500ms.
	// It should fail with an internal error.
	wg.Add(1)
	var err1 error
	go func() {
		qry := `
			SELECT *
			FROM a
			LEFT JOIN b ON b.b14 = a.a1 
			LEFT JOIN c ON b.b13 = c.c1
			LEFT JOIN d ON b.b12 = d.d1
			WHERE d.d2 IN ('');
		`
		err1 = olap.WithConnection(context.Background(), 0, func(ctx, ensuredCtx context.Context) error {
			time.Sleep(500 * time.Millisecond)
			return olap.Exec(ctx, &drivers.Statement{Query: qry})
		})
		wg.Done()
	}()

	// Func 2 acquires conn immediately, runs query after 1000ms
	// It should fail with a fatal error, because the DB has been invalidated by the previous query.
	wg.Add(1)
	var err2 error
	go func() {
		qry := `SELECT * FROM a;`
		err2 = olap.WithConnection(context.Background(), 0, func(ctx, ensuredCtx context.Context) error {
			time.Sleep(1000 * time.Millisecond)
			return olap.Exec(ctx, &drivers.Statement{Query: qry})
		})
		wg.Done()
	}()

	// Func 3 acquires conn after 250ms and runs query immediately. It will be enqueued (because the OLAP conns limit is rill_pool_size-1 = 2).
	// By the time it's dequeued, the DB will have been invalidated, and it will wait for the reopen before returning a conn. So the query should succeed.
	wg.Add(1)
	var err3 error
	go func() {
		time.Sleep(250 * time.Millisecond)
		qry := `SELECT * FROM a;`
		err3 = olap.WithConnection(context.Background(), 0, func(ctx, ensuredCtx context.Context) error {
			return olap.Exec(ctx, &drivers.Statement{Query: qry})
		})
		wg.Done()
	}()

	wg.Wait()

	require.ErrorContains(t, err1, "INTERNAL Error")
	require.ErrorContains(t, err2, "FATAL Error")
	require.NoError(t, err3)

	err = olap.Exec(context.Background(), &drivers.Statement{Query: "SELECT * FROM a"})
	require.NoError(t, err)

	err = handle.Close()
	require.NoError(t, err)
}
