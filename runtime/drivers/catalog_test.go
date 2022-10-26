package drivers_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rilldata/rill/runtime/api"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func testCatalog(t *testing.T, catalog drivers.CatalogStore) {
	ctx := context.Background()
	instanceID := uuid.NewString()

	objs := catalog.FindObjects(ctx, instanceID, drivers.CatalogObjectTypeSource)
	require.Len(t, objs, 0)

	obj1 := &drivers.CatalogObject{
		Name:    "bar",
		Type:    drivers.CatalogObjectTypeSource,
		SQL:     "CREATE SOURCE foo WITH connector = 'file', path = '/path/to/file.csv'",
		Schema:  &api.StructType{Fields: []*api.StructType_Field{{Name: "a", Type: &api.Type{Code: api.Type_CODE_INT64}}}},
		Managed: true,
	}

	obj2 := &drivers.CatalogObject{
		Name:    "baz",
		Type:    drivers.CatalogObjectTypeSource,
		SQL:     "CREATE SOURCE foo WITH connector = 'file', path = '/path/to/file.csv'",
		Schema:  &api.StructType{Fields: []*api.StructType_Field{{Name: "a", Type: &api.Type{Code: api.Type_CODE_INT64}}}},
		Managed: true,
	}

	obj3 := &drivers.CatalogObject{
		Name:    "foo",
		Type:    drivers.CatalogObjectTypeTable,
		SQL:     "",
		Schema:  &api.StructType{Fields: []*api.StructType_Field{{Name: "a", Type: &api.Type{Code: api.Type_CODE_INT64}}}},
		Managed: false,
	}

	err := catalog.CreateObject(ctx, instanceID, obj1)
	require.NoError(t, err)
	require.Greater(t, time.Minute, time.Since(obj1.CreatedOn))
	require.Greater(t, time.Minute, time.Since(obj1.UpdatedOn))

	err = catalog.CreateObject(ctx, instanceID, obj1)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "UNIQUE constraint failed"))

	err = catalog.CreateObject(ctx, instanceID, obj2)
	require.NoError(t, err)

	err = catalog.CreateObject(ctx, instanceID, obj3)
	require.NoError(t, err)

	objs = catalog.FindObjects(ctx, instanceID, drivers.CatalogObjectTypeSource)
	require.Len(t, objs, 2)
	require.Equal(t, objs[0].Name, obj1.Name)
	require.Equal(t, objs[0].Type, obj1.Type)
	require.Equal(t, objs[0].SQL, obj1.SQL)
	require.True(t, proto.Equal(objs[0].Schema, obj1.Schema))
	require.Equal(t, objs[0].Managed, obj1.Managed)
	require.Equal(t, objs[1].Name, obj2.Name)
	require.Equal(t, objs[1].Type, obj2.Type)
	require.Equal(t, objs[1].SQL, obj2.SQL)
	require.True(t, proto.Equal(objs[1].Schema, obj2.Schema))
	require.Equal(t, objs[1].Managed, obj2.Managed)

	objs = catalog.FindObjects(ctx, instanceID, drivers.CatalogObjectTypeUnspecified)
	require.Len(t, objs, 3)
	require.Equal(t, objs[0].Name, obj1.Name)
	require.Equal(t, objs[0].Type, obj1.Type)
	require.Equal(t, objs[1].Name, obj2.Name)
	require.Equal(t, objs[1].Type, obj2.Type)
	require.Equal(t, objs[2].Name, obj3.Name)
	require.Equal(t, objs[2].Type, obj3.Type)

	obj1.Type = drivers.CatalogObjectTypeMetricsView
	err = catalog.UpdateObject(ctx, instanceID, obj1)
	require.NoError(t, err)

	obj, found := catalog.FindObject(ctx, instanceID, "bar")
	require.True(t, found)
	require.Equal(t, obj.Name, "bar")
	require.Equal(t, obj.Type, drivers.CatalogObjectTypeMetricsView)

	err = catalog.DeleteObject(ctx, instanceID, "bar")
	require.NoError(t, err)

	obj, found = catalog.FindObject(ctx, instanceID, "bar")
	require.False(t, found)
	require.Nil(t, obj)
}
