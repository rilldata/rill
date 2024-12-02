package duckdb

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/rduckdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gocloud.dev/blob/memblob"
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

func Test_safeSQLString(t *testing.T) {
	conn := prepareConn(t)
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "let's t@st \"weird\" dirs")
	err := os.Mkdir(path, fs.ModePerm)
	require.NoError(t, err)

	// dbFile := filepath.Join(path, "st@g3's.db")
	err = conn.db.CreateTableAsSelect(context.Background(), "foo", "SELECT 'a' AS bar, 1 AS baz", &rduckdb.CreateTableOptions{
		// InitSQL: fmt.Sprintf("ATTACH %s", safeSQLString(dbFile)),
	})
	require.NoError(t, err)
}

func prepareConn(t *testing.T) *connection {
	conn, err := Driver{}.Open("default", map[string]any{"data_dir": t.TempDir(), "pool_size": 4}, activity.NewNoopClient(), memblob.OpenBucket(nil), zap.NewNop())
	require.NoError(t, err)

	olap, ok := conn.AsOLAP("")
	require.True(t, ok)

	ctx := context.Background()
	err = olap.CreateTableAsSelect(ctx, "foo", false, "SELECT 'a' AS bar, 1 AS baz UNION ALL SELECT 'a', 2 UNION ALL SELECT 'b', 3 UNION ALL SELECT 'c', 4", nil)
	require.NoError(t, err)

	err = olap.CreateTableAsSelect(ctx, "bar", false, "SELECT 'a' AS bar, 1 AS baz UNION ALL SELECT 'a', 2 UNION ALL SELECT 'b', 3 UNION ALL SELECT 'c', 4", nil)
	require.NoError(t, err)
	return conn.(*connection)
}
