---
title: "Canvas Dashboards"
description: Creating Canvas Dashboards in Rill
sidebar_label: "Getting Started"
sidebar_position: 6
---
:::note
Canvas Dashboards are still not released yet and is hidden behind a feature flag. Breaking changes may be pushed at any time and may cause your dashboards to no longer function. Please proceed with that in mind!
:::

## Creating Canvas Dashboards in Rill 

In this section, we'll cover how to create Canvas dashboards in Rill Developer and publish these to Rill Cloud,


### Step 1: Let's return to our project my-rill-tutorial in Rill Developer

:::tip Friends from ClickHouse
If you are coming from the Rill and ClickHouse course, we are using the following datasets!

```
gs://rilldata-public/github-analytics/Clickhouse/*/*/modified_files_*.parquet
gs://rilldata-public/github-analytics/Clickhouse/*/*/commits_*.parquet
```
You will need to comment out the olap_connector value in your rill.yaml
```yaml
#olap_connector: clickhouse
```
You can add the following key pair to your dashboard to continue using ClickHouse:
```yaml
connector: clickhouse
```
:::
From the terminal, let's start rill

```
rill start 
```

### Step 2: Select More! 

Under the + Add dropdown, select More to find the chart and custom dashboard components.

![img](/img/tutorials/301/add-custom-dashboard.png)

Once you select either of these, a dedicated folder `components` and `canvas-dashboards` will be created.


import DocsRating from '@site/src/components/DocsRating';

---
<DocsRating />