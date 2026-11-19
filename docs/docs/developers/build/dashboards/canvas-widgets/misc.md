---
title: "Miscellaneous Widgets"
sidebar_label: "Miscellaneous"
sidebar_position: 20
---

import ImageCodeToggle from '@site/src/components/ImageCodeToggle';

Miscellaneous widgets in Rill Canvas provide additional functionality for text, images, and other non-data elements. These widgets help enhance your dashboards with rich content. For more information, refer to our [Components reference doc](/reference/project-files/component).

## Text/Markdown

Text widgets allow you to add formatted text, markdown content, and documentation directly to your dashboards. You can also use `metrics_sql` and `metrics_sql_rows` to display data from Metrics SQL queries.

<ImageCodeToggle
  image="/img/build/dashboard/canvas/components/text.png"
  imageAlt="Text component showing markdown formatting"
  code={`- markdown:
      alignment:
        horizontal: left
        vertical: middle
      content: |-
        # Markdown
        *Italic*  
        **Bold**  
        ***Bold Italic***  
        ~~Strikethrough~~

        [Rill Home](https://rilldata.com)
        Inline code: \`print("Hello")\`

        Block code:
        \`\`\`python
        def greet():
            return "Hello, Markdown!"
      width: 6`}
  codeLanguage="yaml"
/>

### Styling

Use `alignment` to position the text within the widget. `horizontal` is `left` (default), `center`, or `right`, and `vertical` is `top`, `middle` (default), or `bottom`.

Text size follows the markdown heading level (`#` to `####`). To style one block of text beyond that, such as its size, weight, or color, write it as HTML with a `style` attribute:

```yaml
- markdown:
    alignment:
      horizontal: center
      vertical: middle
    content: |-
      <div style="font-size: 28px; font-weight: 600; color: #0F766E;">Quarterly review</div>

      Updated every Monday.
  width: 6
```

A color set this way stays the same in light and dark mode. To change colors across the whole dashboard, use a [theme](/developers/build/dashboards/customization#changing-themes--colors) instead. HTML in `content` is sanitized, so scripts and event handlers are removed.

### Dynamic Markdown

The `metrics_sql` template function allows you to execute a Metrics SQL query and display the results as formatted text in your dashboard. This is useful for displaying aggregated data or custom metrics.

The query must return a single value: one row with one column. Otherwise, the widget shows an error instead of its text.

```yaml
- markdown:
    content: |-
      Advertisers placed **{{ metrics_sql "SELECT total_bids FROM bids_metrics" }}** bids
      and spent **{{ metrics_sql "SELECT overall_spend FROM bids_metrics" }}** in this period.
    width: 6
```

The `metrics_sql_rows` template function allows you to execute a Metrics SQL query and iterate over the results, displaying each row with custom formatting. This is useful for displaying multiple data points in a structured format.

Use `metrics_sql_rows` to get query results and then iterate over them using Go template syntax. You can access columns using dot notation (`.column_name`) or by assigning a variable name in the range:

<ImageCodeToggle
  image="/img/build/dashboard/canvas/components/markdown.png"
  imageAlt="Metrics SQL rows component showing multiple data rows"
  code={`- markdown:
      content: |-
        {{ $rows := metrics_sql_rows "SELECT total_bids, adomain FROM bids_metrics ORDER BY total_bids DESC LIMIT 3" }}
        {{ range $index, $row := $rows }}
        <div>Domain: {{ $row.adomain }}</div>
        <div>Total: {{ $row.total_bids }}</div>
        {{ end }}
      width: 6`}
  codeLanguage="yaml"
/>

Alternatively, you can use dot notation directly without assigning a variable:

```yaml
- markdown:
    content: |-
      {{ $rows := metrics_sql_rows "SELECT publisher, total_records FROM ad_bids_metrics LIMIT 5" }}
      {{ range $rows }}
      <div>{{ .publisher }}: {{ .total_records }}</div>
      {{ end }}
```

The `metrics_sql_rows` function returns an array of rows that you can iterate over with `{{ range }}`. Access columns using either `{{ $row.column_name }}` (when using named variables) or `{{ .column_name }}` (when using dot notation).

Give the query a `LIMIT` that fits the widget and is at or below the project's [`rill.interactive_sql_row_limit`](/reference/project-files/rill-yaml#env), 10,000 by default. Without a `LIMIT`, the query fails if it returns more rows than the cap.

For more information on Metrics SQL syntax and capabilities, see the [Metrics SQL API documentation](/developers/build/metrics-view/metrics-sql).

### Measure Formatting

`metrics_sql` and `metrics_sql_rows` insert values as the query returns them, such as `6279`. To format measure values the way the metrics view formats them, set `apply_formatting: true`, or turn on **Apply measure value formatting** in the widget's properties panel. Each measure is then shown with its own `format_d3` or `format_preset`, so a currency measure reads `$6.3k`. Dimension values are not changed.

```yaml
- markdown:
    apply_formatting: true
    content: |-
      Advertisers spent **{{ metrics_sql "SELECT overall_spend FROM bids_metrics" }}** in this period.
    width: 6
```

### Canvas Filters and Time Range

Queries in a Markdown widget follow the canvas, like the queries of its other widgets. When a viewer changes the time range, the time zone, or a filter in the [filter bar](/developers/build/dashboards/canvas#filters), Rill runs the queries again and updates the text.

- **Time range**: each query is limited to the canvas time range, applied to the primary time dimension of the metrics view it selects from.
- **Filters**: each query gets the canvas's dimension and measure filters for the metrics view it selects from, including [default filters](/developers/build/dashboards/canvas#default-filters). As for other widgets, a filter applies to every metrics view that has the filtered dimension or measure and to no other. Text that queries two metrics views can therefore show one value filtered and the other unfiltered.
- **The query's own conditions**: a `WHERE` clause in the query is combined with the canvas time range and filters. It can narrow what the canvas selects, but not reach outside of it.

In this example, both values follow the canvas, and the second one counts mobile bids only:

```yaml
- markdown:
    content: |-
      **{{ metrics_sql "SELECT total_bids FROM bids_metrics" }}** bids in this period,
      **{{ metrics_sql "SELECT total_bids FROM bids_metrics WHERE device_type = 'mobile'" }}** of them on mobile.
    width: 6
```

A Markdown widget has no **Time & filters** tab in the properties panel. In YAML, `time_filters` gives its queries their own time range, as on [data widgets](/developers/build/dashboards/canvas-widgets/data#kpi-grid), but `dimension_filters` is ignored; put the condition in the query's `WHERE` clause instead.

```yaml
- markdown:
    # Always the last 7 days, whatever time range the canvas shows
    time_filters: tr=P7D
    content: |-
      **{{ metrics_sql "SELECT total_bids FROM bids_metrics" }}** bids in the last 7 days.
    width: 6
```

Keep these limits in mind:

- **The metrics view needs another widget.** A canvas builds its time range and filter bar from the metrics views of its data, chart and map widgets, not from Markdown. A metrics view that only Markdown queries gets the canvas time range but none of its filters, and `apply_formatting` shows a raw `RILL__FORMAT(...)` placeholder in place of its values. A canvas that has only Markdown widgets has no time range and no filters to choose from, so its queries cover all the data. Adding one widget on the metrics view, such as a KPI, lifts all of these.
- **No time comparison.** Queries get the canvas time range but not its comparison period, so the text can't show a change against the previous period.

## Image

Image widgets let you embed images, logos, and visual elements into your dashboards. Put files in your public/ folder and reference them directly in the `url` as `public/image.png`.

To show a different image when the dashboard is viewed in dark mode, set `dark_url`. When `dark_url` is not set, `url` is used in both modes.

```yaml
- image:
    url: public/logo-light.png
    dark_url: public/logo-dark.png
  width: 6
```

<ImageCodeToggle
  image="/img/build/dashboard/canvas/components/image.png"
  imageAlt="Image component showing embedded logo"
  code={`- image:
      url: https://cdn.prod.website-files.com/659ddac460dbacbdc813b204/660b0f85094eb576187342cf_rill_logo_sq_gradient.svg
    width: 6`}
  codeLanguage="yaml"
/>

### Alignment

An image is scaled to fit the widget without cropping, and centered in it. Use `alignment` to anchor it to a side instead, with the same `horizontal` and `vertical` values as the text widget:

```yaml
- image:
    url: public/logo-light.png
    alignment:
      horizontal: left
      vertical: top
  width: 6
```

## Component

Reference a reusable component created outside of the canvas dashboard.

<ImageCodeToggle
  image="/img/build/dashboard/canvas/components/component.png"
  imageAlt="Table showing detailed data columns"
  code={`- component: my_component`}
  codeLanguage="yaml"
/>

To reuse one chart with different metrics views and fields, define it as a [custom viz](/developers/build/dashboards/canvas-widgets/custom-viz): a component that declares params, which each canvas item binds.

```yaml
- component: bullet_chart
  params:
    metrics_view: bids_metrics
    y_axis: device_type
    x_axis: avg_bid_price
    goal: avg_bid_floor
  width: 6
```

A component file also keeps a large widget out of the canvas file. Rill doesn't parse a project file larger than 128 KB, and the limit applies to each file on its own, so a chart with a long Vega-Lite spec is better kept in its own file than inline. See [Move a Large Chart Out of the Canvas File](/developers/build/dashboards/canvas-widgets/custom-viz#move-a-large-chart-out-of-the-canvas-file).
