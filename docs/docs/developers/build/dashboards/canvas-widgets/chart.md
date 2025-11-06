---
title: "Chart Widgets"
sidebar_label: "Chart"
sidebar_position: 10
---

import ImageCodeToggle from '@site/src/components/ImageCodeToggle';

Chart widgets in Rill Canvas allow you to visualize your data in various formats. You can create charts dynamically in the Canvas Dashboard or through individual component files. For more information, refer to our [Components reference doc](/reference/project-files/component).

## Bar Chart

Bar charts are ideal for comparing values across different categories.

<ImageCodeToggle
  image="/img/build/dashboard/canvas/components/bar.png"
  imageAlt="Bar chart example showing sales data"
  code={`- bar_chart:
      metrics_view: bids_metrics
      color: primary
      x:
        field: advertiser_name
        limit: 20
        showNull: true
        type: nominal
        sort: -y
      y:
        field: total_bids
        type: quantitative
        zeroBasedOrigin: true`}
  codeLanguage="yaml"
         
/>

## Line Chart

Line charts are perfect for showing trends over time.

<ImageCodeToggle
  image="/img/build/dashboard/canvas/components/line.png"
  imageAlt="Line chart example showing revenue trends"
  code={`- line_chart:
      metrics_view: bids_metrics
      color:
        field: device_os
        limit: 3
        type: nominal
      x:
        field: __time
        limit: 20
        sort: -y
        type: temporal
      y:
        field: total_bids
        type: quantitative
        zeroBasedOrigin: true`}
  codeLanguage="yaml"
/>

## Stacked Area Chart

Area charts show the magnitude of change over time with filled areas.

<ImageCodeToggle
  image="/img/build/dashboard/canvas/components/stacked-area.png"
  imageAlt="Area chart showing filled areas over time"
  code={`- area_chart:
      color:
        field: app_or_site
        type: nominal
      metrics_view: auction_metrics
      x:
        field: __time
        limit: 20
        showNull: true
        type: temporal
      y:
        field: requests
        type: quantitative
        zeroBasedOrigin: true`}
  codeLanguage="yaml"
/>

## Stacked Bar Chart

Stacked bar charts show multiple data series stacked on top of each other.

<ImageCodeToggle
  image="/img/build/dashboard/canvas/components/stacked-bar.png"
  imageAlt="Stacked bar chart showing multiple measures"
  code={`- stacked_bar:
      color:
        field: rill_measures
        legendOrientation: top
        type: value
      metrics_view: bids_metrics
      x:
        field: __time
        limit: 20
        type: temporal
      y:
        field: clicks
        fields:
          - video_starts
          - video_completes
          - ctr
          - clicks
          - ecpm
          - impressions
        type: quantitative
        zeroBasedOrigin: true`}
  codeLanguage="yaml"
/>

## Stacked Bar Normalized

Normalized stacked bars show proportions instead of absolute values.

<ImageCodeToggle
  image="/img/build/dashboard/canvas/components/stacked-bar-normalized.png"
  imageAlt="Normalized stacked bar chart showing proportions"
  code={`- stacked_bar_normalized:
      color:
        field: username
        limit: 3
        type: nominal
      metrics_view: rill_commits_metrics
      x:
        field: date
        limit: 20
        type: temporal
      y:
        field: number_of_commits
        type: quantitative
        zeroBasedOrigin: true`}
  codeLanguage="yaml"
/>

## Donut Chart

Donut charts display data as segments of a circle with a hollow center.

<ImageCodeToggle
  image="/img/build/dashboard/canvas/components/donut.png"
  imageAlt="Donut chart showing data segments"
  code={`- donut_chart:
      color:
        field: username
        limit: 20
        type: nominal
      innerRadius: 50
      measure:
        field: number_of_commits
        type: quantitative
      metrics_view: rill_commits_metrics`}
  codeLanguage="yaml"
/>

## Funnel Chart

Funnel charts show the flow through a process with decreasing/increasing values at each stage. 

<ImageCodeToggle
  image="/img/build/dashboard/canvas/components/funnel.png"
  imageAlt="Funnel chart showing process flow"
  code={`- funnel_chart:
      color: stage
      measure:
        field: total_users_measure
        type: quantitative
      metrics_view: Funnel_Dataset_metrics
      mode: width
      stage:
        field: stage
        limit: 15
        type: nominal`}
  codeLanguage="yaml"
/>

## Heat Map

Heat maps visualize data density using color intensity across two dimensions.

<ImageCodeToggle
  image="/img/build/dashboard/canvas/components/heatmap.png"
  imageAlt="Heat map showing data density"
  code={`- heatmap:
      color:
        field: total_bids
        type: quantitative
      metrics_view: bids_metrics
      x:
        field: day
        limit: 10
        type: nominal
        sort:
          - Sunday
          - Monday
          - Tuesday
          - Wednesday
          - Thursday
          - Friday
          - Saturday
      y:
        field: hour
        limit: 24
        type: nominal
        sort: y`}
  codeLanguage="yaml"
/>

## Combo Chart

Combo charts combine different chart types (like bars and lines) in a single visualization.

<ImageCodeToggle
  image="/img/build/dashboard/canvas/components/combo.png"
  imageAlt="Combo chart combining bars and lines"
  code={`- combo_chart:
      color:
        field: measures
        legendOrientation: top
        type: value
      metrics_view: auction_metrics
      x:
        field: __time
        limit: 20
        type: temporal
      y1:
        field: 1d_qps
        mark: bar
        type: quantitative
        zeroBasedOrigin: true
      y2:
        field: requests
        mark: line
        type: quantitative
        zeroBasedOrigin: true`}
  codeLanguage="yaml"
/>

