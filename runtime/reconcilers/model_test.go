package reconcilers_test

import (
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"

	_ "github.com/rilldata/rill/runtime/resolvers"
)

func TestPatchModeManualTrigger(t *testing.T) {
	rt, instanceID := testruntime.NewInstance(t)

	// Create a model with patch mode
	testruntime.PutFiles(t, rt, instanceID, map[string]string{
		"rill.yaml": ``,
		"models/patch_model.yaml": `
type: model
incremental: true
change_mode: patch
partitions:
  sql: SELECT now() AS now
sql: SELECT '{{.partition.now}}::TIMESTAMP' AS now
`,
	})
	testruntime.ReconcileParserAndWait(t, rt, instanceID)
	testruntime.RequireReconcileState(t, rt, instanceID, 2, 0, 0)

	// Check there's one row
	testruntime.RequireResolve(t, rt, instanceID, &testruntime.RequireResolveOptions{
		Resolver:   "sql",
		Properties: map[string]any{"sql": `SELECT COUNT(*) AS count FROM patch_model`},
		Result:     []map[string]any{{"count": 1}},
	})

	// Create a manual trigger
	testruntime.RefreshAndWait(t, rt, instanceID, &runtimev1.ResourceName{Kind: runtime.ResourceKindModel, Name: "patch_model"})

	// Check there's now two rows
	testruntime.RequireResolve(t, rt, instanceID, &testruntime.RequireResolveOptions{
		Resolver:   "sql",
		Properties: map[string]any{"sql": `SELECT COUNT(*) AS count FROM patch_model`},
		Result:     []map[string]any{{"count": 2}},
	})
}

func TestModelTests(t *testing.T) {
	rt, instanceID := testruntime.NewInstance(t)

	// Create the model with tests
	testruntime.PutFiles(t, rt, instanceID, map[string]string{
		"rill.yaml": ``,
		"models/test_model.yaml": `
type: model
sql: SELECT * FROM range(5)

tests:
  # Test that all values are in valid range
  - name: Valid Range
    assert: range >= 0 AND range <= 4

  # Test that all values are in valid range (sql)
  - name: Valid Range SQL
    sql: SELECT * FROM test_model WHERE range < 0 OR range > 4

  # Test row count is exactly 5
  - name: Exact Row Count
    sql: SELECT 'Wrong row count' as error WHERE (SELECT COUNT(*) FROM test_model) != 5

  # Test no null values exist
  - name: No Nulls
    assert: range IS NOT NULL

  # Test all values are non-negative
  - name: Non-negative Values
    assert: range >= 0

  # Test maximum value doesn't exceed 4
  - name: Max Value Check
    assert: range <= 4

  # Test using BETWEEN syntax
  - name: Range Between
    assert: range BETWEEN 0 AND 4

  # Test that specific values exist
  - name: Value 0 Exists
    sql: SELECT 'Value 0 missing' WHERE (SELECT COUNT(*) FROM test_model WHERE range = 0) = 0

  - name: Value 4 Exists
    sql: SELECT 'Value 4 missing' WHERE (SELECT COUNT(*) FROM test_model WHERE range = 4) = 0

  # Test no duplicates (each value appears exactly once)
  - name: No Duplicates
    sql: SELECT range, COUNT(*) as count FROM test_model GROUP BY range HAVING COUNT(*) > 1

  # Test arithmetic properties
  - name: Sum Check
    sql: SELECT 'Sum should be 10' WHERE (SELECT SUM(range) FROM test_model) != 10

  - name: Average Check
    sql: SELECT 'Average should be 2' WHERE (SELECT AVG(range) FROM test_model) != 2.0

  # Test min/max values
  - name: Min Value Check
    sql: SELECT 'Min should be 0' WHERE (SELECT MIN(range) FROM test_model) != 0

  - name: Max Value Check SQL
    sql: SELECT 'Max should be 4' WHERE (SELECT MAX(range) FROM test_model) != 4

  # Test using IN clause
  - name: Valid Values Only
    assert: range IN (0, 1, 2, 3, 4)

  # Test data completeness
  - name: All Expected Values Present
    sql: |
      SELECT missing_value FROM (
        VALUES (0), (1), (2), (3), (4)
      ) AS expected(missing_value)
      WHERE missing_value NOT IN (SELECT range FROM test_model)
`,
	})

	testruntime.ReconcileParserAndWait(t, rt, instanceID)

	model := testruntime.GetResource(t, rt, instanceID, runtime.ResourceKindModel, "test_model").GetModel()
	require.NotNil(t, model)

	require.NotEmpty(t, model.State.ResultTable)
	require.Empty(t, model.State.TestErrors, "All tests should pass")

	require.Len(t, model.Spec.Tests, 16, "Should have 16 tests defined")

	testNames := make(map[string]bool)
	for _, test := range model.Spec.Tests {
		testNames[test.Name] = true
	}

	expectedTests := []string{
		"Valid Range",
		"Valid Range SQL",
		"Exact Row Count",
		"No Nulls",
		"Non-negative Values",
		"Max Value Check",
		"Range Between",
		"Value 0 Exists",
		"Value 4 Exists",
		"No Duplicates",
		"Sum Check",
		"Average Check",
		"Min Value Check",
		"Max Value Check SQL",
		"Valid Values Only",
		"All Expected Values Present",
	}

	for _, testName := range expectedTests {
		require.True(t, testNames[testName], "Test %q should be defined", testName)
	}

	// Verify test hash is computed
	require.NotEmpty(t, model.State.TestHash, "Test hash should be computed")

	// Verify no test errors (all tests should pass)
	require.Empty(t, model.State.TestErrors, "No test errors expected for valid data")
}

func TestModelTestsWithFailures(t *testing.T) {
	rt, instanceID := testruntime.NewInstance(t)

	// Create a model that will fail some tests
	testruntime.PutFiles(t, rt, instanceID, map[string]string{
		"rill.yaml": ``,
		"models/failing_model.yaml": `
type: model
sql: SELECT range FROM range(10) WHERE range > 5  -- Only values 6,7,8,9

tests:
  # This will fail - expects 0-4 but we have 6-9
  - name: Valid Range
    assert: range >= 0 AND range <= 4

  # This will pass - all values are non-negative
  - name: Non-negative Values
    assert: range >= 0

  # This will fail - expects 5 rows but we have 4
  - name: Row Count Check
    sql: SELECT 'Wrong count' WHERE (SELECT COUNT(*) FROM failing_model) != 5
`,
	})

	// Reconcile should complete but with test errors
	testruntime.ReconcileParserAndWait(t, rt, instanceID)

	// Get the model resource
	model := testruntime.GetResource(t, rt, instanceID, runtime.ResourceKindModel, "failing_model").GetModel()
	require.NotNil(t, model)

	// Verify model executed successfully (data was produced)
	require.NotEmpty(t, model.State.ResultTable)

	// Verify tests are defined
	require.Len(t, model.Spec.Tests, 3)

	// Verify test errors are recorded
	require.NotEmpty(t, model.State.TestErrors, "Should have test errors")

	// Check specific test failures
	testErrorStr := model.State.TestErrors[0]
	require.Contains(t, testErrorStr, "test did not pass", "Should contain test failure message")
}

func TestModelTestAssertion(t *testing.T) {
	rt, instanceID := testruntime.NewInstance(t)

	// Test assertion SQL generation
	testruntime.PutFiles(t, rt, instanceID, map[string]string{
		"rill.yaml": ``,
		"models/assertion_model.yaml": `
type: model
sql: SELECT 1 as id, 'test' as name

tests:
  # Test assertion gets converted to "query for bad data"
  - name: ID Positive
    assert: id > 0
  
  # Equivalent SQL test
  - name: ID Positive SQL
    sql: SELECT * FROM assertion_model WHERE NOT (id > 0)
`,
	})

	testruntime.ReconcileParserAndWait(t, rt, instanceID)

	model := testruntime.GetResource(t, rt, instanceID, runtime.ResourceKindModel, "assertion_model").GetModel()
	require.NotNil(t, model)

	require.Empty(t, model.State.TestErrors, "Both assertion and SQL tests should pass")

	require.Len(t, model.Spec.Tests, 2)

	for _, test := range model.Spec.Tests {
		require.Equal(t, "sql", test.Resolver)
		require.NotNil(t, test.ResolverProperties)

		// Check that SQL is present in resolver properties
		props := test.ResolverProperties.AsMap()
		require.Contains(t, props, "sql")
		require.NotEmpty(t, props["sql"])
	}
}

func TestModelTriggeredPartitions(t *testing.T) {
	rt, instanceID := testruntime.NewInstance(t)

	// Create a partitioned model
	testruntime.PutFiles(t, rt, instanceID, map[string]string{
		"rill.yaml": ``,
		"models/partitioned_model.yaml": `
type: model
incremental: true
partitions:
  sql: SELECT 'partition1' AS id UNION ALL SELECT 'partition2' AS id
sql: SELECT '{{.partition.id}}' AS partition_id, now() AS created_at
`,
	})
	testruntime.ReconcileParserAndWait(t, rt, instanceID)
	testruntime.RequireReconcileState(t, rt, instanceID, 2, 0, 0)

	// Get the initial model state
	model := testruntime.GetResource(t, rt, instanceID, runtime.ResourceKindModel, "partitioned_model").GetModel()
	require.NotNil(t, model)
	require.NotEmpty(t, model.State.ResultTable)

	// Verify initially no triggered partitions
	require.Empty(t, model.State.TriggeredPartitions, "Should have no triggered partitions initially")

	// Verify model executed successfully
	require.NotEmpty(t, model.State.ResultTable, "Model should have executed and produced results")
}

func TestModelNoTriggeredPartitions(t *testing.T) {
	rt, instanceID := testruntime.NewInstance(t)

	// Create a regular incremental model
	testruntime.PutFiles(t, rt, instanceID, map[string]string{
		"rill.yaml": ``,
		"models/regular_model.yaml": `
type: model
incremental: true
sql: SELECT now() AS created_at
`,
	})
	testruntime.ReconcileParserAndWait(t, rt, instanceID)
	testruntime.RequireReconcileState(t, rt, instanceID, 2, 0, 0)

	// Get the model state
	model := testruntime.GetResource(t, rt, instanceID, runtime.ResourceKindModel, "regular_model").GetModel()
	require.NotNil(t, model)
	require.NotEmpty(t, model.State.ResultTable)

	// Verify no triggered partitions are set
	require.Empty(t, model.State.TriggeredPartitions, "Should have no triggered partitions initially")

	// Trigger a manual refresh (normal incremental refresh)
	testruntime.RefreshAndWait(t, rt, instanceID, &runtimev1.ResourceName{Kind: runtime.ResourceKindModel, Name: "regular_model"})

	// Get updated model state
	updatedModel := testruntime.GetResource(t, rt, instanceID, runtime.ResourceKindModel, "regular_model").GetModel()
	require.NotNil(t, updatedModel)

	// Verify still no triggered partitions (normal refresh doesn't set them)
	require.Empty(t, updatedModel.State.TriggeredPartitions, "Should still have no triggered partitions")

	// Verify model executed successfully
	require.NotEmpty(t, updatedModel.State.ResultTable, "Model should have executed normally")
}
