package server_test

import (
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestCreateResource(t *testing.T) {
	server, instanceID := getTestServer(t)
	ctx := testCtx()

	// Create a simple model resource
	modelSpec, err := structpb.NewStruct(map[string]interface{}{
		"sql": "SELECT 1 as value",
	})
	require.NoError(t, err)

	resource := &runtimev1.Resource{
		Meta: &runtimev1.ResourceMeta{
			Name: &runtimev1.ResourceName{
				Kind: runtime.ResourceKindModel,
				Name: "test_model",
			},
		},
		Resource: &runtimev1.Resource_Model{
			Model: &runtimev1.Model{
				Spec: &runtimev1.ModelSpec{
					OutputConnector:  "duckdb",
					OutputProperties: modelSpec,
				},
				State: &runtimev1.ModelState{},
			},
		},
	}

	resp, err := server.CreateResource(ctx, &runtimev1.CreateResourceRequest{
		InstanceId: instanceID,
		Resource:   resource,
	})
	require.NoError(t, err)
	require.NotNil(t, resp.Resource)
	require.Equal(t, runtime.ResourceKindModel, resp.Resource.Meta.Name.Kind)
	require.Equal(t, "test_model", resp.Resource.Meta.Name.Name)
}

func TestGetResource(t *testing.T) {
	server, instanceID := getTestServer(t)
	ctx := testCtx()

	// Create a resource
	modelSpec, err := structpb.NewStruct(map[string]interface{}{
		"sql": "SELECT 1 as value",
	})
	require.NoError(t, err)

	resource := &runtimev1.Resource{
		Meta: &runtimev1.ResourceMeta{
			Name: &runtimev1.ResourceName{
				Kind: runtime.ResourceKindModel,
				Name: "test_model",
			},
		},
		Resource: &runtimev1.Resource_Model{
			Model: &runtimev1.Model{
				Spec: &runtimev1.ModelSpec{
					OutputConnector:  "duckdb",
					OutputProperties: modelSpec,
				},
				State: &runtimev1.ModelState{},
			},
		},
	}

	_, err = server.CreateResource(ctx, &runtimev1.CreateResourceRequest{
		InstanceId: instanceID,
		Resource:   resource,
	})
	require.NoError(t, err)

	// Now get it
	resp, err := server.GetResource(ctx, &runtimev1.GetResourceRequest{
		InstanceId: instanceID,
		Name: &runtimev1.ResourceName{
			Kind: runtime.ResourceKindModel,
			Name: "test_model",
		},
	})
	require.NoError(t, err)
	require.NotNil(t, resp.Resource)
	require.Equal(t, runtime.ResourceKindModel, resp.Resource.Meta.Name.Kind)
	require.Equal(t, "test_model", resp.Resource.Meta.Name.Name)
}

func TestUpdateResource(t *testing.T) {
	server, instanceID := getTestServer(t)
	ctx := testCtx()

	// First create a resource
	modelSpec, err := structpb.NewStruct(map[string]interface{}{
		"sql": "SELECT 1 as value",
	})
	require.NoError(t, err)

	resource := &runtimev1.Resource{
		Meta: &runtimev1.ResourceMeta{
			Name: &runtimev1.ResourceName{
				Kind: runtime.ResourceKindModel,
				Name: "test_model",
			},
		},
		Resource: &runtimev1.Resource_Model{
			Model: &runtimev1.Model{
				Spec: &runtimev1.ModelSpec{
					OutputConnector:  "duckdb",
					OutputProperties: modelSpec,
				},
				State: &runtimev1.ModelState{},
			},
		},
	}

	_, err = server.CreateResource(ctx, &runtimev1.CreateResourceRequest{
		InstanceId: instanceID,
		Resource:   resource,
	})
	require.NoError(t, err)

	// Update the resource
	updatedModelSpec, err := structpb.NewStruct(map[string]interface{}{
		"sql": "SELECT 2 as value",
	})
	require.NoError(t, err)

	updatedResource := &runtimev1.Resource{
		Meta: &runtimev1.ResourceMeta{
			Name: &runtimev1.ResourceName{
				Kind: runtime.ResourceKindModel,
				Name: "test_model",
			},
		},
		Resource: &runtimev1.Resource_Model{
			Model: &runtimev1.Model{
				Spec: &runtimev1.ModelSpec{
					OutputConnector:  "duckdb",
					OutputProperties: updatedModelSpec,
				},
				State: &runtimev1.ModelState{},
			},
		},
	}

	resp, err := server.UpdateResource(ctx, &runtimev1.UpdateResourceRequest{
		InstanceId: instanceID,
		Kind:       runtime.ResourceKindModel,
		Name:       "test_model",
		Resource:   updatedResource,
	})
	require.NoError(t, err)
	require.NotNil(t, resp.Resource)
}

func TestDeleteResource(t *testing.T) {
	server, instanceID := getTestServer(t)
	ctx := testCtx()

	// First create a resource
	modelSpec, err := structpb.NewStruct(map[string]interface{}{
		"sql": "SELECT 1 as value",
	})
	require.NoError(t, err)

	resource := &runtimev1.Resource{
		Meta: &runtimev1.ResourceMeta{
			Name: &runtimev1.ResourceName{
				Kind: runtime.ResourceKindModel,
				Name: "test_model",
			},
		},
		Resource: &runtimev1.Resource_Model{
			Model: &runtimev1.Model{
				Spec: &runtimev1.ModelSpec{
					OutputConnector:  "duckdb",
					OutputProperties: modelSpec,
				},
				State: &runtimev1.ModelState{},
			},
		},
	}

	_, err = server.CreateResource(ctx, &runtimev1.CreateResourceRequest{
		InstanceId: instanceID,
		Resource:   resource,
	})
	require.NoError(t, err)

	// Delete the resource
	_, err = server.DeleteResource(ctx, &runtimev1.DeleteResourceRequest{
		InstanceId: instanceID,
		Kind:       runtime.ResourceKindModel,
		Name:       "test_model",
	})
	require.NoError(t, err)

	// Verify it's gone
	_, err = server.GetResource(ctx, &runtimev1.GetResourceRequest{
		InstanceId: instanceID,
		Name: &runtimev1.ResourceName{
			Kind: runtime.ResourceKindModel,
			Name: "test_model",
		},
	})
	require.Error(t, err) // Should return not found
}
