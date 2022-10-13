package server

import (
	"context"
	"testing"

	"github.com/google/uuid"
	_ "github.com/rilldata/rill/runtime/drivers/sqlite"
	"github.com/stretchr/testify/require"
)

func TestConnectionCache(t *testing.T) {
	ctx := context.Background()
	id := uuid.NewString()

	c := newConnectionCache(10)
	conn1, err := c.openAndMigrate(ctx, id, "sqlite", ":memory:")
	require.NoError(t, err)
	require.NotNil(t, conn1)

	conn2, err := c.openAndMigrate(ctx, id, "sqlite", ":memory:")
	require.NoError(t, err)
	require.NotNil(t, conn2)

	conn3, err := c.openAndMigrate(ctx, uuid.NewString(), "sqlite", ":memory:")
	require.NoError(t, err)
	require.NotNil(t, conn3)

	require.True(t, conn1 == conn2)
	require.False(t, conn2 == conn3)
}
