package gcs_test

import (
	"context"
	"sort"
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
	t.Run("testListObjectsForGlobPagination/pageSize=1", func(t *testing.T) { testListObjectsForGlobPagination(t, objectStore, bucket, 1) })
	t.Run("testListObjectsForGlobPagination/pageSize=2", func(t *testing.T) { testListObjectsForGlobPagination(t, objectStore, bucket, 2) })
	t.Run("testListObjectsForGlobPagination/pageSize=3", func(t *testing.T) { testListObjectsForGlobPagination(t, objectStore, bucket, 3) })
	t.Run("testMatchDirectoriesFromGlobTest", func(t *testing.T) { testMatchDirectoriesFromGlobTest(t, objectStore, bucket) })
	t.Run("testMatchFilesWithLeafWildcardGlobTest", func(t *testing.T) { testMatchFilesWithLeafWildcardGlobTest(t, objectStore, bucket) })
	t.Run("testMatchFilesWithDoubleStarGlobTest", func(t *testing.T) { testMatchFilesWithDoubleStarGlobTest(t, objectStore, bucket) })
	t.Run("testListDirectoriesForGlobPagination_pageSize1", func(t *testing.T) { testListDirectoriesForGlobPagination(t, objectStore, bucket, 1) })
	t.Run("testListDirectoriesForGlobPagination_pageSize2", func(t *testing.T) { testListDirectoriesForGlobPagination(t, objectStore, bucket, 2) })
	t.Run("testTrailingSlashNormalized", func(t *testing.T) { testTrailingSlashNormalized(t, objectStore, bucket) })
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
	t.Run("testListObjectsForGlobPagination/pageSize=1", func(t *testing.T) { testListObjectsForGlobPagination(t, objectStore, bucket, 1) })
	t.Run("testListObjectsForGlobPagination/pageSize=2", func(t *testing.T) { testListObjectsForGlobPagination(t, objectStore, bucket, 2) })
	t.Run("testListObjectsForGlobPagination/pageSize=3", func(t *testing.T) { testListObjectsForGlobPagination(t, objectStore, bucket, 3) })
	t.Run("testMatchDirectoriesFromGlobTest", func(t *testing.T) { testMatchDirectoriesFromGlobTest(t, objectStore, bucket) })
	t.Run("testMatchFilesWithLeafWildcardGlobTest", func(t *testing.T) { testMatchFilesWithLeafWildcardGlobTest(t, objectStore, bucket) })
	t.Run("testMatchFilesWithDoubleStarGlobTest", func(t *testing.T) { testMatchFilesWithDoubleStarGlobTest(t, objectStore, bucket) })
	t.Run("testListDirectoriesForGlobPagination_pageSize1", func(t *testing.T) { testListDirectoriesForGlobPagination(t, objectStore, bucket, 1) })
	t.Run("testListDirectoriesForGlobPagination_pageSize2", func(t *testing.T) { testListDirectoriesForGlobPagination(t, objectStore, bucket, 2) })
	t.Run("testTrailingSlashNormalized", func(t *testing.T) { testTrailingSlashNormalized(t, objectStore, bucket) })
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

func testListObjectsForGlobPagination(t *testing.T, objectStore drivers.ObjectStore, bucket string, pageSize uint32) {
	ctx := context.Background()
	path := "glob_test/y=202*/*"

	expected := []string{
		"glob_test/y=2023/aab.csv",
		"glob_test/y=2024/aaa.csv",
		"glob_test/y=2024/bbb.csv",
	}

	var pageToken string
	var collected []string
	var pageCount int // number of non-final pages
	for {
		objects, nextToken, err := objectStore.ListObjectsForGlob(ctx, bucket, path, pageSize, pageToken)
		require.NoError(t, err)

		if nextToken != "" {
			require.Len(t, objects, int(pageSize))
		} else {
			require.NotEmpty(t, objects)
			require.LessOrEqual(t, len(objects), int(pageSize))
		}

		for _, obj := range objects {
			collected = append(collected, obj.Path)
		}

		if nextToken == "" {
			break
		}

		pageCount++
		pageToken = nextToken
	}

	require.Equal(t, expected, collected)
	expectedPages := (len(expected) + int(pageSize) - 1) / int(pageSize)
	require.Equal(t, expectedPages, pageCount+1)
}

func testMatchDirectoriesFromGlobTest(t *testing.T, objectStore drivers.ObjectStore, bucket string) {
	ctx := context.Background()
	// Using the existing glob_test structure: glob_test/y=202X/...
	path := "glob_test/y=*"

	objects, nextToken, err := objectStore.ListObjectsForGlob(ctx, bucket, path, 100, "")
	require.NoError(t, err)
	require.Empty(t, nextToken)

	// Should match directories like: glob_test/y=2023/, glob_test/y=2024/
	expected := []string{
		"glob_test/y=2010/",
		"glob_test/y=2023/",
		"glob_test/y=2024/",
	}

	var collected []string
	for _, obj := range objects {
		require.True(t, obj.IsDir, "Expected directory, got file: %s", obj.Path)
		collected = append(collected, obj.Path)
	}

	sort.Strings(collected)
	sort.Strings(expected)
	require.Equal(t, expected, collected)
}

func testMatchFilesWithLeafWildcardGlobTest(t *testing.T, objectStore drivers.ObjectStore, bucket string) {
	ctx := context.Background()
	path := "glob_test/y=*/*"

	objects, nextToken, err := objectStore.ListObjectsForGlob(ctx, bucket, path, 100, "")
	require.NoError(t, err)
	require.Empty(t, nextToken)

	expected := []string{
		"glob_test/y=2010/aac.csv",
		"glob_test/y=2023/aab.csv",
		"glob_test/y=2024/aaa.csv",
		"glob_test/y=2024/bbb.csv",
	}

	var collected []string
	for _, obj := range objects {
		require.False(t, obj.IsDir, "Expected file, got directory: %s", obj.Path)
		collected = append(collected, obj.Path)
	}

	require.Equal(t, expected, collected)
}

func testMatchFilesWithDoubleStarGlobTest(t *testing.T, objectStore drivers.ObjectStore, bucket string) {
	ctx := context.Background()
	path := "glob_test/**/*"

	objects, _, err := objectStore.ListObjectsForGlob(ctx, bucket, path, 100, "")
	require.NoError(t, err)

	for _, obj := range objects {
		require.False(t, obj.IsDir, "Double star should match files, not directories: %s", obj.Path)
	}

	require.GreaterOrEqual(t, len(objects), 3)
}

func testListDirectoriesForGlobPagination(t *testing.T, objectStore drivers.ObjectStore, bucket string, pageSize uint32) {
	ctx := context.Background()
	path := "glob_test/y=*"

	// Expected directories based on existing test data
	expected := []string{
		"glob_test/y=2010/",
		"glob_test/y=2023/",
		"glob_test/y=2024/",
	}

	var pageToken string
	var collected []string
	var pageCount int
	for {
		objects, nextToken, err := objectStore.ListObjectsForGlob(ctx, bucket, path, pageSize, pageToken)
		require.NoError(t, err)

		if nextToken != "" {
			require.Len(t, objects, int(pageSize), "Non-final page should return exactly pageSize results")
		} else {
			require.NotEmpty(t, objects, "Final page should not be empty")
			require.LessOrEqual(t, len(objects), int(pageSize), "Final page should not exceed pageSize")
		}

		for _, obj := range objects {
			require.True(t, obj.IsDir, "Expected directory, got file: %s", obj.Path)
			collected = append(collected, obj.Path)
		}

		if nextToken == "" {
			break
		}

		pageCount++
		pageToken = nextToken
	}

	sort.Strings(collected)
	sort.Strings(expected)
	require.Equal(t, expected, collected, "Collected directories should match expected")

	expectedPages := (len(expected) + int(pageSize) - 1) / int(pageSize)
	require.Equal(t, expectedPages, pageCount+1, "Number of pages should match expected")
}

func testTrailingSlashNormalized(t *testing.T, objectStore drivers.ObjectStore, bucket string) {
	ctx := context.Background()
	pathWithSlash := "glob_test/y=*/"
	pathWithoutSlash := "glob_test/y=*"

	objsWithSlash, _, err := objectStore.ListObjectsForGlob(ctx, bucket, pathWithSlash, 100, "")
	require.NoError(t, err)

	objsWithoutSlash, _, err := objectStore.ListObjectsForGlob(ctx, bucket, pathWithoutSlash, 100, "")
	require.NoError(t, err)

	// Both should return same directories
	require.Equal(t, len(objsWithSlash), len(objsWithoutSlash))

	// Compare paths
	pathsWithSlash := make([]string, len(objsWithSlash))
	for i, obj := range objsWithSlash {
		pathsWithSlash[i] = obj.Path
	}

	pathsWithoutSlash := make([]string, len(objsWithoutSlash))
	for i, obj := range objsWithoutSlash {
		pathsWithoutSlash[i] = obj.Path
	}
	require.Equal(t, pathsWithSlash, pathsWithoutSlash)
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
	require.NotEmpty(t, nextToken)
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
