package reconcilers_test

import (
	"testing"

	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
)

func TestResourceMetadata(t *testing.T) {
	rt, id := testruntime.NewInstance(t)
	testruntime.PutFiles(t, rt, id, map[string]string{
		"models/m1.sql": `SELECT 1 AS id`,
		"metrics_views/mv1.yaml": `
version: 1
type: metrics_view
model: m1
metadata:
  owner: data-team
  tier: 1
dimensions:
- column: id
measures:
- name: count
  expression: count(*)
`,
	})

	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 3, 0, 0)

	mv := testruntime.GetResource(t, rt, id, runtime.ResourceKindMetricsView, "mv1")
	require.Equal(t, map[string]string{"owner": "data-team", "tier": "1"}, mv.Meta.Metadata)
	specVersion := mv.Meta.SpecVersion

	// Changing only the metadata must be detected and bump the spec version.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"metrics_views/mv1.yaml": `
version: 1
type: metrics_view
model: m1
metadata:
  owner: platform-team
dimensions:
- column: id
measures:
- name: count
  expression: count(*)
`,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 3, 0, 0)

	mv = testruntime.GetResource(t, rt, id, runtime.ResourceKindMetricsView, "mv1")
	require.Equal(t, map[string]string{"owner": "platform-team"}, mv.Meta.Metadata)
	require.Greater(t, mv.Meta.SpecVersion, specVersion)

	// Removing the metadata clears it.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"metrics_views/mv1.yaml": `
version: 1
type: metrics_view
model: m1
dimensions:
- column: id
measures:
- name: count
  expression: count(*)
`,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 3, 0, 0)

	mv = testruntime.GetResource(t, rt, id, runtime.ResourceKindMetricsView, "mv1")
	require.Empty(t, mv.Meta.Metadata)
}
