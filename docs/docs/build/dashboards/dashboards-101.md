---
title: Dashboards in Rill
description: Understanding Explore vs Canvas dashboards and when to use each
sidebar_label: Dashboard Types
sidebar_position: 00
---

Rill offers two distinct dashboard experiences, each optimized for different use cases and user needs. The key difference lies in how they handle **metrics views** - the data sources that power your dashboards.

- **Explore Dashboards** - Standardized _opinionated_ dashboards with consistent data components, visualizations, and layout structure that promotes slice-and-dice discovery and interactive data exploration
- **Canvas Dashboards** - Traditional dashboards that allow users to customize visualizations and layout according to their specific requirements, enabling comprehensive multi-metric reporting


## Built on Metrics Views

All dashboards in Rill are built on [metrics views](/build/metrics-view/metrics-view.md). This ensures that your defined measures and dimensions are standard throughout all of your visualizations. The primary difference between Explore and Canvas dashboards is that **Explore is built on a single metrics view**, while **Canvas can combine multiple metrics views** in one dashboard.

## Explore Dashboards

**[Explore dashboards](/build/dashboards/explore)** are designed for interactive data exploration and analysis. They provide a powerful "slice and dice" experience focused on a single metrics view. Some key features of our Explore Dashboards include:

- [Pivot Tables](/explore/dashboard-101/pivot)
- [Time Dimension Detail](/explore/dashboard-101/tdd)
- [Leaderboards](/explore/dashboard-101/multi-metrics)

<img src = '/img/build/dashboard/explore.png' class='rounded-gif' />

## Canvas Dashboards

**[Canvas dashboards](/build/dashboards/canvas)** provide a traditional dashboard experience where you can combine multiple metrics views and create custom layouts. See our [list of components](/build/dashboards/canvas-components) to start creating your custom dashboard.


<img src = '/img/build/canvas/RTBAds-Canvas-2.png' class='rounded-gif' />

## Key Features

Both dashboard types share common capabilities that make them powerful for data analysis:

- **Time filtering and Comaparisons** - Navigate through time periods and compare different time ranges
- **Dimension filtering** - Filter data by specific dimensions and values

## Learn More About Using Dashboards

For comprehensive guidance on using dashboards as an analyst, see our [Analyst's Guide to Dashboards](/explore/dashboard-101), which covers:

- [Dashboard navigation and interface](/explore/dashboard-101)
- [Filtering and comparisons](/explore/filters)
- [Pivot tables](/explore/dashboard-101/pivot)
- [Time dimension details](/explore/dashboard-101/tdd)
- [Bookmarks and sharing](/explore/bookmarks)
- [Alerts and notifications](/explore/alerts)

<!-- Separate this list into two once explore changes merged -->
<!-- Canvas Dashboard Specific: -->
  
## Live Demo

See both dashboard types in action with our interactive demos:

- **[E-commerce Explore Dashboard](https://ui.rilldata.com/demo/ezcommerce-demo/explore/data_metrics_explore)** - Interactive exploration of sales data
- **[Programmatic Advertising Canvas Dashboard](https://ui.rilldata.com/demo/rill-openrtb-prog-ads/canvas/executive_overview)** - Multi-metrics executive dashboard



