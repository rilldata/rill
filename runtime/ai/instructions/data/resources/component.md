---
description: Detailed instructions and examples for developing component (custom viz) resources in Rill
---

# Instructions for developing a component in Rill

## Introduction

Components are reusable visualizations ("custom viz") defined as standalone `type: component` files, conventionally under `viz_library/`. A component declares typed **params** and a **renderer**; canvas dashboards reference the component by name and bind values to its params. This makes one visualization reusable across many dashboards with different metrics views, measures, and dimensions.

Prefer creating a standalone component over an inline `custom_chart` block in a canvas whenever the visualization is non-trivial or could plausibly be reused. Inline blocks bloat the canvas file and cannot be shared.

## Anatomy

A component file's only top-level keys are `type`, `display_name`, `description` (optional), `params`, and one renderer block such as `custom_chart` (which contains `metrics_sql` and `vega_spec`). Never place `metrics_sql` or `vega_spec` at the top level of the YAML, and never wrap the renderer under a `renderer:` key; a file without a proper renderer block parses as an empty draft and renders nothing.

```yaml
# viz_library/measure_trend.yaml
type: component
display_name: Measure trend
description: Line chart of any measure over a time dimension.

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
  - name: limit
    type: number
    default: 500
  - name: smooth
    type: boolean
    default: false

custom_chart:
  metrics_sql: |
    SELECT {{ .params.time_dim }}, {{ .params.measure }}
    FROM {{ .params.metrics_view }}
    ORDER BY {{ .params.time_dim }}
    LIMIT {{ .params.limit }}
  vega_spec: |
    {
      "$schema": "https://vega.github.io/schema/vega-lite/v5.json",
      "width": "container",
      "height": "container",
      "autosize": {"type": "fit"},
      "data": {"name": "query1"},
      "mark": {"type": "line", "interpolate": {"expr": "smooth ? 'monotone' : 'linear'"}},
      "encoding": {
        "x": {"field": "{{ .params.time_dim }}", "type": "temporal"},
        "y": {"field": "{{ .params.measure }}", "type": "quantitative"}
      }
    }
```

The renderer block (`custom_chart:`) is a top-level key of the YAML, a sibling of `params` — never nest it under a `renderer:` key. Exactly one renderer block must be present.

A canvas references it and binds values:

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
          smooth: true
        width: 8
```

## Params

Each param has:

- `name` (required): a valid identifier, referenced in templates as `{{ .params.<name> }}`. The names `title`, `description`, `datum`, `item`, `event`, and `parent` are reserved.
- `type` (required): one of `string`, `number`, `boolean`, `metrics_view`, `measure`, `dimension`, `time_dimension`.
- `required`: if true, every canvas item referencing the component must bind a value. Mutually exclusive with `default`.
- `default`: the value used when the param is not bound. Prefer defaults over `required` when evolving an existing component, so existing dashboards don't break.
- `description`: shown in the visual editor.
- `options`: for scalar params, the allowed values (renders as a select input).
- `metrics_view`: for `measure`/`dimension`/`time_dimension` params, the name of the sibling `metrics_view` param whose bound metrics view the field must belong to. Auto-inferred when exactly one `metrics_view` param is declared.

Constraints enforced by the parser:

- Params of type `metrics_view` must be named `metrics_view` or end with `_metrics_view`.
- Every `{{ .params.X }}` reference in the renderer properties must be a declared param.
- Canvas bindings are validated at reconcile time: unknown params, missing required params, type mismatches, and fields that don't exist in the bound metrics view are all errors.

## Templating vs native Vega-Lite params

There are two complementary mechanisms; use each for what it is for:

1. **Field-shaped params** (`metrics_view`, `measure`, `dimension`, `time_dimension`) must be substituted through templating: `{{ .params.measure }}` inside `metrics_sql` and, where field names appear, inside `vega_spec`. Vega-Lite cannot parameterize field names natively.
2. **Scalar params** (`string`, `number`, `boolean`) are automatically injected as native [Vega-Lite params](https://vega.github.io/vega-lite/docs/parameter.html) at render time. Reference them directly by name in Vega expressions, predicates, and extents (e.g. `"expr": "smooth ? 'monotone' : 'linear'"`); no templating needed. If the `vega_spec` already declares a top-level param with the same name, only its `value` is overridden, so author-supplied `bind`/`expr` are preserved. Scalar params may additionally be templated when needed in SQL (e.g. `LIMIT {{ .params.limit }}`).

## metrics_sql rules

`metrics_sql` queries metrics views as virtual tables, and it is a deliberately restricted dialect — treat it as "pick columns", not general SQL:

- Keep queries trivial: `SELECT <dimensions and measures> FROM <metrics view>` with an `ORDER BY` and `LIMIT`. For parameterized components that means `SELECT {{ .params.<dim> }}, {{ .params.<measure> }} FROM {{ .params.metrics_view }}`.
- Make the sort customizable: declare a required `string` param named `order_by` and write `ORDER BY {{ .params.order_by }} DESC`. The visual editor treats `order_by` specially by name: it renders a dropdown of the field values bound to the component's other params and defaults it to the measure. Time-series charts instead order by the time dimension ascending, without an `order_by` param.
- Declare a `number` param named `limit` (not required, with a default) and end the SQL with `LIMIT {{ .params.limit }}` so dashboards can tune the row count. Choose the default from the chart's visual density: around 10 where each row is a large mark (bars, arcs, text), up to 100 for charts that stay readable with many small marks (point/circle/square scatterplots, rect heatmaps, tick plots), and up to 2000 for time series.
- Every field param the `vega_spec` encodings or the `ORDER BY` reference must also appear in the `SELECT` list; ordering by an unselected field fails at query time, and encodings reading unselected fields render empty.
- NOT supported: CTEs (`WITH`), JOINs, subqueries, window functions, `CASE` expressions, and aggregate functions. Measures are already aggregated; never wrap them in `SUM()`/`AVG()`/`COUNT()`. Grouping by the selected dimensions is implicit.
- Column aliases (`AS x`) are not reflected in the output: result columns always keep the metrics view field names, so reference `{{ .params.<name> }}` in the Vega encodings and `ORDER BY` the field itself, never an alias.
- `date_trunc('<grain>', <time_dimension>)` is the one supported shaping function, for time bucketing.
- Query results are bound to the Vega spec as named datasets `query1`, `query2`, ... in order.

If a visualization needs cross-row derivations (running sums, per-group offsets, rankings), they cannot be expressed here or in Vega: simplify the chart to what direct field encodings support instead of emulating the derivation. Row-local arithmetic on a row's own values (sign flips, ratios, differences) belongs in a Vega `calculate` transform, not in the SQL; see the transforms tip below.

## Extracting an inline custom chart into a component

To convert an existing inline `custom_chart` canvas item into a reusable component:

1. Create `viz_library/<name>.yaml` with `type: component` and move the `custom_chart` block into it verbatim.
2. Declare a `metrics_view` param and replace the metrics view name in each query's `FROM` clause with `{{ .params.metrics_view }}`.
3. For each field that should be configurable, declare a `measure`/`dimension`/`time_dimension` param and replace occurrences of the field name in both `metrics_sql` and `vega_spec` with `{{ .params.<name> }}`. Leave visualization-specific fields hardcoded.
4. Replace the inline block in the canvas item with `component: <name>` plus a `params:` map binding the original values. The dashboard must render identically before and after.

## Tips

- Give the component a clear `display_name` and `description`; they are shown in the visual editor's widget picker.
- Remove Vega transforms that reshape or aggregate the dataset (`filter`, `aggregate`, `joinaggregate`, `window`, `stack`, `fold`, `pivot`, `lookup`, `bin`, `density`, `regression`, ...) and do not try to reproduce their derived values in `metrics_sql` (the dialect can't express them). Simplify the chart to direct field encodings, dropping encodings that depend on them; transforms written against specific data values do not survive rebinding the params to a different metrics view anyway.
- KEEP row-local `calculate` transforms whose derived values the chart's geometry needs: diverging sign flips (population pyramids, diverging stacked bars), ratios/percentages, differences, midpoints, absolute values. Rewrite their `datum` references against the params and keep the `as` name and the encodings that read it. Expose any literal category value the expression compares against as a required `string` param; scalar params are injected as native Vega-Lite params, so reference them directly in the expression. For example, a population pyramid declares `negate_when` (the dimension value plotted on the negative side) and uses:

  ```json
  "transform": [{
    "calculate": "datum['{{ .params.color }}'] === negate_when ? -datum['{{ .params.measure }}'] : datum['{{ .params.measure }}']",
    "as": "signed_value"
  }]
  ```

  with the diverging axis encoded from `"signed_value"` (quantitative) and an axis `labelExpr` showing absolute values.
- `timeUnit` is an encoding property, not a transform: preserve it. When several channels derive from one temporal field via different `timeUnit`s (calendar heatmaps, punch cards, day/hour grids), declare a single `time_dimension` param and reference it from every such channel with its original `timeUnit` and `type`; select the field once in the SQL and order by it ascending with a time-series limit (up to 2000).
- Match every encoding's `type` to the param bound to it: `dimension` → `"nominal"`/`"ordinal"`, `time_dimension` → `"temporal"` (or the original ordinal + `timeUnit`), `measure` → `"quantitative"`. A `"temporal"` type on a plain dimension renders NaN axes. Remove `aggregate` from encodings (measures arrive pre-aggregated), except where a `timeUnit` groups several rows into one mark; keep the example's aggregate there so cells don't overplot.
- Never `bin` a dimension param or type it `"quantitative"`: metrics-view dimensions are categorical strings, so binning them renders a blank chart. When adapting a chart that bins raw numeric axis fields (histograms, binned scatterplots), make those channels nominal/ordinal dimension params without `bin` and keep count/size channels on measure params.
- When adapting an example, copy its `mark` and `config` blocks verbatim; do not invent config or sizing properties the example does not have (invalid config properties silently render a blank chart).
- When converting a Vega-Lite example or an existing chart, name params after their role in the chart (`x_axis`, `y_axis`, `color`, `size`, ...) rather than after the underlying column names, and set a `description` on each param.
- Do not alias templated fields in the SQL: Metrics SQL output columns always keep the metrics view field names, so `AS` aliases are not reflected in the query results. Reference the params in the Vega encodings too (e.g. `"field": "{{ .params.measure }}"`), and `ORDER BY` the field itself (e.g. `ORDER BY {{ .params.time_dim }}`).
- New params on an already-used component should be optional or have a default, otherwise existing dashboards break.
- Components are validated leniently while they contain unresolved `{{ .params.* }}` placeholders; the fully-bound instance is validated in each canvas that references it.
