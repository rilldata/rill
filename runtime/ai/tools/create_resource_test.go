package tools

import (
	"context"
	"testing"

	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func TestCreateAndReconcileResource(t *testing.T) {
	t.Skip("Skipping since file watcher is not available in test runtime")
	ctx := context.Background()

	// Create a minimal runtime using testruntime
	rt, instanceID := testruntime.NewInstance(t)

	// Create the tool
	tool := CreateAndReconcileResource(instanceID, rt)
	if tool == nil {
		t.Fatal("Expected tool to be non-nil")
	}

	// Test parameters for creating a simple model
	params := map[string]any{
		"resource_name": "test_model",
		"resource_type": "model",
		"contents": `SELECT 
    1 as id,
    'test' as name,
    100.0 as amount
LIMIT 10`,
		"path": "models/test_model.sql",
	}

	// Execute the tool
	result, err := tool.Execute(ctx, params)
	if err != nil {
		t.Fatalf("CreateAndReconcileResource failed: %v", err)
	}

	// Check that we got a successful result
	resultMap, ok := result.(map[string]any)
	if !ok {
		t.Fatal("Expected result to be a map")
	}

	// Check for success message
	if resultValue, exists := resultMap["result"]; exists {
		t.Logf("Success: %v", resultValue)
	} else if errorValue, exists := resultMap["error"]; exists {
		t.Fatalf("Tool returned error: %v", errorValue)
	} else {
		t.Fatal("Unexpected result format")
	}

	olap, release, err := rt.OLAP(ctx, instanceID, "")
	if err != nil {
		require.NoError(t, err, "Failed to get OLAP connection")
	}
	defer release()

	// Verify the model was created
	_, err = olap.InformationSchema().Lookup(ctx, "", "", "test_model")
	require.NoError(t, err, "Expected model 'test_model' to be created")
}
