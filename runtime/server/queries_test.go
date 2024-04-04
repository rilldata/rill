package server

import (
	"context"
	"testing"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestServer_InsertLimit_SELECT(t *testing.T) {
	t.Parallel()
	olap := prepareOLAPStore(t)
	err := olap.Exec(context.Background(), &drivers.Statement{
		Query: "CREATE TABLE tbl1 (col1 int)",
	})
	require.NoError(t, err)

	transformedSQL, err := ensureLimits(context.Background(), olap, "SELECT col1 FROM (SELECT col1 FROM tbl1) AS sub1 INNER JOIN (SELECT col1 FROM tbl1) AS sub2 ON (sub1.col1 = sub2.col1)", 100)
	require.NoError(t, err)
	require.Equal(t, "SELECT col1 FROM (SELECT col1 FROM tbl1) AS sub1 INNER JOIN (SELECT col1 FROM tbl1) AS sub2 ON ((sub1.col1 = sub2.col1)) LIMIT 100", transformedSQL)
}

func TestServer_UpdateLimit_SELECT(t *testing.T) {
	t.Parallel()
	olap := prepareOLAPStore(t)
	err := olap.Exec(context.Background(), &drivers.Statement{
		Query: "CREATE TABLE tbl1 (col1 int)",
	})
	require.NoError(t, err)

	transformedSQL, err := ensureLimits(context.Background(), olap, "SELECT col1 FROM (SELECT col1 FROM tbl1 LIMIT 2000) AS sub1 INNER JOIN (SELECT col1 FROM tbl1 LIMIT 2000) AS sub2 ON ((sub1.col1 = sub2.col1)) LIMIT 2000", 100)
	require.NoError(t, err)
	require.Equal(t, "SELECT col1 FROM (SELECT col1 FROM tbl1 LIMIT 2000) AS sub1 INNER JOIN (SELECT col1 FROM tbl1 LIMIT 2000) AS sub2 ON ((sub1.col1 = sub2.col1)) LIMIT 100", transformedSQL)
}

func TestServer_InsertLimit_WITH(t *testing.T) {
	t.Parallel()
	olap := prepareOLAPStore(t)
	err := olap.Exec(context.Background(), &drivers.Statement{
		Query: "CREATE TABLE tbl1 (col1 int)",
	})
	require.NoError(t, err)

	transformedSQL, err := ensureLimits(context.Background(), olap, "WITH tbl2 AS (SELECT col1 FROM tbl1), tbl3 AS (SELECT col1 FROM tbl1) SELECT col1 FROM tbl2 UNION ALL SELECT col1 FROM tbl3", 100)
	require.NoError(t, err)
	require.Equal(t, "WITH tbl2 AS (SELECT col1 FROM tbl1), tbl3 AS (SELECT col1 FROM tbl1)(SELECT col1 FROM tbl2) UNION ALL (SELECT col1 FROM tbl3) LIMIT 100", transformedSQL)
}

func TestServer_UpdateLimit_WITH(t *testing.T) {
	t.Parallel()
	olap := prepareOLAPStore(t)
	err := olap.Exec(context.Background(), &drivers.Statement{
		Query: "CREATE TABLE tbl1 (col1 int)",
	})
	require.NoError(t, err)

	transformedSQL, err := ensureLimits(context.Background(), olap, "WITH tbl2 AS (SELECT col1 FROM tbl1 LIMIT 2000), tbl3 AS (SELECT col1 FROM tbl1 LIMIT 2000)(SELECT col1 FROM tbl2 LIMIT 2000) UNION ALL (SELECT col1 FROM tbl3 LIMIT 2000) LIMIT 2000", 100)
	require.NoError(t, err)
	require.Equal(t, "WITH tbl2 AS (SELECT col1 FROM tbl1 LIMIT 2000), tbl3 AS (SELECT col1 FROM tbl1 LIMIT 2000)(SELECT col1 FROM tbl2 LIMIT 2000) UNION ALL (SELECT col1 FROM tbl3 LIMIT 2000) LIMIT 100", transformedSQL)
}

func TestServer_InsertLimit_SELECT_WHERE(t *testing.T) {
	t.Parallel()
	olap := prepareOLAPStore(t)
	err := olap.Exec(context.Background(), &drivers.Statement{
		Query: "CREATE TABLE tbl1 (col1 int)",
	})
	require.NoError(t, err)

	transformedSQL, err := ensureLimits(context.Background(), olap, "SELECT col1 FROM tbl1 WHERE col1 = 1 ORDER BY 1", 100)
	require.NoError(t, err)
	require.Equal(t, "SELECT col1 FROM tbl1 WHERE (col1 = 1) ORDER BY 1 LIMIT 100", transformedSQL)
}

func TestServer_UpdateLimit_SELECT_WHERE(t *testing.T) {
	t.Parallel()
	olap := prepareOLAPStore(t)
	err := olap.Exec(context.Background(), &drivers.Statement{
		Query: "CREATE TABLE tbl1 (col1 int)",
	})
	require.NoError(t, err)

	transformedSQL, err := ensureLimits(context.Background(), olap, "SELECT col1 FROM tbl1 WHERE (col1 = 1) ORDER BY 1 LIMIT 2000", 100)
	require.NoError(t, err)
	require.Equal(t, "SELECT col1 FROM tbl1 WHERE (col1 = 1) ORDER BY 1 LIMIT 100", transformedSQL)
}

func TestServer_UpdateLimit_args(t *testing.T) {
	t.Parallel()
	olap := prepareOLAPStore(t)
	err := olap.Exec(context.Background(), &drivers.Statement{
		Query: "CREATE TABLE tbl1 (col1 int)",
	})
	require.NoError(t, err)

	transformedSQL, err := ensureLimits(context.Background(), olap, "SELECT col1 FROM tbl1 WHERE col1 = ? ORDER BY 1 LIMIT 2000", 100)
	require.NoError(t, err)
	require.Equal(t, "SELECT col1 FROM tbl1 WHERE (col1 = $1) ORDER BY 1 LIMIT 100", transformedSQL)

	transformedSQL, err = ensureLimits(context.Background(), olap, "SELECT col1 FROM tbl1 WHERE col1 = $1 ORDER BY 1 LIMIT 2000", 100)
	require.NoError(t, err)
	require.Equal(t, "SELECT col1 FROM tbl1 WHERE (col1 = $1) ORDER BY 1 LIMIT 100", transformedSQL)
}

func TestServer_UpdateLimit_LIMIT_args(t *testing.T) {
	t.Parallel()
	olap := prepareOLAPStore(t)
	err := olap.Exec(context.Background(), &drivers.Statement{
		Query: "CREATE TABLE tbl1 (col1 int)",
	})
	require.NoError(t, err)

	transformedSQL, err := ensureLimits(context.Background(), olap, "SELECT col1 FROM tbl1 WHERE col1 = 1 ORDER BY 1 LIMIT ?", 100)
	require.NoError(t, err)
	require.Equal(t, "SELECT col1 FROM tbl1 WHERE (col1 = 1) ORDER BY 1 LIMIT 100", transformedSQL)
}

func TestServer_UpdateLimit_UNION(t *testing.T) {
	t.Parallel()
	olap := prepareOLAPStore(t)
	err := olap.Exec(context.Background(), &drivers.Statement{
		Query: "CREATE TABLE tbl1 (col1 int)",
	})
	require.NoError(t, err)

	transformedSQL, err := ensureLimits(context.Background(), olap, "SELECT col1 FROM tbl1 UNION ALL SELECT col1 FROM tbl1", 100)
	require.NoError(t, err)
	require.Equal(t, "(SELECT col1 FROM tbl1) UNION ALL (SELECT col1 FROM tbl1) LIMIT 100", transformedSQL)
}

func prepareOLAPStore(t *testing.T) drivers.OLAPStore {
	conn, err := drivers.Open("duckdb", map[string]any{"dsn": "?access_mode=read_write", "data_dir": t.TempDir(), "pool_size": 4}, false, activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	olap, ok := conn.AsOLAP("")
	require.True(t, ok)
	return olap
}
