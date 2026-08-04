---
title: "Custom Viz Components"
sidebar_label: "Custom Viz"
sidebar_position: 15
---

Custom viz components are reusable visualizations defined as standalone `type: component` files. A component declares typed **params** and renders a [custom chart](/reference/project-files/component) built from a Metrics SQL query and a chart spec. Canvas dashboards reference the component by name and bind values to its params, so one visualization can be reused across many dashboards with different metrics views, measures, and dimensions.

Charts are described declaratively: you state the chart type and which field goes on which channel, and Rill derives the scales, axes, number formats, sorting, stacking, colors, and layout. Rill supplies the semantics of each field from your metrics view — a currency measure formats as currency, the time dimension is treated as temporal at the dashboard's grain — so the spec stays short and the same component adapts when it is bound to a different metrics view.

Unlike an inline `custom_chart` widget, a custom viz lives in its own file (conventionally under `viz_library/`), has a declared parameter contract that Rill validates, and can be developed in a dedicated editor with a live preview.

## Creating a custom viz

In Rill Developer, click **Add → Custom viz → Blank** (or create a file under `viz_library/` by hand):

```yaml
# viz_library/measure_trend.yaml
type: component
display_name: Measure trend

params:
  - name: metrics_view
    type: metrics_view
    required: true
  - name: measure
    type: measure
    required: true
  - name: time_dim
    type: time_dimension
    required: true

custom_chart:
  metrics_sql: |
    SELECT {{ .params.time_dim }}, {{ .params.measure }}
    FROM {{ .params.metrics_view }}
    ORDER BY {{ .params.time_dim }}
  spec:
    chartType: Line Chart
    encodings:
      x: { field: "{{ .params.time_dim }}" }
      y: { field: "{{ .params.measure }}" }
```

There is no schema, sizing, mark, or encoding type to write: `chartType` and `encodings` are the whole spec. An encoding accepts only `field`, `type`, `aggregate`, `sortOrder`, `sortBy`, and `scheme` — and `type` and `aggregate` are rarely needed, since Rill already knows that measures are quantitative and pre-aggregated. Presentation tuning that a chart type exposes (`innerRadius`, `stackMode`, `interpolate`, …) goes under an optional `chartProperties` mapping.

The component editor shows a live preview. Use the **Test values** panel to bind preview-only values to the params; dashboards set their own values. The **Used by** panel lists every dashboard referencing the component.

## Params

Params are the component's contract with dashboards. Each param has a `name`, a `type`, and optionally `required`, `default`, `description`, and (for scalars) `options`:

| Type | Bound value | Notes |
|---|---|---|
| `metrics_view` | A metrics view name | Must be named `metrics_view` or end with `_metrics_view` |
| `measure` | A measure of the bound metrics view | Validated at reconcile time |
| `dimension` | A dimension of the bound metrics view | Validated at reconcile time |
| `time_dimension` | A time dimension of the bound metrics view | The primary time dimension or a time-typed dimension |
| `string`, `number`, `boolean` | A scalar value | Useful for row limits and `chartProperties` |

Params are substituted into `metrics_sql` and `spec` through templating: `{{ .params.measure }}`. A value that is exactly one placeholder keeps the param's type, so a `number` param bound to `innerRadius: "{{ .params.inner_radius }}"` arrives as a number rather than a string. Partial interpolations such as `LIMIT {{ .params.limit }}` are ordinary strings.

When evolving a component that dashboards already use, prefer adding params with a `default` (or `optional`) so existing references keep working. Renaming or removing a param surfaces a validation error on every dashboard that binds it.

## Referencing from a canvas

In the canvas visual editor, open **Add widget → Your components** and pick the component: Rill pre-selects sensible bindings (a metrics view with a time dimension, its first measure, and so on), and the inspector shows a form generated from the declared params. In YAML, a reference looks like:

```yaml
# dashboards/overview.yaml
type: canvas
rows:
  - items:
      - component: measure_trend
        params:
          metrics_view: bids_metrics
          measure: total_bids
          time_dim: __time
        width: 8
      - component: measure_trend
        params:
          metrics_view: bids_metrics
          measure: avg_bid_price
          time_dim: __time
        width: 4
```

Bindings are validated when the canvas reconciles: unknown params, missing required params, and fields that don't exist in the bound metrics view are reported as errors on the dashboard. Canvas-level time and dimension filters apply to custom viz like any other widget.

A time dimension selected by the `metrics_sql` is bucketed at the dashboard's time grain, so a time series re-buckets when the time controls change, just like a native chart. Write `date_trunc('<grain>', <field>) AS <field>` in the query only when a chart should stay at one fixed bucket regardless of the dashboard; the alias is required, or the column comes out named after the expression and the encodings can no longer reference it. The component editor's preview has no time controls to inherit from and renders at day grain.

## Starting from a chart type

**Add → Custom viz** opens a gallery of every chart type Rill can render, grouped by family (bars, lines and areas, distributions, circular, points, tables and maps) and labelled with the fields each one needs — for example a heatmap needs two categorical fields and a measure. Picking one writes a component with a required `metrics_view` param, one typed param per channel named after its chart role (`x_axis`, `color`, …), a templated `metrics_sql`, and a row limit sized to the chart's visual density.

When the AI assistant is enabled, the chart type is handed to it instead, and it authors the same component against your project's own metrics views so the params come pre-bound to real fields. You're taken to the component immediately and watch an "Importing chart using AI…" screen while it generates; the editor appears once the component reconciles.

## Limits

A component issues exactly one Metrics SQL query, so charts that layer several datasets cannot be expressed. Cross-row derivations (running totals, rankings, per-group offsets) are not available in Metrics SQL either — use a chart type that computes them, such as `Waterfall Chart`, `Bump Chart`, or `ECDF Plot`.

Rows that overflow what an axis can legibly fit are dropped and reported above the chart, so a row limit set far above a chart's density silently truncates rather than crowds.

Inline `custom_chart` widgets inside a canvas are a separate, Vega-Lite-based widget and continue to use `vega_spec`. A component file must use `spec`.

:::note Feature flag
Custom viz is currently gated behind the `customComponents` feature flag. Enable it in `rill.yaml`:

```yaml
features:
  - customComponents
```
:::
