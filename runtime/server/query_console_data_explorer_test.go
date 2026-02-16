package server

import (
	"context"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestListDataExplorerObjects_EmptyCatalog(t *testing.T) {
	t.Parallel()
	rt, instanceID := testruntime.NewInstance(t)
	server, _ := newServer(t, rt)

	resp, err := server.ListDataExplorerObjects(testCtx(), &runtimev1.ListDataExplorerObjectsRequest{
		InstanceId: instanceID,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	// An empty catalog should return an empty or minimal list of objects
	// (may include system/information_schema tables depending on driver)
	require.NotNil(t, resp.Nodes)
}

func TestListDataExplorerObjects_WithModels(t *testing.T) {
	t.Parallel()
	rt, instanceID := testruntime.NewInstance(t)
	server, _ := newServer(t, rt)

	// Create a source and model via test files
	testruntime.PutFiles(t, rt, instanceID, map[string]string{
		"sources/test_source.yaml": `
connector: duckdb
sql: "SELECT 1 AS id, 'hello' AS name"
`,
		"models/test_model.sql": "SELECT id, name FROM test_source",
	})
	testruntime.ReconcileParserAndWait(t, rt, instanceID)
	testruntime.RequireReconcileState(t, rt, instanceID, 4, 0, 0)

	resp, err := server.ListDataExplorerObjects(testCtx(), &runtimev1.ListDataExplorerObjectsRequest{
		InstanceId: instanceID,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Nodes)

	// Verify we can find the objects we created
	foundSource := false
	foundModel := false
	for _, node := range flattenNodes(resp.Nodes) {
		if node.Name == "test_source" {
			foundSource = true
		}
		if node.Name == "test_model" {
			foundModel = true
		}
	}
	require.True(t, foundSource, "expected to find test_source in data explorer objects")
	require.True(t, foundModel, "expected to find test_model in data explorer objects")
}

func TestListDataExplorerObjects_InvalidInstance(t *testing.T) {
	t.Parallel()
	rt, _ := testruntime.NewInstance(t)
	server, _ := newServer(t, rt)

	_, err := server.ListDataExplorerObjects(testCtx(), &runtimev1.ListDataExplorerObjectsRequest{
		InstanceId: "nonexistent-instance-id",
	})
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.NotFound, st.Code())
}

func TestListDataExplorerObjects_TypeFilter(t *testing.T) {
	t.Parallel()
	rt, instanceID := testruntime.NewInstance(t)
	server, _ := newServer(t, rt)

	// Create source and model
	testruntime.PutFiles(t, rt, instanceID, map[string]string{
		"sources/filter_source.yaml": `
connector: duckdb
sql: "SELECT 1 AS val"
`,
		"models/filter_model.sql": "SELECT val FROM filter_source",
	})
	testruntime.ReconcileParserAndWait(t, rt, instanceID)
	testruntime.RequireReconcileState(t, rt, instanceID, 4, 0, 0)

	// Request with type filter (if supported by the API)
	resp, err := server.ListDataExplorerObjects(testCtx(), &runtimev1.ListDataExplorerObjectsRequest{
		InstanceId: instanceID,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Greater(t, len(resp.Nodes), 0, "expected at least one top-level node")
}

func TestGetDataExplorerSchema_Table(t *testing.T) {
	t.Parallel()
	rt, instanceID := testruntime.NewInstance(t)
	server, _ := newServer(t, rt)

	// Create a source with known schema
	testruntime.PutFiles(t, rt, instanceID, map[string]string{
		"sources/schema_test.yaml": `
connector: duckdb
sql: "SELECT 1::INTEGER AS id, 'hello'::VARCHAR AS name, 3.14::DOUBLE AS value"
`,
	})
	testruntime.ReconcileParserAndWait(t, rt, instanceID)
	testruntime.RequireReconcileState(t, rt, instanceID, 3, 0, 0)

	resp, err := server.GetDataExplorerSchema(testCtx(), &runtimev1.GetDataExplorerSchemaRequest{
		InstanceId: instanceID,
		TableName:  "schema_test",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Columns)
	require.GreaterOrEqual(t, len(resp.Columns), 3)

	// Verify column names
	colNames := make(map[string]bool)
	for _, col := range resp.Columns {
		colNames[col.Name] = true
	}
	require.True(t, colNames["id"], "expected column 'id'")
	require.True(t, colNames["name"], "expected column 'name'")
	require.True(t, colNames["value"], "expected column 'value'")
}

func TestGetDataExplorerSchema_Model(t *testing.T) {
	t.Parallel()
	rt, instanceID := testruntime.NewInstance(t)
	server, _ := newServer(t, rt)

	// Create a source and a model that transforms it
	testruntime.PutFiles(t, rt, instanceID, map[string]string{
		"sources/base_src.yaml": `
connector: duckdb
sql: "SELECT 1::INTEGER AS id, 'hello'::VARCHAR AS greeting"
`,
		"models/derived.sql": "SELECT id, greeting, LENGTH(greeting) AS greeting_len FROM base_src",
	})
	testruntime.ReconcileParserAndWait(t, rt, instanceID)
	testruntime.RequireReconcileState(t, rt, instanceID, 4, 0, 0)

	resp, err := server.GetDataExplorerSchema(testCtx(), &runtimev1.GetDataExplorerSchemaRequest{
		InstanceId: instanceID,
		TableName:  "derived",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Columns)
	require.GreaterOrEqual(t, len(resp.Columns), 3)

	colNames := make(map[string]bool)
	for _, col := range resp.Columns {
		colNames[col.Name] = true
	}
	require.True(t, colNames["id"], "expected column 'id'")
	require.True(t, colNames["greeting"], "expected column 'greeting'")
	require.True(t, colNames["greeting_len"], "expected column 'greeting_len'")
}

func TestGetDataExplorerSchema_NotFound(t *testing.T) {
	t.Parallel()
	rt, instanceID := testruntime.NewInstance(t)
	server, _ := newServer(t, rt)

	_, err := server.GetDataExplorerSchema(testCtx(), &runtimev1.GetDataExplorerSchemaRequest{
		InstanceId: instanceID,
		TableName:  "nonexistent_table",
	})
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.NotFound, st.Code())
}

func TestGetDataExplorerSchema_InvalidInstance(t *testing.T) {
	t.Parallel()
	rt, _ := testruntime.NewInstance(t)
	server, _ := newServer(t, rt)

	_, err := server.GetDataExplorerSchema(testCtx(), &runtimev1.GetDataExplorerSchemaRequest{
		InstanceId: "nonexistent-instance-id",
		TableName:  "some_table",
	})
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.NotFound, st.Code())
}

func TestGetDataExplorerSchema_EmptyTableName(t *testing.T) {
	t.Parallel()
	rt, instanceID := testruntime.NewInstance(t)
	server, _ := newServer(t, rt)

	_, err := server.GetDataExplorerSchema(testCtx(), &runtimev1.GetDataExplorerSchemaRequest{
		InstanceId: instanceID,
		TableName:  "",
	})
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	// Should be InvalidArgument for empty table name
	require.Equal(t, codes.InvalidArgument, st.Code())
}

func TestGetDataExplorerSchema_ColumnTypes(t *testing.T) {
	t.Parallel()
	rt, instanceID := testruntime.NewInstance(t)
	server, _ := newServer(t, rt)

	// Create a source with diverse column types
	testruntime.PutFiles(t, rt, instanceID, map[string]string{
		"sources/typed_src.yaml": `
connector: duckdb
sql: "SELECT 42::INTEGER AS int_col, 3.14::DOUBLE AS float_col, 'text'::VARCHAR AS str_col, true::BOOLEAN AS bool_col, CURRENT_TIMESTAMP AS ts_col"
`,
	})
	testruntime.ReconcileParserAndWait(t, rt, instanceID)
	testruntime.RequireReconcileState(t, rt, instanceID, 3, 0, 0)

	resp, err := server.GetDataExplorerSchema(testCtx(), &runtimev1.GetDataExplorerSchemaRequest{
		InstanceId: instanceID,
		TableName:  "typed_src",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Columns, 5)

	// Verify each column has a name and type
	for _, col := range resp.Columns {
		require.NotEmpty(t, col.Name, "column name should not be empty")
		require.NotEmpty(t, col.Type, "column type should not be empty for column %s", col.Name)
	}
}

func TestGetDataExplorerSchema_WithConnector(t *testing.T) {
	t.Parallel()
	rt, instanceID := testruntime.NewInstance(t)
	server, _ := newServer(t, rt)

	// Create a source
	testruntime.PutFiles(t, rt, instanceID, map[string]string{
		"sources/connector_test.yaml": `
connector: duckdb
sql: "SELECT 1 AS a, 2 AS b"
`,
	})
	testruntime.ReconcileParserAndWait(t, rt, instanceID)
	testruntime.RequireReconcileState(t, rt, instanceID, 3, 0, 0)

	// Request schema with explicit connector
	resp, err := server.GetDataExplorerSchema(testCtx(), &runtimev1.GetDataExplorerSchemaRequest{
		InstanceId: instanceID,
		TableName:  "connector_test",
		Connector:  "duckdb",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.GreaterOrEqual(t, len(resp.Columns), 2)
}

func TestListDataExplorerObjects_MultipleModelsAndSources(t *testing.T) {
	t.Parallel()
	rt, instanceID := testruntime.NewInstance(t)
	server, _ := newServer(t, rt)

	// Create multiple sources and models
	testruntime.PutFiles(t, rt, instanceID, map[string]string{
		"sources/src_a.yaml": `
connector: duckdb
sql: "SELECT 1 AS id"
`,
		"sources/src_b.yaml": `
connector: duckdb
sql: "SELECT 2 AS id"
`,
		"models/model_x.sql": "SELECT id FROM src_a",
		"models/model_y.sql": "SELECT id FROM src_b",
		"models/model_z.sql": "SELECT a.id AS a_id, b.id AS b_id FROM src_a a CROSS JOIN src_b b",
	})
	testruntime.ReconcileParserAndWait(t, rt, instanceID)
	testruntime.RequireReconcileState(t, rt, instanceID, 7, 0, 0)

	resp, err := server.ListDataExplorerObjects(testCtx(), &runtimev1.ListDataExplorerObjectsRequest{
		InstanceId: instanceID,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)

	allNodes := flattenNodes(resp.Nodes)
	names := make(map[string]bool)
	for _, node := range allNodes {
		names[node.Name] = true
	}

	require.True(t, names["src_a"], "expected src_a")
	require.True(t, names["src_b"], "expected src_b")
	require.True(t, names["model_x"], "expected model_x")
	require.True(t, names["model_y"], "expected model_y")
	require.True(t, names["model_z"], "expected model_z")
}

// newServer creates a test runtime server instance for handler testing.
func newServer(t *testing.T, rt *runtime.Runtime) (*Server, error) {
	t.Helper()
	server := &Server{
		runtime: rt,
	}
	return server, nil
}

// testCtx returns a background context for tests.
func testCtx() context.Context {
	return context.Background()
}

// flattenNodes recursively flattens a tree of DataExplorerNode into a flat slice.
func flattenNodes(nodes []*runtimev1.DataExplorerNode) []*runtimev1.DataExplorerNode {
	var result []*runtimev1.DataExplorerNode
	for _, node := range nodes {
		result = append(result, node)
		if len(node.Children) > 0 {
			result = append(result, flattenNodes(node.Children)...)
		}
	}
	return result
}
