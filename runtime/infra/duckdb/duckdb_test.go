package duckdb

import (
	"context"
	"testing"
	"time"

	"github.com/rilldata/rill/runtime/infra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

func TestQuery(t *testing.T) {
	conn := prepareConn(t)

	rows, err := conn.Execute(context.Background(), &infra.Statement{Query: "SELECT COUNT(*) FROM foo"})
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

	tables, err := conn.InformationSchema().All(context.Background())
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
	ctx := context.Background()

	table, err := conn.InformationSchema().Lookup(ctx, "foo")
	require.NoError(t, err)
	require.Equal(t, "foo", table.Name)

	_, err = conn.InformationSchema().Lookup(ctx, "bad")
	require.Equal(t, infra.ErrNotFound, err)
}

func TestPriorityQueue(t *testing.T) {
	conn := prepareConn(t)
	defer conn.Close()

	n := 100
	results := make(chan int, n)
	var g errgroup.Group

	for i := n; i > 0; i-- {
		priority := i
		g.Go(func() error {
			rows, err := conn.Execute(context.Background(), &infra.Statement{
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

	err := g.Wait()
	require.NoError(t, err)

	for i := n; i > 0; i-- {
		x := <-results
		assert.Equal(t, i, x)
	}
}

func TestCancel(t *testing.T) {
	conn := prepareConn(t)
	defer conn.Close()

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

			rows, err := conn.Execute(ctx, &infra.Statement{
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
	conn := prepareConn(t)

	n := 100
	results := make(chan int, n)
	var g errgroup.Group

	for i := n; i > 0; i-- {
		priority := i
		g.Go(func() error {
			rows, err := conn.Execute(context.Background(), &infra.Statement{
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

	g.Go(func() error {
		err := conn.Close()
		require.NoError(t, err)
		return nil
	})

	err := g.Wait()
	require.Equal(t, infra.ErrClosed, err)

	x := <-results
	require.Greater(t, x, 0)
}

func prepareConn(t *testing.T) infra.Connection {
	conn, err := driver{}.Open("?access_mode=read_write")
	require.NoError(t, err)

	rows, err := conn.Execute(context.Background(), &infra.Statement{
		Query: "CREATE TABLE foo(bar VARCHAR, baz INTEGER)",
	})
	require.NoError(t, err)
	require.NoError(t, rows.Close())

	rows, err = conn.Execute(context.Background(), &infra.Statement{
		Query: "INSERT INTO foo VALUES ('a', 1), ('a', 2), ('b', 3), ('c', 4)",
	})
	require.NoError(t, err)
	require.NoError(t, rows.Close())

	rows, err = conn.Execute(context.Background(), &infra.Statement{
		Query: "CREATE TABLE bar(bar VARCHAR, baz INTEGER)",
	})
	require.NoError(t, err)
	require.NoError(t, rows.Close())

	rows, err = conn.Execute(context.Background(), &infra.Statement{
		Query: "INSERT INTO bar VALUES ('a', 1), ('a', 2), ('b', 3), ('c', 4)",
	})

	require.NoError(t, err)
	require.NoError(t, rows.Close())

	return conn
}
