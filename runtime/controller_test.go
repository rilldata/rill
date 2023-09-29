package runtime_test

import (
	"context"
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/queries"
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

func TestCacheInvalidation(t *testing.T) {
	// Add source and model
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
		"/models/bar.sql": `SELECT * FROM foo`,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 3, 0, 0)
	testruntime.RequireTableRowCount(t, rt, id, "bar", 3)

	testruntime.PutFiles(t, rt, id, map[string]string{
		"/models/bar.sql": `SELECT * FROM foo LIMIT`,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 2, 1, 1)

	time.Sleep(time.Second) // this is needed since we add second to the cache key
	testruntime.PutFiles(t, rt, id, map[string]string{
		"/models/bar.sql": `SELECT * FROM foo LIMIT 1`,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 3, 0, 0)
	testruntime.RequireTableRowCount(t, rt, id, "bar", 1) // limit in model should override the limit in the query
}

func TestSourceRefreshSchedule(t *testing.T) {
	// Add source refresh schedule
	rt, id := testruntime.NewInstance(t)
	testruntime.PutFiles(t, rt, id, map[string]string{
		"/data/foo.csv": `a,b,c,d,e
1,2,3,4,5
1,2,3,4,5
1,2,3,4,5`,
		"/sources/foo.yaml": `
type: local_file
path: data/foo.csv
refresh:
  every: 1
`,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 2, 0, 0)
	testruntime.RequireTableRowCount(t, rt, id, "foo", 3)

	// update the data file with only 2 rows
	testruntime.PutFiles(t, rt, id, map[string]string{
		"/data/foo.csv": `a,b,c,d,e
1,2,3,4,5
1,2,3,4,5`,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	// no change in data just yet
	testruntime.RequireTableRowCount(t, rt, id, "foo", 3)

	// wait to make sure the data is ingested
	time.Sleep(2 * time.Second) // TODO: is there a way to decrease this wait time?
	testruntime.ReconcileParserAndWait(t, rt, id)
	// data has changed
	testruntime.RequireTableRowCount(t, rt, id, "foo", 2)
}

func TestSourceAndModelNameCollission(t *testing.T) {
	// Add source for a file
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
	testruntime.RequireOLAPTable(t, rt, id, "foo")

	// Create a source with same name within a different folder
	testruntime.PutFiles(t, rt, id, map[string]string{
		"/other_folder/foo.yaml": `
kind: source
type: local_file
path: data/foo.csv
`,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 2, 1, 1)
	testruntime.RequireOLAPTable(t, rt, id, "foo")

	// Deleting the other file marks the other as valid
	testruntime.DeleteFiles(t, rt, id, "/other_folder/foo.yaml")
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 2, 0, 0)
	testruntime.RequireOLAPTable(t, rt, id, "foo")

	// Create a source with same name using `name` annotation
	testruntime.PutFiles(t, rt, id, map[string]string{
		"/sources/foo_1.yaml": `
name: foo
type: local_file
path: data/foo.csv
`,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 2, 1, 1)
	testruntime.RequireOLAPTable(t, rt, id, "foo")

	// Deleting the other file marks the other as valid
	testruntime.DeleteFiles(t, rt, id, "/sources/foo_1.yaml")
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 2, 0, 0)
	testruntime.RequireOLAPTable(t, rt, id, "foo")

	// Create a model with same name as the source
	testruntime.PutFiles(t, rt, id, map[string]string{
		"/models/foo.sql": `SELECT 1`,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 2, 1, 1)
	// Data is from the source and model did not override it
	testruntime.RequireTableRowCount(t, rt, id, "foo", 3)

	// TODO: any other cases?
}

func TestModelMaterialize(t *testing.T) {
	// Add a simple model
	rt, id := testruntime.NewInstance(t)
	testruntime.PutFiles(t, rt, id, map[string]string{
		"/models/bar.sql": `
select 1
`,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 2, 0, 0)

	olap, done, err := rt.OLAP(context.Background(), id)
	require.NoError(t, err)
	defer done()

	// Assert that the model is a view
	testruntime.RequireIsView(t, olap, "bar", true)

	// Mark the model as materialized
	testruntime.PutFiles(t, rt, id, map[string]string{
		"/models/bar.sql": `
-- @materialize: true
select 1
`,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 2, 0, 0)
	// Assert that the model is a table now
	testruntime.RequireIsView(t, olap, "bar", false)

	// Mark the model as not materialized
	testruntime.PutFiles(t, rt, id, map[string]string{
		"/models/bar.sql": `
-- @materialize: false
select 1
`,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 2, 0, 0)
	// Assert that the model is back to being a view
	testruntime.RequireIsView(t, olap, "bar", true)
}

func TestModelCTE(t *testing.T) {
	// Create a model that references a source
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
		"/models/bar.sql": `SELECT * FROM foo`,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 3, 0, 0)
	falsy := false
	model := &runtimev1.ModelV2{
		Spec: &runtimev1.ModelSpec{
			Connector:   "duckdb",
			Sql:         "SELECT * FROM foo",
			Materialize: &falsy,
		},
		State: &runtimev1.ModelState{
			Connector: "duckdb",
			Table:     "bar",
		},
	}
	modelRes := &runtimev1.Resource{
		Meta: &runtimev1.ResourceMeta{
			Name:      &runtimev1.ResourceName{Kind: runtime.ResourceKindModel, Name: "bar"},
			Refs:      []*runtimev1.ResourceName{{Kind: runtime.ResourceKindSource, Name: "foo"}},
			Owner:     runtime.GlobalProjectParserName,
			FilePaths: []string{"/models/bar.sql"},
		},
		Resource: &runtimev1.Resource_Model{
			Model: model,
		},
	}
	testruntime.RequireResource(t, rt, id, modelRes)
	testruntime.RequireOLAPTable(t, rt, id, "bar")

	// Update model to have a CTE with alias different from the source
	testruntime.PutFiles(t, rt, id, map[string]string{
		"/models/bar.sql": `with CTEAlias as (select * from foo) select * from CTEAlias`,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 3, 0, 0)
	model.Spec.Sql = `with CTEAlias as (select * from foo) select * from CTEAlias`
	testruntime.RequireResource(t, rt, id, modelRes)
	testruntime.RequireOLAPTable(t, rt, id, "bar")

	// Update model to have a CTE with alias same as the source
	testruntime.PutFiles(t, rt, id, map[string]string{
		"/models/bar.sql": `with foo as (select * from foo) select * from foo`,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 3, 0, 0)
	model.Spec.Sql = `with foo as (select * from foo) select * from foo`
	modelRes.Meta.Refs = []*runtimev1.ResourceName{}
	testruntime.RequireResource(t, rt, id, modelRes)
	// Refs are removed but the model is valid.
	// TODO: is this expected?
	testruntime.RequireOLAPTable(t, rt, id, "bar")
}

func TestRename(t *testing.T) {
	// Rename model A to B and model B to A, verify success
	// Rename model A to B and source B to A, verify success

	// Create source and model
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
		"/models/bar.sql": `SELECT * FROM foo`,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 3, 0, 0)
	falsy := false
	model := &runtimev1.ModelV2{
		Spec: &runtimev1.ModelSpec{
			Connector:   "duckdb",
			Sql:         "SELECT * FROM foo",
			Materialize: &falsy,
		},
		State: &runtimev1.ModelState{
			Connector: "duckdb",
			Table:     "bar",
		},
	}
	modelRes := &runtimev1.Resource{
		Meta: &runtimev1.ResourceMeta{
			Name:      &runtimev1.ResourceName{Kind: runtime.ResourceKindModel, Name: "bar"},
			Refs:      []*runtimev1.ResourceName{{Kind: runtime.ResourceKindSource, Name: "foo"}},
			Owner:     runtime.GlobalProjectParserName,
			FilePaths: []string{"/models/bar.sql"},
		},
		Resource: &runtimev1.Resource_Model{
			Model: model,
		},
	}
	testruntime.RequireResource(t, rt, id, modelRes)
	testruntime.RequireOLAPTable(t, rt, id, "bar")

	// Rename the model
	testruntime.RenameFile(t, rt, id, "/models/bar.sql", "/models/bar_new.sql")
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 3, 0, 0)
	modelRes.Meta.Name.Name = "bar_new"
	modelRes.Meta.FilePaths[0] = "/models/bar_new.sql"
	model.State.Table = "bar_new"
	testruntime.RequireResource(t, rt, id, modelRes)
	testruntime.RequireOLAPTable(t, rt, id, "bar_new")
	testruntime.RequireNoOLAPTable(t, rt, id, "bar")

	// Rename the model to same name but different case
	testruntime.RenameFile(t, rt, id, "/models/bar_new.sql", "/models/Bar_New.sql")
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 3, 0, 0)
	modelRes.Meta.Name.Name = "Bar_New"
	modelRes.Meta.FilePaths[0] = "/models/Bar_New.sql"
	model.State.Table = "Bar_New"
	testruntime.RequireResource(t, rt, id, modelRes)
	testruntime.RequireOLAPTable(t, rt, id, "Bar_New")

	// Create a model referencing the new model name from before
	testruntime.PutFiles(t, rt, id, map[string]string{
		"/models/bar_another.sql": `SELECT * FROM Bar_New`,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 4, 0, 0)
	anotherModel := &runtimev1.ModelV2{
		Spec: &runtimev1.ModelSpec{
			Connector:   "duckdb",
			Sql:         "SELECT * FROM Bar_New",
			Materialize: &falsy,
		},
		State: &runtimev1.ModelState{
			Connector: "duckdb",
			Table:     "bar_another",
		},
	}
	anotherModelRes := &runtimev1.Resource{
		Meta: &runtimev1.ResourceMeta{
			Name:      &runtimev1.ResourceName{Kind: runtime.ResourceKindModel, Name: "bar_another"},
			Refs:      []*runtimev1.ResourceName{{Kind: runtime.ResourceKindModel, Name: "Bar_New"}},
			Owner:     runtime.GlobalProjectParserName,
			FilePaths: []string{"/models/bar_another.sql"},
		},
		Resource: &runtimev1.Resource_Model{
			Model: anotherModel,
		},
	}
	testruntime.RequireResource(t, rt, id, anotherModelRes)
	testruntime.RequireOLAPTable(t, rt, id, "bar_another")

	// Rename the source to the model's name
	testruntime.RenameFile(t, rt, id, "/sources/foo.yaml", "/sources/Bar_New.yaml")
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 3, 1, 1)
	testruntime.RequireOLAPTable(t, rt, id, "Bar_New")
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

func assertTableRows(t testing.TB, rt *runtime.Runtime, id, table string, limit int) {
	q := &queries.TableHead{
		TableName: table,
		Limit:     3,
	}
	require.NoError(t, rt.Query(context.Background(), id, q, 5))
	require.Len(t, q.Result, limit)
}

func assertIsView(t testing.TB, olap drivers.OLAPStore, tableName string, isView bool) {
	table, err := olap.InformationSchema().Lookup(context.Background(), tableName)
	require.NoError(t, err)
	// Assert that the model is a table now
	require.Equal(t, table.View, isView)
}
