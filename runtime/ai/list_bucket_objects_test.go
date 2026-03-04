package ai_test

import (
	"testing"

	"github.com/rilldata/rill/runtime/ai"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/rilldata/rill/runtime/testruntime/testmode"
	"github.com/stretchr/testify/require"
)

func TestListBucketObjects(t *testing.T) {
	testmode.Expensive(t)

	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		TestConnectors: []string{"s3"},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 1, 0, 0)

	s := newSession(t, rt, instanceID)

	t.Run("list objects in bucket", func(t *testing.T) {
		var res *ai.ListBucketObjectsResult
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.ListBucketObjectsName, &res, &ai.ListBucketObjectsArgs{
			Connector: "s3",
			Bucket:    "integration-test.rilldata.com",
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.NotNil(t, res.Objects)
		require.Greater(t, len(res.Objects), 0)
	})

	t.Run("list objects with path prefix", func(t *testing.T) {
		var res *ai.ListBucketObjectsResult
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.ListBucketObjectsName, &res, &ai.ListBucketObjectsArgs{
			Connector: "s3",
			Bucket:    "integration-test.rilldata.com",
			Path:      "glob_test/",
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.NotNil(t, res.Objects)
	})

	t.Run("missing connector", func(t *testing.T) {
		var res *ai.ListBucketObjectsResult
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.ListBucketObjectsName, &res, &ai.ListBucketObjectsArgs{
			Bucket: "some-bucket",
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "connector name is required")
	})

	t.Run("missing bucket", func(t *testing.T) {
		var res *ai.ListBucketObjectsResult
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.ListBucketObjectsName, &res, &ai.ListBucketObjectsArgs{
			Connector: "s3",
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "bucket name is required")
	})

	t.Run("invalid connector", func(t *testing.T) {
		var res *ai.ListBucketObjectsResult
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.ListBucketObjectsName, &res, &ai.ListBucketObjectsArgs{
			Connector: "duckdb",
			Bucket:    "some-bucket",
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "does not implement object store")
	})

	t.Run("glob matching directories", func(t *testing.T) {
		var res *ai.ListBucketObjectsResult
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.ListBucketObjectsName, &res, &ai.ListBucketObjectsArgs{
			Connector: "s3",
			Bucket:    "integration-test.rilldata.com",
			Path:      "glob_test/y=*",
			PageSize:  100,
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		// The glob should match files inside all y=* directories.
		require.Greater(t, len(res.Objects), 0)
		for _, obj := range res.Objects {
			require.Contains(t, obj.Path, "glob_test/y=")
		}
	})

	t.Run("glob matching csv files", func(t *testing.T) {
		var res *ai.ListBucketObjectsResult
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.ListBucketObjectsName, &res, &ai.ListBucketObjectsArgs{
			Connector: "s3",
			Bucket:    "integration-test.rilldata.com",
			Path:      "glob_test/y=*/*.csv",
			PageSize:  100,
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.Greater(t, len(res.Objects), 0)
		for _, obj := range res.Objects {
			require.Contains(t, obj.Path, "glob_test/y=")
			require.True(t, len(obj.Path) > 4 && obj.Path[len(obj.Path)-4:] == ".csv", "expected .csv file, got %s", obj.Path)
			require.False(t, obj.IsDir)
		}
	})

	t.Run("glob with year wildcard", func(t *testing.T) {
		var res *ai.ListBucketObjectsResult
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.ListBucketObjectsName, &res, &ai.ListBucketObjectsArgs{
			Connector: "s3",
			Bucket:    "integration-test.rilldata.com",
			Path:      "glob_test/y=202*/*",
			PageSize:  100,
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		// Should match files in y=2023 and y=2024 but not y=2010
		paths := make([]string, 0, len(res.Objects))
		for _, obj := range res.Objects {
			paths = append(paths, obj.Path)
		}
		require.ElementsMatch(t, []string{
			"glob_test/y=2023/aab.csv",
			"glob_test/y=2024/aaa.csv",
			"glob_test/y=2024/bbb.csv",
		}, paths)
	})

	t.Run("glob no match", func(t *testing.T) {
		var res *ai.ListBucketObjectsResult
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.ListBucketObjectsName, &res, &ai.ListBucketObjectsArgs{
			Connector: "s3",
			Bucket:    "integration-test.rilldata.com",
			Path:      "glob_test/y=999*/*.csv",
			PageSize:  100,
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.Empty(t, res.Objects)
	})
}
