---
title: "Custom Dashboards"
description: Creating Custom Dashboards in Rill
sidebar_label: "Getting Started"
sidebar_position: 6
---

## Creating Custom Dashboards in Rill 

In this section, we'll cover how to create custom dashboards in Rill Developer and publish these to Rill Cloud,

:::note
This feature is not currently publicly released and is behind a feature flag. In case of specific issues, please reach out to us for assistance. 
:::
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

### Step 2: Adding feature flag (no longer required once GA)

Edit the rill.yaml via the UI and add the following:
```
features:
 - customDashboards
```


### Step 3: Notice the changes in the UI

You will now notice that under the `+Add` UI, two new options are available.

![img](/img/tutorials/301/add-custom-dashboard.png)

Once you select either of these, a dedicated folder `charts` and `custom-dashboards` will be created.

import DocsRating from '@site/src/components/DocsRating';

---
<DocsRating />