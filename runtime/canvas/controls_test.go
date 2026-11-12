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
	props, err := structpb.NewStruct(map[string]any{
		"metrics_view": "bids",
		"x":            map[string]any{"field": "device", "type": "nominal"},
		"y":            map[string]any{"field": "spend", "type": "quantitative"},
	})
	require.NoError(t, err)
	components := map[string]*runtimev1.ComponentSpec{"chart": {Renderer: "bar_chart", RendererProperties: props}}
	mvs := map[string]*runtimev1.MetricsViewSpec{"bids": {
		Dimensions: []*runtimev1.MetricsViewSpec_Dimension{{Name: "device"}, {Name: "country"}},
		Measures:   []*runtimev1.MetricsViewSpec_Measure{{Name: "spend"}, {Name: "impressions"}},
	}}
	spec := func(c *runtimev1.CanvasControl) *runtimev1.CanvasSpec {
		return &runtimev1.CanvasSpec{Rows: []*runtimev1.CanvasRow{{Items: []*runtimev1.CanvasItem{{Component: "chart", Name: "chart", Controls: []*runtimev1.CanvasControl{c}}}}}}
	}

	ok := &runtimev1.CanvasControl{Kind: ControlKindMeasure, Target: "spend", Name: "spend", Options: []string{"spend", "impressions"}}
	require.NoError(t, ValidateControls(spec(ok), components, mvs))

	bad := &runtimev1.CanvasControl{Kind: ControlKindMeasure, Target: "spend", Name: "spend", Options: []string{"spend", "win_rate"}}
	require.ErrorContains(t, ValidateControls(spec(bad), components, mvs), `option "win_rate"`)

	// A swap of a field the item doesn't use is not checked against it.
	unused := &runtimev1.CanvasControl{Kind: ControlKindMeasure, Target: "clicks", Name: "clicks", Options: []string{"clicks", "win_rate"}}
	require.NoError(t, ValidateControls(spec(unused), components, mvs))

	slot := &runtimev1.CanvasControl{Kind: ControlKindSlot, Target: "x.field", Name: "x_field", Options: []string{"device", "nope"}}
	require.ErrorContains(t, ValidateControls(spec(slot), components, mvs), `option "nope"`)
}
