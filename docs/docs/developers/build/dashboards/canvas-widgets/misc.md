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

### Dynamic Markdown

The `metrics_sql` template function allows you to execute a Metrics SQL query and display the results as formatted text in your dashboard. This is useful for displaying aggregated data or custom metrics.


The `metrics_sql_rows` template function allows you to execute a Metrics SQL query and iterate over the results, displaying each row with custom formatting. This is useful for displaying multiple data points in a structured format.

Use `metrics_sql_rows` to get query results and then iterate over them using Go template syntax. You can access columns using dot notation (`.column_name`) or by assigning a variable name in the range:

<ImageCodeToggle
  image="/img/build/dashboard/canvas/components/markdown.png"
  imageAlt="Metrics SQL rows component showing multiple data rows"
  code={`- markdown:
      content: |-
        {{ $rows := metrics_sql_rows "SELECT total_bids, adomain FROM bids_metrics ORDER BY total_bids DESC LIMIT 3" }}
        {{ range $index, $row := $rows }}
        <div style="font-weight: 600;">{{ $row.publisher }}</div>
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

For more information on Metrics SQL syntax and capabilities, see the [Metrics SQL API documentation](/developers/build/metrics-view/metrics-sql).

## Image

Image widgets let you embed images, logos, and visual elements into your dashboards. The `url` is either an absolute URL or the path of an image file in your project.

<ImageCodeToggle
  image="/img/build/dashboard/canvas/components/image.png"
  imageAlt="Image component showing embedded logo"
  code={`- image:
      url: https://cdn.prod.website-files.com/659ddac460dbacbdc813b204/660b0f85094eb576187342cf_rill_logo_sq_gradient.svg
    width: 6`}
  codeLanguage="yaml"
/>

### Images from Your Project

Put image files in the `public/` folder of your project, in any subfolder, and reference them by their path from the project root:

```yaml
- image:
    url: public/img/brand/logo.svg
  width: 6
```

Rill loads these files with the viewer's access to the project, so the same path works in Rill Developer, on Rill Cloud, and in embedded dashboards. PNG, JPEG, GIF, WebP, AVIF, and SVG images are supported.

To serve files from other folders, list them in [`public_paths`](/reference/project-files/rill-yaml#public_paths) in `rill.yaml`.

### Dark Mode

To show a different image when the dashboard is viewed in dark mode, set `dark_url`. When `dark_url` is not set, `url` is used in both modes.

```yaml
- image:
    url: public/logo-light.png
    dark_url: public/logo-dark.png
  width: 6
```

### Images in Markdown

Project images also work in [Text/Markdown](#textmarkdown) widgets, with markdown or HTML syntax. Use HTML when you need to set a size or place a logo inline with text:

```yaml
- markdown:
    content: |-
      ![Logo](public/img/brand/logo.svg)

      <img src="public/img/brand/logo.svg" height="32" alt="Logo" /> **Quarterly review**
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