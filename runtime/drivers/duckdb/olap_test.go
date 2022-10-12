package duckdb

import (
	"context"
	"testing"
	"time"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

func TestQuery(t *testing.T) {
	conn := prepareConn(t)
	olap, _ := conn.OLAPStore()

	rows, err := olap.Execute(context.Background(), &drivers.Statement{Query: "SELECT COUNT(*) FROM foo"})
	require.NoError(t, err)

	var count int
	rows.Next()
	require.NoError(t, rows.Scan(&count))
	require.Equal(t, 4, count)
	require.NoError(t, rows.Close())

	err = conn.Close()
	require.NoError(t, err)
	err = conn.(*connection).db.Ping()
	require.Error(t, err)
}

func TestInformationSchemaAll(t *testing.T) {
	conn := prepareConn(t)
	olap, _ := conn.OLAPStore()

	tables, err := olap.InformationSchema().All(context.Background())
	require.NoError(t, err)
	require.Equal(t, 2, len(tables))

	require.Equal(t, "bar", tables[0].Name)
	require.Equal(t, "foo", tables[1].Name)
	require.Equal(t, 2, len(tables[1].Columns))
	require.Equal(t, "bar", tables[1].Columns[0].Name)
	require.Equal(t, "VARCHAR", tables[1].Columns[0].Type)
	require.Equal(t, "baz", tables[1].Columns[1].Name)
	require.Equal(t, "INTEGER", tables[1].Columns[1].Type)
}

func TestInformationSchemaLookup(t *testing.T) {
	conn := prepareConn(t)
	olap, _ := conn.OLAPStore()
	ctx := context.Background()

	table, err := olap.InformationSchema().Lookup(ctx, "foo")
	require.NoError(t, err)
	require.Equal(t, "foo", table.Name)

	_, err = olap.InformationSchema().Lookup(ctx, "bad")
	require.Equal(t, drivers.ErrNotFound, err)
}

func TestPriorityQueue(t *testing.T) {
	if testing.Short() {
		t.Skip("duckdb: skipping test in short mode")
	}

	conn := prepareConn(t)
	olap, _ := conn.OLAPStore()
	defer conn.Close()

	// pause the priority worker to allow the queue to fill up
	conn.(*connection).worker.Pause()

	n := 100
	results := make(chan int, n)
	var g errgroup.Group

	for i := n; i > 0; i-- {
		priority := i
		g.Go(func() error {
			rows, err := olap.Execute(context.Background(), &drivers.Statement{
				Query:    "SELECT ?",
				Args:     []any{priority},
				Priority: priority,
			})
			if err != nil {
				return err
			}

			var x int
			rows.Next()
			rows.Scan(&x)
			results <- x

			return rows.Close()
		})
	}

	// give the queue plenty of time to fill up, then unpause
	time.Sleep(1000 * time.Millisecond)
	conn.(*connection).worker.Unpause()

	err := g.Wait()
	require.NoError(t, err)

	for i := n; i > 0; i-- {
		x := <-results
		assert.Equal(t, i, x)
	}
}

func TestCancel(t *testing.T) {
	if testing.Short() {
		t.Skip("duckdb: skipping test in short mode")
	}

	conn := prepareConn(t)
	olap, _ := conn.OLAPStore()
	defer conn.Close()

	// pause the priority worker to allow the queue to fill up
	conn.(*connection).worker.Pause()

	n := 100
	cancelIdx := 50
	results := make(chan int, n)
	var g errgroup.Group

	for i := n; i > 0; i-- {
		priority := i
		g.Go(func() error {
			ctx := context.Background()

			if priority == cancelIdx {
				cctx, cancel := context.WithCancel(ctx)
				ctx = cctx
				go func() {
					time.Sleep(time.Millisecond)
					cancel()
				}()
			}

			rows, err := olap.Execute(ctx, &drivers.Statement{
				Query:    "SELECT ?",
				Args:     []any{priority},
				Priority: priority,
			})

			if priority == cancelIdx {
				require.Error(t, err)
				return nil
			} else if err != nil {
				return err
			}

			var x int
			rows.Next()
			rows.Scan(&x)
			results <- x

			return rows.Close()
		})
	}

	// give the queue plenty of time to fill up, then unpause
	time.Sleep(1000 * time.Millisecond)
	conn.(*connection).worker.Unpause()

	err := g.Wait()
	require.NoError(t, err)

	for i := n; i > 0; i-- {
		if i == cancelIdx {
			continue
		}
		x := <-results
		assert.Equal(t, i, x)
	}
}

func TestClose(t *testing.T) {
	if testing.Short() {
		t.Skip("duckdb: skipping test in short mode")
	}

	conn := prepareConn(t)
	olap, _ := conn.OLAPStore()

	// pause the priority worker to allow the queue to fill up
	conn.(*connection).worker.Pause()

	n := 100
	results := make(chan int, n)
	var g errgroup.Group

	for i := n; i > 0; i-- {
		priority := i
		g.Go(func() error {
			rows, err := olap.Execute(context.Background(), &drivers.Statement{
				Query:    "SELECT ?",
				Args:     []any{priority},
				Priority: priority,
			})
			if err != nil {
				return err
			}

			var x int
			rows.Next()
			rows.Scan(&x)
			results <- x

			return rows.Close()
		})
	}

	// unpause the queue, so it con process a bit before closing
	conn.(*connection).worker.Unpause()

	g.Go(func() error {
		err := conn.Close()
		require.NoError(t, err)
		return nil
	})

	err := g.Wait()
	require.Equal(t, drivers.ErrClosed, err)

	x := <-results
	require.Greater(t, x, 0)
}

func prepareConn(t *testing.T) drivers.Connection {
	conn, err := driver{}.Open("?access_mode=read_write")
	require.NoError(t, err)

	olap, ok := conn.OLAPStore()
	require.True(t, ok)

	rows, err := olap.Execute(context.Background(), &drivers.Statement{
		Query: "CREATE TABLE foo(bar VARCHAR, baz INTEGER)",
	})
	require.NoError(t, err)
	require.NoError(t, rows.Close())

	rows, err = olap.Execute(context.Background(), &drivers.Statement{
		Query: "INSERT INTO foo VALUES ('a', 1), ('a', 2), ('b', 3), ('c', 4)",
	})
	require.NoError(t, err)
	require.NoError(t, rows.Close())

	rows, err = olap.Execute(context.Background(), &drivers.Statement{
		Query: "CREATE TABLE bar(bar VARCHAR, baz INTEGER)",
	})
	require.NoError(t, err)
	require.NoError(t, rows.Close())

	rows, err = olap.Execute(context.Background(), &drivers.Statement{
		Query: "INSERT INTO bar VALUES ('a', 1), ('a', 2), ('b', 3), ('c', 4)",
	})

	require.NoError(t, err)
	require.NoError(t, rows.Close())

	return conn
}
