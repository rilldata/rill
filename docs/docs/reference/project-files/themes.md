---
note: GENERATED. DO NOT EDIT.
title: Theme YAML
sidebar_position: 40
---

In your Rill project directory, create a `<theme_name>.yaml` file in any directory containing `type: theme`. Rill will automatically ingest the theme next time you run `rill start` or deploy to Rill Cloud.

To apply that theme to a dashboard, add `theme: <name of theme>` to the YAML file for that dashboard. Alternatively, you can add this to the end of the URL in your browser: `?theme=<name of theme>`


## Properties

### `type`

_[string]_ - Refers to the resource type and must be `theme` _(required)_

### `colors`

_[object]_ - Color palette for the theme

  - **`primary`** - _[string]_ - Primary color

  - **`secondary`** - _[string]_ - Secondary color

### `light`

_[object]_ - Light theme color configuration

  - **`primary`** - _[string]_ - Primary color for light theme. Can have any hex, [named colors](https://www.w3.org/TR/css-color-4/#named-colors) or hsl() formats.

  - **`secondary`** - _[string]_ - Secondary color for light theme. Can have any hex, [named colors](https://www.w3.org/TR/css-color-4/#named-colors) or hsl() formats.

  - **`kpi-positive`** - _[string]_ - Color for positive KPI delta values in light theme. Defaults to fg-secondary (gray).

  - **`kpi-negative`** - _[string]_ - Color for negative KPI delta values in light theme. Defaults to red.

  - **`variables`** - _[object]_ - CSS variables for the light theme, as a map from variable name to value. See [Variables](/reference/project-files/themes#variables) for the supported names. A variable can also be set directly in `light`, next to `primary`; if it is set in both places, the value under `variables` is used.

### `dark`

_[object]_ - Dark theme color configuration

  - **`primary`** - _[string]_ - Primary color for dark theme. Can have any hex, [named colors](https://www.w3.org/TR/css-color-4/#named-colors) or hsl() formats.

  - **`secondary`** - _[string]_ - Secondary color for dark theme. Can have any hex, [named colors](https://www.w3.org/TR/css-color-4/#named-colors) or hsl() formats.

  - **`kpi-positive`** - _[string]_ - Color for positive KPI delta values in dark theme. Defaults to fg-secondary (gray).

  - **`kpi-negative`** - _[string]_ - Color for negative KPI delta values in dark theme. Defaults to red.

  - **`variables`** - _[object]_ - CSS variables for the dark theme, as a map from variable name to value. See [Variables](/reference/project-files/themes#variables) for the supported names. A variable can also be set directly in `dark`, next to `primary`; if it is set in both places, the value under `variables` is used.

## Common Properties

### `name`

_[string]_ - Name is usually inferred from the filename, but can be specified manually.

### `refs`

_[array of string]_ - List of resource references

### `tags`

_[array of string]_ - Tags for organizing and filtering the resource (e.g. on the project dashboards list).

### `metadata`

_[object]_ - User-defined key-value metadata attached to the resource. Rill does not read or write it; it is exposed as-is on the resource's meta over the API for external tooling. Values are strings, with numbers and booleans coerced. Not supported in rill.yaml defaults.

### `dev`

_[object]_ - Overrides any properties in development environment.

### `prod`

_[object]_ - Overrides any properties in production environment.

## Variables

The variables below can be set in `light`, `dark` or both. Each one can be set under `variables` or directly in the block, next to `primary`. The two forms are equivalent and can be mixed; if a variable is set in both places, the value under `variables` is used.

Colors can be hex values, [named colors](https://www.w3.org/TR/css-color-4/#named-colors) or `hsl()` values. A variable that a theme does not set keeps the default listed here, and a variable set only in `light` keeps its default in dark mode.

### Brand colors

The two colors most themes start with. Rill generates a palette of shades from each, and uses those shades for hover and selected states, links, chips and other accents. They are usually set directly in the block, but are also accepted under `variables`.

| Variable | What it styles | Light default | Dark default |
| --- | --- | --- | --- |
| `primary` | Primary buttons, links, hover and selected states, dimension chips, and the lines, bars and areas of charts that show a single series. | `#5655ff` | `#5b5ef3` |
| `secondary` | Measure chips and, together with `primary`, the loading spinner. | `#04b8ff` | `#00a9eb` |

### Surfaces

Background colors, from the page to the elements layered on top of it.

| Variable | What it styles | Light default | Dark default |
| --- | --- | --- | --- |
| `surface-background` | Background of the dashboard itself: the canvas behind its components, the explore page and pivot table cells. | `#ffffff` | `#1e2125` |
| `surface-card` | Background of canvas components, including their header and hover toolbar. | `#ffffff` | `#1e2125` |
| `surface-subtle` | Background of the filter bar and of filter pills. | `#f9fafb` | `#17191c` |
| `surface-base` | Background of table rows in the pivot table and the time dimension detail table. | `#ffffff` | `#0d0e0e` |
| `surface-muted` | Background of pivot table header cells and neutral buttons. | `#f5f5f5` | `#2b3037` |
| `surface-overlay` | Background of dialogs and the time range picker. | `#ffffff` | `#24282e` |
| `surface-hover` | Background of buttons, menu items and table rows while hovered. | Shade 50 of `primary` | `#3e4651` |
| `surface-active` | Background of buttons while pressed, and of the excluded part of leaderboard bars. | Shade 100 of `primary` | `#2b3037` |

### Text

Text colors. If a theme sets `fg-primary` but not the others, Rill derives `fg-secondary`, `fg-tertiary`, `fg-muted` and `fg-disabled` from it by lowering its opacity, and uses it unchanged for `fg-inverse` and `fg-accent`.

| Variable | What it styles | Light default | Dark default |
| --- | --- | --- | --- |
| `fg-primary` | Main text: KPI values, leaderboard rows, pivot table cells, inputs and menu items. | `#0a0a0a` | `#e5e5e5` |
| `fg-secondary` | Supporting text: labels, component descriptions, filter text and the time axis of explore charts. | `#364153` | `#d1d5dc` |
| `fg-tertiary` | Labels of form fields. | `#6a7282` | `#99a1af` |
| `fg-muted` | Low-emphasis text: leaderboard headers, KPI comparison values, placeholders and the value axis of explore charts. | `#737373` | `#a1a1a1` |
| `fg-disabled` | Text of excluded leaderboard rows and disabled options. | `#a1a1a1` | `#525252` |
| `fg-inverse` | Text on tooltips and primary buttons. | `#fafafa` | `#fafafa` |
| `fg-accent` | Text of the selected item in menus and dropdowns. Follows `primary` unless set. | Shade 900 of `primary` | Shade 100 of `primary` |

### Accents and icons

Brand-colored emphasis and icon colors. The accents follow `primary` unless set.

| Variable | What it styles | Light default | Dark default |
| --- | --- | --- | --- |
| `accent-primary` | Fill of primary buttons. | Shade 500 of `primary` | Shade 600 of `primary` |
| `accent-primary-action` | Links, and the text of secondary and link-style buttons. | Shade 500 of `primary` | Shade 400 of `primary` |
| `icon-default` | Drag handles in the dimension and measure selectors. | `#404040` | `#fafafa` |
| `icon-muted` | Low-emphasis icons, such as those in pivot table chips and the settings button. | `#6a7282` | `#6a7282` |

### Borders, inputs and corners

Lines around and between elements, form fields and corner rounding.

| Variable | What it styles | Light default | Dark default |
| --- | --- | --- | --- |
| `border` | Borders and dividers: the outline of canvas components and KPI tiles, table borders and separators. | `#e5e5e5` | `#ffffff` at 8% opacity |
| `input` | Background of text inputs and search fields. | `#ffffff` at 5% opacity | `#31373f` at 50% opacity |
| `radius` | Corner radius, as a length in `px`, `rem` or `em` (`0` for square corners). Dialogs use the value as given; menus, popovers, tooltips and canvas tab groups use 2px less; canvas components, menu items, measure chips and the pivot table use 4px less. Buttons, inputs and dimension chips have a fixed shape. | `0.5rem` | `0.5rem` |

### KPIs

Colors of the change shown next to a value when a comparison period is on: in KPIs, big numbers, leaderboards and pivot tables.

| Variable | What it styles | Light default | Dark default |
| --- | --- | --- | --- |
| `kpi-positive` | Positive change. | `#364153` | `#d1d5dc` |
| `kpi-negative` | Negative change. | `#fb2c36` | `#ec2933` |

### Menus and tooltips

Elements that open on top of the dashboard.

| Variable | What it styles | Light default | Dark default |
| --- | --- | --- | --- |
| `popover` | Background of menus, dropdowns and popovers. | `#ffffff` | `#24282e` |
| `popover-foreground` | Text in menus, dropdowns and popovers. | `#0a0a0a` | `#fafafa` |
| `popover-accent` | Background of the highlighted item in a menu, and of leaderboard rows while hovered. Follows `surface-hover` unless set. | Shade 50 of `primary` | `#3e4651` |
| `popover-footer` | Background of the footer row of a menu, such as the one in the dimension filter. | `#fafafa` | `#24282e` |
| `tooltip` | Background of tooltips, including chart tooltips. | `#2b3037` | `#2b3037` |

### Errors and destructive actions

Colors of error states and of actions that delete or discard something.

| Variable | What it styles | Light default | Dark default |
| --- | --- | --- | --- |
| `destructive` | Error text, the outline of a canvas component that failed to load, and the fill of destructive buttons. | `#e7000a` | `#ff6467` at 90% opacity |
| `destructive-foreground` | Text on destructive buttons. | `#fef2f2` | `#fef2f2` |

### Sequential palette

Nine colors for values that run from low (`color-sequential-1`) to high (`color-sequential-9`). Used by heatmaps, map color scales and conditional formatting in pivot tables.

| Variable | Light default | Dark default |
| --- | --- | --- |
| `color-sequential-1` | `#e4effc` | `#33404d` |
| `color-sequential-2` | `#c5dbf2` | `#394d60` |
| `color-sequential-3` | `#98c3f0` | `#3e5974` |
| `color-sequential-4` | `#60a3e6` | `#42668a` |
| `color-sequential-5` | `#2388e1` | `#4573a1` |
| `color-sequential-6` | `#006edb` | `#4680b9` |
| `color-sequential-7` | `#0056b8` | `#538cc6` |
| `color-sequential-8` | `#004094` | `#6199d1` |
| `color-sequential-9` | `#002b70` | `#70a6db` |

### Diverging palette

Eleven colors for values that spread out from a midpoint (`color-diverging-6`) towards two extremes (`color-diverging-1` and `color-diverging-11`). Used in the same places as the sequential palette when a diverging color scheme is selected.

| Variable | Light default | Dark default |
| --- | --- | --- |
| `color-diverging-1` | `#e51029` | `#4d2e6b` |
| `color-diverging-2` | `#ff643d` | `#4c3181` |
| `color-diverging-3` | `#ffab66` | `#443399` |
| `color-diverging-4` | `#fde0a5` | `#3f3fa6` |
| `color-diverging-5` | `#f7f7e8` | `#4d5db3` |
| `color-diverging-6` | `#b8eeff` | `#647fb4` |
| `color-diverging-7` | `#7bd1f9` | `#88b34d` |
| `color-diverging-8` | `#36aaec` | `#a69e3f` |
| `color-diverging-9` | `#007be0` | `#996e33` |
| `color-diverging-10` | `#004cc7` | `#815331` |
| `color-diverging-11` | `#1b2193` | `#6b3d2e` |

### Qualitative palette

Twenty-four colors for categories. Used for chart series split by a dimension, the measures of a combo chart, and dimension comparisons in explores. Colors are assigned in order, so the first few matter most.

| Variable | Light default | Dark default |
| --- | --- | --- |
| `color-qualitative-1` | `#40c992` | `#39ac8c` |
| `color-qualitative-2` | `#ff9d4d` | `#ccb233` |
| `color-qualitative-3` | `#00b0eb` | `#3449b2` |
| `color-qualitative-4` | `#d99ce7` | `#bf4095` |
| `color-qualitative-5` | `#83d770` | `#39c668` |
| `color-qualitative-6` | `#f2c86e` | `#96c653` |
| `color-qualitative-7` | `#ed7d91` | `#b24934` |
| `color-qualitative-8` | `#6b90cc` | `#6039ac` |
| `color-qualitative-9` | `#00b88a` | `#34b29d` |
| `color-qualitative-10` | `#fb916a` | `#c69739` |
| `color-qualitative-11` | `#9297e8` | `#9540bf` |
| `color-qualitative-12` | `#bbce83` | `#59c059` |
| `color-qualitative-13` | `#e05299` | `#d14775` |
| `color-qualitative-14` | `#efe039` | `#cfdd3c` |
| `color-qualitative-15` | `#10c6c6` | `#26bbd9` |
| `color-qualitative-16` | `#a675d7` | `#b561d1` |
| `color-qualitative-17` | `#33cc66` | `#47d18c` |
| `color-qualitative-18` | `#f59e47` | `#e4a144` |
| `color-qualitative-19` | `#ed5e6a` | `#e05d52` |
| `color-qualitative-20` | `#478cd1` | `#5c85d6` |
| `color-qualitative-21` | `#22c3a8` | `#2dd2c5` |
| `color-qualitative-22` | `#e6bf4c` | `#d7ca42` |
| `color-qualitative-23` | `#8d7cde` | `#9e75d7` |
| `color-qualitative-24` | `#8bc95e` | `#74d161` |

### Names without an effect

These names are accepted, but no dashboard element reads them at present, so setting them changes nothing visible. Dimension and measure chips take their colors from `primary` and `secondary`.

`accent-secondary`, `accent-secondary-action`, `icon-accent`, `icon-disabled`, `ring-focus`, `ring-offset`, `dimension`, `dimension-foreground`, `dimension-border`, `measure`, `measure-foreground`, `measure-border`

### Deprecated names

Names from earlier versions of the theme format. They are still accepted: Rill sets the current variable instead, unless the theme also sets that variable itself.

| Deprecated name | Sets instead |
| --- | --- |
| `background` | `surface-subtle` |
| `surface` | `surface-background`, `surface-base` |
| `card` | `surface-card` |
| `muted` | `surface-muted` |
| `accent` | `popover-accent` |
| `foreground` | `fg-primary` |
| `card-foreground` | `fg-secondary` |
| `muted-foreground` | `fg-muted` |
| `accent-foreground` | `fg-accent` |
| `ring` | `ring-focus` |
| `primary-foreground` | Nothing: it no longer has an effect |
| `secondary-foreground` | Nothing: it no longer has an effect |

## Examples

```yaml
# Example: Basic theme with light and dark mode colors
type: theme
light:
    primary: "#4F46E5" # Indigo-600
    secondary: "#8B5CF6" # Purple-500
dark:
    primary: "#818CF8" # Indigo-400
    secondary: "#A78BFA" # Purple-400
```

```yaml
# Example: Branded surfaces, text, borders and corners
type: theme
light:
    primary: "#0F766E" # Teal-700
    secondary: "#F59E0B" # Amber-500
    variables:
        surface-background: "#F8FAFC" # Page behind the components
        surface-card: "#FFFFFF" # Canvas components
        fg-primary: "#0F172A" # Main text
        border: "#CBD5E1"
        radius: 1rem # Rounder corners; use 0 for square ones
dark:
    primary: "#2DD4BF" # Teal-400
    secondary: "#FBBF24" # Amber-400
    variables:
        surface-background: "#0B1120"
        surface-card: "#111827"
        fg-primary: "#F1F5F9"
        border: "#334155"
        radius: 1rem
```

```yaml
# Example: Custom chart color palettes
type: theme
light:
    primary: "#14B8A6" # Teal
    secondary: "#10B981" # Emerald
    variables:
        color-sequential-1: "hsl(180deg 80% 95%)"
        color-sequential-5: "hsl(180deg 80% 50%)"
        color-sequential-9: "hsl(180deg 80% 25%)"
        color-qualitative-1: "hsl(156deg 56% 52%)"
        color-qualitative-2: "hsl(27deg 100% 65%)"
dark:
    primary: "2DD4BF"
    secondary: "34D399"
    variables:
        color-sequential-1: "hsl(180deg 40% 30%)"
        color-sequential-5: "hsl(180deg 50% 50%)"
        color-sequential-9: "hsl(180deg 60% 70%)"
```

```yaml
# Example: KPI delta colors, with the variables set directly in the block instead of under `variables`
type: theme
light:
    primary: "#6366f1"
    secondary: "#8b5cf6"
    kpi-positive: "#16a34a" # Green for positive deltas (default: gray)
    kpi-negative: "#dc2626" # Red for negative deltas (default: red)
dark:
    primary: "#818cf8"
    secondary: "#a78bfa"
    kpi-positive: "#4ade80" # Brighter green for dark mode
    kpi-negative: "#f87171" # Brighter red for dark mode
```
