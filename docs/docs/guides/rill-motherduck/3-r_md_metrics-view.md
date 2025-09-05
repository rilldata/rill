---
title: "3. Create Metrics View in Rill"
sidebar_label: "3. Create Metrics View in Rill"
sidebar_position: 3
hide_table_of_contents: false
tags:
  - OLAP:MotherDuck
  - Tutorial
---

You'll need to use a table that exists in your MotherDuck database. In this tutorial, we'll be using `rill_auction_data`.

:::note
Don't have any good dataset to use? See [Ingest into MotherDuck](./r_md_ingest.md) to ingest directly into MotherDuck from Rill.
:::

### Create metrics view

Let's create a metrics view based on the table using the `Generate metrics via AI` feature.

<img src = '/img/tutorials/md/MotherDuck-metrics-ai.png' class='rounded-gif' />
<br />

### What are we looking at?

This is our metrics view, where we can define measures and dimensions to be used in dashboards.  

```yaml
# Metrics view YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/explore-dashboards
# This file was generated using AI.

version: 1
type: metrics_view

display_name: Auction Data Raw Metrics
connector: motherduck
table: auction_data_raw
timeseries: __time


dimensions:
    ...

measures:
    ...
```




While we go into more details in our [Rill Basics course](/guides/rill-basics/dashboard) and [our documentation](https://docs.rilldata.com/build/dashboards/), let's go over it quickly.

---

`timeseries` - This is our time column that is used as our x-axis for graphs.

`connector` - This is our manually defined MotherDuck connector

`dimensions` - These are our categorical columns that we can use in the dashboard to filter and slice

`measures` - These are our numerical aggregates defined in the metrics layer. We can see functions such as MAX(), COUNT(), and AVG() used on the underlying table.