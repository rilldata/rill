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
}
