package clickhouse

import (
	"context"
	"fmt"
	"testing"

	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/clickhouse"
	"go.uber.org/zap"
)

func TestRenameTable(t *testing.T) {
	if testing.Short() {
		t.Skip("clickhouse: skipping test in short mode")
	}

	ctx := context.Background()
	clickHouseContainer, err := clickhouse.RunContainer(ctx,
		testcontainers.WithImage("clickhouse/clickhouse-server:latest"),
		clickhouse.WithUsername("clickhouse"),
		clickhouse.WithPassword("clickhouse"),
		clickhouse.WithConfigFile("../../testruntime/testdata/clickhouse-config.xml"),
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		err := clickHouseContainer.Terminate(ctx)
		require.NoError(t, err)
	})

	host, err := clickHouseContainer.Host(ctx)
	require.NoError(t, err)
	port, err := clickHouseContainer.MappedPort(ctx, "9000/tcp")
	require.NoError(t, err)

	conn, err := driver{}.Open("default", map[string]any{"dsn": fmt.Sprintf("clickhouse://clickhouse:clickhouse@%v:%v", host, port.Port())}, activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	prepareConn(t, conn)

	olap, ok := conn.AsOLAP("default")
	require.True(t, ok)
	err = olap.RenameTable(ctx, "foo", "foo1", false)
	require.NoError(t, err)

	err = olap.RenameTable(ctx, "foo1", "bar", false)
	require.NoError(t, err)

	var exist bool
	require.NoError(t, conn.(*connection).db.QueryRowContext(ctx, "EXISTS foo1").Scan(&exist))
	require.False(t, exist)
	require.NoError(t, conn.(*connection).db.QueryRowContext(ctx, "EXISTS foo1").Scan(&exist))
	require.False(t, exist)
}
