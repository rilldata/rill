package resolvers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"

	_ "github.com/rilldata/rill/runtime/drivers/mock"
)

func TestGlob(t *testing.T) {
	rt, instanceID := prepareGlobTest(t, "mock", map[string]string{
		"file1.csv":      ``,
		"path/file2.csv": ``,
		"path/file3.csv": ``,
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

	log.Printf("ROWS: %v", rows)
	t.Fail()
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
