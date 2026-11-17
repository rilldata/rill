---
title: "Customization & Themes"
description: Alter dashboard look and feel
sidebar_label: "Customization & Themes"
sidebar_position: 30
---

Below are some common customizations and dashboard configurations available for end users. 

:::info Dashboard properties

For a full list of available dashboard properties and configurations, please see our [Dashboard YAML](/reference/project-files/explore-dashboards) reference page.

:::

## Define Dashboard Access

Along with [metrics views security policies](/developers/build/metrics-view/security), you can set access on the dashboard level. Access policies will be combined with metrics view policies using a logical AND, so if a user doesn’t pass both, they won’t get access to the dashboard.  Only the `access` key can be set in the dashboard.

```yaml
security:
  access: "{{ .user.admin }} OR '{{ .user.domain }}' == 'example.com'"
```


## Setting Default Views for Dashboards
### Default Time Range

Default time range controls the data analyzed on initial page load. Setting the default time range improves user experience by setting it to the most frequently used period— in particular, avoiding `all time` if you have a large data source but only analyze more recent data.

The value must be either a valid [ISO 8601 duration](https://en.wikipedia.org/wiki/ISO_8601#Durations) (for example, `PT12H` for 12 hours, `P1M` for 1 month, or `P26W` for 26 weeks) or one of the [Rill ISO 8601 extensions](/reference/time-syntax/rill-iso-extensions#extensions).


### Default Dimensions

For dashboards with wide tables, setting default dimensions is a good way to make sure that users can focus on the primary analyses and ensure a positive first experience. Each dimension listed under the `dimensions` setting will appear on the screen, while the remainder of the dimensions will be hidden (and still available for selection under filters). Common use cases for setting default fields include simplifying dashboards on initial load, narrowing the dashboard to the most used fields, and avoiding high cardinality fields (that may take longer to load, but are used less often, so this improves performance). An example addition to the `dashboard.yaml` file is below.

```yaml
defaults:
  dimensions:
    - column1
    - column2
```

:::warning Column vs. Name Usage
The `column` property is used by default from the column name in your underlying source. If you decide to use the `name` property, you'd replace the column above with the field name.
:::

### Default Measures

A list of measures that should be visible by default. Operates the same as the `default_dimensions` configuration. When selecting measures, by default, consider hiding more computationally intensive measures like `count distinct` or other complicated expressions to improve performance.

```yaml
defaults:
  measures:
    - measure_1
    - measure_1
```

### Default Comparison Modes

It is also possible to set up a default comparison mode for your dashboard. In Rill, we support both [time comparison](/guide/dashboards/time-series#time-comparisons) and [dimension comparison.](/guide/dashboards/filters#filter-by-dimensions) Note that only one of these comparisons can be set as default. 

```yaml
defaults:
  comparison_mode: time
  # comparison_mode: dimension
  # comparison_dimension: action

```



## Time Ranges

One of the more important configurations, available time ranges allow you to change the defaults in the time dropdown for periods to select. Updating this list allows users to quickly change between the most common analyses, like day over day, recent weeks, or period to date. The range must be a valid [ISO 8601 duration](https://en.wikipedia.org/wiki/ISO_8601#Durations) or one of the [Rill ISO 8601 extensions](/reference/time-syntax/rill-iso-extensions#extensions).

```yaml
time_ranges:
  - PT15M 
  - PT1H
  - P7D
  - P4W
  - rill-TD ## Today
  - rill-WTD ## Week-To-date
```

## Time Zones

Rill will automatically select several time zones that should be pinned to the top of the time zone selector. It should be a list of [IANA time zone identifiers](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones). You can add or remove relevant time zones for your team from this list.

```yaml
time_zones:
  - America/Los_Angeles
  - America/Chicago
  - America/New_York
  - Europe/London
  - Europe/Paris
  - Asia/Jerusalem
  - Europe/Moscow
  - Asia/Kolkata
  - Asia/Shanghai
  - Asia/Tokyo
  - Australia/Sydney  
```

## Changing Themes & Colors

In your Rill project directory, create a `<theme_name>.yaml` file in any directory. Rill will automatically ingest the theme next time you run `rill start` or deploy to Rill Cloud and change the color scheme of your dashboard. All properties in the file are optional—any values you omit will fall back to Rill's standard theme defaults.

First, create the YAML file as below. You can define separate color schemes for light and dark modes:

```yaml
type: theme
light:
  primary: "#4F46E5"  # Indigo-600
  secondary: "#8B5CF6"  # Purple-500
  variables:
    # Sequential palette - for ordered data that progresses from low to high
    color-sequential-1: "hsl(211deg 79% 94%)"
    color-sequential-2: "hsl(211deg 63% 86%)"
    color-sequential-3: "hsl(211deg 75% 77%)"
    color-sequential-4: "hsl(210deg 73% 64%)"
    color-sequential-5: "hsl(208deg 76% 51%)"
    color-sequential-6: "hsl(210deg 100% 43%)"
    color-sequential-7: "hsl(212deg 100% 36%)"
    color-sequential-8: "hsl(214deg 100% 29%)"
    color-sequential-9: "hsl(217deg 100% 22%)"

    # Diverging palette - for data that diverges from a critical midpoint
    color-diverging-1: "hsl(353deg 87% 48%)"
    color-diverging-2: "hsl(12deg 100% 62%)"
    color-diverging-3: "hsl(27deg 100% 70%)"
    color-diverging-4: "hsl(40deg 96% 82%)"
    color-diverging-5: "hsl(59deg 48% 94%)"
    color-diverging-6: "hsl(194deg 100% 86%)"
    color-diverging-7: "hsl(199deg 91% 73%)"
    color-diverging-8: "hsl(202deg 83% 57%)"
    color-diverging-9: "hsl(207deg 100% 44%)"
    color-diverging-10: "hsl(217deg 100% 39%)"
    color-diverging-11: "hsl(237deg 69% 34%)"

    # Qualitative palette - for categorical data (showing first 12 of 24)
    color-qualitative-1: "hsl(156deg 56% 52%)"
    color-qualitative-2: "hsl(27deg 100% 65%)"
    color-qualitative-3: "hsl(195deg 100% 46%)"
    color-qualitative-4: "hsl(289deg 61% 76%)"
    color-qualitative-5: "hsl(109deg 56% 64%)"
    color-qualitative-6: "hsl(41deg 83% 69%)"
    color-qualitative-7: "hsl(349deg 76% 71%)"
    color-qualitative-8: "hsl(217deg 49% 61%)"
    color-qualitative-9: "hsl(165deg 100% 36%)"
    color-qualitative-10: "hsl(16deg 95% 70%)"
    color-qualitative-11: "hsl(236deg 65% 74%)"
    color-qualitative-12: "hsl(75deg 43% 66%)"
    # ... up to color-qualitative-24

dark:
  primary: "#818CF8"  # Indigo-400
  secondary: "#A78BFA"  # Purple-400
  variables:
    # Sequential palette - adjusted for dark backgrounds
    color-sequential-1: "hsl(210deg 20% 25%)"
    color-sequential-2: "hsl(210deg 25% 30%)"
    color-sequential-3: "hsl(210deg 30% 35%)"
    color-sequential-4: "hsl(210deg 35% 40%)"
    color-sequential-5: "hsl(210deg 40% 45%)"
    color-sequential-6: "hsl(210deg 45% 50%)"
    color-sequential-7: "hsl(210deg 50% 55%)"
    color-sequential-8: "hsl(210deg 55% 60%)"
    color-sequential-9: "hsl(210deg 60% 65%)"
    # ... (diverging and qualitative palettes also available)
```

The `light` and `dark` properties allow you to customize:
- **primary**: Primary color used for charts, buttons, and interactive elements
- **secondary**: Secondary color (used for loading spinners and accents)
- **variables**: Custom CSS variables for complete control over color palettes
  - **Sequential palette** (color-sequential-1 through 9): For ordered data progressions
  - **Diverging palette** (color-diverging-1 through 11): For data diverging from a midpoint
  - **Qualitative palette** (color-qualitative-1 through 24): For categorical data

Colors can be specified using:
- Hex values (with or without the '#' character, e.g., `4F46E5`. `"#4F46E5"`)
- Named colors (e.g., `plum`, `violet`)
- HSL format (e.g., `hsl(180, 100%, 50%)`)

Once you have created that file, update the `dashboard.yaml` with the following configuration (we typically add this at the top along with time zones, time series and other configurations):

`theme: <name of theme yaml file>`

:::info Theme properties

For more details about configuring themes, you can refer to our [Theme YAML](/reference/project-files/themes) reference page.

:::

### Styling Charts

The `chart` section of a theme sets how charts are drawn, so you style them once instead of per chart. It applies to every chart on a canvas dashboard that uses the theme: the [chart widgets](/developers/build/dashboards/canvas-widgets/chart), [custom viz components](/developers/build/dashboards/canvas-widgets/custom-viz) and [Vega-Lite charts](/developers/build/dashboards/canvas-widgets/custom-viz/vega-lite).

For example, this theme draws dashed gridlines and thicker lines:

```yaml
type: theme
chart:
  grid:
    style: dashed
  line:
    width: 2.5
```

Every property is optional, and a chart keeps Rill's default for anything the theme leaves out.

| Property | What it styles |
|---|---|
| `font_family` | The font of all chart text. The font must be available in the viewer's browser. |
| `text` | The `size`, `weight` and `color` of text, by role: `axis_label`, `axis_title`, `legend_label`, `legend_title`, `value_label` (values printed on marks) and `title` (a title drawn inside a chart; a widget's own title is not part of the chart). |
| `grid` | Gridlines: their `style` (`solid`, `dashed` or `dotted`), `dash` pattern, `width` and `color`, and which ones are drawn (`measure`, `category`). |
| `axis` | Whether an axis has a `line` and `ticks`, and their `line_width` and `color`. |
| `line` | The `width` of lines. |
| `area` | The `fill_opacity` of areas, and the `line_width` of the line along their top. |
| `bar` | The `opacity` and `corner_radius` of bars. |
| `point` | The `size` and `opacity` of points. |
| `legend` | The default `position` of legends and the `symbol` beside each entry. |

Gridlines are named for what their axis holds rather than where it sits. `grid.measure` draws the lines that values are read off, which are horizontal on a line chart and vertical on a horizontal bar chart. `grid.category` draws the lines across the category or time axis.

#### Light and Dark Mode

Properties set directly under `chart` apply in both modes. Put what differs, usually the colors, under `chart.light` and `chart.dark`. Both take the same properties as `chart`:

```yaml
type: theme
chart:
  grid:
    style: dashed
  light:
    grid:
      color: "#e0e0e0"
  dark:
    grid:
      color: "#393939"
```

#### A Complete Theme

This theme sets brand colors and palettes for each mode, then styles the charts drawn with them:

```yaml
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
  light:
    text:
      axis_label:
        color: "#525252"
      axis_title:
        color: "#161616"
    grid:
      color: "#e0e0e0"
    axis:
      color: "#8d8d8d"
  dark:
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

#### What a Chart Sets Itself

The theme provides defaults, and a chart's own settings take precedence over them:

- A chart widget that sets a legend position keeps it; `legend.position` only places the legends of charts that don't.
- A chart widget's `vl_config` and a Vega-Lite chart's `config` override the theme property by property, so a spec can restyle one thing and take the rest from the theme.
- Series colors come from the theme's `color-qualitative-*`, `color-sequential-*` and `color-diverging-*` palettes, unless a chart maps its own colors.

#### Presets for Custom Viz

Custom viz components are compiled by [Flint](https://github.com/microsoft/flint-chart), which ships a set of complete chart designs. `extends` starts from one of them, and the other `chart` properties go over it:

```yaml
type: theme
chart:
  extends: economist
  line:
    width: 3
```

The presets are `nyt`, `economist`, `swiss`, `nature`, `mckinsey`, `datawrapper`, `powerbi`, `powerbi-light`, `pop`, `cartoon`, `apple`, `de-stijl`, `bbc` and `material`. A custom viz reports an error if the theme names a preset that Flint does not have.

A preset is a design of its own: a custom viz drawn with one keeps the preset's fonts, surface, palette and legend layout, except where the theme sets a `chart` property or a `color-*` palette. Chart widgets and Vega-Lite charts have no presets, so they ignore `extends` and take only the other `chart` properties. A preset has one look for both modes; to use one in a single mode, or a different one in each, set `extends` under `chart.light` or `chart.dark`.

The [Theme YAML](/reference/project-files/themes#chart) reference lists every `chart` property.

## Example

```yaml
# Explore YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/explores

type: explore

display_name: "Programmatic Ads Auction"
metrics_view: auction_metrics

dimensions:
  expr: "*"
measures:
  - requests
  - avg_bid_floor
  - 1d_qps

defaults:
  measures:
    - avg_bid_floor
    - requests
  dimensions:
    - app_site_cat
    - app_site_domain
    - app_site_name
    - pub_name
  comparison_mode: time
  time_range: P7D

time_ranges:
  - rill-TD
  - rill-WTD
  - rill-MTD
  - rill-QTD
  - rill-YTD
  - rill-PDC
  - rill-PWC
  - rill-PMC
  - rill-PQC
  - rill-PYC
theme:
  light:
    primary: "#14B8A6"  # Teal-500
    secondary: "#10B981"  # Emerald-500
    variables:
      color-sequential-1: "hsl(180deg 80% 95%)"
      color-sequential-5: "hsl(180deg 80% 50%)"
      color-sequential-9: "hsl(180deg 80% 25%)"
  dark:
    primary: "#2DD4BF"  # Teal-400
    secondary: "#34D399"  # Emerald-400

security:
    access: "{{ .user.admin }} OR '{{ .user.domain }}' == 'example.com'"  # only access can be set on dashboard level, see metric view for detailed access policies '{{ .user.domain }}' == 'example.com'"  # only access can be set on dashboard level, see metric view for detailed access policies
```