package resolvers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"

	_ "github.com/rilldata/rill/runtime/drivers/mock"
)

func TestGlobUnpartitioned(t *testing.T) {
	rt, instanceID := prepareGlobTest(t, "mock", map[string]string{
		"file1.csv":     ``,
		"dir/file2.csv": ``,
		"dir/file3.csv": ``,
	})

	res, err := rt.Resolve(context.Background(), &runtime.ResolveOptions{
		InstanceID: instanceID,
		Resolver:   "glob",
		ResolverProperties: map[string]any{
			"connector": "mock",
			"path":      "mock://bucket/**/*.csv",
		},
		Args:   nil,
		Claims: &runtime.SecurityClaims{},
	})
	require.NoError(t, err)
	defer res.Close()

	var rows []map[string]interface{}
	require.NoError(t, json.Unmarshal(must(res.MarshalJSON()), &rows))

	require.Len(t, rows, 3)
	require.Equal(t, "dir/file2.csv", rows[0]["path"])
	require.Equal(t, "dir/file3.csv", rows[1]["path"])
	require.Equal(t, "file1.csv", rows[2]["path"])
}

func TestGlobDirectoryPartitioned(t *testing.T) {
	rt, instanceID := prepareGlobTest(t, "mock", map[string]string{
		"dir/file1.csv":        ``,
		"dir/subdir/file2.csv": ``,
		"dir/subdir/file3.csv": ``,
	})

	res, err := rt.Resolve(context.Background(), &runtime.ResolveOptions{
		InstanceID: instanceID,
		Resolver:   "glob",
		ResolverProperties: map[string]any{
			"connector":    "mock",
			"path":         "mock://bucket/**/*.csv",
			"partition":    "directory",
			"rollup_files": true,
		},
		Args:   nil,
		Claims: &runtime.SecurityClaims{},
	})
	require.NoError(t, err)
	defer res.Close()

	var rows []map[string]interface{}
	require.NoError(t, json.Unmarshal(must(res.MarshalJSON()), &rows))

	for _, row := range rows {
		delete(row, "updated_on")
	}

	require.Equal(t, []map[string]interface{}{
		{"uri": "mock://bucket/dir", "path": "dir", "files": []any{"dir/file1.csv"}},
		{"uri": "mock://bucket/dir/subdir", "path": "dir/subdir", "files": []any{"dir/subdir/file2.csv", "dir/subdir/file3.csv"}},
	}, rows)
}

func TestGlobHivePartitioned(t *testing.T) {
	rt, instanceID := prepareGlobTest(t, "mock", map[string]string{
		"dir/year=2024/month=02/file1.csv": ``,
		"dir/year=2024/month=03/file2.csv": ``,
		"dir/year=2024/month=03/file3.csv": ``,
	})

	res, err := rt.Resolve(context.Background(), &runtime.ResolveOptions{
		InstanceID: instanceID,
		Resolver:   "glob",
		ResolverProperties: map[string]any{
			"connector": "mock",
			"path":      "mock://bucket/**/*.csv",
			"partition": "hive",
		},
		Args:   nil,
		Claims: &runtime.SecurityClaims{},
	})
	require.NoError(t, err)
	defer res.Close()

	var rows []map[string]interface{}
	require.NoError(t, json.Unmarshal(must(res.MarshalJSON()), &rows))

	for _, row := range rows {
		delete(row, "updated_on")
	}

	require.Equal(t, []map[string]interface{}{
		{"uri": "mock://bucket/dir/year=2024/month=02", "path": "dir/year=2024/month=02", "year": "2024", "month": "02"},
		{"uri": "mock://bucket/dir/year=2024/month=03", "path": "dir/year=2024/month=03", "year": "2024", "month": "03"},
	}, rows)
}

func TestGlobHivePartitionedTransformSQL(t *testing.T) {
	rt, instanceID := prepareGlobTest(t, "mock", map[string]string{
		"dir/year=2024/month=02/file1.csv": ``,
		"dir/year=2024/month=03/file2.csv": ``,
		"dir/year=2024/month=03/file3.csv": ``,
	})

	res, err := rt.Resolve(context.Background(), &runtime.ResolveOptions{
		InstanceID: instanceID,
		Resolver:   "glob",
		ResolverProperties: map[string]any{
			"connector":     "mock",
			"path":          "mock://bucket/**/*.csv",
			"partition":     "hive",
			"rollup_files":  true,
			"transform_sql": "SELECT path, LAG(path) OVER (ORDER BY path) AS prev_path, len(files) AS num_files FROM {{ .table }}",
		},
		Args:   nil,
		Claims: &runtime.SecurityClaims{},
	})
	require.NoError(t, err)
	defer res.Close()

	var rows []map[string]interface{}
	require.NoError(t, json.Unmarshal(must(res.MarshalJSON()), &rows))

	for _, row := range rows {
		delete(row, "updated_on")
	}

	require.Equal(t, []map[string]interface{}{
		{"path": "dir/year=2024/month=02", "prev_path": nil, "num_files": float64(1)},
		{"path": "dir/year=2024/month=03", "prev_path": "dir/year=2024/month=02", "num_files": float64(2)},
	}, rows)
}

func prepareGlobTest(t *testing.T, connector string, files map[string]string) (*runtime.Runtime, string) {
	// Write the provided file contents into a temporary directory.
	tempDir := t.TempDir()
	for k, v := range files {
		subdir := filepath.Dir(k)
		if subdir != "" {
			err := os.MkdirAll(filepath.Join(tempDir, subdir), 0755)
			require.NoError(t, err)
		}

		err := os.WriteFile(filepath.Join(tempDir, k), []byte(v), 0644)
		require.NoError(t, err)
	}

	// Prepare a mock_object_store connector that serves files from the temporary directory.
	connectorYAML := fmt.Sprintf(`
type: connector
name: %s
driver: mock_object_store
path: %s
`, connector, tempDir)

	// Initialize the test runtime
	rt, instanceID := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml":      ``,
			"connector.yaml": connectorYAML,
		},
	})
	testruntime.RequireReconcileState(t, rt, instanceID, 2, 0, 0)
	return rt, instanceID
}
