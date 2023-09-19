package runtime_test

import (
	"context"
	"strings"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestSource(t *testing.T) {
	rt, id := testruntime.NewInstance(t)
	putFiles(t, rt, id, map[string]string{
		"/data/foo.csv": `
a,b,c,d,e
1,2,3,4,5
1,2,3,4,5
1,2,3,4,5
`,
		"/sources/foo.yaml": `
type: local_file
path: data/foo.csv
`,
	})
	reconcileAndWait(t, rt, id)
	dumpCatalog(t, rt, id)
	requireCatalog(t, rt, id, 2, 0, 0)
	requireResource(t, rt, id, &runtimev1.Resource{
		Meta: &runtimev1.ResourceMeta{
			Name:      &runtimev1.ResourceName{Kind: runtime.ResourceKindSource, Name: "foo"},
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
}

func putFiles(t testing.TB, rt *runtime.Runtime, id string, files map[string]string) {
	ctx := context.Background()
	repo, release, err := rt.Repo(ctx, id)
	require.NoError(t, err)
	defer release()

	for path, data := range files {
		err := repo.Put(ctx, path, strings.NewReader(data))
		require.NoError(t, err)
	}
}

func deleteFiles(t testing.TB, rt *runtime.Runtime, id string, files ...string) {
	ctx := context.Background()
	repo, release, err := rt.Repo(ctx, id)
	require.NoError(t, err)
	defer release()

	for _, path := range files {
		err := repo.Delete(ctx, path)
		require.NoError(t, err)
	}
}

func reconcileAndWait(t testing.TB, rt *runtime.Runtime, id string) {
	ctrl, err := rt.Controller(id)
	require.NoError(t, err)

	err = ctrl.Reconcile(context.Background(), runtime.GlobalProjectParserName)
	require.NoError(t, err)

	err = ctrl.WaitUntilIdle(context.Background())
	require.NoError(t, err)
}

func requireResource(t testing.TB, rt *runtime.Runtime, id string, a *runtimev1.Resource) {
	ctrl, err := rt.Controller(id)
	require.NoError(t, err)

	b, err := ctrl.Get(context.Background(), a.Meta.Name, false)
	require.NoError(t, err)

	require.Equal(t, a.Meta.Name, b.Meta.Name)
	require.Equal(t, a.Meta.Refs, b.Meta.Refs)
	require.Equal(t, a.Meta.Owner, b.Meta.Owner)
	require.Equal(t, a.Meta.FilePaths, b.Meta.FilePaths)
	require.Greater(t, b.Meta.Version, 0)
	require.Greater(t, b.Meta.SpecVersion, 0)
	require.Greater(t, b.Meta.StateVersion, 0)
	require.NotEmpty(t, b.Meta.CreatedOn.AsTime())
	require.NotEmpty(t, b.Meta.SpecUpdatedOn.AsTime())
	require.NotEmpty(t, b.Meta.StateUpdatedOn.AsTime())
	require.Equal(t, b.Meta.DeletedOn, nil)
	// require.Equal(t, a.Meta.ReconcileStatus, b.Meta.ReconcileStatus)
	// require.Equal(t, a.Meta.ReconcileError, b.Meta.ReconcileError)
	// require.Equal(t, a.Meta.ReconcileOn, b.Meta.ReconcileOn)
	// require.Equal(t, a.Meta.RenamedFrom, b.Meta.RenamedFrom)

	require.Equal(t, a.Resource, b.Resource, "for resource %q", a.Meta.Name)
}

func requireCatalog(t testing.TB, rt *runtime.Runtime, id string, resources, reconcileErrs, parseErrs int) {
	ctrl, err := rt.Controller(id)
	require.NoError(t, err)

	rs, err := ctrl.List(context.Background(), "", false)
	require.NoError(t, err)

	var seenReconcileErrs, seenParseErrs int
	for _, r := range rs {
		if r.Meta.ReconcileError != "" {
			seenReconcileErrs++
		}

		if r.Meta.Name.Kind == runtime.ResourceKindProjectParser {
			seenParseErrs += len(r.GetProjectParser().State.ParseErrors)
		}
	}

	require.Equal(t, resources, len(rs), "resources")
	require.Equal(t, reconcileErrs, seenReconcileErrs, "reconcile errors")
	require.Equal(t, parseErrs, seenParseErrs, "parse errors")
}

func dumpCatalog(t testing.TB, rt *runtime.Runtime, id string) {
	ctrl, err := rt.Controller(id)
	require.NoError(t, err)

	rs, err := ctrl.List(context.Background(), "", false)
	require.NoError(t, err)

	for _, r := range rs {
		t.Logf("%s/%s: status=%d, stateversion=%d, error=%q", r.Meta.Name.Kind, r.Meta.Name.Name, r.Meta.ReconcileStatus, r.Meta.StateVersion, r.Meta.ReconcileError)
	}
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
