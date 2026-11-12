package canvas

import (
	"fmt"
	"slices"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
)

// Kinds of canvas controls.
const (
	ControlKindMeasure   = "measure"
	ControlKindDimension = "dimension"
	ControlKindSlot      = "slot"
	ControlKindFilter    = "filter"
	ControlKindTimeRange = "time_range"
	ControlKindTimeGrain = "time_grain"
)

// ItemControls holds an item and the controls in effect for it.
type ItemControls struct {
	Item *runtimev1.CanvasItem
	// Controls in effect for the item, nearest level first.
	Controls []*runtimev1.CanvasControl
}

// ResolveItemControls returns the controls in effect for each item of the canvas.
// An item's own controls come first, then those of its tab, tab group or container, then the canvas's.
// For each target (see ControlTargetKey), only the nearest control is kept.
// An item's ignore_controls drops controls of the enclosing levels, never its own.
//
// The frontend implements the same rules in web-common/src/features/canvas/controls/controls.ts; keep them in sync.
func ResolveItemControls(spec *runtimev1.CanvasSpec) []ItemControls {
	var res []ItemControls
	var walk func(rows []*runtimev1.CanvasRow, enclosing [][]*runtimev1.CanvasControl)
	walk = func(rows []*runtimev1.CanvasRow, enclosing [][]*runtimev1.CanvasControl) {
		for _, row := range rows {
			if g := row.GetTabGroup(); g != nil {
				for _, t := range g.Tabs {
					// Nearest level first: the tab, then its group, then what encloses the group.
					walk(t.Rows, append([][]*runtimev1.CanvasControl{t.Controls, g.Controls}, enclosing...))
				}
				continue
			}
			if c := row.GetContainer(); c != nil {
				walk(c.Rows, append([][]*runtimev1.CanvasControl{c.Controls}, enclosing...))
				continue
			}
			for _, item := range row.Items {
				res = append(res, ItemControls{Item: item, Controls: resolveItem(item, enclosing)})
			}
		}
	}
	walk(spec.Rows, [][]*runtimev1.CanvasControl{spec.Controls})
	return res
}

func resolveItem(item *runtimev1.CanvasItem, enclosing [][]*runtimev1.CanvasControl) []*runtimev1.CanvasControl {
	var res []*runtimev1.CanvasControl
	seen := make(map[string]bool)
	for _, c := range item.Controls {
		if key := ControlTargetKey(c); !seen[key] {
			seen[key] = true
			res = append(res, c)
		}
	}
	if item.IgnoreAllControls {
		return res
	}
	for _, level := range enclosing {
		for _, c := range level {
			if slices.Contains(item.IgnoreControls, c.Name) {
				continue
			}
			if key := ControlTargetKey(c); !seen[key] {
				seen[key] = true
				res = append(res, c)
			}
		}
	}
	return res
}

// ControlTargetKey identifies what a control changes. Two controls with the same key compete, and the nearest one wins.
func ControlTargetKey(c *runtimev1.CanvasControl) string {
	return c.Kind + ":" + c.Target
}

// swapScalarKeys are the renderer property keys whose string value is a field name.
var swapScalarKeys = []string{"field", "name", "measure", "dimension", "sort_by"}

// swapListKeys are the renderer property keys whose value is a list of field names or field objects.
var swapListKeys = []string{"fields", "measures", "dimensions", "columns", "row_dimensions", "col_dimensions"}

// SwapField returns a copy of the renderer properties with the field `from` replaced by `to` wherever it is used as a field,
// and whether it was used at all. Other strings that happen to equal `from` (titles, the metrics view name, filters) are left as they are.
//
// The frontend implements the same rules in web-common/src/features/canvas/controls/apply-fields.ts; keep them in sync.
func SwapField(props map[string]any, from, to string) (map[string]any, bool) {
	used := false
	var walk func(v any, key string, inList bool) any
	walk = func(v any, key string, inList bool) any {
		switch v := v.(type) {
		case string:
			if v == from && (inList || slices.Contains(swapScalarKeys, key)) {
				used = true
				return to
			}
			return v
		case map[string]any:
			out := make(map[string]any, len(v))
			for k, child := range v {
				out[k] = walk(child, k, false)
			}
			return out
		case []any:
			isFieldList := slices.Contains(swapListKeys, key)
			out := make([]any, len(v))
			for i, child := range v {
				out[i] = walk(child, key, isFieldList)
			}
			return out
		default:
			return v
		}
	}
	res, _ := walk(props, "", false).(map[string]any)
	return res, used
}

// ValidateControls checks the controls of a canvas against the components and metrics views they apply to:
// every option of a control that swaps a field must work in every item that uses the field.
// components maps component names to their specs, and metricsViews maps metrics view names to their valid specs;
// items whose component or metrics view is missing are skipped, since they are reported elsewhere.
func ValidateControls(spec *runtimev1.CanvasSpec, components map[string]*runtimev1.ComponentSpec, metricsViews map[string]*runtimev1.MetricsViewSpec) error {
	for _, ic := range ResolveItemControls(spec) {
		comp := components[ic.Item.Component]
		if comp == nil || comp.RendererProperties == nil {
			continue
		}
		props := comp.RendererProperties.AsMap()
		mvName, _ := props["metrics_view"].(string)
		mv := metricsViews[mvName]
		if mv == nil {
			continue
		}
		for _, c := range ic.Controls {
			switch c.Kind {
			case ControlKindMeasure, ControlKindDimension:
				for _, opt := range c.Options {
					if opt == c.Target {
						continue
					}
					swapped, used := SwapField(props, c.Target, opt)
					if !used {
						break
					}
					if err := ValidateRendererProperties(comp.Renderer, swapped, metricsViews, false); err != nil {
						return fmt.Errorf("control %q option %q doesn't work in component %q: %w", c.Name, opt, ic.Item.Component, err)
					}
				}
			case ControlKindSlot:
				for _, opt := range c.Options {
					if !metricsViewHasMeasure(mv, opt) && !metricsViewHasDimension(mv, opt) {
						return fmt.Errorf("control %q option %q is not a measure or dimension of metrics view %q (component %q)", c.Name, opt, mvName, ic.Item.Component)
					}
				}
			case ControlKindFilter:
				dim := c.Target
				if d, ok := c.Fields[mvName]; ok {
					dim = d
				}
				// A filter only applies to items whose metrics view has the dimension, so a missing one is fine
				// unless the control explicitly maps the dimension for this metrics view.
				if _, mapped := c.Fields[mvName]; mapped && !metricsViewHasDimension(mv, dim) {
					return fmt.Errorf("control %q maps to %q, which is not a dimension of metrics view %q", c.Name, dim, mvName)
				}
			}
		}
	}
	return nil
}
