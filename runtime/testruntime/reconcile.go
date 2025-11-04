package testruntime

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func PutFiles(t testing.TB, rt *runtime.Runtime, id string, files map[string]string) {
	ctx := t.Context()
	repo, release, err := rt.Repo(ctx, id)
	require.NoError(t, err)
	defer release()

	for path, data := range files {
		err := repo.Put(ctx, path, strings.NewReader(strings.TrimSpace(data)))
		require.NoError(t, err)
	}
}

func RenameFile(t testing.TB, rt *runtime.Runtime, id, from, to string) {
	ctx := t.Context()
	repo, release, err := rt.Repo(ctx, id)
	require.NoError(t, err)
	defer release()

	require.NoError(t, repo.Rename(ctx, from, to))
}

func DeleteFiles(t testing.TB, rt *runtime.Runtime, id string, files ...string) {
	ctx := t.Context()
	repo, release, err := rt.Repo(ctx, id)
	require.NoError(t, err)
	defer release()

	for _, path := range files {
		err := repo.Delete(ctx, path, false)
		require.NoError(t, err)
	}
}

func ReconcileParserAndWait(t testing.TB, rt *runtime.Runtime, id string) {
	ReconcileAndWait(t, rt, id, runtime.GlobalProjectParserName)
}

func ReconcileAndWait(t testing.TB, rt *runtime.Runtime, id string, n *runtimev1.ResourceName) {
	ctx := t.Context()
	ctrl, err := rt.Controller(ctx, id)
	require.NoError(t, err)

	err = ctrl.Reconcile(ctx, n)
	require.NoError(t, err)

	err = ctrl.WaitUntilIdle(ctx, false)
	require.NoError(t, err)
}

func RefreshAndWait(t testing.TB, rt *runtime.Runtime, id string, n *runtimev1.ResourceName) {
	ctx := t.Context()
	ctrl, err := rt.Controller(ctx, id)
	require.NoError(t, err)

	// Get resource before refresh
	rPrev, err := ctrl.Get(ctx, n, false)
	require.NoError(t, err)

	// Create refresh trigger
	trgName := &runtimev1.ResourceName{Kind: runtime.ResourceKindRefreshTrigger, Name: time.Now().String()}
	err = ctrl.Create(ctx, trgName, nil, nil, nil, false, &runtimev1.Resource{
		Resource: &runtimev1.Resource_RefreshTrigger{
			RefreshTrigger: &runtimev1.RefreshTrigger{
				Spec: &runtimev1.RefreshTriggerSpec{
					Resources: []*runtimev1.ResourceName{n},
				},
			},
		},
	})
	require.NoError(t, err)

	// Wait for refresh to complete
	err = ctrl.WaitUntilIdle(ctx, false)
	require.NoError(t, err)

	// Get resource after refresh
	rNew, err := ctrl.Get(ctx, n, false)
	require.NoError(t, err)

	// Check the resource's spec version has increased
	require.Greater(t, rNew.Meta.SpecVersion, rPrev.Meta.SpecVersion)
}

func RequireReconcileState(t testing.TB, rt *runtime.Runtime, id string, lenResources, lenReconcileErrs, lenParseErrs int) {
	ctx := t.Context()
	ctrl, err := rt.Controller(ctx, id)
	require.NoError(t, err)

	rs, err := ctrl.List(ctx, "", "", false)
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

	var names []string
	for _, r := range rs {
		names = append(names, fmt.Sprintf("%s/%s", r.Meta.Name.Kind, r.Meta.Name.Name))
	}

	if lenParseErrs >= 0 {
		require.Equal(t, lenParseErrs, len(parseErrs), "parse errors: %s", strings.Join(parseErrs, "\n"))
	}
	if lenReconcileErrs >= 0 {
		require.Equal(t, lenReconcileErrs, len(reconcileErrs), "reconcile errors: %s", strings.Join(reconcileErrs, "\n"))
	}
	if lenResources >= 0 {
		require.Equal(t, lenResources, len(rs), "resources: %s", strings.Join(names, "\n"))
	}
}

func GetResource(t testing.TB, rt *runtime.Runtime, id, kind, name string) *runtimev1.Resource {
	ctx := t.Context()
	ctrl, err := rt.Controller(ctx, id)
	require.NoError(t, err)

	r, err := ctrl.Get(ctx, &runtimev1.ResourceName{Kind: kind, Name: name}, true)
	require.NoError(t, err)

	return r
}

func RequireResource(t testing.TB, rt *runtime.Runtime, id string, a *runtimev1.Resource) {
	ctx := t.Context()
	ctrl, err := rt.Controller(ctx, id)
	require.NoError(t, err)

	b, err := ctrl.Get(ctx, a.Meta.Name, true) // Set clone=true because we may manipulate it before comparing
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

	// Checking ReconcileError using Contains instead of Equal
	if a.Meta.ReconcileError == "" {
		require.Empty(t, b.Meta.ReconcileError)
	} else {
		require.Contains(t, b.Meta.ReconcileError, a.Meta.ReconcileError)
	}

	// Not comparing these fields because they are not stable:
	// require.Equal(t, a.Meta.ReconcileStatus, b.Meta.ReconcileStatus)
	// require.Equal(t, a.Meta.ReconcileOn, b.Meta.ReconcileOn)
	// require.Equal(t, a.Meta.RenamedFrom, b.Meta.RenamedFrom)

	// Some kind-specific fields are not stable. We reset those to stable values before comparing.
	switch b.Meta.Name.Kind {
	case runtime.ResourceKindSource:
		state := b.GetSource().State
		state.RefreshedOn = nil
		state.SpecHash = ""
	case runtime.ResourceKindModel:
		state := b.GetModel().State
		state.LatestExecutionDurationMs = 0
		state.TotalExecutionDurationMs = 0
		state.RefreshedOn = nil
		state.SpecHash = ""
		state.RefsHash = ""
		state.TestHash = ""
		state.LatestExecutionDurationMs = 0
		state.TotalExecutionDurationMs = 0
	case runtime.ResourceKindMetricsView:
		state := b.GetMetricsView().State
		state.DataRefreshedOn = nil
	case runtime.ResourceKindExplore:
		state := b.GetExplore().State
		state.DataRefreshedOn = nil
	case runtime.ResourceKindComponent:
		state := b.GetComponent().State
		state.DataRefreshedOn = nil
	case runtime.ResourceKindCanvas:
		state := b.GetCanvas().State
		state.DataRefreshedOn = nil
	case runtime.ResourceKindAlert:
		state := b.GetAlert().State
		state.SpecHash = ""
		state.RefsHash = ""
		for i, e := range state.ExecutionHistory {
			e.StartedOn = nil
			e.FinishedOn = nil
			if a.GetAlert().State.ExecutionHistory[i].ExecutionTime == nil {
				e.ExecutionTime = nil
			}
		}
	case runtime.ResourceKindConnector:
		state := b.GetConnector().State
		state.SpecHash = ""
	}

	// Hack to only compare the Resource field (not Meta)
	name := b.Meta.Name
	a = &runtimev1.Resource{Resource: a.Resource}
	b = &runtimev1.Resource{Resource: b.Resource}

	// Compare!
	require.True(t, proto.Equal(a, b), "for resource %q\nexpected: %v\nactual: %v", name.Name, a.Resource, b.Resource)
}

func DumpResources(t testing.TB, rt *runtime.Runtime, id string) {
	ctx := t.Context()
	ctrl, err := rt.Controller(ctx, id)
	require.NoError(t, err)

	rs, err := ctrl.List(ctx, "", "", false)
	require.NoError(t, err)

	for _, r := range rs {
		t.Logf("%s/%s: status=%d, stateversion=%d, error=%q", r.Meta.Name.Kind, r.Meta.Name.Name, r.Meta.ReconcileStatus, r.Meta.StateVersion, r.Meta.ReconcileError)
	}
}

func RequireParseErrors(t testing.TB, rt *runtime.Runtime, id string, expectedParseErrors map[string]string) {
	ctx := t.Context()
	ctrl, err := rt.Controller(ctx, id)
	require.NoError(t, err)

	pp, err := ctrl.Get(ctx, runtime.GlobalProjectParserName, true)
	require.NoError(t, err)

	parseErrs := map[string]string{}
	for _, pe := range pp.GetProjectParser().State.ParseErrors {
		parseErrs[pe.FilePath] = pe.Message
	}
	require.Len(t, parseErrs, len(expectedParseErrors), "Should have %d parse errors", len(expectedParseErrors))

	for f, pe := range parseErrs {
		// Checking parseError using Contains instead of Equal
		require.Contains(t, pe, expectedParseErrors[f])
	}
}

type RequireResolveOptions struct {
	Resolver           string
	Properties         map[string]any
	Args               map[string]any
	UserAttributes     map[string]any
	AdditionalRules    []*runtimev1.SecurityRule
	SkipSecurityChecks bool

	Result        []map[string]any
	ResultCSV     string
	ErrorContains string
	Update        bool
}

func RequireResolve(t testing.TB, rt *runtime.Runtime, id string, opts *RequireResolveOptions) {
	// Run the resolver.
	ctx := t.Context()
	res, err := rt.Resolve(ctx, &runtime.ResolveOptions{
		InstanceID:         id,
		Resolver:           opts.Resolver,
		ResolverProperties: opts.Properties,
		Args:               opts.Args,
		Claims: &runtime.SecurityClaims{
			UserAttributes:  opts.UserAttributes,
			AdditionalRules: opts.AdditionalRules,
			SkipChecks:      opts.SkipSecurityChecks,
		},
	})

	// If it succeeded, get the result rows.
	// Does a JSON roundtrip to coerce to simple types (easier to compare).
	var rows []map[string]any
	if err == nil {
		data, err2 := res.MarshalJSON()
		if err2 != nil {
			err = err2
		} else {
			err = json.Unmarshal(data, &rows)
		}
	}

	// If the Update flag is set, update the results in opts instead of checking them.
	// The caller can then access the updated values.
	if opts.Update {
		opts.Result = rows
		opts.ResultCSV = ""
		if res != nil {
			opts.ResultCSV = resultToCSV(t, rows, res.Schema())
		}
		opts.ErrorContains = ""
		if err != nil {
			opts.ErrorContains = err.Error()
		}
		return
	}

	// Check if an error was expected.
	if opts.ErrorContains != "" {
		require.Error(t, err)
		require.Contains(t, err.Error(), opts.ErrorContains)
		return
	}
	require.NoError(t, err)

	// We support expressing the expected result as a CSV string, which is more compact.
	// Serialize the result to CSV and compare.
	if opts.ResultCSV != "" {
		actual := resultToCSV(t, rows, res.Schema())
		require.Equal(t, strings.TrimSpace(opts.ResultCSV), strings.TrimSpace(actual))
		return
	}

	// Compare the result rows to the expected result.
	// Like for rows, we do a JSON roundtrip on the expected result (parsed from YAML) to coerce to simple types.
	var expected []map[string]any
	data, err := json.Marshal(opts.Result)
	require.NoError(t, err)
	err = json.Unmarshal(data, &expected)
	require.NoError(t, err)
	if len(expected) != 0 || len(rows) != 0 {
		require.EqualValues(t, expected, rows)
	}
}

func Must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

// resultToCSV serializes the rows to a CSV formatted string.
// It is derived from runtime/drivers/file/model_executor_olap_self.go#writeCSV.
func resultToCSV(t testing.TB, rows []map[string]any, schema *runtimev1.StructType) string {
	buf := &bytes.Buffer{}
	w := csv.NewWriter(buf)

	strs := make([]string, len(schema.Fields))
	for i, f := range schema.Fields {
		strs[i] = f.Name
	}
	err := w.Write(strs)
	require.NoError(t, err)

	for _, row := range rows {
		for i, f := range schema.Fields {
			v, ok := row[f.Name]
			require.True(t, ok, "missing field %q", f.Name)

			var s string
			if v != nil {
				if v2, ok := v.(string); ok {
					s = v2
				} else {
					tmp, err := json.Marshal(v)
					require.NoError(t, err)
					s = string(tmp)
				}
			}

			strs[i] = s
		}

		err = w.Write(strs)
		require.NoError(t, err)
	}

	w.Flush()
	return buf.String()
}
