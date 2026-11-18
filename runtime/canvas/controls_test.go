package canvas

import (
	"testing"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

func control(kind, target, name string) *runtimev1.CanvasControl {
	return &runtimev1.CanvasControl{Kind: kind, Target: target, Name: name}
}

func names(controls []*runtimev1.CanvasControl) []string {
	res := make([]string, len(controls))
	for i, c := range controls {
		res[i] = c.Name
	}
	return res
}

func TestResolveItemControls(t *testing.T) {
	own := &runtimev1.CanvasItem{Name: "own", Controls: []*runtimev1.CanvasControl{control(ControlKindFilter, "publisher", "item_publisher")}}
	plain := &runtimev1.CanvasItem{Name: "plain"}
	ignoreAll := &runtimev1.CanvasItem{Name: "ignore_all", IgnoreAllControls: true, Controls: []*runtimev1.CanvasControl{control(ControlKindTimeGrain, "", "time_grain")}}
	ignoreOne := &runtimev1.CanvasItem{Name: "ignore_one", IgnoreControls: []string{"metric"}}
	inTab := &runtimev1.CanvasItem{Name: "in_tab"}
	top := &runtimev1.CanvasItem{Name: "top"}

	spec := &runtimev1.CanvasSpec{
		Controls: []*runtimev1.CanvasControl{control(ControlKindMeasure, "spend", "metric")},
		Rows: []*runtimev1.CanvasRow{
			{Items: []*runtimev1.CanvasItem{top}},
			{Container: &runtimev1.CanvasContainer{
				Name:     "section",
				Controls: []*runtimev1.CanvasControl{control(ControlKindFilter, "publisher", "publisher")},
				Rows:     []*runtimev1.CanvasRow{{Items: []*runtimev1.CanvasItem{own, plain, ignoreAll, ignoreOne}}},
			}},
			{TabGroup: &runtimev1.CanvasTabGroup{
				Name:     "group",
				Controls: []*runtimev1.CanvasControl{control(ControlKindTimeGrain, "", "time_grain")},
				Tabs: []*runtimev1.CanvasTab{{
					Name:     "mobile",
					Controls: []*runtimev1.CanvasControl{control(ControlKindFilter, "os", "os")},
					Rows:     []*runtimev1.CanvasRow{{Items: []*runtimev1.CanvasItem{inTab}}},
				}},
			}},
		},
	}

	got := make(map[string][]string)
	for _, ic := range ResolveItemControls(spec) {
		got[ic.Item.Name] = names(ic.Controls)
	}
	require.Equal(t, map[string][]string{
		"top":        {"metric"},
		"own":        {"item_publisher", "metric"}, // The item's own publisher filter wins over the container's.
		"plain":      {"publisher", "metric"},
		"ignore_all": {"time_grain"}, // Keeps its own controls only.
		"ignore_one": {"publisher"},
		"in_tab":     {"os", "time_grain", "metric"},
	}, got)
}

func TestSwapField(t *testing.T) {
	props := map[string]any{
		"metrics_view": "spend",
		"title":        "spend",
		"x":            map[string]any{"field": "device"},
		"y":            map[string]any{"field": "spend", "fields": []any{"spend", "clicks"}},
		"columns":      []any{"device", map[string]any{"name": "spend", "width": 10.0}},
		"sort_by":      "spend",
	}
	got, used := SwapField(props, "spend", "impressions")
	require.True(t, used)
	require.Equal(t, map[string]any{
		"metrics_view": "spend",
		"title":        "spend",
		"x":            map[string]any{"field": "device"},
		"y":            map[string]any{"field": "impressions", "fields": []any{"impressions", "clicks"}},
		"columns":      []any{"device", map[string]any{"name": "impressions", "width": 10.0}},
		"sort_by":      "impressions",
	}, got)
	// The input is not modified.
	require.Equal(t, "spend", props["sort_by"])

	_, used = SwapField(props, "missing", "impressions")
	require.False(t, used)
}

func TestValidateControls(t *testing.T) {
	newProps := func(m map[string]any) *structpb.Struct {
		pb, err := structpb.NewStruct(m)
		require.NoError(t, err)
		return pb
	}
	components := map[string]*runtimev1.ComponentSpec{
		"chart": {Renderer: "bar_chart", RendererProperties: newProps(map[string]any{
			"metrics_view": "bids",
			"x":            map[string]any{"field": "device", "type": "nominal"},
			"y":            map[string]any{"field": "spend", "type": "quantitative"},
		})},
		"table": {Renderer: "table", RendererProperties: newProps(map[string]any{"metrics_view": "bids", "columns": []any{"device", "spend", "impressions"}})},
		"kpis":  {Renderer: "kpi_grid", RendererProperties: newProps(map[string]any{"metrics_view": "bids", "measures": []any{"spend", "impressions"}})},
	}
	mvs := map[string]*runtimev1.MetricsViewSpec{"bids": {
		Dimensions: []*runtimev1.MetricsViewSpec_Dimension{{Name: "device"}, {Name: "country"}},
		Measures:   []*runtimev1.MetricsViewSpec_Measure{{Name: "spend"}, {Name: "impressions"}, {Name: "clicks"}},
	}}
	measure := func(target string, options ...string) *runtimev1.CanvasControl {
		return &runtimev1.CanvasControl{Kind: ControlKindMeasure, Target: target, Name: target, Options: options}
	}

	for _, tt := range []struct {
		name      string
		component string
		control   *runtimev1.CanvasControl
		ignore    []string
		err       string
	}{
		{name: "valid swap", component: "chart", control: measure("spend", "spend", "impressions")},
		{name: "option missing from the metrics view", component: "chart", control: measure("spend", "spend", "win_rate"), err: `option "win_rate" doesn't work`},
		{name: "unused field with a missing option", component: "chart", control: measure("clicks", "clicks", "win_rate"), err: `option "win_rate" is not a measure of any metrics view`},
		{name: "unused field with existing options", component: "chart", control: measure("clicks", "clicks", "impressions")},
		{name: "swap repeats a table column", component: "table", control: measure("spend", "spend", "impressions"), err: `option "impressions" is already shown in component "table"`},
		{name: "swap repeats a kpi", component: "kpis", control: measure("spend", "spend", "impressions"), err: `option "impressions" is already shown in component "kpis"`},
		{name: "swap to a new kpi", component: "kpis", control: measure("spend", "spend", "clicks")},
		{name: "slot", component: "chart", control: &runtimev1.CanvasControl{Kind: ControlKindSlot, Target: "x.field", Name: "x_field", Options: []string{"device", "country"}}},
		{name: "slot with a missing option", component: "chart", control: &runtimev1.CanvasControl{Kind: ControlKindSlot, Target: "x.field", Name: "x_field", Options: []string{"device", "nope"}}, err: `option "nope"`},
		{name: "slot the component doesn't set", component: "chart", control: &runtimev1.CanvasControl{Kind: ControlKindSlot, Target: "color.field", Name: "color_field", Options: []string{"device"}}, err: `component "chart" doesn't set "color.field"`},
		{name: "filter", component: "chart", control: &runtimev1.CanvasControl{Kind: ControlKindFilter, Target: "country", Name: "country"}},
		{name: "filter on a missing dimension", component: "chart", control: &runtimev1.CanvasControl{Kind: ControlKindFilter, Target: "region", Name: "region"}, err: `filters on "region", which is not a dimension`},
		{name: "ignored control", component: "chart", ignore: []string{"canvas_metric"}},
		{name: "ignored control that doesn't exist", component: "chart", ignore: []string{"canvas_metrc"}, err: `ignores control "canvas_metrc"`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			// A valid canvas control stays in the spec whatever happens to the item's control.
			canvasControl := &runtimev1.CanvasControl{Kind: ControlKindTimeGrain, Name: "canvas_metric", Options: []string{"day"}}
			item := &runtimev1.CanvasItem{Component: tt.component, Name: tt.component, IgnoreControls: tt.ignore}
			if tt.control != nil {
				item.Controls = []*runtimev1.CanvasControl{tt.control}
			}
			spec := &runtimev1.CanvasSpec{Controls: []*runtimev1.CanvasControl{canvasControl}, Rows: []*runtimev1.CanvasRow{{Items: []*runtimev1.CanvasItem{item}}}}

			valid, err := ValidateControls(spec, components, mvs)
			if tt.err == "" {
				require.NoError(t, err)
				require.Same(t, spec, valid)
				return
			}
			require.ErrorContains(t, err, tt.err)
			require.Equal(t, []string{"canvas_metric"}, names(valid.Controls))
			if tt.control != nil {
				// The control that failed is left out of the valid spec, and the input is not modified.
				require.Empty(t, valid.Rows[0].Items[0].Controls)
				require.Len(t, item.Controls, 1)
			}
		})
	}
}

func TestValidateControlSlots(t *testing.T) {
	mv := &runtimev1.MetricsViewSpec{
		Dimensions: []*runtimev1.MetricsViewSpec_Dimension{{Name: "device"}, {Name: "country"}},
		Measures:   []*runtimev1.MetricsViewSpec_Measure{{Name: "spend"}},
	}
	for _, tt := range []struct {
		name, renderer, target, option string
		multiple, valid                bool
		props                          map[string]any
	}{
		{"dimension", "bar_chart", "x.field", "country", false, true, nil},
		{"measure in dimension slot", "bar_chart", "x.field", "spend", false, false, nil},
		{"list in scalar slot", "bar_chart", "x.field", "device", true, false, nil},
		{"dimension list", "pivot", "row_dimensions", "country", true, true, map[string]any{"measures": []any{"spend"}, "row_dimensions": []any{"device"}}},
		{"measure in dimension list", "pivot", "row_dimensions", "spend", true, false, map[string]any{"measures": []any{"spend"}, "row_dimensions": []any{"device"}}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			props := tt.props
			if props == nil {
				props = map[string]any{"x": map[string]any{"field": "device", "type": "nominal"}, "y": map[string]any{"field": "spend", "type": "quantitative"}}
			}
			props["metrics_view"] = "bids"
			pb, err := structpb.NewStruct(props)
			require.NoError(t, err)
			ctrl := &runtimev1.CanvasControl{Kind: ControlKindSlot, Name: "pick", Target: tt.target, Options: []string{tt.option}, Multiple: tt.multiple}
			spec := &runtimev1.CanvasSpec{Rows: []*runtimev1.CanvasRow{{Items: []*runtimev1.CanvasItem{{Component: "chart", Controls: []*runtimev1.CanvasControl{ctrl}}}}}}
			_, err = ValidateControls(spec, map[string]*runtimev1.ComponentSpec{"chart": {Renderer: tt.renderer, RendererProperties: pb}}, map[string]*runtimev1.MetricsViewSpec{"bids": mv})
			if tt.valid {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, `control "pick" option`)
			}
			require.Equal(t, props, pb.AsMap())
		})
	}
}
