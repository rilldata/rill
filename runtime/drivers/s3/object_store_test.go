package s3_test

import (
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
	cfg := testruntime.AcquireConnector(t, "s3")
	conn, err := drivers.Open("s3", "", "default", cfg, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })

	objectStore, ok := conn.AsObjectStore()
	require.True(t, ok)
	bucket := "integration-test.rilldata.com"
	t.Run("testListObjectsForGlobPagination_pageSize1", func(t *testing.T) { testListObjectsForGlobPagination(t, objectStore, bucket, 1) })
	t.Run("testListObjectsForGlobPagination_pageSize2", func(t *testing.T) { testListObjectsForGlobPagination(t, objectStore, bucket, 2) })
	t.Run("testListObjectsForGlobPagination_pageSize3", func(t *testing.T) { testListObjectsForGlobPagination(t, objectStore, bucket, 3) })
	t.Run("testListObjectsForGlobPagination_pageSize5", func(t *testing.T) { testListObjectsForGlobPagination(t, objectStore, bucket, 5) })
	t.Run("testListDirectoriesForGlobPagination_pageSize1", func(t *testing.T) { testListDirectoriesForGlobPagination(t, objectStore, bucket, 1) })
	t.Run("testListDirectoriesForGlobPagination_pageSize2", func(t *testing.T) { testListDirectoriesForGlobPagination(t, objectStore, bucket, 2) })
	t.Run("testListDirectoriesForGlobPagination_pageSize3", func(t *testing.T) { testListDirectoriesForGlobPagination(t, objectStore, bucket, 3) })
	t.Run("testListDirectoriesForGlobPagination_pageSize5", func(t *testing.T) { testListDirectoriesForGlobPagination(t, objectStore, bucket, 5) })
	t.Run("testListMonthDirectoriesForGlobPagination_pageSize1", func(t *testing.T) { testListMonthDirectoriesForGlobPagination(t, objectStore, bucket, 1) })
	t.Run("testListMonthDirectoriesForGlobPagination_pageSize2", func(t *testing.T) { testListMonthDirectoriesForGlobPagination(t, objectStore, bucket, 2) })
	t.Run("testListMonthDirectoriesForGlobPagination_pageSize3", func(t *testing.T) { testListMonthDirectoriesForGlobPagination(t, objectStore, bucket, 3) })
	t.Run("testListMonthDirectoriesForGlobPagination_pageSize5", func(t *testing.T) { testListMonthDirectoriesForGlobPagination(t, objectStore, bucket, 5) })
	t.Run("testListDayDirectoriesForGlobPagination_pageSize1", func(t *testing.T) { testListDayDirectoriesForGlobPagination(t, objectStore, bucket, 1) })
	t.Run("testListDayDirectoriesForGlobPagination_pageSize2", func(t *testing.T) { testListDayDirectoriesForGlobPagination(t, objectStore, bucket, 2) })
	t.Run("testListDayDirectoriesForGlobPagination_pageSize3", func(t *testing.T) { testListDayDirectoriesForGlobPagination(t, objectStore, bucket, 3) })
	t.Run("testListDayDirectoriesForGlobPagination_pageSize5", func(t *testing.T) { testListDayDirectoriesForGlobPagination(t, objectStore, bucket, 5) })
	t.Run("testMatchDirectoriesFromGlobTest", func(t *testing.T) { testMatchDirectoriesFromGlobTest(t, objectStore, bucket) })
	t.Run("testMatchFilesWithLeafWildcardGlobTest", func(t *testing.T) { testMatchFilesWithLeafWildcardGlobTest(t, objectStore, bucket) })
	t.Run("testMatchFilesWithDoubleStarGlobTest", func(t *testing.T) { testMatchFilesWithDoubleStarGlobTest(t, objectStore, bucket) })
	t.Run("testTrailingSlashNormalized", func(t *testing.T) { testTrailingSlashNormalized(t, objectStore, bucket) })
	t.Run("testGlobIgnoresNonCSVFiles", func(t *testing.T) { testGlobIgnoresNonCSVFiles(t, objectStore, bucket) })
	t.Run("testListObjectsPagination", func(t *testing.T) { testListObjectsPagination(t, objectStore, bucket) })
	t.Run("testListObjectsDelimiter", func(t *testing.T) { testListObjectsDelimiter(t, objectStore, bucket) })
	t.Run("testListObjectsFull", func(t *testing.T) { testListObjectsFull(t, objectStore, bucket) })
	t.Run("testListObjectsEmptyPath", func(t *testing.T) { testListObjectsEmptyPath(t, objectStore, bucket) })
	t.Run("testListObjectsNoMatch", func(t *testing.T) { testListObjectsNoMatch(t, objectStore, bucket) })
}

func TestObjectStorePathPrefixes(t *testing.T) {
	testmode.Expensive(t)
	cfg := testruntime.AcquireConnector(t, "s3")
	cfg["path_prefixes"] = "s3://integration-test.rilldata.com/glob_test/"
	conn, err := drivers.Open("s3", "", "default", cfg, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
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
	ctx := t.Context()
	path := "listing_glob_test/**/*.csv"

	expected := []string{
		"listing_glob_test/year=2024/month=11/day=30/file1.csv",
		"listing_glob_test/year=2024/month=11/day=30/file2.csv",
		"listing_glob_test/year=2024/month=11/day=31/file1.csv",

		"listing_glob_test/year=2025/month=01/day=01/file1.csv",
		"listing_glob_test/year=2025/month=01/day=01/file2.csv",
		"listing_glob_test/year=2025/month=01/day=01/file3.csv",

		"listing_glob_test/year=2025/month=01/day=02/file1.csv",
		"listing_glob_test/year=2025/month=01/day=02/file2.csv",

		"listing_glob_test/year=2025/month=02/day=01/file1.csv",

		"listing_glob_test/year=2025/month=12/day=01/file1.csv",
		"listing_glob_test/year=2025/month=12/day=01/file2.csv",

		"listing_glob_test/year=2026/month=01/day=01/file1.csv",
		"listing_glob_test/year=2026/month=01/day=01/hour=01/file1.csv",
		"listing_glob_test/year=2026/month=01/day=01/hour=02/file1.csv",
	}

	var pageToken string
	var collected []string
	var pageCount int

	for {
		objects, nextToken, err := objectStore.ListObjectsForGlob(ctx, bucket, path, pageSize, pageToken, "", "")
		require.NoError(t, err)

		if nextToken != "" {
			require.Len(t, objects, int(pageSize))
		} else {
			require.NotEmpty(t, objects)
			require.LessOrEqual(t, len(objects), int(pageSize))
		}

		for _, obj := range objects {
			require.False(t, obj.IsDir)
			collected = append(collected, obj.Path)
		}

		if nextToken == "" {
			break
		}

		pageToken = nextToken
		pageCount++
	}

	require.Equal(t, expected, collected)

	expectedPages := (len(expected) + int(pageSize) - 1) / int(pageSize)
	require.Equal(t, expectedPages, pageCount+1)
}

func testMatchDirectoriesFromGlobTest(t *testing.T, objectStore drivers.ObjectStore, bucket string) {
	ctx := t.Context()

	path := "listing_glob_test/year=*"

	objects, nextToken, err := objectStore.ListObjectsForGlob(ctx, bucket, path, 100, "", "", "")

	require.NoError(t, err)
	require.Empty(t, nextToken)

	expected := []string{
		"listing_glob_test/year=2024/",
		"listing_glob_test/year=2025/",
		"listing_glob_test/year=2026/",
	}

	var collected []string
	for _, obj := range objects {
		require.True(t, obj.IsDir)
		collected = append(collected, obj.Path)
	}

	require.Equal(t, expected, collected)
}

func testMatchFilesWithLeafWildcardGlobTest(t *testing.T, objectStore drivers.ObjectStore, bucket string) {
	ctx := t.Context()

	path := "listing_glob_test/year=*/month=*/day=*/*.csv"

	objects, nextToken, err := objectStore.ListObjectsForGlob(ctx, bucket, path, 100, "", "", "")
	require.NoError(t, err)
	require.Empty(t, nextToken)

	var collected []string
	for _, obj := range objects {
		require.False(t, obj.IsDir)
		collected = append(collected, obj.Path)
	}

	require.Contains(t, collected, "listing_glob_test/year=2025/month=01/day=01/file1.csv")
	require.Contains(t, collected, "listing_glob_test/year=2025/month=01/day=02/file1.csv")
	require.Contains(t, collected, "listing_glob_test/year=2026/month=01/day=01/file1.csv")
}

func testMatchFilesWithDoubleStarGlobTest(t *testing.T, objectStore drivers.ObjectStore, bucket string) {
	ctx := t.Context()

	path := "listing_glob_test/**"

	objects, _, err := objectStore.ListObjectsForGlob(ctx, bucket, path, 100, "", "", "")
	require.NoError(t, err)

	var fileCount int
	for _, obj := range objects {
		if !obj.IsDir {
			fileCount++
		}
	}

	// Ensure recursive traversal finds files in nested directories
	require.GreaterOrEqual(t, fileCount, 10)
}

func testListDirectoriesForGlobPagination(t *testing.T, objectStore drivers.ObjectStore, bucket string, pageSize uint32) {
	ctx := t.Context()

	path := "listing_glob_test/year=*"

	expected := []string{
		"listing_glob_test/year=2024/",
		"listing_glob_test/year=2025/",
		"listing_glob_test/year=2026/",
	}

	var pageToken string
	var collected []string
	var pageCount int

	for {
		objects, nextToken, err := objectStore.ListObjectsForGlob(ctx, bucket, path, pageSize, pageToken, "", "")
		require.NoError(t, err)

		if nextToken != "" {
			require.Len(t, objects, int(pageSize))
		}

		for _, obj := range objects {
			require.True(t, obj.IsDir)
			collected = append(collected, obj.Path)
		}

		if nextToken == "" {
			break
		}

		pageToken = nextToken
		pageCount++
	}

	require.Equal(t, expected, collected)

	expectedPages := (len(expected) + int(pageSize) - 1) / int(pageSize)
	require.Equal(t, expectedPages, pageCount+1)
}

func testListMonthDirectoriesForGlobPagination(t *testing.T, objectStore drivers.ObjectStore, bucket string, pageSize uint32) {
	ctx := t.Context()

	path := "listing_glob_test/year=*/month=*"

	expected := []string{
		"listing_glob_test/year=2024/month=11/",
		"listing_glob_test/year=2025/month=01/",
		"listing_glob_test/year=2025/month=02/",
		"listing_glob_test/year=2025/month=12/",
		"listing_glob_test/year=2026/month=01/",
	}

	var pageToken string
	var collected []string

	for {
		objects, nextToken, err := objectStore.ListObjectsForGlob(ctx, bucket, path, pageSize, pageToken, "", "")
		require.NoError(t, err)

		for _, obj := range objects {
			require.True(t, obj.IsDir)
			collected = append(collected, obj.Path)
		}

		if nextToken == "" {
			break
		}

		pageToken = nextToken
	}

	require.Equal(t, expected, collected)
}

func testListDayDirectoriesForGlobPagination(t *testing.T, objectStore drivers.ObjectStore, bucket string, pageSize uint32) {
	ctx := t.Context()
	path := "listing_glob_test/*/month=*/*"

	expected := []string{
		// "listing_glob_test/year=2024/month=11/day=30/",
		"listing_glob_test/year=2024/month=11/day=31/",
		"listing_glob_test/year=2025/month=01/day=01/",
		"listing_glob_test/year=2025/month=01/day=02/",
		"listing_glob_test/year=2025/month=02/day=01/",
		"listing_glob_test/year=2025/month=12/day=01/",
		"listing_glob_test/year=2026/month=01/day=01/",
	}

	var pageToken string
	var collected []string
	var pageCount int

	for {
		objects, nextToken, err := objectStore.ListObjectsForGlob(ctx, bucket, path, pageSize, pageToken, "listing_glob_test/year=2024/month=11/day=31/", "")
		require.NoError(t, err)

		if nextToken != "" {
			require.Len(t, objects, int(pageSize))
		} else {
			require.NotEmpty(t, objects)
			require.LessOrEqual(t, len(objects), int(pageSize))
		}

		for _, obj := range objects {
			require.True(t, obj.IsDir)
			collected = append(collected, obj.Path)
		}

		if nextToken == "" {
			break
		}

		pageToken = nextToken
		pageCount++
	}

	require.Equal(t, expected, collected)

	expectedPages := (len(expected) + int(pageSize) - 1) / int(pageSize)
	require.Equal(t, expectedPages, pageCount+1)
}

func testGlobIgnoresNonCSVFiles(t *testing.T, objectStore drivers.ObjectStore, bucket string) {
	ctx := t.Context()

	path := "listing_glob_test/**/*.csv"

	objects, _, err := objectStore.ListObjectsForGlob(ctx, bucket, path, 100, "", "", "")
	require.NoError(t, err)

	for _, obj := range objects {
		require.NotContains(t, obj.Path, ".txt")
		require.NotContains(t, obj.Path, "_SUCCESS")
		require.NotContains(t, obj.Path, "_metadata")
	}
}

func testTrailingSlashNormalized(t *testing.T, objectStore drivers.ObjectStore, bucket string) {
	ctx := t.Context()
	pathWithSlash := "glob_test/y=*/"
	pathWithoutSlash := "glob_test/y=*"

	objsWithSlash, _, err := objectStore.ListObjectsForGlob(ctx, bucket, pathWithSlash, 100, "", "", "")
	require.NoError(t, err)

	objsWithoutSlash, _, err := objectStore.ListObjectsForGlob(ctx, bucket, pathWithoutSlash, 100, "", "", "")
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
	ctx := t.Context()
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
	ctx := t.Context()
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
	ctx := t.Context()
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
	ctx := t.Context()

	objects, nextToken, err := objectStore.ListObjects(ctx, bucket, "", "/", 2, "")
	require.NoError(t, err)
	require.NotNil(t, objects)
	require.Len(t, objects, 2)
	require.NotEmpty(t, nextToken)
}

func testListObjectsNoMatch(t *testing.T, objectStore drivers.ObjectStore, bucket string) {
	ctx := t.Context()

	objects, nextToken, err := objectStore.ListObjects(ctx, bucket, "nonexistent/", "/", 10, "")
	require.NoError(t, err)
	require.Empty(t, objects)
	require.Empty(t, nextToken)
}

func testPathSameAllowedPrefix(t *testing.T, objectStore drivers.ObjectStore, bucket string) {
	ctx := t.Context()
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
	ctx := t.Context()
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
	ctx := t.Context()
	path := "csv_test/"

	objects, nextToken, err := objectStore.ListObjects(ctx, bucket, path, "/", 10, "")
	require.Error(t, err)
	require.Empty(t, objects)
	require.Empty(t, nextToken)
}

func testPathRootLevelOfAllowedPrefix(t *testing.T, objectStore drivers.ObjectStore, bucket string) {
	ctx := t.Context()
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
