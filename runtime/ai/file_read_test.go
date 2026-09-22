package ai_test

import (
	"testing"

	"github.com/rilldata/rill/runtime/ai"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func TestReadFile(t *testing.T) {
	// Setup a project with a file and a test session
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"models/test_model.sql": "SELECT 1 AS val",
		},
	})
	s := newSession(t, rt, instanceID)

	// Read file with and without a leading slash
	for _, path := range []string{"models/test_model.sql", "/models/test_model.sql"} {
		var res *ai.ReadFileResult
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.ReadFileName, &res, &ai.ReadFileArgs{Path: path})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.Equal(t, "SELECT 1 AS val", res.Contents)
	}

	// Read non-existent file
	var res *ai.ReadFileResult
	_, err := s.CallTool(t.Context(), ai.RoleUser, ai.ReadFileName, &res, &ai.ReadFileArgs{Path: "models/non_existent.sql"})
	require.Error(t, err)

	// Reject paths that traverse outside the project directory
	for _, path := range []string{"../secret.txt", "/../secret.txt", "models/../../secret.txt", "../../../../etc/passwd", "..\\..\\secret.txt"} {
		res = nil
		_, err := s.CallTool(t.Context(), ai.RoleUser, ai.ReadFileName, &res, &ai.ReadFileArgs{Path: path})
		require.ErrorContains(t, err, "must not contain", "path %q", path)
	}
}
