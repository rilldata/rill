package server

import (
	"context"
	"strings"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestPublishModel_SourceModel(t *testing.T) {
	t.Parallel()
	rt, instanceID := testruntime.NewInstance(t)
	server := mustNewServer(t, rt)

	// Create a source table in DuckDB so the SQL is valid
	testruntime.PutFiles(t, rt, instanceID, map[string]string{
		"sources/raw_events.sql": "SELECT 1 AS id, 'click' AS event_type",
	})
	testruntime.ReconcileParserAndWait(t, rt, instanceID)
	testruntime.RequireReconcileState(t, rt, instanceID, 1, 0, 0)

	resp, err := server.PublishModel(context.Background(), &runtimev1.PublishModelRequest{
		InstanceId: instanceID,
		Name:       "my_source_model",
		Sql:        "SELECT * FROM raw_events",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "my_source_model", resp.ModelName)
	require.NotEmpty(t, resp.FilePath)
	require.True(t, strings.HasSuffix(resp.FilePath, ".sql") || strings.HasSuffix(resp.FilePath, ".yaml"))
}

func TestPublishModel_DerivedModel(t *testing.T) {
	t.Parallel()
	rt, instanceID := testruntime.NewInstance(t)
	server := mustNewServer(t, rt)

	// Create an existing model first
	testruntime.PutFiles(t, rt, instanceID, map[string]string{
		"models/base_model.sql": "SELECT 1 AS id, 100 AS amount",
	})
	testruntime.ReconcileParserAndWait(t, rt, instanceID)
	testruntime.RequireReconcileState(t, rt, instanceID, 1, 0, 0)

	resp, err := server.PublishModel(context.Background(), &runtimev1.PublishModelRequest{
		InstanceId: instanceID,
		Name:       "derived_model",
		Sql:        "SELECT id, amount * 2 AS double_amount FROM base_model",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "derived_model", resp.ModelName)
}

func TestPublishModel_DuplicateNameRejected(t *testing.T) {
	t.Parallel()
	rt, instanceID := testruntime.NewInstance(t)
	server := mustNewServer(t, rt)

	// Create an existing model
	testruntime.PutFiles(t, rt, instanceID, map[string]string{
		"models/existing_model.sql": "SELECT 1 AS id",
	})
	testruntime.ReconcileParserAndWait(t, rt, instanceID)
	testruntime.RequireReconcileState(t, rt, instanceID, 1, 0, 0)

	// Attempt to publish with the same name
	_, err := server.PublishModel(context.Background(), &runtimev1.PublishModelRequest{
		InstanceId: instanceID,
		Name:       "existing_model",
		Sql:        "SELECT 2 AS id",
	})
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.AlreadyExists, st.Code())
	require.Contains(t, st.Message(), "already exists")
}

func TestPublishModel_InvalidNameRejected(t *testing.T) {
	t.Parallel()
	rt, instanceID := testruntime.NewInstance(t)
	server := mustNewServer(t, rt)

	tests := []struct {
		name      string
		modelName string
	}{
		{"empty name", ""},
		{"name with spaces", "my model"},
		{"name with special chars", "my-model!"},
		{"name starting with number", "123model"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := server.PublishModel(context.Background(), &runtimev1.PublishModelRequest{
				InstanceId: instanceID,
				Name:       tt.modelName,
				Sql:        "SELECT 1 AS id",
			})
			require.Error(t, err)
			st, ok := status.FromError(err)
			require.True(t, ok)
			require.Equal(t, codes.InvalidArgument, st.Code())
		})
	}
}

func TestPublishModel_EmptySQLRejected(t *testing.T) {
	t.Parallel()
	rt, instanceID := testruntime.NewInstance(t)
	server := mustNewServer(t, rt)

	_, err := server.PublishModel(context.Background(), &runtimev1.PublishModelRequest{
		InstanceId: instanceID,
		Name:       "some_model",
		Sql:        "",
	})
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.InvalidArgument, st.Code())
	require.Contains(t, st.Message(), "sql")
}

func TestPublishModel_YAMLGeneration(t *testing.T) {
	t.Parallel()
	rt, instanceID := testruntime.NewInstance(t)
	server := mustNewServer(t, rt)

	sql := "SELECT 1 AS id, 'hello' AS greeting"
	resp, err := server.PublishModel(context.Background(), &runtimev1.PublishModelRequest{
		InstanceId: instanceID,
		Name:       "yaml_test_model",
		Sql:        sql,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "yaml_test_model", resp.ModelName)

	// Verify the file was written by checking the file exists in the repo
	filePath := resp.FilePath
	require.NotEmpty(t, filePath)

	// Read back the file content to verify YAML correctness
	content, err := rt.GetFile(context.Background(), instanceID, filePath)
	require.NoError(t, err)
	require.NotEmpty(t, content)

	// Verify the file contains the SQL
	require.Contains(t, content, sql)
	// Verify it references the model name or is in the correct path
	require.True(t, strings.Contains(filePath, "yaml_test_model"))
}

func TestPublishModel_OverwriteNotAllowed(t *testing.T) {
	t.Parallel()
	rt, instanceID := testruntime.NewInstance(t)
	server := mustNewServer(t, rt)

	// First publish should succeed
	resp, err := server.PublishModel(context.Background(), &runtimev1.PublishModelRequest{
		InstanceId: instanceID,
		Name:       "overwrite_test",
		Sql:        "SELECT 1 AS id",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Wait for reconciliation to pick up the new file
	testruntime.ReconcileParserAndWait(t, rt, instanceID)

	// Second publish with same name should fail (overwrite_allowed_v1: false)
	_, err = server.PublishModel(context.Background(), &runtimev1.PublishModelRequest{
		InstanceId: instanceID,
		Name:       "overwrite_test",
		Sql:        "SELECT 2 AS new_id",
	})
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.AlreadyExists, st.Code())
}

func TestPublishModel_CaseInsensitiveDuplicateCheck(t *testing.T) {
	t.Parallel()
	rt, instanceID := testruntime.NewInstance(t)
	server := mustNewServer(t, rt)

	// Create a model with lowercase name
	testruntime.PutFiles(t, rt, instanceID, map[string]string{
		"models/mymodel.sql": "SELECT 1 AS id",
	})
	testruntime.ReconcileParserAndWait(t, rt, instanceID)
	testruntime.RequireReconcileState(t, rt, instanceID, 1, 0, 0)

	// Attempt to publish with different casing — should be rejected
	_, err := server.PublishModel(context.Background(), &runtimev1.PublishModelRequest{
		InstanceId: instanceID,
		Name:       "MyModel",
		Sql:        "SELECT 2 AS id",
	})
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.AlreadyExists, st.Code())
}

func TestPublishModel_ModelNameValidChars(t *testing.T) {
	t.Parallel()
	rt, instanceID := testruntime.NewInstance(t)
	server := mustNewServer(t, rt)

	tests := []struct {
		name      string
		modelName string
		wantErr   bool
	}{
		{"simple lowercase", "mymodel", false},
		{"with underscores", "my_model_v2", false},
		{"mixed case", "MyModel", false},
		{"single char", "m", false},
		{"with numbers", "model123", false},
		{"starts with underscore", "_model", false},
		{"hyphenated", "my-model", true},
		{"with dot", "my.model", true},
		{"with slash", "my/model", true},
		{"with space", "my model", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := server.PublishModel(context.Background(), &runtimev1.PublishModelRequest{
				InstanceId: instanceID,
				Name:       tt.modelName,
				Sql:        "SELECT 1 AS id",
			})
			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.InvalidArgument, st.Code())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestPublishModel_ClassificationSourceModel(t *testing.T) {
	t.Parallel()
	rt, instanceID := testruntime.NewInstance(t)
	server := mustNewServer(t, rt)

	// Create a source (external connector table simulation)
	testruntime.PutFiles(t, rt, instanceID, map[string]string{
		"sources/external_data.yaml": `type: source
connector: duckdb
sql: "CREATE TABLE external_data AS SELECT 1 AS id, 'data' AS val"
`,
	})
	testruntime.ReconcileParserAndWait(t, rt, instanceID)

	resp, err := server.PublishModel(context.Background(), &runtimev1.PublishModelRequest{
		InstanceId: instanceID,
		Name:       "published_from_source",
		Sql:        "SELECT * FROM external_data WHERE id > 0",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "published_from_source", resp.ModelName)
}

func TestPublishModel_ClassificationDerivedModel(t *testing.T) {
	t.Parallel()
	rt, instanceID := testruntime.NewInstance(t)
	server := mustNewServer(t, rt)

	// Create two models
	testruntime.PutFiles(t, rt, instanceID, map[string]string{
		"models/model_a.sql": "SELECT 1 AS id, 10 AS value",
		"models/model_b.sql": "SELECT 2 AS id, 20 AS value",
	})
	testruntime.ReconcileParserAndWait(t, rt, instanceID)
	testruntime.RequireReconcileState(t, rt, instanceID, 2, 0, 0)

	// Publish a model that references existing models
	resp, err := server.PublishModel(context.Background(), &runtimev1.PublishModelRequest{
		InstanceId: instanceID,
		Name:       "combined_model",
		Sql:        "SELECT * FROM model_a UNION ALL SELECT * FROM model_b",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "combined_model", resp.ModelName)
}

func TestPublishModel_InvalidInstanceID(t *testing.T) {
	t.Parallel()
	rt, _ := testruntime.NewInstance(t)
	server := mustNewServer(t, rt)

	_, err := server.PublishModel(context.Background(), &runtimev1.PublishModelRequest{
		InstanceId: "nonexistent-instance-id",
		Name:       "some_model",
		Sql:        "SELECT 1",
	})
	require.Error(t, err)
}

func TestPublishModel_ComplexSQL(t *testing.T) {
	t.Parallel()
	rt, instanceID := testruntime.NewInstance(t)
	server := mustNewServer(t, rt)

	// Publish a model with a CTE
	resp, err := server.PublishModel(context.Background(), &runtimev1.PublishModelRequest{
		InstanceId: instanceID,
		Name:       "cte_model",
		Sql:        "WITH base AS (SELECT 1 AS id, 'hello' AS msg) SELECT id, msg, length(msg) AS msg_len FROM base",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, "cte_model", resp.ModelName)
}

// mustNewServer creates a runtime server suitable for testing.
// It creates a Server with a no-op activity client.
func mustNewServer(t *testing.T, rt *runtime.Runtime) *Server {
	t.Helper()
	server, err := NewServer(context.Background(), &Options{
		Runtime:        rt,
		ActivityClient: activity.NewNoopClient(),
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		server.Close()
	})
	return server
}
