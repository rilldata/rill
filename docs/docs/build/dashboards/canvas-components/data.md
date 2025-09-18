---
title: "Data Components"
sidebar_label: "Data"
sidebar_position: 00
---

import ImageCodeToggle from '@site/src/components/ImageCodeToggle';

Data components in Rill Canvas allow you to display raw data in various formats. These components are perfect for showing detailed information, metrics, and tabular data. For more information, refer to our [Components reference doc](/reference/project-files/component).

## KPI Grid

KPI grids display key performance indicators in a compact grid format with comparison capabilities.

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