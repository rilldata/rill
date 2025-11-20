package gcs_test

import (
	"context"
	"testing"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestObjectStore(t *testing.T) {
	cfg := testruntime.AcquireConnector(t, "gcs")
	conn, err := drivers.Open("gcs", "default", cfg, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })

	objectStore, ok := conn.AsObjectStore()
	require.True(t, ok)
	bucket := "integration-test.rilldata.com"
	t.Run("testListObjectsPagination", func(t *testing.T) { testListObjectsPagination(t, objectStore, bucket) })
	t.Run("testListObjectsDelimiter", func(t *testing.T) { testListObjectsDelimiter(t, objectStore, bucket) })
	t.Run("testListObjectsFull", func(t *testing.T) { testListObjectsFull(t, objectStore, bucket) })
	t.Run("testListObjectsEmptyPrefix", func(t *testing.T) { testListObjectsEmptyPrefix(t, objectStore, bucket) })
	t.Run("testListObjectsNoMatch", func(t *testing.T) { testListObjectsNoMatch(t, objectStore, bucket) })
}

func TestObjectStoreHMAC(t *testing.T) {
	cfg := testruntime.AcquireConnector(t, "gcs_s3_compat")
	conn, err := drivers.Open("gcs", "default", cfg, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })

	objectStore, ok := conn.AsObjectStore()
	require.True(t, ok)
	bucket := "integration-test.rilldata.com"
	t.Run("testListObjectsPagination", func(t *testing.T) { testListObjectsPagination(t, objectStore, bucket) })
	t.Run("testListObjectsDelimiter", func(t *testing.T) { testListObjectsDelimiter(t, objectStore, bucket) })
	t.Run("testListObjectsFull", func(t *testing.T) { testListObjectsFull(t, objectStore, bucket) })
	t.Run("testListObjectsEmptyPrefix", func(t *testing.T) { testListObjectsEmptyPrefix(t, objectStore, bucket) })
	t.Run("testListObjectsNoMatch", func(t *testing.T) { testListObjectsNoMatch(t, objectStore, bucket) })
}

func testListObjectsPagination(t *testing.T, objectStore drivers.ObjectStore, bucket string) {
	ctx := context.Background()
	prefix := "glob_test/"
	expected := []string{
		"glob_test/y=2010/aac.csv",
		"glob_test/y=2023/aab.csv",
		"glob_test/y=2024/aaa.csv",
		"glob_test/y=2024/bbb.csv",
	}

	pageSize := 1
	var pageToken string
	var collected []string
	pageCount := 0
	for {
		objects, nextToken, err := objectStore.ListObjects(ctx, bucket, prefix, "", uint32(pageSize), pageToken)
		require.NoError(t, err)

		pageCount++
		require.Len(t, objects, 1, "page %d should return exactly 1 object", pageCount)
		collected = append(collected, objects[0].Path)

		if pageCount < len(expected) {
			require.NotEmpty(t, nextToken, "page %d should have nextPageToken", pageCount)
		} else {
			require.Empty(t, nextToken, "last page should have empty nextPageToken")
		}

		if nextToken == "" {
			break
		}
		pageToken = nextToken
	}

	require.Equal(t, len(expected), pageCount, "unexpected number of pages")
	require.Equal(t, expected, collected, "paginated order mismatch")
}

func testListObjectsDelimiter(t *testing.T, objectStore drivers.ObjectStore, bucket string) {
	ctx := context.Background()
	prefix := "glob_test/"
	delimiter := "/"

	expected := []string{
		"glob_test/y=2010/",
		"glob_test/y=2023/",
		"glob_test/y=2024/",
	}

	objects, nextToken, err := objectStore.ListObjects(ctx, bucket, prefix, delimiter, 10, "")
	require.NoError(t, err)
	require.Empty(t, nextToken)
	require.Len(t, objects, len(expected))
	for i, obj := range objects {
		require.Equal(t, expected[i], obj.Path)
		require.True(t, obj.IsDir)
	}
}

func testListObjectsFull(t *testing.T, objectStore drivers.ObjectStore, bucket string) {
	ctx := context.Background()
	prefix := "glob_test/"

	expected := []string{
		"glob_test/y=2010/aac.csv",
		"glob_test/y=2023/aab.csv",
		"glob_test/y=2024/aaa.csv",
		"glob_test/y=2024/bbb.csv",
	}

	objects, nextToken, err := objectStore.ListObjects(ctx, bucket, prefix, "", 100, "")
	require.NoError(t, err)
	require.Empty(t, nextToken)
	require.Len(t, objects, len(expected))
	for i, obj := range objects {
		require.Equal(t, expected[i], obj.Path)
	}
}

func testListObjectsEmptyPrefix(t *testing.T, objectStore drivers.ObjectStore, bucket string) {
	ctx := context.Background()
	objects, nextToken, err := objectStore.ListObjects(ctx, bucket, "", "", 4, "")
	require.NoError(t, err)
	require.NotNil(t, objects)
	require.Len(t, objects, 4)
	require.NotEmpty(t, nextToken)
}

func testListObjectsNoMatch(t *testing.T, objectStore drivers.ObjectStore, bucket string) {
	ctx := context.Background()
	glob := "nonexistent/*"

	objects, nextToken, err := objectStore.ListObjects(ctx, bucket, glob, "", 10, "")
	require.NoError(t, err)
	require.Empty(t, objects)
	require.Empty(t, nextToken)
}
