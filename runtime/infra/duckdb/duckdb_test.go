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

	return conn
}
