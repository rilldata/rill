package drivers_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
)

func testCatalog(t *testing.T, catalog drivers.CatalogStore) {
	t.Run("Partitions", func(t *testing.T) { testCatalogPartitions(t, catalog) })
}

func testCatalogPartitions(t *testing.T, catalog drivers.CatalogStore) {
	ctx := context.Background()
	modelID := uuid.NewString()

	now, err := time.Parse(time.RFC3339, "2021-01-01T00:00:00Z")
	require.NoError(t, err)

	partitions, err := catalog.FindModelPartitionsByKeys(ctx, modelID, []string{})
	if errors.Is(err, drivers.ErrNotImplemented) {
		t.Skip("Partition management not implemented")
	}
	require.NoError(t, err)
	require.Len(t, partitions, 0)

	partition := drivers.ModelPartition{
		Key:       "hello",
		DataJSON:  []byte(`{"hello": "world"}`),
		Watermark: &now,
		Index:     2,
	}

	err = catalog.InsertModelPartition(ctx, modelID, partition)
	require.NoError(t, err)

	partitions, err = catalog.FindModelPartitions(ctx, &drivers.FindModelPartitionsOptions{ModelID: modelID, WherePending: true, Limit: 10})
	require.NoError(t, err)
	require.Len(t, partitions, 1)
	requirePartitionEqual(t, partition, partitions[0])

	partition.ExecutedOn = &now
	partition.Error = "something"
	partition.Elapsed = time.Second

	err = catalog.UpdateModelPartition(ctx, modelID, partition)
	require.NoError(t, err)

	partitions, err = catalog.FindModelPartitionsByKeys(ctx, modelID, []string{partition.Key})
	require.NoError(t, err)
	require.Len(t, partitions, 1)
	requirePartitionEqual(t, partition, partitions[0])

	partition2 := drivers.ModelPartition{
		Key:       "hello2",
		DataJSON:  []byte(`{"hello": "world"}`),
		Watermark: &now,
		Index:     3,
	}
	err = catalog.InsertModelPartition(ctx, modelID, partition2)
	require.NoError(t, err)

	partition2.ExecutedOn = &now
	partition2.Error = ""
	partition2.Elapsed = 2 * time.Second
	err = catalog.UpdateModelPartition(ctx, modelID, partition2)
	require.NoError(t, err)

	partitions, err = catalog.FindModelPartitions(ctx, &drivers.FindModelPartitionsOptions{ModelID: modelID, Limit: 10})
	require.NoError(t, err)
	require.Len(t, partitions, 2)
	requirePartitionEqual(t, partition, partitions[0])
	requirePartitionEqual(t, partition2, partitions[1])

	err = catalog.DeleteModelPartitions(ctx, modelID)
	require.NoError(t, err)
}

func requirePartitionEqual(t *testing.T, expected, actual drivers.ModelPartition) {
	t.Helper()
	require.Equal(t, expected.Key, actual.Key)
	require.Equal(t, expected.DataJSON, actual.DataJSON)
	require.Equal(t, expected.Index, actual.Index)
	requireTimePtrEqual(t, expected.Watermark, actual.Watermark)
	requireTimePtrEqual(t, expected.ExecutedOn, actual.ExecutedOn)
	require.Equal(t, expected.Error, actual.Error)
	require.Equal(t, expected.Elapsed, actual.Elapsed)
}

func requireTimePtrEqual(t *testing.T, expected, actual *time.Time) {
	t.Helper()
	if expected == nil {
		require.Nil(t, actual)
	} else {
		require.NotNil(t, actual)
		require.True(t, expected.Equal(*actual))
	}
}
