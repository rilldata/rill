package druid_test

import (
	"context"
	"testing"
	"time"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestScan(t *testing.T) {
	_, olap := acquireTestDruid(t)

	rows, err := olap.Execute(context.Background(), &drivers.Statement{Query: "SELECT 1, 'hello world', true, null, CAST('2024-01-01T00:00:00Z' AS TIMESTAMP)"})
	require.NoError(t, err)

	var i int
	var s string
	var b bool
	var n any
	var t1 time.Time
	require.True(t, rows.Next())
	require.NoError(t, rows.Scan(&i, &s, &b, &n, &t1))

	require.Equal(t, 1, i)
	require.Equal(t, "hello world", s)
	require.Equal(t, true, b)
	require.Nil(t, n)
	require.Equal(t, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), t1)

	require.NoError(t, rows.Close())
}

func acquireTestDruid(t *testing.T) (drivers.Handle, drivers.OLAPStore) {
	cfg := testruntime.AcquireConnector(t, "druid")
	conn, err := drivers.Open("druid", "default", cfg, activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })

	olap, ok := conn.AsOLAP("default")
	require.True(t, ok)

	return conn, olap
}
