---
description: Detailed instructions and examples for developing component (custom viz) resources in Rill
---

# Instructions for developing a component in Rill

## Introduction

Components are reusable visualizations ("custom viz") defined as standalone `type: component` files, conventionally under `viz_library/`. A component declares typed **params** and a **renderer**; canvas dashboards reference the component by name and bind values to its params. This makes one visualization reusable across many dashboards with different metrics views, measures, and dimensions.

Charts are described with a declarative chart spec: you state the chart type and which field goes on which channel, and Rill derives the scales, axes, number formats, sorting, stacking, colors and layout from the data and from the metrics view's semantics. **In a chart spec, do not set axis titles, tick counts, colors, fonts, widths or heights anywhere**; they are computed. Reach for Vega-Lite only when the chart spec cannot express what was asked, and then only as described in "Going beyond the chart spec" below.

## Anatomy

A component file's only top-level keys are `type`, `display_name`, `description` (optional), `params`, and one renderer block such as `custom_chart` (which contains `metrics_sql` and `spec`). **Never place `metrics_sql` or `spec` at the top level of the YAML, and never wrap the renderer under a `renderer:` key.** A file without a proper renderer block parses as an empty draft and renders nothing: it reports no error, so a misplaced block is silent.

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

custom_chart:
  metrics_sql: |
    SELECT {{ .params.time_dim }}, {{ .params.measure }}
    FROM {{ .params.metrics_view }}
    ORDER BY {{ .params.time_dim }}
    LIMIT {{ .params.limit }}
  spec:
    chartType: Line Chart
    encodings:
      x: { field: "{{ .params.time_dim }}" }
      y: { field: "{{ .params.measure }}" }
```

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

A value that consists of exactly one `{{ .params.X }}` placeholder keeps the param's type, so a `number` param used as `innerRadius: "{{ .params.inner_radius }}"` arrives as a number, not a string. Partial interpolations such as `LIMIT {{ .params.limit }}` are ordinary strings.

## Creating a new component

A new component is written from a chart type and a set of channels, without being told which data to use:

- Pick the metrics view in the project that best fits the chart's shape, and **bind every field param to a field that metrics view actually declares** — never invent field names, and confirm them by reading the metrics view rather than guessing from the component's own naming.
- **Declare a required `metrics_view` param, plus one required param per channel the chart binds, typed `dimension`, `measure` or `time_dimension`.** Field params are never `string` and never have a default.
- Bind only the channels the chart type calls for; an unused channel left in the spec renders empty.

## The chart spec

`spec` is a YAML mapping with three keys:

- `chartType` (required): one of the chart types listed below. The name must match exactly.
- `encodings` (required): maps a visual channel to a field. Each value is either a field name or a mapping.
- `chartProperties` (optional): per-chart-type presentation tuning. Properties that don't apply to the chart type are ignored.

**An encoding mapping supports exactly these six keys, and nothing else:**

| Key | Purpose |
| --- | --- |
| `field` | The column to bind. Usually `"{{ .params.<name> }}"`. |
| `type` | `quantitative`, `nominal`, `ordinal` or `temporal`. Rarely needed: Rill supplies each field's semantics from the metrics view, so the chart already knows measures are quantitative, dimensions categorical, and the time dimension temporal. |
| `aggregate` | `count`, `sum`, `average` or `mean`. Rarely needed: measures selected from a metrics view are already aggregated. |
| `sortOrder` | `ascending` or `descending`. |
| `sortBy` | Channel or field to sort this channel's values by, e.g. `x` or `y`. |
| `scheme` | Color scheme name for the `color` channel, e.g. `viridis`, `tableau10`. |

The Vega-Lite keys `bin`, `stack`, `timeUnit`, `axis`, `legend`, `format`, `scale`, `mark` and `config` do not exist here. Binning, stacking, time bucketing, axis and legend configuration, and number formatting are all derived; a chart type that needs one applies it itself.

Common channels: `x`, `y`, `color`, `size`, `shape`, `opacity`, `column` and `row` (small multiples), `group` (dodge/cluster category, distinct from `color`), `detail`, `order`, `angle`, `goal` (bullet target). Which channels a chart type accepts varies; bind only the ones it uses.

There is no `title` in a chart spec. Set `display_name` on the component instead.

### Chart types

Points: `Scatter Plot`, `Regression`, `Connected Scatter Plot`, `Ranged Dot Plot`, `Strip Plot`

Bars: `Bar Chart`, `Grouped Bar Chart`, `Stacked Bar Chart`, `Lollipop Chart`, `Waterfall Chart`, `Gantt Chart`, `Bullet Chart`

Distributions: `Histogram`, `Density Plot`, `ECDF Plot`, `Violin Plot`, `Boxplot`, `Pyramid Chart`, `Candlestick Chart`

Lines and areas: `Line Chart`, `Sparkline`, `Bump Chart`, `Slope Chart`, `Area Chart`, `Streamgraph`, `Range Area Chart`

Circular: `Pie Chart`, `Donut Chart`, `Rose Chart`, `Radar Chart`

Tables: `Heatmap`, `Bar Table`

Single numbers are not a chart type here: a canvas has a native KPI widget for those.

### Examples

A ranked bar chart. Metrics SQL picks the top rows; the chart orders the bars:

```yaml
custom_chart:
  metrics_sql: |
    SELECT {{ .params.dimension }}, {{ .params.measure }}
    FROM {{ .params.metrics_view }}
    ORDER BY {{ .params.measure }} DESC
    LIMIT {{ .params.limit }}
  spec:
    chartType: Bar Chart
    encodings:
      y: { field: "{{ .params.dimension }}", sortBy: x, sortOrder: descending }
      x: { field: "{{ .params.measure }}" }
```

A donut, with the hole exposed as a param:

```yaml
  spec:
    chartType: Pie Chart
    encodings:
      size: { field: "{{ .params.measure }}" }
      color: { field: "{{ .params.dimension }}" }
    chartProperties:
      innerRadius: "{{ .params.inner_radius }}"
```

Small multiples: one panel per value of a dimension:

```yaml
  spec:
    chartType: Line Chart
    encodings:
      column: { field: "{{ .params.facet }}" }
      x: { field: "{{ .params.time_dim }}" }
      y: { field: "{{ .params.measure }}" }
      color: { field: "{{ .params.series }}" }
```

## Going beyond the chart spec: ejecting to Vega-Lite

The chart spec is deliberately narrow: a chart type, six encoding keys, and `chartProperties`. When a
request needs something it cannot reach — a mark it has no chart type for, layered or dual-axis
composition, a data transform, an interaction, or specific control over an axis, legend, scale or
number format — the component **ejects**: `spec` is replaced by `vega_spec`, a Vega-Lite v5 spec
written as a string. The two are mutually exclusive; a component that declares both fails to parse.

Try the chart spec first. Ejecting is a one-way trade: Rill stops deriving scales, axes, formats,
sorting, stacking and layout, so the chart no longer adapts to its data, and everything the chart
spec used to compute becomes yours to write and maintain. Eject when the request genuinely cannot be
met otherwise, not to gain incidental control. If a listed chart type does what was asked, use it.
Say plainly in your summary that the component was ejected and what it gave up.

Nothing else about the component changes: keep the `params` and the `metrics_sql` exactly as they
were, so every dashboard already referencing the component keeps working. Only the renderer's chart
description changes.

### Keeping an ejected component parameterized

A component is reusable because its bindings are templated, and an ejected spec must stay that way.
Never write a field name, an axis title or a number format as a literal — bind each one back to the
param it came from:

| What you need | Write |
| --- | --- |
| A field name, anywhere: encodings, transforms, `datum[...]` expressions, tooltips | `{{ .params.<param> }}` |
| A field's label, for an axis, legend or tooltip title | `{{ .fields.<param>.display_name }}` |
| A measure's number format | `formatType: "{{ .fields.<param>.format_type }}"` |

`.fields.<param>` exposes the metrics view metadata of a **field-typed** param (`measure`,
`dimension`, `time_dimension`) and is an error on any other param. Its properties are `name` (the
same value as `{{ .params.<param> }}`), `display_name`, `format_d3`, `format_preset` and
`format_type`. A mistyped property is rejected at parse time.

Prefer `formatType` with `format_type` over a raw `format` string: it routes values through the
formatter Rill registers for that measure, so currency, percent and duration presets render the way
they do everywhere else in the dashboard. Use `format: "{{ .fields.<param>.format_d3 }}"` only when
the spec has to be portable outside Rill.

Scalar params (`string`, `number`, `boolean`) are additionally injected as native Vega-Lite params,
so a spec can reference one as a signal by name in an expression, predicate or extent — in addition
to substituting it textually as `{{ .params.<name> }}`.

### Vega-Lite rules for an ejected component

- Write a valid Vega-Lite v5 spec as a YAML block scalar (`vega_spec: |`).
- Read the query result from `{"data": {"name": "query1"}}`. A component issues exactly one query, so
  there is never a `query2`. Never inline `{"values": [...]}`.
- Set `"width": "container"`, `"height": "container"` and `"autosize": {"type": "fit"}` at the top
  level, so the chart fills its widget instead of overflowing it. `autosize` is not optional: without
  it only the width is fitted and axis and legend decorations run past the bottom of the container.
- Set `type` on every encoding (`quantitative`, `nominal`, `ordinal`, `temporal`). Nothing is derived
  any more, and an unset type silently renders wrong.
- A time dimension selected plainly by the `metrics_sql` is still bucketed at the dashboard's grain,
  so encode it as `"type": "temporal"` and let the query drive the bucket.
- Include a tooltip covering the encoded fields, titled with `display_name` and formatted with
  `format_type`.

```yaml
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
      "mark": {"type": "line", "interpolate": "monotone", "point": true},
      "encoding": {
        "x": {
          "field": "{{ .params.time_dim }}",
          "type": "temporal",
          "title": "{{ .fields.time_dim.display_name }}"
        },
        "y": {
          "field": "{{ .params.measure }}",
          "type": "quantitative",
          "title": "{{ .fields.measure.display_name }}",
          "axis": {"formatType": "{{ .fields.measure.format_type }}"}
        },
        "tooltip": [
          {
            "field": "{{ .params.time_dim }}",
            "type": "temporal",
            "title": "{{ .fields.time_dim.display_name }}"
          },
          {
            "field": "{{ .params.measure }}",
            "type": "quantitative",
            "title": "{{ .fields.measure.display_name }}",
            "formatType": "{{ .fields.measure.format_type }}"
          }
        ]
      }
    }
```

The component editor also has an **Eject to Vega-Lite** action, which compiles the current chart spec
and writes the equivalent parameterized Vega-Lite for the user. It is a good starting point for
someone editing by hand, but it is not available to you: write the `vega_spec` directly.

## metrics_sql rules

`metrics_sql` queries metrics views as virtual tables, and it is a deliberately restricted dialect — treat it as "pick columns", not general SQL:

- Keep queries trivial: `SELECT <dimensions and measures> FROM <metrics view>` with an `ORDER BY` and `LIMIT`. For parameterized components that means `SELECT {{ .params.<dim> }}, {{ .params.<measure> }} FROM {{ .params.metrics_view }}`.
- Order a time series by its time dimension ascending, and anything else by the main measure descending so the `LIMIT` keeps the rows the chart should show. To let a dashboard choose the sort instead, declare a required `string` param named `order_by` and write `ORDER BY {{ .params.order_by }} DESC`: the visual editor treats `order_by` specially by name, rendering a dropdown of the field values bound to the component's other params and defaulting it to the measure.
- Declare a `number` param named `limit` (not required, with a default) and end the SQL with `LIMIT {{ .params.limit }}` so dashboards can tune the row count. Choose the default from the chart's visual density: around 10 where each row is a large mark (bars, arcs, text), up to 100 for charts that stay readable with many small marks (scatterplots, heatmaps, strip plots), and up to 2000 for time series. Rows beyond what the axis can legibly fit are dropped and reported as a warning, so an oversized limit silently truncates the chart.
- **Every field the `spec` encodings or the `ORDER BY` reference must also appear in the `SELECT` list.** Ordering by an unselected field fails at query time, and encodings reading unselected fields render empty.
- NOT supported: CTEs (`WITH`), JOINs, subqueries, window functions, `CASE` expressions, and aggregate functions. Measures are already aggregated; never wrap them in `SUM()`/`AVG()`/`COUNT()`. Grouping by the selected dimensions is implicit.
- Renaming a dimension or measure with `AS` is rejected: result columns keep the metrics view field names, so reference `{{ .params.<name> }}` in the encodings and `ORDER BY` the field itself, never an alias.
- **Select a time dimension plainly; do not bucket it with `date_trunc`.** A selected time dimension is bucketed automatically at the grain of whatever time controls the chart is placed under, so a plain `SELECT {{ .params.<time_dim> }}` is what lets a dashboard drive the chart. Reach for `date_trunc('<grain>', <time_dimension>)` only when the chart is meant to stay at one fixed bucket regardless of the dashboard, and then alias it back to the field's own name — `date_trunc('month', {{ .params.<time_dim> }}) AS {{ .params.<time_dim> }}` — because an unaliased `date_trunc` produces a column named after the expression, which no encoding can reference.
- A component issues exactly one query. Charts that need several datasets cannot be expressed.

If a visualization needs cross-row derivations (running sums, per-group offsets, rankings), they cannot be expressed here: pick a chart type that computes what you need (`Waterfall Chart`, `Bump Chart`, `ECDF Plot`) or simplify the chart to direct field encodings.

## Tips

- Give the component a clear `display_name` and `description`; they are shown in the visual editor's widget picker.
- Name params after their role in the chart (`x_axis`, `y_axis`, `color`, `size`, ...) rather than after the underlying column names, and set a `description` on each param.
- **Do not set `type` or `aggregate` on an encoding** unless the derived behaviour is actually wrong; the semantics Rill supplies are usually right, and hardcoding them is what breaks a component when it is rebound to a different metrics view. A dimension typed `quantitative` renders a blank chart.
- New params on an already-used component should be optional or have a default, otherwise existing dashboards break.
- Components are validated leniently while they contain unresolved `{{ .params.* }}` placeholders; the fully-bound instance is validated in each canvas that references it.
- Inline `custom_chart` blocks inside a canvas are a different, Vega-Lite-based widget and always use `vega_spec`; they never accept `spec`. A component file authors `spec` by default and switches to `vega_spec` only when ejected (see "Going beyond the chart spec").
