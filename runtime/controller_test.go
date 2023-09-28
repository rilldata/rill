package runtime_test

import (
	"context"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestComplete(t *testing.T) {
	// Add source, model, and dashboard
	rt, id := testruntime.NewInstance(t)
	testruntime.PutFiles(t, rt, id, map[string]string{
		"/data/foo.csv": `a,b,c,d,e
1,2,3,4,5
1,2,3,4,5
1,2,3,4,5
`,
		"/sources/foo.yaml": `
type: local_file
path: data/foo.csv
`,
		"/models/bar.sql": `
SELECT * FROM foo
`,
		"/dashboards/foobar.yaml": `
model: bar
dimensions:
- name: a
  column: a
measures:
- name: b
  expression: count(*)
`,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 0, 0)

	// Verify the source
	testruntime.RequireResource(t, rt, id, &runtimev1.Resource{
		Meta: &runtimev1.ResourceMeta{
			Name:      &runtimev1.ResourceName{Kind: runtime.ResourceKindSource, Name: "foo"},
			Owner:     runtime.GlobalProjectParserName,
			FilePaths: []string{"/sources/foo.yaml"},
		},
		Resource: &runtimev1.Resource_Source{
			Source: &runtimev1.SourceV2{
				Spec: &runtimev1.SourceSpec{
					SourceConnector: "local_file",
					SinkConnector:   "duckdb",
					Properties:      must(structpb.NewStruct(map[string]any{"path": "data/foo.csv"})),
				},
				State: &runtimev1.SourceState{
					Connector: "duckdb",
					Table:     "foo",
				},
			},
		},
	})
	testruntime.RequireOLAPTable(t, rt, id, "foo")
	testruntime.RequireOLAPTableCount(t, rt, id, "foo", 3)

	// Verify the model
	falsy := false
	testruntime.RequireResource(t, rt, id, &runtimev1.Resource{
		Meta: &runtimev1.ResourceMeta{
			Name:      &runtimev1.ResourceName{Kind: runtime.ResourceKindModel, Name: "bar"},
			Refs:      []*runtimev1.ResourceName{{Kind: runtime.ResourceKindSource, Name: "foo"}},
			Owner:     runtime.GlobalProjectParserName,
			FilePaths: []string{"/models/bar.sql"},
		},
		Resource: &runtimev1.Resource_Model{
			Model: &runtimev1.ModelV2{
				Spec: &runtimev1.ModelSpec{
					Connector:   "duckdb",
					Sql:         "SELECT * FROM foo",
					Materialize: &falsy,
				},
				State: &runtimev1.ModelState{
					Connector: "duckdb",
					Table:     "bar",
				},
			},
		},
	})
	testruntime.RequireOLAPTable(t, rt, id, "bar")
	testruntime.RequireOLAPTableCount(t, rt, id, "bar", 3)

	// Verify the metrics view
	mvSpec := &runtimev1.MetricsViewSpec{
		Connector:  "duckdb",
		Table:      "bar",
		Dimensions: []*runtimev1.MetricsViewSpec_DimensionV2{{Name: "a", Column: "a"}},
		Measures:   []*runtimev1.MetricsViewSpec_MeasureV2{{Name: "b", Expression: "count(*)"}},
	}
	testruntime.RequireResource(t, rt, id, &runtimev1.Resource{
		Meta: &runtimev1.ResourceMeta{
			Name:      &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: "foobar"},
			Refs:      []*runtimev1.ResourceName{{Kind: runtime.ResourceKindModel, Name: "bar"}},
			Owner:     runtime.GlobalProjectParserName,
			FilePaths: []string{"/dashboards/foobar.yaml"},
		},
		Resource: &runtimev1.Resource_MetricsView{
			MetricsView: &runtimev1.MetricsViewV2{
				Spec: mvSpec,
				State: &runtimev1.MetricsViewState{
					ValidSpec: mvSpec,
				},
			},
		},
	})

	// Add syntax error in source, check downstream resources error
	testruntime.PutFiles(t, rt, id, map[string]string{
		"/sources/foo.yaml": `
type: local_file
path
`,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 3, 3, 1)

	// Verify the model (errored and state cleared)
	testruntime.RequireResource(t, rt, id, &runtimev1.Resource{
		Meta: &runtimev1.ResourceMeta{
			Name:           &runtimev1.ResourceName{Kind: runtime.ResourceKindModel, Name: "bar"},
			Owner:          runtime.GlobalProjectParserName,
			FilePaths:      []string{"/models/bar.sql"},
			ReconcileError: "Table with name foo does not exist",
		},
		Resource: &runtimev1.Resource_Model{
			Model: &runtimev1.ModelV2{
				Spec: &runtimev1.ModelSpec{
					Connector:   "duckdb",
					Sql:         "SELECT * FROM foo",
					Materialize: &falsy,
				},
				State: &runtimev1.ModelState{},
			},
		},
	})
	testruntime.RequireNoOLAPTable(t, rt, id, "bar")

	// Verify the metrics view (errored and state cleared)
	testruntime.RequireResource(t, rt, id, &runtimev1.Resource{
		Meta: &runtimev1.ResourceMeta{
			Name:           &runtimev1.ResourceName{Kind: runtime.ResourceKindMetricsView, Name: "foobar"},
			Refs:           []*runtimev1.ResourceName{{Kind: runtime.ResourceKindModel, Name: "bar"}},
			Owner:          runtime.GlobalProjectParserName,
			FilePaths:      []string{"/dashboards/foobar.yaml"},
			ReconcileError: "does not exist",
		},
		Resource: &runtimev1.Resource_MetricsView{
			MetricsView: &runtimev1.MetricsViewV2{
				Spec:  mvSpec,
				State: &runtimev1.MetricsViewState{},
			},
		},
	})

	// Fix source, check downstream resources succeed
	testruntime.PutFiles(t, rt, id, map[string]string{
		"/sources/foo.yaml": `
type: local_file
path: data/foo.csv
`,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 0, 0)
}

func TestSource(t *testing.T) {
	// Add source
	rt, id := testruntime.NewInstance(t)
	testruntime.PutFiles(t, rt, id, map[string]string{
		"/data/foo.csv": `a,b,c,d,e
1,2,3,4,5
1,2,3,4,5
1,2,3,4,5
`,
		"/sources/foo.yaml": `
type: local_file
path: data/foo.csv
`,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 2, 0, 0)
	testruntime.RequireResource(t, rt, id, &runtimev1.Resource{
		Meta: &runtimev1.ResourceMeta{
			Name:      &runtimev1.ResourceName{Kind: runtime.ResourceKindSource, Name: "foo"},
			Owner:     runtime.GlobalProjectParserName,
			FilePaths: []string{"/sources/foo.yaml"},
		},
		Resource: &runtimev1.Resource_Source{
			Source: &runtimev1.SourceV2{
				Spec: &runtimev1.SourceSpec{
					SourceConnector: "local_file",
					SinkConnector:   "duckdb",
					Properties:      must(structpb.NewStruct(map[string]any{"path": "data/foo.csv"})),
				},
				State: &runtimev1.SourceState{
					Connector: "duckdb",
					Table:     "foo",
				},
			},
		},
	})
	testruntime.RequireOLAPTable(t, rt, id, "foo")
	testruntime.RequireOLAPTableCount(t, rt, id, "foo", 3)

	// Update underlying data and refresh, verify table is updated
	testruntime.PutFiles(t, rt, id, map[string]string{
		"/data/foo.csv": `a,b,c,d,e
1,2,3,4,5
`,
	})
	testruntime.RefreshAndWait(t, rt, id, &runtimev1.ResourceName{Kind: runtime.ResourceKindSource, Name: "foo"})
	testruntime.RequireReconcileState(t, rt, id, 2, 0, 0)
	testruntime.RequireOLAPTable(t, rt, id, "foo")
	testruntime.RequireOLAPTableCount(t, rt, id, "foo", 1)

	// Delete the underlying table
	olap, release, err := rt.OLAP(context.Background(), id)
	require.NoError(t, err)
	err = olap.Exec(context.Background(), &drivers.Statement{Query: "DROP TABLE foo;"})
	require.NoError(t, err)
	release()
	testruntime.RequireNoOLAPTable(t, rt, id, "foo")

	// Reconcile the source and verify the table is added back
	testruntime.ReconcileAndWait(t, rt, id, &runtimev1.ResourceName{Kind: runtime.ResourceKindSource, Name: "foo"})
	testruntime.RequireReconcileState(t, rt, id, 2, 0, 0)
	testruntime.RequireOLAPTable(t, rt, id, "foo")
	testruntime.RequireOLAPTableCount(t, rt, id, "foo", 1)

	// Change the source so it errors
	testruntime.PutFiles(t, rt, id, map[string]string{
		"/sources/foo.yaml": `
type: local_file
path: data/bar.csv
`,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 2, 1, 0)
	testruntime.RequireNoOLAPTable(t, rt, id, "foo")

	// Restore the source, verify it works again
	testruntime.PutFiles(t, rt, id, map[string]string{
		"/sources/foo.yaml": `
type: local_file
path: data/foo.csv
`,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 2, 0, 0)
	testruntime.RequireResource(t, rt, id, &runtimev1.Resource{
		Meta: &runtimev1.ResourceMeta{
			Name:      &runtimev1.ResourceName{Kind: runtime.ResourceKindSource, Name: "foo"},
			Owner:     runtime.GlobalProjectParserName,
			FilePaths: []string{"/sources/foo.yaml"},
		},
		Resource: &runtimev1.Resource_Source{
			Source: &runtimev1.SourceV2{
				Spec: &runtimev1.SourceSpec{
					SourceConnector: "local_file",
					SinkConnector:   "duckdb",
					Properties:      must(structpb.NewStruct(map[string]any{"path": "data/foo.csv"})),
				},
				State: &runtimev1.SourceState{
					Connector: "duckdb",
					Table:     "foo",
				},
			},
		},
	})
	testruntime.RequireOLAPTable(t, rt, id, "foo")
	testruntime.RequireOLAPTableCount(t, rt, id, "foo", 1)
}

func TestSimultaneousDeleteRenameCreate(t *testing.T) {
	// Add bar and foo
	rt, id := testruntime.NewInstance(t)
	testruntime.PutFiles(t, rt, id, map[string]string{
		"/models/bar.sql": `SELECT 10`,
		"/models/foo.sql": `SELECT 20`,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 3, 0, 0)

	// Delete bar, rename foo to foo_two, add bazz
	testruntime.DeleteFiles(t, rt, id,
		"/models/bar.sql",
		"/models/foo.sql",
	)
	testruntime.PutFiles(t, rt, id, map[string]string{
		"/models/foo_two.sql": `SELECT 20`,
		"/models/bazz.sql":    `SELECT 30`,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 3, 0, 0)

	testruntime.RequireNoOLAPTable(t, rt, id, "bar")
	testruntime.RequireOLAPTable(t, rt, id, "foo_two")
	testruntime.RequireOLAPTable(t, rt, id, "bazz")
}

func TestSourceRefreshSchedule(t *testing.T) {
	// Add source with refresh schedule
	// Verify it gets retriggered after the delay
}

func TestSourceAndModelNameCollission(t *testing.T) {

}

func TestModelMaterialize(t *testing.T) {
	// Create model
	// Make materialized, verify is table
	// Make not materialized, verify is view
}

func TestModelCTE(t *testing.T) {
	// Create a model that references a source
	// Add CTE with same name as source, verify no ref to source anymore
}

func TestRename(t *testing.T) {
	// Create source and model
	// Rename the model, verify success
	// Rename the model to different case, verify success
	// Add model referencing new name, Rename the source to new name, verify old model breaks and new one works
	// Rename model A to B and model B to A, verify success
	// Rename model A to B and source B to A, verify success
}

func TestInterdependence(t *testing.T) {
	// Test D -> C, D -> A, C -> A,B (-> = refs)
	// Test error propagation on source error
}

func TestCycles(t *testing.T) {
	// Test A -> B, B -> A
	// Break cycle by deleting, verify changed errors

	// Test A -> B, B -> C, C -> A
	// Break cycle by changing to source, verify success
}

func TestMetricsView(t *testing.T) {
	// Create model and metrics view, verify success
	// Break model, verify metrics view not valid
}

func TestStageChanges(t *testing.T) {
	// Create source, model, metrics view
	// Break source, verify model and metrics view have errors, but are valid
}

func TestWatch(t *testing.T) {
	// Create instance with watch
	// Create source, wait and verify
	// Create model, wait and verify
	// Create metrics view, wait and verify
	// Drop source, wait and verify
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
