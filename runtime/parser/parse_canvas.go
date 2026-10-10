package parser

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/metricsview"
	"github.com/rilldata/rill/runtime/metricsview/metricssql"
	"github.com/rilldata/rill/runtime/pkg/rilltime"
	"github.com/rilldata/rill/runtime/pkg/urlutils"
	"golang.org/x/exp/maps"
	"gopkg.in/yaml.v3"
)

type CanvasYAML struct {
	commonYAML           `yaml:",inline"`       // Not accessed here, only setting it so we can use KnownFields for YAML parsing
	DisplayName          string                 `yaml:"display_name"`
	Title                string                 `yaml:"title"` // Deprecated: use display_name
	Banner               string                 `yaml:"banner"`
	MaxWidth             uint32                 `yaml:"max_width"`
	GapX                 uint32                 `yaml:"gap_x"`
	GapY                 uint32                 `yaml:"gap_y"`
	Theme                yaml.Node              `yaml:"theme"` // Name (string) or inline theme definition (map)
	AllowCustomTimeRange *bool                  `yaml:"allow_custom_time_range"`
	TimeRanges           []ExploreTimeRangeYAML `yaml:"time_ranges"`
	TimeZones            []string               `yaml:"time_zones"`
	Filters              struct {
		Enable   *bool    `yaml:"enable"`
		Pinned   []string `yaml:"pinned"`
		Required []string `yaml:"required"`
	}
	Defaults *struct {
		TimeRange           string            `yaml:"time_range"`
		ComparisonMode      string            `yaml:"comparison_mode"`
		ComparisonDimension string            `yaml:"comparison_dimension"`
		Filters             map[string]string `yaml:"filters"`
	} `yaml:"defaults"`
	Variables   []*ComponentVariableYAML `yaml:"variables"`
	Rows        []*canvasRowYAML         `yaml:"rows"`
	Security    *SecurityPolicyYAML      `yaml:"security"`
	Annotations map[string]string        `yaml:"annotations"`
	AIPrompts   []AIPromptYAML           `yaml:"ai_prompts"`
}

// canvasRowYAML is a single entry in a canvas's (or tab's) rows list.
// It is either a plain row (items) or a tab group (tabs), never both.
type canvasRowYAML struct {
	Height *string           `yaml:"height"`
	Items  []*canvasItemYAML `yaml:"items"`
	// Name is the stable identifier of a tab group. Only used for tab-group entries.
	Name string `yaml:"name"`
	// Tabs, when set, makes this entry a tab group instead of a plain row.
	Tabs []*canvasTabYAML `yaml:"tabs"`
	// If is a templated condition; the row (or tab group) is only shown to viewers for whom it's true.
	If string `yaml:"if"`
}

// canvasTabYAML is a single tab within a tab group. `label` is the display name shown on the
// tab; `name` is the stable URL key (defaulted from the label when omitted).
type canvasTabYAML struct {
	Name  string           `yaml:"name"`
	Label string           `yaml:"label"`
	Rows  []*canvasRowYAML `yaml:"rows"`
	// If is a templated condition; the tab is only shown to viewers for whom it's true.
	If string `yaml:"if"`
}

// canvasItemYAML is a single item within a row.
type canvasItemYAML struct {
	Width     *string `yaml:"width"`
	Component string  `yaml:"component"` // Name of an externally defined component
	// If is a templated condition; the item is only shown to viewers for whom it's true.
	If              string               `yaml:"if"`
	InlineComponent map[string]yaml.Node `yaml:",inline"` // Any other properties are considered an inline component definition
}

func (p *Parser) parseCanvas(node *Node) error {
	// Parse YAML
	tmp := &CanvasYAML{}
	err := p.decodeNodeYAML(node, true, tmp)
	if err != nil {
		return err
	}

	// Validate SQL or connector isn't set
	if node.SQL != "" {
		return fmt.Errorf("canvases cannot have SQL")
	}
	if !node.ConnectorInferred && node.Connector != "" {
		return fmt.Errorf("canvases cannot have a connector")
	}

	// Display name backwards compatibility
	if tmp.Title != "" && tmp.DisplayName == "" {
		tmp.DisplayName = tmp.Title
	}

	// Set default for AllowCustomTimeRange to true if not provided
	allowCustomTimeRange := true
	if tmp.AllowCustomTimeRange != nil {
		allowCustomTimeRange = *tmp.AllowCustomTimeRange
	}

	// Parse theme if present.
	// If it returns a themeSpec, it will be inserted as a separate resource later in this function.
	themeName, themeSpec, err := p.parseThemeRef(&tmp.Theme)
	if err != nil {
		return err
	}
	// Fallback to top-level theme from rill.yaml if no local theme or default theme is set
	if themeName == "" && themeSpec == nil && p.RillYAML != nil && p.RillYAML.Theme != "" {
		themeName = p.RillYAML.Theme
	}
	if themeName != "" && themeSpec == nil {
		node.Refs = append(node.Refs, ResourceName{Kind: ResourceKindTheme, Name: themeName})
	}

	// Build and validate time ranges
	var timeRanges []*runtimev1.ExploreTimeRange
	for _, tr := range tmp.TimeRanges {
		if _, err := rilltime.Parse(tr.Range, rilltime.ParseOptions{}); err != nil {
			return fmt.Errorf("invalid time range %q: %w", tr.Range, err)
		}
		res := &runtimev1.ExploreTimeRange{Range: tr.Range}
		for _, ctr := range tr.ComparisonTimeRanges {
			err = rilltime.ParseCompatibility(ctr.Range, ctr.Offset)
			if err != nil {
				return err
			}
			res.ComparisonTimeRanges = append(res.ComparisonTimeRanges, &runtimev1.ExploreComparisonTimeRange{
				Offset: ctr.Offset,
				Range:  ctr.Range,
			})
		}
		timeRanges = append(timeRanges, res)
	}

	// Validate time zones
	for _, tz := range tmp.TimeZones {
		_, err := time.LoadLocation(tz)
		if err != nil {
			return err
		}
	}

	// Parse variable definitions.
	var variables []*runtimev1.ComponentVariable
	if len(tmp.Variables) > 0 {
		variables = make([]*runtimev1.ComponentVariable, len(tmp.Variables))
	}
	for i, v := range tmp.Variables {
		variables[i], err = v.Proto()
		if err != nil {
			return fmt.Errorf("invalid variable at index %d: %w", i, err)
		}
	}

	// Parse rows and items.
	// Each row entry is either a plain row (items) or a tab group (tabs); tab groups are only allowed at the top level.
	var inlineComponentDefs []*componentDef // Track inline component definitions so we can insert them after we have validated all components
	rows, err := p.parseCanvasRows(node, tmp.Rows, true, "", nil, &inlineComponentDefs)
	if err != nil {
		return err
	}

	// Build and validate presets
	var defaultPreset *runtimev1.CanvasPreset
	if tmp.Defaults != nil {
		if tmp.Defaults.TimeRange != "" {
			if _, err := rilltime.Parse(tmp.Defaults.TimeRange, rilltime.ParseOptions{}); err != nil {
				return fmt.Errorf("invalid time range %q: %w", tmp.Defaults.TimeRange, err)
			}
		}

		mode := runtimev1.ExploreComparisonMode_EXPLORE_COMPARISON_MODE_NONE
		if tmp.Defaults.ComparisonMode != "" {
			var ok bool
			mode, ok = exploreComparisonModes[tmp.Defaults.ComparisonMode]
			if !ok {
				return fmt.Errorf("invalid comparison mode %q (options: %s)", tmp.Defaults.ComparisonMode, strings.Join(maps.Keys(exploreComparisonModes), ", "))
			}
		}

		if tmp.Defaults.ComparisonDimension != "" && mode != runtimev1.ExploreComparisonMode_EXPLORE_COMPARISON_MODE_DIMENSION {
			return errors.New("can only set comparison_dimension when comparison_mode is 'dimension'")
		}

		filterExpr, err := parseFilterExpressions(tmp.Defaults.Filters)
		if err != nil {
			return fmt.Errorf("invalid filter expression in defaults: %w", err)
		}

		defaultPreset = &runtimev1.CanvasPreset{
			TimeRange:           pointerIfNotEmpty(tmp.Defaults.TimeRange),
			ComparisonMode:      mode,
			ComparisonDimension: pointerIfNotEmpty(tmp.Defaults.ComparisonDimension),
			FilterExpr:          filterExpr,
		}
	}

	// Parse security rules
	rules, err := tmp.Security.Proto()
	if err != nil {
		return err
	}
	for _, rule := range rules {
		if rule.GetAccess() == nil {
			return fmt.Errorf("the 'canvas' resource type only supports 'access' security rules")
		}
	}

	// Validate the configured AI prompts
	aiPrompts, err := parseAIPrompts(tmp.AIPrompts)
	if err != nil {
		return err
	}

	// Track canvas
	r, err := p.insertResource(ResourceKindCanvas, node.Name, node.Paths, node.Tags, node.Metadata, node.Refs...)
	if err != nil {
		return err
	}
	// NOTE: After calling insertResource, an error must not be returned. Any validation should be done before calling it.

	r.CanvasSpec.DisplayName = tmp.DisplayName
	if r.CanvasSpec.DisplayName == "" {
		r.CanvasSpec.DisplayName = ToDisplayName(node.Name)
	}
	r.CanvasSpec.Banner = tmp.Banner
	r.CanvasSpec.MaxWidth = tmp.MaxWidth
	r.CanvasSpec.GapX = tmp.GapX
	r.CanvasSpec.GapY = tmp.GapY
	r.CanvasSpec.Theme = themeName
	r.CanvasSpec.AllowCustomTimeRange = allowCustomTimeRange
	r.CanvasSpec.TimeRanges = timeRanges
	r.CanvasSpec.TimeZones = tmp.TimeZones
	r.CanvasSpec.FiltersEnabled = true
	if tmp.Filters.Enable != nil {
		r.CanvasSpec.FiltersEnabled = *tmp.Filters.Enable
	}
	r.CanvasSpec.DefaultPreset = defaultPreset
	r.CanvasSpec.EmbeddedTheme = themeSpec
	r.CanvasSpec.Variables = variables
	r.CanvasSpec.Rows = rows
	r.CanvasSpec.SecurityRules = rules
	r.CanvasSpec.PinnedFilters = tmp.Filters.Pinned
	r.CanvasSpec.RequiredFilters = tmp.Filters.Required
	r.CanvasSpec.Annotations = tmp.Annotations
	r.CanvasSpec.AiPrompts = aiPrompts

	// Track inline components
	for _, def := range inlineComponentDefs {
		r, err := p.insertResource(ResourceKindComponent, def.name, node.Paths, nil, nil, def.refs...)
		if err != nil {
			// Normally we could return the error, but we can't do that here because we've already inserted the canvas.
			// Since the component has been validated with insertDryRun in parseCanvasItemComponent, this error should never happen in practice.
			// So let's panic.
			panic(err)
		}
		r.ComponentSpec = def.spec
	}

	return nil
}

// parseCanvasRows parses a list of canvas row entries. Each entry is either a plain row (items)
// or a tab group (tabs). Tab groups are only allowed when allowTabs is true (the top level);
// a tab's own rows are always plain. posPrefix disambiguates inline component names across tabs.
// conditions are the `if` conditions of the enclosing tab group and tab, which inline components inherit.
func (p *Parser) parseCanvasRows(node *Node, rows []*canvasRowYAML, allowTabs bool, posPrefix string, conditions []string, inlineComponentDefs *[]*componentDef) ([]*runtimev1.CanvasRow, error) {
	var out []*runtimev1.CanvasRow
	// seenGroupNames tracks tab group names so each group has a unique URL key. Only populated at the top level.
	seenGroupNames := make(map[string]bool)
	for i, row := range rows {
		if row == nil {
			return nil, fmt.Errorf("row at index %d is empty", i)
		}

		rowCondition, err := parseCondition(row.If)
		if err != nil {
			return nil, fmt.Errorf("invalid 'if' for row %d: %w", i, err)
		}
		rowConditions := appendCondition(conditions, rowCondition)

		// Dispatch on whether this entry is a tab group. Presence of the `tabs:` key (even if empty)
		// marks an entry as a group, so an empty `tabs: []` is rejected rather than silently treated as a row.
		if row.Tabs != nil {
			if len(row.Items) > 0 {
				return nil, fmt.Errorf("row %d cannot define both 'items' and 'tabs'", i)
			}
			if !allowTabs {
				return nil, fmt.Errorf("tab groups cannot be nested inside a tab (row %d)", i)
			}
			group, err := p.parseCanvasTabGroup(node, row, i, posPrefix, rowConditions, seenGroupNames, inlineComponentDefs)
			if err != nil {
				return nil, err
			}
			out = append(out, &runtimev1.CanvasRow{TabGroup: group, ConditionExpression: rowCondition})
			continue
		}

		var height *uint32
		var heightUnit string
		if row.Height != nil {
			v, u, err := parseItemSize(*row.Height)
			if err != nil {
				return nil, fmt.Errorf("invalid height for row %d: %w", i, err)
			}
			if v != 0 && u != "px" {
				return nil, fmt.Errorf("invalid height unit %q for row %d: unit must be 'px'", u, i)
			}
			height = &v
			heightUnit = u
		}

		var items []*runtimev1.CanvasItem
		for j, item := range row.Items {
			if item == nil {
				return nil, fmt.Errorf("item %d in row %d is empty", j, i)
			}

			var width *uint32
			var widthUnit string
			if item.Width != nil {
				v, u, err := parseItemSize(*item.Width)
				if err != nil {
					return nil, fmt.Errorf("invalid width for item %d in row %d: %w", j, i, err)
				}
				if u != "" {
					return nil, fmt.Errorf("invalid width unit %q for item %d in row %d: 'width' cannot have a unit", u, j, i)
				}
				width = &v
				widthUnit = u
			}

			itemCondition, err := parseCondition(item.If)
			if err != nil {
				return nil, fmt.Errorf("invalid 'if' for item %d in row %d: %w", j, i, err)
			}

			// Validate that exactly one of Component and InlineComponent are set
			if item.Component == "" && len(item.InlineComponent) == 0 {
				return nil, fmt.Errorf("item %d in row %d is missing a component definition", j, i)
			}
			if item.Component != "" && len(item.InlineComponent) > 0 {
				return nil, fmt.Errorf("item %d in row %d has properties incompatible with 'component'", j, i)
			}

			// Parse inline component definition if present and assign into item.Component
			var definedInCanvs bool
			if len(item.InlineComponent) > 0 {
				name, def, err := p.parseCanvasInlineComponent(node.Name, fmt.Sprintf("%s%d-%d", posPrefix, i, j), item.InlineComponent)
				if err != nil {
					return nil, fmt.Errorf("invalid component for item %d in row %d: %w", j, i, err)
				}

				// The component inherits the conditions of the item and its ancestors, so it can't be read by viewers who can't see it.
				def.spec.ConditionExpression = combineConditions(appendCondition(rowConditions, itemCondition))

				item.Component = name
				*inlineComponentDefs = append(*inlineComponentDefs, def)
				definedInCanvs = true
			}

			items = append(items, &runtimev1.CanvasItem{
				Component:           item.Component,
				DefinedInCanvas:     definedInCanvs,
				Width:               width,
				WidthUnit:           widthUnit,
				ConditionExpression: itemCondition,
			})

			node.Refs = append(node.Refs, ResourceName{Kind: ResourceKindComponent, Name: item.Component})
		}

		out = append(out, &runtimev1.CanvasRow{
			Height:              height,
			HeightUnit:          heightUnit,
			Items:               items,
			ConditionExpression: rowCondition,
		})
	}

	return out, nil
}

// parseCanvasTabGroup parses a single tab group entry (a row with tabs).
// conditions are the `if` conditions of the group and its ancestors, which inline components inherit.
func (p *Parser) parseCanvasTabGroup(node *Node, row *canvasRowYAML, rowIdx int, posPrefix string, conditions []string, seenGroupNames map[string]bool, inlineComponentDefs *[]*componentDef) (*runtimev1.CanvasTabGroup, error) {
	if len(row.Tabs) == 0 {
		return nil, fmt.Errorf("tab group at row %d must have at least one tab", rowIdx)
	}

	groupName := row.Name
	if groupName == "" {
		groupName = fmt.Sprintf("group-%d", rowIdx)
	}
	if seenGroupNames[groupName] {
		return nil, fmt.Errorf("duplicate tab group name %q at row %d", groupName, rowIdx)
	}
	seenGroupNames[groupName] = true

	var tabs []*runtimev1.CanvasTab
	seenNames := make(map[string]bool, len(row.Tabs))
	for t, tab := range row.Tabs {
		if tab == nil {
			return nil, fmt.Errorf("tab %d in tab group at row %d is empty", t, rowIdx)
		}

		if tab.Label == "" {
			return nil, fmt.Errorf("tab %d in tab group at row %d is missing a label", t, rowIdx)
		}

		// Prefer the explicit name as the URL key, then a slug of the label, then a positional
		// fallback. uniqueName uniquifies the result against earlier tabs in this group.
		target := tab.Name
		if target == "" {
			target = urlutils.Slugify(tab.Label)
		}
		tabName := uniqueName(target, fmt.Sprintf("tab-%d", t), seenNames)
		seenNames[tabName] = true

		tabCondition, err := parseCondition(tab.If)
		if err != nil {
			return nil, fmt.Errorf("invalid 'if' for tab %q in tab group at row %d: %w", tab.Label, rowIdx, err)
		}

		tabRows, err := p.parseCanvasRows(node, tab.Rows, false, fmt.Sprintf("%sg%d-t%d-", posPrefix, rowIdx, t), appendCondition(conditions, tabCondition), inlineComponentDefs)
		if err != nil {
			return nil, fmt.Errorf("invalid tab %q in tab group at row %d: %w", tab.Label, rowIdx, err)
		}

		tabs = append(tabs, &runtimev1.CanvasTab{
			Name:                tabName,
			DisplayName:         tab.Label,
			Rows:                tabRows,
			ConditionExpression: tabCondition,
		})
	}

	return &runtimev1.CanvasTabGroup{
		Name: groupName,
		Tabs: tabs,
	}, nil
}

// uniqueName returns name if it is non-empty and unused, otherwise it derives a unique
// alternative: first the supplied fallback, then fallback with a numeric suffix.
func uniqueName(name, fallback string, seen map[string]bool) string {
	if name == "" {
		name = fallback
	}
	if !seen[name] {
		return name
	}
	for n := 2; ; n++ {
		candidate := fmt.Sprintf("%s-%d", name, n)
		if !seen[candidate] {
			return candidate
		}
	}
}

// parseCondition validates the `if` condition of a canvas row, tab or item and returns it trimmed.
// The condition uses the same templating and expression syntax as a security policy's `access`.
// Conditions that reference custom user attributes can't be evaluated without a real user,
// so for those only the template syntax is checked.
func parseCondition(condition string) (string, error) {
	condition = strings.TrimSpace(condition)
	if condition == "" {
		return "", nil
	}
	resolved, err := ResolveTemplate(condition, validationTemplateData, false)
	if err != nil {
		return "", fmt.Errorf("templating is not valid: %w", err)
	}
	if strings.Contains(resolved, "<no value>") {
		return condition, nil
	}
	if _, err := EvaluateBoolExpression(resolved); err != nil {
		return "", fmt.Errorf("expression error: %w", err)
	}
	return condition, nil
}

// appendCondition returns conditions with condition appended, without modifying the provided slice.
func appendCondition(conditions []string, condition string) []string {
	if condition == "" {
		return conditions
	}
	res := make([]string, 0, len(conditions)+1)
	res = append(res, conditions...)
	return append(res, condition)
}

// combineConditions combines templated boolean conditions with AND.
func combineConditions(conditions []string) string {
	switch len(conditions) {
	case 0:
		return ""
	case 1:
		return conditions[0]
	default:
		return "(" + strings.Join(conditions, ") AND (") + ")"
	}
}

// parseCanvasInlineComponent parses an inline component definition in a canvas item.
// posKey uniquely identifies the item's position in the canvas (including any tab path).
func (p *Parser) parseCanvasInlineComponent(canvasName, posKey string, props map[string]yaml.Node) (string, *componentDef, error) {
	var n yaml.Node
	err := n.Encode(props)
	if err != nil {
		return "", nil, fmt.Errorf("invalid component at %s: %w", posKey, err)
	}

	tmp := &ComponentYAML{}
	err = n.Decode(tmp)
	if err != nil {
		return "", nil, err
	}

	spec, refs, err := p.parseComponentYAML(tmp)
	if err != nil {
		return "", nil, err
	}

	spec.DefinedInCanvas = true

	name := fmt.Sprintf("%s--component-%s", canvasName, posKey)

	err = p.insertDryRun(ResourceKindComponent, name)
	if err != nil {
		name = fmt.Sprintf("%s--component-%s-%s", canvasName, posKey, uuid.New())
		err = p.insertDryRun(ResourceKindComponent, name)
		if err != nil {
			return "", nil, err
		}
	}

	def := &componentDef{
		name: name,
		refs: refs,
		spec: spec,
	}

	return name, def, nil
}

type componentDef struct {
	name string
	refs []ResourceName
	spec *runtimev1.ComponentSpec
}

// itemSizeRegex is used for parseItemSize.
var itemSizeRegex = regexp.MustCompile(`^(\d+)\s*(.*)$`)

// parseItemSize parses a string of the format "<int><space?><unit?>".
// Examples: "100", "100px", "100 px".
func parseItemSize(s string) (uint32, string, error) {
	if s == "" {
		return 0, "", nil
	}

	matches := itemSizeRegex.FindStringSubmatch(s)
	if matches == nil {
		return 0, "", fmt.Errorf("invalid size %q", s)
	}

	size, err := strconv.ParseUint(matches[1], 10, 32)
	if err != nil {
		return 0, "", fmt.Errorf("invalid size %q: %w", s, err)
	}

	return uint32(size), matches[2], nil
}

func pointerIfNotEmpty(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}

func parseFilterExpressions(filterMap map[string]string) (map[string]*runtimev1.DefaultMetricsSQLFilter, error) {
	result := make(map[string]*runtimev1.DefaultMetricsSQLFilter)

	for key, filterStr := range filterMap {
		expr, err := metricssql.ParseFilter(filterStr)
		if err != nil {
			return nil, fmt.Errorf("invalid filter expression for key %q: %w", key, err)
		}

		converted := metricsview.ExpressionToProto(expr)

		result[key] = &runtimev1.DefaultMetricsSQLFilter{
			Expression: converted,
		}
	}

	return result, nil
}
