---
title: "Custom Viz Components"
sidebar_label: "Custom Viz"
sidebar_position: 15
---

Custom viz components are reusable visualizations defined as standalone `type: component` files. A component declares typed **params** and renders a [custom chart](/reference/project-files/component) built from Metrics SQL queries and a Vega-Lite spec. Canvas dashboards reference the component by name and bind values to its params, so one visualization can be reused across many dashboards with different metrics views, measures, and dimensions.

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
  vega_spec: |
    {
      "$schema": "https://vega.github.io/schema/vega-lite/v5.json",
      "width": "container",
      "height": "container",
      "autosize": {"type": "fit"},
      "data": {"name": "query1"},
      "mark": "line",
      "encoding": {
        "x": {"field": "{{ .params.time_dim }}", "type": "temporal"},
        "y": {"field": "{{ .params.measure }}", "type": "quantitative"}
      }
    }
```

The component editor shows a live preview. Use the **Test values** panel to bind preview-only values to the params; dashboards set their own values. The **Used by** panel lists every dashboard referencing the component.

## Params

Params are the component's contract with dashboards. Each param has a `name`, a `type`, and optionally `required`, `default`, `description`, and (for scalars) `options`:

| Type | Bound value | Notes |
|---|---|---|
| `metrics_view` | A metrics view name | Must be named `metrics_view` or end with `_metrics_view` |
| `measure` | A measure of the bound metrics view | Validated at reconcile time |
| `dimension` | A dimension of the bound metrics view | Validated at reconcile time |
| `time_dimension` | A time dimension of the bound metrics view | The primary time dimension or a time-typed dimension |
| `string`, `number`, `boolean` | A scalar value | Injected as native [Vega-Lite params](https://vega.github.io/vega-lite/docs/parameter.html) |

Field-typed params are substituted into `metrics_sql` and `vega_spec` through templating: `{{ .params.measure }}`. Scalar params are additionally available directly in Vega expressions by name (e.g. `"expr": "smooth ? 'monotone' : 'linear'"`), and if the spec declares a Vega-Lite param with the same name, its `bind` and `expr` are preserved and only its value is overridden.

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

## Extracting an inline custom chart

To turn an existing inline `custom_chart` widget into a reusable component, open the widget's **⋯** menu and choose **Save as custom viz…**. Rill writes the component file, lifts the metrics view (and any fields you check) into params, and replaces the inline widget with a reference — the dashboard renders identically. The inverse also exists: **Detach copy** on a referenced widget materializes its current bindings back into an inline chart you can tweak independently.

## Starting from a Vega-Lite example

**Add → Custom viz → From Vega-Lite example…** opens a gallery of examples from the [Vega-Lite example gallery](https://vega.github.io/vega-lite/examples/) (vendored, so it works offline). Importing an example takes only its visual design: the example's spec is handed to the AI assistant, which authors the component against your project's data — a required `metrics_view` param plus one typed param per encoded field (`dimension`, `measure`, or `time_dimension`), params named after their chart role (`x_axis`, `color`, …), a simple templated `metrics_sql`, and a transform-free Vega spec. You're taken to the component immediately and watch an "Importing example using AI…" screen while it generates; the editor appears once the component reconciles. When the AI assistant is unavailable, simple encoding-driven examples are converted deterministically instead, and complex ones import verbatim with their sample data for manual conversion.

:::note Feature flag
Custom viz is currently gated behind the `customComponents` feature flag. Enable it in `rill.yaml`:

```yaml
features:
  - customComponents
```
:::
