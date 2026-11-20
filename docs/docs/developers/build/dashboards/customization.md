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

Along with [metrics views security policies](/developers/build/metrics-view/security), you can set access on the dashboard level. Access policies will be combined with metrics view policies using a logical AND, so if a user doesn’t pass both, they won’t get access to the dashboard. Only the `access` key can be set in the dashboard.

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

### Default Filters

A default filter is applied when a viewer first opens the dashboard, and again when they go back to its default view. Set it under `defaults.filters` as a [Metrics SQL](/developers/build/metrics-view/metrics-sql) filter expression, written the way it would follow `WHERE`:

```yaml
defaults:
  filters: "country = 'US' AND device IN ('mobile', 'tablet')"
```

The expression can only reference dimensions and measures shown in the dashboard. Viewers can change or clear the default filter. A shared link or [bookmark](/guide/dashboards/bookmarks) opens the dashboard in the state it was saved with, including its filters, so the default filter is not applied on top of it.

An explore dashboard has a single metrics view, so `defaults.filters` is one expression. On a [canvas dashboard](/guide/dashboards/canvas), `defaults.filters` is instead a map with an expression per metrics view.

To set a default [measure filter](/guide/dashboards/filters#filter-by-measures), select the dimension values to keep with a subquery. This example keeps the publishers with more than 1,000 in total spend:

```yaml
defaults:
  filters: "publisher IN (SELECT publisher FROM metrics_view HAVING total_spend > 1000)"
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

## Pinned and Required Filters

You can keep specific filters in the filter bar, and hold the dashboard back until the viewer picks a value for them. This is useful when a dashboard only makes sense for one customer, account or region at a time.

| Setting      | Effect for viewers                                                                                                                 |
| ------------ | ---------------------------------------------------------------------------------------------------------------------------------- |
| **Pinned**   | The filter always appears in the filter bar, even when it has no value. Viewers can change its value but cannot remove the filter. |
| **Required** | The dashboard does not load until the filter has a value. A required filter is always pinned.                                      |

List dimension or measure names under `filters`:

```yaml
filters:
  pinned:
    - device
  required:
    - advertiser
```

Each name must be a measure or a dimension shown in the dashboard, other than a time dimension. If it is not, the dashboard fails to reconcile with an error like:

```
required filter "advertiser" is not a dimension or measure in this explore
```

Until every required filter has a value, the leaderboards, charts, tables and pivot are replaced by a **Select a value to continue** message that lists the missing filters. **Clear filters** removes the values but keeps the pinned and required filters in the filter bar, so clearing a required filter blocks the dashboard again. Alerts cannot be created from the dashboard while a required filter has no value.

To open the dashboard with a value already selected, combine a required filter with a [default filter](#default-filters):

```yaml
filters:
  required:
    - advertiser
defaults:
  filters: "advertiser = 'Acme'"
```

The same `filters` and `defaults.filters` properties work in the `explore` block of a [metrics view](/reference/project-files/metrics-views).

### Filter Order

Pinned and required filters appear in the filter bar in the order you list them, so you can arrange them by hierarchy:

```yaml
filters:
  pinned:
    - region
    - country
    - city
```

Required filters that are not also listed under `pinned` come first, in the order of `required`. To show a required filter elsewhere, list it under `pinned` as well: its position in `pinned` then decides where it appears. Filters that viewers add come after the pinned and required ones, sorted by name. Measure filters always appear after dimension filters.

### Configure Filters in the Editor

When you edit an explore dashboard in Rill Developer, each filter's dropdown shows two extra controls in its top-right corner: the **asterisk** marks the filter as required, and the **pin** icon pins it. These controls change only your current editing session. To write them to the file, click **Save dashboard state as default** in the editing sidebar. This saves the pinned and required filters to `filters.pinned` and `filters.required`, and saves the current filter, time range, comparison mode and visible fields as the dashboard's `defaults`. Pinning a filter in the editor adds it to the end of `filters.pinned`; to change the order, edit the list in the file.

:::note Required filters are not access control
Pinned and required filters shape what viewers see in the dashboard. They do not restrict the data a viewer can query, for example through a report or the API. To limit which rows a user can access, use [security policies](/developers/build/metrics-view/security).
:::

## Single-Select Filters

Some dashboards should only ever be viewed for one value of a dimension at a time, for example one advertiser or one employee, because their measures are not meaningful across several. List those dimensions under `filters.single_select`:

```yaml
filters:
  single_select:
    - advertiser
```

The filter for each listed dimension holds one value. Picking a value in the filter bar, a leaderboard or the dimension table replaces the current one, and viewers cannot exclude values or match on text. A link or bookmark with several values opens with only the last of them. Each name must be a dimension shown in the explore, other than a time dimension.

To always show exactly one value, combine it with a [required filter](#pinned-and-required-filters) and a default value:

```yaml
filters:
  required:
    - advertiser
  single_select:
    - advertiser
defaults:
  filters: "advertiser = 'Acme'"
```

Single-select is a dashboard setting, not an access restriction: use [row access policies](/developers/build/metrics-view/security) to limit which data a viewer can query.

## Changing Themes & Colors

In your Rill project directory, create a `<theme_name>.yaml` file in any directory. Rill will automatically ingest the theme next time you run `rill start` or deploy to Rill Cloud and change the color scheme of your dashboard. All properties in the file are optional—any values you omit will fall back to Rill's standard theme defaults.

First, create the YAML file as below. You can define separate colors for light and dark modes:

```yaml
type: theme
light:
  primary: "#4F46E5"  # Indigo-600
  secondary: "#8B5CF6"  # Purple-500
dark:
  primary: "#818CF8"  # Indigo-400
  secondary: "#A78BFA"  # Purple-400
```

The `light` and `dark` properties allow you to customize:
- **primary**: Primary color used for buttons, links, selected states, and charts that show a single series
- **secondary**: Secondary color (used for measure chips and the loading spinner)
- **variables**: CSS variables for finer control over backgrounds, text, borders, corners, and chart color palettes (see [Theme Variables](#theme-variables))

Colors can be specified using:
- Hex values (with or without the '#' character, e.g., `4F46E5`. `"#4F46E5"`)
- Named colors (e.g., `plum`, `violet`)
- HSL format (e.g., `hsl(180, 100%, 50%)`)

Once you have created that file, update the `dashboard.yaml` with the following configuration (we typically add this at the top along with time zones, time series and other configurations):

`theme: <name of theme yaml file>`

### Theme Variables

Beyond the two brand colors, a theme can set variables for the dashboard's backgrounds, text, borders, and corner rounding. Add them under `variables` in the `light` and `dark` blocks:

```yaml
type: theme
light:
  primary: "#0F766E"  # Teal-700
  secondary: "#F59E0B"  # Amber-500
  variables:
    surface-background: "#F8FAFC"  # Page behind the components
    surface-card: "#FFFFFF"  # Canvas components
    fg-primary: "#0F172A"  # Main text
    border: "#CBD5E1"  # Borders and dividers
    radius: 1rem  # Corner radius; use 0 for square corners
dark:
  primary: "#2DD4BF"  # Teal-400
  secondary: "#FBBF24"  # Amber-400
  variables:
    surface-background: "#0B1120"
    surface-card: "#111827"
    fg-primary: "#F1F5F9"
    border: "#334155"
    radius: 1rem
```

Every variable takes a color, except `radius`, which takes a length in `px`, `rem`, or `em`. Light and dark modes are independent: a variable set only under `light` keeps Rill's default in dark mode.

A variable can also be set directly in the `light` or `dark` block, next to `primary`, without the `variables` key:

```yaml
type: theme
light:
  primary: "#0F766E"
  surface-background: "#F8FAFC"
  radius: 1rem
```

The two forms are equivalent and can be mixed. If the same variable is set in both places, the value under `variables` is used.

For the full list of variables, what each one styles, and its default in light and dark mode, see the [Theme YAML](/reference/project-files/themes#variables) reference page.

### Chart Color Palettes

Charts take their colors from three palettes, which are set with variables as well:
- **Sequential palette** (`color-sequential-1` through `9`): For ordered data progressions
- **Diverging palette** (`color-diverging-1` through `11`): For data diverging from a midpoint
- **Qualitative palette** (`color-qualitative-1` through `24`): For categorical data

```yaml
type: theme
light:
  primary: "#4F46E5" # Indigo-600
  secondary: "#8B5CF6" # Purple-500
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
  primary: "#818CF8" # Indigo-400
  secondary: "#A78BFA" # Purple-400
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

:::info Theme properties

For more details about configuring themes, you can refer to our [Theme YAML](/reference/project-files/themes) reference page.

:::

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
    primary: "#14B8A6" # Teal-500
    secondary: "#10B981" # Emerald-500
    variables:
      color-sequential-1: "hsl(180deg 80% 95%)"
      color-sequential-5: "hsl(180deg 80% 50%)"
      color-sequential-9: "hsl(180deg 80% 25%)"
  dark:
    primary: "#2DD4BF" # Teal-400
    secondary: "#34D399" # Emerald-400

security:
  access: "{{ .user.admin }} OR '{{ .user.domain }}' == 'example.com'" # only access can be set on dashboard level, see metric view for detailed access policies '{{ .user.domain }}' == 'example.com'"  # only access can be set on dashboard level, see metric view for detailed access policies
```
