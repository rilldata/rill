package parser

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/canvas"
	"github.com/rilldata/rill/runtime/pkg/rilltime"
	"gopkg.in/yaml.v3"
)

// canvasControlNameRegex restricts the names of controls and items to identifiers, since they are segments of URL keys.
var canvasControlNameRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// canvasControlGrains are the time grains a time grain control can offer.
var canvasControlGrains = []string{"minute", "hour", "day", "week", "month", "quarter", "year"}

// parseCanvasControls parses the `controls` list of a canvas, container, tab group, tab or item.
// Slot controls are only allowed on items, since a slot is a property of one renderer.
func parseCanvasControls(controls []*canvasControlYAML, onItem bool) ([]*runtimev1.CanvasControl, error) {
	if len(controls) == 0 {
		return nil, nil
	}
	res := make([]*runtimev1.CanvasControl, 0, len(controls))
	seen := make(map[string]bool, len(controls))
	for i, c := range controls {
		if c == nil {
			return nil, fmt.Errorf("control %d is empty", i)
		}
		ctrl, err := parseCanvasControl(c, onItem)
		if err != nil {
			return nil, fmt.Errorf("control %d: %w", i, err)
		}
		if seen[ctrl.Name] {
			return nil, fmt.Errorf("control %d: duplicate control name %q: set a different 'name' on one of them", i, ctrl.Name)
		}
		seen[ctrl.Name] = true
		res = append(res, ctrl)
	}
	return res, nil
}

func parseCanvasControl(c *canvasControlYAML, onItem bool) (*runtimev1.CanvasControl, error) {
	// Resolve the kind and target from the one target key that is set.
	var kind, target, name string
	n := 0
	if c.Measure != "" {
		n++
		kind, target, name = canvas.ControlKindMeasure, c.Measure, c.Measure
	}
	if c.Dimension != "" {
		n++
		kind, target, name = canvas.ControlKindDimension, c.Dimension, c.Dimension
	}
	if c.Slot != "" {
		n++
		kind, target, name = canvas.ControlKindSlot, c.Slot, strings.ReplaceAll(c.Slot, ".", "_")
	}
	if c.Filter != "" {
		n++
		kind, target, name = canvas.ControlKindFilter, c.Filter, c.Filter
	}
	if c.Time != "" {
		n++
		switch c.Time {
		case "range":
			kind, name = canvas.ControlKindTimeRange, "time_range"
		case "grain":
			kind, name = canvas.ControlKindTimeGrain, "time_grain"
		default:
			return nil, fmt.Errorf(`invalid 'time' value %q: expected "range" or "grain"`, c.Time)
		}
	}
	if n != 1 {
		return nil, errors.New("must set exactly one of 'measure', 'dimension', 'slot', 'filter' or 'time'")
	}
	if kind == canvas.ControlKindSlot && !onItem {
		return nil, errors.New("'slot' controls can only be declared on an item")
	}

	if c.Name != "" {
		name = c.Name
	}
	if !canvasControlNameRegex.MatchString(name) {
		return nil, fmt.Errorf("invalid control name %q: set 'name' to a valid identifier", name)
	}

	// Parse the default, which is a scalar or a list.
	var def []string
	switch c.Default.Kind {
	case 0:
	case yaml.ScalarNode:
		def = []string{c.Default.Value}
	case yaml.SequenceNode:
		for _, v := range c.Default.Content {
			if v.Kind != yaml.ScalarNode {
				return nil, errors.New("'default' must be a value or a list of values")
			}
			def = append(def, v.Value)
		}
	default:
		return nil, errors.New("'default' must be a value or a list of values")
	}

	// Validate the options, default and multiple per kind.
	switch kind {
	case canvas.ControlKindMeasure, canvas.ControlKindDimension:
		if c.Multiple {
			return nil, fmt.Errorf("'multiple' is not supported for %s controls; use a 'slot' control on an item instead", kind)
		}
		if len(c.Options) == 0 {
			return nil, errors.New("'options' must list the fields viewers can choose from")
		}
		if !slices.Contains(c.Options, target) {
			return nil, fmt.Errorf("'options' must include %q, which is what the items show by default", target)
		}
		if len(def) > 0 && (len(def) != 1 || def[0] != target) {
			return nil, fmt.Errorf("'default' is not supported for %s controls: the default is %q", kind, target)
		}
		def = nil
	case canvas.ControlKindSlot:
		if len(c.Options) == 0 {
			return nil, errors.New("'options' must list the fields viewers can choose from")
		}
		if len(def) > 0 {
			return nil, errors.New("'default' is not supported for slot controls: the default is the fields set on the item")
		}
	case canvas.ControlKindFilter:
	case canvas.ControlKindTimeRange:
		if c.Multiple {
			return nil, errors.New("'multiple' is not supported for time controls")
		}
		if len(c.Options) == 0 {
			return nil, errors.New("'options' must list the time ranges viewers can choose from")
		}
		for _, o := range c.Options {
			if _, err := rilltime.Parse(o, rilltime.ParseOptions{}); err != nil {
				return nil, fmt.Errorf("invalid time range %q in 'options': %w", o, err)
			}
		}
	case canvas.ControlKindTimeGrain:
		if c.Multiple {
			return nil, errors.New("'multiple' is not supported for time controls")
		}
		if len(c.Options) == 0 {
			return nil, errors.New("'options' must list the time grains viewers can choose from")
		}
		for _, o := range c.Options {
			if !slices.Contains(canvasControlGrains, o) {
				return nil, fmt.Errorf("invalid time grain %q in 'options': expected one of %s", o, strings.Join(canvasControlGrains, ", "))
			}
		}
	}
	if len(def) > 1 && !c.Multiple {
		return nil, errors.New("'default' has several values, but 'multiple' is not set")
	}
	if len(c.Options) > 0 {
		for _, d := range def {
			if !slices.Contains(c.Options, d) {
				return nil, fmt.Errorf("default %q is not one of the 'options'", d)
			}
		}
	}
	if len(c.Fields) > 0 && kind != canvas.ControlKindFilter {
		return nil, errors.New("'fields' is only supported for filter controls")
	}
	if c.Display != "" && !slices.Contains([]string{"dropdown", "buttons"}, c.Display) {
		return nil, fmt.Errorf(`invalid 'display' value %q: expected "dropdown" or "buttons"`, c.Display)
	}

	return &runtimev1.CanvasControl{
		Name:     name,
		Kind:     kind,
		Target:   target,
		Label:    c.Label,
		Options:  c.Options,
		Default:  def,
		Multiple: c.Multiple,
		Display:  c.Display,
		Fields:   c.Fields,
	}, nil
}

// parseIgnoreControls parses an item's `ignore_controls`, which is `true` to ignore every enclosing control, or a list of control names.
func parseIgnoreControls(n *yaml.Node) (bool, []string, error) {
	switch n.Kind {
	case 0:
		return false, nil, nil
	case yaml.ScalarNode:
		var all bool
		if err := n.Decode(&all); err != nil {
			return false, nil, errors.New("'ignore_controls' must be true or a list of control names")
		}
		return all, nil, nil
	case yaml.SequenceNode:
		var names []string
		if err := n.Decode(&names); err != nil {
			return false, nil, errors.New("'ignore_controls' must be true or a list of control names")
		}
		return false, names, nil
	default:
		return false, nil, errors.New("'ignore_controls' must be true or a list of control names")
	}
}

// validateCanvasControlScopes checks the names that make up the URL keys of controls across the whole canvas:
// items and blocks (tab groups, containers) share one namespace, and neither may be named like a canvas-level control.
func validateCanvasControlScopes(spec *runtimev1.CanvasSpec) error {
	seen := make(map[string]string)
	for _, c := range spec.Controls {
		seen[c.Name] = "canvas control"
	}
	claim := func(name, kind string) error {
		if other, ok := seen[name]; ok {
			return fmt.Errorf("the name %q is used by both a %s and a %s: names of items, tab groups, containers and canvas-level controls must be unique", name, other, kind)
		}
		seen[name] = kind
		return nil
	}
	var err error
	canvas.WalkRows(spec.Rows, canvas.Visit{
		Row: func(row *runtimev1.CanvasRow) {
			if err != nil {
				return
			}
			if g := row.GetTabGroup(); g != nil {
				err = claim(g.Name, "tab group")
			} else if c := row.GetContainer(); c != nil {
				err = claim(c.Name, "container")
			}
		},
		Item: func(item *runtimev1.CanvasItem) {
			if err == nil && item.Name != "" {
				err = claim(item.Name, "item")
			}
		},
	})
	return err
}
