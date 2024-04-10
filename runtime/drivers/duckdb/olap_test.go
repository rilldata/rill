package duckdb

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func TestQuery(t *testing.T) {
	conn := prepareConn(t)
	olap, _ := conn.AsOLAP("")

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
	if testing.Short() {
		t.Skip()
	}

	conn := prepareConn(t)
	olap, _ := conn.AsOLAP("")
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
	if testing.Short() {
		t.Skip()
	}

	conn := prepareConn(t)
	olap, _ := conn.AsOLAP("")
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
	if testing.Short() {
		t.Skip()
	}

	conn := prepareConn(t)
	olap, _ := conn.AsOLAP("instanceID string")

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

	require.Error(t, err)
	isConnErr := strings.Contains(err.Error(), "database/sql/driver: could not connect to database")
	isClosedErr := strings.Contains(err.Error(), "sql: database is closed")
	require.True(t, isConnErr || isClosedErr, "Error should be either connection error or database closed error")

	x := <-results
	require.Greater(t, x, 0)
}

func prepareConn(t *testing.T) drivers.Handle {
	conn, err := Driver{}.Open(map[string]any{"dsn": ":memory:?access_mode=read_write", "pool_size": 4}, false, activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)

	olap, ok := conn.AsOLAP("")
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

func Test_safeSQLString(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "let's t@st \"weird\" dirs")
	err := os.Mkdir(path, fs.ModePerm)
	require.NoError(t, err)

	dbFile := filepath.Join(path, "st@g3's.db")
	conn, err := Driver{}.Open(map[string]any{"path": dbFile}, false, activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	require.NoError(t, conn.Close())

	conn, err = Driver{}.Open(map[string]any{}, false, activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)

	olap, ok := conn.AsOLAP("")
	require.True(t, ok)

	err = olap.Exec(context.Background(), &drivers.Statement{Query: fmt.Sprintf("ATTACH '%s'", dbFile)})
	require.Error(t, err)

	err = olap.Exec(context.Background(), &drivers.Statement{Query: fmt.Sprintf("ATTACH %s", safeSQLString(dbFile))})
	require.NoError(t, err)
}
