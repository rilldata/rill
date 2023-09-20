package runtime_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestSource(t *testing.T) {
	rt, id := testruntime.NewInstance(t)
	putFiles(t, rt, id, map[string]string{
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
	reconcileAndWait(t, rt, id)
	requireCatalog(t, rt, id, 2, 0, 0)
	requireResource(t, rt, id, &runtimev1.Resource{
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

	b, err := ctrl.Get(context.Background(), a.Meta.Name, true) // Set clone=true because we may manipulate it before comparing
	require.NoError(t, err)

	require.True(t, proto.Equal(a.Meta.Name, b.Meta.Name), "expected: %v\nactual: %v", a.Meta.Name, b.Meta.Name)
	require.ElementsMatch(t, a.Meta.Refs, b.Meta.Refs)
	require.True(t, proto.Equal(a.Meta.Owner, b.Meta.Owner), "expected: %v\nactual: %v", a.Meta.Owner, b.Meta.Owner)
	require.ElementsMatch(t, a.Meta.FilePaths, b.Meta.FilePaths)
	require.Greater(t, b.Meta.Version, int64(0))
	require.Greater(t, b.Meta.SpecVersion, int64(0))
	require.Greater(t, b.Meta.StateVersion, int64(0))
	require.NotEmpty(t, b.Meta.CreatedOn.AsTime())
	require.NotEmpty(t, b.Meta.SpecUpdatedOn.AsTime())
	require.NotEmpty(t, b.Meta.StateUpdatedOn.AsTime())
	require.Nil(t, b.Meta.DeletedOn)
	// require.Equal(t, a.Meta.ReconcileStatus, b.Meta.ReconcileStatus)
	// require.Equal(t, a.Meta.ReconcileError, b.Meta.ReconcileError)
	// require.Equal(t, a.Meta.ReconcileOn, b.Meta.ReconcileOn)
	// require.Equal(t, a.Meta.RenamedFrom, b.Meta.RenamedFrom)

	// Some kind-specific fields are ephemeral. We reset those to stable values before comparing.
	switch b.Meta.Name.Kind {
	case runtime.ResourceKindSource:
		state := b.GetSource().State
		state.RefreshedOn = nil
		state.SpecHash = ""
	}

	// Hack to only compare the Resource field (not Meta)
	name := b.Meta.Name
	a = &runtimev1.Resource{Resource: a.Resource}
	b = &runtimev1.Resource{Resource: b.Resource}

	// Compare!
	require.True(t, proto.Equal(a, b), "for resource %q\nexpected: %v\nactual: %v", name.Name, a.Resource, b.Resource)
}

func requireCatalog(t testing.TB, rt *runtime.Runtime, id string, resources, lenReconcileErrs, lenParseErrs int) {
	ctrl, err := rt.Controller(id)
	require.NoError(t, err)

	rs, err := ctrl.List(context.Background(), "", false)
	require.NoError(t, err)

	var reconcileErrs, parseErrs []string
	for _, r := range rs {
		if r.Meta.ReconcileError != "" {
			reconcileErrs = append(reconcileErrs, fmt.Sprintf("%s/%s: %s", r.Meta.Name.Kind, r.Meta.Name.Name, r.Meta.ReconcileError))
		}

		if r.Meta.Name.Kind == runtime.ResourceKindProjectParser {
			for _, pe := range r.GetProjectParser().State.ParseErrors {
				parseErrs = append(parseErrs, fmt.Sprintf("%s: %s", pe.FilePath, pe.Message))
			}
		}
	}

	require.Equal(t, lenParseErrs, len(parseErrs), "parse errors: %s", strings.Join(parseErrs, "\n"))
	require.Equal(t, lenReconcileErrs, len(reconcileErrs), "reconcile errors: %s", strings.Join(reconcileErrs, "\n"))
	require.Equal(t, resources, len(rs), "resources")
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
