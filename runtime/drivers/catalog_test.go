package drivers_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func testCatalog(t *testing.T, catalog drivers.CatalogStore) {
	ctx := context.Background()
	instanceID := uuid.NewString()

	objs := catalog.FindEntries(ctx, instanceID, drivers.ObjectTypeSource)
	require.Len(t, objs, 0)

	obj1 := &drivers.CatalogEntry{
		Name:          "bar",
		Type:          drivers.ObjectTypeSource,
		Path:          "sources/bar.yaml",
		BytesIngested: 1029,
		Object: &runtimev1.Source{
			Name:       "bar",
			Connector:  "local_file",
			Properties: &structpb.Struct{Fields: map[string]*structpb.Value{"path": structpb.NewStringValue("/path/to/file.csv")}},
			Schema:     &runtimev1.StructType{Fields: []*runtimev1.StructType_Field{{Name: "a", Type: &runtimev1.Type{Code: runtimev1.Type_CODE_INT64}}}},
		},
	}

	obj2 := &drivers.CatalogEntry{
		Name: "foo",
		Type: drivers.ObjectTypeTable,
		Object: &runtimev1.Table{
			Name:    "foo",
			Managed: false,
			Schema:  &runtimev1.StructType{Fields: []*runtimev1.StructType_Field{{Name: "a", Type: &runtimev1.Type{Code: runtimev1.Type_CODE_INT64}}}},
		},
	}

	err := catalog.CreateEntry(ctx, instanceID, obj1)
	require.NoError(t, err)
	require.Greater(t, time.Minute, time.Since(obj1.CreatedOn))
	require.Greater(t, time.Minute, time.Since(obj1.UpdatedOn))

	err = catalog.CreateEntry(ctx, instanceID, obj1)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "duplicate key") ||
		strings.Contains(err.Error(), "Duplicate key") ||
		strings.Contains(err.Error(), "UNIQUE constraint failed"))

	err = catalog.CreateEntry(ctx, instanceID, obj2)
	require.NoError(t, err)

	objs = catalog.FindEntries(ctx, instanceID, drivers.ObjectTypeSource)
	require.Len(t, objs, 1)
	require.Equal(t, objs[0].Name, obj1.Name)
	require.Equal(t, objs[0].Type, obj1.Type)
	require.Equal(t, objs[0].Path, obj1.Path)
	require.Equal(t, objs[0].BytesIngested, obj1.BytesIngested)
	require.True(t, proto.Equal(objs[0].GetSource(), obj1.GetSource()))

	objs = catalog.FindEntries(ctx, instanceID, drivers.ObjectTypeUnspecified)
	require.Len(t, objs, 2)
	require.Equal(t, objs[0].Name, obj1.Name)
	require.Equal(t, objs[0].Type, obj1.Type)
	require.Equal(t, objs[1].Name, obj2.Name)
	require.Equal(t, objs[1].Type, obj2.Type)

	obj1.Type = drivers.ObjectTypeMetricsView
	err = catalog.UpdateEntry(ctx, instanceID, obj1)
	require.NoError(t, err)

	obj, found := catalog.FindEntry(ctx, instanceID, "bar")
	require.True(t, found)
	require.Equal(t, obj.Name, "bar")
	require.Equal(t, obj.Type, drivers.ObjectTypeMetricsView)

	err = catalog.DeleteEntry(ctx, instanceID, "bar")
	require.NoError(t, err)

	obj, found = catalog.FindEntry(ctx, instanceID, "bar")
	require.False(t, found)
	require.Nil(t, obj)
}
