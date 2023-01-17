package duckdb

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
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
}

func TestPriorityQueue(t *testing.T) {
	conn := prepareConn(t)
	olap, _ := conn.OLAPStore()
	defer conn.Close()

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

	// give the queue plenty of time to fill up
	time.Sleep(1000 * time.Millisecond)

	err := g.Wait()
	require.NoError(t, err)

	actual := 0
	expected := 0
	for i := n; i > 0; i-- {
		actual += <-results
		expected += i
	}
	assert.Equal(t, expected, actual)
}

func TestCancel(t *testing.T) {
	conn := prepareConn(t)
	olap, _ := conn.OLAPStore()
	defer conn.Close()

	n := 100
	cancelIdx := 50
	cancelCh := make(chan struct{})

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
					// ensure the queue fills up before cancelling the context
					time.Sleep(100 * time.Millisecond)
					cancel()
					cancelCh <- struct{}{}
				}()
			}

			if priority == cancelIdx {
				// wait until context is cancelled
				<-cancelCh
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

	// give the queue plenty of time to fill up
	time.Sleep(1000 * time.Millisecond)

	err := g.Wait()
	require.NoError(t, err)

	actual := 0
	expected := 0
	for i := n; i > 0; i-- {
		if i == cancelIdx {
			continue
		}
		actual += <-results
		expected += i
	}
	assert.Equal(t, expected, actual)
}

func TestClose(t *testing.T) {
	conn := prepareConn(t)
	olap, _ := conn.OLAPStore()

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

	g.Go(func() error {
		err := conn.Close()
		require.NoError(t, err)
		return nil
	})

	err := g.Wait()
	require.Equal(t, errors.New("sql: database is closed"), err)

	x := <-results
	require.Greater(t, x, 0)
}

func prepareConn(t *testing.T) drivers.Connection {
	conn, err := Driver{}.Open("?access_mode=read_write&rill_pool_size=4", zap.NewNop())
	require.NoError(t, err)

	olap, ok := conn.OLAPStore()
	require.True(t, ok)

	err = olap.Exec(context.Background(), &drivers.Statement{
		Query: "CREATE TABLE foo(bar VARCHAR, baz INTEGER)",
	})
	require.NoError(t, err)

	err = olap.Exec(context.Background(), &drivers.Statement{
		Query: "INSERT INTO foo VALUES ('a', 1), ('a', 2), ('b', 3), ('c', 4)",
	})
	require.NoError(t, err)

	err = olap.Exec(context.Background(), &drivers.Statement{
		Query: "CREATE TABLE bar(bar VARCHAR, baz INTEGER)",
	})
	require.NoError(t, err)

	err = olap.Exec(context.Background(), &drivers.Statement{
		Query: "INSERT INTO bar VALUES ('a', 1), ('a', 2), ('b', 3), ('c', 4)",
	})
	require.NoError(t, err)

	return conn
}
