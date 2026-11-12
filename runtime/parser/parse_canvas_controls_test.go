package parser

import (
	"context"
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func parseControlsYAML(t *testing.T, src string, onItem bool) ([]*runtimev1.CanvasControl, error) {
	var controls []*canvasControlYAML
	require.NoError(t, yaml.Unmarshal([]byte(src), &controls))
	return parseCanvasControls(controls, onItem)
}

func TestParseCanvasControls(t *testing.T) {
	got, err := parseControlsYAML(t, `
- measure: spend
  label: Metric
  options: [spend, impressions]
- dimension: device
  options: [device, country]
  display: buttons
- slot: row_dimensions
  multiple: true
  options: [account, brand]
- filter: publisher
  multiple: true
  default: [Google, Meta]
  fields: { bids_v2: pub }
- time: range
  options: [P7D, P30D]
- time: grain
  options: [day, week]
`, true)
	require.NoError(t, err)
	require.Len(t, got, 6)
	require.Equal(t, "measure", got[0].Kind)
	require.Equal(t, "spend", got[0].Name)
	require.Equal(t, "Metric", got[0].Label)
	require.Equal(t, "row_dimensions", got[2].Name)
	require.Equal(t, []string{"Google", "Meta"}, got[3].Default)
	require.Equal(t, "pub", got[3].Fields["bids_v2"])
	require.Equal(t, "time_range", got[4].Name)
	require.Equal(t, "time_grain", got[5].Kind)
}

func TestParseCanvasControlsErrors(t *testing.T) {
	cases := []struct {
		name, yaml, err string
		onItem          bool
	}{
		{"no target", `[{label: x}]`, "exactly one of", true},
		{"two targets", `[{measure: a, filter: b, options: [a]}]`, "exactly one of", true},
		{"slot off item", `[{slot: measures, options: [a]}]`, "only be declared on an item", false},
		{"bad time", `[{time: zone, options: [a]}]`, "invalid 'time' value", true},
		{"swap without options", `[{measure: a}]`, "'options' must list", true},
		{"swap options miss target", `[{measure: a, options: [b, c]}]`, `must include "a"`, true},
		{"swap multiple", `[{measure: a, options: [a, b], multiple: true}]`, "'multiple' is not supported", true},
		{"bad range", `[{time: range, options: [nonsense]}]`, "invalid time range", true},
		{"bad grain", `[{time: grain, options: [fortnight]}]`, "invalid time grain", true},
		{"several defaults", `[{filter: a, default: [x, y]}]`, "'multiple' is not set", true},
		{"default not an option", `[{filter: a, options: [x], default: y}]`, "is not one of the 'options'", true},
		{"duplicate names", `[{filter: a}, {filter: a}]`, "duplicate control name", true},
		{"bad name", `[{slot: x.field, name: "x field", options: [a]}]`, "invalid control name", true},
		{"bad display", `[{filter: a, display: slider}]`, "invalid 'display'", true},
		{"fields on swap", `[{measure: a, options: [a], fields: {m: b}}]`, "'fields' is only supported", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parseControlsYAML(t, tc.yaml, tc.onItem)
			require.ErrorContains(t, err, tc.err)
		})
	}
}

func TestValidateCanvasControlScopes(t *testing.T) {
	spec := &runtimev1.CanvasSpec{
		Controls: []*runtimev1.CanvasControl{{Name: "metric"}},
		Rows: []*runtimev1.CanvasRow{
			{Items: []*runtimev1.CanvasItem{{Name: "chart"}, {}}},
			{Container: &runtimev1.CanvasContainer{Name: "section", Rows: []*runtimev1.CanvasRow{{Items: []*runtimev1.CanvasItem{{Name: "other"}}}}}},
		},
	}
	require.NoError(t, validateCanvasControlScopes(spec))

	spec.Rows[0].Items[1].Name = "section"
	require.ErrorContains(t, validateCanvasControlScopes(spec), `the name "section"`)

	spec.Rows[0].Items[1].Name = "metric"
	require.ErrorContains(t, validateCanvasControlScopes(spec), `the name "metric"`)
}

func TestCanvasControlsInYAML(t *testing.T) {
	ctx := context.Background()
	repo := makeRepo(t, map[string]string{
		`rill.yaml`: ``,
		`canvases/d1.yaml`: `
type: canvas
controls:
- measure: spend
  label: Metric
  options: [spend, impressions]
rows:
- items:
  - name: intro
    ignore_controls: true
    markdown:
      content: "Intro"
  - name: chart
    ignore_controls: [metric]
    controls:
    - filter: publisher
      multiple: true
    markdown:
      content: "Chart"
- name: section
  controls:
  - time: range
    options: [P7D, P30D]
  rows:
  - items:
    - markdown:
        content: "In section"
- name: platforms
  controls:
  - time: grain
    options: [day, week]
  tabs:
  - label: Mobile
    controls:
    - filter: os
    rows:
    - items:
      - markdown:
          content: "Mobile"
`,
	})

	p, err := Parse(ctx, repo, "", "", "duckdb", true)
	require.NoError(t, err)
	require.Empty(t, p.Errors)

	spec := p.Resources[ResourceName{Kind: ResourceKindCanvas, Name: "d1"}].CanvasSpec
	require.Len(t, spec.Controls, 1)
	require.Equal(t, "spend", spec.Controls[0].Name)

	intro, chart := spec.Rows[0].Items[0], spec.Rows[0].Items[1]
	require.Equal(t, "intro", intro.Name)
	require.True(t, intro.IgnoreAllControls)
	require.Equal(t, []string{"metric"}, chart.IgnoreControls)
	require.Equal(t, "filter", chart.Controls[0].Kind)
	// The item-level keys are not part of the inline component definition.
	for _, r := range p.Resources {
		if r.ComponentSpec != nil {
			require.Equal(t, "markdown", r.ComponentSpec.Renderer)
		}
	}

	require.Equal(t, "time_range", spec.Rows[1].Container.Controls[0].Name)
	group := spec.Rows[2].TabGroup
	require.Equal(t, "time_grain", group.Controls[0].Name)
	require.Equal(t, "os", group.Tabs[0].Controls[0].Name)
}

func TestCanvasControlsInYAMLErrors(t *testing.T) {
	ctx := context.Background()
	cases := map[string]string{
		"needs a 'name'": `
type: canvas
rows:
- items:
  - controls:
    - filter: publisher
    markdown:
      content: "x"
`,
		"cannot declare 'controls'": `
type: canvas
rows:
- controls:
  - filter: publisher
  items:
  - markdown:
      content: "x"
`,
		`the name "section"`: `
type: canvas
rows:
- items:
  - name: section
    markdown:
      content: "x"
- name: section
  rows: []
`,
	}
	for want, src := range cases {
		repo := makeRepo(t, map[string]string{`rill.yaml`: ``, `canvases/d1.yaml`: src})
		p, err := Parse(ctx, repo, "", "", "duckdb", true)
		require.NoError(t, err)
		require.Len(t, p.Errors, 1)
		require.Contains(t, p.Errors[0].Message, want)
	}
}
