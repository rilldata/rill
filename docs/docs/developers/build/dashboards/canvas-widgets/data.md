---
title: "Data Widgets"
sidebar_label: "Data"
sidebar_position: 00
---

import ImageCodeToggle from '@site/src/components/ImageCodeToggle';

Data widgets in Rill Canvas allow you to display raw data in various formats. These widgets are perfect for showing detailed information, metrics, and tabular data. For more information, refer to our [Components reference doc](/reference/project-files/component).

## KPI Grid

KPI grids display key performance indicators in a compact grid format with comparison capabilities. You can select up to 10 concurrent measures to display in a single widget

<ImageCodeToggle
  image="/img/build/dashboard/canvas/components/kpi.png"
  imageAlt="KPI Grid showing metrics with comparisons"
  code={`- kpi_grid:
    comparison:
      - delta
      - percent_change
    metrics_view: auction_metrics
    measures:
      - requests`}
  codeLanguage="yaml"
/>

Widgets follow the canvas time range and time comparison by default. Override either one with `time_filters`, which takes the same `tr` and `compare_tr` parameters as an explore URL. `tr=inherit` keeps the canvas time range, and once `tr` is set a missing `compare_tr` turns delta and comparison off for that widget only. `compare_tr=inherit` follows the canvas comparison, and a comparison range such as `rill-PW` (previous week) or `rill-PY` (previous year) compares the widget against its own previous period. This works on KPI grids, tables, pivots, leaderboards, and time-series charts.

```yaml
- kpi_grid:
    metrics_view: auction_metrics
    measures:
      - requests
    # Canvas time range, comparison off
    time_filters: tr=inherit
```

## Leaderboard

Leaderboards show ranked data with the top performers highlighted.

<ImageCodeToggle
  image="/img/build/dashboard/canvas/components/leaderboard.png"
  imageAlt="Leaderboard showing top performers"
  code={`- leaderboard:
     measures:
       - requests
     metrics_view: auction_metrics
     num_rows: 7
     dimensions:
       - app_site_name`}
  codeLanguage="yaml"
/>

## Pivot/Table

Tables display detailed data in a structured format with customizable columns.

<ImageCodeToggle
  image="/img/build/dashboard/canvas/components/table.png"
  imageAlt="Table showing detailed data columns"
  code={`- table:
    columns:
      - app_site_domain
      - pub_name
      - requests
      - avg_bid_floor
      - 1d_qps
    metrics_view: auction_metrics`}
  codeLanguage="yaml"
/>

## Adhoc measures

Tables, pivots, KPIs, KPI grids, leaderboards, and charts can define their own [adhoc measures](/guide/dashboards/explore/adhoc-measures) with `adhoc_measures`. An adhoc measure combines the metrics view's measures with an arithmetic expression, and you reference it by `name` wherever the widget accepts a measure.

```yaml
- pivot:
    metrics_view: sales_metrics
    row_dimensions:
      - region
    measures:
      - revenue
      - margin_pct
    adhoc_measures:
      - name: margin_pct
        display_name: Margin %
        expression: (revenue - cost) / revenue
        format_preset: percentage
```

In the canvas editor, select **Create adhoc measure** at the bottom of a widget's measure selector to open the same dialog as in Explore dashboards. Rill writes the definition to the widget's `adhoc_measures`, and the measure selector shows an edit action next to it.

| Property | Description |
|----------|-------------|
| `name` | Required. The name you reference in `measures`, `columns`, or a chart field. It must start with a letter, contain only letters, numbers, and underscores, and not match a dimension or measure in the metrics view. |
| `expression` | Required. An arithmetic expression over the metrics view's measure names. See the [expression reference](/guide/dashboards/explore/adhoc-measures#expression-reference). |
| `display_name` | The label shown in the widget. Defaults to `name`. |
| `format_preset` | One of `humanize` (default), `none`, `currency_usd`, `currency_eur`, or `percentage`. |
| `description` | Text shown in the measure's tooltip instead of the expression. |

Adhoc measures belong to a single widget: to use the same calculation in two widgets, define it in both. When you write `adhoc_measures` by hand, follow the same rules as the editor:

- Reference only measures that don't use a window function, don't have required dimensions, and aren't time-comparison measures.
- Don't end a `name` with a comparison suffix, such as `_prev`, `_delta`, `_delta_perc`, or `_percent_of_total`, and don't include `_rill_` in it.
- Define at most 10 adhoc measures per widget.

For a calculation that many widgets or dashboards share, add a measure to the [metrics view](/developers/build/metrics-view) instead.

## Navigation

All Data widgets also provide a button to "Go to explore" that can navigate to the Explore dashboard if available.

![Go To Explore](/img/build/dashboard/canvas/go-to-explore.png)