package canvas

import (
	"cmp"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"google.golang.org/protobuf/proto"
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
// The controls in skip are left out, as if the canvas didn't declare them; skip can be nil.
//
// The frontend implements the same rules in web-common/src/features/canvas/controls/controls.ts; keep them in sync.
func ResolveItemControls(spec *runtimev1.CanvasSpec, skip map[*runtimev1.CanvasControl]bool) []ItemControls {
	var res []ItemControls
	walkCanvas(spec, nil, func(item *runtimev1.CanvasItem, enclosing [][]*runtimev1.CanvasControl) {
		res = append(res, ItemControls{Item: item, Controls: resolveItem(item, enclosing, skip)})
	})
	return res
}

// walkCanvas walks the canvas in document order.
// It calls onLevel with the list of controls of each level, from the canvas down to its items; onLevel can replace the list through the pointer.
// It calls onItem with each item and the controls of its enclosing levels, nearest level first.
// Either callback can be nil.
func walkCanvas(spec *runtimev1.CanvasSpec, onLevel func(controls *[]*runtimev1.CanvasControl), onItem func(item *runtimev1.CanvasItem, enclosing [][]*runtimev1.CanvasControl)) {
	if onLevel == nil {
		onLevel = func(*[]*runtimev1.CanvasControl) {}
	}
	if onItem == nil {
		onItem = func(*runtimev1.CanvasItem, [][]*runtimev1.CanvasControl) {}
	}
	var walk func(rows []*runtimev1.CanvasRow, enclosing [][]*runtimev1.CanvasControl)
	walk = func(rows []*runtimev1.CanvasRow, enclosing [][]*runtimev1.CanvasControl) {
		for _, row := range rows {
			if g := row.GetTabGroup(); g != nil {
				onLevel(&g.Controls)
				for _, t := range g.Tabs {
					onLevel(&t.Controls)
					// Nearest level first: the tab, then its group, then what encloses the group.
					walk(t.Rows, append([][]*runtimev1.CanvasControl{t.Controls, g.Controls}, enclosing...))
				}
				continue
			}
			if c := row.GetContainer(); c != nil {
				onLevel(&c.Controls)
				walk(c.Rows, append([][]*runtimev1.CanvasControl{c.Controls}, enclosing...))
				continue
			}
			for _, item := range row.Items {
				onLevel(&item.Controls)
				onItem(item, enclosing)
			}
		}
	}
	onLevel(&spec.Controls)
	walk(spec.Rows, [][]*runtimev1.CanvasControl{spec.Controls})
}

func resolveItem(item *runtimev1.CanvasItem, enclosing [][]*runtimev1.CanvasControl, skip map[*runtimev1.CanvasControl]bool) []*runtimev1.CanvasControl {
	var res []*runtimev1.CanvasControl
	seen := make(map[string]bool)
	for _, c := range item.Controls {
		if skip[c] {
			continue
		}
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
			if skip[c] || slices.Contains(item.IgnoreControls, c.Name) {
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

// ValidateControls checks the controls of a canvas against its items, components and metrics views.
// A control fails when an option of a field swap doesn't work in an item that uses the field or would show a field twice in it,
// when it names a field that none of the canvas's metrics views has,
// or when it is a slot control whose property the item doesn't set.
// An item's ignore_controls must also name controls of its enclosing levels.
//
// It returns the spec without the controls that fail, so a broken control doesn't take down a canvas that works without it,
// and an error describing every problem found; when there are none, it returns the spec as is.
// components maps component names to their specs, and metricsViews maps metrics view names to their valid specs;
// items whose component or metrics view is missing are skipped, since they are reported elsewhere.
func ValidateControls(spec *runtimev1.CanvasSpec, components map[string]*runtimev1.ComponentSpec, metricsViews map[string]*runtimev1.MetricsViewSpec) (*runtimev1.CanvasSpec, error) {
	var errs []error
	invalid := make(map[*runtimev1.CanvasControl]bool)
	fail := func(c *runtimev1.CanvasControl, err error) {
		if !invalid[c] {
			invalid[c] = true
			errs = append(errs, err)
		}
	}

	// The fields of the canvas are only known when every item's metrics view is valid;
	// they include the adhoc measures items define.
	knownFields := true
	adhocMeasures := make(map[string]bool)
	walkCanvas(spec, nil, func(item *runtimev1.CanvasItem, _ [][]*runtimev1.CanvasControl) {
		comp := components[item.Component]
		if comp == nil || comp.RendererProperties == nil {
			knownFields = false
			return
		}
		props := comp.RendererProperties.AsMap()
		mvName, _ := props["metrics_view"].(string)
		mv := metricsViews[mvName]
		if mv == nil {
			knownFields = knownFields && mvName == ""
			return
		}
		if names, err := ephemeralMeasureNames(props, mvName, mv); err == nil {
			maps.Copy(adhocMeasures, names)
		}
	})

	// Check the fields each control names, including controls whose field no item uses.
	if knownFields {
		hasDimension := func(mv *runtimev1.MetricsViewSpec, name string) bool {
			return metricsViewHasDimension(mv, name) || name == mv.TimeDimension || isEncodedTimeDimension(mv, name)
		}
		walkCanvas(spec, func(controls *[]*runtimev1.CanvasControl) {
			for _, c := range *controls {
				switch c.Kind {
				case ControlKindMeasure, ControlKindDimension:
					for _, opt := range c.Options {
						found := c.Kind == ControlKindMeasure && adhocMeasures[opt]
						for _, mv := range metricsViews {
							if c.Kind == ControlKindMeasure {
								found = found || metricsViewHasMeasure(mv, opt)
							} else {
								found = found || hasDimension(mv, opt)
							}
						}
						if !found {
							fail(c, fmt.Errorf("control %q option %q is not a %s of any metrics view of the canvas", c.Name, opt, c.Kind))
							break
						}
					}
				case ControlKindFilter:
					found := false
					for mvName, mv := range metricsViews {
						dim := c.Target
						if d, ok := c.Fields[mvName]; ok {
							dim = d
						}
						found = found || hasDimension(mv, dim)
					}
					if !found {
						fail(c, fmt.Errorf("control %q filters on %q, which is not a dimension of any metrics view of the canvas", c.Name, c.Target))
					}
				}
			}
		}, nil)
	}

	// Check each control against the items it applies to, leaving out the controls that failed.
	// Leaving one out puts the control of an enclosing level it overrode in effect for its items,
	// so resolve the items again and check until no more controls fail.
	for {
		n := len(invalid)
		for _, ic := range ResolveItemControls(spec, invalid) {
			validateItemControls(ic, components, metricsViews, fail)
		}
		if len(invalid) == n {
			break
		}
	}

	// An item can only ignore controls of its enclosing levels.
	walkCanvas(spec, nil, func(item *runtimev1.CanvasItem, enclosing [][]*runtimev1.CanvasControl) {
		for _, name := range item.IgnoreControls {
			found := false
			for _, level := range enclosing {
				found = found || slices.ContainsFunc(level, func(c *runtimev1.CanvasControl) bool { return c.Name == name })
			}
			if !found {
				errs = append(errs, fmt.Errorf("item %q ignores control %q, which none of its enclosing levels declares", cmp.Or(item.Name, item.Component), name))
			}
		}
	})

	if len(invalid) == 0 {
		return spec, errors.Join(errs...)
	}
	return withoutControls(spec, invalid), errors.Join(errs...)
}

// validateItemControls checks the controls in effect for an item against its component, and calls fail for each control that doesn't work in it.
// Items whose component or metrics view is missing are skipped.
func validateItemControls(ic ItemControls, components map[string]*runtimev1.ComponentSpec, metricsViews map[string]*runtimev1.MetricsViewSpec, fail func(c *runtimev1.CanvasControl, err error)) {
	comp := components[ic.Item.Component]
	if comp == nil || comp.RendererProperties == nil {
		return
	}
	props := comp.RendererProperties.AsMap()
	mvName, _ := props["metrics_view"].(string)
	mv := metricsViews[mvName]
	if mv == nil {
		return
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
				if listsFieldTwice(swapped, opt) {
					fail(c, fmt.Errorf("control %q option %q is already shown in component %q, so choosing it would show it twice", c.Name, opt, ic.Item.Component))
					break
				}
				if err := ValidateRendererProperties(comp.Renderer, swapped, metricsViews, false); err != nil {
					fail(c, fmt.Errorf("control %q option %q doesn't work in component %q: %w", c.Name, opt, ic.Item.Component, err))
					break
				}
			}
		case ControlKindSlot:
			// The slot must name a property the item sets, since its default is the item's own value.
			var cur any = props
			for _, key := range strings.Split(c.Target, ".") {
				m, _ := cur.(map[string]any)
				cur = m[key]
			}
			if cur == nil {
				fail(c, fmt.Errorf("control %q: component %q doesn't set %q", c.Name, ic.Item.Component, c.Target))
				continue
			}
			for _, opt := range c.Options {
				if !metricsViewHasMeasure(mv, opt) && !metricsViewHasDimension(mv, opt) {
					fail(c, fmt.Errorf("control %q option %q is not a measure or dimension of metrics view %q (component %q)", c.Name, opt, mvName, ic.Item.Component))
					break
				}
				var value any = opt
				if c.Multiple {
					value = []any{opt}
				}
				selected := setControlSlot(props, strings.Split(c.Target, "."), value)
				if err := ValidateRendererProperties(comp.Renderer, selected, metricsViews, false); err != nil {
					fail(c, fmt.Errorf("control %q option %q doesn't work in component %q: %w", c.Name, opt, ic.Item.Component, err))
					break
				}
			}
		case ControlKindFilter:
			// A filter only applies to items whose metrics view has the dimension, so a missing one is fine
			// unless the control explicitly maps the dimension for this metrics view.
			if dim, mapped := c.Fields[mvName]; mapped && !metricsViewHasDimension(mv, dim) {
				fail(c, fmt.Errorf("control %q maps to %q, which is not a dimension of metrics view %q", c.Name, dim, mvName))
			}
		}
	}
}

// withoutControls returns a copy of the spec without the given controls.
func withoutControls(spec *runtimev1.CanvasSpec, drop map[*runtimev1.CanvasControl]bool) *runtimev1.CanvasSpec {
	// The copy has the same lists in the same order, so mark the controls to drop by position.
	var dropAt []bool
	walkCanvas(spec, func(controls *[]*runtimev1.CanvasControl) {
		for _, c := range *controls {
			dropAt = append(dropAt, drop[c])
		}
	}, nil)
	res := proto.Clone(spec).(*runtimev1.CanvasSpec)
	i := 0
	walkCanvas(res, func(controls *[]*runtimev1.CanvasControl) {
		var kept []*runtimev1.CanvasControl
		for _, c := range *controls {
			if !dropAt[i] {
				kept = append(kept, c)
			}
			i++
		}
		*controls = kept
	}, nil)
	return res
}

// listsFieldTwice reports whether a list of fields in the renderer properties (see swapListKeys) names the field more than once.
func listsFieldTwice(props map[string]any, field string) bool {
	var walk func(v any, key string) bool
	walk = func(v any, key string) bool {
		switch v := v.(type) {
		case map[string]any:
			for k, child := range v {
				if walk(child, k) {
					return true
				}
			}
		case []any:
			n := 0
			for _, child := range v {
				name, ok := child.(string)
				if m, isMap := child.(map[string]any); isMap {
					name, ok = m["name"].(string)
				}
				if ok && name == field && slices.Contains(swapListKeys, key) {
					n++
				}
				if walk(child, key) {
					return true
				}
			}
			return n > 1
		}
		return false
	}
	return walk(props, "")
}

// setControlSlot copies the maps along the slot path, matching the frontend's setPath.
func setControlSlot(props map[string]any, path []string, value any) map[string]any {
	out := make(map[string]any, len(props)+1)
	for key, child := range props {
		out[key] = child
	}
	if len(path) == 1 {
		out[path[0]] = value
	} else {
		child, _ := props[path[0]].(map[string]any)
		out[path[0]] = setControlSlot(child, path[1:], value)
	}
	return out
}
