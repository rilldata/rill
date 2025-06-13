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
			expectedErr:  "number is at least 5",
		},
		{
			name: "single WHERE test (syntactic sugar)",
			modelYAML: `
type: model
sql: SELECT range AS number FROM range(0, 10)
tests:
  - name: Number is lower than 10
    where: number >= 10
`,
			expectedTest: 1,
			expectedErr:  "",
		},
		{
			name: "multiple WHERE tests (syntactic sugar)",
			modelYAML: `
type: model
sql: SELECT range AS number FROM range(0, 10)
tests:
  - name: Number is lower than 10
    where: number >= 10
  - name: Number is at least 0
    where: number < 0
`,
			expectedTest: 2,
			expectedErr:  "",
		},
		{
			name: "failing WHERE test (syntactic sugar)",
			modelYAML: `
type: model
sql: SELECT range AS number FROM range(0, 10)
tests:
  - name: Number is at least 5
    where: number < 5
`,
			expectedTest: 1,
			expectedErr:  "number is at least 5",
		},
		{
			name: "mixed SQL and WHERE tests",
			modelYAML: `
type: model
sql: SELECT range AS number FROM range(0, 10)
tests:
  - name: Number is lower than 10
    sql: SELECT 1 FROM m1 WHERE number >= 10
  - name: Number is at least 0
    where: number < 0
`,
			expectedTest: 2,
			expectedErr:  "",
		},
		{
			name: "WHERE test with templating",
			modelYAML: `
type: model
sql: SELECT range AS number FROM range(0, 10)
tests:
  - name: Number is less than env var
    where: number >= {{ .env.my_value | int64 }}
`,
			expectedTest: 1,
			expectedErr:  "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			files := map[string]string{
				"rill.yaml":      "env:\n  my_value: 10\n",
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

// TestModelPartitionTests validates that partition-level tests work correctly
func TestModelPartitionTests(t *testing.T) {
	files := map[string]string{
		"rill.yaml": "",
		"models/m1.yaml": `
type: model
sql: SELECT range AS number, range % 2 AS partition_key FROM range(0, 10)
partitions:
  sql: SELECT DISTINCT partition_key FROM (SELECT range % 2 AS partition_key FROM range(0, 10))
partition_tests:
  - name: Partition has only even or odd numbers
    sql: SELECT 1 FROM m1 WHERE (partition_key = 0 AND number % 2 != 0) OR (partition_key = 1 AND number % 2 != 1)
  - name: Partition key is not null
    sql: SELECT 1 FROM m1 WHERE partition_key IS NULL
`,
	}
	rt, id := testruntime.NewInstance(t)
	testruntime.PutFiles(t, rt, id, files)
	testruntime.ReconcileParserAndWait(t, rt, id)
	res := testruntime.GetResource(t, rt, id, runtime.ResourceKindModel, "m1")
	partitionTests := res.GetModel().Spec.GetPartitionsTests()
	require.Equal(t, 2, len(partitionTests), "Expected 2 partition_tests, got %d", len(partitionTests))

	// Check that the model reconciled without errors
	modelRes := testruntime.GetResource(t, rt, id, runtime.ResourceKindModel, "m1")
	reconcileErr := modelRes.Meta.ReconcileError
	require.Empty(t, reconcileErr, "Model reconciliation failed: %s", reconcileErr)
}

// TestModelTestsWithTemplating validates that model tests work correctly with templating
func TestModelTestsWithTemplating(t *testing.T) {
	files := map[string]string{
		"rill.yaml": `
env:
  my_value: 10
`,
		"models/m1.yaml": `
type: model
sql: |
  SELECT range AS number FROM range(0, 10)
tests:
  - name: Number is less than env var
    sql: SELECT * FROM m1 WHERE number >= {{ .env.my_value | int64 }}
`,
	}

	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: files,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	res := testruntime.GetResource(t, rt, id, runtime.ResourceKindModel, "m1")
	tests := res.GetModel().Spec.Tests
	require.Equal(t, 1, len(tests), "Expected 1 test, got %d", len(tests))

	modelRes := testruntime.GetResource(t, rt, id, runtime.ResourceKindModel, "m1")
	reconcileErr := modelRes.Meta.ReconcileError
	require.Empty(t, reconcileErr, "Model reconciliation failed: %s", reconcileErr)
}
