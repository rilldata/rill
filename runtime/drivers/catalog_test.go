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
	t.Run("Splits", func(t *testing.T) { testCatalogSplits(t, catalog) })
}

func testCatalogSplits(t *testing.T, catalog drivers.CatalogStore) {
	ctx := context.Background()
	modelID := uuid.NewString()

	now, err := time.Parse(time.RFC3339, "2021-01-01T00:00:00Z")
	require.NoError(t, err)

	splits, err := catalog.FindModelSplitsByKeys(ctx, modelID, []string{})
	if errors.Is(err, drivers.ErrNotImplemented) {
		t.Skip("Split management not implemented")
	}
	require.NoError(t, err)
	require.Len(t, splits, 0)

	split := drivers.ModelSplit{
		Key:       "hello",
		DataJSON:  []byte(`{"hello": "world"}`),
		Watermark: &now,
		Index:     2,
	}

	err = catalog.InsertModelSplit(ctx, modelID, split)
	require.NoError(t, err)

	splits, err = catalog.FindModelSplits(ctx, &drivers.FindModelSplitsOptions{ModelID: modelID, WherePending: true, Limit: 10})
	require.NoError(t, err)
	require.Len(t, splits, 1)
	requireSplitEqual(t, split, splits[0])

	split.ExecutedOn = &now
	split.Error = "something"
	split.Elapsed = time.Second

	err = catalog.UpdateModelSplit(ctx, modelID, split)
	require.NoError(t, err)

	splits, err = catalog.FindModelSplitsByKeys(ctx, modelID, []string{split.Key})
	require.NoError(t, err)
	require.Len(t, splits, 1)
	requireSplitEqual(t, split, splits[0])

	err = catalog.DeleteModelSplits(ctx, modelID)
	require.NoError(t, err)
}

func requireSplitEqual(t *testing.T, expected, actual drivers.ModelSplit) {
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
