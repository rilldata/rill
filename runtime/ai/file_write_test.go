package ai_test

import (
	"testing"

	"github.com/rilldata/rill/runtime/ai"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func TestWriteFile(t *testing.T) {
	// Setup an empty project and test session
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{})
	testruntime.RequireReconcileState(t, rt, instanceID, 1, 0, 0)
	s := newSession(t, rt, instanceID)

	// Create file
	var res *ai.WriteFileResult
	_, err := s.CallTool(t.Context(), ai.RoleUser, ai.WriteFileName, &res, &ai.WriteFileArgs{
		Path:     "models/test_model.sql",
		Contents: "SELECT 1 AS val",
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.True(t, res.IsNewFile)
	require.Contains(t, res.Diff, "+SELECT 1 AS val")
	require.Len(t, res.Resources, 1)
	require.Equal(t, "test_model", res.Resources[0]["name"])
	require.Equal(t, "", res.Resources[0]["reconcile_error"])
	require.Empty(t, res.ParseError)

	// Update file with broken model
	res = nil
	_, err = s.CallTool(t.Context(), ai.RoleUser, ai.WriteFileName, &res, &ai.WriteFileArgs{
		Path:     "models/test_model.sql",
		Contents: "SELECT 2 AS val\nFROM invalid_table",
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.False(t, res.IsNewFile)
	require.Contains(t, res.Diff, "-SELECT 1 AS val")
	require.Contains(t, res.Diff, "+SELECT 2 AS val")
	require.Contains(t, res.Diff, "+FROM invalid_table")
	require.Len(t, res.Resources, 1)
	require.Equal(t, "test_model", res.Resources[0]["name"])
	require.NotEmpty(t, res.Resources[0]["reconcile_error"])
	require.Empty(t, res.ParseError)

	// Update file to fix model
	res = nil
	_, err = s.CallTool(t.Context(), ai.RoleUser, ai.WriteFileName, &res, &ai.WriteFileArgs{
		Path:     "models/test_model.sql",
		Contents: "SELECT 2 AS val",
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.False(t, res.IsNewFile)
	require.Contains(t, res.Diff, "-FROM invalid_table")
	require.Len(t, res.Resources, 1)
	require.Equal(t, "test_model", res.Resources[0]["name"])
	require.Equal(t, "", res.Resources[0]["reconcile_error"])
	require.Empty(t, res.ParseError)

	// Delete file
	res = nil
	_, err = s.CallTool(t.Context(), ai.RoleUser, ai.WriteFileName, &res, &ai.WriteFileArgs{
		Path:   "models/test_model.sql",
		Remove: true,
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.False(t, res.IsNewFile)
	require.Empty(t, res.Resources)
	require.Empty(t, res.ParseError)

	// Delete non-existent file
	res = nil
	_, err = s.CallTool(t.Context(), ai.RoleUser, ai.WriteFileName, &res, &ai.WriteFileArgs{
		Path:   "models/non_existent.sql",
		Remove: true,
	})
	require.Error(t, err)
}
