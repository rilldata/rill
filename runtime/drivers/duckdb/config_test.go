package duckdb

import (
	"testing"

	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
)

func TestConfig(t *testing.T) {
	cfg, err := newConfig("", nil, nil)
	require.NoError(t, err)
	require.Equal(t, "", cfg.DSN)
	require.Equal(t, 1, cfg.PoolSize)

	cfg, err = newConfig("path/to/duck.db", nil, nil)
	require.NoError(t, err)
	require.Equal(t, "path/to/duck.db", cfg.DSN)
	require.Equal(t, "path/to/duck.db", cfg.DBFilePath)
	require.Equal(t, 1, cfg.PoolSize)

	cfg, err = newConfig("path/to/duck.db?rill_pool_size=10", nil, nil)
	require.NoError(t, err)
	require.Equal(t, "path/to/duck.db", cfg.DSN)
	require.Equal(t, "path/to/duck.db", cfg.DBFilePath)
	require.Equal(t, 10, cfg.PoolSize)

	cfg, err = newConfig("path/to/duck.db?rill_pool_size=10&hello=world", nil, nil)
	require.NoError(t, err)
	require.Equal(t, "path/to/duck.db?hello=world", cfg.DSN)
	require.Equal(t, 10, cfg.PoolSize)
	require.Equal(t, "path/to/duck.db", cfg.DBFilePath)

	cfg, err = newConfig("path/to/duck.db?rill_pool_size=abc&hello=world", nil, nil)
	require.Error(t, err)

	cfg, err = newConfig("path/to/duck.db?rill_pool_size=0&hello=world", nil, nil)
	require.Error(t, err)

	cfg, err = newConfig("duck.db", nil, nil)
	require.NoError(t, err)
	require.Equal(t, "duck.db", cfg.DBFilePath)

	cfg, err = newConfig("duck.db?rill_pool_size=10", nil, nil)
	require.NoError(t, err)
	require.Equal(t, "duck.db", cfg.DBFilePath)

	client := activity.NewNoopClient()
	activityDims := []attribute.KeyValue{attribute.String("key", "value")}
	cfg, err = newConfig("path/to/duck.db", client, activityDims)
	require.NoError(t, err)
	require.Equal(t, client, cfg.Activity)
	require.Equal(t, activityDims, cfg.ActivityDims)
}
