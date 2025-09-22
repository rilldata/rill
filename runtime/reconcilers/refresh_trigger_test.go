package reconcilers_test

import (
	"context"
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"

	_ "github.com/rilldata/rill/runtime/resolvers"
)

func TestPartitionSpecificRefresh(t *testing.T) {
	rt, instanceID := testruntime.NewInstance(t)

	// Create a partitioned model
	testruntime.PutFiles(t, rt, instanceID, map[string]string{
		"rill.yaml": ``,
		"models/trigger_test_model.yaml": `
type: model
incremental: true
partitions:
  sql: SELECT 'partition1' AS id UNION ALL SELECT 'partition2' AS id
sql: SELECT '{{.partition.id}}' AS partition_id, now() AS created_at
`,
	})
	testruntime.ReconcileParserAndWait(t, rt, instanceID)
	testruntime.RequireReconcileState(t, rt, instanceID, 2, 0, 0)

	// Get controller for creating refresh trigger
	ctx := context.Background()
	ctrl, err := rt.Controller(ctx, instanceID)
	require.NoError(t, err)

	// Create a refresh trigger targeting specific partition
	trgName := &runtimev1.ResourceName{Kind: runtime.ResourceKindRefreshTrigger, Name: time.Now().String()}
	err = ctrl.Create(ctx, trgName, nil, nil, nil, false, &runtimev1.Resource{
		Resource: &runtimev1.Resource_RefreshTrigger{
			RefreshTrigger: &runtimev1.RefreshTrigger{
				Spec: &runtimev1.RefreshTriggerSpec{
					Models: []*runtimev1.RefreshModelTrigger{
						{
							Model:      "trigger_test_model",
							Partitions: []string{"partition1"},
						},
					},
				},
			},
		},
	})
	require.NoError(t, err)

	// Wait for refresh trigger to be processed
	err = ctrl.WaitUntilIdle(ctx, false)
	require.NoError(t, err)

	// Get the model after refresh trigger
	model := testruntime.GetResource(t, rt, instanceID, runtime.ResourceKindModel, "trigger_test_model").GetModel()
	require.NotNil(t, model)

	// Verify the model executed successfully
	require.NotEmpty(t, model.State.ResultTable, "Model should have executed and produced results")

	// Note: The triggered partitions field should be cleared after execution
	// so we can't directly verify it was set, but we can verify the model ran
}

func TestRefreshMultiplePartitions(t *testing.T) {
	rt, instanceID := testruntime.NewInstance(t)

	// Create a partitioned model
	testruntime.PutFiles(t, rt, instanceID, map[string]string{
		"rill.yaml": ``,
		"models/multi_trigger_model.yaml": `
type: model
incremental: true
partitions:
  sql: SELECT 'p1' AS id UNION ALL SELECT 'p2' AS id UNION ALL SELECT 'p3' AS id
sql: SELECT '{{.partition.id}}' AS partition_id, now() AS created_at
`,
	})
	testruntime.ReconcileParserAndWait(t, rt, instanceID)
	testruntime.RequireReconcileState(t, rt, instanceID, 2, 0, 0)

	// Get controller
	ctx := context.Background()
	ctrl, err := rt.Controller(ctx, instanceID)
	require.NoError(t, err)

	// Create refresh trigger with multiple partitions
	trgName := &runtimev1.ResourceName{Kind: runtime.ResourceKindRefreshTrigger, Name: time.Now().String()}
	err = ctrl.Create(ctx, trgName, nil, nil, nil, false, &runtimev1.Resource{
		Resource: &runtimev1.Resource_RefreshTrigger{
			RefreshTrigger: &runtimev1.RefreshTrigger{
				Spec: &runtimev1.RefreshTriggerSpec{
					Models: []*runtimev1.RefreshModelTrigger{
						{
							Model:      "multi_trigger_model",
							Partitions: []string{"p1", "p3"},
						},
					},
				},
			},
		},
	})
	require.NoError(t, err)

	// Wait for trigger to be processed
	err = ctrl.WaitUntilIdle(ctx, false)
	require.NoError(t, err)

	// Verify model executed successfully
	model := testruntime.GetResource(t, rt, instanceID, runtime.ResourceKindModel, "multi_trigger_model").GetModel()
	require.NotNil(t, model)
	require.NotEmpty(t, model.State.ResultTable, "Model should have executed successfully")
}

func TestRefreshAllModels(t *testing.T) {
	rt, instanceID := testruntime.NewInstance(t)

	// Create multiple models
	testruntime.PutFiles(t, rt, instanceID, map[string]string{
		"rill.yaml": ``,
		"models/model1.yaml": `
type: model
incremental: true
sql: SELECT 'model1' AS name, now() AS created_at
`,
		"models/model2.yaml": `
type: model
incremental: true
sql: SELECT 'model2' AS name, now() AS created_at
`,
	})
	testruntime.ReconcileParserAndWait(t, rt, instanceID)
	testruntime.RequireReconcileState(t, rt, instanceID, 3, 0, 0)

	// Get controller
	ctx := context.Background()
	ctrl, err := rt.Controller(ctx, instanceID)
	require.NoError(t, err)

	// Create refresh trigger for all models (no specific partitions)
	trgName := &runtimev1.ResourceName{Kind: runtime.ResourceKindRefreshTrigger, Name: time.Now().String()}
	err = ctrl.Create(ctx, trgName, nil, nil, nil, false, &runtimev1.Resource{
		Resource: &runtimev1.Resource_RefreshTrigger{
			RefreshTrigger: &runtimev1.RefreshTrigger{
				Spec: &runtimev1.RefreshTriggerSpec{
					Models: []*runtimev1.RefreshModelTrigger{
						{Model: "model1"},
						{Model: "model2"},
					},
				},
			},
		},
	})
	require.NoError(t, err)

	// Wait for trigger to be processed
	err = ctrl.WaitUntilIdle(ctx, false)
	require.NoError(t, err)

	// Verify both models exist and executed successfully
	model1 := testruntime.GetResource(t, rt, instanceID, runtime.ResourceKindModel, "model1").GetModel()
	model2 := testruntime.GetResource(t, rt, instanceID, runtime.ResourceKindModel, "model2").GetModel()

	require.NotNil(t, model1)
	require.NotNil(t, model2)
	require.NotEmpty(t, model1.State.ResultTable, "Model1 should have executed successfully")
	require.NotEmpty(t, model2.State.ResultTable, "Model2 should have executed successfully")

	// When no partitions specified, should not set triggered partitions (normal refresh)
	require.Empty(t, model1.State.TriggeredPartitions, "Model1 should have no triggered partitions for normal refresh")
	require.Empty(t, model2.State.TriggeredPartitions, "Model2 should have no triggered partitions for normal refresh")
}
