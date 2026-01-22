package gcs_test

import (
	"context"
	"testing"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/rilldata/rill/runtime/testruntime/testmode"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestObjectStore(t *testing.T) {
	testmode.Expensive(t)
	cfg := testruntime.AcquireConnector(t, "gcs")
	conn, err := drivers.Open("gcs", "default", cfg, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })

	objectStore, ok := conn.AsObjectStore()
	require.True(t, ok)
	bucket := "integration-test.rilldata.com"
	t.Run("testListObjectsForGlobPagination", func(t *testing.T) { testListObjectsForGlobPagination(t, objectStore, bucket) })
	t.Run("testListObjectsPagination", func(t *testing.T) { testListObjectsPagination(t, objectStore, bucket) })
	t.Run("testListObjectsDelimiter", func(t *testing.T) { testListObjectsDelimiter(t, objectStore, bucket) })
	t.Run("testListObjectsFull", func(t *testing.T) { testListObjectsFull(t, objectStore, bucket) })
	t.Run("testListObjectsEmptyPath", func(t *testing.T) { testListObjectsEmptyPath(t, objectStore, bucket) })
	t.Run("testListObjectsNoMatch", func(t *testing.T) { testListObjectsNoMatch(t, objectStore, bucket) })
}

func TestObjectStorePathPrefixes(t *testing.T) {
	testmode.Expensive(t)
	cfg := testruntime.AcquireConnector(t, "gcs")
	cfg["path_prefixes"] = "gcs://integration-test.rilldata.com/glob_test/"
	conn, err := drivers.Open("gcs", "default", cfg, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })

	objectStore, ok := conn.AsObjectStore()
	require.True(t, ok)
	bucket := "integration-test.rilldata.com"

	t.Run("testPathSameAllowedPrefix", func(t *testing.T) { testPathSameAllowedPrefix(t, objectStore, bucket) })
	t.Run("testPathWithInAllowedPrefix", func(t *testing.T) { testPathWithInAllowedPrefix(t, objectStore, bucket) })
	t.Run("testPathOutsideAllowedPrefix", func(t *testing.T) { testPathOutsideAllowedPrefix(t, objectStore, bucket) })
	t.Run("testPathRootLevelOfAllowedPrefix", func(t *testing.T) { testPathRootLevelOfAllowedPrefix(t, objectStore, bucket) })
}

func TestObjectStoreHMAC(t *testing.T) {
	testmode.Expensive(t)
	cfg := testruntime.AcquireConnector(t, "gcs_s3_compat")
	conn, err := drivers.Open("gcs", "default", cfg, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })

	objectStore, ok := conn.AsObjectStore()
	require.True(t, ok)
	bucket := "integration-test.rilldata.com"
	t.Run("testListObjectsForGlobPagination", func(t *testing.T) { testListObjectsForGlobPagination(t, objectStore, bucket) })
	t.Run("testListObjectsPagination", func(t *testing.T) { testListObjectsPagination(t, objectStore, bucket) })
	t.Run("testListObjectsDelimiter", func(t *testing.T) { testListObjectsDelimiter(t, objectStore, bucket) })
	t.Run("testListObjectsFull", func(t *testing.T) { testListObjectsFull(t, objectStore, bucket) })
	t.Run("testListObjectsEmptyPath", func(t *testing.T) { testListObjectsEmptyPath(t, objectStore, bucket) })
	t.Run("testListObjectsNoMatch", func(t *testing.T) { testListObjectsNoMatch(t, objectStore, bucket) })
}

func TestObjectStoreHMACPathPrefixes(t *testing.T) {
	testmode.Expensive(t)
	cfg := testruntime.AcquireConnector(t, "gcs_s3_compat")
	cfg["path_prefixes"] = "gcs://integration-test.rilldata.com/glob_test/"
	conn, err := drivers.Open("gcs", "default", cfg, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })

	objectStore, ok := conn.AsObjectStore()
	require.True(t, ok)
	bucket := "integration-test.rilldata.com"

	t.Run("testPathSameAllowedPrefix", func(t *testing.T) { testPathSameAllowedPrefix(t, objectStore, bucket) })
	t.Run("testPathWithInAllowedPrefix", func(t *testing.T) { testPathWithInAllowedPrefix(t, objectStore, bucket) })
	t.Run("testPathOutsideAllowedPrefix", func(t *testing.T) { testPathOutsideAllowedPrefix(t, objectStore, bucket) })
	t.Run("testPathRootLevelOfAllowedPrefix", func(t *testing.T) { testPathRootLevelOfAllowedPrefix(t, objectStore, bucket) })
}

func testListObjectsForGlobPagination(t *testing.T, objectStore drivers.ObjectStore, bucket string) {
	ctx := context.Background()
	Path := "glob_*/y=202*/*"
	expected := []string{
		"glob_test/y=2023/aab.csv",
		"glob_test/y=2024/aaa.csv",
		"glob_test/y=2024/bbb.csv",
	}

	pageSize := 1

	var pageToken string
	var collected []string
	pageCount := 0

	for {
		objects, nextToken, err := objectStore.ListObjectsForGlob(ctx, bucket, Path, uint32(pageSize), pageToken)
		require.NoError(t, err)

		if nextToken == "" {
			break
		}
		pageCount++
		require.Len(t, objects, 1)
		collected = append(collected, objects[0].Path)
		pageToken = nextToken
	}

	require.Equal(t, expected, collected)
	require.Equal(t, len(expected), pageCount)
}

func testListObjectsPagination(t *testing.T, objectStore drivers.ObjectStore, bucket string) {
	ctx := context.Background()
	Path := "glob_test/"
	expected := []string{
		"glob_test/y=2010/",
		"glob_test/y=2023/",
		"glob_test/y=2024/",
	}

	pageSize := 1
	delimiter := "/"

	var pageToken string
	var collected []string
	pageCount := 0

	for {
		objects, nextToken, err := objectStore.ListObjects(ctx, bucket, Path, delimiter, uint32(pageSize), pageToken)
		require.NoError(t, err)

		pageCount++
		require.Len(t, objects, 1)
		collected = append(collected, objects[0].Path)

		if nextToken == "" {
			break
		}
		pageToken = nextToken
	}

	require.Equal(t, expected, collected)
	require.Equal(t, len(expected), pageCount)
}

func testListObjectsDelimiter(t *testing.T, objectStore drivers.ObjectStore, bucket string) {
	ctx := context.Background()
	Path := "glob_test/"
	delimiter := "/"

	expected := []string{
		"glob_test/y=2010/",
		"glob_test/y=2023/",
		"glob_test/y=2024/",
	}

	objects, nextToken, err := objectStore.ListObjects(ctx, bucket, Path, delimiter, 10, "")
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
	Path := "glob_test/"

	expected := []string{
		"glob_test/y=2010/",
		"glob_test/y=2023/",
		"glob_test/y=2024/",
	}

	objects, nextToken, err := objectStore.ListObjects(ctx, bucket, Path, "/", 100, "")
	require.NoError(t, err)
	require.Empty(t, nextToken)
	require.Len(t, objects, len(expected))

	for i, obj := range objects {
		require.Equal(t, expected[i], obj.Path)
		require.True(t, obj.IsDir)
	}
}

func testListObjectsEmptyPath(t *testing.T, objectStore drivers.ObjectStore, bucket string) {
	ctx := context.Background()

	objects, nextToken, err := objectStore.ListObjects(ctx, bucket, "", "/", 4, "")
	require.NoError(t, err)
	require.NotNil(t, objects)
	require.Empty(t, nextToken)
}

func testListObjectsNoMatch(t *testing.T, objectStore drivers.ObjectStore, bucket string) {
	ctx := context.Background()

	objects, nextToken, err := objectStore.ListObjects(ctx, bucket, "nonexistent/", "/", 10, "")
	require.NoError(t, err)
	require.Empty(t, objects)
	require.Empty(t, nextToken)
}

func testPathSameAllowedPrefix(t *testing.T, objectStore drivers.ObjectStore, bucket string) {
	ctx := context.Background()
	path := "glob_test/"

	expected := []string{
		"glob_test/y=2010/",
		"glob_test/y=2023/",
		"glob_test/y=2024/",
	}

	objects, nextToken, err := objectStore.ListObjects(ctx, bucket, path, "/", 100, "")
	require.NoError(t, err)
	require.Empty(t, nextToken)
	require.Len(t, objects, len(expected))

	for i, obj := range objects {
		require.Equal(t, expected[i], obj.Path)
		require.True(t, obj.IsDir)
	}
}

func testPathWithInAllowedPrefix(t *testing.T, objectStore drivers.ObjectStore, bucket string) {
	ctx := context.Background()
	path := "glob_test/y=2024/"

	expected := []string{
		"glob_test/y=2024/aaa.csv",
		"glob_test/y=2024/bbb.csv",
	}

	objects, nextToken, err := objectStore.ListObjects(ctx, bucket, path, "/", 100, "")
	require.NoError(t, err)
	require.Empty(t, nextToken)
	require.Len(t, objects, len(expected))

	for i, obj := range objects {
		require.Equal(t, expected[i], obj.Path)
		require.False(t, obj.IsDir)
	}
}

func testPathOutsideAllowedPrefix(t *testing.T, objectStore drivers.ObjectStore, bucket string) {
	ctx := context.Background()
	path := "csv_test/"

	objects, nextToken, err := objectStore.ListObjects(ctx, bucket, path, "/", 10, "")
	require.Error(t, err)
	require.Empty(t, objects)
	require.Empty(t, nextToken)
}

func testPathRootLevelOfAllowedPrefix(t *testing.T, objectStore drivers.ObjectStore, bucket string) {
	ctx := context.Background()
	path := "/"

	expected := []string{
		"glob_test/",
	}

	objects, nextToken, err := objectStore.ListObjects(ctx, bucket, path, "/", 100, "")
	require.NoError(t, err)
	require.Empty(t, nextToken)
	require.Len(t, objects, len(expected))

	for i, obj := range objects {
		require.Equal(t, expected[i], obj.Path)
		require.True(t, obj.IsDir)
	}
}
