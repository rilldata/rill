---
description: How to chart a metrics view with the create_chart tool, which chart type fits which question, and the pitfalls that produce misleading charts
---

## Charts with create_chart

The `create_chart` tool draws a chart from a metrics view. You describe the chart (`chart_spec`: a chart type and which field goes on which channel) and what to query (`data`: the metrics view, time range, grain and filter); Rill runs the query, formats the measures and renders the chart. Never pass rows, colors or axis settings. The tool description lists the chart types, their channels and the argument shape.

### Before you chart

- Know the fields: use `get_metrics_view` for the measures, dimensions and the time dimension, and reference them by their exact names.
- Look at the data first with `query_metrics_view` whenever the shape matters: how many distinct values a dimension has, whether a measure is a count, an amount or a rate, and whether a dimension mixes a total with its parts. Chart one level, never a total next to its parts.
- One chart answers one question. Prefer fewer series: cap a dimension with `limit`, and split with `row` or `column` facets instead of piling categories onto `color`.
- Pick the grain that gives roughly 10 to 100 points for the range: days for a few weeks, weeks for a quarter, months for a year or more.

### Choosing a chart

- A trend over time: Line Chart, with several measures on `y` or a dimension on `color`; Area Chart to stack parts; Bump Chart for ranks; Slope Chart for two periods.
- Categories compared: Bar Chart for one measure; Stacked Bar Chart with the parts on `color`; Grouped Bar Chart with the clustering dimension on `group`; Bar Table for many categories with value labels; Lollipop Chart for a ranking.
- Share of a whole: Pie Chart with the measure on `size` and the dimension on `color`, `innerRadius` for a donut; only for a handful of values that add up to a whole.
- How a total changed: Waterfall Chart with the step on `x` and the signed change on `y`; leave `color` unbound, it is reserved for a start/delta/end type field.
- Two measures per category: Scatter Plot with the dimension on `detail`, optional `size` and `color`; Heatmap for two dimensions and one measure.
- A distribution: Histogram, Boxplot, Violin Plot or Density Plot need one row per observation, so put the finest dimension on `detail`.

### Titles

Set `title` to the finding in a sentence and `subtitle` to what is measured, of whom, when and in what unit, for example "Spend fell 18% after the March price change" over "Daily total spend by device, Jan–Apr 2024, USD". Axes and legends take their labels from the metrics view, so do not restate field names.

### Pitfalls

- Parts of a whole go on a Stacked Bar Chart with `color`; values compared side by side go on a Grouped Bar Chart with `group`. A Bar Chart with a second dimension on `color` stacks.
- Counts and amounts on one axis share a scale and the smaller one flattens; use `row` facets or separate charts.
- More than about ten values on `color` is unreadable; use `limit`, a Bar Table or facets.
- A dimension with a single value in the range gives one mark; choose another breakdown.
- A chart is a snapshot of its arguments: when the user asks for another period, grain or breakdown, create a new chart.
- Do not restate the chart's numbers in prose; name the insight. The chart shows its data and an "Open in Rill" link.
- When the result reports that values were cut, coarsen `time_grain` or lower the `limit`s and create the chart again.
