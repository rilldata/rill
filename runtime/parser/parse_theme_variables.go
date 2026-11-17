package parser

import "fmt"

// ThemeVariable describes a CSS variable that a theme can set in its `light` and `dark` blocks.
// The descriptions are user-facing: the theme reference docs are generated from them.
type ThemeVariable struct {
	// Name of the variable as written in the theme YAML.
	Name string
	// Description of what the variable styles in a dashboard.
	Description string
	// Length is true if the value is a CSS length instead of a color.
	Length bool
	// Unused is true for names that are accepted, but that no dashboard element reads at present.
	Unused bool
	// Deprecated is true for names from earlier versions of the theme format.
	Deprecated bool
	// ReplacedBy lists the variables that a deprecated name sets instead.
	ReplacedBy []string
}

// ThemeVariableGroup is a set of related theme variables that are documented together.
type ThemeVariableGroup struct {
	Title       string
	Description string
	Variables   []ThemeVariable
}

// ThemeVariableGroups lists every CSS variable that a theme can set.
// It is the allow-list the parser validates against, and the source of the variable reference in the theme docs:
// run `make docs.generate` after changing it.
var ThemeVariableGroups = []ThemeVariableGroup{
	{
		Title:       "Brand colors",
		Description: "The two colors most themes start with. Rill generates a palette of shades from each, and uses those shades for hover and selected states, links, chips and other accents. They are usually set directly in the block, but are also accepted under `variables`.",
		Variables: []ThemeVariable{
			{Name: "primary", Description: "Primary buttons, links, hover and selected states, dimension chips, and the lines, bars and areas of charts that show a single series."},
			{Name: "secondary", Description: "Measure chips and, together with `primary`, the loading spinner."},
		},
	},
	{
		Title:       "Surfaces",
		Description: "Background colors, from the page to the elements layered on top of it.",
		Variables: []ThemeVariable{
			{Name: "surface-background", Description: "Background of the dashboard itself: the canvas behind its components, the explore page and pivot table cells."},
			{Name: "surface-card", Description: "Background of canvas components, including their header and hover toolbar."},
			{Name: "surface-subtle", Description: "Background of the filter bar and of filter pills."},
			{Name: "surface-base", Description: "Background of table rows in the pivot table and the time dimension detail table."},
			{Name: "surface-muted", Description: "Background of pivot table header cells and neutral buttons."},
			{Name: "surface-overlay", Description: "Background of dialogs and the time range picker."},
			{Name: "surface-hover", Description: "Background of buttons, menu items and table rows while hovered."},
			{Name: "surface-active", Description: "Background of buttons while pressed, and of the excluded part of leaderboard bars."},
		},
	},
	{
		Title:       "Text",
		Description: "Text colors. If a theme sets `fg-primary` but not the others, Rill derives `fg-secondary`, `fg-tertiary`, `fg-muted` and `fg-disabled` from it by lowering its opacity, and uses it unchanged for `fg-inverse` and `fg-accent`.",
		Variables: []ThemeVariable{
			{Name: "fg-primary", Description: "Main text: KPI values, leaderboard rows, pivot table cells, inputs and menu items."},
			{Name: "fg-secondary", Description: "Supporting text: labels, component descriptions, filter text and the time axis of explore charts."},
			{Name: "fg-tertiary", Description: "Labels of form fields."},
			{Name: "fg-muted", Description: "Low-emphasis text: leaderboard headers, KPI comparison values, placeholders and the value axis of explore charts."},
			{Name: "fg-disabled", Description: "Text of excluded leaderboard rows and disabled options."},
			{Name: "fg-inverse", Description: "Text on tooltips and primary buttons."},
			{Name: "fg-accent", Description: "Text of the selected item in menus and dropdowns. Follows `primary` unless set."},
		},
	},
	{
		Title:       "Accents and icons",
		Description: "Brand-colored emphasis and icon colors. The accents follow `primary` unless set.",
		Variables: []ThemeVariable{
			{Name: "accent-primary", Description: "Fill of primary buttons."},
			{Name: "accent-primary-action", Description: "Links, and the text of secondary and link-style buttons."},
			{Name: "icon-default", Description: "Drag handles in the dimension and measure selectors."},
			{Name: "icon-muted", Description: "Low-emphasis icons, such as those in pivot table chips and the settings button."},
		},
	},
	{
		Title:       "Borders, inputs and corners",
		Description: "Lines around and between elements, form fields and corner rounding.",
		Variables: []ThemeVariable{
			{Name: "border", Description: "Borders and dividers: the outline of canvas components and KPI tiles, table borders and separators."},
			{Name: "input", Description: "Background of text inputs and search fields."},
			{Name: "radius", Length: true, Description: "Corner radius, as a length in `px`, `rem` or `em` (`0` for square corners). Dialogs use the value as given; menus, popovers, tooltips and canvas tab groups use 2px less; canvas components, menu items, measure chips and the pivot table use 4px less. Buttons, inputs and dimension chips have a fixed shape."},
		},
	},
	{
		Title:       "KPIs",
		Description: "Colors of the change shown next to a value when a comparison period is on: in KPIs, big numbers, leaderboards and pivot tables.",
		Variables: []ThemeVariable{
			{Name: "kpi-positive", Description: "Positive change."},
			{Name: "kpi-negative", Description: "Negative change."},
		},
	},
	{
		Title:       "Menus and tooltips",
		Description: "Elements that open on top of the dashboard.",
		Variables: []ThemeVariable{
			{Name: "popover", Description: "Background of menus, dropdowns and popovers."},
			{Name: "popover-foreground", Description: "Text in menus, dropdowns and popovers."},
			{Name: "popover-accent", Description: "Background of the highlighted item in a menu, and of leaderboard rows while hovered. Follows `surface-hover` unless set."},
			{Name: "popover-footer", Description: "Background of the footer row of a menu, such as the one in the dimension filter."},
			{Name: "tooltip", Description: "Background of tooltips, including chart tooltips."},
		},
	},
	{
		Title:       "Errors and destructive actions",
		Description: "Colors of error states and of actions that delete or discard something.",
		Variables: []ThemeVariable{
			{Name: "destructive", Description: "Error text, the outline of a canvas component that failed to load, and the fill of destructive buttons."},
			{Name: "destructive-foreground", Description: "Text on destructive buttons."},
		},
	},
	{
		Title:       "Sequential palette",
		Description: "Nine colors for values that run from low (`color-sequential-1`) to high (`color-sequential-9`). Used by heatmaps, map color scales and conditional formatting in pivot tables.",
		Variables:   themePaletteVariables("color-sequential", 9),
	},
	{
		Title:       "Diverging palette",
		Description: "Eleven colors for values that spread out from a midpoint (`color-diverging-6`) towards two extremes (`color-diverging-1` and `color-diverging-11`). Used in the same places as the sequential palette when a diverging color scheme is selected.",
		Variables:   themePaletteVariables("color-diverging", 11),
	},
	{
		Title:       "Qualitative palette",
		Description: "Twenty-four colors for categories. Used for chart series split by a dimension, the measures of a combo chart, and dimension comparisons in explores. Colors are assigned in order, so the first few matter most.",
		Variables:   themePaletteVariables("color-qualitative", 24),
	},
	{
		Title:       "Names without an effect",
		Description: "These names are accepted, but no dashboard element reads them at present, so setting them changes nothing visible. Dimension and measure chips take their colors from `primary` and `secondary`.",
		Variables: []ThemeVariable{
			{Name: "accent-secondary", Unused: true},
			{Name: "accent-secondary-action", Unused: true},
			{Name: "icon-accent", Unused: true},
			{Name: "icon-disabled", Unused: true},
			{Name: "ring-focus", Unused: true},
			{Name: "ring-offset", Unused: true},
			{Name: "dimension", Unused: true},
			{Name: "dimension-foreground", Unused: true},
			{Name: "dimension-border", Unused: true},
			{Name: "measure", Unused: true},
			{Name: "measure-foreground", Unused: true},
			{Name: "measure-border", Unused: true},
		},
	},
	{
		Title:       "Deprecated names",
		Description: "Names from earlier versions of the theme format. They are still accepted: Rill sets the current variable instead, unless the theme also sets that variable itself.",
		Variables: []ThemeVariable{
			{Name: "background", Deprecated: true, ReplacedBy: []string{"surface-subtle"}},
			{Name: "surface", Deprecated: true, ReplacedBy: []string{"surface-background", "surface-base"}},
			{Name: "card", Deprecated: true, ReplacedBy: []string{"surface-card"}},
			{Name: "muted", Deprecated: true, ReplacedBy: []string{"surface-muted"}},
			{Name: "accent", Deprecated: true, ReplacedBy: []string{"popover-accent"}},
			{Name: "foreground", Deprecated: true, ReplacedBy: []string{"fg-primary"}},
			{Name: "card-foreground", Deprecated: true, ReplacedBy: []string{"fg-secondary"}},
			{Name: "muted-foreground", Deprecated: true, ReplacedBy: []string{"fg-muted"}},
			{Name: "accent-foreground", Deprecated: true, ReplacedBy: []string{"fg-accent"}},
			{Name: "ring", Deprecated: true, ReplacedBy: []string{"ring-focus"}},
			{Name: "primary-foreground", Deprecated: true},
			{Name: "secondary-foreground", Deprecated: true},
		},
	},
}

// themeVariables indexes ThemeVariableGroups by variable name.
var themeVariables = func() map[string]ThemeVariable {
	res := make(map[string]ThemeVariable)
	for _, g := range ThemeVariableGroups {
		for _, v := range g.Variables {
			if _, ok := res[v.Name]; ok {
				panic(fmt.Sprintf("theme variable %q is listed twice", v.Name))
			}
			res[v.Name] = v
		}
	}
	return res
}()

// themePaletteVariables returns the variables for a chart palette with n numbered colors.
func themePaletteVariables(prefix string, n int) []ThemeVariable {
	res := make([]ThemeVariable, n)
	for i := range res {
		res[i] = ThemeVariable{Name: fmt.Sprintf("%s-%d", prefix, i+1)}
	}
	return res
}
