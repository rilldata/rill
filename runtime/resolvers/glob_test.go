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

func TestGlobTrimsWhitespace(t *testing.T) {
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
			"path":      "\n mock://bucket/**/*.csv \n",
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
			"transform_sql": "SELECT path, LAG(path) OVER (ORDER BY path) AS prev_path, len(files) AS num_files, updated_on FROM {{ .table }}",
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

func TestGlobPatternYYYYMMDD(t *testing.T) {
	rt, instanceID := prepareGlobTest(t, "mock", map[string]string{
		"data/2024/01/15/file1.parquet": ``,
		"data/2024/01/16/file2.parquet": ``,
		"data/2024/02/01/file3.parquet": ``,
	})

	res, err := rt.Resolve(context.Background(), &runtime.ResolveOptions{
		InstanceID: instanceID,
		Resolver:   "glob",
		ResolverProperties: map[string]any{
			"connector": "mock",
			"path":      "mock://bucket/data/*/*/*/*.parquet",
			"pattern":   "YYYY/MM/DD",
			"partition": "directory",
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
		{
			"uri":      "mock://bucket/data/2024/01/15",
			"path":     "data/2024/01/15",
			"year":     "2024",
			"month":    "01",
			"day":      "15",
			"day_path": "data/2024/01/15",
			"date":     "2024-01-15",
		},
		{
			"uri":      "mock://bucket/data/2024/01/16",
			"path":     "data/2024/01/16",
			"year":     "2024",
			"month":    "01",
			"day":      "16",
			"day_path": "data/2024/01/16",
			"date":     "2024-01-16",
		},
		{
			"uri":      "mock://bucket/data/2024/02/01",
			"path":     "data/2024/02/01",
			"year":     "2024",
			"month":    "02",
			"day":      "01",
			"day_path": "data/2024/02/01",
			"date":     "2024-02-01",
		},
	}, rows)
}

func TestGlobPatternYYYYMMDDHH(t *testing.T) {
	rt, instanceID := prepareGlobTest(t, "mock", map[string]string{
		"pmp/2025/09/01/14/file1.parquet": ``,
		"pmp/2025/09/01/15/file2.parquet": ``,
		"pmp/2025/09/02/10/file3.parquet": ``,
	})

	res, err := rt.Resolve(context.Background(), &runtime.ResolveOptions{
		InstanceID: instanceID,
		Resolver:   "glob",
		ResolverProperties: map[string]any{
			"connector": "mock",
			"path":      "mock://bucket/pmp/*/*/*/*/*.parquet",
			"pattern":   "YYYY/MM/DD/HH",
			"partition": "directory",
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
		{
			"uri":      "mock://bucket/pmp/2025/09/01/14",
			"path":     "pmp/2025/09/01/14",
			"year":     "2025",
			"month":    "09",
			"day":      "01",
			"hour":     "14",
			"day_path": "pmp/2025/09/01",
			"date":     "2025-09-01",
		},
		{
			"uri":      "mock://bucket/pmp/2025/09/01/15",
			"path":     "pmp/2025/09/01/15",
			"year":     "2025",
			"month":    "09",
			"day":      "01",
			"hour":     "15",
			"day_path": "pmp/2025/09/01",
			"date":     "2025-09-01",
		},
		{
			"uri":      "mock://bucket/pmp/2025/09/02/10",
			"path":     "pmp/2025/09/02/10",
			"year":     "2025",
			"month":    "09",
			"day":      "02",
			"hour":     "10",
			"day_path": "pmp/2025/09/02",
			"date":     "2025-09-02",
		},
	}, rows)
}

func TestGlobPatternWithEquals(t *testing.T) {
	rt, instanceID := prepareGlobTest(t, "mock", map[string]string{
		"data/y=2024/m=01/d=15/file1.parquet": ``,
		"data/y=2024/m=02/d=01/file2.parquet": ``,
	})

	res, err := rt.Resolve(context.Background(), &runtime.ResolveOptions{
		InstanceID: instanceID,
		Resolver:   "glob",
		ResolverProperties: map[string]any{
			"connector": "mock",
			"path":      "mock://bucket/data/*/*/*/*.parquet",
			"pattern":   "y=YYYY/m=MM/d=DD",
			"partition": "directory",
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
		{
			"uri":      "mock://bucket/data/y=2024/m=01/d=15",
			"path":     "data/y=2024/m=01/d=15",
			"year":     "2024",
			"month":    "01",
			"day":      "15",
			"day_path": "data/2024/01/15",
			"date":     "2024-01-15",
		},
		{
			"uri":      "mock://bucket/data/y=2024/m=02/d=01",
			"path":     "data/y=2024/m=02/d=01",
			"year":     "2024",
			"month":    "02",
			"day":      "01",
			"day_path": "data/2024/02/01",
			"date":     "2024-02-01",
		},
	}, rows)
}

func TestGlobPatternWithTransformSQL(t *testing.T) {
	rt, instanceID := prepareGlobTest(t, "mock", map[string]string{
		"pmp/2025/09/01/file1.parquet": ``,
		"pmp/2025/09/02/file2.parquet": ``,
	})

	res, err := rt.Resolve(context.Background(), &runtime.ResolveOptions{
		InstanceID: instanceID,
		Resolver:   "glob",
		ResolverProperties: map[string]any{
			"connector":     "mock",
			"path":          "mock://bucket/pmp/*/*/*/*.parquet",
			"pattern":       "YYYY/MM/DD",
			"partition":     "directory",
			"transform_sql": "SELECT day_path, date, path, uri FROM {{ .table }} ORDER BY path",
		},
		Args:   nil,
		Claims: &runtime.SecurityClaims{},
	})
	require.NoError(t, err)
	defer res.Close()

	var rows []map[string]interface{}
	require.NoError(t, json.Unmarshal(must(res.MarshalJSON()), &rows))

	require.Equal(t, []map[string]interface{}{
		{
			"day_path": "pmp/2025/09/01",
			"date":     "2025-09-01",
			"path":     "pmp/2025/09/01",
			"uri":      "mock://bucket/pmp/2025/09/01",
		},
		{
			"day_path": "pmp/2025/09/02",
			"date":     "2025-09-02",
			"path":     "pmp/2025/09/02",
			"uri":      "mock://bucket/pmp/2025/09/02",
		},
	}, rows)
}

func TestGlobPatternNoMatch(t *testing.T) {
	rt, instanceID := prepareGlobTest(t, "mock", map[string]string{
		"data/invalid/path/file1.parquet": ``,
	})

	res, err := rt.Resolve(context.Background(), &runtime.ResolveOptions{
		InstanceID: instanceID,
		Resolver:   "glob",
		ResolverProperties: map[string]any{
			"connector": "mock",
			"path":      "mock://bucket/**/*.parquet",
			"pattern":   "YYYY/MM/DD",
			"partition": "directory",
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

	// Should still return the partition but without date extraction
	require.Equal(t, []map[string]interface{}{
		{
			"uri":  "mock://bucket/data/invalid/path",
			"path": "data/invalid/path",
		},
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
