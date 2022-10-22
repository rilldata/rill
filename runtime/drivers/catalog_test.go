package drivers_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
)

func testCatalog(t *testing.T, catalog drivers.CatalogStore) {
	ctx := context.Background()
	instanceID := uuid.NewString()

	objs := catalog.FindObjects(ctx, instanceID)
	require.Len(t, objs, 0)

	obj := &drivers.CatalogObject{
		Name: "foo",
		Type: drivers.CatalogObjectTypeSource,
		SQL:  "CREATE SOURCE foo WITH connector = 'file', path = '/path/to/file.csv'",
	}
	err := catalog.CreateObject(ctx, instanceID, obj)
	require.NoError(t, err)
	require.Greater(t, time.Minute, time.Since(obj.CreatedOn))
	require.Greater(t, time.Minute, time.Since(obj.UpdatedOn))

	err = catalog.CreateObject(ctx, instanceID, obj)
	require.Error(t, err)
	require.Contains(t, err.Error(), "duplicate key")

	obj.Name = "bar"
	err = catalog.CreateObject(ctx, instanceID, obj)
	require.NoError(t, err)

	obj.Type = drivers.CatalogObjectTypeMetricsView
	err = catalog.UpdateObject(ctx, instanceID, obj)
	require.NoError(t, err)

	objs = catalog.FindObjects(ctx, instanceID)
	require.Len(t, objs, 2)
	require.Equal(t, objs[0].Name, "bar")
	require.Equal(t, objs[0].Type, drivers.CatalogObjectTypeMetricsView)
	require.Equal(t, objs[1].Name, "foo")
	require.Equal(t, objs[1].Type, drivers.CatalogObjectTypeSource)

	obj, found := catalog.FindObject(ctx, instanceID, "foo")
	require.True(t, found)
	require.Equal(t, obj.Name, "foo")
	require.Equal(t, obj.Type, drivers.CatalogObjectTypeSource)

	err = catalog.DeleteObject(ctx, instanceID, "foo")
	require.NoError(t, err)

	obj, found = catalog.FindObject(ctx, instanceID, "foo")
	require.False(t, found)
	require.Nil(t, obj)
}
