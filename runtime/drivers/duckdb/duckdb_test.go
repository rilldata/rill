package duckdb

import (
	"context"
	"database/sql"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestOpenDrop(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tmp.db")
	walpath := path + ".wal"
	dsn := path

	handle, err := Driver{}.Open(map[string]any{"dsn": dsn, "pool_size": 2}, false, activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)

	olap, ok := handle.AsOLAP("")
	require.True(t, ok)

	err = olap.Exec(context.Background(), &drivers.Statement{Query: "CREATE TABLE foo (bar INTEGER)"})
	require.NoError(t, err)

	err = handle.Close()
	require.NoError(t, err)
	require.FileExists(t, path)

	err = Driver{}.Drop(map[string]any{"dsn": dsn}, zap.NewNop())
	require.NoError(t, err)
	require.NoFileExists(t, path)
	require.NoFileExists(t, walpath)
}

func TestNoFatalErr(t *testing.T) {
	// NOTE: Using this issue to create a fatal error: https://github.com/duckdb/duckdb/issues/7905

	dsn := filepath.Join(t.TempDir(), "tmp.db")

	handle, err := Driver{}.Open(map[string]any{"dsn": dsn, "pool_size": 2}, false, activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)

	olap, ok := handle.AsOLAP("")
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
	require.NoError(t, err)

	err = olap.Exec(context.Background(), &drivers.Statement{Query: "SELECT * FROM a"})
	require.NoError(t, err)

	err = handle.Close()
	require.NoError(t, err)
}

func TestNoFatalErrConcurrent(t *testing.T) {
	// NOTE: Using this issue to create a fatal error: https://github.com/duckdb/duckdb/issues/7905

	dsn := filepath.Join(t.TempDir(), "tmp.db")

	handle, err := Driver{}.Open(map[string]any{"dsn": dsn, "pool_size": 3}, false, activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)

	olap, ok := handle.AsOLAP("")
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
		err1 = olap.WithConnection(context.Background(), 0, false, false, func(ctx, ensuredCtx context.Context, _ *sql.Conn) error {
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
		err2 = olap.WithConnection(context.Background(), 0, false, false, func(ctx, ensuredCtx context.Context, _ *sql.Conn) error {
			time.Sleep(1000 * time.Millisecond)
			return olap.Exec(ctx, &drivers.Statement{Query: qry})
		})
		wg.Done()
	}()

	// Func 3 acquires conn after 250ms and runs query immediately. It will be enqueued (because the OLAP conns limit is pool_size-1 = 2).
	// By the time it's dequeued, the DB will have been invalidated, and it will wait for the reopen before returning a conn. So the query should succeed.
	wg.Add(1)
	var err3 error
	go func() {
		time.Sleep(250 * time.Millisecond)
		qry := `SELECT * FROM a;`
		err3 = olap.WithConnection(context.Background(), 0, false, false, func(ctx, ensuredCtx context.Context, _ *sql.Conn) error {
			return olap.Exec(ctx, &drivers.Statement{Query: qry})
		})
		wg.Done()
	}()

	wg.Wait()

	require.NoError(t, err1)
	require.NoError(t, err2)
	require.NoError(t, err3)

	err = olap.Exec(context.Background(), &drivers.Statement{Query: "SELECT * FROM a"})
	require.NoError(t, err)

	err = handle.Close()
	require.NoError(t, err)
}

func TestHumanReadableSizeToBytes(t *testing.T) {
	tests := []struct {
		input     string
		expected  float64
		shouldErr bool
	}{
		{"1 byte", 1, false},
		{"2 bytes", 2, false},
		{"1KB", 1000, false},
		{"1.5KB", 1500, false},
		{"1MB", 1000 * 1000, false},
		{"2.5MB", 2.5 * 1000 * 1000, false},
		{"1GB", 1000 * 1000 * 1000, false},
		{"1.5GB", 1.5 * 1000 * 1000 * 1000, false},
		{"1TB", 1000 * 1000 * 1000 * 1000, false},
		{"1.5TB", 1.5 * 1000 * 1000 * 1000 * 1000, false},
		{"1PB", 1000 * 1000 * 1000 * 1000 * 1000, false},
		{"1.5PB", 1.5 * 1000 * 1000 * 1000 * 1000 * 1000, false},
		{"invalid", 0, true},
		{"123invalid", 0, true},
		{"123 ZZ", 0, true},
	}

	for _, tt := range tests {
		result, err := humanReadableSizeToBytes(tt.input)
		if (err != nil) != tt.shouldErr {
			t.Errorf("expected error: %v, got error: %v for input: %s", tt.shouldErr, err, tt.input)
		}

		if !tt.shouldErr && result != tt.expected {
			t.Errorf("expected: %v, got: %v for input: %s", tt.expected, result, tt.input)
		}
	}
}
