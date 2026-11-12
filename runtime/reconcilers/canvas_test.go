package reconcilers_test

import (
	"context"
	"strings"
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/testruntime"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	_ "github.com/rilldata/rill/runtime/resolvers"
)

func TestCanvasValidState(t *testing.T) {
	// Create an instance with StageChanges==true
	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files:        map[string]string{"rill.yaml": ""},
		StageChanges: true,
	})

	// Create basic model + metrics_view + canvas
	testruntime.PutFiles(t, rt, id, map[string]string{
		"m1.sql": `SELECT 'foo' as foo, 1 as x`,
		"mv1.yaml": `
version: 1
type: metrics_view
model: m1
dimensions:
- column: foo
measures:
- name: x
  expression: sum(x)
`,
		"c1.yaml": `
type: canvas
rows:
  - items:
      - kpi_grid:
          metrics_view: mv1
          measures:
            - x
  - items:
      - kpi_grid:
          metrics_view: mv1
          measures:
            - x
`,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 6, 0, 0)
	c1 := testruntime.GetResource(t, rt, id, runtime.ResourceKindCanvas, "c1")
	require.NotNil(t, c1.GetCanvas().State.ValidSpec)

	// Change the model so it breaks the metrics view and canvas, check valid spec is preserved
	testruntime.PutFiles(t, rt, id, map[string]string{"m1.sql": `SELECT 'bar' as bar, 2 as y`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 6, 4, 0)
	mv1 := testruntime.GetResource(t, rt, id, runtime.ResourceKindMetricsView, "mv1")
	require.NotNil(t, mv1.GetMetricsView().State.ValidSpec)
	c1 = testruntime.GetResource(t, rt, id, runtime.ResourceKindCanvas, "c1")
	require.NotNil(t, c1.GetCanvas().State.ValidSpec)
	require.Len(t, c1.Meta.Refs, 2)
	for _, componentName := range c1.Meta.Refs {
		r := testruntime.GetResource(t, rt, id, runtime.ResourceKindComponent, componentName.Name)
		require.NotNil(t, r.GetComponent().State.ValidSpec)
	}

	// Fix everything
	testruntime.PutFiles(t, rt, id, map[string]string{"m1.sql": `SELECT 'foo' as foo, 1 as x`})
	testruntime.ReconcileParserAndWait(t, rt, id)

	// Break one canvas component. Check valid spec is preserved.
	testruntime.PutFiles(t, rt, id, map[string]string{"c1.yaml": `
type: canvas
rows:
  - items:
      - kpi_grid:
          metrics_view: doesnt_exist
          measures:
            - x
  - items:
      - kpi_grid:
          metrics_view: mv1
          measures:
            - x
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 6, 2, 0)
	mv1 = testruntime.GetResource(t, rt, id, runtime.ResourceKindMetricsView, "mv1")
	require.NotNil(t, mv1.GetMetricsView().State.ValidSpec)
	c1 = testruntime.GetResource(t, rt, id, runtime.ResourceKindCanvas, "c1")
	require.NotNil(t, c1.GetCanvas().State.ValidSpec)
	require.NotEmpty(t, c1.Meta.ReconcileError)
	require.Len(t, c1.Meta.Refs, 2)
	var valid, invalid int
	for _, componentName := range c1.Meta.Refs {
		r := testruntime.GetResource(t, rt, id, runtime.ResourceKindComponent, componentName.Name)
		if r.GetComponent().State.ValidSpec == nil {
			invalid++
		} else {
			valid++
		}
	}
	require.Equal(t, 1, valid)
	require.Equal(t, 1, invalid)
}

func TestCanvasValidateMetricsViewTimeConsistency(t *testing.T) {
	// Create an instance with StageChanges==true
	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files:        map[string]string{"rill.yaml": ""},
		StageChanges: true,
	})

	// Create two models, two metrics views with consistent time settings, and a canvas referencing both
	testruntime.PutFiles(t, rt, id, map[string]string{
		"m1.sql": `SELECT 'foo' as foo, 1 as x`,
		"m2.sql": `SELECT 'bar' as bar, 2 as y`,
		"mv1.yaml": `
version: 1
type: metrics_view
model: m1
dimensions:
- column: foo
measures:
- name: x
  expression: sum(x)
first_day_of_week: 2
first_month_of_year: 3
`,
		"mv2.yaml": `
version: 1
type: metrics_view
model: m2
dimensions:
- column: bar
measures:
- name: y
  expression: sum(y)
first_day_of_week: 2
first_month_of_year: 3
`,
		"c1.yaml": `
type: canvas
rows:
  - items:
      - kpi_grid:
          metrics_view: mv1
          measures:
            - x
  - items:
      - kpi_grid:
          metrics_view: mv2
          measures:
            - y
`,
	})

	// Reconcile and verify success
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 8, 0, 0)

	// Verify that the canvas and metrics views have valid specs
	c1 := testruntime.GetResource(t, rt, id, runtime.ResourceKindCanvas, "c1")
	require.NotNil(t, c1.GetCanvas().State.ValidSpec)

	mv1 := testruntime.GetResource(t, rt, id, runtime.ResourceKindMetricsView, "mv1")
	require.NotNil(t, mv1.GetMetricsView().State.ValidSpec)
	require.Equal(t, uint32(2), mv1.GetMetricsView().State.ValidSpec.FirstDayOfWeek)
	require.Equal(t, uint32(3), mv1.GetMetricsView().State.ValidSpec.FirstMonthOfYear)

	mv2 := testruntime.GetResource(t, rt, id, runtime.ResourceKindMetricsView, "mv2")
	require.NotNil(t, mv2.GetMetricsView().State.ValidSpec)
	require.Equal(t, uint32(2), mv2.GetMetricsView().State.ValidSpec.FirstDayOfWeek)
	require.Equal(t, uint32(3), mv2.GetMetricsView().State.ValidSpec.FirstMonthOfYear)

	// Change one metrics view to have inconsistent time settings
	testruntime.PutFiles(t, rt, id, map[string]string{
		"mv2.yaml": `
version: 1
type: metrics_view
model: m2
dimensions:
- column: bar
measures:
- name: y
  expression: sum(y)
first_day_of_week: 1
first_month_of_year: 1
`,
	})

	// Reconcile and verify that the metrics view gets updated
	testruntime.ReconcileParserAndWait(t, rt, id)

	// Verify that mv2's valid spec got updated
	mv2 = testruntime.GetResource(t, rt, id, runtime.ResourceKindMetricsView, "mv2")
	require.NotNil(t, mv2.GetMetricsView().State.ValidSpec)
	require.Equal(t, uint32(1), mv2.GetMetricsView().State.ValidSpec.FirstDayOfWeek)
	require.Equal(t, uint32(1), mv2.GetMetricsView().State.ValidSpec.FirstMonthOfYear)

	// Verify that the canvas reconciliation fails with the expected error
	c1 = testruntime.GetResource(t, rt, id, runtime.ResourceKindCanvas, "c1")
	require.NotEmpty(t, c1.Meta.ReconcileError)
	require.Contains(t, c1.Meta.ReconcileError, "inconsistent first_day_of_week")

	// With StageChanges==true, the valid spec should be preserved
	require.NotNil(t, c1.GetCanvas().State.ValidSpec)

	// Fix the metrics view to have consistent time settings again
	testruntime.PutFiles(t, rt, id, map[string]string{
		"mv2.yaml": `
version: 1
type: metrics_view
model: m2
dimensions:
- column: bar
measures:
- name: y
  expression: sum(y)
first_day_of_week: 2
first_month_of_year: 3
`,
	})

	// Reconcile and verify success again
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 8, 0, 0)

	// Verify that mv2's valid spec got updated again
	mv2 = testruntime.GetResource(t, rt, id, runtime.ResourceKindMetricsView, "mv2")
	require.NotNil(t, mv2.GetMetricsView().State.ValidSpec)
	require.Equal(t, uint32(2), mv2.GetMetricsView().State.ValidSpec.FirstDayOfWeek)
	require.Equal(t, uint32(3), mv2.GetMetricsView().State.ValidSpec.FirstMonthOfYear)

	// Verify that the canvas reconciles successfully again
	c1 = testruntime.GetResource(t, rt, id, runtime.ResourceKindCanvas, "c1")
	require.NotNil(t, c1.GetCanvas().State.ValidSpec)
	require.Empty(t, c1.Meta.ReconcileError)
}

func TestCanvasDataRefreshedOn(t *testing.T) {
	// Create an instance with StageChanges==true
	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files:        map[string]string{"rill.yaml": ""},
		StageChanges: true,
	})

	// Create basic model + metrics_view + canvas
	testruntime.PutFiles(t, rt, id, map[string]string{
		"m1.sql": `
-- @materialize: true
SELECT 'foo' as foo, 1 as x
`,
		"mv1.yaml": `
version: 1
type: metrics_view
model: m1
dimensions:
- column: foo
measures:
- name: x
  expression: sum(x)
`,
		"c1.yaml": `
type: canvas
rows:
  - items:
    - kpi_grid:
        metrics_view: mv1
        measures:
          - x
`,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 5, 0, 0)

	getAndCheckRefreshedOn := func() time.Time {
		c1 := testruntime.GetResource(t, rt, id, runtime.ResourceKindCanvas, "c1")
		require.NotNil(t, c1.GetCanvas().State.DataRefreshedOn)

		comp1 := testruntime.GetResource(t, rt, id, runtime.ResourceKindComponent, c1.Meta.Refs[0].Name)
		require.NotNil(t, comp1.GetComponent().State.DataRefreshedOn)

		mv1 := testruntime.GetResource(t, rt, id, runtime.ResourceKindMetricsView, "mv1")
		require.NotNil(t, mv1.GetMetricsView().State.DataRefreshedOn)

		require.Equal(t, c1.GetCanvas().State.DataRefreshedOn, comp1.GetComponent().State.DataRefreshedOn)
		require.Equal(t, c1.GetCanvas().State.DataRefreshedOn, mv1.GetMetricsView().State.DataRefreshedOn)

		return c1.GetCanvas().State.DataRefreshedOn.AsTime()
	}

	refreshedOn1 := getAndCheckRefreshedOn()
	testruntime.RefreshAndWait(t, rt, id, &runtimev1.ResourceName{Kind: runtime.ResourceKindModel, Name: "m1"})
	refreshedOn2 := getAndCheckRefreshedOn()
	require.Greater(t, refreshedOn2, refreshedOn1)
}

func TestCanvasResolveTransitiveAccess(t *testing.T) {
	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{"rill.yaml": ""},
	})

	// Create three models, three metrics views, and a canvas with components using metrics_view, metrics_sql, and markdown.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"m1.sql": `SELECT 'foo' as foo, 1 as x`,
		"m2.sql": `SELECT 'bar' as bar, 2 as y`,
		"m3.sql": `SELECT 'baz' as baz, 3 as z`,
		"m4.sql": `SELECT 'qux' as qux, 4 as w`,
		"mv1.yaml": `
version: 1
type: metrics_view
model: m1
dimensions:
- column: foo
measures:
- name: x
  expression: sum(x)
`,
		"mv2.yaml": `
version: 1
type: metrics_view
model: m2
dimensions:
- column: bar
measures:
- name: y
  expression: sum(y)
`,
		"mv3.yaml": `
version: 1
type: metrics_view
model: m3
dimensions:
- column: baz
measures:
- name: z
  expression: sum(z)
`,
		"mv4.yaml": `
version: 1
type: metrics_view
model: m4
dimensions:
- column: qux
measures:
- name: w
  expression: sum(w)
`,
		"c1.yaml": `
type: canvas
rows:
  - items:
      - kpi_grid:
          metrics_view: mv1
          measures:
            - x
  - items:
      - custom_chart:
          metrics_sql: "SELECT bar, y FROM mv2"
  - items:
      - custom_chart:
          metrics_sql:
            - "SELECT bar, y FROM mv2"
            - "SELECT qux, w FROM mv4"
  - items:
      - markdown:
          content: 'Total z: {{ metrics_sql "SELECT z FROM mv3" }}'
`,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 14, 0, 0)

	// Build claims with a transitive access rule on the canvas
	ctx := t.Context()
	claims := &runtime.SecurityClaims{
		AdditionalRules: []*runtimev1.SecurityRule{
			{
				Rule: &runtimev1.SecurityRule_TransitiveAccess{
					TransitiveAccess: &runtimev1.SecurityRuleTransitiveAccess{
						Resource: &runtimev1.ResourceName{
							Kind: runtime.ResourceKindCanvas,
							Name: "c1",
						},
					},
				},
			},
		},
	}

	// Resolve security for mv1 (referenced via metrics_view); should be accessible
	mv1 := testruntime.GetResource(t, rt, id, runtime.ResourceKindMetricsView, "mv1")
	sec, err := rt.ResolveSecurity(ctx, id, claims, mv1)
	require.NoError(t, err)
	require.True(t, sec.CanAccess())

	// Resolve security for mv2 (referenced via metrics_sql); should be accessible
	mv2 := testruntime.GetResource(t, rt, id, runtime.ResourceKindMetricsView, "mv2")
	sec, err = rt.ResolveSecurity(ctx, id, claims, mv2)
	require.NoError(t, err)
	require.True(t, sec.CanAccess())

	// Resolve security for mv3 (referenced via metrics_sql in markdown content); should be accessible
	mv3 := testruntime.GetResource(t, rt, id, runtime.ResourceKindMetricsView, "mv3")
	sec, err = rt.ResolveSecurity(ctx, id, claims, mv3)
	require.NoError(t, err)
	require.True(t, sec.CanAccess())

	// Resolve security for mv4 (referenced via a metrics_sql list in a multi-query custom chart); should be accessible
	mv4 := testruntime.GetResource(t, rt, id, runtime.ResourceKindMetricsView, "mv4")
	sec, err = rt.ResolveSecurity(ctx, id, claims, mv4)
	require.NoError(t, err)
	require.True(t, sec.CanAccess())
}

func TestCanvasComponentAccessAnnotationsAfterReparse(t *testing.T) {
	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "",
			// An admin-managed canvas, such as a personal canvas.
			"personal/p1.yaml": `
type: canvas
annotations:
  admin_managed: "true"
  admin_embed_owner_id: ext_owner
rows:
- items:
  - markdown:
      content: "Private"
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, id, 3, 0, 0)

	ctx := t.Context()
	ctrl, err := rt.Controller(ctx, id)
	require.NoError(t, err)
	name := &runtimev1.ResourceName{Kind: runtime.ResourceKindComponent, Name: "p1--component-0-0"}
	want := map[string]string{"admin_managed": "true", "admin_embed_owner_id": "ext_owner"}
	canAccess := func(userID string) bool {
		cmp, err := ctrl.Get(ctx, name, false)
		require.NoError(t, err)
		sec, err := rt.ResolveSecurity(ctx, id, &runtime.SecurityClaims{UserID: userID, UserAttributes: map[string]any{"embed": true}}, cmp)
		require.NoError(t, err)
		return sec.CanAccess()
	}

	// The inline component inherits the canvas's access annotations, so only the canvas's owner can access it.
	cmp, err := ctrl.Get(ctx, name, true)
	require.NoError(t, err)
	require.Equal(t, want, cmp.GetComponent().Spec.CanvasAnnotations)
	require.Equal(t, want, cmp.GetComponent().State.ValidSpec.CanvasAnnotations)
	require.True(t, canAccess("ext_owner"))
	require.False(t, canAccess("ext_other"))

	// Remove the annotations from the component, like a component that was parsed by a version that didn't add them.
	cmp.GetComponent().Spec.CanvasAnnotations = nil
	require.NoError(t, ctrl.UpdateSpec(ctx, name, cmp))
	require.NoError(t, ctrl.WaitUntilIdle(ctx, false))
	cmp, err = ctrl.Get(ctx, name, false)
	require.NoError(t, err)
	require.Empty(t, cmp.GetComponent().Spec.CanvasAnnotations)
	require.Empty(t, cmp.GetComponent().State.ValidSpec.CanvasAnnotations)

	// Parsing the project again, which also happens when the runtime restarts, adds them without a change to the canvas's file.
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 3, 0, 0)
	cmp, err = ctrl.Get(ctx, name, false)
	require.NoError(t, err)
	require.Equal(t, want, cmp.GetComponent().Spec.CanvasAnnotations)
	require.Equal(t, want, cmp.GetComponent().State.ValidSpec.CanvasAnnotations)
	require.True(t, canAccess("ext_owner"))
	require.False(t, canAccess("ext_other"))
}

func TestCanvasComponentAccessAnnotationsOnCatalogLoad(t *testing.T) {
	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{
			"rill.yaml": "",
			"c1.yaml": `
type: component
markdown:
  content: "Shared"
`,
			// A canvas that isn't admin-managed.
			"d1.yaml": `
type: canvas
rows:
- items:
  - markdown:
      content: "Public"
`,
			// Admin-managed canvases, such as personal canvases.
			"personal/p1.yaml": `
type: canvas
annotations:
  admin_managed: "true"
  admin_embed_owner_id: ext_owner
  admin_nonce: "2026-01-01T00:00:00Z"
rows:
- items:
  - markdown:
      content: "Private"
  - component: c1
- tabs:
  - label: Details
    rows:
    - items:
      - markdown:
          content: "Private tab"
`,
			"personal/p2.yaml": `
type: canvas
annotations:
  admin_managed: "true"
  admin_owner_user_id: u1
rows:
- items:
  - markdown:
      content: "Private"
`,
		},
	})
	testruntime.RequireReconcileState(t, rt, id, 9, 0, 0)

	ctx := t.Context()
	ctrl, err := rt.Controller(ctx, id)
	require.NoError(t, err)
	component := func(c *runtime.Controller, name string) *runtimev1.Resource {
		cmp, err := c.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindComponent, Name: name}, true)
		require.NoError(t, err)
		return cmp
	}

	// Remove the annotations from the spec and valid spec of p1's components and persist them, like a catalog written by a version that didn't add them.
	// The component of p2 keeps its annotations.
	for _, name := range []string{"p1--component-0-0", "p1--component-g1-t0-0-0"} {
		cmp := component(ctrl, name)
		cmp.GetComponent().Spec.CanvasAnnotations = nil
		require.NoError(t, ctrl.UpdateSpec(ctx, cmp.Meta.Name, cmp))
	}
	require.NoError(t, ctrl.WaitUntilIdle(ctx, false))
	for _, name := range []string{"p1--component-0-0", "p1--component-g1-t0-0-0"} {
		cmp := component(ctrl, name)
		require.Empty(t, cmp.GetComponent().Spec.CanvasAnnotations)
		require.Empty(t, cmp.GetComponent().State.ValidSpec.CanvasAnnotations)
	}
	require.NoError(t, ctrl.Flush(ctx))

	// Load the catalog in a controller that isn't run, so nothing is parsed or reconciled.
	// That's the state in which a restarted runtime starts serving requests.
	loaded, err := runtime.NewController(ctx, rt, id, zap.NewNop(), activity.NewNoopClient())
	require.NoError(t, err)
	t.Cleanup(func() {
		// Running the controller with a cancelled context closes it.
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_ = loaded.Run(ctx)
	})
	canAccess := func(name string, claims *runtime.SecurityClaims) bool {
		sec, err := rt.ResolveSecurity(ctx, id, claims, component(loaded, name))
		require.NoError(t, err)
		return sec.CanAccess()
	}
	owner := &runtime.SecurityClaims{UserID: "ext_owner", UserAttributes: map[string]any{"embed": true}}
	other := &runtime.SecurityClaims{UserID: "ext_other", UserAttributes: map[string]any{"embed": true}}

	// The components of the admin-managed canvas were loaded with its access annotations, without its nonce, so only its owner can access them.
	want := map[string]string{"admin_managed": "true", "admin_embed_owner_id": "ext_owner"}
	for _, name := range []string{"p1--component-0-0", "p1--component-g1-t0-0-0"} {
		cmp := component(loaded, name)
		require.Equal(t, want, cmp.GetComponent().Spec.CanvasAnnotations, name)
		require.Equal(t, want, cmp.GetComponent().State.ValidSpec.CanvasAnnotations, name)
		require.True(t, canAccess(name, owner), name)
		require.False(t, canAccess(name, other), name)
	}

	// The other components were loaded as they were persisted:
	// the component defined in its own file, the inline component of the canvas that isn't admin-managed, and the component that had its annotations.
	for _, name := range []string{"c1", "d1--component-0-0"} {
		cmp := component(loaded, name)
		require.Empty(t, cmp.GetComponent().Spec.CanvasAnnotations, name)
		require.Empty(t, cmp.GetComponent().State.ValidSpec.CanvasAnnotations, name)
		require.True(t, canAccess(name, other), name)
	}
	require.Equal(t, map[string]string{"admin_managed": "true", "admin_owner_user_id": "u1"}, component(loaded, "p2--component-0-0").GetComponent().Spec.CanvasAnnotations)
	require.False(t, canAccess("p2--component-0-0", other))
}

func TestCanvasComponentAccessAnnotationsOnCatalogLoadOfLegacyStates(t *testing.T) {
	private := map[string]string{"admin_managed": "true", "admin_owner_user_id": "u1"}
	shared := map[string]string{"admin_managed": "true", "admin_owner_user_id": "u1", "admin_shared": "true"}
	embedOwned := map[string]string{"admin_managed": "true", "admin_embed_owner_id": "ext_owner"}
	owner := &runtime.SecurityClaims{UserID: "u1", UserAttributes: map[string]any{"id": "u1", "admin": false}}
	other := &runtime.SecurityClaims{UserID: "u2", UserAttributes: map[string]any{"id": "u2", "admin": false}}
	embedOwner := &runtime.SecurityClaims{UserID: "ext_owner", UserAttributes: map[string]any{"embed": true}}

	// canvasSpec returns a canvas spec with the given annotations and one item per component; the names of inline components end in a position.
	canvasSpec := func(annotations map[string]string, components ...string) *runtimev1.CanvasSpec {
		row := &runtimev1.CanvasRow{}
		for _, name := range components {
			row.Items = append(row.Items, &runtimev1.CanvasItem{Component: name, DefinedInCanvas: strings.Contains(name, "--component-")})
		}
		return &runtimev1.CanvasSpec{Annotations: annotations, Rows: []*runtimev1.CanvasRow{row}}
	}
	componentSpec := func(content string, inline bool) *runtimev1.ComponentSpec {
		props, err := structpb.NewStruct(map[string]any{"content": content})
		require.NoError(t, err)
		return &runtimev1.ComponentSpec{Renderer: "markdown", RendererProperties: props, DefinedInCanvas: inline}
	}
	meta := func(kind, name, path string) *runtimev1.ResourceMeta {
		return &runtimev1.ResourceMeta{
			Name:           &runtimev1.ResourceName{Kind: kind, Name: name},
			Owner:          runtime.GlobalProjectParserName,
			FilePaths:      []string{path},
			CreatedOn:      timestamppb.Now(),
			SpecUpdatedOn:  timestamppb.Now(),
			StateUpdatedOn: timestamppb.Now(),
		}
	}
	canvas := func(name, path string, spec, validSpec *runtimev1.CanvasSpec) *runtimev1.Resource {
		return &runtimev1.Resource{
			Meta:     meta(runtime.ResourceKindCanvas, name, path),
			Resource: &runtimev1.Resource_Canvas{Canvas: &runtimev1.Canvas{Spec: spec, State: &runtimev1.CanvasState{ValidSpec: validSpec}}},
		}
	}
	deleted := func(r *runtimev1.Resource) *runtimev1.Resource {
		r.Meta.DeletedOn = timestamppb.Now()
		return r
	}
	// component returns a component with the given contents in its spec and valid spec; an empty valid content omits the valid spec.
	component := func(name, path, content, validContent string) *runtimev1.Resource {
		inline := strings.Contains(name, "--component-")
		state := &runtimev1.ComponentState{}
		if validContent != "" {
			state.ValidSpec = componentSpec(validContent, inline)
		}
		return &runtimev1.Resource{
			Meta:     meta(runtime.ResourceKindComponent, name, path),
			Resource: &runtimev1.Resource_Component{Component: &runtimev1.Component{Spec: componentSpec(content, inline), State: state}},
		}
	}

	// want describes a component after the catalog is loaded:
	// the annotations in its spec and valid spec, and whether someone other than the canvas's owner can access it.
	type want struct {
		spec, validSpec map[string]string
		other           bool
	}
	tests := []struct {
		name      string
		resources []*runtimev1.Resource
		owner     *runtime.SecurityClaims
		want      map[string]want
	}{
		{
			// The owner shared the canvas and edited it, and the catalog was stored before the canvas and its component were reconciled.
			// The component's valid spec is from before the canvas was shared.
			name: "canvas shared in its spec but not in its valid spec",
			resources: []*runtimev1.Resource{
				canvas("p1", "/personal/p1.yaml", canvasSpec(shared, "p1--component-0-0"), canvasSpec(private, "p1--component-0-0")),
				component("p1--component-0-0", "/personal/p1.yaml", "Shared", "Private"),
			},
			want: map[string]want{"p1--component-0-0": {spec: shared, validSpec: private, other: false}},
		},
		{
			name: "canvas shared in its valid spec but not in its spec",
			resources: []*runtimev1.Resource{
				canvas("p1", "/personal/p1.yaml", canvasSpec(private, "p1--component-0-0"), canvasSpec(shared, "p1--component-0-0")),
				component("p1--component-0-0", "/personal/p1.yaml", "Private", "Shared"),
			},
			want: map[string]want{"p1--component-0-0": {spec: private, validSpec: shared, other: false}},
		},
		{
			// There's no valid version of the canvas to tell under which annotations the component's valid spec was defined.
			name: "shared canvas without a valid spec",
			resources: []*runtimev1.Resource{
				canvas("p1", "/personal/p1.yaml", canvasSpec(shared, "p1--component-0-0", "p1--component-0-1"), nil),
				component("p1--component-0-0", "/personal/p1.yaml", "Shared", "Private"),
				component("p1--component-0-1", "/personal/p1.yaml", "Shared", ""),
			},
			want: map[string]want{
				"p1--component-0-0": {spec: shared, validSpec: private, other: false},
				"p1--component-0-1": {spec: shared, other: true},
			},
		},
		{
			// The canvas's file was moved, and only the canvas was stored with its new path.
			name: "canvas and component stored with different paths",
			resources: []*runtimev1.Resource{
				canvas("p1", "/archive/p1.yaml", canvasSpec(embedOwned, "p1--component-0-0"), canvasSpec(embedOwned, "p1--component-0-0")),
				component("p1--component-0-0", "/personal/p1.yaml", "Private", "Private"),
			},
			owner: embedOwner,
			want:  map[string]want{"p1--component-0-0": {spec: embedOwned, validSpec: embedOwned, other: false}},
		},
		{
			// A file that defined an admin-managed canvas now defines a component, and the canvas and its inline component are still being deleted.
			name: "component defined in the file of a deleted canvas",
			resources: []*runtimev1.Resource{
				deleted(canvas("x", "/x.yaml", canvasSpec(private, "x--component-0-0"), canvasSpec(private, "x--component-0-0"))),
				component("x--component-0-0", "/x.yaml", "Private", "Private"),
				component("x", "/x.yaml", "Public", "Public"),
			},
			want: map[string]want{
				"x--component-0-0": {spec: private, validSpec: private, other: false},
				"x":                {other: true},
			},
		},
		{
			// The canvas's second item was removed from its spec; its component is listed by the canvas's valid spec only.
			// The third item was removed earlier, and its component was stored although no version of the canvas lists it any more.
			name: "components the canvas no longer lists",
			resources: []*runtimev1.Resource{
				canvas("p1", "/personal/p1.yaml", canvasSpec(shared, "p1--component-0-0"), canvasSpec(shared, "p1--component-0-0", "p1--component-0-1")),
				component("p1--component-0-0", "/personal/p1.yaml", "Shared", "Shared"),
				component("p1--component-0-1", "/personal/p1.yaml", "Removed", "Removed"),
				component("p1--component-0-2", "/personal/p1.yaml", "Removed earlier", "Removed earlier"),
			},
			want: map[string]want{
				"p1--component-0-0": {spec: shared, validSpec: shared, other: true},
				"p1--component-0-1": {spec: private, validSpec: shared, other: false},
				"p1--component-0-2": {spec: private, validSpec: private, other: false},
			},
		},
		{
			// A canvas that isn't admin-managed, with a component that no version of it lists any more,
			// and an admin-managed canvas that references a component defined in its own file.
			name: "ordinary canvases and components defined in their own file",
			resources: []*runtimev1.Resource{
				canvas("d1", "/d1.yaml", canvasSpec(nil, "d1--component-0-0", "c1"), canvasSpec(nil, "d1--component-0-0", "c1")),
				component("d1--component-0-0", "/d1.yaml", "Public", "Public"),
				component("d1--component-0-1", "/d1.yaml", "Removed", "Removed"),
				component("c1", "/c1.yaml", "Public", "Public"),
				canvas("p1", "/personal/p1.yaml", canvasSpec(private, "p1--component-0-0", "c1"), canvasSpec(private, "p1--component-0-0", "c1")),
				component("p1--component-0-0", "/personal/p1.yaml", "Private", ""),
			},
			want: map[string]want{
				"d1--component-0-0": {other: true},
				"d1--component-0-1": {other: true},
				"c1":                {other: true},
				"p1--component-0-0": {spec: private, other: false},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{Files: map[string]string{"rill.yaml": ""}})
			testruntime.RequireReconcileState(t, rt, id, 1, 0, 0)
			ctx := t.Context()

			// Store the resources in the catalog, like a version that didn't add the annotations.
			// The instance's controller has nothing left to write, so taking over its catalog doesn't interfere with it.
			ctrl, err := rt.Controller(ctx, id)
			require.NoError(t, err)
			require.NoError(t, ctrl.Flush(ctx))
			store, release, err := rt.Catalog(ctx, id)
			require.NoError(t, err)
			defer release()
			version, err := store.NextControllerVersion(ctx)
			require.NoError(t, err)
			for _, r := range tt.resources {
				data, err := proto.Marshal(r)
				require.NoError(t, err)
				require.NoError(t, store.CreateResource(ctx, version, drivers.Resource{Kind: r.Meta.Name.Kind, Name: r.Meta.Name.Name, Data: data}))
			}

			// Load the catalog in a controller that isn't run, so nothing is parsed or reconciled.
			loaded, err := runtime.NewController(ctx, rt, id, zap.NewNop(), activity.NewNoopClient())
			require.NoError(t, err)
			t.Cleanup(func() {
				// Running the controller with a cancelled context closes it.
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				_ = loaded.Run(ctx)
			})
			canAccess := func(r *runtimev1.Resource, claims *runtime.SecurityClaims) bool {
				sec, err := rt.ResolveSecurity(ctx, id, claims, r)
				require.NoError(t, err)
				return sec.CanAccess()
			}

			if tt.owner == nil {
				tt.owner = owner
			}
			for name, want := range tt.want {
				cmp, err := loaded.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindComponent, Name: name}, false)
				require.NoError(t, err)
				if len(want.spec) == 0 {
					require.Empty(t, cmp.GetComponent().Spec.CanvasAnnotations, name)
				} else {
					require.Equal(t, want.spec, cmp.GetComponent().Spec.CanvasAnnotations, name)
				}
				if len(want.validSpec) == 0 {
					require.Empty(t, cmp.GetComponent().State.ValidSpec.GetCanvasAnnotations(), name)
				} else {
					require.Equal(t, want.validSpec, cmp.GetComponent().State.ValidSpec.GetCanvasAnnotations(), name)
				}
				require.True(t, canAccess(cmp, tt.owner), name)
				require.Equal(t, want.other, canAccess(cmp, other), name)
			}

			// A component is never accessible to someone who can't access any of the canvases that list it as an inline component.
			canvases, err := loaded.List(ctx, runtime.ResourceKindCanvas, "", false)
			require.NoError(t, err)
			for _, c := range canvases {
				if canAccess(c, other) {
					continue
				}
				for _, spec := range []*runtimev1.CanvasSpec{c.GetCanvas().Spec, c.GetCanvas().State.ValidSpec} {
					var items []*runtimev1.CanvasItem
					runtime.CollectCanvasItems(spec.GetRows(), &items)
					for _, item := range items {
						if !item.DefinedInCanvas {
							continue
						}
						cmp, err := loaded.Get(ctx, &runtimev1.ResourceName{Kind: runtime.ResourceKindComponent, Name: item.Component}, false)
						require.NoError(t, err)
						require.False(t, canAccess(cmp, other), "%s is accessible, but canvas %s isn't", item.Component, c.Meta.Name.Name)
					}
				}
			}
		})
	}
}

func TestCanvasItemParamBindings(t *testing.T) {
	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files: map[string]string{"rill.yaml": ""},
	})

	// Create a model, a metrics view, a parameterized component, and a canvas binding values to its params.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"m1.sql": `SELECT '2025-01-01T00:00:00Z'::TIMESTAMP AS ts, 'foo' as foo, 1 as x`,
		"mv1.yaml": `
version: 1
type: metrics_view
model: m1
timeseries: ts
dimensions:
- column: foo
measures:
- name: x
  expression: sum(x)
`,
		"trend.yaml": `
type: component
params:
  - name: metrics_view
    type: metrics_view
    required: true
  - name: measure
    type: measure
    required: true
  - name: time_dim
    type: time_dimension
    required: true
  - name: limit
    type: number
    default: 500
custom_chart:
  metrics_sql: SELECT {{ .params.time_dim }} AS ts, {{ .params.measure }} AS value FROM {{ .params.metrics_view }} LIMIT {{ .params.limit }}
  spec:
    chartType: Line Chart
    encodings:
      x: {field: ts}
      y: {field: value}
`,
		"c1.yaml": `
type: canvas
rows:
  - items:
      - component: trend
        params:
          metrics_view: mv1
          measure: x
          time_dim: ts
`,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 5, 0, 0)

	// The canvas should be valid and have a parser-added ref to the param-bound metrics view.
	c1 := testruntime.GetResource(t, rt, id, runtime.ResourceKindCanvas, "c1")
	require.NotNil(t, c1.GetCanvas().State.ValidSpec)
	var hasMVRef bool
	for _, ref := range c1.Meta.Refs {
		if ref.Kind == runtime.ResourceKindMetricsView && ref.Name == "mv1" {
			hasMVRef = true
		}
	}
	require.True(t, hasMVRef, "expected canvas to have a ref to the param-bound metrics view")

	// The param-bound metrics view should be accessible through transitive access on the canvas.
	claims := &runtime.SecurityClaims{
		AdditionalRules: []*runtimev1.SecurityRule{
			{
				Rule: &runtimev1.SecurityRule_TransitiveAccess{
					TransitiveAccess: &runtimev1.SecurityRuleTransitiveAccess{
						Resource: &runtimev1.ResourceName{Kind: runtime.ResourceKindCanvas, Name: "c1"},
					},
				},
			},
		},
	}
	mv1 := testruntime.GetResource(t, rt, id, runtime.ResourceKindMetricsView, "mv1")
	sec, err := rt.ResolveSecurity(t.Context(), id, claims, mv1)
	require.NoError(t, err)
	require.True(t, sec.CanAccess())

	// Invalid: a bound measure that doesn't exist in the metrics view.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: canvas
rows:
  - items:
      - component: trend
        params:
          metrics_view: mv1
          measure: nonexistent
          time_dim: ts
`,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 5, 1, 0)
	testruntime.RequireReconcileErrorContains(t, rt, id, runtime.ResourceKindCanvas, "c1", `not a measure in metrics view "mv1"`)

	// Invalid: a missing required param.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: canvas
rows:
  - items:
      - component: trend
        params:
          metrics_view: mv1
          time_dim: ts
`,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 5, 1, 0)
	testruntime.RequireReconcileErrorContains(t, rt, id, runtime.ResourceKindCanvas, "c1", `missing value for required param "measure"`)

	// Invalid: passing params to a component that doesn't declare any.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"plain.yaml": `
type: component
kpi:
  metrics_view: mv1
  measure: x
`,
		"c1.yaml": `
type: canvas
rows:
  - items:
      - component: plain
        params:
          foo: 1
`,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 6, 1, 0)
	testruntime.RequireReconcileErrorContains(t, rt, id, runtime.ResourceKindCanvas, "c1", "does not declare any params")

	// Valid again: fix the canvas and check it recovers.
	testruntime.PutFiles(t, rt, id, map[string]string{
		"c1.yaml": `
type: canvas
rows:
  - items:
      - component: trend
        params:
          metrics_view: mv1
          measure: x
          time_dim: ts
          limit: 100
      - component: plain
`,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 6, 0, 0)
}

func TestCanvasValidateDefaultFilters(t *testing.T) {
	// Create an instance with StageChanges==true
	rt, id := testruntime.NewInstanceWithOptions(t, testruntime.InstanceOptions{
		Files:        map[string]string{"rill.yaml": ""},
		StageChanges: true,
	})

	// Create two metrics views, but a canvas that only references mv1 and sets a default filter on it
	testruntime.PutFiles(t, rt, id, map[string]string{
		"m1.sql": `SELECT 'foo' as foo, 1 as x`,
		"m2.sql": `SELECT 'bar' as bar, 2 as y`,
		"mv1.yaml": `
version: 1
type: metrics_view
model: m1
dimensions:
- column: foo
measures:
- name: x
  expression: sum(x)
`,
		"mv2.yaml": `
version: 1
type: metrics_view
model: m2
dimensions:
- column: bar
measures:
- name: y
  expression: sum(y)
`,
		"c1.yaml": `
type: canvas
defaults:
  filters:
    mv1: "foo = 'foo'"
rows:
  - items:
      - kpi_grid:
          metrics_view: mv1
          measures:
            - x
`,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 7, 0, 0)
	c1 := testruntime.GetResource(t, rt, id, runtime.ResourceKindCanvas, "c1")
	require.NotNil(t, c1.GetCanvas().State.ValidSpec)
	require.Empty(t, c1.Meta.ReconcileError)

	// Add a default filter for a metrics view that exists, but is not referenced by any component
	testruntime.PutFiles(t, rt, id, map[string]string{"c1.yaml": `
type: canvas
defaults:
  filters:
    mv1: "foo = 'foo'"
    mv2: "bar = 'bar'"
rows:
  - items:
      - kpi_grid:
          metrics_view: mv1
          measures:
            - x
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	c1 = testruntime.GetResource(t, rt, id, runtime.ResourceKindCanvas, "c1")
	require.Contains(t, c1.Meta.ReconcileError, `metrics view "mv2" referenced in default filter does not exist`)

	// With StageChanges==true, the valid spec should be preserved
	require.NotNil(t, c1.GetCanvas().State.ValidSpec)

	// Fix it by adding a component that references mv2
	testruntime.PutFiles(t, rt, id, map[string]string{"c1.yaml": `
type: canvas
defaults:
  filters:
    mv1: "foo = 'foo'"
    mv2: "bar = 'bar'"
rows:
  - items:
      - kpi_grid:
          metrics_view: mv1
          measures:
            - x
  - items:
      - kpi_grid:
          metrics_view: mv2
          measures:
            - y
`})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 8, 0, 0)
	c1 = testruntime.GetResource(t, rt, id, runtime.ResourceKindCanvas, "c1")
	require.NotNil(t, c1.GetCanvas().State.ValidSpec)
	require.Empty(t, c1.Meta.ReconcileError)
}

func TestCanvasAIPrompts(t *testing.T) {
	rt, id := testruntime.NewInstance(t)
	testruntime.PutFiles(t, rt, id, map[string]string{
		"m1.sql": `SELECT 'foo' as foo, 1 as x`,
		"mv1.yaml": `
version: 1
type: metrics_view
model: m1
dimensions:
- column: foo
measures:
- name: x
  expression: sum(x)
`,
		"c1.yaml": `
type: canvas
ai_prompts:
  - label: Total
    prompt: What is the total x?
rows:
  - items:
      - kpi_grid:
          metrics_view: mv1
          measures:
            - x
`,
	})
	testruntime.ReconcileParserAndWait(t, rt, id)
	testruntime.RequireReconcileState(t, rt, id, 5, 0, 0)
	c1 := testruntime.GetResource(t, rt, id, runtime.ResourceKindCanvas, "c1").GetCanvas()
	require.NotNil(t, c1.State.ValidSpec)
	require.Len(t, c1.State.ValidSpec.AiPrompts, 1)
	require.Equal(t, "Total", c1.State.ValidSpec.AiPrompts[0].Label)
	require.Equal(t, "What is the total x?", c1.State.ValidSpec.AiPrompts[0].Prompt)
}
