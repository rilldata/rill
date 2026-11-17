---
note: GENERATED. DO NOT EDIT.
title: Theme YAML
sidebar_position: 40
---

In your Rill project directory, create a `<theme_name>.yaml` file in any directory containing `type: theme`. Rill will automatically ingest the theme next time you run `rill start` or deploy to Rill Cloud.

To apply that theme to a dashboard, add `default_theme: <name of theme>` to the yaml file for that dashboard. Alternatively, you can add this to the end of the URL in your browser: `?theme=<name of theme>`


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

  - **`variables`** - _[object]_ - Custom CSS variables for light theme

### `dark`

_[object]_ - Dark theme color configuration

  - **`primary`** - _[string]_ - Primary color for dark theme. Can have any hex, [named colors](https://www.w3.org/TR/css-color-4/#named-colors) or hsl() formats.

  - **`secondary`** - _[string]_ - Secondary color for dark theme. Can have any hex, [named colors](https://www.w3.org/TR/css-color-4/#named-colors) or hsl() formats.

  - **`kpi-positive`** - _[string]_ - Color for positive KPI delta values in dark theme. Defaults to fg-secondary (gray).

  - **`kpi-negative`** - _[string]_ - Color for negative KPI delta values in dark theme. Defaults to red.

  - **`variables`** - _[object]_ - Custom CSS variables for dark theme

### `chart`

_[object]_ - Styling for charts. It applies to built-in canvas charts, to Vega-Lite charts and to custom charts. Properties set directly under `chart` apply in light and dark mode; `light` and `dark` override them for one mode. Anything left unset keeps Rill's default. A chart's own settings, such as a Vega-Lite spec's `config`, take precedence.

  - **`extends`** - _[string]_ - Name of a theme preset that ships with Flint, the compiler behind custom charts, to start from (for example `economist`, `nyt`, `swiss`, `material`). Only custom charts have presets; built-in and Vega-Lite charts ignore this property. The other `chart` properties override the preset.

  - **`font_family`** - _[string]_ - CSS font family for all chart text, for example `"IBM Plex Sans, sans-serif"`. The font must be available in the viewer's browser.

  - **`text`** - _[object]_ - Size, weight and color of chart text, by the role the text plays.

    - **`title`** - _[object]_ - Titles drawn inside a chart, such as the `title` of a Vega-Lite spec. A canvas widget's own title is not part of the chart.

      - **`size`** - _[number]_ - Font size in pixels.

      - **`weight`** - _[integer, string]_ - Font weight, as a multiple of 100 from 100 to 900 or a name (`thin`, `extralight`, `light`, `regular`, `medium`, `semibold`, `bold`, `extrabold`, `black`).

      - **`color`** - _[string]_ - Text color. Can have any hex, [named colors](https://www.w3.org/TR/css-color-4/#named-colors) or hsl() formats.

    - **`axis_label`** - _[object]_ - Axis tick labels and facet labels.

      - **`size`** - _[number]_ - Font size in pixels.

      - **`weight`** - _[integer, string]_ - Font weight, as a multiple of 100 from 100 to 900 or a name (`thin`, `extralight`, `light`, `regular`, `medium`, `semibold`, `bold`, `extrabold`, `black`).

      - **`color`** - _[string]_ - Text color. Can have any hex, [named colors](https://www.w3.org/TR/css-color-4/#named-colors) or hsl() formats.

    - **`axis_title`** - _[object]_ - Axis titles and facet titles.

      - **`size`** - _[number]_ - Font size in pixels.

      - **`weight`** - _[integer, string]_ - Font weight, as a multiple of 100 from 100 to 900 or a name (`thin`, `extralight`, `light`, `regular`, `medium`, `semibold`, `bold`, `extrabold`, `black`).

      - **`color`** - _[string]_ - Text color. Can have any hex, [named colors](https://www.w3.org/TR/css-color-4/#named-colors) or hsl() formats.

    - **`legend_label`** - _[object]_ - Legend entries.

      - **`size`** - _[number]_ - Font size in pixels.

      - **`weight`** - _[integer, string]_ - Font weight, as a multiple of 100 from 100 to 900 or a name (`thin`, `extralight`, `light`, `regular`, `medium`, `semibold`, `bold`, `extrabold`, `black`).

      - **`color`** - _[string]_ - Text color. Can have any hex, [named colors](https://www.w3.org/TR/css-color-4/#named-colors) or hsl() formats.

    - **`legend_title`** - _[object]_ - Legend titles.

      - **`size`** - _[number]_ - Font size in pixels.

      - **`weight`** - _[integer, string]_ - Font weight, as a multiple of 100 from 100 to 900 or a name (`thin`, `extralight`, `light`, `regular`, `medium`, `semibold`, `bold`, `extrabold`, `black`).

      - **`color`** - _[string]_ - Text color. Can have any hex, [named colors](https://www.w3.org/TR/css-color-4/#named-colors) or hsl() formats.

    - **`value_label`** - _[object]_ - Values printed on or beside marks.

      - **`size`** - _[number]_ - Font size in pixels.

      - **`weight`** - _[integer, string]_ - Font weight, as a multiple of 100 from 100 to 900 or a name (`thin`, `extralight`, `light`, `regular`, `medium`, `semibold`, `bold`, `extrabold`, `black`).

      - **`color`** - _[string]_ - Text color. Can have any hex, [named colors](https://www.w3.org/TR/css-color-4/#named-colors) or hsl() formats.

  - **`grid`** - _[object]_ - Gridlines.

    - **`measure`** - _[boolean]_ - Whether to draw the gridlines that values are read off, which run across the measure axis.

    - **`category`** - _[boolean]_ - Whether to draw the gridlines across the category or time axis.

    - **`style`** - _[string]_ - Line style, one of `solid`, `dashed` or `dotted`.

    - **`dash`** - _[array of number]_ - Alternating dash and gap lengths in pixels, for example `[4, 2]`. Overrides the dash pattern of `style`.

    - **`width`** - _[number]_ - Stroke width in pixels.

    - **`color`** - _[string]_ - Color of the gridlines. Can have any hex, [named colors](https://www.w3.org/TR/css-color-4/#named-colors) or hsl() formats.

  - **`axis`** - _[object]_ - Axis lines and tick marks. Axis text is styled under `text`.

    - **`line`** - _[boolean]_ - Whether to draw the axis line.

    - **`line_width`** - _[number]_ - Stroke width of the axis line in pixels.

    - **`ticks`** - _[boolean]_ - Whether to draw tick marks.

    - **`color`** - _[string]_ - Color of the axis line and tick marks.

  - **`line`** - _[object]_ - Line marks.

    - **`width`** - _[number]_ - Stroke width in pixels.

  - **`area`** - _[object]_ - Area marks.

    - **`fill_opacity`** - _[number]_ - Opacity of the fill, from 0 to 1.

    - **`line_width`** - _[number]_ - Stroke width of the line along the top of an area, in pixels.

  - **`bar`** - _[object]_ - Bar marks.

    - **`opacity`** - _[number]_ - Opacity of the bars, from 0 to 1.

    - **`corner_radius`** - _[number]_ - Radius of the corners at the value end of a bar, in pixels.

  - **`point`** - _[object]_ - Point marks, such as the dots of a scatter plot.

    - **`size`** - _[number]_ - Area of a point in square pixels.

    - **`opacity`** - _[number]_ - Opacity of the points, from 0 to 1.

  - **`legend`** - _[object]_ - Legends. Legend text is styled under `text`.

    - **`position`** - _[string]_ - Default legend position, one of `top`, `bottom`, `left` or `right`. A chart that sets its own legend position keeps it.

    - **`symbol`** - _[string]_ - Symbol drawn beside each legend entry, one of `circle`, `square` or `line`.

  - **`light`** - _[object]_ - Overrides for light mode. Takes the same properties as `chart`, except `light` and `dark`.

  - **`dark`** - _[object]_ - Overrides for dark mode. Takes the same properties as `chart`, except `light` and `dark`.

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
# Example: Advanced theme with custom color palettes
type: theme
light:
    primary: "#14B8A6" # Teal
    secondary: "#10B981" # Emerald
    color-sequential-1: "hsl(180deg 80% 95%)"
    color-sequential-5: "hsl(180deg 80% 50%)"
    color-sequential-9: "hsl(180deg 80% 25%)"
    color-qualitative-1: "hsl(156deg 56% 52%)"
    color-qualitative-2: "hsl(27deg 100% 65%)"
dark:
    primary: "2DD4BF"
    secondary: "34D399"
    color-sequential-1: "hsl(180deg 40% 30%)"
    color-sequential-5: "hsl(180deg 50% 50%)"
    color-sequential-9: "hsl(180deg 60% 70%)"
```

```yaml
# Example: Theme with KPI delta colors (green for positive, red for negative)
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

```yaml
# Example: Theme with chart styling
type: theme
light:
    primary: "#0f62fe"
    secondary: "#8a3ffc"
    color-qualitative-1: "#0f62fe"
    color-qualitative-2: "#ee5396"
    color-qualitative-3: "#08bdba"
    color-qualitative-4: "#f1c21b"
dark:
    primary: "#78a9ff"
    secondary: "#be95ff"
    color-qualitative-1: "#78a9ff"
    color-qualitative-2: "#ff7eb6"
    color-qualitative-3: "#3ddbd9"
    color-qualitative-4: "#fddc69"
chart:
    font_family: "IBM Plex Sans, sans-serif"
    text:
        axis_label:
            size: 12
            weight: medium
        axis_title:
            size: 12
            weight: semibold
        legend_label:
            size: 12
    grid:
        style: dashed
        dash: [4, 4]
        width: 1
    axis:
        line: true
        ticks: true
    line:
        width: 2.5
    area:
        fill_opacity: 0.25
        line_width: 2
    bar:
        corner_radius: 2
    legend:
        position: bottom
        symbol: line
    light: # Colors for light mode
        text:
            axis_label:
                color: "#525252"
            axis_title:
                color: "#161616"
        grid:
            color: "#e0e0e0"
        axis:
            color: "#8d8d8d"
    dark: # Colors for dark mode
        text:
            axis_label:
                color: "#c6c6c6"
            axis_title:
                color: "#f4f4f4"
        grid:
            color: "#393939"
        axis:
            color: "#6f6f6f"
```
