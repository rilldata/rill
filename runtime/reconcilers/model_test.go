package reconcilers_test

import (
	"testing"

	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

// TestModelSQLTests validates that model tests are correctly reconciled
func TestModelSQLTests(t *testing.T) {
	testCases := []struct {
		name         string
		modelYAML    string
		expectedTest int
		expectedErr  string
	}{
		{
			name: "no SQL tests",
			modelYAML: `
type: model
sql: SELECT range AS number FROM range(0, 10)
`,
			expectedTest: 0,
			expectedErr:  "",
		},
		{
			name: "single SQL test",
			modelYAML: `
type: model
sql: SELECT range AS number FROM range(0, 10)
tests:
  - name: Number is lower than 10
    sql: SELECT 1 FROM m1 WHERE number >= 10
`,
			expectedTest: 1,
			expectedErr:  "",
		},
		{
			name: "multiple SQL tests",
			modelYAML: `
type: model
sql: SELECT range AS number FROM range(0, 10)
tests:
  - name: Number is lower than 10
    sql: SELECT 1 FROM m1 WHERE number >= 10
  - name: Number is at least 0
    sql: SELECT 1 FROM m1 WHERE number < 0
`,
			expectedTest: 2,
			expectedErr:  "",
		},
		{
			name: "failing SQL test",
			modelYAML: `
type: model
sql: SELECT range AS number FROM range(0, 10)
tests:
  - name: Number is at least 5
    sql: SELECT 1 FROM m1 WHERE number < 5
`,
			expectedTest: 1,
			expectedErr:  "Number is at least 5",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			files := map[string]string{
				"rill.yaml":      "",
				"models/m1.yaml": tc.modelYAML,
			}
			rt, id := testruntime.NewInstance(t)
			testruntime.PutFiles(t, rt, id, files)
			testruntime.ReconcileParserAndWait(t, rt, id)
			res := testruntime.GetResource(t, rt, id, runtime.ResourceKindModel, "m1")
			tests := res.GetModel().Spec.Tests
			require.Equal(t, tc.expectedTest, len(tests), "Expected %d tests, got %d", tc.expectedTest, len(tests))

			// Check that the model reconciled
			modelRes := testruntime.GetResource(t, rt, id, runtime.ResourceKindModel, "m1")
			reconcileErr := modelRes.Meta.ReconcileError
			if tc.expectedErr == "" {
				require.Empty(t, reconcileErr, "Model reconciliation failed: %s", reconcileErr)
			} else {
				require.Contains(t, reconcileErr, tc.expectedErr, "Expected error message to contain: %s", tc.expectedErr)
			}
		})
	}
}

// TestModelSQLTestsWithPartition validates that model tests work correctly with partitions
func TestModelSQLTestsWithPartition(t *testing.T) {
	files := map[string]string{
		"rill.yaml": "",
		"models/m1.yaml": `
type: model
sql: SELECT range AS number, range % 2 AS partition_key FROM range(0, 10)
partition_by: partition_key
tests:
  - name: All numbers are less than 10
    sql: SELECT 1 FROM m1 WHERE number >= 10
  - name: Partition key is 0 or 1
    sql: SELECT 1 FROM m1 WHERE partition_key NOT IN (0, 1)
`,
	}
	rt, id := testruntime.NewInstance(t)
	testruntime.PutFiles(t, rt, id, files)
	testruntime.ReconcileParserAndWait(t, rt, id)
	res := testruntime.GetResource(t, rt, id, runtime.ResourceKindModel, "m1")
	tests := res.GetModel().Spec.Tests
	require.Equal(t, 2, len(tests), "Expected 2 tests, got %d", len(tests))

	// Check that the model reconciled without errors
	modelRes := testruntime.GetResource(t, rt, id, runtime.ResourceKindModel, "m1")
	reconcileErr := modelRes.Meta.ReconcileError
	require.Empty(t, reconcileErr, "Model reconciliation failed: %s", reconcileErr)
}
