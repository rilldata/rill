---
title: Dashboards in Rill
description: Understanding Explore vs Canvas dashboards and when to use each
sidebar_label: Overview
sidebar_position: 00
---

Rill offers two distinct dashboard experiences, each optimized for different use cases and user needs. While both are used to visualize your data, the key difference lies in how they handle [**metrics views**](/build/metrics-view) - the data sources that power your dashboards.

- **Explore Dashboards** - Standardized _opinionated_ dashboards with consistent data components, visualizations, and layout structure that promote slice-and-dice discovery and interactive data exploration. These are built over a **single metrics view** using the [One Big Table approach](/build/models/models-101#one-big-table-and-dashboarding).
- **Canvas Dashboards** - Traditional dashboards that allow users to customize visualizations and layout according to their specific requirements utilizing **multiple metrics views**.

## Explore Dashboards

**[Explore dashboards](/build/dashboards/explore)** are designed for interactive data exploration and analysis. They provide a powerful "slice and dice" experience focused on a single metrics view. Some key features of our Explore dashboards include:

- [Pivot Tables](/explore/dashboard-101/pivot)
- [Time Dimension Detail](/explore/dashboard-101/tdd)
- [Leaderboards](/explore/dashboard-101/multi-metrics)

<img src = '/img/build/dashboard/explore.png' class='rounded-gif' /> <br/>

### Live Demo

See both dashboard types in action with our interactive demos:


- **[Programmatic Advertising Explore Dashboard](https://ui.rilldata.com/demo/rill-openrtb-prog-ads/explore/auction_explore?tr=-24h+to+ref+as+of+latest%2FD&tz=UTC&grain=day&measures=1d_qps%2Cavg_bid_floor%2Crequests&dims=pub_name%2Capp_site_name%2Cad_size%2Capp_site_domain%2Cdevice_osv&leaderboard_measures=1d_qps%2Cavg_bid_floor%2Crequests)** - Real-time bidding and auction data exploration
- **[Mobile App Engagement Explore Dashboard](https://ui.rilldata.com/demo/rill-app-engagement/explore/mobile_events_explore)** - Mobile app user behavior and engagement metrics
- **[GitHub Analytics Explore Dashboard](https://ui.rilldata.com/demo/rill-github-analytics/explore/rill_commits_explore?tr=P12M&grain=week&dims=username%2Cfile_path%2Cfilename%2Cfile_extension%2Cfirst_directory%2Csecond_directory%2Cprevious_file_path%2Cis_merge_commit)** - Code repository and commit analysis
- **[E-commerce Explore Dashboard](https://ui.rilldata.com/demo/ezcommerce-demo/explore/data_metrics_explore)** - Interactive exploration of sales data

## Canvas Dashboards

**[Canvas dashboards](/build/dashboards/canvas)** provide a traditional dashboard experience where you can combine multiple metrics views and create custom layouts. See our [list of components](/build/dashboards/canvas-widgets) to start creating your custom dashboard.


<img src = '/img/build/dashboard/canvas/RTBAds-Canvas-2.png' class='rounded-gif' /> <br/>

### Live Demo

See both dashboard types in action with our interactive demos:
- **[E-commerce demo dashboard](https://ui.rilldata.com/demo/ezcommerce-demo/canvas/canvas)** - Sales analytics and revenue tracking with multiple visualizations
- **[Programmatic advertising demo dashboard](https://ui.rilldata.com/demo/rill-openrtb-prog-ads/canvas/executive_overview)** - Real-time bidding metrics and campaign performance analysis
- **[New York City demo dashboard](https://ui.rilldata.com/demo/nyc-canvas-jam/canvas/scorecard%20canvas)** - Urban analytics and city performance metrics
- **[NYC party demo dashboard 🎉](https://ui.rilldata.com/demo/nyc-canvas-jam/canvas/Leaderboard)** - Fun leaderboard showcasing various data insights


## Key Features

Both dashboard types share common capabilities that make them powerful for data analysis:

- **Time filtering and comparisons** - Navigate through time periods and compare different time ranges
- **Dimension filtering** - Filter data by specific dimensions and values

## Learn More About Using Dashboards

For comprehensive guidance on using dashboards as an analyst, see our [Analyst's Guide to Dashboards](/explore/dashboard-101), which covers:

- [Dashboard navigation and interface](/explore/dashboard-101)
- [Filtering and comparisons](/explore/filters)
- [Pivot tables](/explore/dashboard-101/pivot)
- [Time dimension details](/explore/dashboard-101/tdd)
- [Bookmarks and sharing](/explore/bookmarks)
- [Alerts and notifications](/explore/alerts)

